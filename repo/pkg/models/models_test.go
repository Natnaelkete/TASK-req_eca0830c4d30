package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserTableName(t *testing.T) {
	assert.Equal(t, "users", User{}.TableName())
}

func TestPlotTableName(t *testing.T) {
	assert.Equal(t, "plots", Plot{}.TableName())
}

func TestDeviceTableName(t *testing.T) {
	assert.Equal(t, "devices", Device{}.TableName())
}

func TestMetricTableName(t *testing.T) {
	assert.Equal(t, "metrics", Metric{}.TableName())
}

func TestTaskTableName(t *testing.T) {
	assert.Equal(t, "tasks", Task{}.TableName())
}

func TestAuditLogTableName(t *testing.T) {
	assert.Equal(t, "audit_logs", AuditLog{}.TableName())
}

func TestUserStruct(t *testing.T) {
	u := User{
		ID:           1,
		Username:     "admin",
		Email:        "admin@test.com",
		PasswordHash: "hash",
		Role:         "admin",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	assert.Equal(t, uint(1), u.ID)
	assert.Equal(t, "admin", u.Username)
	assert.Equal(t, "admin@test.com", u.Email)
	assert.Equal(t, "admin", u.Role)
}

func TestPlotStruct(t *testing.T) {
	p := Plot{
		ID:       1,
		Name:     "Field A",
		Location: "North Campus",
		Area:     2.5,
		SoilType: "Clay",
		CropType: "Wheat",
		UserID:   1,
	}
	assert.Equal(t, "Field A", p.Name)
	assert.Equal(t, 2.5, p.Area)
	assert.Equal(t, "Wheat", p.CropType)
}

func TestDeviceStruct(t *testing.T) {
	now := time.Now()
	d := Device{
		ID:           1,
		Name:         "Soil Sensor",
		Type:         "sensor",
		SerialNumber: "SN-001",
		PlotID:       1,
		Status:       "active",
		InstalledAt:  &now,
	}
	assert.Equal(t, "Soil Sensor", d.Name)
	assert.Equal(t, "SN-001", d.SerialNumber)
	assert.NotNil(t, d.InstalledAt)
}

func TestMetricStruct(t *testing.T) {
	m := Metric{
		ID:         1,
		DeviceID:   1,
		MetricType: "temperature",
		Value:      23.5,
		Unit:       "celsius",
		EventTime:  time.Now(),
	}
	assert.Equal(t, "temperature", m.MetricType)
	assert.Equal(t, 23.5, m.Value)
}

func TestTaskStruct(t *testing.T) {
	due := time.Now().Add(24 * time.Hour)
	tk := Task{
		ID:          1,
		Title:       "Evaluate crop yield",
		Description: "Measure output from Field A",
		Status:      "pending",
		AssignedTo:  1,
		DueDate:     &due,
	}
	assert.Equal(t, "Evaluate crop yield", tk.Title)
	assert.Equal(t, "pending", tk.Status)
}

func TestAuditLogStruct(t *testing.T) {
	a := AuditLog{
		ID:         1,
		UserID:     1,
		Action:     "CREATE",
		Resource:   "plots",
		ResourceID: 5,
		IPAddress:  "192.168.1.1",
	}
	assert.Equal(t, "CREATE", a.Action)
	assert.Equal(t, "plots", a.Resource)
}

func TestAlertTableName(t *testing.T) {
	assert.Equal(t, "alerts", Alert{}.TableName())
}

func TestAlertStruct(t *testing.T) {
	a := Alert{
		ID:         1,
		DeviceID:   2,
		MetricType: "temperature",
		Value:      35.0,
		Threshold:  30.0,
		Level:      "critical",
		Message:    "threshold exceeded",
		Resolved:   false,
	}
	assert.Equal(t, "critical", a.Level)
	assert.Equal(t, 35.0, a.Value)
	assert.False(t, a.Resolved)
}

func TestMessageTableName(t *testing.T) {
	assert.Equal(t, "messages", Message{}.TableName())
}

func TestMessageStruct(t *testing.T) {
	plotID := uint(1)
	m := Message{
		ID:         1,
		SenderID:   1,
		ReceiverID: 2,
		PlotID:     &plotID,
		Content:    "Check field A",
		Read:       false,
	}
	assert.Equal(t, uint(1), m.SenderID)
	assert.Equal(t, "Check field A", m.Content)
	assert.False(t, m.Read)
}

func TestResultTableName(t *testing.T) {
	assert.Equal(t, "results", Result{}.TableName())
}

func TestResultStruct(t *testing.T) {
	taskID := uint(1)
	r := Result{
		ID:        1,
		PlotID:    1,
		TaskID:    &taskID,
		Title:     "Yield Analysis",
		Summary:   "Positive results",
		Data:      `{"yield":500}`,
		CreatedBy: 1,
	}
	assert.Equal(t, "Yield Analysis", r.Title)
	assert.Equal(t, uint(1), r.PlotID)
	assert.NotNil(t, r.TaskID)
}

func TestMonitoringDataTableName(t *testing.T) {
	assert.Equal(t, "monitoring_data", MonitoringData{}.TableName())
}

func TestMonitoringDataStruct(t *testing.T) {
	now := time.Now()
	m := MonitoringData{
		ID:         1,
		SourceID:   "sensor-001-20260115",
		DeviceID:   5,
		PlotID:     3,
		MetricCode: "temperature",
		Value:      23.5,
		Unit:       "celsius",
		EventTime:  now,
		Tags:       `{"location":"field-A","zone":"north"}`,
	}
	assert.Equal(t, "sensor-001-20260115", m.SourceID)
	assert.Equal(t, uint(5), m.DeviceID)
	assert.Equal(t, uint(3), m.PlotID)
	assert.Equal(t, "temperature", m.MetricCode)
	assert.Equal(t, 23.5, m.Value)
	assert.Equal(t, "celsius", m.Unit)
	assert.Equal(t, now, m.EventTime)
	assert.Contains(t, m.Tags, "field-A")
}

func TestDashboardConfigTableName(t *testing.T) {
	assert.Equal(t, "dashboard_configs", DashboardConfig{}.TableName())
}

func TestDashboardConfigStruct(t *testing.T) {
	cfg := DashboardConfig{
		ID:     1,
		UserID: 2,
		Name:   "My Dashboard",
		Config: `{"plots":[1,2],"metrics":["temperature","humidity"],"time_window":"24h","chart_type":"line"}`,
	}
	assert.Equal(t, uint(2), cfg.UserID)
	assert.Equal(t, "My Dashboard", cfg.Name)
	assert.Contains(t, cfg.Config, "temperature")
	assert.Contains(t, cfg.Config, "line")
}
