// Author: pp
// Created: 2025-03-21 0:17
// Description: 

package map_service

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	map_config "walk/apps/map/config"
	utils_map "walk/apps/map/utils"

)


type MapWebService struct{
	key 				string		// 用户key
	signature 			string		// 签名私钥
	geocodeBaseURL		string		// 地理编码baseurl
	staticMapBaseURL	string 		// 静态地图baseurl
}

func NewMapWebService(amap *map_config.Amap) *MapWebService {
    return &MapWebService{
        key:            	amap.Web.Key,
        signature:      	amap.Web.Signature,
        geocodeBaseURL: 	amap.Web.GeocodeBaseURL,
		staticMapBaseURL:	amap.Web.StaticMapBaseURL,
    } 
}

// 调用地理编码API服务
// @param addr string: 位置 
func (s *MapWebService) GetGeoCode(addr string) (*http.Response, error) {
	// 请求参数
	params := map[string]string {
		"address": addr,
		"key": s.key,
	}

	encrypotParams, err := utils_map.ProcessParams(params, s.signature)
	if err != nil {
		log.Printf("handle amap request params error: %v", err)
	}

	// 拼接完整的geocode_url
	params["sig"] = encrypotParams
	queryParams := url.Values{}
	for k, v := range params {
		queryParams.Add(k, v)
	}
	geocodeURL := s.geocodeBaseURL + "?" + queryParams.Encode()

	// test 
	fmt.Printf("geocodeURL: %s\n", geocodeURL)

	// 发送get请求
	resp, err := http.Get(geocodeURL)
    if err != nil {
        fmt.Println("get request error:", err)
        return nil, err
    }

	return resp, nil
}

// 调用静态地图API服务
// @param location string: 中心点坐标: 经度, 纬度 [精确度不超过小数点后6位]
// @param zoom string: 地图缩放级别[1, 17]
// @param size string: 图片大小 h*w
// @param markers string: markersStyle:location1;location2..locationN | 可自定义
// 						  marksStyle: size[small|mid|large], color[0x000000, 0xffffff], label[0-9]|[A-Z]|[单个中文字]当size为small时不显示label
func (s *MapWebService) GetStaticMap(location, size, markers, zoom string) (*http.Response, error) {
	// 请求参数
	params := map[string]string {
		"location": location,
		"zoom": zoom,
		"size": size,
		"makers": markers,
		"key": s.key,
	}
	// 请求参数签名
	encrypotParams, err := utils_map.ProcessParams(params, s.signature)
	if err != nil {
		log.Printf("handle amap request params error: %v", err)
	}

	// 拼接完整的staticmap_url
	params["sig"] = encrypotParams
	queryParams := url.Values{}
	for k, v := range params {
		queryParams.Add(k, v)
	}
	staticMapURL := s.staticMapBaseURL + "?" + queryParams.Encode()

	// test 
	fmt.Printf("staticMapURL: %s\n", staticMapURL)

	// 发送get请求
	resp, err := http.Get(staticMapURL)
	if err != nil {
		fmt.Println("get request error:", err)
		return nil, err
	}
	// defer resp.Body.Close()

	return resp, nil
}