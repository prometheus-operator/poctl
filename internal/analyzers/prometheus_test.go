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

package analyzers

import (
	"context"
	"testing"

	"github.com/prometheus-operator/poctl/internal/k8sutil"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/fake"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

func getPrometheusClusterRoleBinding(namespace string) []rbacv1.ClusterRoleBinding {
	return []rbacv1.ClusterRoleBinding{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "prometheus",
				Labels: map[string]string{
					"prometheus": "prometheus",
				},
			},
			RoleRef: rbacv1.RoleRef{
				Name: "prometheus",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "prometheus",
					Namespace: namespace,
				},
			},
		},
	}
}

func TestPrometheusAnalyzer(t *testing.T) {
	type testCase struct {
		name                string
		namespace           string
		getMockedClientSets func(tc testCase) k8sutil.ClientSets
		shouldFail          bool
	}

	tests := []testCase{
		{
			name:       "PrometheusRoleBindingListError",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.PrometheusList{})
				mClient.PrependReactor("get", "prometheuses", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Prometheus{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
					}, nil
				})

				kClient := fake.NewSimpleClientset(&rbacv1.ClusterRoleBindingList{})
				kClient.PrependReactor("list", "clusterrolebindings", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.NewInternalError(nil)
				})

				return k8sutil.ClientSets{
					MClient: mClient,
					KClient: kClient,
				}
			},
		},
		{
			name:       "PrometheusServiceAccountNotFound",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.PrometheusList{})
				mClient.PrependReactor("get", "prometheuses", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Prometheus{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
					}, nil
				})

				kClient := fake.NewSimpleClientset(&rbacv1.ClusterRoleBindingList{})
				kClient.PrependReactor("list", "clusterrolebindings", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &rbacv1.ClusterRoleBindingList{
						Items: getPrometheusClusterRoleBinding(tc.namespace),
					}, nil
				})

				return k8sutil.ClientSets{
					MClient: mClient,
					KClient: kClient,
				}
			},
		},
		{
			name:       "ConfigMapsVerbsNotFoundInClusterRole",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.PrometheusList{})
				mClient.PrependReactor("get", "prometheuses", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Prometheus{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
					}, nil
				})

				kClient := fake.NewSimpleClientset(&rbacv1.ClusterRoleBindingList{})
				kClient.PrependReactor("list", "clusterrolebindings", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &rbacv1.ClusterRoleBindingList{
						Items: getPrometheusClusterRoleBinding(tc.namespace),
					}, nil
				})

				kClient.PrependReactor("get", "clusterroles", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &rbacv1.ClusterRole{
						ObjectMeta: metav1.ObjectMeta{
							Name: "prometheus",
						},
						Rules: []rbacv1.PolicyRule{
							{
								Resources: []string{"configmaps"},
								Verbs:     []string{"list", "watch"},
							},
						},
					}, nil
				})

				return k8sutil.ClientSets{
					MClient: mClient,
					KClient: kClient,
				}
			},
		},
		{
			name:       "RequiredVerbsNotFoundInClusterRole",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.PrometheusList{})
				mClient.PrependReactor("get", "prometheuses", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Prometheus{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
					}, nil
				})

				kClient := fake.NewSimpleClientset(&rbacv1.ClusterRoleBindingList{})
				kClient.PrependReactor("list", "clusterrolebindings", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &rbacv1.ClusterRoleBindingList{
						Items: getPrometheusClusterRoleBinding(tc.namespace),
					}, nil
				})

				kClient.PrependReactor("get", "clusterroles", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &rbacv1.ClusterRole{
						ObjectMeta: metav1.ObjectMeta{
							Name: "prometheus",
						},
						Rules: []rbacv1.PolicyRule{
							{
								Resources: []string{"nodes", "pods"},
								Verbs:     []string{"list", "watch"},
								APIGroups: []string{""},
							},
						},
					}, nil
				})

				return k8sutil.ClientSets{
					MClient: mClient,
					KClient: kClient,
				}
			},
		},
		{
			name:       "NonResourceURLsNotFoundInClusterRole",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.PrometheusList{})
				mClient.PrependReactor("get", "prometheuses", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Prometheus{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
					}, nil
				})

				kClient := fake.NewSimpleClientset(&rbacv1.ClusterRoleBindingList{})
				kClient.PrependReactor("list", "clusterrolebindings", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &rbacv1.ClusterRoleBindingList{
						Items: getPrometheusClusterRoleBinding(tc.namespace),
					}, nil
				})

				kClient.PrependReactor("get", "clusterroles", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &rbacv1.ClusterRole{
						ObjectMeta: metav1.ObjectMeta{
							Name: "prometheus",
						},
						Rules: []rbacv1.PolicyRule{
							{
								NonResourceURLs: []string{"/metrics"},
								Verbs:           []string{"post"},
							},
						},
					}, nil
				})

				return k8sutil.ClientSets{
					MClient: mClient,
					KClient: kClient,
				}
			},
		},
		{
			name:       "NamespaceSelectorNull",
			namespace:  "test",
			shouldFail: false,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.PrometheusList{})
				mClient.PrependReactor("get", "prometheuses", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Prometheus{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
						Spec: monitoringv1.PrometheusSpec{
							CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
								ScrapeConfigNamespaceSelector: nil,
							},
						},
					}, nil
				})
				return k8sutil.ClientSets{
					MClient: mClient,
				}
			},
		},
		{
			name:       "NamespaceSelectorEmpty",
			namespace:  "test",
			shouldFail: false,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.PrometheusList{})
				mClient.PrependReactor("get", "prometheuses", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Prometheus{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
						Spec: monitoringv1.PrometheusSpec{
							CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
								PodMonitorNamespaceSelector: &metav1.LabelSelector{
									MatchLabels:      map[string]string{},
									MatchExpressions: []metav1.LabelSelectorRequirement{},
								},
							},
						},
					}, nil
				})
				return k8sutil.ClientSets{
					MClient: mClient,
				}
			},
		},
		{
			name:       "NamespaceSelectorWithoutMatchLabels",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.PrometheusList{})
				mClient.PrependReactor("get", "prometheuses", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Prometheus{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
						Spec: monitoringv1.PrometheusSpec{
							CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
								ServiceMonitorNamespaceSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{"environment": "test"},
								},
							},
						},
					}, nil
				})

				kClient := fake.NewSimpleClientset(&corev1.Namespace{})
				kClient.PrependReactor("list", "namespaces", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &corev1.NamespaceList{
						Items: []corev1.Namespace{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:   "test-namespace",
									Labels: map[string]string{"another": "label"},
								},
							},
						},
					}, nil
				})

				return k8sutil.ClientSets{
					MClient: mClient,
					KClient: kClient,
				}
			},
		},
		{
			name:       "NamespaceSelectorWithtMatchLabels",
			namespace:  "test",
			shouldFail: false,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.PrometheusList{})
				mClient.PrependReactor("get", "prometheuses", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Prometheus{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
						Spec: monitoringv1.PrometheusSpec{
							CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
								ProbeNamespaceSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{"environment": "test"},
								},
							},
						},
					}, nil
				})

				kClient := fake.NewSimpleClientset(&corev1.Namespace{})
				kClient.PrependReactor("list", "namespaces", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &corev1.NamespaceList{
						Items: []corev1.Namespace{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:   "test-namespace",
									Labels: map[string]string{"environment": "test"},
								},
							},
						},
					}, nil
				})

				return k8sutil.ClientSets{
					MClient: mClient,
					KClient: kClient,
				}
			},
		},
		{
			name:       "ServiceSelectorsEmpty",
			namespace:  "test",
			shouldFail: false,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.PrometheusList{})
				mClient.PrependReactor("get", "prometheuses", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Prometheus{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
						Spec: monitoringv1.PrometheusSpec{
							CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
								PodMonitorSelector: &metav1.LabelSelector{
									MatchLabels:      map[string]string{},
									MatchExpressions: []metav1.LabelSelectorRequirement{},
								},
							},
						},
					}, nil
				})

				return k8sutil.ClientSets{
					MClient: mClient,
				}
			},
		},
		{
			name:       "ServiceSelectorsNull",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.PrometheusList{})
				mClient.PrependReactor("get", "prometheuses", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Prometheus{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
						Spec: monitoringv1.PrometheusSpec{
							CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
								ServiceMonitorSelector: nil,
							},
						},
					}, nil
				})

				return k8sutil.ClientSets{
					MClient: mClient,
				}
			},
		},
		{
			name:       "ServiceSelectorsWithoutMatchLabels",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.PrometheusList{})
				mClient.PrependReactor("get", "prometheuses", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Prometheus{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
						Spec: monitoringv1.PrometheusSpec{
							CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
								ScrapeConfigSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{"app": "label"},
								},
							},
						},
					}, nil
				})

				mClient.PrependReactor("list", "scrapeconfigs", func(_ clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &monitoringv1alpha1.ScrapeConfigList{
						Items: []*monitoringv1alpha1.ScrapeConfig{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "scrapeconfig-crd",
									Namespace: tc.namespace,
									Labels:    map[string]string{"service": "notest"},
								},
							},
						},
					}, nil
				})

				return k8sutil.ClientSets{
					MClient: mClient,
				}
			},
		},
		{
			name:       "ServiceSelectorsWithMatchLabels",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.PrometheusList{})
				mClient.PrependReactor("get", "prometheuses", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Prometheus{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
						Spec: monitoringv1.PrometheusSpec{
							CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
								ProbeSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{"app": "label"},
								},
							},
						},
					}, nil
				})

				mClient.PrependReactor("list", "probes", func(_ clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &monitoringv1.ProbeList{
						Items: []*monitoringv1.Probe{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "probes-crd",
									Namespace: tc.namespace,
									Labels:    map[string]string{"app": "label"},
								},
							},
						},
					}, nil
				})

				return k8sutil.ClientSets{
					MClient: mClient,
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clientSets := tc.getMockedClientSets(tc)
			err := RunPrometheusAnalyzer(context.Background(), &clientSets, tc.name, tc.namespace)
			if tc.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
