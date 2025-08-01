# Format

```bash mdox-exec="go run main.go fmt --help" mdox-expect-exit-code=0
The format command in poctl formats PromQL expressions found in YAML manifests using the upstream Prometheus formatter.

Usage:
  poctl fmt {file | directory} [flags]

Flags:
  -h, --help   help for fmt

Global Flags:
      --kubeconfig string   path to the kubeconfig file, defaults to $KUBECONFIG
      --log-format string   Log format (default "text")
      --log-level string    Log level (default "DEBUG")
```
