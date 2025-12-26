package saver

import (
	"context"
	"os"
	"testing"

	"github.com/AA122AA/metring/internal/server/repository"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	ctx := context.Background()
	dir := os.TempDir()
	file, err := os.CreateTemp(dir, "metrics.json")
	require.NoError(t, err)
	cfg := Config{
		StoreInterval:   10,
		FileStoragePath: file.Name(),
	}
	repo := repository.NewMemStorage()

	saver := NewSaver(ctx, cfg, repo)
	err = saver.load()
	require.NoError(t, err)
}
