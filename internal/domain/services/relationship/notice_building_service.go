package http_relationship_service

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"gorm.io/gorm"
)

type InterfaceNoticeBuildingService interface {
	BindBuildings(noticeID uint, buildingIDs []uint) error
	UnbindBuildings(noticeID uint, buildingIDs []uint) error
	GetBuildingsByNoticeID(noticeID uint) ([]base_models.Building, error)
	GetNoticesByBuildingID(buildingID uint) ([]base_models.Notice, error)
	BulkCheckBuildingsExistence(buildingIDs []uint) ([]uint, error)
	BuildingExists(buildingID uint) (bool, error)
	NoticeExists(noticeID uint) (bool, error)
}

type NoticeBuildingService struct {
	db *gorm.DB
}

func NewNoticeBuildingService(db *gorm.DB) InterfaceNoticeBuildingService {
	log.Debug("初始化NoticeBuildingService")
	return &NoticeBuildingService{db: db}
}

func (s *NoticeBuildingService) BindBuildings(noticeID uint, buildingIDs []uint) error {
	log.Info("绑定通知到建筑 | 通知ID: %d | 建筑IDs: %v", noticeID, buildingIDs)
	return s.db.Transaction(func(tx *gorm.DB) error {
		var notice base_models.Notice
		if err := tx.Preload("Buildings").First(&notice, noticeID).Error; err != nil {
			log.Warn("通知不存在 | 通知ID: %d | 错误: %v", noticeID, err)
			return err
		}

		var buildings []base_models.Building
		if err := tx.Find(&buildings, buildingIDs).Error; err != nil {
			log.Error("查询建筑失败 | 建筑IDs: %v | 错误: %v", buildingIDs, err)
			return err
		}

		if len(buildings) != len(buildingIDs) {
			log.Warn("部分建筑未找到 | 请求建筑数: %d | 找到建筑数: %d", len(buildingIDs), len(buildings))
		}

		if err := tx.Model(&notice).Association("Buildings").Append(&buildings); err != nil {
			log.Error("绑定通知到建筑失败 | 通知ID: %d | 建筑IDs: %v | 错误: %v", noticeID, buildingIDs, err)
			return err
		}

		log.Info("成功绑定通知到建筑 | 通知ID: %d | 建筑数量: %d", noticeID, len(buildings))
		return nil
	})
}

func (s *NoticeBuildingService) UnbindBuildings(noticeID uint, buildingIDs []uint) error {
	log.Info("解绑通知与建筑 | 通知ID: %d | 建筑IDs: %v", noticeID, buildingIDs)
	return s.db.Transaction(func(tx *gorm.DB) error {
		var notice base_models.Notice
		if err := tx.Preload("Buildings").First(&notice, noticeID).Error; err != nil {
			log.Warn("通知不存在 | 通知ID: %d | 错误: %v", noticeID, err)
			return err
		}

		var buildings []base_models.Building
		if err := tx.Find(&buildings, buildingIDs).Error; err != nil {
			log.Error("查询建筑失败 | 建筑IDs: %v | 错误: %v", buildingIDs, err)
			return err
		}

		if len(buildings) != len(buildingIDs) {
			log.Warn("部分建筑未找到 | 请求建筑数: %d | 找到建筑数: %d", len(buildingIDs), len(buildings))
		}

		if err := tx.Model(&notice).Association("Buildings").Delete(&buildings); err != nil {
			log.Error("解绑通知与建筑失败 | 通知ID: %d | 建筑IDs: %v | 错误: %v", noticeID, buildingIDs, err)
			return err
		}

		log.Info("成功解绑通知与建筑 | 通知ID: %d | 建筑数量: %d", noticeID, len(buildings))
		return nil
	})
}

func (s *NoticeBuildingService) GetBuildingsByNoticeID(noticeID uint) ([]base_models.Building, error) {
	log.Info("获取通知关联的建筑 | 通知ID: %d", noticeID)
	var notice base_models.Notice
	if err := s.db.Preload("Buildings").First(&notice, noticeID).Error; err != nil {
		log.Warn("通知不存在 | 通知ID: %d | 错误: %v", noticeID, err)
		return nil, err
	}
	if notice.Buildings == nil {
		log.Debug("通知未关联任何建筑 | 通知ID: %d", noticeID)
		return []base_models.Building{}, nil
	}
	log.Info("成功获取通知关联的建筑 | 通知ID: %d | 建筑数量: %d", noticeID, len(notice.Buildings))
	return notice.Buildings, nil
}

func (s *NoticeBuildingService) GetNoticesByBuildingID(buildingID uint) ([]base_models.Notice, error) {
	log.Info("获取建筑关联的通知 | 建筑ID: %d", buildingID)
	var building base_models.Building
	if err := s.db.
		Preload("Notices.File").
		Preload("Notices").
		First(&building, buildingID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("建筑不存在 | 建筑ID: %d", buildingID)
			return []base_models.Notice{}, nil
		}
		log.Error("查询建筑失败 | 建筑ID: %d | 错误: %v", buildingID, err)
		return nil, err
	}
	log.Info("成功获取建筑关联的通知 | 建筑ID: %d | 通知数量: %d", buildingID, len(building.Notices))
	return building.Notices, nil
}

func (s *NoticeBuildingService) NoticeExists(noticeID uint) (bool, error) {
	log.Debug("检查通知是否存在 | 通知ID: %d", noticeID)
	var notice base_models.Notice
	err := s.db.First(&notice, noticeID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Debug("通知不存在 | 通知ID: %d", noticeID)
			return false, nil
		}
		log.Error("检查通知存在性失败 | 通知ID: %d | 错误: %v", noticeID, err)
		return false, err
	}
	return true, nil
}

func (s *NoticeBuildingService) BuildingExists(buildingID uint) (bool, error) {
	log.Debug("检查建筑是否存在 | 建筑ID: %d", buildingID)
	var building base_models.Building
	err := s.db.First(&building, buildingID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Debug("建筑不存在 | 建筑ID: %d", buildingID)
			return false, nil
		}
		log.Error("检查建筑存在性失败 | 建筑ID: %d | 错误: %v", buildingID, err)
		return false, err
	}
	return true, nil
}

func (s *NoticeBuildingService) BulkCheckBuildingsExistence(buildingIDs []uint) ([]uint, error) {
	log.Debug("批量检查建筑存在性 | 建筑IDs: %v", buildingIDs)
	var count int64
	err := s.db.Model(&base_models.Building{}).Where("id IN ?", buildingIDs).Count(&count).Error
	if err != nil {
		log.Error("批量检查建筑存在性失败 | 错误: %v", err)
		return nil, err
	}
	if int(count) == len(buildingIDs) {
		log.Debug("所有建筑都存在 | 建筑数量: %d", len(buildingIDs))
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

	log.Warn("部分建筑不存在 | 请求建筑数: %d | 缺失建筑数: %d | 缺失IDs: %v",
		len(buildingIDs), len(missingIDs), missingIDs)
	return missingIDs, nil
}
