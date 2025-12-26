package saver

import (
	"bufio"
	"context"
	"encoding/json"
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

func (s *Saver) load() error {
	file, err := os.Open(s.FileStoragePath)
	if err != nil {
		return fmt.Errorf("can not load file: %w", err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := scanner.Text()
		if data == "[" || data == "]" {
			continue
		}
		data, _ = strings.CutSuffix(data, ",")

		var metric models.Metrics
		err := json.Unmarshal([]byte(data), &metric)
		if err != nil {
			return fmt.Errorf("error while unmarshaling: %w", err)
		}

		err = s.repo.Write(metric.ID, &metric)
		if err != nil {
			return fmt.Errorf("error while writing to repo: %w", err)
		}
	}

	s.lg.Debug("loaded data")

	return nil
}

func (s *Saver) store() error {
	data, err := s.repo.GetAll()
	if err != nil {
		if strings.Contains(err.Error(), "no metrics") {
			return nil
		}
		return fmt.Errorf("err while getting metrics from repo: %w", err)
	}

	file, err := os.OpenFile(s.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		return fmt.Errorf("err while creating file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString("[\n")
	if err != nil {
		return fmt.Errorf("err while writing to file: %w", err)
	}

	i := 0
	for _, v := range data {
		i++
		metric, err := json.Marshal(&v)
		if err != nil {
			return fmt.Errorf("err while encoding data: %w", err)
		}
		if i != len(data) {
			metric = append(metric, ',')
		}
		metric = append(metric, '\n')
		_, err = file.Write(metric)
		if err != nil {
			return fmt.Errorf("err while writing to file: %w", err)
		}
	}

	_, err = file.WriteString("]")
	if err != nil {
		return fmt.Errorf("err while writing to file: %w", err)
	}

	s.lg.Debug("stored in", zap.String("file", s.FileStoragePath))

	return nil
}

func (s *Saver) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Duration(s.StoreInterval) * time.Second)
	if s.Restore {
		err := s.load()
		if err != nil {
			s.lg.Error("error while restoring data from file", zap.Error(err))
		}
	}
	err := s.store()
	if err != nil {
		s.lg.Error("error while storing", zap.Error(err))
	}
	for {
		select {
		case <-ctx.Done():
			s.lg.Info("got cancellation, returning")
			return
		case <-ticker.C:
			s.store()
		}
	}
}
