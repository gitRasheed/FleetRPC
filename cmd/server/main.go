package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gitRasheed/FleetRPC/internal/service/proto/protoconnect"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	svc := protoconnect.NewDeviceServiceServer()
	path, handler := protoconnect.NewDeviceServiceHandler(svc)

	mux := http.NewServeMux()
	mux.Handle(path, handler)
	mux.Handle("/metrics", promhttp.Handler())

	slog.Info("Starting FleetRPC server", "grpc_path", path, "metrics", "/metrics", "port", ":8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		slog.Error("Server failed", "err", err)
		os.Exit(1)
	}
}
