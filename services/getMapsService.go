package services

import (
	"GetMapsService2/models"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// GetMapsService 定义
type GetMapsService struct {
	Config      *ConfigService   `inject:""`
	Download    *DownloadService `inject:""`
	Save        *SaveService     `inject:""`
	downloading bool
}

// GetMaps 定义
func (service *GetMapsService) GetMaps(message []byte) {
	if service.downloading == true {
		return
	}
	go service.doGet(message)
}

// GetDownloadAreas 定义
func (service *GetMapsService) GetDownloadAreas(message []byte) (areas models.AreasStruct) {
	var rep models.AreasResponseStruct
	var err error
	if err = json.Unmarshal(message, &rep); err != nil {
		fmt.Println(err.Error())
	}

	areas.MinZoomLevel, err = strconv.Atoi(*rep.MinZoomLevel)
	if err != nil {
		fmt.Println(err.Error())
		areas.MinZoomLevel = 3
	}
	areas.MaxZoomLevel, err = strconv.Atoi(*rep.MaxZoomLevel)
	if err != nil {
		fmt.Println(err.Error())
		areas.MaxZoomLevel = 18
	}
	areas.Provinces = strings.Split(*rep.Province, ",")
	if err != nil {
		fmt.Println(err.Error())
		areas.Provinces = nil
	}
	for _, province := range areas.Provinces {
		for _, value := range service.Config.configStruct.ProvinceInformation {
			if province == value.province {
				var rect models.RectStruct
				longitude := value.area.longitude
				latitude := value.area.latitude
				if longitude[0] < longitude[1] {
					rect.Left = longitude[0]
					rect.Right = longitude[1]
				} else {
					rect.Left = longitude[1]
					rect.Right = longitude[0]
				}
				if latitude[0] < latitude[1] {
					rect.Top = latitude[0]
					rect.Bottom = latitude[1]
				} else {
					rect.Top = latitude[1]
					rect.Bottom = latitude[0]
				}
				if areas.Rect == nil {
					areas.Rect = make([]models.RectStruct, 1, 100)
					areas.Rect[0] = rect
				} else {
					areas.Rect = append(areas.Rect, rect)
				}
			}
		}
	}
	return
}

// doGet 定义
func (service *GetMapsService) doGet(message []byte) {
	service.downloading = true
	areas := service.GetDownloadAreas(message)
	service.Download.Baidu.OnBaiduTileDownloaded = service.Save.Baidu.UploadATile
	service.Download.Baidu.DownLoadMaps(areas)
	service.downloading = false
}
