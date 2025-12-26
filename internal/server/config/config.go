package config

import (
	"flag"
	"fmt"

	"github.com/AA122AA/metring/internal/flags"
	"github.com/AA122AA/metring/internal/server/service/saver"
)

type Config struct {
	HostAddr     string `json:"hostAddr" yaml:"hostAddr" env:"ADDRESS" default:"localhost:8080"`
	TemplatePath string `json:"templatePath" yaml:"templatePath" env:"TEMPLATE_PATH" default:"internal/server/templates/*.html"`
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
	flag.IntVar(
		&c.SaverCfg.StoreInterval,
		"i",
		300,
		"store interval - time to save metrics to disk",
	)
	flag.StringVar(
		&c.SaverCfg.FileStoragePath,
		"f",
		"data/metrics.json",
		"file where old metrics are stored",
	)
	flag.BoolVar(
		&c.SaverCfg.Restore,
		"r",
		false,
		"should server restore old metrics or not",
	)
	fmt.Printf("file storage in Parseconfig: %v\n", c.SaverCfg.FileStoragePath)
	flag.Parse()
}
