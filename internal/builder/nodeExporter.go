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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	applyCofongiAppsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	applyConfigCorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	applyConfigMetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/utils/ptr"
)

const LatestNodeExporterVersion = "1.8.2"

type NodeExporterBuilder struct {
	labels         map[string]string
	labelSelectors map[string]string
	namespace      string
	manifests      NodexExporterManifests
	version        string
}

type NodexExporterManifests struct {
	ServiceAccount *applyConfigCorev1.ServiceAccountApplyConfiguration
	DaemonSet      *applyCofongiAppsv1.DaemonSetApplyConfiguration
	PodMonitor     *monitoringv1.PodMonitorApplyConfiguration
}

func NewNodeExporterBuilder(namespace, version string) *NodeExporterBuilder {
	return &NodeExporterBuilder{
		namespace: namespace,
		version:   version,
		labels: map[string]string{
			"app.kubernetes.io/name": "node-exporter",
		},
		labelSelectors: map[string]string{
			"app.kubernetes.io/name": "node-exporter",
		},
	}
}

func (n *NodeExporterBuilder) WithServiceAccount() *NodeExporterBuilder {
	n.manifests.ServiceAccount = &applyConfigCorev1.ServiceAccountApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("ServiceAccount"),
			APIVersion: ptr.To("v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("node-exporter"),
			Labels:    n.labels,
			Namespace: ptr.To(n.namespace),
		},
	}
	return n
}

func (n *NodeExporterBuilder) WithPodMonitor() *NodeExporterBuilder {
	n.manifests.PodMonitor = &monitoringv1.PodMonitorApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("PodMonitor"),
			APIVersion: ptr.To("monitoring.coreos.com/v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("node-exporter"),
			Labels:    n.labels,
			Namespace: ptr.To(n.namespace),
		},
		Spec: &monitoringv1.PodMonitorSpecApplyConfiguration{
			JobLabel: ptr.To("app.kubernetes.io/name"),
			Selector: &applyConfigMetav1.LabelSelectorApplyConfiguration{
				MatchLabels: n.labelSelectors,
			},
			NamespaceSelector: &monitoringv1.NamespaceSelectorApplyConfiguration{
				MatchNames: []string{n.namespace},
			},
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpointApplyConfiguration{
				{
					Port:          ptr.To("metrics"),
					HonorLabels:   ptr.To(true),
					FilterRunning: ptr.To(true),
				},
			},
		},
	}
	return n
}

var nodeExporterArgs = []string{
	"--web.listen-address=0.0.0.0:9100",
	"--path.sysfs=/host/sys",
	"--path.rootfs=/host/root",
	"--path.udev.data=/host/root/run/udev/data",
	"--no-collector.wifi",
	"--no-collector.hwmon",
	"--no-collector.btrfs",
	"--collector.filesystem.mount-points-exclude=^/(dev|proc|sys|run/k3s/containerd/.+|var/lib/docker/.+|var/lib/kubelet/pods/.+)($|/)",
	"--collector.netclass.ignored-devices=^(veth.*|[a-f0-9]{15})$",
	"--collector.netdev.device-exclude=^(veth.*|[a-f0-9]{15})$",
}

func (n *NodeExporterBuilder) WithDaemonSet() *NodeExporterBuilder {
	n.manifests.DaemonSet = &applyCofongiAppsv1.DaemonSetApplyConfiguration{
		TypeMetaApplyConfiguration: applyConfigMetav1.TypeMetaApplyConfiguration{
			Kind:       ptr.To("DaemonSet"),
			APIVersion: ptr.To("apps/v1"),
		},
		ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
			Name:      ptr.To("node-exporter"),
			Labels:    n.labels,
			Namespace: ptr.To(n.namespace),
		},
		Spec: &applyCofongiAppsv1.DaemonSetSpecApplyConfiguration{
			Selector: &applyConfigMetav1.LabelSelectorApplyConfiguration{
				MatchLabels: n.labelSelectors,
			},
			Template: &applyConfigCorev1.PodTemplateSpecApplyConfiguration{
				ObjectMetaApplyConfiguration: &applyConfigMetav1.ObjectMetaApplyConfiguration{
					Labels: n.labelSelectors,
				},
				Spec: &applyConfigCorev1.PodSpecApplyConfiguration{
					ServiceAccountName:           n.manifests.ServiceAccount.Name,
					AutomountServiceAccountToken: ptr.To(true),
					Containers: []applyConfigCorev1.ContainerApplyConfiguration{
						{
							Name:  ptr.To("node-exporter"),
							Image: ptr.To(fmt.Sprintf("quay.io/prometheus/node-exporter:v%s", n.version)),
							Args:  nodeExporterArgs,
							Ports: []applyConfigCorev1.ContainerPortApplyConfiguration{
								{
									Name:          ptr.To("metrics"),
									ContainerPort: ptr.To(int32(9100)),
									Protocol:      ptr.To(corev1.ProtocolTCP),
								},
							},
							Resources: &applyConfigCorev1.ResourceRequirementsApplyConfiguration{
								Requests: &corev1.ResourceList{
									"cpu":    resource.MustParse("200m"),
									"memory": resource.MustParse("200Mi"),
								},
								Limits: &corev1.ResourceList{
									"cpu":    resource.MustParse("200m"),
									"memory": resource.MustParse("200Mi"),
								},
							},
							SecurityContext: &applyConfigCorev1.SecurityContextApplyConfiguration{
								AllowPrivilegeEscalation: ptr.To(false),
								ReadOnlyRootFilesystem:   ptr.To(true),
								Capabilities: &applyConfigCorev1.CapabilitiesApplyConfiguration{
									Add: []corev1.Capability{
										"SYS_TIME",
									},
									Drop: []corev1.Capability{
										"ALL",
									},
								},
							},
							VolumeMounts: []applyConfigCorev1.VolumeMountApplyConfiguration{
								{
									MountPath:        ptr.To("/host/sys"),
									MountPropagation: ptr.To(corev1.MountPropagationHostToContainer),
									Name:             ptr.To("sys"),
									ReadOnly:         ptr.To(true),
								},
								{
									MountPath:        ptr.To("/host/root"),
									MountPropagation: ptr.To(corev1.MountPropagationHostToContainer),
									Name:             ptr.To("root"),
									ReadOnly:         ptr.To(true),
								},
							},
						},
					},
					HostNetwork: ptr.To(true),
					HostPID:     ptr.To(true),
					NodeSelector: map[string]string{
						"kubernetes.io/os": "linux",
					},
					PriorityClassName: ptr.To("system-node-critical"),
					SecurityContext: &applyConfigCorev1.PodSecurityContextApplyConfiguration{
						RunAsGroup:   ptr.To(int64(65534)),
						RunAsNonRoot: ptr.To(true),
						RunAsUser:    ptr.To(int64(65534)),
					},
					Tolerations: []applyConfigCorev1.TolerationApplyConfiguration{
						{
							Operator: ptr.To(corev1.TolerationOpExists),
						},
					},
					Volumes: []applyConfigCorev1.VolumeApplyConfiguration{
						{
							Name: ptr.To("sys"),
							VolumeSourceApplyConfiguration: applyConfigCorev1.VolumeSourceApplyConfiguration{
								HostPath: &applyConfigCorev1.HostPathVolumeSourceApplyConfiguration{
									Path: ptr.To("/sys"),
								},
							},
						},
						{
							Name: ptr.To("root"),
							VolumeSourceApplyConfiguration: applyConfigCorev1.VolumeSourceApplyConfiguration{
								HostPath: &applyConfigCorev1.HostPathVolumeSourceApplyConfiguration{
									Path: ptr.To("/"),
								},
							},
						},
					},
				},
			},
			UpdateStrategy: &applyCofongiAppsv1.DaemonSetUpdateStrategyApplyConfiguration{
				Type: ptr.To(appsv1.RollingUpdateDaemonSetStrategyType),
				RollingUpdate: &applyCofongiAppsv1.RollingUpdateDaemonSetApplyConfiguration{
					MaxUnavailable: ptr.To(intstr.FromString("10%")),
				},
			},
		},
	}
	return n
}

func (n *NodeExporterBuilder) Build() NodexExporterManifests {
	return n.manifests
}
