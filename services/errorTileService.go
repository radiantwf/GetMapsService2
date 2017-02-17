package services

import (
	"GetMapsService2/models"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// ErrorTileService 定义
type ErrorTileService struct {
}

var errorListCaption = 1000

// BaiduErrorTileService 定义
type BaiduErrorTileService struct {
	writtingErrorFile     *os.File
	writtingErrorFileName string
	writtingErrorList     []models.BaiduProperties
	readingErrorFile      *os.File
	reader                *bufio.Reader
	mu                    sync.Mutex
}

// InitSave 定义
func (errorMaps *BaiduErrorTileService) InitSave(downloadthreadCounter int) {
	tmpPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	downloadPathName := fmt.Sprintf("%s/error/", tmpPath)
	if downloadthreadCounter == 1 {
		os.RemoveAll(downloadPathName)
		os.Mkdir(downloadPathName, 0777)
	}
	errorMaps.mu.Lock()
	defer errorMaps.mu.Unlock()
	errorMaps.writtingErrorList = make([]models.BaiduProperties, 0, errorListCaption)
	errorFileName := fmt.Sprintf("%s/baidu-errLst%d.err", downloadPathName, downloadthreadCounter)
	errorMaps.writtingErrorFileName = errorFileName
	errorMaps.writtingErrorFile, err = os.OpenFile(errorMaps.writtingErrorFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		fmt.Println(err.Error())
		errorMaps.writtingErrorFile = nil
		errorMaps.writtingErrorList = nil
		return
	}
}

// InitLoad 定义
func (errorMaps *BaiduErrorTileService) InitLoad(downloadthreadCounter int) {
	tmpPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	downloadPathName := fmt.Sprintf("%s/error/", tmpPath)

	errorMaps.mu.Lock()
	defer errorMaps.mu.Unlock()
	errorFileName := fmt.Sprintf("%s/baidu-errLst%d.err", downloadPathName, downloadthreadCounter)
	errorMaps.readingErrorFile, err = os.Open(errorFileName)
	if err != nil {
		fmt.Println(err.Error())
		errorMaps.readingErrorFile = nil
		errorMaps.reader = nil
	}
	errorMaps.reader = bufio.NewReader(errorMaps.readingErrorFile)
}

// saveLog 定义
func (errorMaps *BaiduErrorTileService) saveLog() {
	if errorMaps.writtingErrorList != nil && len(errorMaps.writtingErrorList) != 0 {

		var buf bytes.Buffer
		for _, value := range errorMaps.writtingErrorList {
			message := fmt.Sprintf("%d,%d,%d\t", value.ZoomLevel, value.X, value.Y)
			buf.WriteString(message)
		}
		fmt.Fprintln(errorMaps.writtingErrorFile, buf.String())
	}
}

// Append 定义
func (errorMaps *BaiduErrorTileService) Append(baiduProperties []models.BaiduProperties) {
	errorMaps.mu.Lock()
	defer errorMaps.mu.Unlock()
	if errorMaps.writtingErrorList == nil {
		return
	}
	for _, value := range baiduProperties {
		errorMaps.writtingErrorList = append(errorMaps.writtingErrorList, value)
		if len(errorMaps.writtingErrorList) >= errorListCaption {
			errorMaps.saveLog()
			errorMaps.writtingErrorList = make([]models.BaiduProperties, 0, errorListCaption)
		}
	}
}

// ReadLine 定义
func (errorMaps *BaiduErrorTileService) ReadLine() (baiduPropertiesList []models.BaiduProperties) {
	errorMaps.mu.Lock()
	defer errorMaps.mu.Unlock()
	if errorMaps.readingErrorFile == nil {
		return
	}
	buf, err := errorMaps.reader.ReadString('\n')

	if err == io.EOF {
		return
	} else if err != nil {
		fmt.Println(err.Error())
	}
	errorDatas := strings.Split(buf, "\t")
	baiduPropertiesList = make([]models.BaiduProperties, 0, errorListCaption)
	for _, value := range errorDatas {
		var zoomLevel int
		var x, y int64
		_, err = fmt.Sscanf(value, "%d,%d,%d", &zoomLevel, &x, &y)
		if err == nil {
			baiduProperties := models.BaiduProperties{zoomLevel, x, y}
			baiduPropertiesList = append(baiduPropertiesList, baiduProperties)
		}
	}
	return
}

// CloseSave 定义
func (errorMaps *BaiduErrorTileService) CloseSave() {
	errorMaps.mu.Lock()
	defer errorMaps.mu.Unlock()
	errorMaps.saveLog()
	errorMaps.writtingErrorFile.Close()
	errorMaps.writtingErrorFileName = ""
	errorMaps.writtingErrorFile = nil
}

// CloseRead 定义
func (errorMaps *BaiduErrorTileService) CloseRead() {
	errorMaps.reader = nil
	errorMaps.readingErrorFile.Close()
	errorMaps.readingErrorFile = nil
}
