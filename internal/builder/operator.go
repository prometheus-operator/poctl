package builder

import (
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	Deployment         *appsv1.Deployment
	Service            *apiv1.Service
	ServiceAccount     *apiv1.ServiceAccount
	ClusterRole        *rbac.ClusterRole
	ClusterRoleBinding *rbac.ClusterRoleBinding
	ServiceMonitor     *monitoringv1.ServiceMonitor
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
	o.manifets.ServiceAccount = &apiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus-operator",
			Labels:    o.labels,
			Namespace: o.namespace,
		},
	}
	return o
}

func (o *OperatorBuilder) WithClusterRole() *OperatorBuilder {
	o.manifets.ClusterRole = &rbac.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus-operator",
			Labels:    o.labels,
			Namespace: o.namespace,
		},
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string{"monitoring.coreos.com"},
				Resources: []string{
					"alertmanagers", "alertmanagers/finalizers", "alertmanagers/status",
					"alertmanagerconfigs", "prometheuses", "prometheuses/finalizers",
					"prometheuses/status", "prometheusagents", "prometheusagents/finalizers",
					"prometheusagents/status", "thanosrulers", "thanosrulers/finalizers",
					"thanosrulers/status", "scrapeconfigs", "servicemonitors", "podmonitors",
					"probes", "prometheusrules",
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
	o.manifets.ClusterRoleBinding = &rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus-operator",
			Labels:    o.labels,
			Namespace: o.namespace,
		},
		RoleRef: rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "prometheus-operator",
		},
		Subjects: []rbac.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      o.manifets.ServiceAccount.Name,
				Namespace: o.namespace,
			},
		},
	}
	return o
}

func (o *OperatorBuilder) WithDeployment() *OperatorBuilder {
	o.manifets.Deployment = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus-operator",
			Labels:    o.labels,
			Namespace: o.namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(1)),
			Selector: &metav1.LabelSelector{
				MatchLabels: o.labelSelectors,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: o.labels,
					Annotations: map[string]string{
						"kubectl.kubernetes.io/default-container": "prometheus-operator",
					},
				},
				Spec: apiv1.PodSpec{
					AutomountServiceAccountToken: ptr.To(true),
					Containers: []apiv1.Container{
						{
							Name:  "prometheus-operator",
							Image: fmt.Sprintf("quay.io/prometheus-operator/prometheus-operator:v%s", o.version),
							Args: []string{
								"--kubelet-service=kube-system/kubelet",
								fmt.Sprintf("--prometheus-config-reloader=quay.io/prometheus-operator/prometheus-config-reloader:v%s", o.version),
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "GOGC",
									Value: "30",
								},
							},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8080,
								},
							},
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("100m"),
									apiv1.ResourceMemory: resource.MustParse("100Mi"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("200m"),
									apiv1.ResourceMemory: resource.MustParse("200Mi"),
								},
							},
							SecurityContext: &apiv1.SecurityContext{
								ReadOnlyRootFilesystem:   ptr.To(true),
								AllowPrivilegeEscalation: ptr.To(false),
								Capabilities: &apiv1.Capabilities{
									Drop: []apiv1.Capability{
										"ALL",
									},
								},
							},
						},
					},
					NodeSelector: map[string]string{
						"kubernetes.io/os": "linux",
					},
					SecurityContext: &apiv1.PodSecurityContext{
						RunAsNonRoot: ptr.To(true),
						RunAsUser:    ptr.To(int64(65534)),
						SeccompProfile: &apiv1.SeccompProfile{
							Type: "RuntimeDefault",
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
	o.manifets.Service = &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus-operator",
			Labels:    o.labels,
			Namespace: o.namespace,
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name:        "http",
					Port:        8080,
					TargetPort:  intstr.FromString("http"),
					AppProtocol: ptr.To("http"),
				},
			},
			Selector: o.labelSelectors,
		},
	}
	return o
}

func (o *OperatorBuilder) WithServiceMonitor() *OperatorBuilder {
	o.manifets.ServiceMonitor = &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus-operator",
			Labels:    o.labels,
			Namespace: o.namespace,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: o.labelSelectors,
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					HonorLabels: true,
					Port:        "http",
				},
			},
		},
	}
	return o
}

func (o *OperatorBuilder) Build() OperatorManifests {
	return o.manifets
}
