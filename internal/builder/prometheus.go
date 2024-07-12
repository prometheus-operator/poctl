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
	metav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	applyConfigRbacv1 "k8s.io/client-go/applyconfigurations/rbac/v1"
	"k8s.io/utils/ptr"
)

type PrometheusBuilder struct {
	labels         map[string]string
	labelSelectors map[string]string
	namespace      string
	manifests      PrometheusManifests
}

type PrometheusManifests struct {
	ServiceAccount     *applyConfigCorev1.ServiceAccountApplyConfiguration
	ClusterRole        *applyConfigRbacv1.ClusterRoleApplyConfiguration
	ClusterRoleBinding *applyConfigRbacv1.ClusterRoleBindingApplyConfiguration
	Prometheus         *monitoringv1.PrometheusApplyConfiguration
	Service            *applyConfigCorev1.ServiceApplyConfiguration
	ServiceMonitor     *monitoringv1.ServiceMonitorApplyConfiguration
}

func NewPrometheus(namespace string) *PrometheusBuilder {
	return &PrometheusBuilder{
		labels: map[string]string{
			"prometheus": "prometheus",
		},
		labelSelectors: map[string]string{
			"prometheus": "prometheus",
		},
		namespace: namespace,
	}
}

func (p *PrometheusBuilder) WithServiceAccount() *PrometheusBuilder {
	p.manifests.ServiceAccount = &applyConfigCorev1.ServiceAccountApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("ServiceAccount"),
			APIVersion: ptr.To("v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("prometheus"),
			Labels:    p.labels,
			Namespace: ptr.To(p.namespace),
		},
	}
	return p
}

func (p *PrometheusBuilder) WithClusterRole() *PrometheusBuilder {
	p.manifests.ClusterRole = &applyConfigRbacv1.ClusterRoleApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("ClusterRole"),
			APIVersion: ptr.To("rbac.authorization.k8s.io/v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("prometheus"),
			Labels:    p.labels,
			Namespace: ptr.To(p.namespace),
		},
		Rules: []applyConfigRbacv1.PolicyRuleApplyConfiguration{
			{
				APIGroups: []string{""},
				Resources: []string{"nodes", "nodes/metrics", "services", "endpoints", "pods"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups: []string{"networking.k8s.io"},
				Resources: []string{"ingresses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				NonResourceURLs: []string{"/metrics"},
				Verbs:           []string{"get"},
			},
		},
	}
	return p
}

func (p *PrometheusBuilder) WithClusterRoleBinding() *PrometheusBuilder {
	p.manifests.ClusterRoleBinding = &applyConfigRbacv1.ClusterRoleBindingApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("ClusterRoleBinding"),
			APIVersion: ptr.To("rbac.authorization.k8s.io/v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("prometheus"),
			Labels:    p.labels,
			Namespace: ptr.To(p.namespace),
		},
		RoleRef: &applyConfigRbacv1.RoleRefApplyConfiguration{
			APIGroup: ptr.To("rbac.authorization.k8s.io"),
			Kind:     ptr.To("ClusterRole"),
			Name:     p.manifests.ClusterRole.Name,
		},
		Subjects: []applyConfigRbacv1.SubjectApplyConfiguration{
			{
				Kind:      ptr.To("ServiceAccount"),
				Name:      p.manifests.ServiceAccount.Name,
				Namespace: ptr.To(p.namespace),
			},
		},
	}
	return p
}

func (p *PrometheusBuilder) WithPrometheus() *PrometheusBuilder {
	p.manifests.Prometheus = &monitoringv1.PrometheusApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("Prometheus"),
			APIVersion: ptr.To("monitoring.coreos.com/v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("prometheus"),
			Labels:    p.labels,
			Namespace: ptr.To(p.namespace),
		},
		Spec: &monitoringv1.PrometheusSpecApplyConfiguration{
			CommonPrometheusFieldsApplyConfiguration: monitoringv1.CommonPrometheusFieldsApplyConfiguration{
				ServiceAccountName:              p.manifests.ServiceAccount.Name,
				ServiceMonitorSelector:          &metav1.LabelSelectorApplyConfiguration{},
				ServiceMonitorNamespaceSelector: &metav1.LabelSelectorApplyConfiguration{},
				PodMonitorSelector:              &metav1.LabelSelectorApplyConfiguration{},
				PodMonitorNamespaceSelector:     &metav1.LabelSelectorApplyConfiguration{},
				ProbeSelector:                   &metav1.LabelSelectorApplyConfiguration{},
				ProbeNamespaceSelector:          &metav1.LabelSelectorApplyConfiguration{},
				ScrapeConfigSelector:            &metav1.LabelSelectorApplyConfiguration{},
				ScrapeConfigNamespaceSelector:   &metav1.LabelSelectorApplyConfiguration{},
				ImagePullPolicy:                 ptr.To(corev1.PullIfNotPresent),
				Replicas:                        ptr.To(int32(2)),
			},
			RuleSelector:          &metav1.LabelSelectorApplyConfiguration{},
			RuleNamespaceSelector: &metav1.LabelSelectorApplyConfiguration{},
			Alerting: &monitoringv1.AlertingSpecApplyConfiguration{
				Alertmanagers: []monitoringv1.AlertmanagerEndpointsApplyConfiguration{
					{
						Namespace: ptr.To(p.namespace),
						Name:      ptr.To(AlertManagerName),
						Port:      ptr.To(intstr.FromString("http-web")),
					},
				},
			},
		},
	}
	return p
}

func (p *PrometheusBuilder) WithService() *PrometheusBuilder {
	p.manifests.Service = &applyConfigCorev1.ServiceApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("Service"),
			APIVersion: ptr.To("v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("prometheus"),
			Labels:    p.labels,
			Namespace: ptr.To(p.namespace),
		},
		Spec: &applyConfigCorev1.ServiceSpecApplyConfiguration{
			Ports: []applyConfigCorev1.ServicePortApplyConfiguration{
				{
					Name:       ptr.To("web"),
					Port:       ptr.To(int32(9090)),
					TargetPort: ptr.To(intstr.FromString("web")),
				},
			},
			Selector: p.labelSelectors,
		},
	}
	return p
}

func (p *PrometheusBuilder) WithServiceMonitor() *PrometheusBuilder {
	p.manifests.ServiceMonitor = &monitoringv1.ServiceMonitorApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("ServiceMonitor"),
			APIVersion: ptr.To("monitoring.coreos.com/v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("prometheus"),
			Labels:    p.labels,
			Namespace: ptr.To(p.namespace),
		},
		Spec: &monitoringv1.ServiceMonitorSpecApplyConfiguration{
			Selector: &metav1.LabelSelectorApplyConfiguration{
				MatchLabels: p.labelSelectors,
			},
			Endpoints: []monitoringv1.EndpointApplyConfiguration{
				{
					HonorLabels: ptr.To(true),
					Port:        ptr.To("web"),
				},
			},
		},
	}
	return p
}

func (p *PrometheusBuilder) Build() PrometheusManifests {
	return p.manifests
}
