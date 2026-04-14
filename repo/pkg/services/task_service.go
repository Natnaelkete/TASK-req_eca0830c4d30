package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/mindflow/agri-platform/pkg/models"
	"gorm.io/gorm"
)

var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrTaskInvalidStatus  = errors.New("invalid task status transition")
)

// OverdueThreshold is the default overdue period (7 days after DueEnd).
const OverdueThreshold = 7 * 24 * time.Hour

type TaskService struct {
	db *gorm.DB
}

func NewTaskService(db *gorm.DB) *TaskService {
	return &TaskService{db: db}
}

type CreateTaskInput struct {
	Title       string `json:"title"       binding:"required,max=255"`
	Description string `json:"description"`
	ObjectID    uint   `json:"object_id"   binding:"required"`
	ObjectType  string `json:"object_type" binding:"required,max=100"`
	CycleType   string `json:"cycle_type"  binding:"omitempty,oneof=daily weekly monthly quarterly"`
	AssignedTo  uint   `json:"assigned_to"`
	ReviewerID  *uint  `json:"reviewer_id"`
	DueStart    string `json:"due_start"`  // RFC3339
	DueEnd      string `json:"due_end"`    // RFC3339
}

type UpdateTaskInput struct {
	Title       *string `json:"title"       binding:"omitempty,max=255"`
	Description *string `json:"description"`
	AssignedTo  *uint   `json:"assigned_to"`
	ReviewerID  *uint   `json:"reviewer_id"`
	DueStart    *string `json:"due_start"`
	DueEnd      *string `json:"due_end"`
}

func (s *TaskService) Create(ctx context.Context, in CreateTaskInput) (*models.Task, error) {
	task := models.Task{
		Title:       in.Title,
		Description: in.Description,
		ObjectID:    in.ObjectID,
		ObjectType:  in.ObjectType,
		CycleType:   in.CycleType,
		Status:      "pending",
		AssignedTo:  in.AssignedTo,
		ReviewerID:  in.ReviewerID,
	}

	if in.DueStart != "" {
		t, err := time.Parse(time.RFC3339, in.DueStart)
		if err != nil {
			return nil, fmt.Errorf("invalid due_start format: %w", err)
		}
		task.DueStart = &t
	}
	if in.DueEnd != "" {
		t, err := time.Parse(time.RFC3339, in.DueEnd)
		if err != nil {
			return nil, fmt.Errorf("invalid due_end format: %w", err)
		}
		task.DueEnd = &t
	}

	if err := s.db.WithContext(ctx).Create(&task).Error; err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	return &task, nil
}

// GenerateTasks creates tasks for a given object and cycle.
type GenerateTasksInput struct {
	ObjectID   uint   `json:"object_id"   binding:"required"`
	ObjectType string `json:"object_type" binding:"required"`
	CycleType  string `json:"cycle_type"  binding:"required,oneof=daily weekly monthly quarterly"`
	Title      string `json:"title"       binding:"required"`
	AssignedTo uint   `json:"assigned_to" binding:"required"`
	ReviewerID *uint  `json:"reviewer_id"`
	Count      int    `json:"count"       binding:"required,min=1,max=52"`
}

func (s *TaskService) GenerateTasks(ctx context.Context, in GenerateTasksInput) ([]models.Task, error) {
	var tasks []models.Task
	now := time.Now()

	for i := 0; i < in.Count; i++ {
		var dueStart, dueEnd time.Time
		switch in.CycleType {
		case "daily":
			dueStart = now.AddDate(0, 0, i)
			dueEnd = dueStart.AddDate(0, 0, 1)
		case "weekly":
			dueStart = now.AddDate(0, 0, i*7)
			dueEnd = dueStart.AddDate(0, 0, 7)
		case "monthly":
			dueStart = now.AddDate(0, i, 0)
			dueEnd = dueStart.AddDate(0, 1, 0)
		case "quarterly":
			dueStart = now.AddDate(0, i*3, 0)
			dueEnd = dueStart.AddDate(0, 3, 0)
		}

		task := models.Task{
			Title:       fmt.Sprintf("%s #%d", in.Title, i+1),
			ObjectID:    in.ObjectID,
			ObjectType:  in.ObjectType,
			CycleType:   in.CycleType,
			Status:      "pending",
			AssignedTo:  in.AssignedTo,
			ReviewerID:  in.ReviewerID,
			DueStart:    &dueStart,
			DueEnd:      &dueEnd,
		}
		if err := s.db.WithContext(ctx).Create(&task).Error; err != nil {
			return nil, fmt.Errorf("generate task %d: %w", i+1, err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// Submit moves a task to submitted status and records submission time.
func (s *TaskService) Submit(ctx context.Context, id uint) (*models.Task, error) {
	var task models.Task
	if err := s.db.WithContext(ctx).First(&task, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("find task: %w", err)
	}

	if task.Status != "pending" && task.Status != "in_progress" {
		return nil, fmt.Errorf("%w: cannot submit from status %s", ErrTaskInvalidStatus, task.Status)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       "submitted",
		"submitted_at": now,
	}
	if err := s.db.WithContext(ctx).Model(&task).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("submit task: %w", err)
	}

	s.db.WithContext(ctx).First(&task, id)
	return &task, nil
}

// Review moves a submitted task to under_review status.
func (s *TaskService) Review(ctx context.Context, id, reviewerID uint) (*models.Task, error) {
	var task models.Task
	if err := s.db.WithContext(ctx).First(&task, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("find task: %w", err)
	}

	if task.Status != "submitted" {
		return nil, fmt.Errorf("%w: cannot review from status %s", ErrTaskInvalidStatus, task.Status)
	}

	updates := map[string]interface{}{
		"status":      "under_review",
		"reviewer_id": reviewerID,
	}
	if err := s.db.WithContext(ctx).Model(&task).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("review task: %w", err)
	}

	s.db.WithContext(ctx).First(&task, id)
	return &task, nil
}

// Complete marks a task as completed (from under_review).
func (s *TaskService) Complete(ctx context.Context, id uint) (*models.Task, error) {
	var task models.Task
	if err := s.db.WithContext(ctx).First(&task, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("find task: %w", err)
	}

	if task.Status != "under_review" {
		return nil, fmt.Errorf("%w: cannot complete from status %s", ErrTaskInvalidStatus, task.Status)
	}

	if err := s.db.WithContext(ctx).Model(&task).Update("status", "completed").Error; err != nil {
		return nil, fmt.Errorf("complete task: %w", err)
	}

	s.db.WithContext(ctx).First(&task, id)
	return &task, nil
}

// MarkOverdueTasks finds tasks past their DueEnd + 7 days and marks them delayed.
// This is called by a background goroutine.
func (s *TaskService) MarkOverdueTasks(ctx context.Context) (int64, error) {
	cutoff := time.Now().Add(-OverdueThreshold)
	res := s.db.WithContext(ctx).
		Model(&models.Task{}).
		Where("status IN ? AND due_end IS NOT NULL AND due_end < ? AND overdue_at IS NULL",
			[]string{"pending", "in_progress"}, cutoff).
		Updates(map[string]interface{}{
			"status":     "delayed",
			"overdue_at": time.Now(),
		})
	if res.Error != nil {
		return 0, fmt.Errorf("mark overdue tasks: %w", res.Error)
	}
	return res.RowsAffected, nil
}

// StartOverdueChecker starts a background goroutine that checks for overdue tasks every minute.
func (s *TaskService) StartOverdueChecker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("overdue checker stopped")
				return
			case <-ticker.C:
				count, err := s.MarkOverdueTasks(ctx)
				if err != nil {
					log.Printf("overdue check error: %v", err)
				} else if count > 0 {
					log.Printf("marked %d tasks as delayed", count)
				}
			}
		}
	}()
}

type TaskListParams struct {
	Page       int
	PageSize   int
	Status     string
	AssignedTo uint
	ObjectID   uint
	ObjectType string
}

type PaginatedTasks struct {
	Data       []models.Task `json:"data"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

func (s *TaskService) List(ctx context.Context, p TaskListParams) (*PaginatedTasks, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}

	q := s.db.WithContext(ctx).Model(&models.Task{})
	if p.Status != "" {
		q = q.Where("status = ?", p.Status)
	}
	if p.AssignedTo > 0 {
		q = q.Where("assigned_to = ?", p.AssignedTo)
	}
	if p.ObjectID > 0 {
		q = q.Where("object_id = ?", p.ObjectID)
	}
	if p.ObjectType != "" {
		q = q.Where("object_type = ?", p.ObjectType)
	}

	// Filter by visibility window
	now := time.Now()
	q = q.Where("(due_start IS NULL OR due_start <= ?) AND (due_end IS NULL OR due_end >= ? OR status = 'delayed')", now, now)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count tasks: %w", err)
	}

	var tasks []models.Task
	offset := (p.Page - 1) * p.PageSize
	if err := q.Order("id DESC").Offset(offset).Limit(p.PageSize).Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	totalPages := int(total) / p.PageSize
	if int(total)%p.PageSize != 0 {
		totalPages++
	}

	return &PaginatedTasks{
		Data: tasks, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: totalPages,
	}, nil
}

func (s *TaskService) GetByID(ctx context.Context, id uint) (*models.Task, error) {
	var task models.Task
	if err := s.db.WithContext(ctx).First(&task, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("get task: %w", err)
	}
	return &task, nil
}

func (s *TaskService) Update(ctx context.Context, id uint, in UpdateTaskInput) (*models.Task, error) {
	var task models.Task
	if err := s.db.WithContext(ctx).First(&task, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("find task: %w", err)
	}

	updates := make(map[string]interface{})
	if in.Title != nil {
		updates["title"] = *in.Title
	}
	if in.Description != nil {
		updates["description"] = *in.Description
	}
	if in.AssignedTo != nil {
		updates["assigned_to"] = *in.AssignedTo
	}
	if in.ReviewerID != nil {
		updates["reviewer_id"] = *in.ReviewerID
	}
	if in.DueStart != nil {
		t, err := time.Parse(time.RFC3339, *in.DueStart)
		if err != nil {
			return nil, fmt.Errorf("invalid due_start: %w", err)
		}
		updates["due_start"] = t
	}
	if in.DueEnd != nil {
		t, err := time.Parse(time.RFC3339, *in.DueEnd)
		if err != nil {
			return nil, fmt.Errorf("invalid due_end: %w", err)
		}
		updates["due_end"] = t
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&task).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update task: %w", err)
		}
	}

	s.db.WithContext(ctx).First(&task, id)
	return &task, nil
}

func (s *TaskService) Delete(ctx context.Context, id uint) error {
	res := s.db.WithContext(ctx).Delete(&models.Task{}, id)
	if res.Error != nil {
		return fmt.Errorf("delete task: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil
}
