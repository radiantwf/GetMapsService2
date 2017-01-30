package services

// SaveService 定义
type SaveService struct {
	Config *ConfigService `inject:""`
}

// UploadATile 定义
func (service *SaveService) UploadATile(tile []byte) {
}
