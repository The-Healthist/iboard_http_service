package base_models

// File 文件模型
type File struct {
	ModelFields
	Size         int64  `json:"size"`
	Md5          string `json:"md5"          gorm:"size:255;uniqueIndex"`
	Path         string `json:"path"         gorm:"size:255;uniqueIndex"`
	MimeType     string `json:"mimeType"     gorm:"size:100"`
	Oss          string `json:"oss"          gorm:"size:50"`  //区域
	Uploader     string `json:"uploader"     gorm:"size:255"` //上传者
	UploaderID   uint   `json:"uploaderId"   gorm:"index"`    //上传者ID
	UploaderType string `json:"uploaderType" gorm:"size:50"`  // user, superadmin
}
