package models

// Advertisement 广告模型
type Advertisement struct {
	ModelFields
	Title       string     `json:"title"        gorm:"size:255;not null"`
	Description string     `json:"description"  gorm:"type:text"`
	Type        string     `json:"type"         gorm:"size:50"` // video, image
	FileID      uint       `json:"fileId"`
	File        File       `json:"file"         gorm:"foreignKey:FileID"`
	Active      bool       `json:"active"       gorm:"default:true"`
	Duration    int        `json:"duration"`
	Display     string     `json:"display"      gorm:"size:50"` // full, top, topfull
	Buildings   []Building `json:"buildings"    gorm:"many2many:building_advertisements;"`
}
