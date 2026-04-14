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
	in := CreateResultInput{
		Type:    "paper",
		PlotID:  1,
		Title:   "Yield Report",
		Summary: "Good",
		Fields:  `{"abstract":"test"}`,
	}
	assert.Equal(t, "paper", in.Type)
	assert.Equal(t, "Yield Report", in.Title)
	assert.Equal(t, uint(1), in.PlotID)
}

func TestUpdateResultInput_Pointers(t *testing.T) {
	s := "Updated summary"
	in := UpdateResultInput{Summary: &s}
	assert.Equal(t, "Updated summary", *in.Summary)
	assert.Nil(t, in.Title)
}

func TestIsValidTransition(t *testing.T) {
	tests := []struct {
		from, to string
		valid    bool
	}{
		{"draft", "submitted", true},
		{"submitted", "returned", true},
		{"submitted", "approved", true},
		{"returned", "submitted", true},
		{"approved", "archived", true},
		// Invalid transitions
		{"draft", "approved", false},
		{"draft", "archived", false},
		{"submitted", "draft", false},
		{"submitted", "archived", false},
		{"approved", "draft", false},
		{"archived", "draft", false},
		{"archived", "submitted", false},
		{"unknown", "submitted", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.valid, IsValidTransition(tt.from, tt.to),
			"from=%s to=%s", tt.from, tt.to)
	}
}

func TestCreateFieldRuleInput_Fields(t *testing.T) {
	in := CreateFieldRuleInput{
		ResultType: "paper",
		FieldName:  "abstract",
		Required:   true,
		MaxLength:  5000,
		EnumValues: "",
	}
	assert.Equal(t, "paper", in.ResultType)
	assert.True(t, in.Required)
	assert.Equal(t, 5000, in.MaxLength)
}

func TestErrResultNotFound(t *testing.T) {
	assert.EqualError(t, ErrResultNotFound, "result not found")
}

func TestErrInvalidTransition(t *testing.T) {
	assert.EqualError(t, ErrInvalidTransition, "invalid status transition")
}

func TestErrResultArchived(t *testing.T) {
	assert.EqualError(t, ErrResultArchived, "archived results cannot be modified")
}

func TestErrInvalidationNoReason(t *testing.T) {
	assert.EqualError(t, ErrInvalidationNoReason, "invalidation requires a reason")
}

func TestErrFieldValidationFailed(t *testing.T) {
	assert.EqualError(t, ErrFieldValidationFailed, "field validation failed")
}

func TestErrResultForbidden(t *testing.T) {
	assert.EqualError(t, ErrResultForbidden, "not authorized to access this result")
}

func TestResultListParams_IsolationFields(t *testing.T) {
	p := ResultListParams{
		Page:     1,
		PageSize: 20,
		UserID:   42,
		Role:     "researcher",
	}
	assert.Equal(t, uint(42), p.UserID)
	assert.Equal(t, "researcher", p.Role)
}
