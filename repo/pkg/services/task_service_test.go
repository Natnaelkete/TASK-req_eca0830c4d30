package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTaskService(t *testing.T) {
	svc := NewTaskService(nil)
	assert.NotNil(t, svc)
}

func TestCreateTaskInput_Fields(t *testing.T) {
	in := CreateTaskInput{
		Title:      "Test crop",
		ObjectID:   1,
		ObjectType: "plot",
		CycleType:  "monthly",
		AssignedTo: 1,
		DueEnd:     "2026-05-01T00:00:00Z",
	}
	assert.Equal(t, "Test crop", in.Title)
	assert.Equal(t, uint(1), in.ObjectID)
	assert.Equal(t, "plot", in.ObjectType)
}

func TestUpdateTaskInput_Pointers(t *testing.T) {
	s := "New title"
	in := UpdateTaskInput{Title: &s}
	assert.Equal(t, "New title", *in.Title)
	assert.Nil(t, in.Description)
}

func TestGenerateTasksInput_Fields(t *testing.T) {
	in := GenerateTasksInput{
		ObjectID:   1,
		ObjectType: "result",
		CycleType:  "weekly",
		Title:      "Weekly Review",
		AssignedTo: 3,
		Count:      4,
	}
	assert.Equal(t, "weekly", in.CycleType)
	assert.Equal(t, 4, in.Count)
}

func TestOverdueThreshold(t *testing.T) {
	assert.Equal(t, 7*24*time.Hour, OverdueThreshold)
}

func TestErrTaskNotFound(t *testing.T) {
	assert.EqualError(t, ErrTaskNotFound, "task not found")
}

func TestErrTaskInvalidStatus(t *testing.T) {
	assert.EqualError(t, ErrTaskInvalidStatus, "invalid task status transition")
}
