package device

import "time"

type Device struct {
	ID         string
	Type       string
	ReservedBy string
	ReservedAt time.Time
	ExpiresAt  time.Time
}

func IsAvailable(d *Device) bool {
	return d.ReservedBy == "" || time.Now().After(d.ExpiresAt)
}
