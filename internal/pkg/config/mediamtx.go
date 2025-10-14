package config

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

//go:embed default_mediamtx.yml
var defaultMediaMtxConfigContent []byte

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
		// 如果配置文件不存在，创建默认配置
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("mediamtx.yml 配置文件不存在，正在创建默认配置文件...")
			if err := createDefaultMediaMtxConfig(); err != nil {
				panic(fmt.Errorf("创建默认 mediamtx.yml 配置文件失败: %w", err))
			}
			fmt.Println("默认 mediamtx.yml 配置文件创建成功: ./config/mediamtx.yml")
			// 重新读取配置
			if err := viper.ReadInConfig(); err != nil {
				panic(fmt.Errorf("读取配置文件失败: %w", err))
			}
		} else {
			panic(fmt.Errorf("读取配置文件失败: %w", err))
		}
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

// createDefaultMediaMtxConfig 创建默认 mediamtx 配置文件
func createDefaultMediaMtxConfig() error {
	// 确保配置目录存在
	if err := os.MkdirAll("./config", 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 将嵌入的默认配置写入文件
	if err := os.WriteFile("./config/mediamtx.yml", defaultMediaMtxConfigContent, 0644); err != nil {
		return fmt.Errorf("写入 mediamtx.yml 配置文件失败: %w", err)
	}

	return nil
}
