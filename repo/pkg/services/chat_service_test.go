package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewChatService(t *testing.T) {
	svc := NewChatService(nil)
	assert.NotNil(t, svc)
}

func TestSendMessageInput_Fields(t *testing.T) {
	plotID := uint(1)
	in := SendMessageInput{ReceiverID: 2, PlotID: &plotID, Content: "Hello"}
	assert.Equal(t, uint(2), in.ReceiverID)
	assert.Equal(t, "Hello", in.Content)
	assert.NotNil(t, in.PlotID)
}

func TestErrMessageNotFound(t *testing.T) {
	assert.EqualError(t, ErrMessageNotFound, "message not found")
}
