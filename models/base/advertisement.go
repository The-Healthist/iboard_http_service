package base_models

// Advertisement 广告模型
type Advertisement struct {
	ModelFields
	Title       string     `json:"title"        gorm:"size:255;not null"`
	Description string     `json:"description"  gorm:"type:text"`
	Type        string     `json:"type"         gorm:"size:50"` // video, image
	Active      bool       `json:"active"       gorm:"default:true"`
	Duration    int        `json:"duration"`
	Display     string     `json:"display"      gorm:"size:50"` // full, top, topfull
	FileID      *uint      `json:"file_id"      gorm:"default:null"`
	File        File       `json:"file"         gorm:"foreignKey:FileID"`
	IsPublic    bool       `json:"is_public"    gorm:"default:true"`
	Buildings   []Building `json:"buildings"    gorm:"many2many:advertisement_buildings;"`
}
