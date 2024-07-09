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

	"github.com/google/go-github/v62/github"
	"github.com/prometheus-operator/poctl/internal/create"
	"github.com/prometheus-operator/poctl/internal/k8sutil"
	"github.com/prometheus-operator/poctl/internal/log"
	"github.com/spf13/cobra"
)

var (
	stackCmd = &cobra.Command{
		Use:   "stack",
		Short: "create a stack of Prometheus Operator resources.",
		Long:  `create a stack of Prometheus Operator resources.`,
		RunE:  runStack,
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

func runStack(cmd *cobra.Command, _ []string) error {
	logger, err := log.NewLogger()
	if err != nil {
		return fmt.Errorf("error while creating logger: %v", err)
	}

	version, err := cmd.Flags().GetString("version")
	if err != nil {
		logger.Error("error while getting version flag", "error", err)
		return err
	}

	logger.Info(version)

	clientSets, err := k8sutil.GetClientSets(kubeconfig)
	if err != nil {
		logger.Error("error while getting client sets", "err", err)
		return err
	}

	gitHubClient := github.NewClient(nil)

	if err := create.RunCreateStack(context.Background(), logger, clientSets, gitHubClient, version); err != nil {
		logger.Error("error while creating Prometheus Operator stack", "err", err)
	}

	logger.Info("Prometheus Operator stack created successfully.")
	return nil
}
