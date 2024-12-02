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
	"k8s.io/apimachinery/pkg/api/errors"
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
