package protoconnect

import (
	"context"
	"log/slog"

	connect "connectrpc.com/connect"
	proto "github.com/gitRasheed/FleetRPC/internal/service/proto"
)

// DeviceServiceServer implements the DeviceServiceHandler interface.
type DeviceServiceServer struct{}

// NewDeviceServiceServer creates a new DeviceServiceServer.
func NewDeviceServiceServer() *DeviceServiceServer {
	return &DeviceServiceServer{}
}

// ReserveDevice handles device reservation requests.
func (s *DeviceServiceServer) ReserveDevice(ctx context.Context, req *connect.Request[proto.ReserveRequest]) (*connect.Response[proto.ReserveResponse], error) {
	slog.Info("ReserveDevice called", "user", req.Msg.User)
	return connect.NewResponse(&proto.ReserveResponse{
		DeviceId: "",
		Status:   "stubbed",
	}), nil
}

// ReleaseDevice handles device release requests.
func (s *DeviceServiceServer) ReleaseDevice(ctx context.Context, req *connect.Request[proto.ReleaseRequest]) (*connect.Response[proto.ReleaseResponse], error) {
	slog.Info("ReleaseDevice called", "device_id", req.Msg.DeviceId)
	return connect.NewResponse(&proto.ReleaseResponse{Status: "stubbed"}), nil
}

// WatchDevices streams device status updates to clients.
func (s *DeviceServiceServer) WatchDevices(ctx context.Context, req *connect.Request[proto.WatchRequest], stream *connect.ServerStream[proto.DeviceStatus]) error {
	slog.Info("WatchDevices started")
	return nil
}
