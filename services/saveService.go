package services

import (
	"GetMapsService2/models"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"
)

// SaveService 定义
type SaveService struct {
	Config *ConfigService       `inject:""`
	Baidu  *BaiduMapSaveService `inject:""`
}

// BaiduMapSaveService 定义
type BaiduMapSaveService struct {
	Config    *ConfigService    `inject:""`
	WebSocket *WebSocketService `inject:""`
}

// UploadATile 定义
func (baidu *BaiduMapSaveService) UploadATile(tile []byte, baiduProperties models.BaiduProperties) (err error) {
	url := fmt.Sprintf(baidu.Config.configStruct.BaiduMapFileSystem.URL, baiduProperties.ZoomLevel, baiduProperties.X, baiduProperties.Y)
	repeatCounter := 0
	for {
		err1 := baidu.postTileToURL(url, tile)
		if err1 == nil {
			break
		}
		repeatCounter++
		if repeatCounter%100 == 0 {
			msg := fmt.Sprintf("文件上传错误，已重试了%d次。上传地址：%s\n错误信息：%s", repeatCounter, url, err1.Error())
			baidu.WebSocket.BroadcastMessage(msg)
			if repeatCounter%10000 == 0 {
				err = errors.New(msg)
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

// postTileToURL 定义
func (baidu *BaiduMapSaveService) postTileToURL(url string, tile []byte) (err error) {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	part, err := writer.CreateFormField(url)
	part.Write(tile)
	if err != nil {
		return
	}
	contentType := writer.FormDataContentType()
	err = writer.Close()
	if err != nil {
		return
	}

	resp, err1 := http.Post(url, contentType, buf)
	if err1 != nil {
		err = err1
		return
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	resp.Body.Close()
	statusCode := resp.StatusCode
	if statusCode < 200 || statusCode > 299 {
		message := fmt.Sprintf("StatusCode is error! URL: %s", url)
		err = errors.New(message)
		return
	}

	return
}
