// Copyright 2024 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package analyzers

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"github.com/prometheus-operator/poctl/internal/k8sutil"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RunPrometheusAnalyzer(ctx context.Context, clientSets *k8sutil.ClientSets, name, namespace string) error {
	prometheus, err := clientSets.MClient.MonitoringV1().Prometheuses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("Prometheus %s not found in namespace %s", name, namespace)
		}
		return fmt.Errorf("error while getting Prometheus: %v", err)
	}

	cRb, err := clientSets.KClient.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{
		LabelSelector: "prometheus=prometheus",
	})
	if err != nil {
		return fmt.Errorf("failed to list RoleBindings: %w", err)
	}

	if !doesServiceAccountBoundToRoleBindingList(cRb, prometheus.Spec.ServiceAccountName) {
		return fmt.Errorf("ServiceAccount %s is not bound to any RoleBindings", prometheus.Spec.ServiceAccountName)
	}

	for _, crb := range cRb.Items {
		cr, err := clientSets.KClient.RbacV1().ClusterRoles().Get(ctx, crb.RoleRef.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get ClusterRole %s", crb.RoleRef.Name)
		}

		err = checkClusterRoleRules(crb, cr)
		if err != nil {
			return err
		}
	}

	namespaceSelectors := map[string]interface{}{
		"PodMonitorNamespaceSelector":      prometheus.Spec.PodMonitorNamespaceSelector,
		"ServiceMonitorNamespaceSelector":  prometheus.Spec.ServiceMonitorNamespaceSelector,
		"ScrapeConfigNamespaceSelector":    prometheus.Spec.ScrapeConfigNamespaceSelector,
		"ProbeNamespaceSelector":           prometheus.Spec.ProbeNamespaceSelector,
	}

	namespaceSelectorStatus, nilNamespaceSelectors := checkPrometheusNamespaceSelectorsStatus(namespaceSelectors)

	if !namespaceSelectorStatus {
		if len(nilNamespaceSelectors) > 0 {
			//fmt.Printf("No %s is defined, defaulting to the same namespace of Prometheus\n", strings.Join(nilNamespaceSelectors, ", "))
			checkSelectorsCreated := checkServiceDiscovery(ctx,clientSets,namespace)
			if !checkSelectorsCreated{
				fmt.Printf("No Service Selectors are created yet")
			}
		}
	} else {
			fmt.Println("All namespace selectors are either empty or properly defined.")
	}

	serviceSelectors := map[string]interface{}{
		"ServiceMonitorSelector":      prometheus.Spec.ServiceMonitorSelector,
		"PodMonitorSelector":          prometheus.Spec.PodMonitorSelector,
		"ProbeSelector":               prometheus.Spec.ProbeSelector,
		"ScrapeConfigSelector":        prometheus.Spec.ScrapeConfigSelector,
	}

	serviceSelectorStatus, emptyServiceSelectors := checkPrometheusServiceSelectorsStatus(serviceSelectors)

	if !serviceSelectorStatus{
		if emptyServiceSelectors!= nil {
			//fmt.Printf("Selectors are defined: %s\n", strings.Join(emptyServiceSelectors, ", "))
			checkServiceSelectorsCreated := checkServiceDiscovery(ctx,clientSets,namespace)
			if !checkServiceSelectorsCreated{
				fmt.Printf("No Service Selectors are created yet")
			}
		}
	} 

	slog.Info("Prometheus is compliant, no issues found", "name", name, "namespace", namespace)
	return nil
}

func checkClusterRoleRules(crb v1.ClusterRoleBinding, cr *v1.ClusterRole) error {
	cmVerbsStatus := false
	nonresourcesStatus := false
	verbsStatus := true
	verbsToCheck := []string{"get", "list", "watch"}

	for _, rule := range cr.Rules {
		for _, resource := range rule.Resources {
			if resource == "configmaps" {
				for _, verb := range rule.Verbs {
					if verb == "get" {
						cmVerbsStatus = true
					}
				}
			} else {
				for range rule.APIGroups {
					if !containsVerb(rule.Verbs, verbsToCheck) {
						verbsStatus = false
					}
				}
			}
		}
		for _, nonresources := range rule.NonResourceURLs {
			if nonresources == "/metrics" {
				for _, verb := range rule.Verbs {
					if verb == "get" {
						nonresourcesStatus = true
					}
				}
			}
		}
	}

	if !cmVerbsStatus {
		return fmt.Errorf("ClusterRole %s does not does not include 'configmaps' with 'get' in its verbs", crb.RoleRef.Name)
	}

	if !nonresourcesStatus {
		return fmt.Errorf("ClusterRole %s does not have proper verbs for NonResourceURLs", crb.RoleRef.Name)
	}

	if !verbsStatus {
		return fmt.Errorf("ClusterRole %s does not have all the necessary verbs for APIGroups ", crb.RoleRef.Name)
	}
	return nil
}

func containsVerb(ruleVerbs, verbsToCheck []string) bool {
	verbSet := make(map[string]bool)
	for _, v := range ruleVerbs {
		verbSet[v] = true
	}
	for _, v := range verbsToCheck {
		if verbSet[v] {
			return true
		}
	}
	return false
}

func doesServiceAccountBoundToRoleBindingList(clusterRoleBindings *v1.ClusterRoleBindingList, serviceAccountName string) bool {
	for _, roleBinding := range clusterRoleBindings.Items {
		if roleBinding.Subjects != nil {
			for _, subject := range roleBinding.Subjects {
				if subject.Kind == "ServiceAccount" && subject.Name == serviceAccountName {
					return true
				}
			}
		}
	}
	return false
}

func checkPrometheusNamespaceSelectorsStatus(namespaceSelectors map[string]interface{}) (bool, []string) {
    var nilNamespaceSelectors []string
    var emptyNamespaceSelectors []string

    for selectorName, selector := range namespaceSelectors {
        if selector == nil {
            nilNamespaceSelectors = append(nilNamespaceSelectors, selectorName)
        } else {
            selectorStruct, ok := selector.(*metav1.LabelSelector)
            if ok {
                if len(selectorStruct.MatchLabels) == 0 && len(selectorStruct.MatchExpressions) == 0 {
                    emptyNamespaceSelectors = append(emptyNamespaceSelectors, selectorName)
                } else {
                    if len(selectorStruct.MatchLabels) > 0 {
						for labelKey, labelValue := range selectorStruct.MatchLabels {
							fmt.Printf("%s has MatchLabel: %s=%s", selectorName, labelKey, labelValue)
						}
                    }
					if len(selectorStruct.MatchExpressions) > 0 {
						for _, expression := range selectorStruct.MatchExpressions {
							fmt.Printf("%s has MatchExpression: %s %s %v", selectorName, expression.Key, expression.Operator, expression.Values)
						}
					}
                }
            }
        }
    }

    if len(emptyNamespaceSelectors) == 4 {
		//fmt.Printf("%s is empty, Prometheus is watching all namespaces.\n", strings.Join(emptyNamespaceSelectors, ", "))
        return true, nil
    }

    if len(nilNamespaceSelectors) > 0 {
		return false, nilNamespaceSelectors 
    }
    return true, nil
}

func checkPrometheusServiceSelectorsStatus(serviceSelectors map[string]interface{}) (bool, []string) {
	var nilServiceSelectors []string
	var emptyServiceSelectors []string

    for selectorName, selector := range serviceSelectors {
        if selector == nil {
            nilServiceSelectors = append(nilServiceSelectors, selectorName)
        } else {
            selectorStruct, ok := selector.(*metav1.LabelSelector)
            if ok {
                if len(selectorStruct.MatchLabels) == 0 && len(selectorStruct.MatchExpressions) == 0 {
                    emptyServiceSelectors = append(emptyServiceSelectors, selectorName)
                } else {
                    if len(selectorStruct.MatchLabels) > 0 {
						for labelKey, labelValue := range selectorStruct.MatchLabels {
							fmt.Printf("%s has MatchLabel: %s=%s", selectorName, labelKey, labelValue)
						}
                    }
					if len(selectorStruct.MatchExpressions) > 0 {
						for _, expression := range selectorStruct.MatchExpressions {
							fmt.Printf("%s has MatchExpression: %s %s %v", selectorName, expression.Key, expression.Operator, expression.Values)
						}
					}
                }
            }
        }
    }

	if len(nilServiceSelectors) == 4 {
		fmt.Printf("No %s are defined, the Prometheus matches no objects,configuration is unmanaged\n", strings.Join(nilServiceSelectors, ", "))
		return false, nil
	}
	
	if len(emptyServiceSelectors) > 0 {
		//fmt.Printf("%s are empty, Prometheus matches all objects.\n", strings.Join(emptyServiceSelectors, ", "))
		return true, emptyServiceSelectors
	}
	return true, nil
}

func checkServiceDiscovery(ctx context.Context, clientSets *k8sutil.ClientSets, namespace string) bool {
	var existingSelectors []string
	podMonitor, err := clientSets.MClient.MonitoringV1().PodMonitors(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Errorf("Error listing PodMonitors: %v\n", err)
	}
	if len(podMonitor.Items) == 0 {
		existingSelectors = append(existingSelectors,"podMonitor")
	}	

	serviceMonitor, err := clientSets.MClient.MonitoringV1().ServiceMonitors(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Errorf("Error listing serviceMonitor: %v\n", err)
	}
	if len(serviceMonitor.Items) == 0 {
		existingSelectors = append(existingSelectors,"serviceMonitor")
	}
	
	probe, err := clientSets.MClient.MonitoringV1().Probes(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Errorf("Error listing Probes: %v\n", err)
	}
	if len(probe.Items) == 0 {
		existingSelectors = append(existingSelectors,"probe")
	}

	scrapeConfig, err := clientSets.MClient.MonitoringV1alpha1().ScrapeConfigs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Errorf("Error listing scrapeConfig: %v\n", err)
	}
	if len(scrapeConfig.Items) == 0 {
		existingSelectors = append(existingSelectors,"probe")
	}

	if len(existingSelectors) == 0 {
		return false
	}
	return true
}
