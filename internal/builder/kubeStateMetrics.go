// Copyright 2024 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package builder

import (
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/applyconfiguration/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	applyCofongiAppsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	applyConfigCorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	applyConfigMetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	applyConfigRbacv1 "k8s.io/client-go/applyconfigurations/rbac/v1"
	"k8s.io/utils/ptr"
)

const LatestKubeStateMetricsVersion = "2.12.0"

type KubeStateMetricsBuilder struct {
	labels         map[string]string
	labelSelectors map[string]string
	namespace      string
	manifests      KubeStateMetricsManifests
	version        string
}

type KubeStateMetricsManifests struct {
	Deployment         *applyCofongiAppsv1.DeploymentApplyConfiguration
	Service            *applyConfigCorev1.ServiceApplyConfiguration
	ServiceAccount     *applyConfigCorev1.ServiceAccountApplyConfiguration
	ClusterRole        *applyConfigRbacv1.ClusterRoleApplyConfiguration
	ClusterRoleBinding *applyConfigRbacv1.ClusterRoleBindingApplyConfiguration
	ServiceMonitor     *monitoringv1.ServiceMonitorApplyConfiguration
}

func NewKubeStateMetricsBuilder(namespace, version string) *KubeStateMetricsBuilder {
	return &KubeStateMetricsBuilder{
		labels: map[string]string{
			"app.kubernetes.io/name": "kube-state-metrics",
		},
		labelSelectors: map[string]string{
			"app.kubernetes.io/name": "kube-state-metrics",
		},
		namespace: namespace,
		version:   LatestKubeStateMetricsVersion,
	}
}

func (k *KubeStateMetricsBuilder) WithServiceAccount() *KubeStateMetricsBuilder {
	k.manifests.ServiceAccount = &applyConfigCorev1.ServiceAccountApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("ServiceAccount"),
			APIVersion: ptr.To("v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("kube-state-metrics"),
			Labels:    k.labels,
			Namespace: ptr.To(k.namespace),
		},
	}
	return k
}

func (k *KubeStateMetricsBuilder) WithClusterRole() *KubeStateMetricsBuilder {
	k.manifests.ClusterRole = &applyConfigRbacv1.ClusterRoleApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("ClusterRole"),
			APIVersion: ptr.To("rbac.authorization.k8s.io/v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("kube-state-metrics"),
			Labels:    k.labels,
			Namespace: ptr.To(k.namespace),
		},
		Rules: []applyConfigRbacv1.PolicyRuleApplyConfiguration{
			{
				APIGroups: []string{""},
				Resources: []string{
					"configmaps",
					"secrets",
					"nodes",
					"pods",
					"services",
					"serviceaccounts",
					"resourcequotas",
					"replicationcontrollers",
					"limitranges",
					"persistentvolumeclaims",
					"persistentvolumes",
					"namespaces",
					"endpoints",
				},
				Verbs: []string{"list", "watch"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{
					"statefulsets",
					"daemonsets",
					"deployments",
					"replicasets",
				},
				Verbs: []string{"list", "watch"},
			},
			{
				APIGroups: []string{"batch"},
				Resources: []string{"cronjobs", "jobs"},
				Verbs:     []string{"list", "watch"},
			},
			{
				APIGroups: []string{"autoscaling"},
				Resources: []string{"horizontalpodautoscalers"},
				Verbs:     []string{"list", "watch"},
			},
			{
				APIGroups: []string{"authentication.k8s.io"},
				Resources: []string{"tokenreviews"},
				Verbs:     []string{"create"},
			},
			{
				APIGroups: []string{"authorization.k8s.io"},
				Resources: []string{"subjectaccessreviews"},
				Verbs:     []string{"create"},
			},
			{
				APIGroups: []string{"policy"},
				Resources: []string{"poddisruptionbudgets"},
				Verbs:     []string{"list", "watch"},
			},
			{
				APIGroups: []string{"certificates.k8s.io"},
				Resources: []string{"certificatesigningrequests"},
				Verbs:     []string{"list", "watch"},
			},
			{
				APIGroups: []string{"discovery.k8s.io"},
				Resources: []string{"endpointslices"},
				Verbs:     []string{"list", "watch"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"storageclasses", "volumeattachments"},
				Verbs:     []string{"list", "watch"},
			},
			{
				APIGroups: []string{"admissionregistration.k8s.io"},
				Resources: []string{"mutatingwebhookconfigurations", "validatingwebhookconfigurations"},
				Verbs:     []string{"list", "watch"},
			},
			{
				APIGroups: []string{"networking.k8s.io"},
				Resources: []string{"networkpolicies", "ingressclasses", "ingresses"},
				Verbs:     []string{"list", "watch"},
			},
			{
				APIGroups: []string{"coordination.k8s.io"},
				Resources: []string{"leases"},
				Verbs:     []string{"list", "watch"},
			},
			{
				APIGroups: []string{"rbac.authorization.k8s.io"},
				Resources: []string{"clusterrolebindings", "clusterroles", "rolebindings", "roles"},
				Verbs:     []string{"list", "watch"},
			},
		},
	}
	return k
}

func (k *KubeStateMetricsBuilder) WithClusterRoleBinding() *KubeStateMetricsBuilder {
	k.manifests.ClusterRoleBinding = &applyConfigRbacv1.ClusterRoleBindingApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("ClusterRoleBinding"),
			APIVersion: ptr.To("rbac.authorization.k8s.io/v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("kube-state-metrics"),
			Labels:    k.labels,
			Namespace: ptr.To(k.namespace),
		},
		RoleRef: &applyConfigRbacv1.RoleRefApplyConfiguration{
			APIGroup: ptr.To("rbac.authorization.k8s.io"),
			Kind:     ptr.To("ClusterRole"),
			Name:     ptr.To("kube-state-metrics"),
		},
		Subjects: []applyConfigRbacv1.SubjectApplyConfiguration{
			{
				Kind:      ptr.To("ServiceAccount"),
				Name:      k.manifests.ServiceAccount.Name,
				Namespace: ptr.To(k.namespace),
			},
		},
	}
	return k
}

func (k *KubeStateMetricsBuilder) WithService() *KubeStateMetricsBuilder {
	k.manifests.Service = &applyConfigCorev1.ServiceApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("Service"),
			APIVersion: ptr.To("v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("kube-state-metrics"),
			Labels:    k.labels,
			Namespace: ptr.To(k.namespace),
		},
		Spec: &applyConfigCorev1.ServiceSpecApplyConfiguration{
			Ports: []applyConfigCorev1.ServicePortApplyConfiguration{
				{
					Name:       ptr.To("http"),
					Port:       ptr.To(int32(8080)),
					TargetPort: ptr.To(intstr.FromString("http")),
				},
				{
					Name:       ptr.To("metrics"),
					Port:       ptr.To(int32(8081)),
					TargetPort: ptr.To(intstr.FromString("metrics")),
				},
			},
			Selector: k.labelSelectors,
		},
	}
	return k
}

func (k *KubeStateMetricsBuilder) WithDeployment() *KubeStateMetricsBuilder {
	k.manifests.Deployment = &applyCofongiAppsv1.DeploymentApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("Deployment"),
			APIVersion: ptr.To("apps/v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("kube-state-metrics"),
			Labels:    k.labels,
			Namespace: ptr.To(k.namespace),
		},
		Spec: &applyCofongiAppsv1.DeploymentSpecApplyConfiguration{
			Replicas: ptr.To(int32(1)),
			Selector: &applyConfigMetav1.LabelSelectorApplyConfiguration{
				MatchLabels: k.labelSelectors,
			},
			Template: &applyConfigCorev1.PodTemplateSpecApplyConfiguration{
				ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
					Labels: k.labelSelectors,
				},
				Spec: &applyConfigCorev1.PodSpecApplyConfiguration{
					ServiceAccountName: k.manifests.ServiceAccount.Name,
					Containers: []applyConfigCorev1.ContainerApplyConfiguration{
						{
							Name:  ptr.To("kube-state-metrics"),
							Image: ptr.To(fmt.Sprintf("registry.k8s.io/kube-state-metrics/kube-state-metrics:v%v", k.version)),
							Args: []string{
								"--port=8080",
							},
							Ports: []applyConfigCorev1.ContainerPortApplyConfiguration{
								{
									Name:          ptr.To("http"),
									ContainerPort: ptr.To(int32(8080)),
								},
								{
									Name:          ptr.To("metrics"),
									ContainerPort: ptr.To(int32(8081)),
								},
							},
							SecurityContext: &applyConfigCorev1.SecurityContextApplyConfiguration{
								AllowPrivilegeEscalation: ptr.To(false),
								Capabilities: &applyConfigCorev1.CapabilitiesApplyConfiguration{
									Drop: []corev1.Capability{
										"ALL",
									},
								},
								ReadOnlyRootFilesystem: ptr.To(true),
								RunAsUser:              ptr.To(int64(65534)),
								RunAsNonRoot:           ptr.To(true),
								RunAsGroup:             ptr.To(int64(65534)),
								SeccompProfile: &applyConfigCorev1.SeccompProfileApplyConfiguration{
									Type: ptr.To(corev1.SeccompProfileTypeRuntimeDefault),
								},
							},
						},
					},
				},
			},
		},
	}
	return k
}

func (k *KubeStateMetricsBuilder) WithServiceMonitor() *KubeStateMetricsBuilder {
	k.manifests.ServiceMonitor = &monitoringv1.ServiceMonitorApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("ServiceMonitor"),
			APIVersion: ptr.To("monitoring.coreos.com/v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("kube-state-metrics"),
			Labels:    k.labels,
			Namespace: ptr.To(k.namespace),
		},
		Spec: &monitoringv1.ServiceMonitorSpecApplyConfiguration{
			Selector: &applyConfigMetav1.LabelSelectorApplyConfiguration{
				MatchLabels: k.labelSelectors,
			},
			Endpoints: []monitoringv1.EndpointApplyConfiguration{
				{
					HonorLabels: ptr.To(true),
					Port:        ptr.To("http"),
				},
				{
					HonorLabels: ptr.To(true),
					Port:        ptr.To("metrics"),
				},
			},
		},
	}
	return k
}

func (k *KubeStateMetricsBuilder) Build() KubeStateMetricsManifests {
	return k.manifests
}
