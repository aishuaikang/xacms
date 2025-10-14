package config

import (
	_ "embed"
	"fmt"
	"os"
	"path"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm/logger"
)

var (
	AppConfig *Config

	//go:embed default_config.yaml
	defaultConfigContent []byte
)

type LogConfig struct {
	Level          zapcore.Level   `yaml:"level"`
	DatabaseLevel  logger.LogLevel `yaml:"databaseLevel"`
	DisableConsole bool            `yaml:"disableConsole"`
}

type Database struct {
	Name string `json:"name" yaml:"name"`
}

type ServerConfig struct {
	Port int `json:"port" yaml:"port"`
}

type JWT struct {
	SigningKey  string `mapstructure:"signing-key" json:"signingKey" yaml:"signing-key"`   // jwt签名
	ExpiresTime string `mapstructure:"expires-time" json:"expireTime" yaml:"expires-time"` // 过期时间
}

type Configuration struct {
	PasswordSalt string `yaml:"passwordSalt"`
	// GpsSwitch         int    `yaml:"gpsSwitch"`
	OrientationSwitch int `yaml:"orientationSwitch"`

	// // 流媒体服务器地址
	StreamMediaUrl string `yaml:"streamMediaUrl"`

	FpvTcpServer     string `yaml:"fpvTcpServer"`
	GpsTcpServer     string `yaml:"gpsTcpServer"`
	ParserTcpServer  string `yaml:"parserTcpServer"`
	StrikerTcpServer string `yaml:"strikerTcpServer"`

	AlertsTTL  int64 `yaml:"alertsTTL"`
	PalertsTTL int64 `yaml:"palertsTTL"`

	SimilarThreshold float64 `yaml:"similarThreshold"`
	LockFrequency    int     `yaml:"lockFrequency"`
	LateralStrategy  int     `yaml:"lateralStrategy"`
	// LateralStrategySwitch int     `yaml:"lateralStrategySwitch"`
}

type TCPConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

type UDPConfig struct {
	Enabled        bool   `yaml:"enabled" mapstructure:"enabled"`
	ServerIP       string `yaml:"server_ip" mapstructure:"server_ip"`
	ServerPort     int    `yaml:"server_port" mapstructure:"server_port"`
	ReportInterval int    `yaml:"report_interval" mapstructure:"report_interval"`
}

type ProtocolConfig struct {
	DetectionID string    `yaml:"detectionID"`
	TCP         TCPConfig `json:"tcp" yaml:"tcp"`
	UDP         UDPConfig `json:"udp" yaml:"udp" mapstructure:"udp"`
}
type MqttConfig struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`
	Broker   string `yaml:"broker" mapstructure:"broker"`
	Port     int    `yaml:"port" mapstructure:"port"`
	Username string `yaml:"username" mapstructure:"username"`
	Password string `yaml:"password" mapstructure:"password"`
}

type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
)

// IsDevelopment 是否为开发环境
func (e Environment) IsDevelopment() bool {
	return e == Development
}

// IsProduction 是否为生产环境
func (e Environment) IsProduction() bool {
	return e == Production
}

type Config struct {
	Environment   Environment    `yaml:"environment"` // 新增
	SN            string         `yaml:"sn"`
	Version       string         `yaml:"version"`
	Log           LogConfig      `json:"log" yaml:"log"`
	Database      Database       `json:"db" yaml:"database"`
	Server        ServerConfig   `json:"server" yaml:"server"`
	Jwt           JWT            `json:"jwt" yaml:"jwt"`
	Configuration Configuration  `json:"configuration" yaml:"configuration"`
	Protocol      ProtocolConfig `json:"protocol" yaml:"protocol"`
	Mqtt          MqttConfig     `json:"mqtt" yaml:"mqtt"`
}

func InitConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	// 允许环境变量覆盖
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP") // 环境变量前缀

	// 绑定特定环境变量
	viper.BindEnv("environment", "APP_ENV")

	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，创建默认配置
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("配置文件不存在，正在创建默认配置文件...")
			if err := createDefaultConfig(); err != nil {
				panic(fmt.Errorf("创建默认配置文件失败: %w", err))
			}
			fmt.Println("默认配置文件创建成功: ./config/config.yaml")
			// 重新读取配置
			if err := viper.ReadInConfig(); err != nil {
				panic(fmt.Errorf("读取配置文件失败: %w", err))
			}
		} else {
			panic(fmt.Errorf("读取配置文件失败: %w", err))
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		panic(fmt.Errorf("解码配置结构体失败: %w", err))
	}

	AppConfig = &config

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		var newConfig Config
		if err := viper.Unmarshal(&newConfig); err != nil {
			panic(fmt.Errorf("解码配置结构体失败: %w", err))
		}
		// 更新全局结构体
		AppConfig = &newConfig
	})

	return &config
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig() error {
	configPath := path.Join(".", "config")

	// 确保配置目录存在
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 将嵌入的默认配置写入文件
	if err := os.WriteFile(path.Join(configPath, "config.yaml"), defaultConfigContent, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}
