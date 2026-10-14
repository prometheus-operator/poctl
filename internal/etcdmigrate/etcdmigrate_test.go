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

package etcdmigrate

import (
	"context"
	"testing"

	"github.com/prometheus-operator/poctl/internal/k8sutil"
	"github.com/stretchr/testify/assert"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	fakeApiExtensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	fakeDynamicClient "k8s.io/client-go/dynamic/fake"
	clienttesting "k8s.io/client-go/testing"
)

func TestMigrateCRDs(t *testing.T) {
	type testCase struct {
		name                string
		namespace           string
		getMockedClientSets func(tc testCase) k8sutil.ClientSets
		shouldFail          bool
	}

	tests := []testCase{
		{
			name:       "FailCRDList",
			shouldFail: true,
			getMockedClientSets: func(_ testCase) k8sutil.ClientSets {
				apiExtensionsClient := fakeApiExtensions.NewSimpleClientset()
				apiExtensionsClient.PrependReactor("list", "customresourcedefinitions", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.NewInternalError(nil)
				})

				return k8sutil.ClientSets{
					APIExtensionsClient: apiExtensionsClient,
				}
			},
		},
		{
			name:       "FailObjectUpdate",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				crd := &apiextensions.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{Name: "testcrd"},
					Spec: apiextensions.CustomResourceDefinitionSpec{
						Group: "monitoring.coreos.com",
						Names: apiextensions.CustomResourceDefinitionNames{
							Plural: "probes",
						},
						Versions: []apiextensions.CustomResourceDefinitionVersion{},
					},
				}

				crInstance := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "monitoring.coreos.com/v1beta1",
						"kind":       "Probe",
						"metadata": map[string]interface{}{
							"name":      "probe",
							"namespace": tc.namespace,
						},
					},
				}

				apiExtensionsClient := fakeApiExtensions.NewSimpleClientset(crd)
				dClient := fakeDynamicClient.NewSimpleDynamicClient(runtime.NewScheme(), crInstance)
				dClient.PrependReactor("update", "probes", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.NewInternalError(nil)
				})

				return k8sutil.ClientSets{
					APIExtensionsClient: apiExtensionsClient,
					DClient:             dClient,
				}
			},
		},
		{
			name:       "SuccessUpdateStorageVersion",
			namespace:  "test",
			shouldFail: false,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				crd := &apiextensions.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{Name: "testcrd"},
					Spec: apiextensions.CustomResourceDefinitionSpec{
						Group: "monitoring.coreos.com",
						Names: apiextensions.CustomResourceDefinitionNames{
							Plural: "probes",
						},
						Versions: []apiextensions.CustomResourceDefinitionVersion{
							{Name: "v1", Storage: true},
						},
					},
				}

				crInstance := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "monitoring.coreos.com/v1beta1",
						"kind":       "Probe",
						"metadata": map[string]interface{}{
							"name":      "probe",
							"namespace": tc.namespace,
						},
					},
				}

				apiExtensionsClient := fakeApiExtensions.NewSimpleClientset(crd)
				dClient := fakeDynamicClient.NewSimpleDynamicClient(runtime.NewScheme(), crInstance)

				dClient.PrependReactor("update", "probes", func(action clienttesting.Action) (bool, runtime.Object, error) {
					updateAction, _ := action.(clienttesting.UpdateAction)
					obj := updateAction.GetObject().(*unstructured.Unstructured)
					apiVersion := obj.GetAPIVersion()

					if apiVersion != "monitoring.coreos.com/v1" {
						return true, nil, errors.NewInternalError(nil)
					}

					return true, obj, nil
				})

				return k8sutil.ClientSets{
					APIExtensionsClient: apiExtensionsClient,
					DClient:             dClient,
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clientSets := tc.getMockedClientSets(tc)

			err := MigrateCRDs(context.Background(), &clientSets)
			if tc.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
