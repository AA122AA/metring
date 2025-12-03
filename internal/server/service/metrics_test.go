package service

import (
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
			mName: "gauge",
			mType: models.Gauge,
			value: "1.25",
			pass:  true,
			repo:  repository.NewMockRepo(),
			want:  "1.25",
		},
		{
			name:  "Positive Counter",
			mName: "counter",
			mType: models.Counter,
			value: "1",
			pass:  true,
			repo:  repository.NewMemStorage(),
			want:  "2",
		},
		{
			name:  "Positive Counter 2",
			mName: "data",
			mType: models.Counter,
			value: "1",
			repo:  repository.NewMockRepo(),
			pass:  true,
		},
		{
			name:  "Negative repo",
			mName: "error",
			mType: models.Counter,
			value: "1",
			repo:  repository.NewMockRepo(),
			pass:  false,
		},
		{
			name:  "Negative, wrong value",
			mName: "Alloc",
			mType: models.Gauge,
			value: "a",
			repo:  repository.NewMockRepo(),
			pass:  false,
		},
		{
			name:  "Negative, wrong type",
			mName: "Alloc",
			mType: "lol",
			value: "1",
			repo:  repository.NewMockRepo(),
			pass:  false,
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			if tCase.mName == "counter" {
				v := int64(1)
				tCase.repo.Write(tCase.mName, &models.Metrics{
					MType: tCase.mType,
					Delta: &v,
				})
			}

			m := NewMetrics(tCase.repo)
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
