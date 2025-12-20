package agent

import (
	"context"
	"testing"
)

func TestGatherMetrics(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		PollInterval: 2,
	}
	ma := NewMetricAgent(ctx, cfg)
	ma.GatherMetrics()
}
