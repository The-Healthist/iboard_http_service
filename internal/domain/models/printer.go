package models

// Printer represents a printer device that can be associated with a display device
type Printer struct {
	ModelFields
	DeviceID     *uint   `json:"deviceId,omitempty" gorm:"index"` // 外键关联到 Device
	DisplayName  *string `json:"display_name,omitempty" gorm:"size:255"`
	IPAddress    *string `json:"ip_address,omitempty" gorm:"size:255"`
	Name         *string `json:"name,omitempty" gorm:"size:255"`
	URI          *string `json:"uri,omitempty" gorm:"size:500"` // 打印机 URI，放在 Name 后面
	State        *string `json:"state,omitempty" gorm:"size:100"`
	Status       *string `json:"status,omitempty" gorm:"size:100"`        // 打印机网络状态: "online" 或 "offline"
	Reason       *string `json:"reason,omitempty" gorm:"size:500"`        // 状态原因，online时为空，offline时包含失败原因
	MarkerLevels *string `json:"marker_levels,omitempty" gorm:"size:255"` // 墨盒墨水量，格式如 "30,20"
}
