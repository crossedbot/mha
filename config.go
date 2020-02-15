package main

import (
	"github.com/crossedbot/common/golang/config"
)

type ServerConfig struct {
	Host         string `toml:"host"`
	Port         int    `toml:"port"`
	ReadTimeout  int    `toml:"read_timeout"`  // in seconds
	WriteTimeout int    `toml:"write_timeout"` // in seconds
}

type LoggingConfig struct {
	File string `toml:"file"`
}

type Config struct {
	Server  ServerConfig  `toml:"server"`
	Logging LoggingConfig `toml:"logging"`
}

func Load(path string, c *Config) error {
	config.Path(path)
	return config.Load(c)
}
