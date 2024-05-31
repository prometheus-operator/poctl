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

	"github.com/prometheus-operator/poctl/internal/builder"
	"github.com/prometheus-operator/poctl/internal/k8sutil"
	"github.com/prometheus-operator/poctl/internal/log"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// stackCmd represents the stack command.
var stackCmd = &cobra.Command{
	Use:   "stack",
	Short: "create a stack of Prometheus Operator resources.",
	Long:  `create a stack of Prometheus Operator resources.`,
	Run: func(cmd *cobra.Command, _ []string) {
		logger, err := log.NewLogger()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

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

		mclient, err := monitoringclient.NewForConfig(restConfig)
		if err != nil {
			logger.With("error", err.Error()).Error("error while creating Prometheus Operator client")
			os.Exit(1)
		}

		if err := createPrometheusOperator(cmd.Context(), logger, kclient, mclient, metav1.NamespaceDefault, "0.73.2"); err != nil {
			logger.With("error", err.Error()).Error("error while creating Prometheus Operator")
			os.Exit(1)
		}

		logger.Info("Prometheus Operator stack created successfully.")
	},
}

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

	err := k8sutil.CreateOrUpdateServiceAccount(ctx, logger, k8sClient, namespace, manifests.ServiceAccount)
	if err != nil {
		logger.With("error", err.Error()).Error("error while creating ServiceAccount", "serviceAccount", fmt.Sprintf("%s/%s", namespace, manifests.ServiceAccount.GetName()))
		return err
	}

	err = k8sutil.CreateOrUpdateClusterRole(ctx, logger, k8sClient, manifests.ClusterRole)
	if err != nil {
		logger.With("error", err.Error()).Error("error while creating ClusterRole", "clusterRole", manifests.ClusterRole.GetName())
		return err
	}

	err = k8sutil.CreateOrUpdateClusterRoleBinding(ctx, logger, k8sClient, manifests.ClusterRoleBinding)
	if err != nil {
		logger.With("error", err.Error()).Error("error while creating ClusterRoleBinding", "clusterRoleBinding", manifests.ClusterRoleBinding.GetName())
		return err
	}

	err = k8sutil.CreateOrUpdateService(ctx, logger, k8sClient, namespace, manifests.Service)
	if err != nil {
		logger.With("error", err.Error()).Error("error while creating/updating Service", "service", fmt.Sprintf("%s/%s", namespace, manifests.Service.GetName()))
		return err
	}

	err = k8sutil.CreateOrUpdateServiceMonitor(ctx, logger, poClient, namespace, manifests.ServiceMonitor)
	if err != nil {
		logger.With("error", err.Error()).Error("error while creating ServiceMonitor", "serviceMonitor", fmt.Sprintf("%s/%s", namespace, manifests.ServiceMonitor.GetName()))
		return err
	}

	err = k8sutil.CreateOrUpdateDeployment(ctx, logger, k8sClient, namespace, manifests.Deployment)
	if err != nil {
		logger.With("error", err.Error()).Error("error while creating Deployment", "deployment", fmt.Sprintf("%s/%s", namespace, manifests.Deployment.GetName()))
		return err
	}

	return nil
}
