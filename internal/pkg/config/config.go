package config

import (
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm/logger"
)

type LogConfig struct {
	Level         zapcore.Level   `yaml:"level"`
	DatabaseLevel logger.LogLevel `yaml:"databaseLevel"`
	Enabled       bool            `yaml:"enabled"`
}

type DatabaseConfig struct {
	Name string `yaml:"name"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type Configuration struct {
	TTL        int64 `yaml:"ttl"`
	FPVPort    int   `yaml:"fpvPort"`
	GpsPort    int   `yaml:"gpsPort"`
	ParsePort  int   `yaml:"parsePort"`
	StrikePort int   `yaml:"strikePort"`
}

type JWT struct {
	SigningKey  string `yaml:"singingKey"`
	ExpiresTime string
	BufferTime  string
	Issuer      string
}

type Config struct {
	Log           LogConfig      `yaml:"log"`
	Database      DatabaseConfig `yaml:"database"`
	Server        ServerConfig   `yaml:"server"`
	Configuration Configuration  `yaml:"configuration"`
}

func NewConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("读取配置文件失败: %w", err))
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		panic(fmt.Errorf("解码配置结构体失败: %w", err))
	}

	return &config
}
