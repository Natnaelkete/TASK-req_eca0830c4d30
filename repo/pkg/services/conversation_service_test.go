package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewConversationService(t *testing.T) {
	svc := NewConversationService(nil)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.limiter)
}

func TestCreateOrderInput_Fields(t *testing.T) {
	in := CreateOrderInput{Title: "Test Order"}
	assert.Equal(t, "Test Order", in.Title)
}

func TestPostMessageInput_Fields(t *testing.T) {
	in := PostMessageInput{Message: "Hello"}
	assert.Equal(t, "Hello", in.Message)
}

func TestTransferInput_Fields(t *testing.T) {
	in := TransferInput{TransferToUserID: 5, Reason: "Reassigning"}
	assert.Equal(t, uint(5), in.TransferToUserID)
	assert.Equal(t, "Reassigning", in.Reason)
}

func TestCreateTemplateInput_Fields(t *testing.T) {
	in := CreateTemplateInput{Name: "Welcome", Content: "Welcome to the system!"}
	assert.Equal(t, "Welcome", in.Name)
}

func TestContainsSensitiveWord(t *testing.T) {
	tests := []struct {
		msg    string
		expect string
	}{
		{"This is illegal content", "illegal"},
		{"Stop sending spam messages", "spam"},
		{"This is a fraud attempt", "fraud"},
		{"Normal agricultural data", ""},
		{"Hello world", ""},
		{"ILLEGAL uppercase", "illegal"},
		{"Some Spam here", "spam"},
	}
	for _, tt := range tests {
		result := ContainsSensitiveWord(tt.msg)
		assert.Equal(t, tt.expect, result, "msg=%s", tt.msg)
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(3, time.Minute)

	// First 3 should be allowed
	assert.True(t, limiter.Allow(1))
	assert.True(t, limiter.Allow(1))
	assert.True(t, limiter.Allow(1))

	// 4th should be rejected
	assert.False(t, limiter.Allow(1))

	// Different user should be allowed
	assert.True(t, limiter.Allow(2))
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	// Use a very short window
	limiter := NewRateLimiter(1, 50*time.Millisecond)

	assert.True(t, limiter.Allow(1))
	assert.False(t, limiter.Allow(1))

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	// Should be allowed again
	assert.True(t, limiter.Allow(1))
}

func TestErrOrderNotFound(t *testing.T) {
	assert.EqualError(t, ErrOrderNotFound, "order not found")
}

func TestErrConversationNotFound(t *testing.T) {
	assert.EqualError(t, ErrConversationNotFound, "conversation not found")
}

func TestErrRateLimitExceeded(t *testing.T) {
	assert.EqualError(t, ErrRateLimitExceeded, "rate limit exceeded: max 20 messages per minute")
}

func TestErrSensitiveWord(t *testing.T) {
	assert.EqualError(t, ErrSensitiveWord, "message contains sensitive content and was blocked")
}

func TestErrTemplateNotFound(t *testing.T) {
	assert.EqualError(t, ErrTemplateNotFound, "template not found")
}

func TestOrderListParams_Defaults(t *testing.T) {
	p := OrderListParams{}
	assert.Equal(t, 0, p.Page)
	assert.Equal(t, uint(0), p.UserID)
}
