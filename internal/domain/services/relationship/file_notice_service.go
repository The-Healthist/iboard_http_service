package http_relationship_service

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"gorm.io/gorm"
)

type InterfaceFileNoticeService interface {
	BindFile(noticeID uint, fileID uint) error
	UnbindFile(noticeID uint) error
	GetNoticeByFileID(fileID uint) (*base_models.Notice, error)
	GetFileByNoticeID(noticeID uint) (*base_models.File, error)
	NoticeExists(noticeID uint) (bool, error)
	FileExists(fileID uint) (bool, error)
}

type FileNoticeService struct {
	db *gorm.DB
}

func NewFileNoticeService(db *gorm.DB) InterfaceFileNoticeService {
	log.Debug("初始化FileNoticeService")
	return &FileNoticeService{db: db}
}

func (s *FileNoticeService) BindFile(noticeID uint, fileID uint) error {
	log.Info("绑定文件到通知 | 通知ID: %d | 文件ID: %d", noticeID, fileID)
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 检查通知是否存在
		exists, err := s.NoticeExists(noticeID)
		if err != nil {
			log.Error("检查通知存在性失败 | 通知ID: %d | 错误: %v", noticeID, err)
			return err
		}
		if !exists {
			log.Warn("通知不存在 | 通知ID: %d", noticeID)
			return errors.New("notice not found")
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

		// 更新通知的 FileID
		if err := tx.Model(&base_models.Notice{}).
			Where("id = ?", noticeID).
			Update("file_id", fileID).Error; err != nil {
			log.Error("绑定文件到通知失败 | 通知ID: %d | 文件ID: %d | 错误: %v", noticeID, fileID, err)
			return err
		}

		log.Info("成功绑定文件到通知 | 通知ID: %d | 文件ID: %d", noticeID, fileID)
		return nil
	})
}

func (s *FileNoticeService) UnbindFile(noticeID uint) error {
	log.Info("解绑通知的文件 | 通知ID: %d", noticeID)

	// 检查通知是否存在
	exists, err := s.NoticeExists(noticeID)
	if err != nil {
		log.Error("检查通知存在性失败 | 通知ID: %d | 错误: %v", noticeID, err)
		return err
	}
	if !exists {
		log.Warn("通知不存在 | 通知ID: %d", noticeID)
		return errors.New("notice not found")
	}

	// 获取当前绑定的文件ID用于日志记录
	file, err := s.GetFileByNoticeID(noticeID)
	var fileID uint
	if err == nil && file != nil {
		fileID = file.ID
		log.Debug("当前通知绑定的文件ID | 通知ID: %d | 文件ID: %d", noticeID, fileID)
	}

	if err := s.db.Model(&base_models.Notice{}).
		Where("id = ?", noticeID).
		Update("file_id", nil).Error; err != nil {
		log.Error("解绑通知文件失败 | 通知ID: %d | 错误: %v", noticeID, err)
		return err
	}

	log.Info("成功解绑通知文件 | 通知ID: %d | 原文件ID: %d", noticeID, fileID)
	return nil
}

func (s *FileNoticeService) GetNoticeByFileID(fileID uint) (*base_models.Notice, error) {
	log.Info("通过文件ID获取通知 | 文件ID: %d", fileID)
	var notice base_models.Notice
	if err := s.db.Where("file_id = ?", fileID).First(&notice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Debug("未找到使用该文件的通知 | 文件ID: %d", fileID)
			return nil, nil
		}
		log.Error("查询通知失败 | 文件ID: %d | 错误: %v", fileID, err)
		return nil, err
	}
	log.Debug("找到使用该文件的通知 | 文件ID: %d | 通知ID: %d", fileID, notice.ID)
	return &notice, nil
}

func (s *FileNoticeService) GetFileByNoticeID(noticeID uint) (*base_models.File, error) {
	log.Info("获取通知关联的文件 | 通知ID: %d", noticeID)
	var notice base_models.Notice
	if err := s.db.Preload("File").First(&notice, noticeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("通知不存在 | 通知ID: %d", noticeID)
			return nil, errors.New("notice not found")
		}
		log.Error("查询通知失败 | 通知ID: %d | 错误: %v", noticeID, err)
		return nil, err
	}

	if notice.File == nil {
		log.Debug("通知未关联文件 | 通知ID: %d", noticeID)
		return nil, nil
	}

	log.Debug("成功获取通知关联的文件 | 通知ID: %d | 文件ID: %d", noticeID, notice.File.ID)
	return notice.File, nil
}

func (s *FileNoticeService) NoticeExists(noticeID uint) (bool, error) {
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

func (s *FileNoticeService) FileExists(fileID uint) (bool, error) {
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
