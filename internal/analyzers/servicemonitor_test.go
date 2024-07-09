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
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/fake"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

func TestServiceMonitorAnalyzer(t *testing.T) {
	type testCase struct {
		name                string
		namespace           string
		getMockedClientSets func(tc testCase) k8sutil.ClientSets
		shouldFail          bool
	}

	// table test
	tests := []testCase{
		{
			name:       "ServiceMonitorEmptySelector",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.ServiceMonitorList{})
				mClient.PrependReactor("get", "servicemonitors", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.ServiceMonitor{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
						Spec: monitoringv1.ServiceMonitorSpec{
							Endpoints: []monitoringv1.Endpoint{
								{
									Port: "http",
								},
							},
							Selector: metav1.LabelSelector{
								MatchLabels:      map[string]string{},
								MatchExpressions: []metav1.LabelSelectorRequirement{},
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
			name:       "ServiceMonitorMatchLabelsNoServices",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.ServiceMonitorList{})
				mClient.PrependReactor("get", "servicemonitors", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.ServiceMonitor{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
						Spec: monitoringv1.ServiceMonitorSpec{
							Endpoints: []monitoringv1.Endpoint{
								{
									Port: "http",
								},
							},
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app": "test",
								},
							},
						},
					}, nil
				})

				kClient := fake.NewSimpleClientset(&v1.ServiceList{})
				kClient.PrependReactor("list", "services", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &v1.ServiceList{
						Items: []v1.Service{},
					}, nil
				})

				return k8sutil.ClientSets{
					MClient: mClient,
					KClient: kClient,
				}
			},
		},
		{
			name:       "ServiceMonitorMatchLabelsNoMatchingPorts",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.ServiceMonitorList{})
				mClient.PrependReactor("get", "servicemonitors", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.ServiceMonitor{
						ObjectMeta: metav1.ObjectMeta{
							Name:      tc.name,
							Namespace: tc.namespace,
						},
						Spec: monitoringv1.ServiceMonitorSpec{
							Endpoints: []monitoringv1.Endpoint{
								{
									Port: "http",
								},
							},
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app": "test",
								},
							},
						},
					}, nil
				})

				kClient := fake.NewSimpleClientset(&v1.ServiceList{})
				kClient.PrependReactor("list", "services", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &v1.ServiceList{
						Items: []v1.Service{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "test",
									Namespace: tc.namespace,
									Labels: map[string]string{
										"app": "test",
									},
								},
								Spec: v1.ServiceSpec{
									Ports: []v1.ServicePort{
										{
											Name: "https",
										},
									},
									Selector: map[string]string{
										"app": "test",
									},
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clientSets := tc.getMockedClientSets(tc)
			err := RunServiceMonitorAnalyzer(context.Background(), &clientSets, tc.name, tc.namespace)
			if tc.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
