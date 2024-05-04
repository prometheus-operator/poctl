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
