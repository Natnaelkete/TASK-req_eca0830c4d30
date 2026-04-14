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
	reviewerID := uint(2)
	tk := Task{
		ID:          1,
		Title:       "Evaluate crop yield",
		Description: "Measure output from Field A",
		ObjectID:    10,
		ObjectType:  "plot",
		CycleType:   "monthly",
		Status:      "pending",
		AssignedTo:  1,
		ReviewerID:  &reviewerID,
		DueEnd:      &due,
	}
	assert.Equal(t, "Evaluate crop yield", tk.Title)
	assert.Equal(t, "pending", tk.Status)
	assert.Equal(t, uint(10), tk.ObjectID)
	assert.Equal(t, "plot", tk.ObjectType)
	assert.NotNil(t, tk.ReviewerID)
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
		ID:          1,
		Type:        "paper",
		PlotID:      1,
		TaskID:      &taskID,
		Title:       "Yield Analysis",
		Summary:     "Positive results",
		Fields:      `{"abstract":"test"}`,
		Status:      "draft",
		SubmitterID: 1,
		CreatedBy:   1,
	}
	assert.Equal(t, "paper", r.Type)
	assert.Equal(t, "Yield Analysis", r.Title)
	assert.Equal(t, "draft", r.Status)
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

func TestOrderTableName(t *testing.T) {
	assert.Equal(t, "orders", Order{}.TableName())
}

func TestOrderStruct(t *testing.T) {
	o := Order{ID: 1, ResearcherID: 2, Title: "Research Order", Status: "open", AssignedTo: 2}
	assert.Equal(t, "Research Order", o.Title)
	assert.Equal(t, "open", o.Status)
}

func TestConversationTableName(t *testing.T) {
	assert.Equal(t, "conversations", Conversation{}.TableName())
}

func TestConversationStruct(t *testing.T) {
	transferTo := uint(5)
	c := Conversation{ID: 1, OrderID: 1, UserID: 2, Message: "Hello", TransferredTo: &transferTo}
	assert.Equal(t, "Hello", c.Message)
	assert.NotNil(t, c.TransferredTo)
}

func TestTemplateMessageTableName(t *testing.T) {
	assert.Equal(t, "template_messages", TemplateMessage{}.TableName())
}

func TestTemplateMessageStruct(t *testing.T) {
	tm := TemplateMessage{ID: 1, Name: "Welcome", Content: "Welcome!"}
	assert.Equal(t, "Welcome", tm.Name)
}

func TestSensitiveWordLogTableName(t *testing.T) {
	assert.Equal(t, "sensitive_word_logs", SensitiveWordLog{}.TableName())
}

func TestSensitiveWordLogStruct(t *testing.T) {
	sl := SensitiveWordLog{ID: 1, UserID: 2, OrderID: 3, Content: "illegal stuff", Word: "illegal"}
	assert.Equal(t, "illegal", sl.Word)
}

func TestFieldRuleTableName(t *testing.T) {
	assert.Equal(t, "field_rules", FieldRule{}.TableName())
}

func TestFieldRuleStruct(t *testing.T) {
	fr := FieldRule{ID: 1, ResultType: "paper", FieldName: "abstract", Required: true, MaxLength: 5000}
	assert.True(t, fr.Required)
	assert.Equal(t, 5000, fr.MaxLength)
}

func TestResultStatusLogTableName(t *testing.T) {
	assert.Equal(t, "result_status_logs", ResultStatusLog{}.TableName())
}

func TestResultStatusLogStruct(t *testing.T) {
	rsl := ResultStatusLog{ID: 1, ResultID: 10, FromStatus: "draft", ToStatus: "submitted", ChangedBy: 2, Reason: "Ready"}
	assert.Equal(t, "submitted", rsl.ToStatus)
	assert.Equal(t, "Ready", rsl.Reason)
}

func TestSystemNotificationTableName(t *testing.T) {
	assert.Equal(t, "system_notifications", SystemNotification{}.TableName())
}

func TestSystemNotificationStruct(t *testing.T) {
	sn := SystemNotification{ID: 1, Type: "capacity", Message: "Disk usage at 85%", Level: "warning"}
	assert.Equal(t, "capacity", sn.Type)
	assert.Equal(t, "warning", sn.Level)
}
