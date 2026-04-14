# grpc-health-proxy

A lightweight sidecar that exposes gRPC health checks as HTTP endpoints for Kubernetes readiness probes.

---

## Installation

```bash
go install github.com/yourorg/grpc-health-proxy@latest
```

Or pull the Docker image:

```bash
docker pull ghcr.io/yourorg/grpc-health-proxy:latest
```

---

## Usage

Run the proxy alongside your gRPC service:

```bash
grpc-health-proxy \
  --grpc-addr=localhost:50051 \
  --http-port=8086 \
  --service=myapp.v1.MyService
```

The proxy will forward HTTP GET requests to `/healthz` and translate the response from the gRPC [Health Checking Protocol](https://grpc.io/docs/guides/health-checking/).

### Kubernetes Readiness Probe

```yaml
readinessProbe:
  httpGet:
    path: /healthz
    port: 8086
  initialDelaySeconds: 5
  periodSeconds: 10
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--grpc-addr` | `localhost:50051` | Address of the upstream gRPC server |
| `--http-port` | `8086` | Port to expose the HTTP health endpoint |
| `--service` | `""` | gRPC service name to check (empty checks overall server health) |
| `--timeout` | `5s` | Timeout for upstream gRPC health requests |
| `--tls` | `false` | Enable TLS when connecting to the gRPC server |

---

## Contributing

Issues and pull requests are welcome. Please open an issue first for significant changes.

---

## License

[MIT](LICENSE)