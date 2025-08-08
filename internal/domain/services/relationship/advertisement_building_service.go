package http_relationship_service

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	"github.com/The-Healthist/iboard_http_service/pkg/log"
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

		// check if the relationship already exists
		for _, buildingID := range buildingIDs {
			var count int64
			if err := tx.Table("advertisement_buildings").
				Where("advertisement_id = ? AND building_id = ?", advertisementID, buildingID).
				Count(&count).Error; err != nil {
				log.Error("检查广告与建筑关系失败 | 广告ID: %d | 建筑ID: %d | 错误: %v", advertisementID, buildingID, err)
				return err
			}

			// if not exists, insert
			if count == 0 {
				if err := tx.Exec(
					"INSERT INTO advertisement_buildings (advertisement_id, building_id) VALUES (?, ?)",
					advertisementID, buildingID,
				).Error; err != nil {
					log.Error("插入广告与建筑关系失败 | 广告ID: %d | 建筑ID: %d | 错误: %v", advertisementID, buildingID, err)
					return err
				}
				log.Info("成功绑定广告与建筑 | 广告ID: %d | 建筑ID: %d", advertisementID, buildingID)
			} else {
				log.Debug("广告与建筑关系已存在 | 广告ID: %d | 建筑ID: %d", advertisementID, buildingID)
			}
		}

		log.Info("成功完成广告与建筑绑定 | 广告ID: %d | 建筑数量: %d", advertisementID, len(buildingIDs))
		return nil
	})
}

func (s *AdvertisementBuildingService) UnbindBuildings(advertisementID uint, buildingIDs []uint) error {
	log.Info("解绑广告与多个建筑 | 广告ID: %d | 建筑数量: %d", advertisementID, len(buildingIDs))
	return s.db.Transaction(func(tx *gorm.DB) error {
		var advertisement base_models.Advertisement
		if err := tx.Preload("Buildings").First(&advertisement, advertisementID).Error; err != nil {
			log.Error("查找广告失败 | 广告ID: %d | 错误: %v", advertisementID, err)
			return err
		}

		var buildings []base_models.Building
		if err := tx.Find(&buildings, buildingIDs).Error; err != nil {
			log.Error("查找建筑失败 | 建筑IDs: %v | 错误: %v", buildingIDs, err)
			return err
		}

		if err := tx.Model(&advertisement).Association("Buildings").Delete(&buildings); err != nil {
			log.Error("解除广告与建筑关联失败 | 广告ID: %d | 错误: %v", advertisementID, err)
			return err
		}

		log.Info("成功解绑广告与多个建筑 | 广告ID: %d | 建筑数量: %d", advertisementID, len(buildingIDs))
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
