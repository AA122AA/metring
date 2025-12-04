// пакеты исполняемых приложений должны называться main
package main

import (
	"github.com/AA122AA/metring/internal/server"
	"github.com/AA122AA/metring/internal/server/config"
)

// функция main вызывается автоматически при запуске приложения
func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

// функция run будет полезна при инициализации зависимостей сервера перед запуском
func run() error {
	cfg := &config.Config{}
	cfg.ParseConfig()

	server := server.NewServer(cfg)

	return server.Run()
}
