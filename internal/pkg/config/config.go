package config

import (
	"fmt"

	"github.com/gofiber/fiber/v2/log"
	"github.com/spf13/viper"
)

type LogConfig struct {
	Level   log.Level `yaml:"level"`
	Enabled bool      `yaml:"enabled"`
}

type DatabaseConfig struct {
	Name string `yaml:"name"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type Configuration struct {
	FpvPort   int `yaml:"fpvPort"`
	GpsPort   int `yaml:"gpsPort"`
	ParsePort int `yaml:"parsePort"`
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
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("读取配置文件失败: %w", err))
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		panic(fmt.Errorf("解码配置结构体失败: %w", err))
	}

	log.Infof("配置文件: %+v", config)

	return &config
}
