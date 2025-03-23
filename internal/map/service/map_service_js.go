// Author: pp
// Created: 2025-03-21 0:17
// Description: This file contains the implementation of the map service js package.

package map_service

import (
	"log"
	"net/http"

	"walk/configs/map"
	// "walk/pkg/amap-sdk"
)

type MapWebJSService struct {
	key 		   string // 用户key
	privateKey     string // 用户私钥
}

// 初始化动态地图服务配置
func NewMapWebJSService(amap *map_config.Amap) *MapWebJSService {
	return &MapWebJSService{
		key:        amap.WebJS.Key,
		privateKey: amap.WebJS.PrivateKey,
	}
}

// 获取动态地图

