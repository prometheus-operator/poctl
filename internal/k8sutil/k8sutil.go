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

package k8sutil

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	v1 "k8s.io/api/rbac/v1"
	apiv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiExtensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
)

const (
	ServiceMonitor = "ServiceMonitor"
	PodMonitor     = "PodMonitor"
	Probe          = "Probe"
	ScrapeConfig   = "ScrapeConfig"
	PrometheusRule = "PrometheusRule"
)

var ApplyOption = metav1.ApplyOptions{
	FieldManager: "application/apply-patch",
}

func getKubeConfig() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	kubeConfig := filepath.Clean(fmt.Sprintf("%v/.kube/config", usr.HomeDir))

	if _, err := os.Stat(kubeConfig); err != nil {
		return "", err
	}

	return kubeConfig, nil
}

func GetRestConfig(kubeConfig string) (*rest.Config, error) {
	var config *rest.Config
	var err error

	if kubeConfig == "" {
		kubeConfig, err = getKubeConfig()
		if err != nil {
			return nil, fmt.Errorf("error while getting kubeconfig: %v", err)
		}
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error while creating k8s client config: %v", err)
	}

	return config, nil
}

func CrdDeserilezer(logger *slog.Logger, reader io.ReadCloser) (runtime.Object, error) {
	sch := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(sch)
	_ = apiextv1beta1.AddToScheme(sch)
	_ = apiv1.AddToScheme(sch)

	_ = monitoringv1.AddToScheme(sch)
	_ = monitoringv1alpha1.AddToScheme(sch)

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		logger.Error("error while reading CRD", "error", err)
		return &runtime.Unknown{}, err
	}

	decode := serializer.NewCodecFactory(sch).UniversalDeserializer().Decode

	obj, _, err := decode(buf.Bytes(), nil, nil)
	if err != nil {
		logger.Error("error while decoding CRD", "error", err)
		return &runtime.Unknown{}, err
	}

	return obj, nil
}

type ClientSets struct {
	KClient             kubernetes.Interface
	MClient             monitoringclient.Interface
	DClient             dynamic.Interface
	APIExtensionsClient apiExtensions.Interface
}

func GetClientSets(kubeconfig string) (*ClientSets, error) {
	restConfig, err := GetRestConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error while getting k8s client config: %v", err)

	}

	kclient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error while creating k8s client: %v", err)
	}

	mclient, err := monitoringclient.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error while creating Prometheus Operator client: %v", err)
	}

	kdynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error while creating dynamic client: %v", err)
	}

	apiExtensions, err := apiExtensions.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error while creating apiextensions client: %v", err)
	}

	return &ClientSets{
		KClient:             kclient,
		MClient:             mclient,
		DClient:             kdynamicClient,
		APIExtensionsClient: apiExtensions,
	}, nil
}

func IsServiceAccountBoundToRoleBindingList(clusterRoleBindings *v1.ClusterRoleBindingList, serviceAccountName string) bool {
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

func CheckResourceNamespaceSelectors(ctx context.Context, clientSets ClientSets, labelSelector *metav1.LabelSelector) error {
	if labelSelector == nil {
		return nil
	}

	if len(labelSelector.MatchLabels) == 0 && len(labelSelector.MatchExpressions) == 0 {
		return nil
	}

	labelMap, err := metav1.LabelSelectorAsMap(labelSelector)
	if err != nil {
		return fmt.Errorf("invalid label selector format in %s: %v", labelSelector, err)
	}

	namespaces, err := clientSets.KClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{LabelSelector: labels.SelectorFromSet(labelMap).String()})

	if err != nil {
		return fmt.Errorf("failed to list Namespaces in %s: %v", labelSelector, err)
	}

	if len(namespaces.Items) == 0 {
		return fmt.Errorf("no namespaces match the selector %s", labelSelector)
	}
	return nil
}

func CheckResourceLabelSelectors(ctx context.Context, clientSets ClientSets, labelSelector *metav1.LabelSelector, resourceName, namespace string) error {
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

func CheckPrometheusClusterRoleRules(crb v1.ClusterRoleBinding, cr *v1.ClusterRole) error {
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
