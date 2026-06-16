package ratelimit

import (
	"context"
	"time"
)

type Phase string

const (
	PhasePostAuth Phase = "post_auth"
	PhasePreAuth  Phase = "pre_auth"
)

type Policy struct {
	Name   string
	Limit  int
	Burst  int
	Window time.Duration
	Phase  Phase
	Scopes []KeyScope
}

type Result struct {
	Allowed    bool
	Limit      int
	Remaining  int
	RetryAfter time.Duration
	ResetAfter time.Duration
}

type Limiter interface {
	Allow(ctx context.Context, key string, policy Policy) (*Result, error)
}
