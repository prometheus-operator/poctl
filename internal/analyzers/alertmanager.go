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

func RunAlertmanagerAnalyzer(ctx context.Context, clientSets *k8sutil.ClientSets, name, namespace string) error {
	alertmanager, err := clientSets.MClient.MonitoringV1().Alertmanagers(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("alertmanager %s not found in namespace %s", name, namespace)
		}
		return fmt.Errorf("error while getting Alertmanager: %v", err)
	}

	_, err = clientSets.KClient.CoreV1().ServiceAccounts(namespace).Get(ctx, alertmanager.Spec.ServiceAccountName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to list ServiceAcounts: %w", err)
	}

	if alertmanager.Spec.AlertmanagerConfigSelector == nil && alertmanager.Spec.AlertmanagerConfiguration == nil {
		if alertmanager.Spec.ConfigSecret != "" {
			// use provided config secret
			if err := checkAlertmanagerSecret(ctx, clientSets, alertmanager.Spec.ConfigSecret, namespace, "alertmanager.yaml"); err != nil {
				return fmt.Errorf("error checking Alertmanager secret: %w", err)
			}
		} else if alertmanager.Spec.ConfigSecret == "" {
			// use the default generated secret from pkg/alertmanager/statefulset.go
			amConfigSecretName := fmt.Sprintf("alertmanager-%s-generated", alertmanager.Name)
			if err := checkAlertmanagerSecret(ctx, clientSets, amConfigSecretName, namespace, "alertmanager.yaml.gz"); err != nil {
				return fmt.Errorf("error checking Alertmanager secret: %w", err)
			}
		}
	}
	// If 'AlertmanagerConfigNamespaceSelector' is nil, only check own namespace.
	if alertmanager.Spec.AlertmanagerConfigNamespaceSelector != nil {
		if err := k8sutil.CheckResourceNamespaceSelectors(ctx, *clientSets, alertmanager.Spec.AlertmanagerConfigNamespaceSelector); err != nil {
			return fmt.Errorf("AlertmanagerConfigNamespaceSelector is not properly defined: %s", err)
		}
	} //else if alertmanager.Spec.AlertmanagerConfigNamespaceSelector == nil {

	//}

	slog.Info("Alertmanager is compliant, no issues found", "name", name, "namespace", namespace)
	return nil
}

func checkAlertmanagerSecret(ctx context.Context, clientSets *k8sutil.ClientSets, secretName, namespace string, secretData string) error {
	alertmanagerSecret, err := clientSets.KClient.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get alertmanager secret %s %v", secretName, err)
	}
	if len(alertmanagerSecret.Data) == 0 {
		return fmt.Errorf("alertmanager Secret %s is empty", secretName)
	}
	_, found := alertmanagerSecret.Data[secretData]
	if !found {
		return fmt.Errorf("the %s key not found in Secret %s", secretData, secretName)
	}
	return nil
}
