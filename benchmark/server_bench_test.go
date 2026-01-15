package benchmark

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"connectrpc.com/connect"

	"github.com/gitRasheed/FleetRPC/internal/device"
	proto "github.com/gitRasheed/FleetRPC/internal/service/proto"
	"github.com/gitRasheed/FleetRPC/internal/service/proto/protoconnect"
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})))
}

func BenchmarkServerReserveDevice(b *testing.B) {
	pool := device.NewDevicePool("iphone", b.N)

	mux := http.NewServeMux()
	svc := protoconnect.NewDeviceServiceServerWithPool(pool)
	path, handler := protoconnect.NewDeviceServiceHandler(svc)
	mux.Handle(path, handler)

	server := httptest.NewServer(mux)
	defer server.Close()

	client := protoconnect.NewDeviceServiceClient(http.DefaultClient, server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.ReserveDevice(context.Background(), connect.NewRequest(&proto.ReserveRequest{
			User:       fmt.Sprintf("user%d", i),
			DeviceType: "iphone",
		}))
		if err != nil {
			b.Fatalf("ReserveDevice failed: %v", err)
		}
	}
}
