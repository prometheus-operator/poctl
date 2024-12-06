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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

func TestAlertmanagerAnalyzer(t *testing.T) {
	type testCase struct {
		name                string
		namespace           string
		getMockedClientSets func(tc testCase) k8sutil.ClientSets
		shouldFail          bool
	}

	tests := []testCase{
		{
			name:       "AlertmanagerNotFound",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.AlertmanagerList{})
				mClient.PrependReactor("get", "alertmanagers", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.NewNotFound(monitoringv1.Resource("alertmanagers"), tc.name)
				})

				kClient := fake.NewSimpleClientset()
				return k8sutil.ClientSets{
					MClient: mClient,
					KClient: kClient,
				}
			},
		},
		{
			name:       "AlertmanagerMissingServiceAccount",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.AlertmanagerList{})
				mClient.PrependReactor("get", "alertmanagers", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Alertmanager{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
						Spec: monitoringv1.AlertmanagerSpec{
							ServiceAccountName: "test-sa",
						},
					}, nil
				})

				kClient := fake.NewSimpleClientset(&corev1.ServiceAccount{})
				kClient.PrependReactor("get", "serviceaccount", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.NewInternalError(nil)
				})
				return k8sutil.ClientSets{
					MClient: mClient,
					KClient: kClient,
				}
			},
		},
		{
			name:       "AlertmanagerFailToGetConfigSecret",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.AlertmanagerList{})
				mClient.PrependReactor("get", "alertmanagers", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Alertmanager{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
						Spec: monitoringv1.AlertmanagerSpec{
							ConfigSecret: "test-secret",
						},
					}, nil
				})

				kClient := fake.NewSimpleClientset(&corev1.Secret{})
				kClient.PrependReactor("get", "secret", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.NewInternalError(nil)
				})
				return k8sutil.ClientSets{
					MClient: mClient,
					KClient: kClient,
				}
			},
		},
		{
			name:       "AlertmanagerSecretEmptyData",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.AlertmanagerList{})
				mClient.PrependReactor("get", "alertmanagers", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Alertmanager{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
						Spec: monitoringv1.AlertmanagerSpec{
							ConfigSecret: "test-secret",
						},
					}, nil
				})

				kClient := fake.NewSimpleClientset(&corev1.Secret{})
				kClient.PrependReactor("get", "secrets", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-secret",
							Namespace: tc.namespace,
						},
						Data: map[string][]byte{},
					}, nil
				})

				return k8sutil.ClientSets{
					MClient: mClient,
					KClient: kClient,
				}
			},
		},
		{
			name:       "AlertmanagerSecretKeyNotFound",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.AlertmanagerList{})
				mClient.PrependReactor("get", "alertmanagers", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Alertmanager{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
						Spec: monitoringv1.AlertmanagerSpec{
							ConfigSecret: "test-secret",
						},
					}, nil
				})

				kClient := fake.NewSimpleClientset(&corev1.Secret{})
				kClient.PrependReactor("get", "secrets", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-secret",
							Namespace: tc.namespace,
						},
						Data: map[string][]byte{
							"some-other-key": []byte("value"),
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
			name:       "AlertmanagerNamespaceSelectorWithoutMatchLabels",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.AlertmanagerList{})
				mClient.PrependReactor("get", "alertmanagers", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Alertmanager{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
						Spec: monitoringv1.AlertmanagerSpec{
							AlertmanagerConfigNamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"environment": "test"},
							},
						},
					}, nil
				})

				kClient := fake.NewSimpleClientset(&corev1.Namespace{})
				kClient.PrependReactor("get", "namespace", func(_ clienttesting.Action) (bool, runtime.Object, error) {
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
			name:       "AlertmanagerSelectorWithoutMatchLabels",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.AlertmanagerList{})
				mClient.PrependReactor("get", "alertmanagers", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Alertmanager{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
						Spec: monitoringv1.AlertmanagerSpec{
							AlertmanagerConfigNamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"amconfig": "test"},
							},
						},
					}, nil
				})

				mClient.PrependReactor("get", "alertmanagerconfigs", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1alpha1.AlertmanagerConfigList{
						Items: []*monitoringv1alpha1.AlertmanagerConfig{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "alertmanagerconfigs-crd",
									Namespace: tc.namespace,
									Labels:    map[string]string{"label": "another-value"},
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
			name:       "AlertmanagerFailedGetAMConfigs",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.AlertmanagerList{})
				mClient.PrependReactor("get", "alertmanagers", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.Alertmanager{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
						Spec: monitoringv1.AlertmanagerSpec{
							AlertmanagerConfiguration: &monitoringv1.AlertmanagerConfiguration{
								Name: "test-amconfig",
							},
						},
					}, nil
				})

				mClient.PrependReactor("get", "alertmanagerconfigs", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.NewNotFound(monitoringv1alpha1.Resource("alertmanagerconfigs"), tc.name)
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
			err := RunAlertmanagerAnalyzer(context.Background(), &clientSets, tc.name, tc.namespace)
			if tc.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
