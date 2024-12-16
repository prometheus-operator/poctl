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

package cmd

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/prometheus-operator/poctl/internal/analyzers"
	"github.com/prometheus-operator/poctl/internal/k8sutil"
	"github.com/prometheus-operator/poctl/internal/log"
	"github.com/spf13/cobra"
)

type AnalyzeKind string

const (
	ServiceMonitor  AnalyzeKind = "servicemonitor"
	Operator        AnalyzeKind = "operator"
	Prometheus      AnalyzeKind = "prometheus"
	Alertmanager    AnalyzeKind = "alertmanager"
	PrometheusAgent AnalyzeKind = "prometheusagent"
)

type AnalyzeFlags struct {
	Kind      string
	Name      string
	Namespace string
}

var (
	analyzerFlags = AnalyzeFlags{}
	analyzeCmd    = &cobra.Command{
		Use:   "analyze",
		Short: "The analyze command performs an in-depth analysis of Prometheus Operator resources, identifying potential issues and misconfigurations in your monitoring setup. It helps ensure your resources are optimized and error-free.",
		Long:  `The analyze command in poctl is a powerful tool that assesses the health of Prometheus Operator resources in Kubernetes. It detects misconfigurations, issues, and inefficiencies in Prometheus, Alertmanager, and ServiceMonitor resources. By offering actionable insights and recommendations, it helps administrators quickly resolve problems and optimize their monitoring setup for better performance.`,
		RunE:  run,
	}
)

func run(cmd *cobra.Command, _ []string) error {
	if analyzerFlags.Kind == "" {
		return fmt.Errorf("kind is required")
	}

	if analyzerFlags.Name == "" {
		return fmt.Errorf("name is required")
	}

	if analyzerFlags.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	logger, err := log.NewLogger()
	if err != nil {
		return fmt.Errorf("error while creating logger: %v", err)
	}

	slog.SetDefault(logger)

	clientSets, err := k8sutil.GetClientSets(kubeconfig)
	if err != nil {
		return fmt.Errorf("error while getting clientsets: %v", err)
	}

	switch AnalyzeKind(strings.ToLower(analyzerFlags.Kind)) {
	case ServiceMonitor:
		return analyzers.RunServiceMonitorAnalyzer(cmd.Context(), clientSets, analyzerFlags.Name, analyzerFlags.Namespace)
	case Operator:
		return analyzers.RunOperatorAnalyzer(cmd.Context(), clientSets, analyzerFlags.Name, analyzerFlags.Namespace)
	case Prometheus:
		return analyzers.RunPrometheusAnalyzer(cmd.Context(), clientSets, analyzerFlags.Name, analyzerFlags.Namespace)
	case Alertmanager:
		return analyzers.RunAlertmanagerAnalyzer(cmd.Context(), clientSets, analyzerFlags.Name, analyzerFlags.Namespace)
	case PrometheusAgent:
		return analyzers.RunPrometheusAgentAnalyzer(cmd.Context(), clientSets, analyzerFlags.Name, analyzerFlags.Namespace)
	default:
		return fmt.Errorf("kind %s not supported", analyzerFlags.Kind)
	}
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.PersistentFlags().StringVarP(&analyzerFlags.Kind, "kind", "k", "", "The kind of object to analyze. For example, ServiceMonitor")
	analyzeCmd.PersistentFlags().StringVarP(&analyzerFlags.Name, "name", "n", "", "The name of the object to analyze")
	analyzeCmd.PersistentFlags().StringVarP(&analyzerFlags.Namespace, "namespace", "s", "", "The namespace of the object to analyze")
}
