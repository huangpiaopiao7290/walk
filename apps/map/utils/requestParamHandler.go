// Author: pp
// Created: 2025-03-21 0:17
// Description: the mehtods of handling request parameters about amap service
//              

package map_utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"os"
	"net/http"
	"io"
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


// 处理Amap地图服务响应: 保存返回的图片
// @param reqURL string: request url
// @param resp *http.Response: response
// @param filepath string: 保存路径
func SaveImageFromResponse(reqURL string, resp *http.Response, filepath string) error {
	// 判断响应是否正常
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("for request to url: %v. got %d", reqURL, resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("for request to url: %v,  respone is ok but repone body is empty: ", err)
	}

	defer resp.Body.Close()

	// 将数据写入文件
	err = os.WriteFile(filepath, body, 0644)
	if err != nil {
		return fmt.Errorf("error writing image file to %s", filepath)
	}

	roorDIR, err :=  os.Getwd()
	if err != nil {
		fmt.Println("get root dir fail", err)
	}
	fmt.Println("root dir: ", roorDIR)

	return nil

}