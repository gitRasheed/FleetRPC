package device

import (
	"fmt"
	"sync"
	"time"
)

type DevicePool struct {
	mu      sync.RWMutex
	devices []*Device
}

func NewDevicePool(deviceType string, count int) *DevicePool {
	pool := &DevicePool{}
	for i := 0; i < count; i++ {
		pool.devices = append(pool.devices, &Device{
			ID:   fmt.Sprintf("%s-%d", deviceType, i),
			Type: deviceType,
		})
	}
	return pool
}

func (p *DevicePool) Reserve(user string, ttl time.Duration) (*Device, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, d := range p.devices {
		if IsAvailable(d) {
			now := time.Now()
			d.ReservedBy = user
			d.ReservedAt = now
			d.ExpiresAt = now.Add(ttl)
			return d, true
		}
	}
	return nil, false
}

func (p *DevicePool) Release(deviceID string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, d := range p.devices {
		if d.ID == deviceID && !IsAvailable(d) {
			d.ReservedBy = ""
			return true
		}
	}
	return false
}

func (p *DevicePool) CleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		p.mu.Lock()
		now := time.Now()
		for _, d := range p.devices {
			if d.ReservedBy != "" && now.After(d.ExpiresAt) {
				d.ReservedBy = ""
			}
		}
		p.mu.Unlock()
	}
}

func (p *DevicePool) All() []*Device {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*Device, len(p.devices))
	copy(result, p.devices)
	return result
}
