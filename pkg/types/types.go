package types

import "time"

type Target struct {
	IP   string
	Port int
}

type Result struct {
	Target Target
	Loss   float64
	Min    time.Duration
	Max    time.Duration
	Avg    time.Duration
}
