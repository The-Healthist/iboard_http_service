package models

// Building 建筑模型
type Building struct {
	ModelFields
	Name           string          `json:"name"         gorm:"size:255;not null;index"`
	IsmartID       string          `json:"ismartId"     gorm:"size:255;index"`
	Password       string          `json:"password"     gorm:"size:255"`
	Remark         string          `json:"remark"       gorm:"type:text"`
	Admins         []BuildingAdmin `json:"admins"       gorm:"many2many:building_admins_buildings;"`
	Notices        []Notice        `json:"notices"      gorm:"many2many:building_notices;"`
	Advertisements []Advertisement `json:"advertisements" gorm:"many2many:building_advertisements;"`
}
