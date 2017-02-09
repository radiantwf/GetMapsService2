package models

// AreasResponseStruct 定义
type AreasResponseStruct struct {
	MinZoomLevel *string `json:"MinZoomLevel"`
	MaxZoomLevel *string `json:"MaxZoomLevel"`
	Province     *string `json:"Province"`
}

// AreasStruct 定义
type AreasStruct struct {
	MinZoomLevel, MaxZoomLevel int
	Provinces                  []string
	Rect                       []RectStruct
}

// RectStruct 定义
type RectStruct struct {
	Top    float64
	Bottom float64
	Left   float64
	Right  float64
}
