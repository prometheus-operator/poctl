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
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getKubeConfig() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	kubeConfig := filepath.Clean(fmt.Sprintf("%v/.kube/config", usr.HomeDir))

	if _, err := os.Stat(kubeConfig); err != nil {
		return "", err
	}

	return kubeConfig, nil
}

func GetRestConfig(logger *slog.Logger, kubeConfig string) (*rest.Config, error) {
	var config *rest.Config
	var err error

	if kubeConfig == "" {
		kubeConfig, err = getKubeConfig()
		if err != nil {
			logger.With("error", err.Error()).Error("error while getting kubeconfig")
			return nil, err
		}
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		logger.With("error", err.Error()).Error("error while creating k8s client config")
		return nil, err
	}

	return config, nil
}
