package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTaskService(t *testing.T) {
	svc := NewTaskService(nil)
	assert.NotNil(t, svc)
}

func TestCreateTaskInput_Fields(t *testing.T) {
	in := CreateTaskInput{Title: "Test crop", Description: "Desc", AssignedTo: 1, DueDate: "2026-05-01T00:00:00Z"}
	assert.Equal(t, "Test crop", in.Title)
	assert.Equal(t, uint(1), in.AssignedTo)
}

func TestUpdateTaskInput_Pointers(t *testing.T) {
	s := "completed"
	in := UpdateTaskInput{Status: &s}
	assert.Equal(t, "completed", *in.Status)
	assert.Nil(t, in.Title)
}

func TestErrTaskNotFound(t *testing.T) {
	assert.EqualError(t, ErrTaskNotFound, "task not found")
}
