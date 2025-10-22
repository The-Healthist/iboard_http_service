package base_services

import (
	"errors"

	"github.com/The-Healthist/iboard_http_service/internal/domain/models"
	"gorm.io/gorm"
)

type InterfacePrinterService interface {
	Create(printer *models.Printer) error
	Get(query map[string]interface{}, paginate map[string]interface{}) ([]models.Printer, models.PaginationResult, error)
	Update(id uint, updates map[string]interface{}) (*models.Printer, error)
	Delete(ids []uint) error
	GetByID(id uint) (*models.Printer, error)
	BatchUpdateOrCreate(deviceID uint, printers []models.Printer) error
	GetByDeviceID(deviceID uint) ([]models.Printer, error)
	SyncPrinters(deviceID uint, printers []models.Printer) error
}

type PrinterService struct {
	db *gorm.DB
}

func NewPrinterService(db *gorm.DB) InterfacePrinterService {
	return &PrinterService{
		db: db,
	}
}

func (s *PrinterService) Create(printer *models.Printer) error {
	return s.db.Create(printer).Error
}

func (s *PrinterService) Get(query map[string]interface{}, paginate map[string]interface{}) ([]models.Printer, models.PaginationResult, error) {
	var printers []models.Printer
	var total int64
	db := s.db.Model(&models.Printer{})

	if search, ok := query["search"].(string); ok && search != "" {
		db = db.Where("name LIKE ? OR ip_address LIKE ?", "%"+search+"%", "%"+search+"%")
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

	if err := db.Limit(pageSize).Offset(offset).Find(&printers).Error; err != nil {
		return nil, models.PaginationResult{}, err
	}

	return printers, models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *PrinterService) Update(id uint, updates map[string]interface{}) (*models.Printer, error) {
	var printer models.Printer
	if err := s.db.First(&printer, id).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&printer).Updates(updates).Error; err != nil {
		return nil, err
	}

	return &printer, nil
}

func (s *PrinterService) Delete(ids []uint) error {
	result := s.db.Delete(&models.Printer{}, ids)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no records found to delete")
	}
	return nil
}

func (s *PrinterService) GetByID(id uint) (*models.Printer, error) {
	var printer models.Printer
	if err := s.db.First(&printer, id).Error; err != nil {
		return nil, err
	}
	return &printer, nil
}

// GetByDeviceID 获取指定设备的所有打印机
func (s *PrinterService) GetByDeviceID(deviceID uint) ([]models.Printer, error) {
	var printers []models.Printer
	if err := s.db.Where("device_id = ?", deviceID).Find(&printers).Error; err != nil {
		return nil, err
	}
	return printers, nil
}

// BatchUpdateOrCreate 批量更新或创建打印机（用于健康检查和打印回调）
func (s *PrinterService) BatchUpdateOrCreate(deviceID uint, printers []models.Printer) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, printer := range printers {
			// 设置 deviceID
			printer.DeviceID = &deviceID

			// 根据 IP 地址或名称查找现有打印机
			var existing models.Printer
			err := tx.Where("device_id = ? AND (ip_address = ? OR name = ?)",
				deviceID, printer.IPAddress, printer.Name).First(&existing).Error

			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}

			if err == gorm.ErrRecordNotFound {
				// 创建新打印机
				if err := tx.Create(&printer).Error; err != nil {
					return err
				}
			} else {
				// 更新现有打印机
				updates := map[string]interface{}{}
				if printer.DisplayName != nil {
					updates["display_name"] = printer.DisplayName
				}
				if printer.IPAddress != nil {
					updates["ip_address"] = printer.IPAddress
				}
				if printer.Name != nil {
					updates["name"] = printer.Name
				}
				if printer.State != nil {
					updates["state"] = printer.State
				}
				if printer.URI != nil {
					updates["uri"] = printer.URI
				}
				if printer.Status != nil {
					updates["status"] = printer.Status
				}
				if printer.Reason != nil {
					updates["reason"] = printer.Reason
				}
				if printer.MarkerLevels != nil {
					updates["marker_levels"] = printer.MarkerLevels
				}

				if err := tx.Model(&existing).Updates(updates).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// SyncPrinters 同步打印机列表（健康检查专用）
// 与数据库比对：多的添加，少的删除，存在的更新
func (s *PrinterService) SyncPrinters(deviceID uint, printers []models.Printer) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 获取数据库中该设备的所有打印机
		var existingPrinters []models.Printer
		if err := tx.Where("device_id = ?", deviceID).Find(&existingPrinters).Error; err != nil {
			return err
		}

		// 2. 创建映射：用于快速查找
		// 使用 IP 地址作为唯一标识
		existingMap := make(map[string]models.Printer)
		for _, p := range existingPrinters {
			if p.IPAddress != nil {
				existingMap[*p.IPAddress] = p
			}
		}

		newMap := make(map[string]models.Printer)
		for _, p := range printers {
			if p.IPAddress != nil {
				newMap[*p.IPAddress] = p
			}
		}

		// 3. 找出要删除的打印机（数据库中有，但新列表中没有）
		var toDelete []uint
		for ip, existing := range existingMap {
			if _, found := newMap[ip]; !found {
				toDelete = append(toDelete, existing.ID)
			}
		}

		// 4. 删除不再存在的打印机
		if len(toDelete) > 0 {
			if err := tx.Delete(&models.Printer{}, toDelete).Error; err != nil {
				return err
			}
		}

		// 5. 添加或更新打印机
		for _, printer := range printers {
			if printer.IPAddress == nil {
				continue // 跳过没有 IP 地址的打印机
			}

			printer.DeviceID = &deviceID

			if existing, found := existingMap[*printer.IPAddress]; found {
				// 更新现有打印机
				updates := map[string]interface{}{}
				if printer.DisplayName != nil {
					updates["display_name"] = printer.DisplayName
				}
				if printer.Name != nil {
					updates["name"] = printer.Name
				}
				if printer.State != nil {
					updates["state"] = printer.State
				}
				if printer.URI != nil {
					updates["uri"] = printer.URI
				}
				if printer.Status != nil {
					updates["status"] = printer.Status
				}
				if printer.Reason != nil {
					updates["reason"] = printer.Reason
				}
				if printer.MarkerLevels != nil {
					updates["marker_levels"] = printer.MarkerLevels
				}

				if err := tx.Model(&existing).Updates(updates).Error; err != nil {
					return err
				}
			} else {
				// 创建新打印机
				if err := tx.Create(&printer).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}
