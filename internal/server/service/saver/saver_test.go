package saver

import (
	"context"
	"os"
	"testing"

	"github.com/AA122AA/metring/internal/server/repository"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	testJSON := `
[
    {
        "id": "PauseTotalNs",
        "type": "gauge",
        "value": 0
    },
    {
        "id": "GCSys",
        "type": "gauge",
        "value": 1737440
    }
]`
	ctx := context.Background()
	dir := os.TempDir()
	file, err := os.CreateTemp(dir, "metrics.json")
	require.NoError(t, err)

	_, err = file.WriteString(testJSON)
	require.NoError(t, err)
	defer file.Close()

	cfg := Config{
		StoreInterval:   10,
		FileStoragePath: file.Name(),
	}
	repo := repository.NewMemStorage()

	saver := NewSaver(ctx, cfg, repo)
	err = saver.load()
	require.NoError(t, err)
}
