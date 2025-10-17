package models

import "gorm.io/datatypes"

// Device represents a display device in a building
type Device struct {
	ModelFields
	DeviceID   string         `json:"deviceId" gorm:"size:255;not null;unique"`
	Building   Building       `json:"building" gorm:"foreignKey:BuildingID"`
	BuildingID uint           `json:"buildingId" `
	Printers   []Printer      `json:"-" gorm:"foreignKey:DeviceID"`                        // 一对多关系，不直接序列化
	OrangePi   OrangePiInfo   `json:"orangePi" gorm:"embedded;embedded_prefix:orange_pi_"` // 包含打印机信息
	Settings   DeviceSettings `json:"settings" gorm:"embedded"`
	// 轮播顺序管理列表（JSON 数组，存储 ID 顺序）
	TopAdvertisementCarouselList  datatypes.JSON `json:"topAdvertisementCarouselList" gorm:"type:json"`
	FullAdvertisementCarouselList datatypes.JSON `json:"fullAdvertisementCarouselList" gorm:"type:json"`
	NoticeCarouselList            datatypes.JSON `json:"noticeCarouselList" gorm:"type:json"`
	Status                        string         `json:"status" gorm:"-"` // 设备在线状态，不存储在数据库中
}

// OrangePiInfo 香橙派服务信息（包含打印机列表）
type OrangePiInfo struct {
	IP           *string   `json:"ip,omitempty" gorm:"column:orange_pi_ip;size:255"`
	Port         *int      `json:"port,omitempty" gorm:"column:orange_pi_port"`
	Status       string    `json:"status" gorm:"column:orange_pi_status;size:50;default:'not_configured'"` // online, offline, not_configured
	ResponseTime *int      `json:"response_time,omitempty" gorm:"column:orange_pi_response_time"`          // 响应时间（毫秒）
	Reason       *string   `json:"reason,omitempty" gorm:"column:orange_pi_reason;size:500"`
	ErrorCode    *string   `json:"error_code,omitempty" gorm:"column:orange_pi_error_code;size:100"`
	Printers     []Printer `json:"printers" gorm:"-"` // 打印机列表（运行时填充，不存储）
}

// DeviceSettings contains all settings for a device
type DeviceSettings struct {
	// Update durations (in minutes)
	ArrearageUpdateDuration     int `json:"arrearageUpdateDuration" gorm:"default:5"`      // 欠款更新时间间隔
	NoticeUpdateDuration        int `json:"noticeUpdateDuration" gorm:"default:10"`        // 通知更新时间间隔
	AdvertisementUpdateDuration int `json:"advertisementUpdateDuration" gorm:"default:15"` // 广告更新时间间隔

	AppUpdateDuration int `json:"appUpdateDuration" gorm:"default:600"` // 应用更新时间间隔
	// durations (in seconds)
	AdvertisementPlayDuration int `json:"advertisementPlayDuration" gorm:"default:30"` // 广告播放时间
	// NoticePlayDuration        int `json:"noticePlayDuration" gorm:"default:30"`        // 通知播放时间
	// SpareDuration             int `json:"spareDuration" gorm:"default:10"`             // 手动操作过时时间
	NoticeStayDuration int `json:"noticeStayDuration" gorm:"default:10"` // 通知停留时间

	BottomCarouselDuration                        int    `json:"bottomCarouselDuration" gorm:"default:10"`                        // 底部轮播切换时间
	PaymentTableOnePageDuration                   int    `json:"paymentTableOnePageDuration" gorm:"default:5"`                    // 缴费表格单页停留时间
	NormalToAnnouncementCarouselDuration          int    `json:"normalToAnnouncementCarouselDuration" gorm:"default:10"`          // 正常播放到公告轮播时间
	AnnouncementCarouselToFullAdsCarouselDuration int    `json:"announcementCarouselToFullAdsCarouselDuration" gorm:"default:10"` // 公告轮播到全屏广告轮播时间
	PrintPassWord                                 string `json:"printPassWord" gorm:"default:'1090119'"`                          // 打印密码
}
