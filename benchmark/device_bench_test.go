package benchmark

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gitRasheed/FleetRPC/internal/device"
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})))
}

func BenchmarkReserveDevice(b *testing.B) {
	pool := device.NewDevicePool("iphone", b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, ok := pool.Reserve(fmt.Sprintf("user%d", i), "iphone", 1*time.Minute)
		if !ok {
			b.Fatalf("reservation failed at %d", i)
		}
	}
}

func BenchmarkReleaseDevice(b *testing.B) {
	pool := device.NewDevicePool("iphone", b.N)

	for i := 0; i < b.N; i++ {
		pool.Reserve(fmt.Sprintf("user%d", i), "iphone", 5*time.Minute)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ok := pool.Release(fmt.Sprintf("iphone-%d", i))
		if !ok {
			b.Fatalf("release failed at %d", i)
		}
	}
}

func BenchmarkWatchDevices(b *testing.B) {
	pool := device.NewDevicePool("iphone", 100)

	pool.Reserve("user0", "iphone", 5*time.Minute)
	pool.Reserve("user1", "iphone", 5*time.Minute)
	pool.Reserve("user2", "iphone", 5*time.Minute)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		devices := pool.All()
		for _, d := range devices {
			_ = device.IsAvailable(d)
		}
	}
}

func BenchmarkConcurrentReserve(b *testing.B) {
	pool := device.NewDevicePool("iphone", 100)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			pool.Reserve(fmt.Sprintf("user%d", i), "iphone", 100*time.Millisecond)
			i++
		}
	})
}

func BenchmarkConcurrentReserveRelease(b *testing.B) {
	pool := device.NewDevicePool("iphone", 50)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			dev, ok := pool.Reserve(fmt.Sprintf("user%d", i), "iphone", 1*time.Second)
			if ok {
				pool.Release(dev.ID)
			}
			i++
		}
	})
}

func BenchmarkHighContention(b *testing.B) {
	pool := device.NewDevicePool("iphone", 5)
	var wg sync.WaitGroup
	numWorkers := 16
	opsPerWorker := b.N / numWorkers

	b.ResetTimer()

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < opsPerWorker; i++ {
				dev, ok := pool.Reserve(fmt.Sprintf("w%d-u%d", workerID, i), "iphone", 50*time.Millisecond)
				if ok {
					pool.Release(dev.ID)
				}
			}
		}(w)
	}
	wg.Wait()
}
