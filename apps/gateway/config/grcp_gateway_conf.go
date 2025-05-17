// apps/gateway/config/config.go
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type GatewayConfig struct {
	Servers []ServerConfig `mapstructure:"servers"`
	Etcd    EtcdConfig     `mapstructure:"etcd"`
}

type ServerConfig struct {
	Name      string         `mapstructure:"name"`
	Address   string         `mapstructure:"address"` // 测试用地址
	Endpoints []EndpointItem `mapstructure:"endpoints"`
}

type EndpointItem struct {
	Method        string `mapstructure:"method"`
	Path          string `mapstructure:"path"`
	GrpcService   string `mapstructure:"grpc_service"`
	GrpcMethod    string `mapstructure:"grpc_method"`
	AuthRequired  bool   `mapstructure:"auth_required"`
	SkipRefresh   bool   `mapstructure:"skip_refresh"` // 是否跳过 refreshToken 自动刷新
}

type EtcdConfig struct {
	Endpoints   []string `mapstructure:"endpoints"`
	DialTimeout int      `mapstructure:"dial_timeout"`
}

const DefaultConfigPath = "./config.yaml"

func InitGatewayConfig(path string) (*GatewayConfig, error) {
	if path == "" {
		path = DefaultConfigPath
		return nil, fmt.Errorf("config path is empty, using default: %s", path)
	}

	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg GatewayConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}