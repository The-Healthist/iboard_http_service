package base_models

// Admin 管理员模型
type BuildingAdmin struct {
	ModelFields
	Email     string     `json:"email"       gorm:"size:255;not null;unique"`
	Password  string     `json:"-"           gorm:"size:255;not null"`
	Status    bool       `json:"status"      gorm:"default:true"`
	Buildings []Building `json:"buildings"   gorm:"many2many:building_admins_buildings;"`
}
