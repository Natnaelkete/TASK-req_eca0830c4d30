package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mindflow/agri-platform/pkg/models"
	"gorm.io/gorm"
)

var (
	ErrOrderNotFound       = errors.New("order not found")
	ErrOrderForbidden      = errors.New("not authorized to access this order")
	ErrConversationNotFound = errors.New("conversation not found")
	ErrRateLimitExceeded   = errors.New("rate limit exceeded: max 20 messages per minute")
	ErrSensitiveWord       = errors.New("message contains sensitive content and was blocked")
	ErrTemplateNotFound    = errors.New("template not found")
)

// sensitiveWords is the default list of words to intercept.
var sensitiveWords = []string{"illegal", "spam", "fraud", "scam", "hack"}

// ConversationService handles order conversations, transfers, templates, sensitive words, and rate limiting.
type ConversationService struct {
	db      *gorm.DB
	limiter *RateLimiter
}

func NewConversationService(db *gorm.DB) *ConversationService {
	return &ConversationService{
		db:      db,
		limiter: NewRateLimiter(20, time.Minute),
	}
}

// --- Orders ---

type CreateOrderInput struct {
	Title string `json:"title" binding:"required,max=255"`
}

func (s *ConversationService) CreateOrder(ctx context.Context, researcherID uint) (*models.Order, error) {
	return s.CreateOrderWithTitle(ctx, researcherID, CreateOrderInput{Title: "New Order"})
}

func (s *ConversationService) CreateOrderWithTitle(ctx context.Context, researcherID uint, in CreateOrderInput) (*models.Order, error) {
	order := models.Order{
		ResearcherID: researcherID,
		Title:        in.Title,
		Status:       "open",
		AssignedTo:   researcherID,
	}
	if err := s.db.WithContext(ctx).Create(&order).Error; err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}
	return &order, nil
}

func (s *ConversationService) GetOrder(ctx context.Context, id, userID uint, role string) (*models.Order, error) {
	var order models.Order
	if err := s.db.WithContext(ctx).First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order: %w", err)
	}
	// Object-level check: only owner, assigned user, or admin/customer_service can access
	if role != "admin" && role != "customer_service" && order.ResearcherID != userID && order.AssignedTo != userID {
		return nil, ErrOrderForbidden
	}
	return &order, nil
}

// checkOrderAccess verifies user has access to the given order.
func (s *ConversationService) checkOrderAccess(ctx context.Context, orderID, userID uint, role string) (*models.Order, error) {
	var order models.Order
	if err := s.db.WithContext(ctx).First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order: %w", err)
	}
	if role != "admin" && role != "customer_service" && order.ResearcherID != userID && order.AssignedTo != userID {
		return nil, ErrOrderForbidden
	}
	return &order, nil
}

type OrderListParams struct {
	Page     int
	PageSize int
	UserID   uint
	Status   string
}

type PaginatedOrders struct {
	Data       []models.Order `json:"data"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

func (s *ConversationService) ListOrders(ctx context.Context, p OrderListParams) (*PaginatedOrders, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}

	q := s.db.WithContext(ctx).Model(&models.Order{})
	if p.UserID > 0 {
		q = q.Where("researcher_id = ? OR assigned_to = ?", p.UserID, p.UserID)
	}
	if p.Status != "" {
		q = q.Where("status = ?", p.Status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count orders: %w", err)
	}

	var orders []models.Order
	offset := (p.Page - 1) * p.PageSize
	if err := q.Order("updated_at DESC").Offset(offset).Limit(p.PageSize).Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}

	totalPages := int(total) / p.PageSize
	if int(total)%p.PageSize != 0 {
		totalPages++
	}

	return &PaginatedOrders{
		Data: orders, Total: total, Page: p.Page, PageSize: p.PageSize, TotalPages: totalPages,
	}, nil
}

// --- Conversations (Messages in Order Threads) ---

type PostMessageInput struct {
	Message string `json:"message" binding:"required,max=2000"`
}

// PostMessage posts a message to an order conversation thread.
// Enforces rate limiting (20/min) and sensitive word filtering.
func (s *ConversationService) PostMessage(ctx context.Context, orderID, userID uint, role string, in PostMessageInput) (*models.Conversation, error) {
	// Rate limit check
	if !s.limiter.Allow(userID) {
		return nil, ErrRateLimitExceeded
	}

	// Sensitive word check
	if word := containsSensitiveWord(in.Message); word != "" {
		// Log the intercepted attempt
		logEntry := models.SensitiveWordLog{
			UserID:  userID,
			OrderID: orderID,
			Content: in.Message,
			Word:    word,
		}
		s.db.WithContext(ctx).Create(&logEntry)
		return nil, ErrSensitiveWord
	}

	// Verify order exists and user has access
	if _, err := s.checkOrderAccess(ctx, orderID, userID, role); err != nil {
		return nil, err
	}

	conv := models.Conversation{
		OrderID: orderID,
		UserID:  userID,
		Message: in.Message,
	}
	if err := s.db.WithContext(ctx).Create(&conv).Error; err != nil {
		return nil, fmt.Errorf("post message: %w", err)
	}
	return &conv, nil
}

// ListMessages returns paginated conversation messages for an order.
func (s *ConversationService) ListMessages(ctx context.Context, orderID, userID uint, role string, page, pageSize int) ([]models.Conversation, int64, error) {
	// Verify access to order
	if _, err := s.checkOrderAccess(ctx, orderID, userID, role); err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	q := s.db.WithContext(ctx).Model(&models.Conversation{}).Where("order_id = ?", orderID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count messages: %w", err)
	}

	var msgs []models.Conversation
	offset := (page - 1) * pageSize
	if err := q.Order("created_at ASC").Offset(offset).Limit(pageSize).Find(&msgs).Error; err != nil {
		return nil, 0, fmt.Errorf("list messages: %w", err)
	}

	return msgs, total, nil
}

// MarkRead marks a conversation message as read by the given user.
func (s *ConversationService) MarkRead(ctx context.Context, msgID, userID uint) error {
	now := time.Now()
	res := s.db.WithContext(ctx).Model(&models.Conversation{}).
		Where("id = ? AND user_id != ?", msgID, userID). // can only mark others' messages as read
		Where("read_at IS NULL").
		Update("read_at", now)
	if res.Error != nil {
		return fmt.Errorf("mark read: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrConversationNotFound
	}
	return nil
}

// --- Ticket Transfers ---

type TransferInput struct {
	TransferToUserID uint   `json:"transfer_to_user_id" binding:"required"`
	Reason           string `json:"reason"`
}

// TransferTicket transfers an order to another user and records the transfer.
func (s *ConversationService) TransferTicket(ctx context.Context, orderID, fromUserID uint, role string, in TransferInput) (*models.Conversation, error) {
	order, err := s.checkOrderAccess(ctx, orderID, fromUserID, role)
	if err != nil {
		return nil, err
	}

	// Update order assignment
	if err := s.db.WithContext(ctx).Model(order).Update("assigned_to", in.TransferToUserID).Error; err != nil {
		return nil, fmt.Errorf("update order assignment: %w", err)
	}

	// Record transfer as a conversation entry
	msg := fmt.Sprintf("Ticket transferred to user %d", in.TransferToUserID)
	if in.Reason != "" {
		msg += ": " + in.Reason
	}
	conv := models.Conversation{
		OrderID:       orderID,
		UserID:        fromUserID,
		Message:       msg,
		TransferredTo: &in.TransferToUserID,
	}
	if err := s.db.WithContext(ctx).Create(&conv).Error; err != nil {
		return nil, fmt.Errorf("record transfer: %w", err)
	}

	return &conv, nil
}

// --- Template Messages ---

type CreateTemplateInput struct {
	Name    string `json:"name"    binding:"required,max=255"`
	Content string `json:"content" binding:"required"`
}

func (s *ConversationService) CreateTemplate(ctx context.Context, in CreateTemplateInput) (*models.TemplateMessage, error) {
	tmpl := models.TemplateMessage{
		Name:    in.Name,
		Content: in.Content,
	}
	if err := s.db.WithContext(ctx).Create(&tmpl).Error; err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}
	return &tmpl, nil
}

func (s *ConversationService) ListTemplates(ctx context.Context) ([]models.TemplateMessage, error) {
	var templates []models.TemplateMessage
	if err := s.db.WithContext(ctx).Order("name ASC").Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	return templates, nil
}

// SendTemplate sends a template message into an order conversation.
func (s *ConversationService) SendTemplate(ctx context.Context, orderID, userID, templateID uint, role string) (*models.Conversation, error) {
	var tmpl models.TemplateMessage
	if err := s.db.WithContext(ctx).First(&tmpl, templateID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, fmt.Errorf("get template: %w", err)
	}

	return s.PostMessage(ctx, orderID, userID, role, PostMessageInput{Message: tmpl.Content})
}

// --- Sensitive Word Helpers ---

func containsSensitiveWord(msg string) string {
	lower := strings.ToLower(msg)
	for _, w := range sensitiveWords {
		if strings.Contains(lower, w) {
			return w
		}
	}
	return ""
}

// ContainsSensitiveWord is exported for testing.
func ContainsSensitiveWord(msg string) string {
	return containsSensitiveWord(msg)
}

// --- Rate Limiter (sliding window per user) ---

// RateLimiter implements a sliding window rate limiter per user ID.
type RateLimiter struct {
	maxRequests int
	window      time.Duration
	mu          sync.Mutex
	timestamps  map[uint][]time.Time
}

func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		maxRequests: maxRequests,
		window:      window,
		timestamps:  make(map[uint][]time.Time),
	}
}

// Allow returns true if the user is within rate limits.
func (r *RateLimiter) Allow(userID uint) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	// Filter expired timestamps
	valid := make([]time.Time, 0)
	for _, ts := range r.timestamps[userID] {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}

	if len(valid) >= r.maxRequests {
		r.timestamps[userID] = valid
		return false
	}

	r.timestamps[userID] = append(valid, now)
	return true
}
