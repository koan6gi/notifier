package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Request RequestConfig `mapstructure:"request"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type RequestConfig struct {
	Host        string        `mapstructure:"host"`
	Path        string        `mapstructure:"path"`
	PointFromID int           `mapstructure:"point_from_id"`
	PointToID   int           `mapstructure:"point_to_id"`
	Date        string        `mapstructure:"date"`
	DirectionID int           `mapstructure:"direction_id"`
	Interval    time.Duration `mapstructure:"interval"`
	SignalPath  string        `mapstructure:"signal"`
	StartTime   string        `mapstructure:"start_time"`
	EndTime     string        `mapstructure:"end_time"`
}

func Parse(path string) (*Config, error) {
	viper.SetConfigFile(path)

	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	cfg := &Config{}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return cfg, nil
}
