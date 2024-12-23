package base_models

import "github.com/The-Healthist/iboard_http_service/utils/field"

// Notice 通知模型
type Notice struct {
	ModelFields
	Title       string           `json:"title"        gorm:"size:255;not null"`
	Description string           `json:"description"  gorm:"type:text"`
	Type        field.NoticeType `json:"type"         gorm:"size:50"` //urgent , common ,system, government
	FileID      *uint            `json:"fileId"       gorm:"default:null"`
	File        File             `json:"file"         gorm:"foreignKey:FileID"`
	FileType    field.FileType   `json:"fileType"     gorm:"size:50" default:"pdf"`
	IsPublic    bool             `json:"is_public"    gorm:"default:true"`
	Buildings   []Building       `json:"-"            gorm:"many2many:notice_buildings;"`
}
