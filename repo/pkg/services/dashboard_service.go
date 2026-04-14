package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/mindflow/agri-platform/pkg/models"
	"gorm.io/gorm"
)

var (
	ErrDashboardNotFound = errors.New("dashboard config not found")
	ErrDashboardForbidden = errors.New("not authorized to access this dashboard")
)

type DashboardService struct {
	db *gorm.DB
}

func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{db: db}
}

type CreateDashboardInput struct {
	Name   string `json:"name"   binding:"required,max=255"`
	Config string `json:"config" binding:"required"` // JSON string
}

type UpdateDashboardInput struct {
	Name   string `json:"name"   binding:"max=255"`
	Config string `json:"config"`
}

func (s *DashboardService) Create(ctx context.Context, userID uint, in CreateDashboardInput) (*models.DashboardConfig, error) {
	cfg := models.DashboardConfig{
		UserID: userID,
		Name:   in.Name,
		Config: in.Config,
	}
	if err := s.db.WithContext(ctx).Create(&cfg).Error; err != nil {
		return nil, fmt.Errorf("create dashboard config: %w", err)
	}
	return &cfg, nil
}

type DashboardListParams struct {
	Page     int
	PageSize int
	UserID   uint
}

type PaginatedDashboards struct {
	Data       []models.DashboardConfig `json:"data"`
	Total      int64                    `json:"total"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	TotalPages int                      `json:"total_pages"`
}

func (s *DashboardService) List(ctx context.Context, p DashboardListParams) (*PaginatedDashboards, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}

	q := s.db.WithContext(ctx).Model(&models.DashboardConfig{})
	if p.UserID > 0 {
		q = q.Where("user_id = ?", p.UserID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count dashboards: %w", err)
	}

	var configs []models.DashboardConfig
	offset := (p.Page - 1) * p.PageSize
	if err := q.Order("updated_at DESC").Offset(offset).Limit(p.PageSize).Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("list dashboards: %w", err)
	}

	totalPages := int(total) / p.PageSize
	if int(total)%p.PageSize != 0 {
		totalPages++
	}

	return &PaginatedDashboards{
		Data:       configs,
		Total:      total,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *DashboardService) GetByID(ctx context.Context, id, userID uint) (*models.DashboardConfig, error) {
	var cfg models.DashboardConfig
	if err := s.db.WithContext(ctx).First(&cfg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDashboardNotFound
		}
		return nil, fmt.Errorf("get dashboard: %w", err)
	}
	if cfg.UserID != userID {
		return nil, ErrDashboardForbidden
	}
	return &cfg, nil
}

func (s *DashboardService) Update(ctx context.Context, id, userID uint, in UpdateDashboardInput) (*models.DashboardConfig, error) {
	var cfg models.DashboardConfig
	if err := s.db.WithContext(ctx).First(&cfg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDashboardNotFound
		}
		return nil, fmt.Errorf("get dashboard for update: %w", err)
	}
	if cfg.UserID != userID {
		return nil, ErrDashboardForbidden
	}

	updates := map[string]interface{}{}
	if in.Name != "" {
		updates["name"] = in.Name
	}
	if in.Config != "" {
		updates["config"] = in.Config
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&cfg).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update dashboard: %w", err)
		}
	}

	// Reload
	s.db.WithContext(ctx).First(&cfg, id)
	return &cfg, nil
}

func (s *DashboardService) Delete(ctx context.Context, id, userID uint) error {
	var cfg models.DashboardConfig
	if err := s.db.WithContext(ctx).First(&cfg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDashboardNotFound
		}
		return fmt.Errorf("get dashboard for delete: %w", err)
	}
	if cfg.UserID != userID {
		return ErrDashboardForbidden
	}

	if err := s.db.WithContext(ctx).Delete(&cfg).Error; err != nil {
		return fmt.Errorf("delete dashboard: %w", err)
	}
	return nil
}
