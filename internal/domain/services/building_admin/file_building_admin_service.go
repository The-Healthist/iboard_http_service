package building_admin_services

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"gorm.io/gorm"
)

type InterfaceBuildingAdminFileService interface {
	Create(file *base_models.File, email string) error
	Get(email string, query map[string]interface{}, paginate map[string]interface{}) ([]base_models.File, base_models.PaginationResult, error)
	Update(id uint, email string, updates map[string]interface{}) error
	Delete(id uint, email string) error
	GetByID(id uint, email string) (*base_models.File, error)
}

type BuildingAdminFileService struct {
	db *gorm.DB
}

func NewBuildingAdminFileService(db *gorm.DB) InterfaceBuildingAdminFileService {
	return &BuildingAdminFileService{db: db}
}

func (s *BuildingAdminFileService) Create(file *base_models.File, email string) error {
	file.Uploader = email
	file.UploaderType = "building_admin"
	return s.db.Create(file).Error
}

func (s *BuildingAdminFileService) Get(email string, query map[string]interface{}, paginate map[string]interface{}) ([]base_models.File, base_models.PaginationResult, error) {
	var files []base_models.File
	var total int64

	// 首先获取 building admin 关联的所有 building IDs
	var buildingAdmin base_models.BuildingAdmin
	if err := s.db.Where("email = ?", email).Preload("Buildings").First(&buildingAdmin).Error; err != nil {
		log.Error("获取建筑管理员信息失败: %v", err)
		return nil, base_models.PaginationResult{}, err
	}

	var buildingIDs []uint
	for _, building := range buildingAdmin.Buildings {
		buildingIDs = append(buildingIDs, building.ID)
	}
	log.Debug("建筑管理员(%s)关联的建筑IDs: %v", email, buildingIDs)

	// 获取广告相关的文件 IDs
	var adFileIDs []uint
	adQuery := s.db.Table("advertisements").
		Select("DISTINCT file_id").
		Joins("JOIN advertisement_buildings ON advertisements.id = advertisement_buildings.advertisement_id").
		Where("advertisement_buildings.building_id IN ?", buildingIDs).
		Where("file_id IS NOT NULL")
	if err := adQuery.Pluck("file_id", &adFileIDs).Error; err != nil {
		log.Error("获取广告相关文件ID失败: %v", err)
		return nil, base_models.PaginationResult{}, err
	}

	// 获取通知相关的文件 IDs
	var noticeFileIDs []uint
	noticeQuery := s.db.Table("notices").
		Select("DISTINCT file_id").
		Joins("JOIN notice_buildings ON notices.id = notice_buildings.notice_id").
		Where("notice_buildings.building_id IN ?", buildingIDs).
		Where("file_id IS NOT NULL")
	if err := noticeQuery.Pluck("file_id", &noticeFileIDs).Error; err != nil {
		log.Error("获取通知相关文件ID失败: %v", err)
		return nil, base_models.PaginationResult{}, err
	}

	// 合并所有文件 IDs
	allFileIDs := append(adFileIDs, noticeFileIDs...)
	log.Debug("广告文件IDs: %v, 通知文件IDs: %v, 合并后: %v", adFileIDs, noticeFileIDs, allFileIDs)

	// 构建主查询
	db := s.db.Model(&base_models.File{}).
		Where(
			s.db.Where("(uploader = ? AND uploader_type = ?)", email, "building_admin").
				Or("id IN ?", allFileIDs),
		)

	// 应用过滤条件
	if mimeType, ok := query["mimeType"].(string); ok && mimeType != "" {
		db = db.Where("mime_type = ?", mimeType)
		log.Debug("应用MIME类型过滤: %s", mimeType)
	}

	if oss, ok := query["oss"].(string); ok && oss != "" {
		db = db.Where("oss = ?", oss)
		log.Debug("应用OSS过滤: %s", oss)
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		log.Error("获取文件总数失败: %v", err)
		return nil, base_models.PaginationResult{}, err
	}

	// 应用分页
	pageSize := paginate["pageSize"].(int)
	pageNum := paginate["pageNum"].(int)
	offset := (pageNum - 1) * pageSize

	if desc, ok := paginate["desc"].(bool); ok && desc {
		db = db.Order("created_at DESC")
	} else {
		db = db.Order("created_at ASC")
	}

	// 使用日志记录SQL语句
	log.Debug("执行文件查询 | 管理员: %s | 页码: %d | 每页数量: %d | 总数: %d",
		email, pageNum, pageSize, total)

	if err := db.Limit(pageSize).Offset(offset).Find(&files).Error; err != nil {
		log.Error("查询文件列表失败: %v", err)
		return nil, base_models.PaginationResult{}, err
	}

	log.Info("成功获取文件列表 | 管理员: %s | 文件数量: %d", email, len(files))
	return files, base_models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *BuildingAdminFileService) Update(id uint, email string, updates map[string]interface{}) error {
	result := s.db.Model(&base_models.File{}).
		Where("id = ? AND uploader = ? AND uploader_type = ?", id, email, "building_admin").
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("file not found or no permission")
	}
	return nil
}

func (s *BuildingAdminFileService) Delete(id uint, email string) error {
	result := s.db.Where("id = ? AND uploader = ? AND uploader_type = ?", id, email, "building_admin").
		Delete(&base_models.File{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("file not found or no permission")
	}
	return nil
}

func (s *BuildingAdminFileService) GetByID(id uint, email string) (*base_models.File, error) {
	var file base_models.File

	// 首先获取 building admin 关联的所有 building IDs
	var buildingAdmin base_models.BuildingAdmin
	if err := s.db.Where("email = ?", email).Preload("Buildings").First(&buildingAdmin).Error; err != nil {
		log.Error("获取建筑管理员信息失败 | 邮箱: %s | 错误: %v", email, err)
		return nil, err
	}

	var buildingIDs []uint
	for _, building := range buildingAdmin.Buildings {
		buildingIDs = append(buildingIDs, building.ID)
	}
	log.Debug("建筑管理员(%s)关联的建筑IDs: %v", email, buildingIDs)

	// 获取广告相关的文件 IDs
	var adFileIDs []uint
	adQuery := s.db.Table("advertisements").
		Select("DISTINCT file_id").
		Joins("JOIN advertisement_buildings ON advertisements.id = advertisement_buildings.advertisement_id").
		Where("advertisement_buildings.building_id IN ?", buildingIDs).
		Where("file_id IS NOT NULL")
	if err := adQuery.Pluck("file_id", &adFileIDs).Error; err != nil {
		log.Error("获取广告相关文件ID失败: %v", err)
		return nil, err
	}

	// 获取通知相关的文件 IDs
	var noticeFileIDs []uint
	noticeQuery := s.db.Table("notices").
		Select("DISTINCT file_id").
		Joins("JOIN notice_buildings ON notices.id = notice_buildings.notice_id").
		Where("notice_buildings.building_id IN ?", buildingIDs).
		Where("file_id IS NOT NULL")
	if err := noticeQuery.Pluck("file_id", &noticeFileIDs).Error; err != nil {
		log.Error("获取通知相关文件ID失败: %v", err)
		return nil, err
	}

	// 合并所有文件 IDs
	allFileIDs := append(adFileIDs, noticeFileIDs...)
	log.Debug("查询文件ID: %d | 广告文件IDs: %v | 通知文件IDs: %v", id, adFileIDs, noticeFileIDs)

	// 使用日志记录SQL查询
	log.Debug("查询文件详情 | 文件ID: %d | 管理员: %s", id, email)
	db := s.db.Model(&base_models.File{}).
		Where("id = ?", id).
		Where(
			s.db.Where("(uploader = ? AND uploader_type = ?)", email, "building_admin").
				Or("id IN ?", allFileIDs),
		)

	if err := db.First(&file).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn("文件未找到或无权限 | 文件ID: %d | 管理员: %s", id, email)
			return nil, errors.New("file not found or no permission")
		}
		log.Error("查询文件详情失败 | 文件ID: %d | 错误: %v", id, err)
		return nil, err
	}

	log.Info("成功获取文件详情 | 文件ID: %d | 管理员: %s", id, email)
	return &file, nil
}
