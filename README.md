# FleetRPC

gRPC device fleet reservation service built with Connect.

## Quick Start

```bash
go run ./cmd/server
```

## CLI Client

```bash
go run ./cmd/client reserve --user USER --type iphone
go run ./cmd/client release --device-id iphone-2
go run ./cmd/client watch
```

## Tests

```bash
# Integration tests
go test -v ./test

# With race detection
go test -v -race ./test
```

## Benchmarks

```bash
# All benchmarks
go test -bench . -benchmem ./benchmark

# Server RPC benchmark only
go test -bench BenchmarkServer -benchmem ./benchmark

# With CPU profiling
cd benchmark
go test -run=NONE -bench=BenchmarkServer -benchtime=3s -cpuprofile cpu.prof
go tool pprof -text cpu.prof
```

## Metrics

```bash
# Bash/Linux/macOS
curl localhost:8080/metrics | grep devicefleet
```

```powershell
# PowerShell/Windows
(curl localhost:8080/metrics).Content | Select-String devicefleet
```

## Build

```bash
go build ./...
go build -race ./...
```

## Proto Generation

```bash
buf generate
```

## Endpoints

| Endpoint | Description |
|----------|-------------|
| `:8080/devicefleet.v1.DeviceService/*` | Connect RPCs |
| `:8080/metrics` | Prometheus metrics |
