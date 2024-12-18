package services

import (
	"errors"

	"github.com/The-Healthist/iboard_http_service/models"
	"gorm.io/gorm"
)

type InterfaceNoticeService interface {
	Create(notice *models.Notice, buildingIDs []uint) error
	Get(query map[string]interface{}, paginate map[string]interface{}) ([]models.Notice, models.PaginationResult, error)
	Update(id uint, updates map[string]interface{}, buildingIDs []uint) error
	Delete(ids []uint) error
}

type NoticeService struct {
	db *gorm.DB
}

func NewNoticeService(db *gorm.DB) InterfaceNoticeService {
	return &NoticeService{db: db}
}

func (s *NoticeService) Create(notice *models.Notice, buildingIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(notice).Error; err != nil {
			return err
		}

		if len(buildingIDs) > 0 {
			var buildings []models.Building
			if err := tx.Where("id IN ?", buildingIDs).Find(&buildings).Error; err != nil {
				return err
			}
			if err := tx.Model(notice).Association("Buildings").Replace(buildings); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *NoticeService) Get(query map[string]interface{}, paginate map[string]interface{}) ([]models.Notice, models.PaginationResult, error) {
	var notices []models.Notice
	var total int64
	db := s.db.Model(&models.Notice{}).
		Preload("Buildings").
		Preload("File")

	if search, ok := query["search"].(string); ok && search != "" {
		db = db.Where("title LIKE ? OR description LIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	if noticeType, ok := query["type"].(string); ok && noticeType != "" {
		db = db.Where("type = ?", noticeType)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, models.PaginationResult{}, err
	}

	pageSize := paginate["pageSize"].(int)
	pageNum := paginate["pageNum"].(int)
	offset := (pageNum - 1) * pageSize

	if desc, ok := paginate["desc"].(bool); ok && desc {
		db = db.Order("created_at DESC")
	} else {
		db = db.Order("created_at ASC")
	}

	if err := db.Limit(pageSize).Offset(offset).Find(&notices).Error; err != nil {
		return nil, models.PaginationResult{}, err
	}

	return notices, models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *NoticeService) Update(id uint, updates map[string]interface{}, buildingIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		notice := &models.Notice{}
		if err := tx.First(notice, id).Error; err != nil {
			return err
		}

		if err := tx.Model(notice).Updates(updates).Error; err != nil {
			return err
		}

		if buildingIDs != nil {
			var buildings []models.Building
			if err := tx.Where("id IN ?", buildingIDs).Find(&buildings).Error; err != nil {
				return err
			}
			if err := tx.Model(notice).Association("Buildings").Replace(buildings); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *NoticeService) Delete(ids []uint) error {
	result := s.db.Delete(&models.Notice{}, ids)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no records found to delete")
	}
	return nil
}
