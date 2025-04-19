package map_config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// map 结构体
type MapConfig struct {
	Server  Server  `yaml:"server"`
	Amap 	Amap    `yaml:"amap"`
}

type Server struct {
	Port string `yaml:"port"`
}

type Amap struct {
	Web Web `yaml:"web"`
	WebJS WebJS `yaml:"web_js" mapstructure:"web_js"`
}

type Web struct {
    Key                 string `yaml:"key"`               							  // web服务应用key
    Signature           string `yaml:"signature"`         							  // 服务签名
    GeocodeBaseURL      string `yaml:"geocodeBaseURL"`    							  // 地理/逆地理编码API
    StaticMapBaseURL    string `yaml:"staticMapBaseURL"`  							  // 静态地图API
}

type WebJS struct {
    Key                 string `yaml:"key"`               							  // web端(js api)服务应用key
    PrivateKey          string `yaml:"private_key" mapstructure:"private_key"`        // web端(js api)服务应用密钥
}


// 读取地图服务配置文件
// @param mapConfigFile string 配置文件路径
// @return (*MapConfig, error) 配置文件内容, 错误信息
func GetMapConfig(mapConfigFile string) (*MapConfig, error) {
	if mapConfigFile == "" {
		log.Fatal("please input map config file path")
	}
	viper.SetConfigFile(mapConfigFile)
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()		// Find and read the config file
	if err != nil {
		return nil,  fmt.Errorf("read map config error: %w", err)
	}
	
	// 将配置文件内容解析到结构体中
	var mapConfig MapConfig
	if err := viper.Unmarshal(&mapConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config in viper: %w", err)
	}

	return &mapConfig, err
}

