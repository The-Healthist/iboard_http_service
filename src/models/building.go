package models

import (
	"time"
)

type Building struct {
	ID        uint   `gorm:"primarykey" json:"id"`
	Name      string `gorm:"size:255;not null" json:"name"`
	IsmartID  string `gorm:"size:255" json:"ismartId"`
	Password  string `gorm:"size:255" json:"password"`
	Remark    string `gorm:"type:text" json:"remark"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
