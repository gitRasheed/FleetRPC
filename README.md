# FleetRPC

gRPC device fleet reservation service built with Connect.

## Quick Start

```bash
# Run server
go run ./cmd/server

# Build
go build ./...
```

## CLI Client

```bash
# Reserve a device
go run ./cmd/client reserve --user USER --type iphone

# Release a device
go run ./cmd/client release --device-id iphone-2

# Watch live status (Ctrl+C to stop)
go run ./cmd/client watch
```

## Endpoints

| Endpoint | Description |
|----------|-------------|
| `:8080/devicefleet.v1.DeviceService/*` | Connect RPCs |
| `:8080/metrics` | Prometheus metrics |

## Metrics

```bash
# Bash/Linux/macOS
curl localhost:8080/metrics | grep devicefleet
```

```powershell
# PowerShell/Windows
(curl localhost:8080/metrics).Content | Select-String devicefleet
```

## Proto Generation

```bash
buf generate
```
