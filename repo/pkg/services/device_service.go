package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/mindflow/agri-platform/pkg/models"
	"gorm.io/gorm"
)

var (
	ErrDeviceNotFound  = errors.New("device not found")
	ErrDuplicateSerial = errors.New("serial number already exists")
	ErrDeviceForbidden = errors.New("not authorized to access this device")
)

type DeviceService struct {
	db *gorm.DB
}

func NewDeviceService(db *gorm.DB) *DeviceService {
	return &DeviceService{db: db}
}

type CreateDeviceInput struct {
	Name         string `json:"name"          binding:"required,max=200"`
	Type         string `json:"type"          binding:"required,max=100"`
	SerialNumber string `json:"serial_number" binding:"required,max=100"`
	PlotID       uint   `json:"plot_id"       binding:"required"`
	Status       string `json:"status"        binding:"omitempty,oneof=active inactive maintenance"`
}

type UpdateDeviceInput struct {
	Name   *string `json:"name"   binding:"omitempty,max=200"`
	Type   *string `json:"type"   binding:"omitempty,max=100"`
	Status *string `json:"status" binding:"omitempty,oneof=active inactive maintenance"`
	PlotID *uint   `json:"plot_id"`
}

func (s *DeviceService) Create(ctx context.Context, in CreateDeviceInput) (*models.Device, error) {
	status := in.Status
	if status == "" {
		status = "active"
	}

	device := models.Device{
		Name:         in.Name,
		Type:         in.Type,
		SerialNumber: in.SerialNumber,
		PlotID:       in.PlotID,
		Status:       status,
	}

	if err := s.db.WithContext(ctx).Create(&device).Error; err != nil {
		if isDuplicateEntry(err) {
			return nil, ErrDuplicateSerial
		}
		return nil, fmt.Errorf("create device: %w", err)
	}
	return &device, nil
}

type DeviceListParams struct {
	Page     int
	PageSize int
	PlotID   uint
	Status   string
	UserID   uint   // authenticated user for scoping
	Role     string // authenticated user role
}

// userOwnedPlotIDs returns the plot IDs owned by the given user.
func (s *DeviceService) userOwnedPlotIDs(ctx context.Context, userID uint) ([]uint, error) {
	var plotIDs []uint
	if err := s.db.WithContext(ctx).Model(&models.Plot{}).Where("user_id = ?", userID).Pluck("id", &plotIDs).Error; err != nil {
		return nil, fmt.Errorf("get user plots: %w", err)
	}
	return plotIDs, nil
}

// checkDeviceOwnership verifies that the user owns the plot the device belongs to.
func (s *DeviceService) checkDeviceOwnership(ctx context.Context, device *models.Device, userID uint, role string) error {
	if role == "admin" {
		return nil
	}
	var plot models.Plot
	if err := s.db.WithContext(ctx).First(&plot, device.PlotID).Error; err != nil {
		return ErrDeviceForbidden
	}
	if plot.UserID != userID {
		return ErrDeviceForbidden
	}
	return nil
}

type PaginatedDevices struct {
	Data       []models.Device `json:"data"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

func (s *DeviceService) List(ctx context.Context, p DeviceListParams) (*PaginatedDevices, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}

	q := s.db.WithContext(ctx).Model(&models.Device{})

	// Object-level isolation: non-admin users only see devices on their plots
	if p.Role != "admin" && p.UserID > 0 {
		plotIDs, err := s.userOwnedPlotIDs(ctx, p.UserID)
		if err != nil {
			return nil, err
		}
		if len(plotIDs) == 0 {
			return &PaginatedDevices{Data: []models.Device{}, Total: 0, Page: p.Page, PageSize: p.PageSize, TotalPages: 0}, nil
		}
		q = q.Where("plot_id IN ?", plotIDs)
	}

	if p.PlotID > 0 {
		q = q.Where("plot_id = ?", p.PlotID)
	}
	if p.Status != "" {
		q = q.Where("status = ?", p.Status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count devices: %w", err)
	}

	var devices []models.Device
	offset := (p.Page - 1) * p.PageSize
	if err := q.Order("id DESC").Offset(offset).Limit(p.PageSize).Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}

	totalPages := int(total) / p.PageSize
	if int(total)%p.PageSize != 0 {
		totalPages++
	}

	return &PaginatedDevices{
		Data:       devices,
		Total:      total,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *DeviceService) GetByID(ctx context.Context, id, userID uint, role string) (*models.Device, error) {
	var device models.Device
	if err := s.db.WithContext(ctx).First(&device, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeviceNotFound
		}
		return nil, fmt.Errorf("get device: %w", err)
	}
	if err := s.checkDeviceOwnership(ctx, &device, userID, role); err != nil {
		return nil, err
	}
	return &device, nil
}

func (s *DeviceService) Update(ctx context.Context, id, userID uint, role string, in UpdateDeviceInput) (*models.Device, error) {
	var device models.Device
	if err := s.db.WithContext(ctx).First(&device, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeviceNotFound
		}
		return nil, fmt.Errorf("find device: %w", err)
	}
	if err := s.checkDeviceOwnership(ctx, &device, userID, role); err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.Type != nil {
		updates["type"] = *in.Type
	}
	if in.Status != nil {
		updates["status"] = *in.Status
	}
	if in.PlotID != nil {
		updates["plot_id"] = *in.PlotID
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&device).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update device: %w", err)
		}
	}

	return &device, nil
}

func (s *DeviceService) Delete(ctx context.Context, id, userID uint, role string) error {
	var device models.Device
	if err := s.db.WithContext(ctx).First(&device, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDeviceNotFound
		}
		return fmt.Errorf("find device: %w", err)
	}
	if err := s.checkDeviceOwnership(ctx, &device, userID, role); err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Delete(&models.Device{}, id).Error; err != nil {
		return fmt.Errorf("delete device: %w", err)
	}
	return nil
}

// isDuplicateEntry checks if a MySQL error is a duplicate entry violation.
func isDuplicateEntry(err error) bool {
	return err != nil && (errors.As(err, new(interface{ Number() uint16 })) ||
		// Fallback: check error string for MySQL duplicate entry code 1062
		containsDuplicateMsg(err.Error()))
}

func containsDuplicateMsg(msg string) bool {
	return len(msg) > 0 && (contains(msg, "Duplicate entry") || contains(msg, "1062"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
