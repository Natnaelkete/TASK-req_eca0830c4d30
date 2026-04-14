package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mindflow/agri-platform/pkg/models"
	"gorm.io/gorm"
)

var (
	ErrResultNotFound        = errors.New("result not found")
	ErrInvalidTransition     = errors.New("invalid status transition")
	ErrResultArchived        = errors.New("archived results cannot be modified")
	ErrInvalidationNoReason  = errors.New("invalidation requires a reason")
	ErrFieldValidationFailed = errors.New("field validation failed")
	ErrResultForbidden       = errors.New("not authorized to access this result")
)

// validTransitions defines the strict status state machine.
// draft → submitted → returned → approved → archived
var validTransitions = map[string][]string{
	"draft":     {"submitted"},
	"submitted": {"returned", "approved"},
	"returned":  {"submitted"},
	"approved":  {"archived"},
}

type ResultService struct {
	db *gorm.DB
}

func NewResultService(db *gorm.DB) *ResultService {
	return &ResultService{db: db}
}

// IsValidTransition checks if a status transition is allowed.
func IsValidTransition(from, to string) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == to {
			return true
		}
	}
	return false
}

type CreateResultInput struct {
	Type    string `json:"type"    binding:"required,oneof=paper project patent"`
	PlotID  uint   `json:"plot_id" binding:"required"`
	TaskID  *uint  `json:"task_id"`
	Title   string `json:"title"   binding:"required,max=255"`
	Summary string `json:"summary"`
	Fields  string `json:"fields"` // JSON string
}

type UpdateResultInput struct {
	Title   *string `json:"title"   binding:"omitempty,max=255"`
	Summary *string `json:"summary"`
	Fields  *string `json:"fields"`
}

func (s *ResultService) Create(ctx context.Context, userID uint, in CreateResultInput) (*models.Result, error) {
	if in.Fields != "" {
		if err := s.validateFields(ctx, in.Type, in.Fields); err != nil {
			return nil, err
		}
	}

	result := models.Result{
		Type:        in.Type,
		PlotID:      in.PlotID,
		TaskID:      in.TaskID,
		Title:       in.Title,
		Summary:     in.Summary,
		Fields:      in.Fields,
		Status:      "draft",
		SubmitterID: userID,
		CreatedBy:   userID,
	}
	if err := s.db.WithContext(ctx).Create(&result).Error; err != nil {
		return nil, fmt.Errorf("create result: %w", err)
	}
	return &result, nil
}

func (s *ResultService) Update(ctx context.Context, id, userID uint, role string, in UpdateResultInput) (*models.Result, error) {
	var result models.Result
	if err := s.db.WithContext(ctx).First(&result, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResultNotFound
		}
		return nil, fmt.Errorf("find result: %w", err)
	}

	// Ownership check: only submitter/creator or admin can update
	if role != "admin" && result.SubmitterID != userID && result.CreatedBy != userID {
		return nil, ErrResultForbidden
	}

	if result.Status == "archived" {
		return nil, ErrResultArchived
	}

	updates := make(map[string]interface{})
	if in.Title != nil {
		updates["title"] = *in.Title
	}
	if in.Summary != nil {
		updates["summary"] = *in.Summary
	}
	if in.Fields != nil {
		if err := s.validateFields(ctx, result.Type, *in.Fields); err != nil {
			return nil, err
		}
		updates["fields"] = *in.Fields
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&result).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update result: %w", err)
		}
	}

	s.db.WithContext(ctx).First(&result, id)
	return &result, nil
}

// Transition changes the status following the strict state machine.
func (s *ResultService) Transition(ctx context.Context, id, userID uint, toStatus, reason string) (*models.Result, error) {
	var result models.Result
	if err := s.db.WithContext(ctx).First(&result, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResultNotFound
		}
		return nil, fmt.Errorf("find result: %w", err)
	}

	if !IsValidTransition(result.Status, toStatus) {
		return nil, fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidTransition, result.Status, toStatus)
	}

	tx := s.db.WithContext(ctx).Begin()

	statusLog := models.ResultStatusLog{
		ResultID:   id,
		FromStatus: result.Status,
		ToStatus:   toStatus,
		ChangedBy:  userID,
		Reason:     reason,
	}
	if err := tx.Create(&statusLog).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("log transition: %w", err)
	}

	updates := map[string]interface{}{"status": toStatus}
	if toStatus == "archived" {
		now := time.Now()
		updates["archived_at"] = now
	}

	if err := tx.Model(&result).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("update status: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit transition: %w", err)
	}

	s.db.WithContext(ctx).First(&result, id)
	return &result, nil
}

// AppendNotes appends retrospective notes to an archived result.
func (s *ResultService) AppendNotes(ctx context.Context, id, userID uint, notes string) (*models.Result, error) {
	var result models.Result
	if err := s.db.WithContext(ctx).First(&result, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResultNotFound
		}
		return nil, fmt.Errorf("find result: %w", err)
	}

	if result.Status != "archived" {
		return nil, fmt.Errorf("notes can only be appended to archived results")
	}

	existing := result.Notes
	if existing != "" {
		existing += "\n"
	}
	existing += fmt.Sprintf("[%s by user %d] %s", time.Now().Format(time.RFC3339), userID, notes)

	if err := s.db.WithContext(ctx).Model(&result).Update("notes", existing).Error; err != nil {
		return nil, fmt.Errorf("append notes: %w", err)
	}

	s.db.WithContext(ctx).First(&result, id)
	return &result, nil
}

// Invalidate marks an archived result as invalidated with reason and full traceability.
func (s *ResultService) Invalidate(ctx context.Context, id, userID uint, reason string) (*models.Result, error) {
	if reason == "" {
		return nil, ErrInvalidationNoReason
	}

	var result models.Result
	if err := s.db.WithContext(ctx).First(&result, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResultNotFound
		}
		return nil, fmt.Errorf("find result: %w", err)
	}

	if result.Status != "archived" {
		return nil, fmt.Errorf("only archived results can be invalidated")
	}

	now := time.Now()
	tx := s.db.WithContext(ctx).Begin()

	statusLog := models.ResultStatusLog{
		ResultID:   id,
		FromStatus: "archived",
		ToStatus:   "invalidated",
		ChangedBy:  userID,
		Reason:     reason,
	}
	if err := tx.Create(&statusLog).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("log invalidation: %w", err)
	}

	updates := map[string]interface{}{
		"invalidated_reason": reason,
		"invalidated_by":     userID,
		"invalidated_at":     now,
	}
	if err := tx.Model(&result).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("invalidate result: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit invalidation: %w", err)
	}

	s.db.WithContext(ctx).First(&result, id)
	return &result, nil
}

type ResultListParams struct {
	Page     int
	PageSize int
	PlotID   uint
	TaskID   uint
	Type     string
	Status   string
	UserID   uint   // authenticated user for scoping
	Role     string // authenticated user role
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

	// Object-level isolation: non-admin users only see results they created/submitted
	if p.Role != "admin" && p.UserID > 0 {
		q = q.Where("submitter_id = ? OR created_by = ?", p.UserID, p.UserID)
	}

	if p.PlotID > 0 {
		q = q.Where("plot_id = ?", p.PlotID)
	}
	if p.TaskID > 0 {
		q = q.Where("task_id = ?", p.TaskID)
	}
	if p.Type != "" {
		q = q.Where("type = ?", p.Type)
	}
	if p.Status != "" {
		q = q.Where("status = ?", p.Status)
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
		Data: results, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: totalPages,
	}, nil
}

func (s *ResultService) GetByID(ctx context.Context, id, userID uint, role string) (*models.Result, error) {
	var result models.Result
	if err := s.db.WithContext(ctx).First(&result, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResultNotFound
		}
		return nil, fmt.Errorf("get result: %w", err)
	}
	// Object-level check: non-admin users must be submitter or creator
	if role != "admin" && result.SubmitterID != userID && result.CreatedBy != userID {
		return nil, ErrResultForbidden
	}
	return &result, nil
}

func (s *ResultService) Delete(ctx context.Context, id, userID uint, role string) error {
	var result models.Result
	if err := s.db.WithContext(ctx).First(&result, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrResultNotFound
		}
		return fmt.Errorf("find result: %w", err)
	}

	// Ownership check: only submitter/creator or admin can delete
	if role != "admin" && result.SubmitterID != userID && result.CreatedBy != userID {
		return ErrResultForbidden
	}

	if result.Status == "archived" {
		return ErrResultArchived
	}

	res := s.db.WithContext(ctx).Delete(&models.Result{}, id)
	if res.Error != nil {
		return fmt.Errorf("delete result: %w", res.Error)
	}
	return nil
}

// --- Field Rule Management ---

type CreateFieldRuleInput struct {
	ResultType string `json:"result_type" binding:"required,oneof=paper project patent"`
	FieldName  string `json:"field_name"  binding:"required,max=100"`
	Required   bool   `json:"required"`
	MaxLength  int    `json:"max_length"`
	EnumValues string `json:"enum_values"`
}

func (s *ResultService) CreateFieldRule(ctx context.Context, in CreateFieldRuleInput) (*models.FieldRule, error) {
	rule := models.FieldRule{
		ResultType: in.ResultType,
		FieldName:  in.FieldName,
		Required:   in.Required,
		MaxLength:  in.MaxLength,
		EnumValues: in.EnumValues,
	}
	if err := s.db.WithContext(ctx).Create(&rule).Error; err != nil {
		return nil, fmt.Errorf("create field rule: %w", err)
	}
	return &rule, nil
}

func (s *ResultService) ListFieldRules(ctx context.Context, resultType string) ([]models.FieldRule, error) {
	var rules []models.FieldRule
	q := s.db.WithContext(ctx).Model(&models.FieldRule{})
	if resultType != "" {
		q = q.Where("result_type = ?", resultType)
	}
	if err := q.Order("field_name ASC").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("list field rules: %w", err)
	}
	return rules, nil
}

func (s *ResultService) validateFields(ctx context.Context, resultType, fieldsJSON string) error {
	if fieldsJSON == "" {
		return nil
	}

	var fields map[string]interface{}
	if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
		return fmt.Errorf("%w: invalid JSON in fields", ErrFieldValidationFailed)
	}

	var rules []models.FieldRule
	if err := s.db.WithContext(ctx).Where("result_type = ?", resultType).Find(&rules).Error; err != nil {
		return fmt.Errorf("retrieve field rules: %w", err)
	}

	for _, rule := range rules {
		val, exists := fields[rule.FieldName]

		if rule.Required && !exists {
			return fmt.Errorf("%w: field %q is required for %s", ErrFieldValidationFailed, rule.FieldName, resultType)
		}

		if exists && rule.MaxLength > 0 {
			if str, ok := val.(string); ok && len(str) > rule.MaxLength {
				return fmt.Errorf("%w: field %q exceeds max length %d", ErrFieldValidationFailed, rule.FieldName, rule.MaxLength)
			}
		}

		if exists && rule.EnumValues != "" {
			if str, ok := val.(string); ok {
				allowed := strings.Split(rule.EnumValues, ",")
				found := false
				for _, a := range allowed {
					if strings.TrimSpace(a) == str {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("%w: field %q must be one of [%s]", ErrFieldValidationFailed, rule.FieldName, rule.EnumValues)
				}
			}
		}
	}

	return nil
}
