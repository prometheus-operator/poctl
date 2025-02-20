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
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RunOverlappingAnalyzer(ctx context.Context, clientSets *k8sutil.ClientSets, _, namespace string) error {
	serviceMonitors, err := clientSets.MClient.MonitoringV1().ServiceMonitors(namespace).List(ctx, metav1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	podMonitors, err := clientSets.MClient.MonitoringV1().PodMonitors(namespace).List(ctx, metav1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if (serviceMonitors == nil || len(serviceMonitors.Items) == 0) && (podMonitors == nil || len(podMonitors.Items) == 0) {
		return nil
	}

	var overlapErrs []string

	for _, servicemonitor := range serviceMonitors.Items {
		if err := checkOverlappingServiceMonitors(ctx, clientSets, servicemonitor); err != nil {
			overlapErrs = append(overlapErrs, err.Error())
		}
	}
	for _, podmonitor := range podMonitors.Items {
		if err := checkOverlappingPodMonitors(ctx, clientSets, podmonitor); err != nil {
			overlapErrs = append(overlapErrs, err.Error())
		}
	}

	if len(overlapErrs) > 0 {
		return fmt.Errorf("multiple issues found:\n%s", strings.Join(overlapErrs, "\n"))
	}

	slog.Info("no overlapping monitoring configurations found in", "namespace", namespace)
	return nil
}

func checkOverlappingServiceMonitors(ctx context.Context, clientSets *k8sutil.ClientSets, servicemonitor *monitoringv1.ServiceMonitor) error {
	selector, err := metav1.LabelSelectorAsSelector(&servicemonitor.Spec.Selector)
	if err != nil {
		return fmt.Errorf("invalid selector in ServiceMonitor %s/%s: %v", servicemonitor.Namespace, servicemonitor.Name, err)
	}

	services, err := clientSets.KClient.CoreV1().Services(servicemonitor.Namespace).List(ctx, metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return fmt.Errorf("error listing services for ServiceMonitor %s/%s: %v", servicemonitor.Namespace, servicemonitor.Name, err)
	}

	serviceOverlaps := make(map[string][]string)
	var overlapErrs []string
	for _, service := range services.Items {
		for _, svcPort := range service.Spec.Ports {
			serviceKey := fmt.Sprintf("%s/%s:%d", service.Namespace, service.Name, svcPort.Port)
			serviceOverlaps[serviceKey] = append(serviceOverlaps[serviceKey], servicemonitor.Name)

			if len(serviceOverlaps[serviceKey]) > 1 {
				overlapErrs = append(overlapErrs, fmt.Sprintf("Overlapping ServiceMonitors found for service/port %s: %v", serviceKey, serviceOverlaps[serviceKey]))
			}
		}
	}

	if len(overlapErrs) > 0 {
		return fmt.Errorf("%s", strings.Join(overlapErrs, "\n"))
	}

	return nil
}

func checkOverlappingPodMonitors(ctx context.Context, clientSets *k8sutil.ClientSets, podmonitor *monitoringv1.PodMonitor) error {
	selector, err := metav1.LabelSelectorAsSelector(&podmonitor.Spec.Selector)
	if err != nil {
		return fmt.Errorf("invalid selector in PodMonitor %s/%s: %v", podmonitor.Namespace, podmonitor.Name, err)
	}

	pods, err := clientSets.KClient.CoreV1().Pods(podmonitor.Namespace).List(ctx, metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return fmt.Errorf("error listing pods for PodMonitor %s/%s: %v", podmonitor.Namespace, podmonitor.Name, err)
	}

	podOverlaps := make(map[string][]string)
	var overlapErrs []string
	for _, pod := range pods.Items {
		for _, podPort := range podmonitor.Spec.PodMetricsEndpoints {
			podKey := fmt.Sprintf("%s/%s:%s", pod.Namespace, pod.Name, podPort.Port)
			podOverlaps[podKey] = append(podOverlaps[podKey], podmonitor.Name)

			if len(podOverlaps[podKey]) > 1 {
				overlapErrs = append(overlapErrs, fmt.Sprintf("Overlapping ServiceMonitors found for service/port %s: %v", podKey, podOverlaps[podKey]))
			}
		}
	}

	if len(overlapErrs) > 0 {
		return fmt.Errorf("%s", strings.Join(overlapErrs, "\n"))
	}

	return nil
}
