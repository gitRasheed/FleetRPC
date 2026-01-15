package protoconnect

import (
	"context"
	"log/slog"
	"time"

	connect "connectrpc.com/connect"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/gitRasheed/FleetRPC/internal/device"
	proto "github.com/gitRasheed/FleetRPC/internal/service/proto"
)

var (
	totalReservations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "devicefleet_reservations_total",
		Help: "Total number of reservation attempts",
	}, []string{"status"})

	currentlyAvailable = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "devicefleet_devices_available",
		Help: "Current number of available devices",
	})
)

type DeviceServiceServer struct {
	pool *device.DevicePool
}

func NewDeviceServiceServer() *DeviceServiceServer {
	pool := device.NewDevicePool("iphone", 10)
	go pool.CleanupExpired()
	updateAvailableMetric(pool)
	return &DeviceServiceServer{pool: pool}
}

func NewDeviceServiceServerWithPool(pool *device.DevicePool) *DeviceServiceServer {
	return &DeviceServiceServer{pool: pool}
}

func updateAvailableMetric(pool *device.DevicePool) {
	count := 0
	for _, d := range pool.All() {
		if device.IsAvailable(d) {
			count++
		}
	}
	currentlyAvailable.Set(float64(count))
}

func (s *DeviceServiceServer) ReserveDevice(ctx context.Context, req *connect.Request[proto.ReserveRequest]) (*connect.Response[proto.ReserveResponse], error) {
	deviceType := req.Msg.DeviceType
	if deviceType == "" {
		deviceType = "iphone"
	}

	dev, ok := s.pool.Reserve(req.Msg.User, deviceType, 2*time.Minute)
	if !ok {
		totalReservations.WithLabelValues("failure").Inc()
		slog.Info("ReserveDevice failed", "user", req.Msg.User, "type", deviceType, "reason", "no devices available")
		return connect.NewResponse(&proto.ReserveResponse{
			Status: "no devices available",
		}), nil
	}

	totalReservations.WithLabelValues("success").Inc()
	updateAvailableMetric(s.pool)
	slog.Info("ReserveDevice success", "user", req.Msg.User, "type", deviceType, "device_id", dev.ID)
	return connect.NewResponse(&proto.ReserveResponse{
		DeviceId: dev.ID,
		Status:   "reserved",
	}), nil
}

func (s *DeviceServiceServer) ReleaseDevice(ctx context.Context, req *connect.Request[proto.ReleaseRequest]) (*connect.Response[proto.ReleaseResponse], error) {
	success := s.pool.Release(req.Msg.DeviceId)
	status := "released"
	if !success {
		status = "not found or already available"
	}
	updateAvailableMetric(s.pool)
	slog.Info("ReleaseDevice", "device_id", req.Msg.DeviceId, "status", status)
	return connect.NewResponse(&proto.ReleaseResponse{Status: status}), nil
}

func (s *DeviceServiceServer) WatchDevices(ctx context.Context, req *connect.Request[proto.WatchRequest], stream *connect.ServerStream[proto.DeviceStatus]) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	slog.Info("WatchDevices started", "client", req.Peer().Addr)

	for {
		select {
		case <-ctx.Done():
			slog.Info("WatchDevices ended", "client", req.Peer().Addr, "reason", ctx.Err())
			return nil
		case <-ticker.C:
			for _, dev := range s.pool.All() {
				err := stream.Send(&proto.DeviceStatus{
					DeviceId:   dev.ID,
					ReservedBy: dev.ReservedBy,
					Available:  device.IsAvailable(dev),
				})
				if err != nil {
					slog.Error("WatchDevices stream error", "client", req.Peer().Addr, "err", err)
					return err
				}
			}
		}
	}
}
