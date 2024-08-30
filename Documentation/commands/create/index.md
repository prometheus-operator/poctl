# Create Stack

The create stack command is used to deploy a monitoring stack to a Kubernetes cluster, including the following components:

- [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator)
- [Prometheus](https://github.com/prometheus/prometheus)
- [Kube-State-Metrics](https://github.com/kubernetes/kube-state-metrics)
- [NodeExporter](https://github.com/prometheus/node_exporter)
- [AlertManager](https://github.com/prometheus/alertmanager)

It installs all the required Custom Resource Definitions using the latest available version of the Prometheus Operator.

```bash mdox-exec="go run main.go create stack --help" mdox-expect-exit-code=2
create a stack of Prometheus Operator resources.

Usage:
  poctl create stack [flags]

Flags:
  -h, --help   help for stack

Global Flags:
      --kubeconfig string   path to the kubeconfig file, defaults to $KUBECONFIG
      --log-format string   Log format (default "text")
      --log-level string    Log level (default "DEBUG")
      --version string      Prometheus Operator version (default "0.74.0")
```

# Create ServiceMonitor

The create service monitor command is used to create a ServiceMonitor object in a Kubernetes cluster, targeting an existing Kubernetes Service, users can provide the namespace, service name, and port of the service to create the ServiceMonitor object.

```bash mdox-exec="go run main.go create servicemonitor --help" mdox-expect-exit-code=2
Create a service monitor object based on user input parameters or taking as source of truth a kubernetes service

Usage:
  poctl create servicemonitor [flags]

Flags:
  -h, --help               help for servicemonitor
  -n, --namespace string   Namespace of the service (default "default")
  -p, --port string        Port of the service
  -s, --service string     Service name to create the service monitor from

Global Flags:
      --kubeconfig string   path to the kubeconfig file, defaults to $KUBECONFIG
      --log-format string   Log format (default "text")
      --log-level string    Log level (default "DEBUG")
      --version string      Prometheus Operator version (default "0.74.0")
```
