package client

import (
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
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

func NewK8sClient(logger *slog.Logger, kubeConfig string) (*kubernetes.Clientset, error) {
	var config *rest.Config = nil
	var err error = nil

	if kubeConfig == "" {
		kubeConfig, err = getKubeConfig()
		if err != nil {
			logger.Error("error while getting kubeconfig", err)
			return nil, err
		}
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)

	if err != nil {
		logger.Error("error while creating k8s client config", err)
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error("error while creating k8s client", err)
		return nil, err
	}

	return clientset, nil
}
