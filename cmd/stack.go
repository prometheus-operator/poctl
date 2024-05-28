/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
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

	_, err := k8sClient.CoreV1().ServiceAccounts(namespace).Create(ctx, manifests.ServiceAccount, metav1.CreateOptions{})
	if err != nil {
		logger.With("error", err.Error()).Error("error while creating ServiceAccount")
		return err
	}

	_, err = k8sClient.RbacV1().ClusterRoles().Create(ctx, manifests.ClusterRole, metav1.CreateOptions{})
	if err != nil {
		logger.With("error", err.Error()).Error("error while creating ClusterRole")
		return err
	}

	_, err = k8sClient.RbacV1().ClusterRoleBindings().Create(ctx, manifests.ClusterRoleBinding, metav1.CreateOptions{})
	if err != nil {
		logger.With("error", err.Error()).Error("error while creating ClusterRoleBinding")
		return err
	}

	_, err = k8sClient.CoreV1().Services(namespace).Create(ctx, manifests.Service, metav1.CreateOptions{})
	if err != nil {
		logger.With("error", err.Error()).Error("error while creating Service")
		return err
	}

	_, err = poClient.MonitoringV1().ServiceMonitors(namespace).Create(ctx, manifests.ServiceMonitor, metav1.CreateOptions{})
	if err != nil {
		logger.With("error", err.Error()).Error("error while creating ServiceMonitor")
		return err
	}

	_, err = k8sClient.AppsV1().Deployments(namespace).Create(ctx, manifests.Deployment, metav1.CreateOptions{})
	if err != nil {
		logger.With("error", err.Error()).Error("error while creating Deployment")
		return err
	}

	logger.Debug("Prometheus Operator manifest created successfully.")
	return nil
}
