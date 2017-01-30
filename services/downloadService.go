package services

import (
	"GetMapsService2/models"
	"math"
	"time"
)

// DownloadService 定义
type DownloadService struct {
	Config           *ConfigService `inject:""`
	OnTileDownloaded TileDownloadedHandler
}

// TileDownloadedHandler 定义
type TileDownloadedHandler func(message []byte)

// DownLoad 定义
func (service *DownloadService) DownLoad(areas models.AreasStruct) {
	service.downLoadBaiduMaps(areas)
}

type BaiduDownloadPara struct {
	serverID int
	x        int
	y        int
	z        int
}

func (service *DownloadService) downLoadBaiduMaps(areas models.AreasStruct) {
	udt := time.Now().Format("20060102")
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

				}
			}
		}
	}
}
