package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTPPort string `yaml:"http_port" validate:"required"`
}

type validate interface {
	Struct(s interface{}) error
}

func Initialize(v validate) (*Config, error) {
	// Reading app file config
	configFile, err := os.Open("./application.yml")

	if err != nil {
		return nil, fmt.Errorf("Cannot open config file: %w", err)
	}

	var config Config

	// Parse config file
	err = yaml.NewDecoder(configFile).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("Cannot unmarshal config data: %w", err)
	}

	// Validate config file
	if err = v.Struct(config); err != nil {
		return nil, fmt.Errorf("Config file is not valid: %w", err)
	}

	return &config, nil
}
