// Author: pp
// Created: 2025-03-21 0:17
// Description: the mehtods of handling request parameters about amap service
//              

package amapsdk

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// 处理Amap地图服务请求参数
// @param params map[string]string: 请求参数
// @param sig string: 签名私钥
func ProcessParams(params map[string]string, sig string) (string, error) {
	if len(params) == 0 {
		return "", fmt.Errorf("params map is empty")
	}

	// 参数名升序排列
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接参数
	var paramStr string
	for _, k := range keys {
		paramStr += fmt.Sprintf("%s=%s&", k, params[k])
	}
	paramStr = strings.TrimRight(paramStr, "&")

	// 生成签名
	signatureStr  := paramStr + sig
	hash := md5.Sum([]byte(signatureStr))
	encroptParams := hex.EncodeToString(hash[:])

	return encroptParams, nil
}	
