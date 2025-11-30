package agent

import (
	"os"

	"sigs.k8s.io/yaml"
)

type Config struct {
	PollInterval   int    `json:"pollInterval" yaml:"pollInterval" default:"2" validate:"required"`
	Url            string `json:"url" yaml:"url" default:"http://localhost:8080" validate:"required"`
	ReportInterval int    `json:"reportInterval" yaml:"reportInterval" default:"10" validate:"required"`
}

func Read(path string) (*Config, error) {
	if path == "" {
		return &Config{
			PollInterval:   2,
			Url:            "http://localhost:8080",
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
