package http_relationship_service

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"gorm.io/gorm"
)

type InterfaceFileAdvertisementService interface {
	BindFile(advertisementID uint, fileID uint) error
	UnbindFile(advertisementID uint) error
	GetAdvertisementByFileID(fileID uint) (*base_models.Advertisement, error)
	GetFileByAdvertisementID(advertisementID uint) (*base_models.File, error)
	AdvertisementExists(advertisementID uint) (bool, error)
	FileExists(fileID uint) (bool, error)
}

type FileAdvertisementService struct {
	db *gorm.DB
}

func NewFileAdvertisementService(db *gorm.DB) InterfaceFileAdvertisementService {
	log.Debug("初始化FileAdvertisementService")
	return &FileAdvertisementService{db: db}
}

func (s *FileAdvertisementService) BindFile(advertisementID uint, fileID uint) error {
	log.Info("绑定文件到广告 | 广告ID: %d | 文件ID: %d", advertisementID, fileID)
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 检查广告是否存在
		exists, err := s.AdvertisementExists(advertisementID)
		if err != nil {
			log.Error("检查广告存在性失败 | 广告ID: %d | 错误: %v", advertisementID, err)
			return err
		}
		if !exists {
			log.Warn("广告不存在 | 广告ID: %d", advertisementID)
			return errors.New("advertisement not found")
		}

		// 检查文件是否存在
		exists, err = s.FileExists(fileID)
		if err != nil {
			log.Error("检查文件存在性失败 | 文件ID: %d | 错误: %v", fileID, err)
			return err
		}
		if !exists {
			log.Warn("文件不存在 | 文件ID: %d", fileID)
			return errors.New("file not found")
		}

		if err := tx.Model(&base_models.Advertisement{}).
			Where("id = ?", advertisementID).
			Update("file_id", fileID).Error; err != nil {
			log.Error("绑定文件到广告失败 | 广告ID: %d | 文件ID: %d | 错误: %v", advertisementID, fileID, err)
			return err
		}

		log.Info("成功绑定文件到广告 | 广告ID: %d | 文件ID: %d", advertisementID, fileID)
		return nil
	})
}

func (s *FileAdvertisementService) UnbindFile(advertisementID uint) error {
	log.Info("解绑广告的文件 | 广告ID: %d", advertisementID)

	// 检查广告是否存在
	exists, err := s.AdvertisementExists(advertisementID)
	if err != nil {
		log.Error("检查广告存在性失败 | 广告ID: %d | 错误: %v", advertisementID, err)
		return err
	}
	if !exists {
		log.Warn("广告不存在 | 广告ID: %d", advertisementID)
		return errors.New("advertisement not found")
	}

	// 获取当前绑定的文件ID用于日志记录
	ad, err := s.GetFileByAdvertisementID(advertisementID)
	var fileID uint
	if err == nil && ad != nil {
		fileID = ad.ID
		log.Debug("当前广告绑定的文件ID | 广告ID: %d | 文件ID: %d", advertisementID, fileID)
	}

	if err := s.db.Model(&base_models.Advertisement{}).
		Where("id = ?", advertisementID).
		Update("file_id", nil).Error; err != nil {
		log.Error("解绑广告文件失败 | 广告ID: %d | 错误: %v", advertisementID, err)
		return err
	}

	log.Info("成功解绑广告文件 | 广告ID: %d | 原文件ID: %d", advertisementID, fileID)
	return nil
}

func (s *FileAdvertisementService) GetAdvertisementByFileID(fileID uint) (*base_models.Advertisement, error) {
	log.Info("通过文件ID获取广告 | 文件ID: %d", fileID)
	var advertisement base_models.Advertisement
	if err := s.db.Where("file_id = ?", fileID).First(&advertisement).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Debug("未找到使用该文件的广告 | 文件ID: %d", fileID)
			return nil, nil
		}
		log.Error("查询广告失败 | 文件ID: %d | 错误: %v", fileID, err)
		return nil, err
	}
	log.Debug("找到使用该文件的广告 | 文件ID: %d | 广告ID: %d", fileID, advertisement.ID)
	return &advertisement, nil
}

func (s *FileAdvertisementService) GetFileByAdvertisementID(advertisementID uint) (*base_models.File, error) {
	log.Info("获取广告关联的文件 | 广告ID: %d", advertisementID)
	var advertisement base_models.Advertisement
	if err := s.db.Preload("File").First(&advertisement, advertisementID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("广告不存在 | 广告ID: %d", advertisementID)
			return nil, errors.New("advertisement not found")
		}
		log.Error("查询广告失败 | 广告ID: %d | 错误: %v", advertisementID, err)
		return nil, err
	}

	if advertisement.File == nil {
		log.Debug("广告未关联文件 | 广告ID: %d", advertisementID)
		return nil, nil
	}

	log.Debug("成功获取广告关联的文件 | 广告ID: %d | 文件ID: %d", advertisementID, advertisement.File.ID)
	return advertisement.File, nil
}

func (s *FileAdvertisementService) AdvertisementExists(advertisementID uint) (bool, error) {
	log.Debug("检查广告是否存在 | 广告ID: %d", advertisementID)
	var advertisement base_models.Advertisement
	err := s.db.First(&advertisement, advertisementID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Debug("广告不存在 | 广告ID: %d", advertisementID)
			return false, nil
		}
		log.Error("检查广告存在性失败 | 广告ID: %d | 错误: %v", advertisementID, err)
		return false, err
	}
	return true, nil
}

func (s *FileAdvertisementService) FileExists(fileID uint) (bool, error) {
	log.Debug("检查文件是否存在 | 文件ID: %d", fileID)
	var file base_models.File
	err := s.db.First(&file, fileID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Debug("文件不存在 | 文件ID: %d", fileID)
			return false, nil
		}
		log.Error("检查文件存在性失败 | 文件ID: %d | 错误: %v", fileID, err)
		return false, err
	}
	return true, nil
}
