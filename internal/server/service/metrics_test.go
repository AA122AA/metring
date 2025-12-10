package service

import (
	"context"
	"testing"

	models "github.com/AA122AA/metring/internal/server/model"
	"github.com/AA122AA/metring/internal/server/repository"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	v1 := float64(1.25)
	v2 := int64(1)
	cases := []struct {
		name   string
		metric *models.Metrics
		value  string
		want   *models.Metrics
		pass   bool
	}{
		{
			name: "Positive Gauge",
			metric: &models.Metrics{
				MType: models.Gauge,
			},
			value: "1.25",
			want: &models.Metrics{
				MType: models.Gauge,
				Value: &v1,
			},
			pass: true,
		},
		{
			name: "Positive Counter",
			metric: &models.Metrics{
				MType: models.Counter,
			},
			value: "1",
			want: &models.Metrics{
				MType: models.Counter,
				Delta: &v2,
			},
			pass: true,
		},
		{
			name: "Negative Gauge",
			metric: &models.Metrics{
				MType: models.Gauge,
			},
			value: "a",
			pass:  false,
		},
		{
			name: "Negative Counter",
			metric: &models.Metrics{
				MType: models.Counter,
			},
			value: "a",
			pass:  false,
		},
		{
			name: "Wrong metric type",
			metric: &models.Metrics{
				MType: "lol",
			},
			pass: false,
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			err := parse(tCase.metric, tCase.value)
			if tCase.pass {
				require.NoError(t, err)
				require.Equal(t, tCase.want, tCase.metric)
				return
			}
			require.Error(t, err)
		})
	}
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name  string
		mName string
		mType string
		value string
		pass  bool
		repo  repository.MetricsRepository
		want  string
	}{
		{
			name:  "Positive Gauge",
			mName: repository.Alloc,
			mType: models.Gauge,
			value: "1.25",
			pass:  true,
			repo:  repository.NewMockRepo(),
			want:  "1.25",
		},
		{
			name:  "Positive Counter",
			mName: repository.PollCount,
			mType: models.Counter,
			value: "1",
			pass:  true,
			repo:  repository.NewMemStorage(),
			want:  "1",
		},
		{
			name:  "Positive Counter 2",
			mName: repository.NoData,
			mType: models.Counter,
			value: "1",
			repo:  repository.NewMockRepo(),
			pass:  true,
		},
		{
			name:  "Negative repo",
			mName: repository.Error,
			mType: models.Counter,
			value: "1",
			repo:  repository.NewMockRepo(),
			pass:  false,
		},
		{
			name:  "Negative, wrong value",
			mName: repository.Alloc,
			mType: models.Gauge,
			value: "a",
			repo:  repository.NewMockRepo(),
			pass:  false,
		},
		{
			name:  "Negative, wrong type",
			mName: repository.Alloc,
			mType: "lol",
			value: "1",
			repo:  repository.NewMockRepo(),
			pass:  false,
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			if tCase.mName == repository.Counter {
				v := int64(1)
				tCase.repo.Write(tCase.mName, &models.Metrics{
					MType: tCase.mType,
					Delta: &v,
				})
			}

			m := NewMetrics(ctx, tCase.repo)
			err := m.Update(tCase.mName, tCase.mType, tCase.value)
			if !tCase.pass {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tCase.want != "" {
				got, err := m.Get(tCase.mType, tCase.mName)
				require.NoError(t, err)
				require.Equal(t, tCase.want, got)
			}
		})
	}
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name  string
		repo  repository.MetricsRepository
		mType string
		mName string
		pass  bool
		want  string
	}{
		{
			name:  "Positive Counter",
			repo:  repository.NewMockRepo(),
			mType: models.Counter,
			mName: repository.PollCount,
			pass:  true,
			want:  "2",
		},
		{
			name:  "Positive Gauge",
			repo:  repository.NewMockRepo(),
			mType: models.Gauge,
			mName: repository.Alloc,
			pass:  true,
			want:  "1.25",
		},
		{
			name:  "Negative NoData",
			repo:  repository.NewMockRepo(),
			mType: models.Gauge,
			mName: repository.NoData,
			pass:  false,
			want:  "err from repo",
		},
		{
			name:  "Negative wrong type",
			repo:  repository.NewMockRepo(),
			mType: models.Counter,
			mName: repository.Alloc,
			pass:  false,
			want:  "wrong metric type",
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			srv := NewMetrics(ctx, tCase.repo)
			got, err := srv.Get(tCase.mType, tCase.mName)
			if tCase.pass {
				require.NoError(t, err)
				require.Equal(t, tCase.want, got)
				return
			}
			require.Error(t, err)
			require.Contains(t, err.Error(), tCase.want)
		})
	}
}
