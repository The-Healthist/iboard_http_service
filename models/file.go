package models

// File 文件模型
type File struct {
	ModelFields
	Size         int64  `json:"size"`
	Path         string `json:"path"         gorm:"size:255;uniqueIndex"`
	MimeType     string `json:"mimeType"     gorm:"size:100"`
	Oss          string `json:"oss"          gorm:"size:50"`
	Uploader     string `json:"uploader"     gorm:"size:255"`
	UploaderType string `json:"uploaderType" gorm:"size:50"` // user, superadmin
}
