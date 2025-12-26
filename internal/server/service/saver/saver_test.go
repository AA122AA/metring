package saver

import (
	"context"
	"testing"

	"github.com/AA122AA/metring/internal/server/repository"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		StoreInterval:   10,
		FileStoragePath: "../../../../data/metrics.json",
	}
	repo := repository.NewMemStorage()

	saver := NewSaver(ctx, cfg, repo)
	err := saver.load()
	require.NoError(t, err)
}
