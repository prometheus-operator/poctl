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
	"strings"

	"github.com/prometheus-operator/poctl/internal/analyzers"
	"github.com/prometheus-operator/poctl/internal/k8sutil"
	"github.com/spf13/cobra"
)

type AnalyzeKind string

const (
	ServiceMonitor AnalyzeKind = "servicemonitor"
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
		Short: "Analyzes the given object and runs a set of rules on it to determine if it is compliant with the given rules.",
		Long: `Analyzes the given object and runs a set of rules on it to determine if it is compliant with the given rules.
		For example:
			- Analyze if the service monitor is selecting any service or using the correct service port.`,
		RunE: run,
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

	clientSets, err := k8sutil.GetClientSets(kubeconfig)
	if err != nil {
		return fmt.Errorf("error while getting clientsets: %v", err)
	}

	switch AnalyzeKind(strings.ToLower(analyzerFlags.Kind)) {
	case ServiceMonitor:
		return analyzers.RunServiceMonitorAnalyzer(cmd.Context(), clientSets, analyzerFlags.Name, analyzerFlags.Namespace)
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
