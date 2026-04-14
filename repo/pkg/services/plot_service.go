package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/mindflow/agri-platform/pkg/models"
	"gorm.io/gorm"
)

var ErrPlotNotFound = errors.New("plot not found")

type PlotService struct {
	db *gorm.DB
}

func NewPlotService(db *gorm.DB) *PlotService {
	return &PlotService{db: db}
}

type CreatePlotInput struct {
	Name     string  `json:"name"      binding:"required,max=200"`
	Location string  `json:"location"  binding:"max=255"`
	Area     float64 `json:"area"      binding:"gte=0"`
	SoilType string  `json:"soil_type" binding:"max=100"`
	CropType string  `json:"crop_type" binding:"max=100"`
}

type UpdatePlotInput struct {
	Name     *string  `json:"name"      binding:"omitempty,max=200"`
	Location *string  `json:"location"  binding:"omitempty,max=255"`
	Area     *float64 `json:"area"      binding:"omitempty,gte=0"`
	SoilType *string  `json:"soil_type" binding:"omitempty,max=100"`
	CropType *string  `json:"crop_type" binding:"omitempty,max=100"`
}

func (s *PlotService) Create(ctx context.Context, userID uint, in CreatePlotInput) (*models.Plot, error) {
	plot := models.Plot{
		Name:     in.Name,
		Location: in.Location,
		Area:     in.Area,
		SoilType: in.SoilType,
		CropType: in.CropType,
		UserID:   userID,
	}
	if err := s.db.WithContext(ctx).Create(&plot).Error; err != nil {
		return nil, fmt.Errorf("create plot: %w", err)
	}
	return &plot, nil
}

type PlotListParams struct {
	Page     int
	PageSize int
	UserID   uint
}

type PaginatedPlots struct {
	Data       []models.Plot `json:"data"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

func (s *PlotService) List(ctx context.Context, p PlotListParams) (*PaginatedPlots, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}

	q := s.db.WithContext(ctx).Model(&models.Plot{})
	if p.UserID > 0 {
		q = q.Where("user_id = ?", p.UserID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count plots: %w", err)
	}

	var plots []models.Plot
	offset := (p.Page - 1) * p.PageSize
	if err := q.Order("id DESC").Offset(offset).Limit(p.PageSize).Find(&plots).Error; err != nil {
		return nil, fmt.Errorf("list plots: %w", err)
	}

	totalPages := int(total) / p.PageSize
	if int(total)%p.PageSize != 0 {
		totalPages++
	}

	return &PaginatedPlots{
		Data:       plots,
		Total:      total,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *PlotService) GetByID(ctx context.Context, id uint) (*models.Plot, error) {
	var plot models.Plot
	if err := s.db.WithContext(ctx).First(&plot, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPlotNotFound
		}
		return nil, fmt.Errorf("get plot: %w", err)
	}
	return &plot, nil
}

func (s *PlotService) Update(ctx context.Context, id uint, in UpdatePlotInput) (*models.Plot, error) {
	var plot models.Plot
	if err := s.db.WithContext(ctx).First(&plot, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPlotNotFound
		}
		return nil, fmt.Errorf("find plot: %w", err)
	}

	updates := make(map[string]interface{})
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.Location != nil {
		updates["location"] = *in.Location
	}
	if in.Area != nil {
		updates["area"] = *in.Area
	}
	if in.SoilType != nil {
		updates["soil_type"] = *in.SoilType
	}
	if in.CropType != nil {
		updates["crop_type"] = *in.CropType
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&plot).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update plot: %w", err)
		}
	}

	return &plot, nil
}

func (s *PlotService) Delete(ctx context.Context, id uint) error {
	res := s.db.WithContext(ctx).Delete(&models.Plot{}, id)
	if res.Error != nil {
		return fmt.Errorf("delete plot: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrPlotNotFound
	}
	return nil
}
