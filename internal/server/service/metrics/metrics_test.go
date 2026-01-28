package metrics

import (
	"context"
	"testing"

	"github.com/AA122AA/metring/internal/server/constants"
	"github.com/AA122AA/metring/internal/server/domain"
	"github.com/AA122AA/metring/internal/server/repository"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	v1 := float64(1.25)
	v2 := int64(1)
	cases := []struct {
		name  string
		mName string
		mType string
		value string
		want  *domain.MetricsJSON
		pass  bool
	}{
		{
			name:  "Positive Gauge",
			mName: "Alloc",
			mType: domain.Gauge,
			value: "1.25",
			want: &domain.MetricsJSON{
				ID:    "Alloc",
				MType: domain.Gauge,
				Value: &v1,
			},
			pass: true,
		},
		{
			name:  "Positive Counter",
			mName: "PollCounter",
			mType: domain.Counter,
			value: "1",
			want: &domain.MetricsJSON{
				ID:    "PollCounter",
				MType: domain.Counter,
				Delta: &v2,
			},
			pass: true,
		},
		{
			name:  "Negative Gauge",
			mName: "wrong gauge",
			mType: domain.Gauge,
			value: "a",
			pass:  false,
		},
		{
			name:  "Negative Counter",
			mName: "wrong counter",
			mType: domain.Counter,
			value: "a",
			pass:  false,
		},
		{
			name:  "Wrong metric type",
			mName: "wrong type",
			mType: "lol",
			pass:  false,
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			srv := NewMetrics(context.Background(), repository.NewMockRepo())
			data, err := srv.Parse(tCase.mType, tCase.mName, tCase.value, constants.Update)
			if tCase.pass {
				require.NoError(t, err)
				require.Equal(t, tCase.want, data)
				return
			}
			require.Error(t, err)
		})
	}
}

func TestUpdate(t *testing.T) {
	v := float64(1.25)
	d := int64(1)
	d1 := int64(1)
	ctx := context.Background()
	cases := []struct {
		name   string
		metric *domain.MetricsJSON
		pass   bool
		repo   repository.MetricsRepository
		want   string
	}{
		{
			name: "Positive Gauge",
			metric: &domain.MetricsJSON{
				ID:    repository.Alloc,
				MType: domain.Gauge,
				Value: &v,
			},
			pass: true,
			repo: repository.NewMockRepo(),
			want: "1.25",
		},
		{
			name: "Positive Counter",
			metric: &domain.MetricsJSON{
				ID:    repository.PollCount,
				MType: domain.Counter,
				Delta: &d,
			},
			pass: true,
			repo: repository.NewMemStorage(),
			want: "2",
		},
		{
			name: "Positive Counter 2",
			metric: &domain.MetricsJSON{
				ID:    repository.NoData,
				MType: domain.Counter,
				Delta: &d1,
			},
			repo: repository.NewMemStorage(),
			pass: true,
			want: "1",
		},
		{
			name: "Negative repo",
			metric: &domain.MetricsJSON{
				ID:    repository.Error,
				MType: domain.Counter,
				Delta: &d,
			},
			repo: repository.NewMockRepo(),
			pass: false,
		},
		{
			name: "Negative, wrong type",
			metric: &domain.MetricsJSON{
				ID:    repository.Alloc,
				MType: "lol",
				Delta: &d,
			},
			repo: repository.NewMockRepo(),
			pass: false,
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			m := NewMetrics(ctx, tCase.repo)

			if tCase.metric.MType == repository.Counter && tCase.metric.ID == repository.PollCount {
				d := int64(1)
				tCase.repo.Write(ctx, tCase.metric.ID, &domain.Metrics{
					MType: tCase.metric.MType,
					Delta: &d,
				})
			}

			err := m.Update(ctx, tCase.metric)
			if !tCase.pass {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tCase.want != "" {
				got, err := m.Get(ctx, tCase.metric)
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
			mType: domain.Counter,
			mName: repository.PollCount,
			pass:  true,
			want:  "2",
		},
		{
			name:  "Positive Gauge",
			repo:  repository.NewMockRepo(),
			mType: domain.Gauge,
			mName: repository.Alloc,
			pass:  true,
			want:  "1.25",
		},
		{
			name:  "Negative NoData",
			repo:  repository.NewMockRepo(),
			mType: domain.Gauge,
			mName: repository.NoData,
			pass:  false,
			want:  "err from repo",
		},
		{
			name:  "Negative wrong type",
			repo:  repository.NewMockRepo(),
			mType: domain.Counter,
			mName: repository.Alloc,
			pass:  false,
			want:  "has different types between data and repo",
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			srv := NewMetrics(ctx, tCase.repo)
			data, err := srv.Parse(tCase.mType, tCase.mName, "", constants.Get)
			require.NoError(t, err)
			got, err := srv.Get(ctx, data)
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
