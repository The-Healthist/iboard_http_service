package models

import "time"

// App 应用版本配置模型
// 用于存储当前使用的版本配置信息
type App struct {
	ModelFields
	CurrentVersionID uint      `json:"currentVersionId" gorm:"default:1;comment:'当前使用的版本ID，1表示使用第一个版本'"`
	CurrentVersion   *Version  `json:"currentVersion" gorm:"foreignKey:CurrentVersionID;comment:'当前版本信息'"`
	LastCheckTime    time.Time `json:"lastCheckTime" gorm:"type:datetime;comment:'最后检查更新时间'"`
	UpdateInterval   int       `json:"updateInterval" gorm:"default:3600;comment:'检查更新间隔(秒)'"`
	AutoUpdate       bool      `json:"autoUpdate" gorm:"default:false;comment:'是否自动更新'"`
	Status           string    `json:"status" gorm:"size:50;default:'active';comment:'应用状态'"`
}
