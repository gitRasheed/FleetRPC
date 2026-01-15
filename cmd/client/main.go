package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"connectrpc.com/connect"

	proto "github.com/gitRasheed/FleetRPC/internal/service/proto"
	"github.com/gitRasheed/FleetRPC/internal/service/proto/protoconnect"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	baseURL := "http://localhost:8080"
	client := protoconnect.NewDeviceServiceClient(http.DefaultClient, baseURL)

	switch os.Args[1] {
	case "reserve":
		handleReserve(client, os.Args[2:])
	case "release":
		handleRelease(client, os.Args[2:])
	case "watch":
		handleWatch(client)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/client/main.go reserve --user USER --type TYPE")
	fmt.Println("  go run cmd/client/main.go release --device-id ID")
	fmt.Println("  go run cmd/client/main.go watch")
}

func handleReserve(client protoconnect.DeviceServiceClient, args []string) {
	fs := flag.NewFlagSet("reserve", flag.ExitOnError)
	user := fs.String("user", "", "user name")
	deviceType := fs.String("type", "iphone", "device type")
	fs.Parse(args)

	if *user == "" {
		fmt.Println("error: --user is required")
		os.Exit(1)
	}

	resp, err := client.ReserveDevice(context.Background(), connect.NewRequest(&proto.ReserveRequest{
		User:       *user,
		DeviceType: *deviceType,
	}))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	if resp.Msg.DeviceId == "" {
		fmt.Printf("failed: %s\n", resp.Msg.Status)
	} else {
		fmt.Printf("reserved: %s\n", resp.Msg.DeviceId)
	}
}

func handleRelease(client protoconnect.DeviceServiceClient, args []string) {
	fs := flag.NewFlagSet("release", flag.ExitOnError)
	deviceID := fs.String("device-id", "", "device ID to release")
	fs.Parse(args)

	if *deviceID == "" {
		fmt.Println("error: --device-id is required")
		os.Exit(1)
	}

	resp, err := client.ReleaseDevice(context.Background(), connect.NewRequest(&proto.ReleaseRequest{
		DeviceId: *deviceID,
	}))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("released: %s (%s)\n", *deviceID, resp.Msg.Status)
}

func handleWatch(client protoconnect.DeviceServiceClient) {
	fmt.Println("watching devices (ctrl+c to stop)")

	stream, err := client.WatchDevices(context.Background(), connect.NewRequest(&proto.WatchRequest{}))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	for stream.Receive() {
		dev := stream.Msg()
		status := "available"
		if !dev.Available {
			status = fmt.Sprintf("reserved by %s", dev.ReservedBy)
		}
		fmt.Printf("%s: %s\n", dev.DeviceId, status)
	}

	if err := stream.Err(); err != nil {
		fmt.Printf("stream error: %v\n", err)
	}
}
