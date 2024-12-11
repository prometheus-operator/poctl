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
// See the License for the specific language governing permissions and
// limitations under the License.

package analyzers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/prometheus-operator/poctl/internal/k8sutil"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RunPrometheusAnalyzer(ctx context.Context, clientSets *k8sutil.ClientSets, name, namespace string) error {
	prometheus, err := clientSets.MClient.MonitoringV1().Prometheuses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("prometheus %s not found in namespace %s", name, namespace)
		}
		return fmt.Errorf("error while getting Prometheus: %v", err)
	}

	cRb, err := clientSets.KClient.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{
		LabelSelector: "prometheus=prometheus",
	})
	if err != nil {
		return fmt.Errorf("failed to list RoleBindings: %w", err)
	}

	if !k8sutil.IsServiceAccountBoundToRoleBindingList(cRb, prometheus.Spec.ServiceAccountName) {
		return fmt.Errorf("serviceAccount %s is not bound to any RoleBindings", prometheus.Spec.ServiceAccountName)
	}

	for _, crb := range cRb.Items {
		cr, err := clientSets.KClient.RbacV1().ClusterRoles().Get(ctx, crb.RoleRef.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get ClusterRole %s", crb.RoleRef.Name)
		}

		err = k8sutil.CheckPrometheusClusterRoleRules(crb, cr)
		if err != nil {
			return err
		}
	}

	if err := k8sutil.CheckResourceNamespaceSelectors(ctx, *clientSets, prometheus.Spec.PodMonitorNamespaceSelector); err != nil {
		return fmt.Errorf("podMonitorNamespaceSelector is not properly defined: %s", err)
	}

	if err := k8sutil.CheckResourceNamespaceSelectors(ctx, *clientSets, prometheus.Spec.ProbeNamespaceSelector); err != nil {
		return fmt.Errorf("probeNamespaceSelector is not properly defined: %s", err)
	}

	if err := k8sutil.CheckResourceNamespaceSelectors(ctx, *clientSets, prometheus.Spec.ServiceMonitorNamespaceSelector); err != nil {
		return fmt.Errorf("serviceMonitorNamespaceSelector is not properly defined: %s", err)
	}

	if err := k8sutil.CheckResourceNamespaceSelectors(ctx, *clientSets, prometheus.Spec.ScrapeConfigNamespaceSelector); err != nil {
		return fmt.Errorf("scrapeConfigNamespaceSelector is not properly defined: %s", err)
	}

	if err := k8sutil.CheckResourceNamespaceSelectors(ctx, *clientSets, prometheus.Spec.RuleNamespaceSelector); err != nil {
		return fmt.Errorf("ruleNamespaceSelector is not properly defined: %s", err)
	}

	if err := k8sutil.CheckResourceLabelSelectors(ctx, *clientSets, prometheus.Spec.ServiceMonitorSelector, k8sutil.ServiceMonitor, namespace); err != nil {
		return fmt.Errorf("serviceMonitorSelector is not properly defined: %s", err)
	}

	if err := k8sutil.CheckResourceLabelSelectors(ctx, *clientSets, prometheus.Spec.PodMonitorSelector, k8sutil.PodMonitor, namespace); err != nil {
		return fmt.Errorf("podMonitorSelector is not properly defined: %s", err)
	}

	if err := k8sutil.CheckResourceLabelSelectors(ctx, *clientSets, prometheus.Spec.ProbeSelector, k8sutil.Probe, namespace); err != nil {
		return fmt.Errorf("probeSelector is not properly defined: %s", err)
	}

	if err := k8sutil.CheckResourceLabelSelectors(ctx, *clientSets, prometheus.Spec.ScrapeConfigSelector, k8sutil.ScrapeConfig, namespace); err != nil {
		return fmt.Errorf("scrapeConfigSelector is not properly defined: %s", err)
	}

	if err := k8sutil.CheckResourceLabelSelectors(ctx, *clientSets, prometheus.Spec.RuleSelector, k8sutil.PrometheusRule, namespace); err != nil {
		return fmt.Errorf("ruleSelector is not properly defined: %s", err)
	}

	slog.Info("Prometheus is compliant, no issues found", "name", name, "namespace", namespace)
	return nil
}
