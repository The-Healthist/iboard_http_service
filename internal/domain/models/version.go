package models

// Version 版本信息模型
type Version struct {
	ModelFields
	VersionNumber string `json:"versionNumber" gorm:"size:50;not null;uniqueIndex"`
	BuildNumber   string `json:"buildNumber" gorm:"size:50;comment:'构建号'"`
	Description   string `json:"description" gorm:"type:text"`
	DownloadUrl   string `json:"downloadUrl" gorm:"size:500;not null"`
	Status        string `json:"status" gorm:"size:50;default:'active';comment:'状态:active,inactive,deprecated'"`
}
