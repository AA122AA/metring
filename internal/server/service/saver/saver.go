package saver

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	models "github.com/AA122AA/metring/internal/server/model"
	"github.com/AA122AA/metring/internal/server/repository"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type Metrics []*models.Metrics

type Saver struct {
	StoreInterval   int
	FileStoragePath string
	Restore         bool

	lg   *zap.Logger
	repo repository.MetricsRepository
}

func NewSaver(ctx context.Context, cfg Config, repo repository.MetricsRepository) *Saver {
	return &Saver{
		StoreInterval:   cfg.StoreInterval,
		FileStoragePath: cfg.FileStoragePath,
		Restore:         cfg.Restore,
		lg:              zctx.From(ctx).Named("Saver service"),
		repo:            repo,
	}
}

func (s *Saver) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	if s.Restore {
		s.restore()
	}

	if s.StoreInterval == 0 {
		return
	}

	ticker := time.NewTicker(time.Duration(s.StoreInterval) * time.Second)

	// err := s.store()
	// if err != nil {
	// 	s.lg.Error("error while storing", zap.Error(err))
	// }

	for {
		select {
		case <-ctx.Done():
			s.lg.Info("got cancellation, returning")
			return
		case <-ticker.C:
			err := s.store()
			if err != nil {
				s.lg.Error("error while storing", zap.Error(err))
			}
		}
	}
}

func (s *Saver) WriteSync(data *models.MetricsJSON) error {
	if s.StoreInterval != 0 {
		return nil
	}
	metric := models.TransformFromJSON(data)

	var pathErr *os.PathError
	metrics, err := s.readFromFile()
	if err != nil {
		if !errors.As(err, &pathErr) {
			return fmt.Errorf("failed to read from file: %w", err)
		}
		metrics = make(Metrics, 1)
		metrics[0] = metric

		err = s.writeToFile(metrics)
		if err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}

		return nil
	}

	if index, ok := contains(metrics, metric); ok {
		metrics[index].Delta = metric.Delta
		metrics[index].Value = metric.Value
	} else {
		metrics = append(metrics, metric)
	}

	err = s.writeToFile(metrics)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func contains(metrics []*models.Metrics, metric *models.Metrics) (int, bool) {
	for i, m := range metrics {
		if m.ID == metric.ID {
			return i, true
		}
	}
	return 0, false
}

func (s *Saver) restore() {
	err := s.load()
	if err != nil {
		s.lg.Error("error while restoring data from file", zap.Error(err))
	}
}

func (s *Saver) store() error {
	data, err := s.repo.GetAll()
	if err != nil {
		if strings.Contains(err.Error(), "no metrics") {
			return nil
		}
		return fmt.Errorf("err while getting metrics from repo: %w", err)
	}

	metrics := make(Metrics, 0, len(data))
	for _, v := range data {
		metrics = append(metrics, v)
	}

	err = s.writeToFile(metrics)
	if err != nil {
		return fmt.Errorf("err while writing to file: %w", err)
	}

	s.lg.Debug("stored in", zap.String("file", s.FileStoragePath))

	return nil
}

func (s *Saver) load() error {
	metrics, err := s.readFromFile()
	if err != nil {
		return fmt.Errorf("failed to read from file: %w", err)
	}

	for _, m := range metrics {
		s.repo.Write(m.ID, m)
	}

	s.lg.Debug("loaded data")

	return nil
}

func (s *Saver) writeToFile(metrics Metrics) error {
	file, err := os.OpenFile(s.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		return fmt.Errorf("err while creating file: %w", err)
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "    ")
	err = enc.Encode(metrics)
	if err != nil {
		return fmt.Errorf("err while marshalling data: %w", err)
	}
	return nil
}

func (s *Saver) readFromFile() (Metrics, error) {
	file, err := os.Open(s.FileStoragePath)
	if err != nil {
		return nil, fmt.Errorf("can not load file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	data, err := reader.ReadBytes(']')
	if err != nil {
		return nil, fmt.Errorf("failed to read from file: %w", err)
	}

	var metrics Metrics
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshaling: %w", err)
	}

	return metrics, nil
}
