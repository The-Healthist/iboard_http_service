package base_services

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"gorm.io/gorm"
)

type InterfaceNoticeService interface {
	Create(notice *base_models.Notice) error
	Get(query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Notice, base_models.PaginationResult, error)
	Update(id uint, updates map[string]interface{}) (*base_models.Notice, error)
	Delete(ids []uint) error
	GetByID(id uint) (*base_models.Notice, error)
}

type NoticeService struct {
	db *gorm.DB
}

func NewNoticeService(db *gorm.DB) InterfaceNoticeService {
	return &NoticeService{db: db}
}

func (s *NoticeService) Create(notice *base_models.Notice) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(notice).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *NoticeService) Get(query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Notice, base_models.PaginationResult, error) {
	var notices []base_models.Notice
	var total int64
	db := s.db.Model(&base_models.Notice{}).
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
		return nil, base_models.PaginationResult{}, err
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
		return nil, base_models.PaginationResult{}, err
	}

	return notices, base_models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *NoticeService) Update(id uint, updates map[string]interface{}) (*base_models.Notice, error) {
	var notice base_models.Notice

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&base_models.Notice{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}

		if err := tx.Preload("File").Preload("Buildings").First(&notice, id).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &notice, nil
}

func (s *NoticeService) Delete(ids []uint) error {
	result := s.db.Delete(&base_models.Notice{}, ids)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no records found to delete")
	}
	return nil
}

func (s *NoticeService) GetByID(id uint) (*base_models.Notice, error) {
	var notice base_models.Notice
	if err := s.db.Preload("File").Preload("Buildings").First(&notice, id).Error; err != nil {
		return nil, err
	}
	return &notice, nil
}
