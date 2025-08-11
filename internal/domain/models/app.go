package models

// App 应用信息模型
// 包含基础字段与应用版本号
type App struct {
    ModelFields
    Version string `json:"version" gorm:"size:50;not null"`
}


