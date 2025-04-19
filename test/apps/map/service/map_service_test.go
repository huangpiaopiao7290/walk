package map_test

import (
	"fmt"
	"os"
	"testing"

	map_config "walk/apps/map/config"
	map_service "walk/apps/map/service"
	map_utils "walk/apps/map/utils"
)

// 初始化地图服务
func setup(t *testing.T) (*map_service.MapWebService, *map_config.MapConfig) {
	filePath := "/home/pp/programs/program_go/timeTrack/walk/configs/map-service.yml"

	cfg, err := map_config.GetMapConfig(filePath)
	if err != nil {
		t.Errorf("GetMapConfig returned unexpected error: %v", err)
	}

	ms := map_service.NewMapWebService(&cfg.Amap)

	return ms, cfg
}

// 测试地理编码
func TestGetGeocode(t *testing.T) {

	ms, _ := setup(t)

	var addr string = "北京市朝阳区阜通东大街6号"

	resp, err := ms.GetGeoCode(addr)
	if err != nil {
		t.Errorf("get map eror,: %v", err)
		return
	}

	defer resp.Body.Close()

	// 处理响应
	fmt.Println("response:")
	fmt.Printf("status: %d\n", resp.StatusCode)

	// 读取响应体
	body := make([]byte, 1024)
	n, _ := resp.Body.Read(body)
	fmt.Printf("response body: %s\n", string(body[:n]))

}


// 测试获取静态地图
func TestGetStaticMap(t *testing.T)  {
	ms, _ := setup(t)

	// 地图参数
	var loc string = "116.481485,39.990464"
	var zoom string = "4"
	var size string = "800*800"
	var makers string = "mid,,A:116.481485,39.990464"

	resp, err := ms.GetStaticMap(loc, size, makers, zoom)
	if err != nil {
		t.Errorf("get map eror,: %v", err)
		return
	}
	defer resp.Body.Close()

		// 处理响应
		fmt.Println("response:")
		fmt.Printf("status: %d\n", resp.StatusCode)
	
		// 图片存储指定位置：xx/test.jpg
		dir, err := os.Getwd()
		if err != nil {
			t.Errorf("error getting current directory: %v", err)
			return
		}
		imgPath := dir + "/test.jpg"
		err = map_utils.SaveImageFromResponse("/getStaticMap", resp, imgPath)
	
		if err != nil {
			t.Errorf("error saving image from response: %v", err)
			return
		}
	
		fmt.Printf("test image has saved in %s", imgPath)
	 

}


