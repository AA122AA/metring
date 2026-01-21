package saver

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/AA122AA/metring/internal/server/domain"
	"github.com/AA122AA/metring/internal/server/repository"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type Metrics []*domain.Metrics

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
		s.restore(ctx)
	}

	if s.StoreInterval == 0 {
		return
	}

	ticker := time.NewTicker(time.Duration(s.StoreInterval) * time.Second)

	for {
		select {
		case <-ctx.Done():
			s.lg.Info("got cancellation, returning")
			return
		case <-ticker.C:
			err := s.store(ctx)
			if err != nil {
				s.lg.Error("error while storing", zap.Error(err))
			}
		}
	}
}

func (s *Saver) WriteSync(data *domain.MetricsJSON) error {
	if s.StoreInterval != 0 {
		return nil
	}

	metrics, empty, err := s.isEmpty()
	if err != nil {
		return err
	}

	metric := domain.TransformFromJSON(data)

	if empty {
		return s.writeNewMetric(metric)
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

func (s *Saver) isEmpty() (Metrics, bool, error) {
	var pathErr *os.PathError
	metrics, err := s.readFromFile()
	if err != nil {
		if errors.Is(err, pathErr) {
			return nil, true, nil
		}
		return nil, true, fmt.Errorf("failed to read from file: %w", err)
	}
	return metrics, false, nil
}

func (s *Saver) writeNewMetric(metric *domain.Metrics) error {
	metrics := make(Metrics, 0, 1)
	metrics[0] = metric

	err := s.writeToFile(metrics)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func (s *Saver) WriteSyncBatch(data []*domain.MetricsJSON) error {
	for _, metric := range data {
		err := s.WriteSync(metric)
		if err != nil {
			return err
		}
	}

	return nil
}

func contains(metrics []*domain.Metrics, metric *domain.Metrics) (int, bool) {
	for i, m := range metrics {
		if m.ID == metric.ID {
			return i, true
		}
	}

	return 0, false
}

func (s *Saver) restore(ctx context.Context) {
	err := s.load(ctx)
	if err != nil {
		s.lg.Error("error while restoring data from file", zap.Error(err))
	}
}

func (s *Saver) store(ctx context.Context) error {
	data, err := s.repo.GetAll(ctx)
	if err != nil {
		// if strings.Contains(err.Error(), "no metrics") {
		var emptyErr *repository.EmptyRepoError
		if errors.Is(err, emptyErr) {
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

func (s *Saver) load(ctx context.Context) error {
	metrics, err := s.readFromFile()
	if err != nil {
		return fmt.Errorf("failed to read from file: %w", err)
	}

	var errs []error
	for _, m := range metrics {
		err := s.repo.Write(ctx, m.ID, m)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	s.lg.Debug("loaded data")

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
