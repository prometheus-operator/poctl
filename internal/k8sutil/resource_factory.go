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

package k8sutil

import (
	"context"
	"fmt"
	"log/slog"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateOrUpdateService(ctx context.Context, logger *slog.Logger, client kubernetes.Interface, namespace string, service *v1.Service) error {
	svcClient := client.CoreV1().Services(namespace)

	_, err := svcClient.Get(ctx, service.Name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		_, err = svcClient.Create(ctx, service, metav1.CreateOptions{})
		if err == nil {
			logger.Info("service created", "service", fmt.Sprintf("%s/%s", namespace, service.Name))
		}
		return err
	}

	_, err = svcClient.Update(ctx, service, metav1.UpdateOptions{})
	if err == nil {
		logger.Info("service updated", "service", fmt.Sprintf("%s/%s", namespace, service.Name))
	}
	return err
}

func CreateOrUpdateServiceAccount(ctx context.Context, logger *slog.Logger, client kubernetes.Interface, namespace string, serviceAccount *v1.ServiceAccount) error {
	saClient := client.CoreV1().ServiceAccounts(namespace)

	_, err := saClient.Get(ctx, serviceAccount.Name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		_, err = saClient.Create(ctx, serviceAccount, metav1.CreateOptions{})
		if err == nil {
			logger.Info("serviceAccount created", "serviceAccount", fmt.Sprintf("%s/%s", namespace, serviceAccount.Name))
		}
		return err
	}

	_, err = saClient.Update(ctx, serviceAccount, metav1.UpdateOptions{})
	if err == nil {
		logger.Info("serviceAccount updated", "serviceAccount", fmt.Sprintf("%s/%s", namespace, serviceAccount.Name))
	}
	return err
}

func CreateOrUpdateClusterRole(ctx context.Context, logger *slog.Logger, client kubernetes.Interface, clusterRole *rbacv1.ClusterRole) error {
	crClient := client.RbacV1().ClusterRoles()

	_, err := crClient.Get(ctx, clusterRole.Name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		_, err = crClient.Create(ctx, clusterRole, metav1.CreateOptions{})
		if err == nil {
			logger.Info("clusterRole created", "clusterRole", clusterRole.Name)
		}
		return err
	}

	_, err = crClient.Update(ctx, clusterRole, metav1.UpdateOptions{})
	if err == nil {
		logger.Info("clusterRole updated", "clusterRole", clusterRole.Name)
	}
	return err
}

func CreateOrUpdateClusterRoleBinding(ctx context.Context, logger *slog.Logger, client kubernetes.Interface, clusterRoleBinding *rbacv1.ClusterRoleBinding) error {
	crbClient := client.RbacV1().ClusterRoleBindings()

	_, err := crbClient.Get(ctx, clusterRoleBinding.Name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		_, err = crbClient.Create(ctx, clusterRoleBinding, metav1.CreateOptions{})
		if err == nil {
			logger.Info("clusterRoleBinding created", "clusterRoleBinding", clusterRoleBinding.Name)
		}
		return err
	}

	_, err = crbClient.Update(ctx, clusterRoleBinding, metav1.UpdateOptions{})
	if err == nil {
		logger.Info("clusterRoleBinding updated", "clusterRoleBinding", clusterRoleBinding.Name)
	}
	return err
}

func CreateOrUpdateDeployment(ctx context.Context, logger *slog.Logger, client kubernetes.Interface, namespace string, deployment *appsv1.Deployment) error {
	deployClient := client.AppsV1().Deployments(namespace)

	_, err := deployClient.Get(ctx, deployment.Name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		_, err = deployClient.Create(ctx, deployment, metav1.CreateOptions{})
		if err == nil {
			logger.Info("deployment created", "deployment", fmt.Sprintf("%s/%s", namespace, deployment.Name))
		}
		return err
	}

	_, err = deployClient.Update(ctx, deployment, metav1.UpdateOptions{})
	if err == nil {
		logger.Info("deployment updated", "deployment", fmt.Sprintf("%s/%s", namespace, deployment.Name))
	}
	return err
}

func CreateOrUpdateServiceMonitor(ctx context.Context, logger *slog.Logger, client *monitoringclient.Clientset, namespace string, serviceMonitor *monitoringv1.ServiceMonitor) error {
	smClient := client.MonitoringV1().ServiceMonitors(namespace)

	_, err := smClient.Get(ctx, serviceMonitor.Name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		_, err = smClient.Create(ctx, serviceMonitor, metav1.CreateOptions{})
		if err == nil {
			logger.Info("serviceMonitor created", "serviceMonitor", fmt.Sprintf("%s/%s", namespace, serviceMonitor.Name))
		}
		return err
	}

	_, err = smClient.Update(ctx, serviceMonitor, metav1.UpdateOptions{})
	if err == nil {
		logger.Info("serviceMonitor updated", "serviceMonitor", fmt.Sprintf("%s/%s", namespace, serviceMonitor.Name))
	}
	return err
}
