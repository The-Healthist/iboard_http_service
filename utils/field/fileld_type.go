package field

// file uploader type
type FileUploaderType string

const (
	UploaderTypeUser       FileUploaderType = "user"
	UploaderTypeSuperAdmin FileUploaderType = "superadmin"
)

// advertisement type
type AdvertisementType string

const (
	AdTypeVideo AdvertisementType = "video"
	AdTypeImage AdvertisementType = "image"
)

// advertisement display type
type AdvertisementDisplay string

const (
	AdDisplayFull    AdvertisementDisplay = "full"
	AdDisplayTop     AdvertisementDisplay = "top"
	AdDisplayTopFull AdvertisementDisplay = "topfull"
)

// notice type
type NoticeType string

const (
	NoticeTypeUrgent     NoticeType = "urgent"
	NoticeTypeCommon     NoticeType = "common"
	NoticeTypeSystem     NoticeType = "system"
	NoticeTypeGovernment NoticeType = "government"
)

// file type
type FileType string

const (
	FileTypePdf FileType = "pdf"
)

// validate method
func IsValidFileUploaderType(t string) bool {
	switch FileUploaderType(t) {
	case UploaderTypeUser, UploaderTypeSuperAdmin:
		return true
	}
	return false
}

func IsValidAdvertisementType(t string) bool {
	switch AdvertisementType(t) {
	case AdTypeVideo, AdTypeImage:
		return true
	}
	return false
}

func IsValidAdvertisementDisplay(d string) bool {
	switch AdvertisementDisplay(d) {
	case AdDisplayFull, AdDisplayTop, AdDisplayTopFull:
		return true
	}
	return false
}

func IsValidNoticeType(t string) bool {
	switch NoticeType(t) {
	case NoticeTypeUrgent, NoticeTypeCommon, NoticeTypeSystem, NoticeTypeGovernment:
		return true
	}
	return false
}
