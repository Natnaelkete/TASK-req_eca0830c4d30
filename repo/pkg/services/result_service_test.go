package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResultService(t *testing.T) {
	svc := NewResultService(nil)
	assert.NotNil(t, svc)
}

func TestCreateResultInput_Fields(t *testing.T) {
	in := CreateResultInput{PlotID: 1, Title: "Yield Report", Summary: "Good", Data: `{"yield":500}`}
	assert.Equal(t, "Yield Report", in.Title)
	assert.Equal(t, uint(1), in.PlotID)
}

func TestUpdateResultInput_Pointers(t *testing.T) {
	s := "Updated summary"
	in := UpdateResultInput{Summary: &s}
	assert.Equal(t, "Updated summary", *in.Summary)
	assert.Nil(t, in.Title)
}

func TestErrResultNotFound(t *testing.T) {
	assert.EqualError(t, ErrResultNotFound, "result not found")
}
