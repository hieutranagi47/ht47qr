package domain

import (
	"fmt"
	"time"
)

type Plan string

const (
	Free     Plan = "free"
	Pro      Plan = "pro"
	Ultimate Plan = "ultimate"
)

type PlanLimits struct {
	Lifetime time.Duration
	MaxLive  int
}

func (p Plan) Limits() (PlanLimits, error) {
	switch p {
	case Free:
		return PlanLimits{Lifetime: 30 * 24 * time.Hour, MaxLive: 5}, nil
	case Pro:
		return PlanLimits{Lifetime: 6 * 30 * 24 * time.Hour, MaxLive: 100}, nil
	case Ultimate:
		return PlanLimits{Lifetime: 365 * 24 * time.Hour, MaxLive: 3000}, nil
	default:
		return PlanLimits{}, fmt.Errorf("unknown plan %q", p)
	}
}
