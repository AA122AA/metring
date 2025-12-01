package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AA122AA/metring/internal/server/repository"
	"github.com/AA122AA/metring/internal/server/service"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	cases := []struct {
		name  string
		mName string
		url   string
		pass  bool
		want  string
	}{
		{
			name:  "Positive",
			mName: "gauge",
			url:   "/get/gauge",
			pass:  true,
			want:  `{"id":"","type":"gauge","value":1.25}`,
		},
		{
			name:  "Negative, not found",
			mName: "data",
			url:   "/get/data",
			pass:  false,
			want:  `Что-то пошло не так`,
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			repo := repository.NewMockRepo()
			srv := service.NewMetrics(repo)
			h := NewMetricsHandler(srv)

			r := httptest.NewRequest(http.MethodGet, tCase.url, nil)
			r.SetPathValue("mName", tCase.mName)
			rec := httptest.NewRecorder()
			h.Get(rec, r)

			res := rec.Result()
			if tCase.want != "" {
				defer res.Body.Close()
				data, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				require.Equal(t, tCase.want, strings.TrimSpace(string(data)))
			}
			if tCase.pass {
				require.Equal(t, res.StatusCode, http.StatusOK)
				return
			}
			require.Equal(t, res.StatusCode, http.StatusInternalServerError)
		})
	}
}

func TestUpdate(t *testing.T) {
	cases := []struct {
		name       string
		url        string
		mName      string
		mType      string
		value      string
		want       string
		statusCode int
		pass       bool
	}{
		{
			name:       "Positive",
			url:        "/update/gauge/gauge/1.25",
			mName:      "gauge",
			mType:      "gauge",
			value:      "1.25",
			statusCode: http.StatusOK,
			pass:       true,
		},
		{
			name:       "Negative, no mName",
			url:        "/update//gauge/1.25",
			mName:      "",
			mType:      "gauge",
			value:      "1.25",
			statusCode: http.StatusNotFound,
			pass:       false,
		},
		{
			name:       "Negative, bad type",
			url:        "/update/gauge/gauge/1.25",
			mName:      "gauge",
			mType:      "lol",
			value:      "1.25",
			statusCode: http.StatusBadRequest,
			pass:       false,
		},
		{
			name:       "Negative, bad value",
			url:        "/update/gauge/gauge/1.25",
			mName:      "gauge",
			mType:      "gauge",
			value:      "a",
			statusCode: http.StatusBadRequest,
			pass:       false,
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			repo := repository.NewMockRepo()
			srv := service.NewMetrics(repo)
			h := NewMetricsHandler(srv)

			r := httptest.NewRequest(http.MethodPost, tCase.url, nil)
			r.SetPathValue("mName", tCase.mName)
			r.SetPathValue("mType", tCase.mType)
			r.SetPathValue("value", tCase.value)
			rec := httptest.NewRecorder()
			h.Update(rec, r)

			res := rec.Result()
			if tCase.want != "" {
				defer res.Body.Close()
				data, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				require.Equal(t, tCase.want, strings.TrimSpace(string(data)))
			}
			if tCase.pass {
				require.Equal(t, tCase.statusCode, res.StatusCode)
				return
			}

			require.Equal(t, tCase.statusCode, res.StatusCode)
		})
	}
}
