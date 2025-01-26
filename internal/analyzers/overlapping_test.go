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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

func TestOverlappingAnalyzer(t *testing.T) {
	type testCase struct {
		name                string
		namespace           string
		getMockedClientSets func(tc testCase) k8sutil.ClientSets
		shouldFail          bool
	}
	tests := []testCase{
		{
			name:       "ErrorListingServiceMonitor",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.ServiceMonitorList{})
				mClient.PrependReactor("list", "servicemonitors", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.NewNotFound(monitoringv1.Resource("servicemonitors"), tc.name)
				})

				return k8sutil.ClientSets{
					MClient: mClient,
				}
			},
		},
		{
			name:       "ErrorListingPodMonitor",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.PodMonitorList{})
				mClient.PrependReactor("list", "podmonitors", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.NewNotFound(monitoringv1.Resource("podmonitors"), tc.name)
				})

				return k8sutil.ClientSets{
					MClient: mClient,
				}
			},
		},
		{
			name:       "OverlapingPodMonitor",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				mClient := monitoringclient.NewSimpleClientset(&monitoringv1.PodMonitorList{})
				mClient.PrependReactor("list", "podmonitors", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &monitoringv1.PodMonitorList{
						Items: []*monitoringv1.PodMonitor{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "podmonitor-1",
									Namespace: "test",
								},
								Spec: monitoringv1.PodMonitorSpec{
									Selector: metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "overlapping-app",
										},
									},
									PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
										{Port: "http-metrics"},
									},
								},
							},
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "podmonitor-2",
									Namespace: "test",
								},
								Spec: monitoringv1.PodMonitorSpec{
									Selector: metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": "overlapping-app",
										},
									},
									PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
										{Port: "http-metrics"},
									},
								},
							},
						},
					}, nil
				})

				kClient := fake.NewSimpleClientset(&corev1.PodList{})
				kClient.PrependReactor("list", "pods", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &corev1.PodList{
						Items: []corev1.Pod{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "overlapping-pod",
									Namespace: "test",
									Labels: map[string]string{
										"app": "overlapping-app",
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
			err := RunOverlappingAnalyzer(context.Background(), &clientSets, tc.name, tc.namespace)
			if tc.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
