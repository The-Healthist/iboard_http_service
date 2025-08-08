package http_relationship_service

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"gorm.io/gorm"
)

type InterfaceDeviceBuildingService interface {
	BindDevices(buildingID uint, deviceIDs []uint) error
	UnbindDevice(deviceID uint) error
	GetDevicesByBuilding(buildingID uint, query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Device, base_models.PaginationResult, error)
	GetBuildingByDevice(deviceID uint) (*base_models.Building, error)
}

type DeviceBuildingService struct {
	db *gorm.DB
}

func NewDeviceBuildingService(db *gorm.DB) InterfaceDeviceBuildingService {
	log.Debug("初始化DeviceBuildingService")
	return &DeviceBuildingService{db: db}
}

func (s *DeviceBuildingService) BindDevices(buildingID uint, deviceIDs []uint) error {
	log.Info("绑定设备到建筑 | 建筑ID: %d | 设备IDs: %v", buildingID, deviceIDs)
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Check if building exists
		var building base_models.Building
		if err := tx.First(&building, buildingID).Error; err != nil {
			log.Warn("建筑不存在 | 建筑ID: %d | 错误: %v", buildingID, err)
			return errors.New("building not found")
		}

		// Check if all devices exist and are not bound to other buildings
		var devices []base_models.Device
		if err := tx.Where("id IN ?", deviceIDs).Find(&devices).Error; err != nil {
			log.Error("查询设备失败 | 设备IDs: %v | 错误: %v", deviceIDs, err)
			return err
		}

		if len(devices) != len(deviceIDs) {
			log.Warn("部分设备未找到 | 请求设备数: %d | 找到设备数: %d", len(deviceIDs), len(devices))
			return errors.New("one or more devices not found")
		}

		// Check if any device is already bound to another building
		for _, device := range devices {
			if device.BuildingID != 0 && device.BuildingID != buildingID {
				log.Warn("设备已绑定到其他建筑 | 设备ID: %d | 当前绑定建筑ID: %d | 目标建筑ID: %d",
					device.ID, device.BuildingID, buildingID)
				return errors.New("one or more devices are already bound to another building")
			}
		}

		// Update all devices' building ID
		if err := tx.Model(&base_models.Device{}).Where("id IN ?", deviceIDs).Update("building_id", buildingID).Error; err != nil {
			log.Error("更新设备建筑ID失败 | 建筑ID: %d | 设备IDs: %v | 错误: %v", buildingID, deviceIDs, err)
			return err
		}

		log.Info("成功绑定设备到建筑 | 建筑ID: %d | 设备数量: %d", buildingID, len(deviceIDs))
		return nil
	})
}

func (s *DeviceBuildingService) UnbindDevice(deviceID uint) error {
	log.Info("解绑设备 | 设备ID: %d", deviceID)
	// Check if device exists
	var device base_models.Device
	if err := s.db.First(&device, deviceID).Error; err != nil {
		log.Warn("设备不存在 | 设备ID: %d | 错误: %v", deviceID, err)
		return errors.New("device not found")
	}

	// Set building_id to null
	if err := s.db.Model(&device).Update("building_id", nil).Error; err != nil {
		log.Error("解绑设备失败 | 设备ID: %d | 错误: %v", deviceID, err)
		return err
	}

	log.Info("成功解绑设备 | 设备ID: %d | 原建筑ID: %d", deviceID, device.BuildingID)
	return nil
}

func (s *DeviceBuildingService) GetDevicesByBuilding(buildingID uint, query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Device, base_models.PaginationResult, error) {
	log.Info("获取建筑的设备列表 | 建筑ID: %d", buildingID)
	var devices []base_models.Device
	var total int64
	db := s.db.Model(&base_models.Device{}).Where("building_id = ?", buildingID)

	if search, ok := query["search"].(string); ok && search != "" {
		db = db.Where("device_id LIKE ?", "%"+search+"%")
		log.Debug("应用搜索过滤 | 关键词: %s", search)
	}

	if err := db.Count(&total).Error; err != nil {
		log.Error("获取设备总数失败 | 建筑ID: %d | 错误: %v", buildingID, err)
		return nil, base_models.PaginationResult{}, err
	}

	pageSize := paginate["pageSize"].(int)
	pageNum := paginate["pageNum"].(int)
	offset := (pageNum - 1) * pageSize
	log.Debug("应用分页 | 页码: %d | 每页数量: %d | 总数: %d", pageNum, pageSize, total)

	if desc, ok := paginate["desc"].(bool); ok && desc {
		db = db.Order("created_at DESC")
	} else {
		db = db.Order("created_at ASC")
	}

	if err := db.Limit(pageSize).Offset(offset).Find(&devices).Error; err != nil {
		log.Error("查询设备列表失败 | 建筑ID: %d | 错误: %v", buildingID, err)
		return nil, base_models.PaginationResult{}, err
	}

	log.Info("成功获取设备列表 | 建筑ID: %d | 设备数量: %d", buildingID, len(devices))
	return devices, base_models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *DeviceBuildingService) GetBuildingByDevice(deviceID uint) (*base_models.Building, error) {
	log.Info("获取设备所属建筑 | 设备ID: %d", deviceID)
	var device base_models.Device
	if err := s.db.Preload("Building").First(&device, deviceID).Error; err != nil {
		log.Warn("设备不存在 | 设备ID: %d | 错误: %v", deviceID, err)
		return nil, errors.New("device not found")
	}

	if device.BuildingID == 0 {
		log.Warn("设备未绑定到任何建筑 | 设备ID: %d", deviceID)
		return nil, errors.New("device is not bound to any building")
	}

	log.Info("成功获取设备所属建筑 | 设备ID: %d | 建筑ID: %d | 建筑名称: %s",
		deviceID, device.BuildingID, device.Building.Name)
	return &device.Building, nil
}
