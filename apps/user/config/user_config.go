// Author: pp
// Created: 2025/3/23 16:35
// Description: This file implements user service config definition from user-config.yml

package user_config

import (
	"fmt"
	"github.com/spf13/viper"
)

type UserConfig struct {
	JWT 		JWT 	 	`mapstructure:"jwt"`
	Email 		Email 	 	`mapstructure:"email"`
	Database 	Database 	`mapstructure:"database"`
	Redis  		Redis 	`mapstructure:"Redis"`
	Etcd 		Etcd 	 	`mapstructure:"etcd"`

}

type JWT struct {
	Secret 				string 		`mapstructure:"secret"`
	Access_token_ttl 	int 		`mapstructure:"access_token_ttl"`
	Refresh_token_ttl   int 		`mapstructure:"refresh_token_ttl"`
}

type Email struct {
	Host 		string 	`mapstructure:"host"`
	Port 		int 	`mapstructure:"port"`
	Username 	string 	`mapstructure:"username"`
	Password 	string 	`mapstructure:"password"`
}

type Database struct {
	Host 		string 	`mapstructure:"host"`
	Port 		int 	`mapstructure:"port"`
	Driver 		string 	`mapstructure:"driver"`
	Username 	string 	`mapstructure:"username"`
	Password 	string 	`mapstructure:"password"`
	DBname 		string 	`mapstructure:"DBname"`
}


type Redis struct {
	Network 		string 	`mapstructure:"network"`
	Addr 			string 	`mapstructure:"addr"`
	Username 		string 	`mapstructure:"username"`
	Password 		string 	`mapstructure:"password"`
	DB 				int 	`mapstructure:"db"`
	PoolSize 		int 	`mapstructure:"poolSize"`
	MinIdleConns 	int 	`mapstructure:"minIdleTimeout"`
	MaxIdleConns 	int 	`mapstructure:"maxIdleTimeout"`
	MaxRetries 		int 	`mapstructure:"maxRetries"`
}

type Etcd struct {
	Endpoints 	   []string 	`mapstructure:"endpoints"`
	DialTimeout    int 		`mapstructure:"dialTimeout"`
	RequestTimeout int 		`mapstructure:"requestTimeout"`
}


// 默认配置文件路径（相对项目根目录）
const defaultConfigPath = "apps/user/config/user-service.yml"

func InitUserConfig(userConfigFile string) (*UserConfig, error) {
	if userConfigFile == "" {
		// 使用默认配置文件路径
		userConfigFile = defaultConfigPath
		return nil, fmt.Errorf("user config file is empty, using default apth: %s", defaultConfigPath)
	}

	viper.SetConfigFile(userConfigFile)
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("read user config error: %w", err)
	}

	var userConfig UserConfig
	err = viper.Unmarshal(&userConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config in viper: %w", err)
	}

	return &userConfig, nil
}


