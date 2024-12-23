package base_models

import "github.com/The-Healthist/iboard_http_service/utils/field"

// Advertisement 广告模型
type Advertisement struct {
	ModelFields
	Title       string                     `json:"title"        gorm:"size:255;not null"`
	Description string                     `json:"description"  gorm:"type:text"`
	Type        field.AdvertisementType    `json:"type"         gorm:"size:50"` // video, image
	Active      bool                       `json:"active"       gorm:"default:true"`
	Duration    int                        `json:"duration"`
	Display     field.AdvertisementDisplay `json:"display"      gorm:"size:50"` // full, top, topfull
	FileID      *uint                      `json:"file_id"      gorm:"default:null"`
	File        File                       `json:"file"         gorm:"foreignKey:FileID"`
	IsPublic    bool                       `json:"is_public"    gorm:"default:true"`
	Buildings   []Building                 `json:"-"            gorm:"many2many:advertisement_buildings;"`
}
