package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDeviceService(t *testing.T) {
	svc := NewDeviceService(nil)
	assert.NotNil(t, svc)
}

func TestCreateDeviceInput_Fields(t *testing.T) {
	in := CreateDeviceInput{
		Name:         "Soil Sensor",
		Type:         "sensor",
		SerialNumber: "SN-001",
		PlotID:       1,
		Status:       "active",
	}
	assert.Equal(t, "Soil Sensor", in.Name)
	assert.Equal(t, uint(1), in.PlotID)
}

func TestUpdateDeviceInput_Pointers(t *testing.T) {
	name := "Updated Sensor"
	status := "inactive"
	in := UpdateDeviceInput{
		Name:   &name,
		Status: &status,
	}
	assert.Equal(t, "Updated Sensor", *in.Name)
	assert.Equal(t, "inactive", *in.Status)
	assert.Nil(t, in.Type)
	assert.Nil(t, in.PlotID)
}

func TestIsDuplicateEntry(t *testing.T) {
	assert.False(t, isDuplicateEntry(nil))
}

func TestContainsDuplicateMsg(t *testing.T) {
	assert.True(t, containsDuplicateMsg("Error 1062: Duplicate entry 'SN-001'"))
	assert.True(t, containsDuplicateMsg("Duplicate entry for key"))
	assert.False(t, containsDuplicateMsg("some other error"))
	assert.False(t, containsDuplicateMsg(""))
}

func TestErrDeviceNotFound(t *testing.T) {
	assert.EqualError(t, ErrDeviceNotFound, "device not found")
}

func TestErrDuplicateSerial(t *testing.T) {
	assert.EqualError(t, ErrDuplicateSerial, "serial number already exists")
}

func TestErrDeviceForbidden(t *testing.T) {
	assert.EqualError(t, ErrDeviceForbidden, "not authorized to access this device")
}

func TestDeviceListParams_IsolationFields(t *testing.T) {
	p := DeviceListParams{
		Page:     1,
		PageSize: 20,
		UserID:   42,
		Role:     "researcher",
	}
	assert.Equal(t, uint(42), p.UserID)
	assert.Equal(t, "researcher", p.Role)
}
