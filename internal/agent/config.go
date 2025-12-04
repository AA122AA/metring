package agent

import (
	"flag"
	"fmt"
	"os"

	"github.com/AA122AA/metring/internal/flags"
	"sigs.k8s.io/yaml"
)

type Config struct {
	PollInterval   int    `json:"pollInterval" yaml:"pollInterval" default:"2" validate:"required"`
	URL            string `json:"url" yaml:"url" default:"http://localhost:8080" validate:"required"`
	ReportInterval int    `json:"reportInterval" yaml:"reportInterval" default:"10" validate:"required"`
}

func (c *Config) ParseFlags() {
	flag.IntVar(&c.ReportInterval, "r", 2, "poll interval value (seconds)")
	flag.IntVar(&c.PollInterval, "p", 10, "report interval value (seconds)")
	flag.Func("a", "ip:port where server will serve", func(flagArgs string) error {
		return flags.ParseAddr(flagArgs, &c.URL)
	})

	flag.Parse()

	if c.URL == "" {
		fmt.Println("address was not specified, using default - localhost:8080")
		c.URL = "localhost:8080"
	}
}

func Read(path string) (*Config, error) {
	if path == "" {
		return &Config{
			PollInterval:   2,
			URL:            "http://localhost:8080",
			ReportInterval: 10,
		}, nil
	}

	f, err := os.ReadFile(path)
	if err != nil {
		return &Config{}, err
	}

	var config *Config
	if err = yaml.UnmarshalStrict(f, &config); err != nil {
		return &Config{}, err
	}

	return config, nil
}
