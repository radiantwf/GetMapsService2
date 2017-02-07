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

var counter uint64

var totalCount uint64

var aliveThreadCounter int64

// DownloadService 定义
type DownloadService struct {
	Config *ConfigService           `inject:""`
	Baidu  *BaiduMapDownloadService `inject:""`
}

// BaiduTileDownloadedHandler 定义
type BaiduTileDownloadedHandler func(tile []byte, x, y int64, z int) (err error)

// BaiduMapDownloadService 定义
type BaiduMapDownloadService struct {
	Config                *ConfigService    `inject:""`
	WebSocket             *WebSocketService `inject:""`
	OnBaiduTileDownloaded BaiduTileDownloadedHandler
	channel               chan int
}

// DownLoadMaps 定义
func (baidu *BaiduMapDownloadService) DownLoadMaps(areas models.AreasStruct) {
	baidu.channel = make(chan int, baidu.Config.configStruct.AllowedThreadCount)

	atomic.StoreUint64(&counter, 0)
	totalCount = 0
	atomic.StoreInt64(&aliveThreadCounter, 0)
	downloadFlag = true

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
	msg := fmt.Sprintf("下载开始，共计%d个文件需要下载。", totalCount)
	baidu.WebSocket.BroadcastMessage(msg)

	go baidu.putProcessingMessage()

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
					baidu.channel <- 0
					atomic.AddInt64(&aliveThreadCounter, 1)

					go baidu.downloadAtile(serverID, xCounter, yCounter, zoomCounter, udt)
					serverID++
					if serverID > baidu.Config.configStruct.BaiduMapServer.MaxServerNo {
						serverID = baidu.Config.configStruct.BaiduMapServer.MinServerNo
					}
				}
			}
		}
	}
	for {
		if atomic.LoadInt64(&aliveThreadCounter) == 0 {
			break
		}
		time.Sleep(3 * time.Second)
	}
	downloadFlag = false
	msg = fmt.Sprintf("下载完成，共计%d个文件，%d个文件下载成功。", totalCount, atomic.LoadUint64(&counter))
	baidu.WebSocket.BroadcastMessage(msg)
}

func (baidu *BaiduMapDownloadService) downloadAtile(serverID int, xCounter, yCounter int64, zoomCounter int, udt string) {
	for {
		url := fmt.Sprintf(baidu.Config.configStruct.BaiduMapServer.URL, serverID, xCounter, yCounter, zoomCounter, udt)
		url = strings.Replace(url, "-", "M", 0)
		raw, err := baidu.getImageFromURL(&url)
		if err == nil {
			err = baidu.OnBaiduTileDownloaded(raw, xCounter, yCounter, zoomCounter)
			break
		}
		serverID++
		if serverID > baidu.Config.configStruct.BaiduMapServer.MaxServerNo {
			serverID = baidu.Config.configStruct.BaiduMapServer.MinServerNo
		}
	}
	atomic.AddUint64(&counter, 1)
	<-baidu.channel
	atomic.AddInt64(&aliveThreadCounter, -1)
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
			msg := fmt.Sprintf("%d个文件下载完成，共计%d个文件。", atomic.LoadUint64(&counter), totalCount)
			baidu.WebSocket.BroadcastMessage(msg)
			time.Sleep(3 * time.Second)
		} else {
			time.Sleep(3 * time.Second)
		}
	}
}
