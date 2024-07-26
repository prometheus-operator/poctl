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
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	fakeApiExtensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/utils/ptr"
)

func getDefaultDeployment(name, namespace string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(1)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "prometheus-operator",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "prometheus-operator",
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "prometheus-operator",
					Containers: []v1.Container{
						{
							Name:  "prometheus-operator",
							Image: "quay.io/coreos/prometheus-operator:v0.38.1",
						},
					},
				},
			},
		},
	}
}

func getDefaultClusterRoleBinding(namespace string) []rbacv1.ClusterRoleBinding {
	return []rbacv1.ClusterRoleBinding{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "prometheus-operator",
				Labels: map[string]string{
					"app.kubernetes.io/name": "prometheus-operator",
				},
			},
			RoleRef: rbacv1.RoleRef{
				Name: "prometheus-operator",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "prometheus-operator",
					Namespace: namespace,
				},
			},
		},
	}
}

func TestOperatorAnalyzer(t *testing.T) {
	type testCase struct {
		name                string
		namespace           string
		getMockedClientSets func(tc testCase) k8sutil.ClientSets
		shouldFail          bool
	}

	tests := []testCase{
		{
			name:       "OperatorDeploymentNotFound",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(_ testCase) k8sutil.ClientSets {
				kClient := fake.NewSimpleClientset(&appsv1.Deployment{})
				kClient.PrependReactor("get", "deployments", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.NewNotFound(appsv1.Resource("deployments"), "NotFound")
				})
				return k8sutil.ClientSets{
					KClient: kClient,
				}
			},
		},
		{
			name:       "OperatorRoleBindingListError",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				kClient := fake.NewSimpleClientset(&appsv1.Deployment{}, &rbacv1.ClusterRoleBindingList{})
				kClient.PrependReactor("get", "deployments", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, getDefaultDeployment(tc.name, tc.namespace), nil
				})
				kClient.PrependReactor("list", "clusterrolebindings", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.NewInternalError(errors.NewInternalError(errors.FromObject(nil)))
				})
				return k8sutil.ClientSets{
					KClient: kClient,
				}
			},
		},
		{
			name:       "ServiceAccountNotBoundToRoleBinding",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				kClient := fake.NewSimpleClientset(&appsv1.Deployment{}, &rbacv1.ClusterRoleBindingList{})
				kClient.PrependReactor("get", "deployments", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					deployment := getDefaultDeployment(tc.name, tc.namespace)
					deployment.Spec.Template.Spec.ServiceAccountName = "not-bound-service-account"
					return true, deployment, nil
				})
				kClient.PrependReactor("list", "clusterrolebindings", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &rbacv1.ClusterRoleBindingList{
						Items: getDefaultClusterRoleBinding(tc.namespace),
					}, nil
				})
				return k8sutil.ClientSets{
					KClient: kClient,
				}
			},
		},
		{
			name:       "ApiGroupNotFoundInClusterRole",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				kClient := fake.NewSimpleClientset(&appsv1.Deployment{}, &rbacv1.ClusterRoleBindingList{})
				kClient.PrependReactor("get", "deployments", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					deployment := getDefaultDeployment(tc.name, tc.namespace)
					return true, deployment, nil
				})
				kClient.PrependReactor("list", "clusterrolebindings", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &rbacv1.ClusterRoleBindingList{
						Items: getDefaultClusterRoleBinding(tc.namespace),
					}, nil
				})

				kClient.PrependReactor("get", "clusterroles", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &rbacv1.ClusterRole{
						ObjectMeta: metav1.ObjectMeta{
							Name: "prometheus-operator",
						},
						Rules: []rbacv1.PolicyRule{
							{
								APIGroups: []string{"not-monitoring.coreos.com"},
								Resources: []string{"prometheuses", "prometheusrules", "servicemonitors", "podmonitors", "thanosrulers"},
							},
						},
					}, nil
				})

				return k8sutil.ClientSets{
					KClient: kClient,
				}
			},
		},
		{
			name:       "CrdNotFoundInClusterRole",
			namespace:  "test",
			shouldFail: true,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				kClient := fake.NewSimpleClientset(&appsv1.Deployment{}, &rbacv1.ClusterRoleBindingList{})

				apiExtensionsClient := fakeApiExtensions.NewSimpleClientset(&apiextensions.CustomResourceDefinition{}, &apiextensions.CustomResourceDefinitionList{})

				kClient.PrependReactor("get", "deployments", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					deployment := getDefaultDeployment(tc.name, tc.namespace)
					return true, deployment, nil
				})

				kClient.PrependReactor("list", "clusterrolebindings", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &rbacv1.ClusterRoleBindingList{
						Items: getDefaultClusterRoleBinding(tc.namespace),
					}, nil
				})

				kClient.PrependReactor("get", "clusterroles", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &rbacv1.ClusterRole{
						ObjectMeta: metav1.ObjectMeta{
							Name: "prometheus-operator",
						},
						Rules: []rbacv1.PolicyRule{
							{
								APIGroups: []string{"monitoring.coreos.com"},
								Resources: []string{"prometheuses", "prometheusrules", "servicemonitors", "podmonitors", "thanosrulers"},
							},
						},
					}, nil
				})

				apiExtensionsClient.PrependReactor("get", "customresourcedefinitions", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &apiextensions.CustomResourceDefinition{
						ObjectMeta: metav1.ObjectMeta{
							Name: "alertmanager.monitoring.coreos.com",
						},
						Spec: apiextensions.CustomResourceDefinitionSpec{
							Names: apiextensions.CustomResourceDefinitionNames{
								Singular: "alertmanager",
								Plural:   "alertmanagers",
							},
						},
					}, nil
				})

				return k8sutil.ClientSets{
					KClient:             kClient,
					APIExtensionsClient: apiExtensionsClient,
				}
			},
		},
		{
			name:       "CrdFoundInClusterRole",
			namespace:  "test",
			shouldFail: false,
			getMockedClientSets: func(tc testCase) k8sutil.ClientSets {
				kClient := fake.NewSimpleClientset(&appsv1.Deployment{}, &rbacv1.ClusterRoleBindingList{})

				apiExtensionsClient := fakeApiExtensions.NewSimpleClientset(&apiextensions.CustomResourceDefinition{}, &apiextensions.CustomResourceDefinitionList{})

				kClient.PrependReactor("get", "deployments", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					deployment := getDefaultDeployment(tc.name, tc.namespace)
					return true, deployment, nil
				})

				kClient.PrependReactor("list", "clusterrolebindings", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &rbacv1.ClusterRoleBindingList{
						Items: getDefaultClusterRoleBinding(tc.namespace),
					}, nil
				})

				kClient.PrependReactor("get", "clusterroles", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &rbacv1.ClusterRole{
						ObjectMeta: metav1.ObjectMeta{
							Name: "prometheus-operator",
						},
						Rules: []rbacv1.PolicyRule{
							{
								APIGroups: []string{"monitoring.coreos.com"},
								Resources: []string{"prometheuses", "prometheusrules", "servicemonitors", "podmonitors", "thanosrulers", "alertmanagers"},
							},
						},
					}, nil
				})

				apiExtensionsClient.PrependReactor("get", "customresourcedefinitions", func(_ clienttesting.Action) (bool, runtime.Object, error) {
					return true, &apiextensions.CustomResourceDefinition{
						ObjectMeta: metav1.ObjectMeta{
							Name: "alertmanager.monitoring.coreos.com",
						},
						Spec: apiextensions.CustomResourceDefinitionSpec{
							Names: apiextensions.CustomResourceDefinitionNames{
								Singular: "alertmanager",
								Plural:   "alertmanagers",
							},
						},
					}, nil
				})

				return k8sutil.ClientSets{
					KClient:             kClient,
					APIExtensionsClient: apiExtensionsClient,
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clientSets := tc.getMockedClientSets(tc)
			err := RunOperatorAnalyzer(context.Background(), &clientSets, tc.name, tc.namespace)
			if tc.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
