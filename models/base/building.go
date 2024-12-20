package base_models

// Building 建筑模型
type Building struct {
	ModelFields
	Name           string          `json:"name" gorm:"size:255;not null"`
	IsmartID       string          `json:"ismartId" gorm:"size:255;not null"`
	Password       string          `json:"password" gorm:"size:255;not null"`
	Remark         string          `json:"remark" gorm:"type:text"`
	BuildingAdmins []BuildingAdmin `json:"admins" gorm:"many2many:building_admin_buildings;"`
	Notices        []Notice        `json:"notices" gorm:"many2many:notice_buildings;"`
	Advertisements []Advertisement `json:"advertisements" gorm:"many2many:advertisement_buildings;"`
}
