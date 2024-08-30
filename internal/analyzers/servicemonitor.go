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

	"github.com/prometheus-operator/poctl/internal/k8sutil"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RunServiceMonitorAnalyzer(ctx context.Context, clientSets *k8sutil.ClientSets, name, namespace string) error {
	serviceMonitor, err := clientSets.MClient.MonitoringV1().ServiceMonitors(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("ServiceMonitor %s not found in namespace %s", name, namespace)
		}
		return fmt.Errorf("error while getting ServiceMonitor: %v", err)
	}

	if len(serviceMonitor.Spec.Selector.MatchLabels) == 0 && len(serviceMonitor.Spec.Selector.MatchExpressions) == 0 {
		return fmt.Errorf("ServiceMonitor %s in namespace %s does not have a selector", name, namespace)
	}

	if len(serviceMonitor.Spec.Selector.MatchLabels) > 0 {
		services, err := clientSets.KClient.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(&serviceMonitor.Spec.Selector),
		})

		if err != nil {
			return fmt.Errorf("error while listing services: %v", err)
		}

		if len(services.Items) == 0 {
			return fmt.Errorf("ServiceMonitor %s in namespace %s has no services matching the selector", name, namespace)
		}

		if err = evaluatePortMatches(serviceMonitor, services, name, namespace); err != nil {
			return err
		}
	}

	if len(serviceMonitor.Spec.Selector.MatchExpressions) > 0 {
		services, err := clientSets.KClient.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(&serviceMonitor.Spec.Selector),
		})

		if err != nil {
			return fmt.Errorf("error while listing services: %v", err)
		}

		if len(services.Items) == 0 {
			return fmt.Errorf("ServiceMonitor %s in namespace %s has no services matching the selector", name, namespace)
		}

		if err = evaluatePortMatches(serviceMonitor, services, name, namespace); err != nil {
			return err
		}
	}

	slog.Info("ServiceMonitor is compliant, no issues found", "name", name, "namespace", namespace)
	return nil
}

func evaluatePortMatches(serviceMonitor *monitoringv1.ServiceMonitor, services *v1.ServiceList, name string, namespace string) error {
	for _, endpoint := range serviceMonitor.Spec.Endpoints {
		found := false
		for _, service := range services.Items {
			for _, port := range service.Spec.Ports {
				if port.Name == endpoint.Port {
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			return fmt.Errorf("ServiceMonitor %s in namespace %s has no services with port %s", name, namespace, endpoint.Port)
		}
	}
	return nil
}
