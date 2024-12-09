# Analyze Command

The analyze command is used to examine Prometheus Operator objects to ensure they comply with specific rules. It checks whether the objects meet certain predefined conditions.

```bash mdox-exec="go run main.go analyze --help" mdox-expect-exit-code=0
The analyze command in poctl is a powerful tool that assesses the health of Prometheus Operator resources in Kubernetes. It detects misconfigurations, issues, and inefficiencies in Prometheus, Alertmanager, and ServiceMonitor resources. By offering actionable insights and recommendations, it helps administrators quickly resolve problems and optimize their monitoring setup for better performance.

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

## Analyze ServiceMonitor

The analyze command can specifically target a ServiceMonitor object within a Kubernetes cluster. Users can specify the namespace and name of the ServiceMonitor to assess its compliance with the predefined rules.

## Rules

The analyze command evaluates objects against a set of rules to determine compliance. These rules are defined in the `analyzer` package and are specifically implemented in the `internal/analyzer/servicemonitor.go` file.

### ServiceMonitor Existence

The ServiceMonitor object must exist in the Kubernetes cluster.

### Selector Presence

The ServiceMonitor object must have a defined selector that selects at least one service.

### Port Matching

Each endpoint within the ServiceMonitor object must have a defined port, and this port should match the port of the service it monitors.

## Analyze Operator

The analyze command can also target the Prometheus Operator deployment within a Kubernetes cluster. Users can specify the namespace and name of the Prometheus Operator to assess its compliance with the predefined rules.

## Rules

The analyze command evaluates objects against a set of rules to determine compliance. These rules are defined in the `analyzer` package and are specifically implemented in the `internal/analyzer/operator.go` file.

### Operator Existence

The Prometheus Operator must be deployed in the Kubernetes cluster, which can be confirmed by checking for the presence of the prometheus-operator deployment in the specified namespace and under the given name.

### RBAC Rules

The Prometheus Operator deployment requires proper RBAC (Role-Based Access Control) rules to function correctly. This means the service account associated with the Prometheus Operator must have permissions aligned with the Prometheus Operator CRDs (Custom Resource Definitions) present in the cluster.

For instance, if the Prometheus Operator is managing only Prometheus instances, the service account should have the necessary permissions to create, update, and delete Prometheus resources, but it should not have permissions to manage other resources like Alertmanager.

## Analyze Prometheus

### Prometheus Existence

The Prometheus must be deployed in the Kubernetes cluster, which can be confirmed by checking for the presence of the Prometheus CRDs (Custom Resource Definitions) in the specified namespace and under the given name.

### Prometheus RBAC Rules

The Prometheus server requires proper RBAC (Role-Based Access Control) rules to function correctly. This means the service account associated with the Prometheus must have permissions aligned with the Prometheus CRDs (Custom Resource Definitions) present in the cluster.

Since Prometheus just reads Objects in the Kubernetes API, it requires the get, list, and watch actions. As Prometheus can also be used to scrape metrics from the Kubernetes apiserver, it also requires access to the /metrics/ endpoint of it. In addition to the rules for Prometheus itself, the Prometheus needs to be able to get configmaps to be able to pull in rule files from configmap objects.

### Prometheus Namespace Selectors and Service Selectors

The Prometheus server requires proper service discovery to be enabled. In order for that we need to ensure that any Namespace Selector defined has a matching existing namespace. The same applies for Service Selectors defined, as if any of then is defined (ServiceMonitor, PodMonitor, ScrapeConfig Probe PrometheusRule) the CRDs (Custom Resource Definitions) needs to exits and properly matched.

## Analyze Alertmanager

### Alertmanager Existence

The Alertmanager must be deployed in the Kubernetes cluster, which can be confirmed by checking for the presence of the Prometheus CRDs (Custom Resource Definitions) in the specified namespace and under the given name.

### Alertmanager Configuration

Alertmanager condifuration needs to be provided, either:

* As a Kubernetes secret provided by the user, that needs to ensure the data is stored in a file called alertmanager.yaml;
* The Operator will provide a default generated Kubernetes secret to use;
* Via the AlertmanagerConfig CRDs (Custom Resource Definitions), that should be matched by a Namespace selector in a given namespace, a ConfigSelector or the ConfigSelector Name.

