package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mindflow/agri-platform/pkg/models"
	"gorm.io/gorm"
)

var (
	ErrIndicatorNotFound = errors.New("indicator not found")
	ErrIndicatorExists   = errors.New("indicator code already exists")
)

// IndicatorService manages indicator definitions and their version history.
type IndicatorService struct {
	db *gorm.DB
}

func NewIndicatorService(db *gorm.DB) *IndicatorService {
	return &IndicatorService{db: db}
}

// CreateIndicatorInput is the payload for creating a new indicator definition.
type CreateIndicatorInput struct {
	Code        string `json:"code"        binding:"required,max=100"`
	Name        string `json:"name"        binding:"required,max=255"`
	Description string `json:"description"`
	Unit        string `json:"unit"        binding:"max=50"`
	Formula     string `json:"formula"`
	Category    string `json:"category"    binding:"max=100"`
}

// UpdateIndicatorInput is the payload for updating an indicator definition.
type UpdateIndicatorInput struct {
	Name        *string `json:"name"        binding:"omitempty,max=255"`
	Description *string `json:"description"`
	Unit        *string `json:"unit"        binding:"omitempty,max=50"`
	Formula     *string `json:"formula"`
	Category    *string `json:"category"    binding:"omitempty,max=100"`
	DiffSummary string  `json:"diff_summary" binding:"required"` // required: describe what changed
}

// Create creates a new indicator definition and records version 1.
func (s *IndicatorService) Create(ctx context.Context, userID uint, in CreateIndicatorInput) (*models.IndicatorDefinition, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.IndicatorDefinition{}).
		Where("code = ?", in.Code).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("check indicator: %w", err)
	}
	if count > 0 {
		return nil, ErrIndicatorExists
	}

	tx := s.db.WithContext(ctx).Begin()

	indicator := models.IndicatorDefinition{
		Code:        in.Code,
		Name:        in.Name,
		Description: in.Description,
		Unit:        in.Unit,
		Formula:     in.Formula,
		Category:    in.Category,
		Status:      "active",
		CreatedBy:   userID,
	}
	if err := tx.Create(&indicator).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create indicator: %w", err)
	}

	// Record initial version
	version := models.IndicatorVersion{
		IndicatorID: indicator.ID,
		Version:     1,
		Name:        in.Name,
		Description: in.Description,
		Unit:        in.Unit,
		Formula:     in.Formula,
		Category:    in.Category,
		DiffSummary: "Initial creation",
		ModifiedBy:  userID,
		ModifiedAt:  time.Now(),
	}
	if err := tx.Create(&version).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create initial version: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit create: %w", err)
	}

	return &indicator, nil
}

// Update updates an indicator definition and records a new version with diff.
func (s *IndicatorService) Update(ctx context.Context, id, userID uint, in UpdateIndicatorInput) (*models.IndicatorDefinition, error) {
	var indicator models.IndicatorDefinition
	if err := s.db.WithContext(ctx).First(&indicator, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIndicatorNotFound
		}
		return nil, fmt.Errorf("find indicator: %w", err)
	}

	// Get current max version
	var maxVersion int
	s.db.WithContext(ctx).Model(&models.IndicatorVersion{}).
		Where("indicator_id = ?", id).
		Select("COALESCE(MAX(version), 0)").Scan(&maxVersion)

	// Build diff and updates
	updates := make(map[string]interface{})
	var changes []string

	if in.Name != nil && *in.Name != indicator.Name {
		changes = append(changes, fmt.Sprintf("name: %q -> %q", indicator.Name, *in.Name))
		updates["name"] = *in.Name
		indicator.Name = *in.Name
	}
	if in.Description != nil && *in.Description != indicator.Description {
		changes = append(changes, "description updated")
		updates["description"] = *in.Description
		indicator.Description = *in.Description
	}
	if in.Unit != nil && *in.Unit != indicator.Unit {
		changes = append(changes, fmt.Sprintf("unit: %q -> %q", indicator.Unit, *in.Unit))
		updates["unit"] = *in.Unit
		indicator.Unit = *in.Unit
	}
	if in.Formula != nil && *in.Formula != indicator.Formula {
		changes = append(changes, "formula updated")
		updates["formula"] = *in.Formula
		indicator.Formula = *in.Formula
	}
	if in.Category != nil && *in.Category != indicator.Category {
		changes = append(changes, fmt.Sprintf("category: %q -> %q", indicator.Category, *in.Category))
		updates["category"] = *in.Category
		indicator.Category = *in.Category
	}

	if len(updates) == 0 {
		return &indicator, nil
	}

	tx := s.db.WithContext(ctx).Begin()

	if err := tx.Model(&models.IndicatorDefinition{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("update indicator: %w", err)
	}

	diffSummary := in.DiffSummary
	if diffSummary == "" {
		diffSummary = strings.Join(changes, "; ")
	}

	version := models.IndicatorVersion{
		IndicatorID: id,
		Version:     maxVersion + 1,
		Name:        indicator.Name,
		Description: indicator.Description,
		Unit:        indicator.Unit,
		Formula:     indicator.Formula,
		Category:    indicator.Category,
		DiffSummary: diffSummary,
		ModifiedBy:  userID,
		ModifiedAt:  time.Now(),
	}
	if err := tx.Create(&version).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create version: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit update: %w", err)
	}

	s.db.WithContext(ctx).First(&indicator, id)
	return &indicator, nil
}

// GetByID returns an indicator definition.
func (s *IndicatorService) GetByID(ctx context.Context, id uint) (*models.IndicatorDefinition, error) {
	var indicator models.IndicatorDefinition
	if err := s.db.WithContext(ctx).First(&indicator, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIndicatorNotFound
		}
		return nil, fmt.Errorf("get indicator: %w", err)
	}
	return &indicator, nil
}

// IndicatorListParams defines pagination and filters for listing indicators.
type IndicatorListParams struct {
	Page     int
	PageSize int
	Category string
	Status   string
}

// PaginatedIndicators holds paginated indicator results.
type PaginatedIndicators struct {
	Data       []models.IndicatorDefinition `json:"data"`
	Total      int64                        `json:"total"`
	Page       int                          `json:"page"`
	PageSize   int                          `json:"page_size"`
	TotalPages int                          `json:"total_pages"`
}

// List returns paginated indicator definitions.
func (s *IndicatorService) List(ctx context.Context, p IndicatorListParams) (*PaginatedIndicators, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}

	q := s.db.WithContext(ctx).Model(&models.IndicatorDefinition{})
	if p.Category != "" {
		q = q.Where("category = ?", p.Category)
	}
	if p.Status != "" {
		q = q.Where("status = ?", p.Status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count indicators: %w", err)
	}

	var indicators []models.IndicatorDefinition
	offset := (p.Page - 1) * p.PageSize
	if err := q.Order("id DESC").Offset(offset).Limit(p.PageSize).Find(&indicators).Error; err != nil {
		return nil, fmt.Errorf("list indicators: %w", err)
	}

	totalPages := int(total) / p.PageSize
	if int(total)%p.PageSize != 0 {
		totalPages++
	}

	return &PaginatedIndicators{
		Data: indicators, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: totalPages,
	}, nil
}

// ListVersions returns all versions for an indicator.
func (s *IndicatorService) ListVersions(ctx context.Context, indicatorID uint) ([]models.IndicatorVersion, error) {
	var versions []models.IndicatorVersion
	if err := s.db.WithContext(ctx).
		Where("indicator_id = ?", indicatorID).
		Order("version DESC").
		Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("list versions: %w", err)
	}
	return versions, nil
}

// GetVersion returns a specific version of an indicator.
func (s *IndicatorService) GetVersion(ctx context.Context, indicatorID uint, version int) (*models.IndicatorVersion, error) {
	var v models.IndicatorVersion
	if err := s.db.WithContext(ctx).
		Where("indicator_id = ? AND version = ?", indicatorID, version).
		First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIndicatorNotFound
		}
		return nil, fmt.Errorf("get version: %w", err)
	}
	return &v, nil
}

// Delete soft-deprecates an indicator (sets status to deprecated).
func (s *IndicatorService) Delete(ctx context.Context, id uint) error {
	res := s.db.WithContext(ctx).Model(&models.IndicatorDefinition{}).
		Where("id = ?", id).
		Update("status", "deprecated")
	if res.Error != nil {
		return fmt.Errorf("deprecate indicator: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrIndicatorNotFound
	}
	return nil
}
