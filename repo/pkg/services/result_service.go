package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/mindflow/agri-platform/pkg/models"
	"gorm.io/gorm"
)

var ErrResultNotFound = errors.New("result not found")

type ResultService struct {
	db *gorm.DB
}

func NewResultService(db *gorm.DB) *ResultService {
	return &ResultService{db: db}
}

type CreateResultInput struct {
	PlotID  uint   `json:"plot_id"  binding:"required"`
	TaskID  *uint  `json:"task_id"`
	Title   string `json:"title"    binding:"required,max=255"`
	Summary string `json:"summary"`
	Data    string `json:"data"`
}

type UpdateResultInput struct {
	Title   *string `json:"title"   binding:"omitempty,max=255"`
	Summary *string `json:"summary"`
	Data    *string `json:"data"`
}

func (s *ResultService) Create(ctx context.Context, userID uint, in CreateResultInput) (*models.Result, error) {
	result := models.Result{
		PlotID:    in.PlotID,
		TaskID:    in.TaskID,
		Title:     in.Title,
		Summary:   in.Summary,
		Data:      in.Data,
		CreatedBy: userID,
	}
	if err := s.db.WithContext(ctx).Create(&result).Error; err != nil {
		return nil, fmt.Errorf("create result: %w", err)
	}
	return &result, nil
}

type ResultListParams struct {
	Page     int
	PageSize int
	PlotID   uint
	TaskID   uint
}

type PaginatedResults struct {
	Data       []models.Result `json:"data"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

func (s *ResultService) List(ctx context.Context, p ResultListParams) (*PaginatedResults, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}

	q := s.db.WithContext(ctx).Model(&models.Result{})
	if p.PlotID > 0 {
		q = q.Where("plot_id = ?", p.PlotID)
	}
	if p.TaskID > 0 {
		q = q.Where("task_id = ?", p.TaskID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count results: %w", err)
	}

	var results []models.Result
	offset := (p.Page - 1) * p.PageSize
	if err := q.Order("id DESC").Offset(offset).Limit(p.PageSize).Find(&results).Error; err != nil {
		return nil, fmt.Errorf("list results: %w", err)
	}

	totalPages := int(total) / p.PageSize
	if int(total)%p.PageSize != 0 {
		totalPages++
	}

	return &PaginatedResults{
		Data:       results,
		Total:      total,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *ResultService) GetByID(ctx context.Context, id uint) (*models.Result, error) {
	var result models.Result
	if err := s.db.WithContext(ctx).First(&result, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResultNotFound
		}
		return nil, fmt.Errorf("get result: %w", err)
	}
	return &result, nil
}

func (s *ResultService) Update(ctx context.Context, id uint, in UpdateResultInput) (*models.Result, error) {
	var result models.Result
	if err := s.db.WithContext(ctx).First(&result, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResultNotFound
		}
		return nil, fmt.Errorf("find result: %w", err)
	}

	updates := make(map[string]interface{})
	if in.Title != nil {
		updates["title"] = *in.Title
	}
	if in.Summary != nil {
		updates["summary"] = *in.Summary
	}
	if in.Data != nil {
		updates["data"] = *in.Data
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&result).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update result: %w", err)
		}
	}

	return &result, nil
}

func (s *ResultService) Delete(ctx context.Context, id uint) error {
	res := s.db.WithContext(ctx).Delete(&models.Result{}, id)
	if res.Error != nil {
		return fmt.Errorf("delete result: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrResultNotFound
	}
	return nil
}
