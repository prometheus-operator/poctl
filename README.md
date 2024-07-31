# Poctl

A Command Line Interface for Prometheus-Operator resources to help you manage, troubleshoot and validate your resources.

## Installation

```bash
go install github.com/prometheus-operator/poctl
```

## Usage

```bash mdox-exec="go run main.go --help" mdox-expect-exit-code=2
poctl is a command line interface for managing Prometheus Operator, allowing you to
	create, delete, and manage Prometheus instances, ServiceMonitors, and more.

Usage:
  poctl [command]

Available Commands:
  analyze     Analyzes the given object and runs a set of rules on it to determine if it is compliant with the given rules.
  completion  Generate the autocompletion script for the specified shell
  create      create is used to create Prometheus Operator resources.
  help        Help about any command

Flags:
  -h, --help                help for poctl
      --kubeconfig string   path to the kubeconfig file, defaults to $KUBECONFIG
      --log-format string   Log format (default "text")
      --log-level string    Log level (default "DEBUG")

Use "poctl [command] --help" for more information about a command.
```

## Available Commands

There are a set of commands available in poctl to help you manage, troubleshoot and validate your Prometheus-Operator resources.

### Create

The `create` allows you either to create a monitor objects from a given kubernetes object or deploy a complete Prometheus Stack.

#### Stack

The `poctl create stack` command allows you to deploy a complete Prometheus Stack containing Prometheus, Alertmanager, node-exporter and kube-state-metrics, with the following flags:

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

#### Service Monitor

The `poctl create servicemonitor` command allows you to create a ServiceMonitor object from a given kubernetes object, withe the following flags:

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

### Analyze

The analyze command allows you to run a set of rules on a given object to determine if it is compliant with the given rules.

```bash mdox-exec="go run main.go analyze --help" mdox-expect-exit-code=2
Analyzes the given object and runs a set of rules on it to determine if it is compliant with the given rules.
		For example:
			- Analyze if the service monitor is selecting any service or using the correct service port.

Usage:
  poctl analyze [flags]

Flags:
  -h, --help               help for analyze
  -k, --kind string        The kind of object to analyze. For example, ServiceMonitor
  -n, --name string        The name of the object to analyze
  -s, --namespace string   The namespace of the object to analyze

Global Flags:
      --kubeconfig string   path to the kubeconfig file, defaults to $KUBECONFIG
      --log-format string   Log format (default "text")
      --log-level string    Log level (default "DEBUG")
```
