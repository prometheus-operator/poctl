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
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/fake"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/kubernetes/fake"
	rbacv1 "k8s.io/api/rbac/v1"
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
			name:      "PrometheusRoleBindingListError",
			namespace: "test",
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
			name:      "PrometheusServiceAccountNotFound",
			namespace: "test",
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
			name:      "ConfigMapsNotFoundInClusterRole",
			namespace: "test",
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
			name:      "RequiredVerbsNotFoundInClusterRole",
			namespace: "test",
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
                                Resources: []string{"configmaps", "pods"},
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
			name:      "NonResourceURLsNotFoundInClusterRole",
			namespace: "test",
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
                                Verbs:     []string{"post"},
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
			name:      "PrometheusMatchesAllNamespaces",
			namespace: "test",
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
								ServiceMonitorNamespaceSelector: &metav1.LabelSelector{
									MatchLabels:      map[string]string{},
									MatchExpressions: []metav1.LabelSelectorRequirement{},
								},
								ProbeNamespaceSelector: &metav1.LabelSelector{
									MatchLabels:      map[string]string{},
									MatchExpressions: []metav1.LabelSelectorRequirement{}, 
								},
								ScrapeConfigNamespaceSelector: &metav1.LabelSelector{
									MatchLabels:      map[string]string{},
									MatchExpressions: []metav1.LabelSelectorRequirement{},
								},
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
			name:      "PrometheusCurrentNamespaceNoSelectors",
			namespace: "test",
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
								ServiceMonitorNamespaceSelector: nil,
								ProbeNamespaceSelector:          nil,
								ScrapeConfigNamespaceSelector:   nil,
								PodMonitorNamespaceSelector:     nil,
							},
						},
					}, nil
				})

				mClient.PrependReactor("list", "podmonitors", func(_ clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &monitoringv1.PodMonitorList{Items: []*monitoringv1.PodMonitor{}}, nil
				})

				mClient.PrependReactor("list", "servicemonitors", func(_ clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &monitoringv1.ServiceMonitorList{Items: []*monitoringv1.ServiceMonitor{}}, nil
				})

				mClient.PrependReactor("list", "probes", func(_ clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &monitoringv1.ProbeList{Items: []*monitoringv1.Probe{}}, nil
				})

				mClient.PrependReactor("list", "scrapeconfigs", func(_ clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &monitoringv1alpha1.ScrapeConfigList{Items: []*monitoringv1alpha1.ScrapeConfig{}}, nil
				})

				return k8sutil.ClientSets{
					MClient: mClient,
				}
			},
		},
		{
			name:      "PrometheusNoServiceSelectors",
			namespace: "test",
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
								ProbeSelector:          nil,
								ScrapeConfigSelector:   nil,
								PodMonitorSelector:     nil,
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
			name:      "PrometheusNoServiceSelectorsCreated",
			namespace: "test",
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
								ServiceMonitorSelector: &metav1.LabelSelector{
									MatchLabels:      map[string]string{},
									MatchExpressions: []metav1.LabelSelectorRequirement{},
								},
								ProbeSelector: &metav1.LabelSelector{
									MatchLabels:      map[string]string{},
									MatchExpressions: []metav1.LabelSelectorRequirement{}, 
								},
								ScrapeConfigSelector: &metav1.LabelSelector{
									MatchLabels:      map[string]string{},
									MatchExpressions: []metav1.LabelSelectorRequirement{},
								},
								PodMonitorSelector: &metav1.LabelSelector{
									MatchLabels:      map[string]string{},
									MatchExpressions: []metav1.LabelSelectorRequirement{},
								},
							},
						},
					}, nil
				})

				mClient.PrependReactor("list", "podmonitors", func(_ clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &monitoringv1.PodMonitorList{Items: []*monitoringv1.PodMonitor{}}, nil
				})

				mClient.PrependReactor("list", "servicemonitors", func(_ clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &monitoringv1.ServiceMonitorList{Items: []*monitoringv1.ServiceMonitor{}}, nil
				})

				mClient.PrependReactor("list", "probes", func(_ clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &monitoringv1.ProbeList{Items: []*monitoringv1.Probe{}}, nil
				})

				mClient.PrependReactor("list", "scrapeconfigs", func(_ clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &monitoringv1alpha1.ScrapeConfigList{Items: []*monitoringv1alpha1.ScrapeConfig{}}, nil
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
