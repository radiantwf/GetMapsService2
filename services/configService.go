package services

import (
	"GetMapsService2/common"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"runtime"
)

// ConfigService 定义
type ConfigService struct {
	configStruct *ConfigStruct
}

// BaiduMapServerJSONStruct 定义
type BaiduMapServerJSONStruct struct {
	MinServerNo int
	MaxServerNo int
	URL         string
}

// ConfigJSONStruct 定义
type ConfigJSONStruct struct {
	Version             string
	UpdateDate          string
	HTMLPath            string
	StaticPath          string
	Port                int
	AllowedThreadCount  int
	BaiduMapServer      BaiduMapServerJSONStruct
	BaiduMapFileSystem  BaiduMapFileSystemStruct
	ProvinceInformation []map[string]interface{}
}

// BaiduMapServerStruct 定义
type BaiduMapServerStruct struct {
	MinServerNo int
	MaxServerNo int
	URL         string
}

// BaiduMapFileSystemStruct 定义
type BaiduMapFileSystemStruct struct {
	URL string
}

// ConfigStruct 定义
type ConfigStruct struct {
	Version             string
	UpdateDate          string
	HTMLPath            string
	StaticPath          string
	Port                int
	AllowedThreadCount  int
	BaiduMapServer      BaiduMapServerStruct
	BaiduMapFileSystem  BaiduMapFileSystemStruct
	ProvinceInformation []ProvinceInfoStruct
}

// ProvinceInfoStruct 定义
type ProvinceInfoStruct struct {
	province string
	area     AreaStruct
}

// AreaStruct 定义
type AreaStruct struct {
	longitude [2]float64
	latitude  [2]float64
}

// NewConfig 定义
func NewConfig() (config ConfigService, err error) {
	var jsonStruct ConfigJSONStruct
	config.configStruct = new(ConfigStruct)
	err = config.loadJSONFile(&jsonStruct)
	if err != nil {
		return
	}
	common.StructDeepCopy(&jsonStruct, config.configStruct)

	if runtime.GOOS == "darwin" {
		if config.configStruct.AllowedThreadCount > 100 {
			config.configStruct.AllowedThreadCount = 100
		}
	}

	config.configStruct.ProvinceInformation = make([]ProvinceInfoStruct, 0, 50)
	for _, value := range jsonStruct.ProvinceInformation {
		var p ProvinceInfoStruct
		p.province = value["province"].(string)
		longitudeInterface := value["area"].(map[string]interface{})["longitude"].([]interface{})
		latitudeInterface := value["area"].(map[string]interface{})["latitude"].([]interface{})
		p.area = AreaStruct{[2]float64{longitudeInterface[0].(float64), longitudeInterface[1].(float64)}, [2]float64{latitudeInterface[0].(float64), latitudeInterface[1].(float64)}}
		config.configStruct.ProvinceInformation = append(config.configStruct.ProvinceInformation, p)
	}
	fmt.Println("当前配置文件信息为：")
	fmt.Println(*config.configStruct)
	return
}

// loadJSONFile 定义
func (config *ConfigService) loadJSONFile(jsonStruct *ConfigJSONStruct) (err error) {
	var jsonStr []byte
	jsonStr, err = ioutil.ReadFile("./resources/config/config.json")
	if err != nil {
		return
	}
	err = json.Unmarshal(jsonStr, jsonStruct)
	return
}
