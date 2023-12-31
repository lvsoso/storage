package main

import (
	"errors"
	"time"
)

var (
	globalConfig = map[string]DiskConfig{
		"local": {
			Driver: "local",
			Root:   "./",
			BackendConfig: map[string]string{
				"copy_links": "true",
			},
		},
	}

	ErrConfigNotFound = errors.New("config not found")
)

type (
	DiskConfig struct {
		Driver        string
		Root          string
		URL           string
		BackendConfig map[string]string
		Timeout       time.Duration
	}
)

func AddDiskConfig(name string, config DiskConfig) {
	globalConfig[name] = config
}

func getDiskConfig(name string) (*DiskConfig, error) {
	cfg, ok := globalConfig[name]
	if !ok {
		return nil, ErrConfigNotFound
	}

	return &cfg, nil
}

func (dc DiskConfig) Get(key string) (value string, ok bool) {
	v, ok := dc.BackendConfig[key]
	return v, ok
}
