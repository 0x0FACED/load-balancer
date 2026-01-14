package config

import (
	"encoding/json"
	"os"

	"github.com/0x0FACED/load-balancer/internal/balancer"
	"github.com/0x0FACED/load-balancer/internal/limiter"
)

type AppConfig struct {
	Balancer    balancer.Config `json:"balancer"`
	RateLimiter limiter.Config  `json:"rate_limiter"`
	Logger      LoggerConfig    `json:"logger"`
	Server      ServerConfig    `json:"server"`
	Database    DatabaseConfig  `json:"database"`
	Redis       RedisConfig     `json:"redis"`
}

type ServerConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
	IdleTimeout  int    `json:"idle_timeout"`
}

type DatabaseConfig struct {
	DSN string `json:"dsn"`
}

type RedisConfig struct {
	Address      string `json:"address"`
	Password     string `json:"password"`
	DB           int    `json:"db"`
	PoolSize     int    `json:"pool_size"`
	DialTimeout  int    `json:"dial_timeout"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
}

type LoggerConfig struct {
	Level   string `json:"level"`
	LogsDir string `json:"logs_dir"`
}

func Load() (*AppConfig, error) {
	var cfg AppConfig

	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "./config/config.json"
	}

	cfgFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer cfgFile.Close()

	if err := json.NewDecoder(cfgFile).Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
