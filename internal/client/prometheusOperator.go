package client

import (
	"log/slog"

	opClientv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewPrometheusOperatorV1(logger *slog.Logger, kubeConfig string) (*opClientv1.MonitoringV1Client, error) {
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
		logger.Error("error while creating Prometheus Operator client config", err)
		return nil, err
	}

	opClientv1, err := opClientv1.NewForConfig(config)
	if err != nil {
		logger.Error("error while creating Prometheus Operator client", err)
		return nil, err
	}

	return opClientv1, nil
}
