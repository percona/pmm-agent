package common

import "time"

type NetworkInformation struct {
	ClockDrift time.Duration
	Ping       time.Duration
	Roundtrip  time.Duration
}

