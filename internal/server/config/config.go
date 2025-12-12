package config

import (
	"flag"

	"github.com/AA122AA/metring/internal/flags"
)

type Config struct {
	HostAddr     string `json:"hostAddr" yaml:"hostAddr" env:"ADDRESS" default:"localhost:8080"`
	TemplatePath string `json:"templatePath" yaml:"templatePath" env:"TEMPLATE_PATH" default:"internal/server/templates/*.html"`
}

func (c *Config) ParseConfig() {
	flag.Func("a", "ip:port where server will serve", func(flagArgs string) error {
		return flags.ParseAddr(flagArgs, &c.HostAddr)
	})
	flag.String(
		"templates",
		"internal/server/templates/*.html",
		"dir where html templates are stored",
	)
	flag.Parse()
}
