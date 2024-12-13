package models

// Permission 权限模型
type Permission struct {
	ModelFields
	Object   string `json:"object"   gorm:"size:255;not null;index"` // building, file
	Type     string `json:"type"     gorm:"size:50;not null;index"`  // r, w
	EntityID uint   `json:"entityId" gorm:"not null;index"`
	APIs     []API  `json:"apis"     gorm:"many2many:api_permissions;"`
}

// API API模型
type API struct {
	ModelFields
	Path        string       `json:"path"        gorm:"size:255;not null"`
	Method      string       `json:"method"      gorm:"size:20;not null"`
	Description string       `json:"description" gorm:"type:text"`
	Permissions []Permission `json:"permissions" gorm:"many2many:api_permissions;"`
}
