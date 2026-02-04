package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type hashResponceWriter struct {
	http.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (hr *hashResponceWriter) Write(p []byte) (int, error) {
	return hr.body.Write(p)
}

func (hr *hashResponceWriter) WriteHeader(statusCode int) {
	hr.status = statusCode
}

func WithHashHeader(lg *zap.Logger, key string) Middleware {
	return func(next http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			originalHeaders := w.Header().Clone()
			hr := &hashResponceWriter{
				ResponseWriter: w,
				body:           bytes.NewBuffer(nil),
				status:         http.StatusOK,
			}

			next.ServeHTTP(hr, r)

			body := hr.body.Bytes()

			if hr.status == http.StatusOK {
				h := hmac.New(sha256.New, []byte(key))
				h.Write(body)
				hash := hex.EncodeToString(h.Sum(nil))
				w.Header()["HashSHA256"] = []string{hash}
			}

			for k, v := range originalHeaders {
				w.Header()[k] = v
			}

			w.WriteHeader(hr.status)
			w.Write(body)
		})
	}
}

func WithHashCheck(lg *zap.Logger, key string) Middleware {
	return func(next http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем, нужно ли вообще что-то делать
			if key == "" {
				lg.Debug("No hash check")
				next.ServeHTTP(w, r)
				return
			}

			// Проверяем, есть ли hash в запросе
			hashFromHeader := r.Header.Get("HashSHA256")
			if hashFromHeader == "" {
				lg.Error("No hash was provided", zap.Any("headers", r.Header.Clone()))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				lg.Error("can not read body", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			hm := hmac.New(sha256.New, []byte(key))
			hm.Write(body)
			hashFromBody := hm.Sum(nil)

			fromHeader, err := hex.DecodeString(hashFromHeader)
			if err != nil {
				lg.Error("can not decode string", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if !hmac.Equal(hashFromBody, fromHeader) {
				lg.Error("hashes are not equal", zap.String("Hash from body", string(hashFromBody)), zap.String("Hash from header", hashFromHeader))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			lg.Debug("hashes are equal")

			// Возвращаем тело в запрос
			r.Body = io.NopCloser(bytes.NewReader(body))

			// передаём управление дальше
			next.ServeHTTP(w, r)
		})
	}
}
