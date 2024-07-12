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

package builder

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/applyconfiguration/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	applyConfigCorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	applyConfigMetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/utils/ptr"
)

type AlertManagerBuilder struct {
	labels         map[string]string
	labelSelectors map[string]string
	namespace      string
	manifets       AlertManagerManifests
}

type AlertManagerManifests struct {
	ServiceAccount *applyConfigCorev1.ServiceAccountApplyConfiguration
	AlertManager   *monitoringv1.AlertmanagerApplyConfiguration
	Service        *applyConfigCorev1.ServiceApplyConfiguration
	ServiceMonitor *monitoringv1.ServiceMonitorApplyConfiguration
}

const AlertManagerName = "alertmanager"

func NewAlertManager(namespace string) *AlertManagerBuilder {
	return &AlertManagerBuilder{
		labels: map[string]string{
			"alertmanager": AlertManagerName,
		},
		labelSelectors: map[string]string{
			"alertmanager": AlertManagerName,
		},
		namespace: namespace,
	}
}

func (a *AlertManagerBuilder) WithServiceAccount() *AlertManagerBuilder {
	a.manifets.ServiceAccount = &applyConfigCorev1.ServiceAccountApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("ServiceAccount"),
			APIVersion: ptr.To("v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To(AlertManagerName),
			Labels:    a.labels,
			Namespace: ptr.To(a.namespace),
		},
	}
	return a
}

func (a *AlertManagerBuilder) WithAlertManager() *AlertManagerBuilder {
	a.manifets.AlertManager = &monitoringv1.AlertmanagerApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("Alertmanager"),
			APIVersion: ptr.To("monitoring.coreos.com/v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To(AlertManagerName),
			Labels:    a.labels,
			Namespace: ptr.To(a.namespace),
		},
		Spec: &monitoringv1.AlertmanagerSpecApplyConfiguration{
			ServiceAccountName:                  a.manifets.ServiceAccount.Name,
			Replicas:                            ptr.To(int32(1)),
			AlertmanagerConfigSelector:          &applyConfigMetav1.LabelSelectorApplyConfiguration{},
			AlertmanagerConfigNamespaceSelector: &applyConfigMetav1.LabelSelectorApplyConfiguration{},
		},
	}
	return a
}

func (a *AlertManagerBuilder) WithService() *AlertManagerBuilder {
	a.manifets.Service = &applyConfigCorev1.ServiceApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("Service"),
			APIVersion: ptr.To("v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To(AlertManagerName),
			Labels:    a.labels,
			Namespace: ptr.To(a.namespace),
		},
		Spec: &applyConfigCorev1.ServiceSpecApplyConfiguration{
			Ports: []applyConfigCorev1.ServicePortApplyConfiguration{
				{
					Name:       ptr.To("http-web"),
					Port:       ptr.To(int32(9093)),
					TargetPort: ptr.To(intstr.FromInt32(9093)),
					Protocol:   ptr.To(corev1.ProtocolTCP),
				},
				{
					Name:       ptr.To("reloader-web"),
					Port:       ptr.To(int32(8080)),
					TargetPort: ptr.To(intstr.FromString("reloader-web")),
				},
			},
			Selector: a.labelSelectors,
		},
	}
	return a
}

func (a *AlertManagerBuilder) WithServiceMonitor() *AlertManagerBuilder {
	a.manifets.ServiceMonitor = &monitoringv1.ServiceMonitorApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("ServiceMonitor"),
			APIVersion: ptr.To("monitoring.coreos.com/v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To(AlertManagerName),
			Labels:    a.labels,
			Namespace: ptr.To(a.namespace),
		},
		Spec: &monitoringv1.ServiceMonitorSpecApplyConfiguration{
			Selector: &applyConfigMetav1.LabelSelectorApplyConfiguration{
				MatchLabels: a.labelSelectors,
			},
			Endpoints: []monitoringv1.EndpointApplyConfiguration{
				{
					HonorLabels: ptr.To(true),
					Port:        ptr.To("http-web"),
				},
			},
		},
	}
	return a
}

func (a *AlertManagerBuilder) Build() AlertManagerManifests {
	return a.manifets
}
