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
	RateLimit      int    `json:"rateLimit" yaml:"rateLimit" env:"RATE_LIMIT"`
}

func (c *Config) ParseFlags() {
	flag.IntVar(&c.ReportInterval, "r", 10, "report interval value (seconds)")
	flag.IntVar(&c.PollInterval, "p", 2, "poll interval value (seconds)")
	flag.Func("a", "ip:port where server will serve", func(flagArgs string) error {
		return flags.ParseAddr(flagArgs, &c.URL)
	})
	flag.StringVar(&c.Key, "k", "", "key for sha256 hash algo")
	flag.IntVar(&c.RateLimit, "l", 3, "how many workers can make requiests")

	flag.Parse()
}
