package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIndicatorService(t *testing.T) {
	svc := NewIndicatorService(nil)
	assert.NotNil(t, svc)
}

func TestCreateIndicatorInput_Fields(t *testing.T) {
	in := CreateIndicatorInput{
		Code:        "soil.ph",
		Name:        "Soil pH",
		Description: "Acidity of soil",
		Unit:        "pH",
		Formula:     "avg(raw)",
		Category:    "soil",
	}
	assert.Equal(t, "soil.ph", in.Code)
	assert.Equal(t, "Soil pH", in.Name)
	assert.Equal(t, "soil", in.Category)
}

func TestUpdateIndicatorInput_Fields(t *testing.T) {
	name := "Soil pH v2"
	in := UpdateIndicatorInput{
		Name:        &name,
		DiffSummary: "renamed",
	}
	assert.NotNil(t, in.Name)
	assert.Equal(t, "renamed", in.DiffSummary)
}

func TestIndicatorListParams_Defaults(t *testing.T) {
	p := IndicatorListParams{}
	assert.Equal(t, 0, p.Page)
	assert.Equal(t, 0, p.PageSize)
	assert.Equal(t, "", p.Category)
}

func TestErrIndicatorNotFound(t *testing.T) {
	assert.EqualError(t, ErrIndicatorNotFound, "indicator not found")
}

func TestErrIndicatorExists(t *testing.T) {
	assert.EqualError(t, ErrIndicatorExists, "indicator code already exists")
}

func TestPaginatedIndicators_Structure(t *testing.T) {
	p := PaginatedIndicators{Total: 10, Page: 2, PageSize: 5, TotalPages: 2}
	assert.Equal(t, int64(10), p.Total)
	assert.Equal(t, 2, p.TotalPages)
}
