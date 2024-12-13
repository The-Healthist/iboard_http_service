package models

// Admin 管理员模型
type BuildingAdmin struct {
	ModelFields
	BuildingID  uint         `json:"buildingId" gorm:"not null;index"`
	Password    string       `json:"-"           gorm:"size:255;not null"`
	Status      bool         `json:"status"      gorm:"default:true"`
	Buildings   []Building   `json:"buildings"   gorm:"many2many:building_admins_buildings;"`
	Permissions []Permission `json:"permissions" gorm:"many2many:admin_permissions;"`
}
