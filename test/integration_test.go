package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/gitRasheed/FleetRPC/internal/device"
	proto "github.com/gitRasheed/FleetRPC/internal/service/proto"
	"github.com/gitRasheed/FleetRPC/internal/service/proto/protoconnect"
)

func setupTestServer(pool *device.DevicePool) (protoconnect.DeviceServiceClient, func()) {
	mux := http.NewServeMux()
	svc := &testServer{pool: pool}
	path, handler := protoconnect.NewDeviceServiceHandler(svc)
	mux.Handle(path, handler)

	server := httptest.NewServer(mux)
	client := protoconnect.NewDeviceServiceClient(http.DefaultClient, server.URL)

	return client, server.Close
}

type testServer struct {
	pool *device.DevicePool
}

func (s *testServer) ReserveDevice(ctx context.Context, req *connect.Request[proto.ReserveRequest]) (*connect.Response[proto.ReserveResponse], error) {
	deviceType := req.Msg.DeviceType
	if deviceType == "" {
		deviceType = "iphone"
	}

	dev, ok := s.pool.Reserve(req.Msg.User, deviceType, 2*time.Minute)
	if !ok {
		return connect.NewResponse(&proto.ReserveResponse{
			Status: "no devices available",
		}), nil
	}

	return connect.NewResponse(&proto.ReserveResponse{
		DeviceId: dev.ID,
		Status:   "reserved",
	}), nil
}

func (s *testServer) ReleaseDevice(ctx context.Context, req *connect.Request[proto.ReleaseRequest]) (*connect.Response[proto.ReleaseResponse], error) {
	success := s.pool.Release(req.Msg.DeviceId)
	status := "released"
	if !success {
		status = "not found or already available"
	}
	return connect.NewResponse(&proto.ReleaseResponse{Status: status}), nil
}

func (s *testServer) WatchDevices(ctx context.Context, req *connect.Request[proto.WatchRequest], stream *connect.ServerStream[proto.DeviceStatus]) error {
	for _, dev := range s.pool.All() {
		err := stream.Send(&proto.DeviceStatus{
			DeviceId:   dev.ID,
			ReservedBy: dev.ReservedBy,
			Available:  device.IsAvailable(dev),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func TestReserveAndRelease(t *testing.T) {
	pool := device.NewDevicePool("iphone", 10)
	client, cleanup := setupTestServer(pool)
	defer cleanup()

	reserveResp, err := client.ReserveDevice(context.Background(), connect.NewRequest(&proto.ReserveRequest{
		User:       "testuser",
		DeviceType: "iphone",
	}))
	if err != nil {
		t.Fatalf("ReserveDevice failed: %v", err)
	}
	if reserveResp.Msg.DeviceId == "" {
		t.Fatalf("expected device ID, got empty string")
	}
	if reserveResp.Msg.Status != "reserved" {
		t.Fatalf("expected status 'reserved', got '%s'", reserveResp.Msg.Status)
	}

	releaseResp, err := client.ReleaseDevice(context.Background(), connect.NewRequest(&proto.ReleaseRequest{
		DeviceId: reserveResp.Msg.DeviceId,
	}))
	if err != nil {
		t.Fatalf("ReleaseDevice failed: %v", err)
	}
	if releaseResp.Msg.Status != "released" {
		t.Fatalf("expected status 'released', got '%s'", releaseResp.Msg.Status)
	}
}

func TestDoubleReleaseFails(t *testing.T) {
	pool := device.NewDevicePool("iphone", 10)
	client, cleanup := setupTestServer(pool)
	defer cleanup()

	reserveResp, err := client.ReserveDevice(context.Background(), connect.NewRequest(&proto.ReserveRequest{
		User:       "testuser",
		DeviceType: "iphone",
	}))
	if err != nil {
		t.Fatalf("ReserveDevice failed: %v", err)
	}

	_, err = client.ReleaseDevice(context.Background(), connect.NewRequest(&proto.ReleaseRequest{
		DeviceId: reserveResp.Msg.DeviceId,
	}))
	if err != nil {
		t.Fatalf("first ReleaseDevice failed: %v", err)
	}

	secondRelease, err := client.ReleaseDevice(context.Background(), connect.NewRequest(&proto.ReleaseRequest{
		DeviceId: reserveResp.Msg.DeviceId,
	}))
	if err != nil {
		t.Fatalf("second ReleaseDevice failed: %v", err)
	}
	if secondRelease.Msg.Status != "not found or already available" {
		t.Fatalf("expected 'not found or already available', got '%s'", secondRelease.Msg.Status)
	}
}

func TestReservationExpiresAfterTTL(t *testing.T) {
	pool := device.NewDevicePool("iphone", 1)
	client, cleanup := setupTestServer(pool)
	defer cleanup()

	pool.Reserve("blocker", "iphone", 100*time.Millisecond)

	secondReserve, err := client.ReserveDevice(context.Background(), connect.NewRequest(&proto.ReserveRequest{
		User:       "waiting",
		DeviceType: "iphone",
	}))
	if err != nil {
		t.Fatalf("ReserveDevice failed: %v", err)
	}
	if secondReserve.Msg.DeviceId != "" {
		t.Fatalf("expected no device available, but got '%s'", secondReserve.Msg.DeviceId)
	}

	time.Sleep(150 * time.Millisecond)

	afterExpiry, err := client.ReserveDevice(context.Background(), connect.NewRequest(&proto.ReserveRequest{
		User:       "afterexpiry",
		DeviceType: "iphone",
	}))
	if err != nil {
		t.Fatalf("ReserveDevice after expiry failed: %v", err)
	}
	if afterExpiry.Msg.DeviceId == "" {
		t.Fatalf("expected device to be available after TTL expiry")
	}
}

func TestNoOverlappingReservations(t *testing.T) {
	pool := device.NewDevicePool("iphone", 10)
	client, cleanup := setupTestServer(pool)
	defer cleanup()

	pool.Reserve("user0", "iphone", 5*time.Minute)
	pool.Reserve("user1", "iphone", 5*time.Minute)
	pool.Reserve("user2", "iphone", 5*time.Minute)
	pool.Reserve("user3", "iphone", 5*time.Minute)
	pool.Reserve("user4", "iphone", 5*time.Minute)
	pool.Reserve("user5", "iphone", 5*time.Minute)
	pool.Reserve("user6", "iphone", 5*time.Minute)
	pool.Reserve("user7", "iphone", 5*time.Minute)
	pool.Reserve("user8", "iphone", 5*time.Minute)
	pool.Reserve("user9", "iphone", 5*time.Minute)

	resp, err := client.ReserveDevice(context.Background(), connect.NewRequest(&proto.ReserveRequest{
		User:       "user10",
		DeviceType: "iphone",
	}))
	if err != nil {
		t.Fatalf("ReserveDevice failed: %v", err)
	}
	if resp.Msg.DeviceId != "" {
		t.Fatalf("expected no devices available, but got '%s'", resp.Msg.DeviceId)
	}
	if resp.Msg.Status != "no devices available" {
		t.Fatalf("expected 'no devices available', got '%s'", resp.Msg.Status)
	}
}

func TestWatchReportsCurrentStatus(t *testing.T) {
	pool := device.NewDevicePool("iphone", 3)
	client, cleanup := setupTestServer(pool)
	defer cleanup()

	pool.Reserve("occupied", "iphone", 5*time.Minute)

	stream, err := client.WatchDevices(context.Background(), connect.NewRequest(&proto.WatchRequest{}))
	if err != nil {
		t.Fatalf("WatchDevices failed: %v", err)
	}

	var devices []*proto.DeviceStatus
	for stream.Receive() {
		devices = append(devices, stream.Msg())
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("stream error: %v", err)
	}

	if len(devices) != 3 {
		t.Fatalf("expected 3 devices, got %d", len(devices))
	}

	reservedCount := 0
	availableCount := 0
	for _, d := range devices {
		if d.Available {
			availableCount++
		} else {
			reservedCount++
		}
	}

	if reservedCount != 1 {
		t.Fatalf("expected 1 reserved device, got %d", reservedCount)
	}
	if availableCount != 2 {
		t.Fatalf("expected 2 available devices, got %d", availableCount)
	}
}
