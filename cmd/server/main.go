// пакеты исполняемых приложений должны называться main
package main

import (
	"net/http"

	"github.com/AA122AA/metring/internal/server/handler"
	"github.com/AA122AA/metring/internal/server/repository"
	"github.com/AA122AA/metring/internal/server/service"
	"github.com/go-chi/chi/v5"
)

// функция main вызывается автоматически при запуске приложения
func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

// функция run будет полезна при инициализации зависимостей сервера перед запуском
func run() error {
	repo := repository.NewMemStorage()
	srv := service.NewMetrics(repo)
	h := handler.NewMetricsHandler(srv)

	// Question: Нужно ли выносить создание роутера в отдельную функцию, если речь
	// о приложения бОльшего масштаба?
	router := chi.NewRouter()
	router.Get("/", h.All)
	router.Get("/value/{mType}/{mName}", h.Get)
	router.Post("/update/{mType}/{mName}/{value}", h.Update)

	// TODO: add config file
	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	return server.ListenAndServe()
}
