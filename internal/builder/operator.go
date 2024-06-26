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
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/applyconfiguration/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	applyCofongiAppsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	applyConfigCorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	applyConfigMetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	applyConfigRbacv1 "k8s.io/client-go/applyconfigurations/rbac/v1"
	"k8s.io/utils/ptr"
)

type OperatorBuilder struct {
	labels         map[string]string
	labelSelectors map[string]string
	namespace      string
	version        string
	manifets       OperatorManifests
}

type OperatorManifests struct {
	Deployment         *applyCofongiAppsv1.DeploymentApplyConfiguration
	Service            *applyConfigCorev1.ServiceApplyConfiguration
	ServiceAccount     *applyConfigCorev1.ServiceAccountApplyConfiguration
	ClusterRole        *applyConfigRbacv1.ClusterRoleApplyConfiguration
	ClusterRoleBinding *applyConfigRbacv1.ClusterRoleBindingApplyConfiguration
	ServiceMonitor     *monitoringv1.ServiceMonitorApplyConfiguration
}

func NewOperator(namespace, version string) *OperatorBuilder {
	return &OperatorBuilder{
		namespace: namespace,
		version:   version,
		labels: map[string]string{
			"app.kubernetes.io/component": "controller",
			"app.kubernetes.io/name":      "prometheus-operator",
			"app.kubernetes.io/version":   version,
		},
		labelSelectors: map[string]string{
			"app.kubernetes.io/component": "controller",
			"app.kubernetes.io/name":      "prometheus-operator",
		},
	}

}

func (o *OperatorBuilder) WithServiceAccount() *OperatorBuilder {
	o.manifets.ServiceAccount = &applyConfigCorev1.ServiceAccountApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("ServiceAccount"),
			APIVersion: ptr.To("v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("prometheus-operator"),
			Labels:    o.labels,
			Namespace: ptr.To(o.namespace),
		},
	}
	return o
}

func (o *OperatorBuilder) WithClusterRole() *OperatorBuilder {
	o.manifets.ClusterRole = &applyConfigRbacv1.ClusterRoleApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("ClusterRole"),
			APIVersion: ptr.To("rbac.authorization.k8s.io/v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:   ptr.To("prometheus-operator"),
			Labels: o.labels,
		},
		Rules: []applyConfigRbacv1.PolicyRuleApplyConfiguration{
			{
				APIGroups: []string{"monitoring.coreos.com"},
				Resources: []string{
					"alertmanagers",
					"alertmanagers/finalizers",
					"alertmanagers/status",
					"alertmanagerconfigs",
					"prometheuses",
					"prometheuses/finalizers",
					"prometheuses/status",
					"prometheusagents",
					"prometheusagents/finalizers",
					"prometheusagents/status",
					"thanosrulers",
					"thanosrulers/finalizers",
					"thanosrulers/status",
					"scrapeconfigs",
					"servicemonitors",
					"podmonitors",
					"probes",
					"prometheusrules",
				},
				Verbs: []string{"*"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"statefulsets"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps", "secrets"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"list", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"services", "services/finalizers", "endpoints"},
				Verbs:     []string{"get", "create", "update", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"namespaces"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"patch", "create"},
			},
			{
				APIGroups: []string{"networking.k8s.io"},
				Resources: []string{"ingresses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"storageclasses"},
				Verbs:     []string{"get"},
			},
		},
	}
	return o
}

func (o *OperatorBuilder) WithClusterRoleBinding() *OperatorBuilder {
	o.manifets.ClusterRoleBinding = &applyConfigRbacv1.ClusterRoleBindingApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("ClusterRoleBinding"),
			APIVersion: ptr.To("rbac.authorization.k8s.io/v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("prometheus-operator"),
			Labels:    o.labels,
			Namespace: ptr.To(o.namespace),
		},
		RoleRef: &applyConfigRbacv1.RoleRefApplyConfiguration{
			APIGroup: ptr.To("rbac.authorization.k8s.io"),
			Kind:     ptr.To("ClusterRole"),
			Name:     ptr.To("prometheus-operator"),
		},
		Subjects: []applyConfigRbacv1.SubjectApplyConfiguration{
			{
				Kind:      ptr.To("ServiceAccount"),
				Name:      o.manifets.ServiceAccount.Name,
				Namespace: ptr.To(o.namespace),
			},
		},
	}
	return o
}

func (o *OperatorBuilder) WithDeployment() *OperatorBuilder {
	o.manifets.Deployment = &applyCofongiAppsv1.DeploymentApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("Deployment"),
			APIVersion: ptr.To("apps/v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("prometheus-operator"),
			Labels:    o.labels,
			Namespace: ptr.To(o.namespace),
		},
		Spec: &applyCofongiAppsv1.DeploymentSpecApplyConfiguration{
			Replicas: ptr.To(int32(1)),
			Selector: &applyConfigMetav1.LabelSelectorApplyConfiguration{
				MatchLabels: o.labelSelectors,
			},
			Template: &applyConfigCorev1.PodTemplateSpecApplyConfiguration{
				ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
					Labels: o.labels,
					Annotations: map[string]string{
						"kubectl.kubernetes.io/default-container": "prometheus-operator",
					},
				},
				Spec: &applyConfigCorev1.PodSpecApplyConfiguration{
					AutomountServiceAccountToken: ptr.To(true),
					Containers: []applyConfigCorev1.ContainerApplyConfiguration{
						{
							Name:  ptr.To("prometheus-operator"),
							Image: ptr.To(fmt.Sprintf("quay.io/prometheus-operator/prometheus-operator:v%s", o.version)),
							Args: []string{
								"--kubelet-service=kube-system/kubelet",
								fmt.Sprintf("--prometheus-config-reloader=quay.io/prometheus-operator/prometheus-config-reloader:v%s", o.version),
							},
							Env: []applyConfigCorev1.EnvVarApplyConfiguration{
								{
									Name:  ptr.To("GOGC"),
									Value: ptr.To("30"),
								},
							},
							Ports: []applyConfigCorev1.ContainerPortApplyConfiguration{
								{
									Name:          ptr.To("http"),
									ContainerPort: ptr.To(int32(8080)),
								},
							},
							Resources: &applyConfigCorev1.ResourceRequirementsApplyConfiguration{
								Requests: &corev1.ResourceList{
									"cpu":    resource.MustParse("100m"),
									"memory": resource.MustParse("100Mi"),
								},
								Limits: &corev1.ResourceList{
									"cpu":    resource.MustParse("200m"),
									"memory": resource.MustParse("200Mi"),
								},
							},
							SecurityContext: &applyConfigCorev1.SecurityContextApplyConfiguration{
								ReadOnlyRootFilesystem:   ptr.To(true),
								AllowPrivilegeEscalation: ptr.To(false),
								Capabilities: &applyConfigCorev1.CapabilitiesApplyConfiguration{
									Drop: []corev1.Capability{"ALL"},
								},
							},
						},
					},
					NodeSelector: map[string]string{
						"kubernetes.io/os": "linux",
					},
					SecurityContext: &applyConfigCorev1.PodSecurityContextApplyConfiguration{
						RunAsNonRoot: ptr.To(true),
						RunAsUser:    ptr.To(int64(65534)),
						SeccompProfile: &applyConfigCorev1.SeccompProfileApplyConfiguration{
							Type: applyConfigCorev1.SeccompProfile().WithType("RuntimeDefault").Type,
						},
					},
					ServiceAccountName: o.manifets.ServiceAccount.Name,
				},
			},
		},
	}
	return o
}

func (o *OperatorBuilder) WithService() *OperatorBuilder {
	o.manifets.Service = &applyConfigCorev1.ServiceApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("Service"),
			APIVersion: ptr.To("v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("prometheus-operator"),
			Labels:    o.labels,
			Namespace: ptr.To(o.namespace),
		},
		Spec: &applyConfigCorev1.ServiceSpecApplyConfiguration{
			Ports: []applyConfigCorev1.ServicePortApplyConfiguration{
				{
					Name:        ptr.To("http"),
					Port:        ptr.To(int32(8080)),
					TargetPort:  ptr.To(intstr.FromString("http")),
					AppProtocol: ptr.To("http"),
				},
			},
			Selector: o.labelSelectors,
		},
	}
	return o
}

func (o *OperatorBuilder) WithServiceMonitor() *OperatorBuilder {
	o.manifets.ServiceMonitor = &monitoringv1.ServiceMonitorApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("ServiceMonitor"),
			APIVersion: ptr.To("monitoring.coreos.com/v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("prometheus-operator"),
			Labels:    o.labels,
			Namespace: ptr.To(o.namespace),
		},
		Spec: &monitoringv1.ServiceMonitorSpecApplyConfiguration{
			Selector: &metav1.LabelSelector{
				MatchLabels: o.labelSelectors,
			},
			Endpoints: []monitoringv1.EndpointApplyConfiguration{
				{
					HonorLabels: ptr.To(true),
					Port:        ptr.To("http"),
				},
			},
		},
	}
	return o
}

func (o *OperatorBuilder) Build() OperatorManifests {
	return o.manifets
}
