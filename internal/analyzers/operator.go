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
	"fmt"

	"github.com/prometheus-operator/poctl/internal/crds"
	"github.com/prometheus-operator/poctl/internal/k8sutil"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RunOperatorAnalyzer(ctx context.Context, clientSets *k8sutil.ClientSets, name, namespace string) error {
	op, err := clientSets.KClient.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get Prometheus Operator deployment: %w", err)
	}

	cRb, err := clientSets.KClient.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=prometheus-operator",
	})

	if err != nil {
		return fmt.Errorf("failed to list RoleBindings: %w", err)
	}

	if !k8sutil.IsServiceAccountBoundToRoleBindingList(cRb, op.Spec.Template.Spec.ServiceAccountName) {
		return fmt.Errorf("ServiceAccount %s is not bound to any RoleBindings", op.Spec.Template.Spec.ServiceAccountName)
	}

	for _, crb := range cRb.Items {
		cr, err := clientSets.KClient.RbacV1().ClusterRoles().Get(ctx, crb.RoleRef.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get ClusterRole %s", crb.RoleRef.Name)
		}

		err = analyzeClusterRoleAndCRDRules(ctx, clientSets, crb, cr)
		if err != nil {
			return err
		}
	}

	return nil
}

func analyzeClusterRoleAndCRDRules(ctx context.Context, clientSets *k8sutil.ClientSets, crb v1.ClusterRoleBinding, cr *v1.ClusterRole) error {
	foundAPIGroup := false
	for _, rule := range cr.Rules {
		for _, apiGroup := range rule.APIGroups {
			if apiGroup == "monitoring.coreos.com" {
				foundAPIGroup = true
				err := analyzeCRDRules(ctx, clientSets, crb, rule)
				if err != nil {
					return err
				}
				break
			}
		}
	}

	if !foundAPIGroup {
		return fmt.Errorf("ClusterRole %s does not have monitoring.coreos.com APIGroup in its rules", crb.RoleRef.Name)
	}

	return nil
}

func analyzeCRDRules(ctx context.Context, clientSets *k8sutil.ClientSets, crb v1.ClusterRoleBinding, rule v1.PolicyRule) error {
	for _, crd := range crds.List {
		crd, err := clientSets.APIExtensionsClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crd, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get CRD %s", crd)
		}

		found := false
		for _, r := range rule.Resources {
			if r == crd.Spec.Names.Plural || r == crd.Spec.Names.Singular || r == crd.Spec.Names.Plural+"/finalizers" || r == crd.Spec.Names.Singular+"/finalizers" {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("ClusterRole %s does not have %s in its rules", crb.RoleRef.Name, crd.Spec.Names.Plural)
		}
	}
	return nil
}
