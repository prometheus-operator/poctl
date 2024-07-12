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

package create

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/go-github/v62/github"
	"github.com/prometheus-operator/poctl/internal/builder"
	"github.com/prometheus-operator/poctl/internal/k8sutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func RunCreateStack(ctx context.Context, logger *slog.Logger, clientSets *k8sutil.ClientSets, gitHubClient *github.Client, version string) error {
	if err := installCRDs(ctx, logger, version, clientSets, gitHubClient); err != nil {
		logger.Error("error while installing CRDs", "error", err)
		return err
	}

	if err := createPrometheusOperator(ctx, clientSets, metav1.NamespaceDefault, version); err != nil {
		logger.Error("error while creating Prometheus Operator", "error", err)
		return err
	}

	if err := createPrometheus(ctx, clientSets, metav1.NamespaceDefault); err != nil {
		logger.Error("error while creating Prometheus", "error", err)
		return err
	}

	if err := createAlertManager(ctx, clientSets, metav1.NamespaceDefault); err != nil {
		logger.Error("error while creating AlertManager", "error", err)
		return err
	}

	if err := createNodeExporter(ctx, clientSets, metav1.NamespaceDefault); err != nil {
		logger.Error("error while creating NodeExporter", "error", err)
		return err
	}

	if err := createKubeStateMetrics(ctx, clientSets, metav1.NamespaceDefault); err != nil {
		logger.Error("error while creating KubeStateMetrics", "error", err)
		return err
	}

	return nil
}

var (
	crds = []string{
		"alertmanagers",
		"alertmanagerconfigs",
		"podmonitors",
		"probes",
		"prometheusagents",
		"prometheuses",
		"prometheusrules",
		"scrapeconfigs",
		"servicemonitors",
		"thanosrulers",
	}
)

func installCRDs(
	ctx context.Context,
	logger *slog.Logger,
	version string,
	clientSets *k8sutil.ClientSets,
	gitHubClient *github.Client) error {

	nodeResource := schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}

	for _, crd := range crds {
		reader, _, err := gitHubClient.Repositories.DownloadContents(
			ctx,
			"prometheus-operator",
			"prometheus-operator",
			fmt.Sprintf("example/prometheus-operator-crd/monitoring.coreos.com_%s.yaml", crd),
			&github.RepositoryContentGetOptions{
				Ref: fmt.Sprintf("v%s", version),
			})

		if err != nil {
			return fmt.Errorf("error while downloading crds: %v", err)
		}

		crds, err := k8sutil.CrdDeserilezer(logger, reader)
		if err != nil {
			return fmt.Errorf("error while deserializing crds: %v", err)
		}

		unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(crds)
		if err != nil {
			return fmt.Errorf("error while converting CRDs to Unstructured: %v", err)
		}

		_, err = clientSets.DClient.Resource(nodeResource).Apply(ctx, fmt.Sprintf("%s.monitoring.coreos.com", crd), &unstructured.Unstructured{Object: unstructuredObj}, k8sutil.ApplyOption)

		if err != nil {
			return fmt.Errorf("error while applying CRD: %v", err)
		}

		logger.Info("applied successfully", "CRD", crd)
	}

	return nil
}

func createPrometheusOperator(
	ctx context.Context,
	clientSets *k8sutil.ClientSets,
	namespace, version string) error {
	manifests := builder.NewOperator(namespace, version).
		WithServiceAccount().
		WithClusterRole().
		WithClusterRoleBinding().
		WithService().
		WithServiceMonitor().
		WithDeployment().
		Build()

	_, err := clientSets.KClient.CoreV1().ServiceAccounts(namespace).Apply(ctx, manifests.ServiceAccount, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating ServiceAccount: %v", err)
	}

	_, err = clientSets.KClient.RbacV1().ClusterRoles().Apply(ctx, manifests.ClusterRole, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating ClusterRole: %v", err)
	}

	_, err = clientSets.KClient.RbacV1().ClusterRoleBindings().Apply(ctx, manifests.ClusterRoleBinding, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating ClusterRoleBinding: %v", err)
	}

	_, err = clientSets.KClient.CoreV1().Services(namespace).Apply(ctx, manifests.Service, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating Service: %v", err)
	}

	_, err = clientSets.MClient.MonitoringV1().ServiceMonitors(namespace).Apply(ctx, manifests.ServiceMonitor, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating ServiceMonitor: %v", err)
	}

	_, err = clientSets.KClient.AppsV1().Deployments(namespace).Apply(ctx, manifests.Deployment, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating Deployment: %v", err)
	}

	return nil
}

func createPrometheus(
	ctx context.Context,
	clientSets *k8sutil.ClientSets,
	namespace string) error {
	manifests := builder.NewPrometheus(namespace).
		WithServiceAccount().
		WithClusterRole().
		WithClusterRoleBinding().
		WithService().
		WithServiceMonitor().
		WithPrometheus().
		Build()

	_, err := clientSets.KClient.CoreV1().ServiceAccounts(namespace).Apply(ctx, manifests.ServiceAccount, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating ServiceAccount: %v", err)
	}

	_, err = clientSets.KClient.RbacV1().ClusterRoles().Apply(ctx, manifests.ClusterRole, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating ClusterRole: %v", err)
	}

	_, err = clientSets.KClient.RbacV1().ClusterRoleBindings().Apply(ctx, manifests.ClusterRoleBinding, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating ClusterRoleBinding: %v", err)
	}

	_, err = clientSets.MClient.MonitoringV1().Prometheuses(namespace).Apply(ctx, manifests.Prometheus, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating Prometheus: %v", err)
	}

	_, err = clientSets.KClient.CoreV1().Services(namespace).Apply(ctx, manifests.Service, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating Service: %v", err)
	}

	_, err = clientSets.MClient.MonitoringV1().ServiceMonitors(namespace).Apply(ctx, manifests.ServiceMonitor, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating ServiceMonitor: %v", err)
	}

	return nil
}

func createAlertManager(
	ctx context.Context,
	clientSets *k8sutil.ClientSets,
	namespace string) error {
	manifests := builder.NewAlertManager(namespace).
		WithServiceAccount().
		WithAlertManager().
		WithService().
		WithServiceMonitor().
		Build()

	_, err := clientSets.KClient.CoreV1().ServiceAccounts(namespace).Apply(ctx, manifests.ServiceAccount, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating ServiceAccount: %v", err)
	}

	_, err = clientSets.MClient.MonitoringV1().Alertmanagers(namespace).Apply(ctx, manifests.AlertManager, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating AlertManager: %v", err)
	}

	_, err = clientSets.KClient.CoreV1().Services(namespace).Apply(ctx, manifests.Service, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating Service: %v", err)
	}

	_, err = clientSets.MClient.MonitoringV1().ServiceMonitors(namespace).Apply(ctx, manifests.ServiceMonitor, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating ServiceMonitor: %v", err)
	}

	return nil
}

func createNodeExporter(ctx context.Context, clientSets *k8sutil.ClientSets, namespace string) error {
	manifests := builder.NewNodeExporterBuilder(namespace, builder.LatestNodeExporterVersion).
		WithServiceAccount().
		WithDaemonSet().
		WithPodMonitor().
		Build()

	_, err := clientSets.KClient.CoreV1().ServiceAccounts(namespace).Apply(ctx, manifests.ServiceAccount, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating ServiceAccount: %v", err)
	}

	_, err = clientSets.KClient.AppsV1().DaemonSets(namespace).Apply(ctx, manifests.DaemonSet, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating DaemonSet: %v", err)
	}

	_, err = clientSets.MClient.MonitoringV1().PodMonitors(namespace).Apply(ctx, manifests.PodMonitor, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating PodMonitor: %v", err)
	}

	return nil
}

func createKubeStateMetrics(ctx context.Context, clientSets *k8sutil.ClientSets, namespace string) error {
	manifests := builder.NewKubeStateMetricsBuilder(namespace, builder.LatestKubeStateMetricsVersion).
		WithServiceAccount().
		WithClusterRole().
		WithClusterRoleBinding().
		WithDeployment().
		WithService().
		WithServiceMonitor().
		Build()

	_, err := clientSets.KClient.CoreV1().ServiceAccounts(namespace).Apply(ctx, manifests.ServiceAccount, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating ServiceAccount: %v", err)
	}

	_, err = clientSets.KClient.RbacV1().ClusterRoles().Apply(ctx, manifests.ClusterRole, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating ClusterRole: %v", err)
	}

	_, err = clientSets.KClient.RbacV1().ClusterRoleBindings().Apply(ctx, manifests.ClusterRoleBinding, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating ClusterRoleBinding: %v", err)
	}

	_, err = clientSets.KClient.AppsV1().Deployments(namespace).Apply(ctx, manifests.Deployment, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating Deployment: %v", err)
	}

	_, err = clientSets.KClient.CoreV1().Services(namespace).Apply(ctx, manifests.Service, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating Service: %v", err)
	}

	_, err = clientSets.MClient.MonitoringV1().ServiceMonitors(namespace).Apply(ctx, manifests.ServiceMonitor, k8sutil.ApplyOption)
	if err != nil {
		return fmt.Errorf("error while creating ServiceMonitor: %v", err)
	}
	return nil
}
