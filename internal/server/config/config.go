package config

import (
	"flag"
	"fmt"

	"github.com/AA122AA/metring/internal/flags"
)

type Config struct {
	HostAddr string
}

func (c *Config) ParseConfig() {
	flag.Func("a", "ip:port where server will serve", func(flagArgs string) error {
		return flags.ParseAddr(flagArgs, &c.HostAddr)
	})
	flag.Parse()

	if c.HostAddr == "" {
		fmt.Println("address was not specified, using default - localhost:8080")
		c.HostAddr = "localhost:8080"
	}
}
