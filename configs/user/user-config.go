// Author: pp
// Created: 2025/3/23 16:35
// Description: This file implements user service config definition from user-config.yml

package user_config

import (
	"fmt"
	"log"
	"github.com/spf13/viper"
)

type UserConfig struct {
	JWT 	JWT 	 `mapstructure:"jwt"`
	Email 	Email 	 `mapstructure:"email"`
	Dtabase Database `mapstructure:"database"`
	Etcd 	Etcd 	 `mapstructure:"etcd"`

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

type Etcd struct {
	Endpoints 	   []string 	`mapstructure:"endpoints"`
	DialTimeout    int 		`mapstructure:"dialTimeout"`
	RequestTimeout int 		`mapstructure:"requestTimeout"`
}


func InitUserConfig(userConfigFile string) (*UserConfig, error) {
	if userConfigFile == "" {
		log.Fatal("please input user config file path")
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



