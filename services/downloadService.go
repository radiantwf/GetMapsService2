package services

import (
	"GetMapsService2/models"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

var downloadFlag bool
var errorCounter uint64
var counter uint64
var totalCount uint64
var timesCounter int

var aliveThreadCounter int64

// DownloadService 定义
type DownloadService struct {
	Config *ConfigService           `inject:""`
	Baidu  *BaiduMapDownloadService `inject:""`
}

// BaiduTileDownloadedHandler 定义
type BaiduTileDownloadedHandler func(tile []byte, baiduProperties models.BaiduProperties) (err error)

// BaiduMapDownloadService 定义
type BaiduMapDownloadService struct {
	Config                *ConfigService    `inject:""`
	WebSocket             *WebSocketService `inject:""`
	OnBaiduTileDownloaded BaiduTileDownloadedHandler
	ErrorList             *BaiduErrorTileService `inject:""`
	channel               chan int
	udt                   string
}

// DownLoadMaps 定义
func (baidu *BaiduMapDownloadService) DownLoadMaps(areas models.AreasStruct) {
	timesCounter = 1
	baidu.ErrorList.InitSave(timesCounter)
	baidu.initDownload()
	baidu.computeCounter(areas)
	msg := fmt.Sprintf("下载开始，共计%d个文件需要下载。", totalCount)
	baidu.WebSocket.BroadcastMessage(msg)

	downloadFlag = true
	go baidu.putProcessingMessage()

	baidu.udt = time.Now().Format("20060102")
	mapBaiduPropertiesList := make([]models.BaiduProperties, 0, baidu.Config.configStruct.ProcessListCapacity)
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

					if len(mapBaiduPropertiesList) >= baidu.Config.configStruct.ProcessListCapacity {
						baidu.channel <- 0
						atomic.AddInt64(&aliveThreadCounter, 1)

						go baidu.downloadTiles(mapBaiduPropertiesList, baidu.udt)

						mapBaiduPropertiesList = make([]models.BaiduProperties, 0, baidu.Config.configStruct.ProcessListCapacity)
					}

					mapProperties := models.BaiduProperties{ZoomLevel: zoomCounter, X: xCounter, Y: yCounter}
					mapBaiduPropertiesList = append(mapBaiduPropertiesList, mapProperties)
				}
			}
		}
	}
	if len(mapBaiduPropertiesList) > 0 {
		baidu.channel <- 0
		atomic.AddInt64(&aliveThreadCounter, 1)

		go baidu.downloadTiles(mapBaiduPropertiesList, baidu.udt)
	}
	for {
		if atomic.LoadInt64(&aliveThreadCounter) == 0 {
			break
		}
		time.Sleep(3 * time.Second)
	}
	baidu.ErrorList.CloseSave()

	msg = fmt.Sprintf("第%d下载完成，共计%d个文件，%d个文件下载成功，%d个文件下载失败。", timesCounter, totalCount, atomic.LoadUint64(&counter), atomic.LoadUint64(&errorCounter))
	baidu.WebSocket.BroadcastMessage(msg)
	if errorCounter > 0 {
		baidu.fetchErrorList(errorCounter)
	}

	downloadFlag = false
}

// initDownload 定义
func (baidu *BaiduMapDownloadService) initDownload() {
	baidu.channel = make(chan int, baidu.Config.configStruct.AllowedThreadCount)

	atomic.StoreUint64(&counter, 0)
	atomic.StoreUint64(&errorCounter, 0)

	totalCount = 0
	atomic.StoreInt64(&aliveThreadCounter, 0)
}

// computeCounter 定义
func (baidu *BaiduMapDownloadService) computeCounter(areas models.AreasStruct) {
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
					totalCount++
				}
			}
		}
	}
}

// downloadAtile 定义
func (baidu *BaiduMapDownloadService) downloadTiles(list []models.BaiduProperties, udt string) {
	serverID := baidu.Config.configStruct.BaiduMapServer.MinServerNo
	errorList := make([]models.BaiduProperties, 0, len(list))
	for _, value := range list {
		repeatCounter := 0
		for {
			serverID++
			if serverID > baidu.Config.configStruct.BaiduMapServer.MaxServerNo {
				serverID = baidu.Config.configStruct.BaiduMapServer.MinServerNo
			}
			url := fmt.Sprintf(baidu.Config.configStruct.BaiduMapServer.URL, serverID, value.X, value.Y, value.ZoomLevel, udt)
			url = strings.Replace(url, "-", "M", 0)
			raw, err := baidu.getImageFromURL(&url)
			if err == nil {
				if baidu.OnBaiduTileDownloaded != nil {
					err2 := baidu.OnBaiduTileDownloaded(raw, value)
					if err2 != nil {
						atomic.AddUint64(&errorCounter, 1)
						errorList = append(errorList, value)
					} else {
						atomic.AddUint64(&counter, 1)
					}
				}
				break
			}
			repeatCounter++
			if repeatCounter%100 == 0 {
				msg := fmt.Sprintf("文件下载错误，已重试了%d次。下载地址：%s\n错误信息：%s", repeatCounter, url, err.Error())
				baidu.WebSocket.BroadcastMessage(msg)
				if repeatCounter%10000 == 0 {
					atomic.AddUint64(&errorCounter, 1)
					errorList = append(errorList, value)
					break
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	baidu.ErrorList.Append(errorList)

	atomic.AddInt64(&aliveThreadCounter, -1)
	<-baidu.channel
}

func (baidu *BaiduMapDownloadService) fetchErrorList(total uint64) {
	baidu.initDownload()
	atomic.StoreUint64(&totalCount, total)
	baiduPropertiesList := make([]models.BaiduProperties, 0, baidu.Config.configStruct.ProcessListCapacity)
	baidu.ErrorList.InitLoad(timesCounter)
	timesCounter++
	baidu.ErrorList.InitSave(timesCounter)
	for {
		lines := baidu.ErrorList.ReadLine()
		if lines == nil {
			break
		}
		for _, value := range lines {
			if len(baiduPropertiesList) >= baidu.Config.configStruct.ProcessListCapacity {
				baidu.channel <- 0
				atomic.AddInt64(&aliveThreadCounter, 1)

				go baidu.downloadTiles(baiduPropertiesList, baidu.udt)

				baiduPropertiesList = make([]models.BaiduProperties, 0, baidu.Config.configStruct.ProcessListCapacity)
			}
			baiduProperties := value
			baiduPropertiesList = append(baiduPropertiesList, baiduProperties)
		}
	}

	if len(baiduPropertiesList) > 0 {
		baidu.channel <- 0
		atomic.AddInt64(&aliveThreadCounter, 1)

		go baidu.downloadTiles(baiduPropertiesList, baidu.udt)
	}
	for {
		if atomic.LoadInt64(&aliveThreadCounter) == 0 {
			break
		}
		time.Sleep(3 * time.Second)
	}
	baidu.ErrorList.CloseRead()
	baidu.ErrorList.CloseSave()

	msg := fmt.Sprintf("第%d下载完成，共计%d个文件，%d个文件下载成功，%d个文件下载失败。", timesCounter, totalCount, atomic.LoadUint64(&counter), atomic.LoadUint64(&errorCounter))
	baidu.WebSocket.BroadcastMessage(msg)
	if errorCounter > 0 {
		baidu.fetchErrorList(errorCounter)
	}
}

// getImageFromURL 定义
func (baidu *BaiduMapDownloadService) getImageFromURL(url *string) (content []byte, err error) {
	resp, err1 := http.Get(*url)
	if err1 != nil {
		err = err1
		return
	}

	defer resp.Body.Close()
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

// putProcessingMessage 定义
func (baidu *BaiduMapDownloadService) putProcessingMessage() {
	for {
		if downloadFlag == false {
			atomic.StoreUint64(&counter, 0)
			break
		}
		if counter > 0 {
			msg := fmt.Sprintf("正在进行第%d轮下载，%d个文件下载成功，共计%d个文件，%d个文件下载失败。", timesCounter, atomic.LoadUint64(&counter), atomic.LoadUint64(&totalCount), atomic.LoadUint64(&errorCounter))

			baidu.WebSocket.BroadcastMessage(msg)
			time.Sleep(3 * time.Second)
		} else {
			time.Sleep(3 * time.Second)
		}
	}
}
