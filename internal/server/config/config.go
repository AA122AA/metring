package config

import (
	"flag"
	"log"

	"github.com/AA122AA/metring/internal/flags"
	"github.com/AA122AA/metring/internal/server/service/saver"
	"github.com/caarlos0/env"
)

type Config struct {
	HostAddr     string `json:"hostAddr" yaml:"hostAddr" env:"ADDRESS" default:"localhost:8080"`
	TemplatePath string `json:"templatePath" yaml:"templatePath" env:"TEMPLATE_PATH" default:"internal/server/templates/*.html"`
	DatabaseDSN  string `json:"databaseDSN" yaml:"databaseDSN" env:"DATABASE_DSN"`
	SaverCfg     saver.Config
}

func (c *Config) ParseConfig() {
	flag.Func("a", "ip:port where server will serve", func(flagArgs string) error {
		return flags.ParseAddr(flagArgs, &c.HostAddr)
	})
	flag.StringVar(
		&c.TemplatePath,
		"templates",
		"internal/server/templates/*.html",
		"dir where html templates are stored",
	)
	flag.StringVar(
		&c.DatabaseDSN,
		"d",
		"",
		"string to connect to database",
	)
	flag.IntVar(
		&c.SaverCfg.StoreInterval,
		"i",
		300,
		"store interval - time to save metrics to disk",
	)
	flag.StringVar(
		&c.SaverCfg.FileStoragePath,
		"f",
		"",
		"file where old metrics are stored",
	)
	flag.BoolVar(
		&c.SaverCfg.Restore,
		"r",
		true,
		"should server restore old metrics or not",
	)
	flag.Parse()
}

func (c *Config) LoadEnv() {
	if err := env.Parse(c); err != nil {
		log.Fatalf("error setting config from env: %v", err)
	}
	if err := env.Parse(&c.SaverCfg); err != nil {
		log.Fatalf("error setting saver config from env: %v", err)
	}
}
