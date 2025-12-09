package config

import (
	"context"
	"flag"

	"github.com/AA122AA/metring/internal/flags"
	"github.com/go-faster/sdk/zctx"
)

type Config struct {
	HostAddr     string `json:"hostAddr" yaml:"hostAddr"`
	TemplatePath string `json:"templatePath" yaml:"templatePath"`
}

func (c *Config) ParseConfig(ctx context.Context) {
	lg := zctx.From(ctx).Named("server config parseConfig")
	flag.Func("a", "ip:port where server will serve", func(flagArgs string) error {
		return flags.ParseAddr(flagArgs, &c.HostAddr)
	})
	flag.String(
		"templates",
		"/home/artem/Documents/development/Yandex.Practicum/metring/internal/server/templates/*.html",
		"dir where html templates are stored",
	)
	flag.Parse()

	if c.HostAddr == "" {
		lg.Debug("address was not specified, using default - localhost:8080")
		c.HostAddr = "localhost:8080"
	}
}
