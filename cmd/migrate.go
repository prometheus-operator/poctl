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

	"github.com/prometheus-operator/poctl/internal/etcdmigrate"
	"github.com/prometheus-operator/poctl/internal/k8sutil"
	"github.com/prometheus-operator/poctl/internal/log"
	"github.com/spf13/cobra"
)

// migrateCmd represents the etcd store objects migration command.
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Automatically update Custom Resources to the latest storage version.",
	Long:  `The migrate command in poctl automates the process of updating Kubernetes Custom Resources to the latest storage version. This is essential when upgrading a CRD that supports multiple API versions.`,
	RunE:  runMigration,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func runMigration(cmd *cobra.Command, _ []string) error {
	logger, err := log.NewLogger()
	if err != nil {
		return fmt.Errorf("error while creating logger: %v", err)
	}
	clientSets, err := k8sutil.GetClientSets(kubeconfig)
	if err != nil {
		logger.Error("error while getting client sets", "err", err)
		return err
	}

	if err := etcdmigrate.MigrateCRDs(cmd.Context(), clientSets); err != nil {
		logger.Error("error while updating etcd store", "err", err)
	}

	logger.Info("Prometheus Operator CRD were update in etcd store ")
	return nil
}
