package agent

import (
	"flag"

	"github.com/AA122AA/metring/internal/flags"
)

type Config struct {
	PollInterval   int    `json:"pollInterval" yaml:"pollInterval" env:"POLL_INTERVAL" default:"2"`
	URL            string `json:"url" yaml:"url" env:"ADDRESS" default:"http://localhost:8080"`
	ReportInterval int    `json:"reportInterval" yaml:"reportInterval" env:"REPORT_INTERVAL" default:"10"`
	Key            string `json:"key" yaml:"key" env:"KEY"`
}

func (c *Config) ParseFlags() {
	flag.IntVar(&c.ReportInterval, "r", 10, "report interval value (seconds)")
	flag.IntVar(&c.PollInterval, "p", 2, "poll interval value (seconds)")
	flag.Func("a", "ip:port where server will serve", func(flagArgs string) error {
		return flags.ParseAddr(flagArgs, &c.URL)
	})
	flag.StringVar(&c.Key, "k", "", "key for sha256 hash algo")

	flag.Parse()
}

// func Read(path string) (*Config, error) {
// 	if path == "" {
// 		return &Config{}, nil
// 	}

// 	f, err := os.ReadFile(path)
// 	if err != nil {
// 		return &Config{}, err
// 	}

// 	var config *Config
// 	if err = yaml.UnmarshalStrict(f, &config); err != nil {
// 		return &Config{}, err
// 	}

// 	return config, nil
// }
