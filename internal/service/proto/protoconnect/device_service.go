package protoconnect

import (
	"context"
	"log/slog"
	"time"

	connect "connectrpc.com/connect"
	"github.com/gitRasheed/FleetRPC/internal/device"
	proto "github.com/gitRasheed/FleetRPC/internal/service/proto"
)

type DeviceServiceServer struct {
	pool *device.DevicePool
}

func NewDeviceServiceServer() *DeviceServiceServer {
	pool := device.NewDevicePool("iphone", 10)
	go pool.CleanupExpired()
	return &DeviceServiceServer{pool: pool}
}

func (s *DeviceServiceServer) ReserveDevice(ctx context.Context, req *connect.Request[proto.ReserveRequest]) (*connect.Response[proto.ReserveResponse], error) {
	dev, ok := s.pool.Reserve(req.Msg.User, 2*time.Minute)
	if !ok {
		slog.Info("ReserveDevice failed", "user", req.Msg.User, "reason", "no devices available")
		return connect.NewResponse(&proto.ReserveResponse{
			Status: "no devices available",
		}), nil
	}

	slog.Info("ReserveDevice success", "user", req.Msg.User, "device_id", dev.ID)
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
	slog.Info("ReleaseDevice", "device_id", req.Msg.DeviceId, "status", status)
	return connect.NewResponse(&proto.ReleaseResponse{Status: status}), nil
}

func (s *DeviceServiceServer) WatchDevices(ctx context.Context, req *connect.Request[proto.WatchRequest], stream *connect.ServerStream[proto.DeviceStatus]) error {
	slog.Info("WatchDevices started")
	return nil
}
