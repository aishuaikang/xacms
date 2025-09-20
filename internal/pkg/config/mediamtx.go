package config

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	MediaMtxAppConfig *MediaMtxConfig
)

type MediaMtxConfig struct {
	Paths map[string]struct {
		Source string `yaml:"source"`
	} `yaml:"paths"`
}

func InitMediaMtxConfig() *MediaMtxConfig {
	viper.SetConfigName("mediamtx")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("读取配置文件失败: %w", err))
	}

	var config MediaMtxConfig
	if err := viper.Unmarshal(&config); err != nil {
		panic(fmt.Errorf("解码配置结构体失败: %w", err))
	}

	MediaMtxAppConfig = &config

	viper.WatchConfig() // 开启监听
	viper.OnConfigChange(func(e fsnotify.Event) {
		var newMediaMtxConfig MediaMtxConfig
		if err := viper.Unmarshal(&newMediaMtxConfig); err != nil {
			panic(fmt.Errorf("解码配置结构体失败: %w", err))
		}

		// 更新全局结构体
		MediaMtxAppConfig = &newMediaMtxConfig
	})

	return &config
}
