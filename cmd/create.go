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
	"github.com/spf13/cobra"
)

// createCmd represents the create command.
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "The create command generates Prometheus Operator resources like Prometheus, Alertmanager, and ServiceMonitor, simplifying the setup of monitoring configurations in Kubernetes.",
	Long:  `The create command in poctl streamlines the process of creating Prometheus Operator resources in Kubernetes. It allows users to easily generate configurations for key components such as Prometheus, Alertmanager, and ServiceMonitor. This tool reduces the complexity of manual configuration by automating resource creation, ensuring proper setups while saving time. Ideal for both new deployments and updates, the create command helps administrators efficiently establish or modify their monitoring infrastructure.`,
	// Run: func(cmd *cobra.Command, args []string) { },
}

const LatestVersion = "0.77.2"

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.PersistentFlags().String("version", LatestVersion, "Prometheus Operator version")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
