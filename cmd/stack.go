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

package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/go-github/v62/github"
	"github.com/prometheus-operator/poctl/internal/builder"
	"github.com/prometheus-operator/poctl/internal/k8sutil"
	"github.com/prometheus-operator/poctl/internal/log"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var (
	stackCmd = &cobra.Command{
		Use:   "stack",
		Short: "create a stack of Prometheus Operator resources.",
		Long:  `create a stack of Prometheus Operator resources.`,
		Run:   run,
	}

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

func init() {
	createCmd.AddCommand(stackCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stackCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stackCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func run(cmd *cobra.Command, _ []string) {
	logger, err := log.NewLogger()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	version, err := cmd.Flags().GetString("version")
	if err != nil {
		logger.With("error", err.Error()).Error("error while getting version flag")
		os.Exit(1)
	}

	logger.Info(version)

	//TODO(nicolastakashi): Replace it when the PR #6623 is merged
	restConfig, err := k8sutil.GetRestConfig(logger, kubeconfig)
	if err != nil {
		logger.With("error", err.Error()).Error("error while getting kubeconfig")
		os.Exit(1)
	}

	kclient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		logger.With("error", err.Error()).Error("error while creating k8s client")
		os.Exit(1)
	}

	kdynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		logger.With("error", err.Error()).Error("error while creating dynamic client")
		os.Exit(1)
	}

	mclient, err := monitoringclient.NewForConfig(restConfig)
	if err != nil {
		logger.With("error", err.Error()).Error("error while creating Prometheus Operator client")
		os.Exit(1)
	}

	gitHubClient := github.NewClient(nil)

	if err := installCRDs(cmd.Context(), logger, version, kdynamicClient, gitHubClient); err != nil {
		logger.With("error", err.Error()).Error("error while installing CRDs")
		os.Exit(1)
	}

	if err := createPrometheusOperator(cmd.Context(), logger, kclient, mclient, metav1.NamespaceDefault, version); err != nil {
		logger.With("error", err.Error()).Error("error while creating Prometheus Operator")
		os.Exit(1)
	}

	if err := createPrometheus(cmd.Context(), logger, kclient, mclient, metav1.NamespaceDefault); err != nil {
		logger.With("error", err.Error()).Error("error while creating Prometheus")
		os.Exit(1)
	}

	if err := createAlertManager(cmd.Context(), logger, kclient, mclient, metav1.NamespaceDefault); err != nil {
		logger.With("error", err.Error()).Error("error while creating AlertManager")
		os.Exit(1)
	}

	logger.Info("Prometheus Operator stack created successfully.")
}

func installCRDs(
	ctx context.Context,
	logger *slog.Logger,
	version string,
	k8sClient *dynamic.DynamicClient,
	gitHubClient *github.Client) error {

	nodeResource := schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}

	for _, crd := range crds {
		l := logger.With("crd", crd)

		reader, _, err := gitHubClient.Repositories.DownloadContents(
			ctx,
			"prometheus-operator",
			"prometheus-operator",
			fmt.Sprintf("example/prometheus-operator-crd/monitoring.coreos.com_%s.yaml", crd),
			&github.RepositoryContentGetOptions{
				Ref: fmt.Sprintf("v%s", version),
			})

		if err != nil {
			l.Error("error while downloading crds", "error", err)
			return err
		}

		crds, err := k8sutil.CrdDeserilezer(logger, reader)
		if err != nil {
			l.Error("error while deserializing crds", "error", err)
			return err
		}

		unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(crds)
		if err != nil {
			l.Error("error while converting CRDs to Unstructured", "error", err)
			return err
		}

		_, err = k8sClient.Resource(nodeResource).Apply(ctx, fmt.Sprintf("%s.monitoring.coreos.com", crd), &unstructured.Unstructured{Object: unstructuredObj}, k8sutil.ApplyOption)

		if err != nil {
			l.Error("error while applying", "error", err)
			return err
		}

		logger.Info("applied successfully", "CRD", crd)
	}

	return nil
}

func createPrometheusOperator(
	ctx context.Context,
	logger *slog.Logger,
	k8sClient *kubernetes.Clientset,
	poClient *monitoringclient.Clientset,
	namespace, version string) error {
	manifests := builder.NewOperator(namespace, version).
		WithServiceAccount().
		WithClusterRole().
		WithClusterRoleBinding().
		WithService().
		WithServiceMonitor().
		WithDeployment().
		Build()

	_, err := k8sClient.CoreV1().ServiceAccounts(namespace).Apply(ctx, manifests.ServiceAccount, k8sutil.ApplyOption)
	if err != nil {
		logger.ErrorContext(ctx, "error while creating ServiceAccount", "error", err.Error())
		return err
	}

	_, err = k8sClient.RbacV1().ClusterRoles().Apply(ctx, manifests.ClusterRole, k8sutil.ApplyOption)
	if err != nil {
		logger.ErrorContext(ctx, "error while creating ClusterRole", "error", err.Error())
		return err
	}

	_, err = k8sClient.RbacV1().ClusterRoleBindings().Apply(ctx, manifests.ClusterRoleBinding, k8sutil.ApplyOption)
	if err != nil {
		logger.ErrorContext(ctx, "error while creating ClusterRoleBinding", "error", err.Error())
		return err
	}

	_, err = k8sClient.CoreV1().Services(namespace).Apply(ctx, manifests.Service, k8sutil.ApplyOption)
	if err != nil {
		logger.ErrorContext(ctx, "error while creating Service", "error", err.Error())
		return err
	}

	_, err = poClient.MonitoringV1().ServiceMonitors(namespace).Apply(ctx, manifests.ServiceMonitor, k8sutil.ApplyOption)
	if err != nil {
		logger.ErrorContext(ctx, "error while creating ServiceMonitor", "error", err.Error())
		return err
	}

	_, err = k8sClient.AppsV1().Deployments(namespace).Apply(ctx, manifests.Deployment, k8sutil.ApplyOption)
	if err != nil {
		logger.ErrorContext(ctx, "error while creating Deployment", "error", err.Error())
		return err
	}

	return nil
}

func createPrometheus(
	ctx context.Context,
	logger *slog.Logger,
	k8sClient *kubernetes.Clientset,
	poClient *monitoringclient.Clientset,
	namespace string) error {
	manifests := builder.NewPrometheus(namespace).
		WithServiceAccount().
		WithClusterRole().
		WithClusterRoleBinding().
		WithService().
		WithServiceMonitor().
		WithPrometheus().
		Build()

	_, err := k8sClient.CoreV1().ServiceAccounts(namespace).Apply(ctx, manifests.ServiceAccount, k8sutil.ApplyOption)
	if err != nil {
		logger.ErrorContext(ctx, "error while creating ServiceAccount", "error", err.Error())
		return err
	}

	_, err = k8sClient.RbacV1().ClusterRoles().Apply(ctx, manifests.ClusterRole, k8sutil.ApplyOption)
	if err != nil {
		logger.ErrorContext(ctx, "error while creating ClusterRole", "error", err.Error())
		return err
	}

	_, err = k8sClient.RbacV1().ClusterRoleBindings().Apply(ctx, manifests.ClusterRoleBinding, k8sutil.ApplyOption)
	if err != nil {
		logger.ErrorContext(ctx, "error while creating ClusterRoleBinding", "error", err.Error())
		return err
	}

	_, err = poClient.MonitoringV1().Prometheuses(namespace).Apply(ctx, manifests.Prometheus, k8sutil.ApplyOption)
	if err != nil {
		logger.ErrorContext(ctx, "error while creating Prometheus", "error", err.Error())
		return err
	}

	_, err = k8sClient.CoreV1().Services(namespace).Apply(ctx, manifests.Service, k8sutil.ApplyOption)
	if err != nil {
		logger.ErrorContext(ctx, "error while creating Service", "error", err.Error())
		return err
	}

	_, err = poClient.MonitoringV1().ServiceMonitors(namespace).Apply(ctx, manifests.ServiceMonitor, k8sutil.ApplyOption)
	if err != nil {
		logger.ErrorContext(ctx, "error while creating ServiceMonitor", "error", err.Error())
		return err
	}

	return nil
}

func createAlertManager(
	ctx context.Context,
	logger *slog.Logger,
	k8sClient *kubernetes.Clientset,
	poClient *monitoringclient.Clientset,
	namespace string) error {
	manifests := builder.NewAlertManager(namespace).
		WithServiceAccount().
		WithAlertManager().
		WithService().
		WithServiceMonitor().
		Build()

	_, err := k8sClient.CoreV1().ServiceAccounts(namespace).Apply(ctx, manifests.ServiceAccount, k8sutil.ApplyOption)
	if err != nil {
		logger.ErrorContext(ctx, "error while creating ServiceAccount", "error", err.Error())
		return err
	}

	_, err = poClient.MonitoringV1().Alertmanagers(namespace).Apply(ctx, manifests.AlertManager, k8sutil.ApplyOption)
	if err != nil {
		logger.ErrorContext(ctx, "error while creating AlertManager", "error", err.Error())
		return err
	}

	_, err = k8sClient.CoreV1().Services(namespace).Apply(ctx, manifests.Service, k8sutil.ApplyOption)
	if err != nil {
		logger.ErrorContext(ctx, "error while creating Service", "error", err.Error())
		return err
	}

	_, err = poClient.MonitoringV1().ServiceMonitors(namespace).Apply(ctx, manifests.ServiceMonitor, k8sutil.ApplyOption)
	if err != nil {
		logger.ErrorContext(ctx, "error while creating ServiceMonitor", "error", err.Error())
		return err
	}

	return nil
}
