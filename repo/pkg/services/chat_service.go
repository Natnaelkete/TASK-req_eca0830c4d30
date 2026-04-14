package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/mindflow/agri-platform/pkg/models"
	"gorm.io/gorm"
)

var ErrMessageNotFound = errors.New("message not found")

type ChatService struct {
	db *gorm.DB
}

func NewChatService(db *gorm.DB) *ChatService {
	return &ChatService{db: db}
}

type SendMessageInput struct {
	ReceiverID uint   `json:"receiver_id" binding:"required"`
	PlotID     *uint  `json:"plot_id"`
	Content    string `json:"content"     binding:"required,max=2000"`
}

func (s *ChatService) Send(ctx context.Context, senderID uint, in SendMessageInput) (*models.Message, error) {
	msg := models.Message{
		SenderID:   senderID,
		ReceiverID: in.ReceiverID,
		PlotID:     in.PlotID,
		Content:    in.Content,
	}
	if err := s.db.WithContext(ctx).Create(&msg).Error; err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}
	return &msg, nil
}

type MessageListParams struct {
	Page     int
	PageSize int
	UserID   uint // messages where user is sender or receiver
	PlotID   uint
}

type PaginatedMessages struct {
	Data       []models.Message `json:"data"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

func (s *ChatService) List(ctx context.Context, p MessageListParams) (*PaginatedMessages, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}

	q := s.db.WithContext(ctx).Model(&models.Message{})
	if p.UserID > 0 {
		q = q.Where("sender_id = ? OR receiver_id = ?", p.UserID, p.UserID)
	}
	if p.PlotID > 0 {
		q = q.Where("plot_id = ?", p.PlotID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count messages: %w", err)
	}

	var msgs []models.Message
	offset := (p.Page - 1) * p.PageSize
	if err := q.Order("id DESC").Offset(offset).Limit(p.PageSize).Find(&msgs).Error; err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}

	totalPages := int(total) / p.PageSize
	if int(total)%p.PageSize != 0 {
		totalPages++
	}

	return &PaginatedMessages{
		Data:       msgs,
		Total:      total,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *ChatService) MarkRead(ctx context.Context, msgID, userID uint) error {
	res := s.db.WithContext(ctx).Model(&models.Message{}).
		Where("id = ? AND receiver_id = ?", msgID, userID).
		Update("read", true)
	if res.Error != nil {
		return fmt.Errorf("mark read: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrMessageNotFound
	}
	return nil
}
