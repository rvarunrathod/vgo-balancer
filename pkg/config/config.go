package config

import (
	"os"
	"fmt"

	"gopkg.in/yaml.v2"
)

func NewConfig(configPath string) (*VgoBalancer, error) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config VgoBalancer
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("invalid service config in %s: %w", configPath, err)
	}

	return &config, err
}