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
	"k8s.io/apimachinery/pkg/labels"
)

const (
	ServiceMonitor = "ServiceMonitor"
	PodMonitor     = "PodMonitor"
	Probe          = "Probe"
	ScrapeConfig   = "ScrapeConfig"
	PrometheusRule = "PrometheusRule"
)

func RunPrometheusAnalyzer(ctx context.Context, clientSets *k8sutil.ClientSets, name, namespace string) error {
	prometheus, err := clientSets.MClient.MonitoringV1().Prometheuses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("prometheus %s not found in namespace %s", name, namespace)
		}
		return fmt.Errorf("error while getting Prometheus: %v", err)
	}

	cRb, err := clientSets.KClient.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{
		LabelSelector: "prometheus=prometheus",
	})
	if err != nil {
		return fmt.Errorf("failed to list RoleBindings: %w", err)
	}

	if !k8sutil.IsServiceAccountBoundToRoleBindingList(cRb, prometheus.Spec.ServiceAccountName) {
		return fmt.Errorf("serviceAccount %s is not bound to any RoleBindings", prometheus.Spec.ServiceAccountName)
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

	if err := k8sutil.CheckResourceNamespaceSelectors(ctx, *clientSets, prometheus.Spec.PodMonitorNamespaceSelector); err != nil {
		return fmt.Errorf("podMonitorNamespaceSelector is not properly defined: %s", err)
	}

	if err := k8sutil.CheckResourceNamespaceSelectors(ctx, *clientSets, prometheus.Spec.ProbeNamespaceSelector); err != nil {
		return fmt.Errorf("probeNamespaceSelector is not properly defined: %s", err)
	}

	if err := k8sutil.CheckResourceNamespaceSelectors(ctx, *clientSets, prometheus.Spec.ServiceMonitorNamespaceSelector); err != nil {
		return fmt.Errorf("serviceMonitorNamespaceSelector is not properly defined: %s", err)
	}

	if err := k8sutil.CheckResourceNamespaceSelectors(ctx, *clientSets, prometheus.Spec.ScrapeConfigNamespaceSelector); err != nil {
		return fmt.Errorf("scrapeConfigNamespaceSelector is not properly defined: %s", err)
	}

	if err := k8sutil.CheckResourceNamespaceSelectors(ctx, *clientSets, prometheus.Spec.RuleNamespaceSelector); err != nil {
		return fmt.Errorf("ruleNamespaceSelector is not properly defined: %s", err)
	}

	if err := checkResourceLabelSelectors(ctx, clientSets, prometheus.Spec.ServiceMonitorSelector, ServiceMonitor, namespace); err != nil {
		return fmt.Errorf("serviceMonitorSelector is not properly defined: %s", err)
	}

	if err := checkResourceLabelSelectors(ctx, clientSets, prometheus.Spec.PodMonitorSelector, PodMonitor, namespace); err != nil {
		return fmt.Errorf("podMonitorSelector is not properly defined: %s", err)
	}

	if err := checkResourceLabelSelectors(ctx, clientSets, prometheus.Spec.ProbeSelector, Probe, namespace); err != nil {
		return fmt.Errorf("probeSelector is not properly defined: %s", err)
	}

	if err := checkResourceLabelSelectors(ctx, clientSets, prometheus.Spec.ScrapeConfigSelector, ScrapeConfig, namespace); err != nil {
		return fmt.Errorf("scrapeConfigSelector is not properly defined: %s", err)
	}

	if err := checkResourceLabelSelectors(ctx, clientSets, prometheus.Spec.RuleSelector, PrometheusRule, namespace); err != nil {
		return fmt.Errorf("ruleSelector is not properly defined: %s", err)
	}

	slog.Info("Prometheus is compliant, no issues found", "name", name, "namespace", namespace)
	return nil
}

func checkClusterRoleRules(crb v1.ClusterRoleBinding, cr *v1.ClusterRole) error {
	var errs []string
	verbsToCheck := []string{"get", "list", "watch"}
	missingVerbs := []string{}

	for _, rule := range cr.Rules {
		for _, resource := range rule.Resources {
			found := false
			if resource == "configmaps" {
				for _, verb := range rule.Verbs {
					if verb == "get" {
						found = true
						break
					}
				}
				if !found {
					errs = append(errs, fmt.Sprintf("ClusterRole %s does not include 'configmaps' with 'get' in its verbs", crb.RoleRef.Name))
				}
				continue
			}
			for range rule.APIGroups {
				for _, requiredVerb := range verbsToCheck {
					found := false
					for _, verb := range rule.Verbs {
						if verb == requiredVerb {
							found = true
							break
						}
					}
					if !found {
						missingVerbs = append(missingVerbs, requiredVerb)
					}
				}
				if len(missingVerbs) > 0 {
					errs = append(errs, fmt.Sprintf("ClusterRole %s is missing necessary verbs for APIGroups: %v", crb.RoleRef.Name, missingVerbs))
				}
			}
		}
		for _, nonResource := range rule.NonResourceURLs {
			if nonResource == "/metrics" {
				hasGet := false
				for _, verb := range rule.Verbs {
					if verb == "get" {
						hasGet = true
						break
					}
				}
				if !hasGet {
					errs = append(errs, fmt.Sprintf("ClusterRole %s does not include 'get' verb for NonResourceURL '/metrics'", crb.RoleRef.Name))
				}
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("multiple errors found:\n%s", strings.Join(errs, "\n"))
	}
	return nil
}

func checkResourceLabelSelectors(ctx context.Context, clientSets *k8sutil.ClientSets, labelSelector *metav1.LabelSelector, resourceName, namespace string) error {
	if labelSelector == nil {
		return fmt.Errorf("%s selector is not defined", resourceName)
	}

	if len(labelSelector.MatchLabels) == 0 && len(labelSelector.MatchExpressions) == 0 {
		return nil
	}

	labelMap, err := metav1.LabelSelectorAsMap(labelSelector)
	if err != nil {
		return fmt.Errorf("invalid label selector format in %s: %v", resourceName, err)
	}

	switch resourceName {
	case ServiceMonitor:
		serviceMonitors, err := clientSets.MClient.MonitoringV1().ServiceMonitors(namespace).List(ctx, metav1.ListOptions{LabelSelector: labels.SelectorFromSet(labelMap).String()})
		if err != nil {
			return fmt.Errorf("failed to list ServiceMonitors in %s: %v", namespace, err)
		}
		if len(serviceMonitors.Items) == 0 {
			return fmt.Errorf("no ServiceMonitors match the provided selector in Prometheus %s", namespace)
		}
	case PodMonitor:
		podMonitors, err := clientSets.MClient.MonitoringV1().PodMonitors(namespace).List(ctx, metav1.ListOptions{LabelSelector: labels.SelectorFromSet(labelMap).String()})
		if err != nil {
			return fmt.Errorf("failed to list PodMonitor in %s: %v", namespace, err)
		}
		if len(podMonitors.Items) == 0 {
			return fmt.Errorf("no PodMonitors match the provided selector in Prometheus %s", namespace)
		}
	case Probe:
		probes, err := clientSets.MClient.MonitoringV1().Probes(namespace).List(ctx, metav1.ListOptions{LabelSelector: labels.SelectorFromSet(labelMap).String()})
		if err != nil {
			return fmt.Errorf("failed to list Probes in %s: %v", namespace, err)
		}
		if len(probes.Items) == 0 {
			return fmt.Errorf("no Probes match the provided selector in Prometheus %s", namespace)
		}
	case ScrapeConfig:
		scrapeConfigs, err := clientSets.MClient.MonitoringV1alpha1().ScrapeConfigs(namespace).List(ctx, metav1.ListOptions{LabelSelector: labels.SelectorFromSet(labelMap).String()})
		if err != nil {
			return fmt.Errorf("failed to list ScrapeConfigs in %s: %v", namespace, err)
		}
		if len(scrapeConfigs.Items) == 0 {
			return fmt.Errorf("no ScrapeConfigs match the provided selector in Prometheus %s", namespace)
		}
	case PrometheusRule:
		promRules, err := clientSets.MClient.MonitoringV1().PrometheusRules(namespace).List(ctx, metav1.ListOptions{LabelSelector: labels.SelectorFromSet(labelMap).String()})
		if err != nil {
			return fmt.Errorf("failed to list Probes in %s: %v", namespace, err)
		}
		if len(promRules.Items) == 0 {
			return fmt.Errorf("no PrometheusRules match the provided selector in Prometheus %s", namespace)
		}
	default:
		return fmt.Errorf("unknown selector type: %s", resourceName)
	}

	return nil
}
