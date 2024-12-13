package models

// Notice 通知模型
type Notice struct {
	ModelFields
	Title       string     `json:"title"        gorm:"size:255;not null"`
	Description string     `json:"description"  gorm:"type:text"`
	Type        string     `json:"type"         gorm:"size:50"`
	FileID      uint       `json:"fileId"`
	File        File       `json:"file"         gorm:"foreignKey:FileID"`
	Field       string     `json:"field"        gorm:"type:text"`
	Buildings   []Building `json:"buildings"    gorm:"many2many:building_notices;"`
}
