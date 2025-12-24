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
	"encoding/json"
	"fmt"

	"github.com/prometheus-operator/poctl/internal/k8sutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func MigrateCRDs(ctx context.Context, clientSets *k8sutil.ClientSets) error {
	crds, err := clientSets.APIExtensionsClient.ApiextensionsV1().CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list CRDs: %w", err)
	}

	for _, crd := range crds.Items {
		if crd.Spec.Group != "monitoring.coreos.com" {
			continue
		}

		var storageVersion string
		for _, version := range crd.Spec.Versions {
			if version.Storage {
				storageVersion = version.Name
				break
			}
		}
		if storageVersion == "" {
			continue
		}

		crdResourceVersion := schema.GroupVersionResource{
			Group:    crd.Spec.Group,
			Version:  storageVersion,
			Resource: crd.Spec.Names.Plural,
		}

		namespaces, err := clientSets.KClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list Namespaces %v", err)
		}
		for _, namespace := range namespaces.Items {
			ns := namespace.Name

			customResourcesInstances, err := clientSets.DClient.Resource(crdResourceVersion).Namespace(ns).List(ctx, metav1.ListOptions{})
			if err != nil {
				continue
			}

			for _, cri := range customResourcesInstances.Items {
				name := cri.GetName()
				apiVersion := cri.GetAPIVersion()

				expectedAPIVersion := fmt.Sprintf("%s/%s", crd.Spec.Group, storageVersion)
				if apiVersion == expectedAPIVersion {
					continue
				}

				crdJSON, err := json.Marshal(cri.Object)
				if err != nil {
					continue
				}

				var unstructuredeObject map[string]interface{}
				if err := json.Unmarshal(crdJSON, &unstructuredeObject); err != nil {
					continue
				}

				unstructuredeObject["apiVersion"] = expectedAPIVersion

				updatedStorageObject := &unstructured.Unstructured{Object: unstructuredeObject}

				_, err = clientSets.DClient.Resource(crdResourceVersion).Namespace(ns).Create(ctx, updatedStorageObject, metav1.CreateOptions{})
				if err != nil {
					return fmt.Errorf("failed to create new version %s %s: %v", ns, name, err)
				}

				err = clientSets.DClient.Resource(crdResourceVersion).Namespace(ns).Delete(ctx, name, metav1.DeleteOptions{})
				if err != nil {
					return fmt.Errorf("failed to delete old version of %s/%s: %v", ns, name, err)
				}
			}
		}
	}

	fmt.Println("CRD migration completed.")
	return nil
}
