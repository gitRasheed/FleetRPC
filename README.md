# FleetRPC

gRPC device fleet reservation service built with Connect.

## Quick Start

```bash
# Run server
go run ./cmd/server

# Build
go build ./...

# Race detection
go build -race ./...
```

## Endpoints

| Endpoint | Description |
|----------|-------------|
| `:8080/devicefleet.v1.DeviceService/*` | Connect RPCs |
| `:8080/metrics` | Prometheus metrics |

## Metrics

```bash
# View metrics
curl localhost:8080/metrics | grep devicefleet

# Live watch
watch -n 1 'curl -s localhost:8080/metrics | grep devicefleet'
```

## Proto Generation

```bash
buf generate
```

## RPCs

- `ReserveDevice(user, device_type)` → reserves an available device
- `ReleaseDevice(device_id)` → releases a reserved device
- `WatchDevices()` → streams device status every 1s
