package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDashboardService(t *testing.T) {
	svc := NewDashboardService(nil)
	assert.NotNil(t, svc)
}

func TestCreateDashboardInput_Fields(t *testing.T) {
	in := CreateDashboardInput{
		Name:   "My Dashboard",
		Config: `{"plots":[1,2],"metrics":["temperature"],"time_window":"24h","chart_type":"line"}`,
	}
	assert.Equal(t, "My Dashboard", in.Name)
	assert.Contains(t, in.Config, "temperature")
}

func TestUpdateDashboardInput_Fields(t *testing.T) {
	in := UpdateDashboardInput{
		Name:   "Updated Dashboard",
		Config: `{"plots":[3],"metrics":["humidity"]}`,
	}
	assert.Equal(t, "Updated Dashboard", in.Name)
	assert.Contains(t, in.Config, "humidity")
}

func TestDashboardListParams_Defaults(t *testing.T) {
	p := DashboardListParams{}
	assert.Equal(t, 0, p.Page)
	assert.Equal(t, 0, p.PageSize)
	assert.Equal(t, uint(0), p.UserID)
}

func TestErrDashboardNotFound(t *testing.T) {
	assert.EqualError(t, ErrDashboardNotFound, "dashboard config not found")
}

func TestErrDashboardForbidden(t *testing.T) {
	assert.EqualError(t, ErrDashboardForbidden, "not authorized to access this dashboard")
}
