package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPlotService(t *testing.T) {
	svc := NewPlotService(nil)
	assert.NotNil(t, svc)
}

func TestPlotListParams_Defaults(t *testing.T) {
	p := PlotListParams{}
	assert.Equal(t, 0, p.Page)
	assert.Equal(t, 0, p.PageSize)
}

func TestCreatePlotInput_Fields(t *testing.T) {
	in := CreatePlotInput{
		Name:     "Field A",
		Location: "North",
		Area:     5.5,
		SoilType: "Clay",
		CropType: "Wheat",
	}
	assert.Equal(t, "Field A", in.Name)
	assert.Equal(t, 5.5, in.Area)
}

func TestUpdatePlotInput_Pointers(t *testing.T) {
	name := "Updated"
	area := 10.0
	in := UpdatePlotInput{
		Name: &name,
		Area: &area,
	}
	assert.Equal(t, "Updated", *in.Name)
	assert.Equal(t, 10.0, *in.Area)
	assert.Nil(t, in.Location)
	assert.Nil(t, in.SoilType)
	assert.Nil(t, in.CropType)
}

func TestErrPlotNotFound(t *testing.T) {
	assert.EqualError(t, ErrPlotNotFound, "plot not found")
}
