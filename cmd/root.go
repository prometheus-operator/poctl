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
	"os"

	"github.com/prometheus-operator/poctl/internal/log"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "poctl",
	Short: "Command Line Interface (CLI) designed specifically for managing Prometheus Operator resources. It streamlines the processes of deploying, troubleshooting, and validating your monitoring infrastructure within Kubernetes environments.",
	Long: `Command Line Interface (CLI) designed specifically for managing Prometheus Operator resources. It streamlines the processes of deploying, troubleshooting, and validating your monitoring infrastructure within Kubernetes environments.

By providing an intuitive interface, poctl allows users to efficiently manage key resources like Prometheus instances, Alertmanager configurations, and ServiceMonitors. It simplifies complex tasks, reduces manual configuration errors, and offers built-in tools for validation and troubleshooting, helping you maintain a healthy and scalable monitoring setup.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var kubeconfig string

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.poctl.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", os.Getenv("KUBECONFIG"), "path to the kubeconfig file, defaults to $KUBECONFIG")
	log.RegisterFlags(rootCmd.PersistentFlags())
}
