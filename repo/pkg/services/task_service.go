package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mindflow/agri-platform/pkg/models"
	"gorm.io/gorm"
)

var ErrTaskNotFound = errors.New("task not found")

type TaskService struct {
	db *gorm.DB
}

func NewTaskService(db *gorm.DB) *TaskService {
	return &TaskService{db: db}
}

type CreateTaskInput struct {
	Title       string `json:"title"       binding:"required,max=255"`
	Description string `json:"description"`
	AssignedTo  uint   `json:"assigned_to"`
	DueDate     string `json:"due_date"` // RFC3339
}

type UpdateTaskInput struct {
	Title       *string `json:"title"       binding:"omitempty,max=255"`
	Description *string `json:"description"`
	Status      *string `json:"status"      binding:"omitempty,oneof=pending in_progress completed cancelled"`
	AssignedTo  *uint   `json:"assigned_to"`
	DueDate     *string `json:"due_date"`
}

func (s *TaskService) Create(ctx context.Context, in CreateTaskInput) (*models.Task, error) {
	task := models.Task{
		Title:       in.Title,
		Description: in.Description,
		Status:      "pending",
		AssignedTo:  in.AssignedTo,
	}

	if in.DueDate != "" {
		t, err := time.Parse(time.RFC3339, in.DueDate)
		if err != nil {
			return nil, fmt.Errorf("invalid due_date format (use RFC3339): %w", err)
		}
		task.DueDate = &t
	}

	if err := s.db.WithContext(ctx).Create(&task).Error; err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	return &task, nil
}

type TaskListParams struct {
	Page       int
	PageSize   int
	Status     string
	AssignedTo uint
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
		Data:       tasks,
		Total:      total,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: totalPages,
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
	if in.Status != nil {
		updates["status"] = *in.Status
	}
	if in.AssignedTo != nil {
		updates["assigned_to"] = *in.AssignedTo
	}
	if in.DueDate != nil {
		t, err := time.Parse(time.RFC3339, *in.DueDate)
		if err != nil {
			return nil, fmt.Errorf("invalid due_date format: %w", err)
		}
		updates["due_date"] = t
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&task).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update task: %w", err)
		}
	}

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
