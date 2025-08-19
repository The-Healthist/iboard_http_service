package http_relationship_service

import (
	"encoding/json"
	"errors"
	"fmt"

	base_models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"github.com/The-Healthist/iboard_http_service/pkg/utils/field"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type InterfaceAdvertisementBuildingService interface {
	BindBuildings(advertisementID uint, buildingIDs []uint) error
	UnbindBuildings(advertisementID uint, buildingIDs []uint) error
	GetBuildingsByAdvertisementID(advertisementID uint) ([]base_models.Building, error)
	GetAdvertisementsByBuildingID(buildingID uint) ([]base_models.Advertisement, error)
	BulkCheckBuildingsExistence(buildingIDs []uint) ([]uint, error)
	BuildingExists(buildingID uint) (bool, error)
	AdvertisementExists(advertisementID uint) (bool, error)
}

type AdvertisementBuildingService struct {
	db *gorm.DB
}

func NewAdvertisementBuildingService(db *gorm.DB) InterfaceAdvertisementBuildingService {
	log.Info("初始化广告-建筑关系服务")
	return &AdvertisementBuildingService{db: db}
}

func (s *AdvertisementBuildingService) BindBuildings(advertisementID uint, buildingIDs []uint) error {
	log.Info("绑定建筑到广告 | 广告ID: %d | 建筑数量: %d", advertisementID, len(buildingIDs))
	return s.db.Transaction(func(tx *gorm.DB) error {
		// check if the advertisement exists
		exists, err := s.AdvertisementExists(advertisementID)
		if err != nil {
			log.Error("检查广告是否存在失败 | 广告ID: %d | 错误: %v", advertisementID, err)
			return err
		}
		if !exists {
			log.Warn("广告不存在 | 广告ID: %d", advertisementID)
			return errors.New("advertisement not found")
		}

		// check if all buildings exist
		missingIDs, err := s.BulkCheckBuildingsExistence(buildingIDs)
		if err != nil {
			log.Error("批量检查建筑存在性失败 | 错误: %v", err)
			return err
		}
		if len(missingIDs) > 0 {
			log.Warn("部分建筑不存在 | 缺失建筑IDs: %v", missingIDs)
			return errors.New("some buildings not found")
		}

		// 检查现有关系
		var existingRelations []struct {
			BuildingID uint `gorm:"column:building_id"`
		}
		if err := tx.Table("advertisement_buildings").
			Select("building_id").
			Where("advertisement_id = ? AND building_id IN ?", advertisementID, buildingIDs).
			Find(&existingRelations).Error; err != nil {
			log.Error("查询现有关系失败 | 广告ID: %d | 错误: %v", advertisementID, err)
			return err
		}

		// 过滤掉已存在的关系
		existingBuildingIDs := make(map[uint]bool)
		for _, relation := range existingRelations {
			existingBuildingIDs[relation.BuildingID] = true
		}

		var newBuildingIDs []uint
		for _, buildingID := range buildingIDs {
			if !existingBuildingIDs[buildingID] {
				newBuildingIDs = append(newBuildingIDs, buildingID)
			}
		}

		if len(newBuildingIDs) == 0 {
			log.Info("所有建筑都已绑定到该广告 | 广告ID: %d", advertisementID)
			return nil
		}

		// 创建新的绑定关系
		for _, buildingID := range newBuildingIDs {
			if err := tx.Exec(
				"INSERT INTO advertisement_buildings (advertisement_id, building_id) VALUES (?, ?)",
				advertisementID, buildingID,
			).Error; err != nil {
				log.Error("创建广告建筑关系失败 | 广告ID: %d | 建筑ID: %d | 错误: %v", advertisementID, buildingID, err)
				return err
			}
		}

		// 获取广告信息以确定其 display 类型
		var advertisement base_models.Advertisement
		if err := tx.First(&advertisement, advertisementID).Error; err != nil {
			log.Error("获取广告信息失败 | 广告ID: %d | 错误: %v", advertisementID, err)
			return err
		}

		// 同步更新相关 device 的轮播列表
		if err := s.syncDeviceCarouselLists(tx, advertisementID, newBuildingIDs, advertisement.Display, true); err != nil {
			log.Error("同步设备轮播列表失败 | 广告ID: %d | 错误: %v", advertisementID, err)
			return err
		}

		log.Info("成功绑定建筑到广告 | 广告ID: %d | 新绑定建筑数量: %d", advertisementID, len(newBuildingIDs))
		return nil
	})
}

func (s *AdvertisementBuildingService) UnbindBuildings(advertisementID uint, buildingIDs []uint) error {
	log.Info("解绑建筑与广告 | 广告ID: %d | 建筑数量: %d", advertisementID, len(buildingIDs))
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取广告信息以确定其 display 类型
		var advertisement base_models.Advertisement
		if err := tx.First(&advertisement, advertisementID).Error; err != nil {
			log.Error("获取广告信息失败 | 广告ID: %d | 错误: %v", advertisementID, err)
			return err
		}

		// 同步更新相关 device 的轮播列表（在删除关系之前）
		if err := s.syncDeviceCarouselLists(tx, advertisementID, buildingIDs, advertisement.Display, false); err != nil {
			log.Error("同步设备轮播列表失败 | 广告ID: %d | 错误: %v", advertisementID, err)
			return err
		}

		// 删除绑定关系
		if err := tx.Exec(
			"DELETE FROM advertisement_buildings WHERE advertisement_id = ? AND building_id IN ?",
			advertisementID, buildingIDs,
		).Error; err != nil {
			log.Error("删除广告建筑关系失败 | 广告ID: %d | 建筑IDs: %v | 错误: %v", advertisementID, buildingIDs, err)
			return err
		}

		log.Info("成功解绑建筑与广告 | 广告ID: %d | 解绑建筑数量: %d", advertisementID, len(buildingIDs))
		return nil
	})
}

func (s *AdvertisementBuildingService) GetBuildingsByAdvertisementID(advertisementID uint) ([]base_models.Building, error) {
	log.Info("获取广告关联的建筑 | 广告ID: %d", advertisementID)
	var advertisement base_models.Advertisement
	if err := s.db.Preload("Buildings").First(&advertisement, advertisementID).Error; err != nil {
		log.Error("查找广告失败 | 广告ID: %d | 错误: %v", advertisementID, err)
		return nil, err
	}
	if advertisement.Buildings == nil {
		log.Debug("广告没有关联的建筑 | 广告ID: %d", advertisementID)
		return []base_models.Building{}, nil
	}
	log.Debug("成功获取广告关联的建筑 | 广告ID: %d | 建筑数量: %d", advertisementID, len(advertisement.Buildings))
	return advertisement.Buildings, nil
}

func (s *AdvertisementBuildingService) GetAdvertisementsByBuildingID(buildingID uint) ([]base_models.Advertisement, error) {
	log.Info("获取建筑关联的广告 | 建筑ID: %d", buildingID)
	var building base_models.Building
	if err := s.db.
		Preload("Advertisements.File").
		Preload("Advertisements").
		First(&building, buildingID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("建筑不存在 | 建筑ID: %d", buildingID)
			return []base_models.Advertisement{}, nil
		}
		log.Error("查找建筑失败 | 建筑ID: %d | 错误: %v", buildingID, err)
		return nil, err
	}
	log.Debug("成功获取建筑关联的广告 | 建筑ID: %d | 广告数量: %d", buildingID, len(building.Advertisements))
	return building.Advertisements, nil
}

func (s *AdvertisementBuildingService) AdvertisementExists(advertisementID uint) (bool, error) {
	log.Debug("检查广告是否存在 | 广告ID: %d", advertisementID)
	var advertisement base_models.Advertisement
	err := s.db.First(&advertisement, advertisementID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Debug("广告不存在 | 广告ID: %d", advertisementID)
			return false, nil
		}
		log.Error("检查广告是否存在失败 | 广告ID: %d | 错误: %v", advertisementID, err)
		return false, err
	}
	return true, nil
}

func (s *AdvertisementBuildingService) BuildingExists(buildingID uint) (bool, error) {
	log.Debug("检查建筑是否存在 | 建筑ID: %d", buildingID)
	var building base_models.Building
	err := s.db.First(&building, buildingID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Debug("建筑不存在 | 建筑ID: %d", buildingID)
			return false, nil
		}
		log.Error("检查建筑是否存在失败 | 建筑ID: %d | 错误: %v", buildingID, err)
		return false, err
	}
	return true, nil
}

func (s *AdvertisementBuildingService) BulkCheckBuildingsExistence(buildingIDs []uint) ([]uint, error) {
	log.Debug("批量检查建筑是否存在 | 建筑IDs: %v", buildingIDs)
	var count int64
	err := s.db.Model(&base_models.Building{}).Where("id IN ?", buildingIDs).Count(&count).Error
	if err != nil {
		log.Error("批量检查建筑存在性失败 | 错误: %v", err)
		return nil, err
	}
	if int(count) == len(buildingIDs) {
		return []uint{}, nil
	}

	var existingIDs []uint
	err = s.db.Model(&base_models.Building{}).Where("id IN ?", buildingIDs).Pluck("id", &existingIDs).Error
	if err != nil {
		log.Error("获取存在的建筑ID失败 | 错误: %v", err)
		return nil, err
	}

	existingMap := make(map[uint]bool)
	for _, id := range existingIDs {
		existingMap[id] = true
	}

	var missingIDs []uint
	for _, id := range buildingIDs {
		if !existingMap[id] {
			missingIDs = append(missingIDs, id)
		}
	}

	if len(missingIDs) > 0 {
		log.Warn("以下建筑ID不存在: %v", missingIDs)
	}

	return missingIDs, nil
}

// syncDeviceCarouselLists 同步更新 device 的轮播列表
func (s *AdvertisementBuildingService) syncDeviceCarouselLists(tx *gorm.DB, advertisementID uint, buildingIDs []uint, display field.AdvertisementDisplay, isBind bool) error {
	// 获取这些建筑下的所有 device
	var devices []base_models.Device
	if err := tx.Where("building_id IN ?", buildingIDs).Find(&devices).Error; err != nil {
		return fmt.Errorf("failed to get devices: %v", err)
	}

	if len(devices) == 0 {
		log.Info("没有找到需要同步的设备 | 建筑IDs: %v", buildingIDs)
		return nil
	}

	for _, device := range devices {
		if err := s.updateDeviceCarouselList(tx, &device, advertisementID, display, isBind); err != nil {
			log.Error("更新设备轮播列表失败 | 设备ID: %d | 广告ID: %d | 错误: %v", device.ID, advertisementID, err)
			return err
		}
	}

	log.Info("成功同步设备轮播列表 | 广告ID: %d | 设备数量: %d | 操作: %s", advertisementID, len(devices), map[bool]string{true: "bind", false: "unbind"}[isBind])
	return nil
}

// updateDeviceCarouselList 更新单个设备的轮播列表
func (s *AdvertisementBuildingService) updateDeviceCarouselList(tx *gorm.DB, device *base_models.Device, advertisementID uint, display field.AdvertisementDisplay, isBind bool) error {
	var currentList []uint
	var err error

	// 根据 display 类型选择对应的轮播列表
	switch display {
	case field.AdDisplayTop:
		currentList, err = s.getCarouselListFromJSON(device.TopAdvertisementCarouselList)
	case field.AdDisplayFull:
		currentList, err = s.getCarouselListFromJSON(device.FullAdvertisementCarouselList)
	case field.AdDisplayTopFull:
		// topfull 类型需要同时更新两个列表
		if err := s.updateDeviceCarouselList(tx, device, advertisementID, field.AdDisplayTop, isBind); err != nil {
			return err
		}
		return s.updateDeviceCarouselList(tx, device, advertisementID, field.AdDisplayFull, isBind)
	default:
		log.Warn("未知的广告显示类型 | 类型: %s", display)
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to parse carousel list: %v", err)
	}

	// 更新列表
	if isBind {
		// 绑定：添加到列表末尾（如果不存在）
		if !s.containsID(currentList, advertisementID) {
			currentList = append(currentList, advertisementID)
		}
	} else {
		// 解绑：从列表中移除
		currentList = s.removeID(currentList, advertisementID)
	}

	// 保存更新后的列表
	if err := s.saveCarouselListToJSON(tx, device.ID, currentList, display); err != nil {
		return fmt.Errorf("failed to save carousel list: %v", err)
	}

	return nil
}

// getCarouselListFromJSON 从 JSON 字段解析轮播列表
func (s *AdvertisementBuildingService) getCarouselListFromJSON(jsonData datatypes.JSON) ([]uint, error) {
	if len(jsonData) == 0 {
		return []uint{}, nil
	}
	var list []uint
	if err := json.Unmarshal(jsonData, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// saveCarouselListToJSON 保存轮播列表到 JSON 字段
func (s *AdvertisementBuildingService) saveCarouselListToJSON(tx *gorm.DB, deviceID uint, list []uint, display field.AdvertisementDisplay) error {
	jsonData, err := json.Marshal(list)
	if err != nil {
		return err
	}

	var fieldName string
	switch display {
	case field.AdDisplayTop:
		fieldName = "top_advertisement_carousel_list"
	case field.AdDisplayFull:
		fieldName = "full_advertisement_carousel_list"
	default:
		return fmt.Errorf("unsupported display type: %s", display)
	}

	if err := tx.Model(&base_models.Device{}).Where("id = ?", deviceID).Update(fieldName, jsonData).Error; err != nil {
		return err
	}

	return nil
}

// containsID 检查列表中是否包含指定 ID
func (s *AdvertisementBuildingService) containsID(list []uint, id uint) bool {
	for _, item := range list {
		if item == id {
			return true
		}
	}
	return false
}

// removeID 从列表中移除指定 ID
func (s *AdvertisementBuildingService) removeID(list []uint, id uint) []uint {
	var result []uint
	for _, item := range list {
		if item != id {
			result = append(result, item)
		}
	}
	return result
}
