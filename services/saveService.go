package services

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

// SaveService 定义
type SaveService struct {
	Config *ConfigService       `inject:""`
	Baidu  *BaiduMapSaveService `inject:""`
}

// BaiduMapSaveService 定义
type BaiduMapSaveService struct {
	Config *ConfigService `inject:""`
}

// UploadATile 定义
func (baidu *BaiduMapSaveService) UploadATile(tile []byte, x, y int64, z int) (err error) {
	url := fmt.Sprintf(baidu.Config.configStruct.BaiduMapFileSystem.URL, z, x, y)
	err = baidu.postTileToURL(url, tile)
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
	defer resp.Body.Close()
	if err1 != nil {
		err = err1
		return
	}
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
