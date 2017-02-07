package services

import (
	"GetMapsService2/models"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"time"
)

// DownloadService 定义
type DownloadService struct {
	Config *ConfigService           `inject:""`
	Baidu  *BaiduMapDownloadService `inject:""`
}

// BaiduTileDownloadedHandler 定义
type BaiduTileDownloadedHandler func(tile []byte, x, y int64, z int) (err error)

// BaiduMapDownloadService 定义
type BaiduMapDownloadService struct {
	Config                *ConfigService `inject:""`
	OnBaiduTileDownloaded BaiduTileDownloadedHandler
}

// DownLoadMaps 定义
func (baidu *BaiduMapDownloadService) DownLoadMaps(areas models.AreasStruct) {
	udt := time.Now().Format("20060102")
	serverID := baidu.Config.configStruct.BaiduMapServer.MinServerNo
	for zoomCounter := areas.MinZoomLevel; zoomCounter <= areas.MaxZoomLevel; zoomCounter++ {
		cV := math.Pow(float64(2), float64(18-zoomCounter))
		unitSize := cV * 256
		for _, areaRect := range areas.Rect {
			minX := int64(math.Floor((111320.7019*areaRect.Left + 0.02068) / unitSize))
			maxX := int64(math.Floor((111320.7019*areaRect.Right + 0.02068) / unitSize))
			minY := int64(math.Floor((137651.4674*areaRect.Top - 673284.9677) / unitSize))
			maxY := int64(math.Floor((137651.4674*areaRect.Bottom - 673284.9677) / unitSize))
			for xCounter := minX; xCounter <= maxX; xCounter++ {
				for yCounter := minY; yCounter <= maxY; yCounter++ {
					for {
						url := fmt.Sprintf(baidu.Config.configStruct.BaiduMapServer.URL, serverID, xCounter, yCounter, zoomCounter, udt)
						url = strings.Replace(url, "-", "M", 0)
						raw, err := baidu.getImageFromURL(&url)
						if err == nil {
							for {
								err = baidu.OnBaiduTileDownloaded(raw, xCounter, yCounter, zoomCounter)
								if err == nil {
									break
								}
							}
							break
						}
						serverID++
						if serverID > baidu.Config.configStruct.BaiduMapServer.MaxServerNo {
							serverID = baidu.Config.configStruct.BaiduMapServer.MinServerNo
						}
					}
				}
			}
		}
	}
}

// getImageFromURL 定义
func (baidu *BaiduMapDownloadService) getImageFromURL(url *string) (content []byte, err error) {
	resp, err1 := http.Get(*url)
	defer resp.Body.Close()
	if err1 != nil {
		err = err1
		return
	}
	data, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		err = err2
		return
	}

	statusCode := resp.StatusCode
	if statusCode != 200 {
		message := fmt.Sprintf("StatusCode is error! URL: %s", *url)
		err = errors.New(message)
		return
	}

	if data != nil && len(data) > 4 {
		if data[1] == 'P' && data[2] == 'N' && data[3] == 'G' {
			content = data
		}
	}
	return
}
