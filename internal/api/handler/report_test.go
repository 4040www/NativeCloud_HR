// report_handlers.go
package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/4040www/NativeCloud_HR/internal/db"
	"github.com/4040www/NativeCloud_HR/internal/model"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New() failed: %v", err)
	}
	dialector := postgres.New(postgres.Config{
		Conn: mockDB,
	})
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		t.Fatalf("gorm.Open() failed: %v", err)
	}

	return db, mock
}

func TestGetMyTodayRecords(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock := setupMockDB(t)

	userID := "user-123"

	// 1. 模擬 access_log 查詢（這個實際上先被呼叫）
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3 ORDER BY access_time asc`).
		WithArgs(userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "direction", "access_time", "gate_name"}).
			AddRow(userID, "IN", time.Date(2025, 5, 23, 9, 0, 0, 0, time.UTC), "A1").
			AddRow(userID, "OUT", time.Date(2025, 5, 23, 18, 0, 0, 0, time.UTC), "B2"))

	// 2. 模擬 employee 查詢（這個後被呼叫）
	mock.ExpectQuery(`SELECT \* FROM "employee" WHERE employee_id = \$1 ORDER BY "employee"."employee_id" LIMIT \$2`).
		WithArgs(userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"employee_id", "first_name", "last_name", "is_manager", "password", "email", "organization_id",
		}).AddRow(userID, "John", "Doe", false, "hashedPassword", "john.doe@example.com", "org-123"))

	req, err := http.NewRequest(http.MethodGet, "/reports/today", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "userID", Value: userID}}

	handler := GetMyTodayRecords(db)
	handler(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", w.Code)
	}

	expected := `{"data":{"date":"2025-05-23","name":"John Doe","clock_in_time":"09:00","clock_out_time":"18:00","clock_in_gate":"A1","clock_out_gate":"B2","status":"Late"}}`

	if strings.TrimSpace(w.Body.String()) != expected {
		t.Errorf("unexpected body: got %s, want %s", w.Body.String(), expected)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
func TestGetMyHistoryRecords(t *testing.T) {
	db, mock := setupMockDB(t)

	userID := "user-123"
	loc, _ := time.LoadLocation("Asia/Taipei")

	clockInTime := time.Date(2025, 4, 24, 9, 0, 0, 0, loc)
	clockOutTime := time.Date(2025, 4, 24, 18, 0, 0, 0, loc)

	// 模擬 access_log 查詢
	accessLogs := sqlmock.NewRows([]string{
		"access_id", "employee_id", "access_time", "direction", "gate_type", "gate_name", "access_result",
	}).AddRow(
		"access-001", userID, clockInTime, "IN", "TypeA", "Gate1", "Success",
	).AddRow(
		"access-002", userID, clockOutTime, "OUT", "TypeB", "Gate2", "Success",
	)

	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3 ORDER BY access_time asc`).
		WithArgs(userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(accessLogs)

	// 模擬 employee 查詢
	employeeRows := sqlmock.NewRows([]string{"employee_id", "first_name", "last_name"}).
		AddRow(userID, "Test", "User")

	mock.ExpectQuery(`SELECT \* FROM "employee" WHERE employee_id = \$1 ORDER BY "employee"."employee_id" LIMIT \$2`).
		WithArgs(userID, 1).
		WillReturnRows(employeeRows)

	// 建立 gin 路由和測試請求
	router := gin.Default()
	router.GET("/history/:userID", GetMyHistoryRecords(db))

	req, _ := http.NewRequest("GET", "/history/"+userID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var resp []model.AttendanceSummary
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	want := []model.AttendanceSummary{
		{
			Date:         "2025-04-24",
			Name:         "Test User",
			ClockInTime:  "09:00",
			ClockOutTime: "18:00",
			ClockInGate:  "Gate1",
			ClockOutGate: "Gate2",
			Status:       "Late",
		},
	}

	if !reflect.DeepEqual(resp, want) {
		t.Errorf("Response = %+v, want %+v", resp, want)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestGetMyPeriodRecords(t *testing.T) {
	db, mock := setupMockDB(t)

	userID := "user-123"
	loc, _ := time.LoadLocation("Asia/Taipei")
	start := time.Date(2025, 4, 24, 0, 0, 0, 0, loc)
	end := start // 只測一天，結束日期同一天

	// 模擬 access_log，9:00 IN，18:00 OUT
	accessLogs := sqlmock.NewRows([]string{
		"access_id", "employee_id", "access_time", "direction", "gate_type", "gate_name", "access_result",
	}).AddRow(
		"access-001", userID, start.Add(9*time.Hour), "IN", "TypeA", "Gate1", "Success",
	).AddRow(
		"access-002", userID, start.Add(18*time.Hour), "OUT", "TypeB", "Gate2", "Success",
	)

	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3 ORDER BY access_time asc`).
		WithArgs(userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(accessLogs)

	mock.ExpectQuery(`SELECT \* FROM "employee" WHERE employee_id = \$1 ORDER BY "employee"."employee_id" LIMIT \$2`).
		WithArgs(userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name"}).
			AddRow(userID, "Test", "User"))

	router := gin.Default()
	router.GET("/period/:userID/:startDate/:endDate", GetMyPeriodRecords(db))

	url := fmt.Sprintf("/period/%s/%s/%s", userID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	req, _ := http.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var got []model.AttendanceSummary
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	want := []model.AttendanceSummary{
		{
			Date:         start.Format("2006-01-02"),
			Name:         "Test User",
			ClockInTime:  "09:00",
			ClockOutTime: "18:00",
			ClockInGate:  "Gate1",
			ClockOutGate: "Gate2",
			Status:       "Late",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Response = %+v, want %+v", got, want)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// new

func TestGetAlertList(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	db.DB = mockDB

	// Mock 員工基本資料
	mock.ExpectQuery(`SELECT \* FROM "employee"`).
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name"}).
			AddRow("E1", "Alice", "Normal").
			AddRow("E2", "Bob", "Warning").
			AddRow("E3", "Charlie", "Alert"))

	// E1: Normal（1天9小時）
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3 ORDER BY access_time asc`).
		WithArgs("E1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction"}).
			AddRow(time.Date(2023, 1, 2, 8, 30, 0, 0, time.UTC), "IN").
			AddRow(time.Date(2023, 1, 2, 17, 30, 0, 0, time.UTC), "OUT"))

	// E2: Warning（2天10.5小時）
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3 ORDER BY access_time asc`).
		WithArgs("E2", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction"}).
			AddRow(time.Date(2023, 1, 1, 8, 30, 0, 0, time.UTC), "IN").
			AddRow(time.Date(2023, 1, 1, 19, 0, 0, 0, time.UTC), "OUT").
			AddRow(time.Date(2023, 1, 2, 8, 30, 0, 0, time.UTC), "IN").
			AddRow(time.Date(2023, 1, 2, 19, 0, 0, 0, time.UTC), "OUT"))

	// E3: Alert（1天13小時）
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3 ORDER BY access_time asc`).
		WithArgs("E3", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction"}).
			AddRow(time.Date(2023, 1, 2, 8, 30, 0, 0, time.UTC), "IN").
			AddRow(time.Date(2023, 1, 2, 21, 30, 0, 0, time.UTC), "OUT"))

	// 建立 HTTP 測試 context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{
		{Key: "startDate", Value: "2023-01-01"},
		{Key: "endDate", Value: "2023-01-02"},
	}

	// 執行 Handler
	handler := GetAlertList(mockDB)
	handler(c)

	// 解析回傳
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	}

	var got []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// 驗證結果
	expected := []map[string]interface{}{
		{
			"EmployeeID": "E1",
			"Name":       "Alice Normal",
			"OTCounts":   float64(1),
			"OTHours":    float64(1),
			"status":     "Normal",
		},
		{
			"EmployeeID": "E2",
			"Name":       "Bob Warning",
			"OTCounts":   float64(2),
			"OTHours":    float64(5),
			"status":     "Warning",
		},
		{
			"EmployeeID": "E3",
			"Name":       "Charlie Alert",
			"OTCounts":   float64(1),
			"OTHours":    float64(5),
			"status":     "Alert",
		},
	}

	gotJSON, _ := json.MarshalIndent(got, "", "  ")
	wantJSON, _ := json.MarshalIndent(expected, "", "  ")
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("handler returned unexpected body:\nGot:\n%s\nWant:\n%s", gotJSON, wantJSON)
	}
}

func TestGetInChargeDepartments(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	db.DB = mockDB // 如果 repository 用的是全域 db

	// 🧪 模擬 user1 管理 A, B 部門
	mock.ExpectQuery(`SELECT distinct organization_id FROM "employee" WHERE employee_id = \$1`).
		WithArgs("user1").
		WillReturnRows(sqlmock.NewRows([]string{"organization_id"}).
			AddRow("A").
			AddRow("B"))

	// 🧪 模擬 user2 沒有管理部門
	mock.ExpectQuery(`SELECT distinct organization_id FROM "employee" WHERE employee_id = \$1`).
		WithArgs("user2").
		WillReturnRows(sqlmock.NewRows([]string{"organization_id"})) // 空結果

	tests := []struct {
		name     string
		userID   string
		expected []string
	}{
		{
			name:     "user with departments",
			userID:   "user1",
			expected: []string{"A", "B"},
		},
		{
			name:     "user with no departments",
			userID:   "user2",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 建立測試 context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = []gin.Param{
				{Key: "userID", Value: tt.userID},
			}

			// 呼叫 handler
			GetInChargeDepartments(c)

			// 驗證回應狀態碼
			if w.Code != http.StatusOK {
				t.Errorf("unexpected status code: got %d, want %d", w.Code, http.StatusOK)
			}

			// 驗證 JSON 回傳
			var got []string
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("invalid JSON response: %v", err)
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("unexpected departments:\nGot:  %v\nWant: %v", got, tt.expected)
			}
		})
	}
}

func TestExportSummaryCSV(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	db.DB = mockDB // 假設 repository 使用全域變數 db

	// 模擬一名員工
	mock.ExpectQuery(`SELECT \* FROM "employee" WHERE organization_id = \$1`).
		WithArgs("Engineering").
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name", "organization_id"}).
			AddRow("E1", "Alice", "Wang", "Engineering"))

	// 模擬打卡資料（一天內完整 IN/OUT）
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3 ORDER BY access_time asc`).
		WithArgs("E1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction", "gate_name"}).
			AddRow(time.Date(2023, 1, 2, 8, 30, 0, 0, time.UTC), "IN", "North").
			AddRow(time.Date(2023, 1, 2, 17, 30, 0, 0, time.UTC), "OUT", "South"))

	// 建立測試 context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{
		{Key: "department", Value: "Engineering"},
		{Key: "startDate", Value: "2023-01-01"},
		{Key: "endDate", Value: "2023-01-02"},
	}

	// 執行 handler
	handler := ExportSummaryCSV(mockDB)
	handler(c)

	// 驗證 HTTP 狀態碼
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want 200", w.Code)
	}

	// 驗證 Content-Type 和 Content-Disposition
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/csv") {
		t.Errorf("unexpected content-type: %s", ct)
	}
	if disp := w.Header().Get("Content-Disposition"); !strings.Contains(disp, "attachment") {
		t.Errorf("expected attachment disposition, got: %s", disp)
	}

	// 驗證 CSV 內容
	r := csv.NewReader(bytes.NewReader(w.Body.Bytes()))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	expected := [][]string{
		{"date", "employee ID", "name", "clock-in time", "clock-in gate", "clock-out time", "clock-out gate", "status"},
		{"2023-01-02", "E1", "Alice Wang", "08:30", "North", "17:30", "South", "On Time"},
	}
	if !reflect.DeepEqual(records, expected) {
		gotJSON, _ := json.MarshalIndent(records, "", "  ")
		wantJSON, _ := json.MarshalIndent(expected, "", "  ")
		t.Errorf("CSV mismatch:\nGot:\n%s\nWant:\n%s", gotJSON, wantJSON)
	}
}

func TestExportSummaryPDF(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	db.DB = mockDB // 若你的 repository 使用全域 db

	// 🧪 模擬一位員工
	mock.ExpectQuery(`SELECT \* FROM "employee" WHERE organization_id = \$1`).
		WithArgs("Engineering").
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name", "organization_id"}).
			AddRow("E1", "Alice", "Wang", "Engineering"))

	// 🧪 模擬打卡資料
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3 ORDER BY access_time asc`).
		WithArgs("E1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction", "gate_name"}).
			AddRow(time.Date(2023, 1, 2, 8, 30, 0, 0, time.UTC), "IN", "North").
			AddRow(time.Date(2023, 1, 2, 17, 30, 0, 0, time.UTC), "OUT", "South"))

	// 建立測試 HTTP 請求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{
		{Key: "department", Value: "Engineering"},
		{Key: "startDate", Value: "2023-01-01"},
		{Key: "endDate", Value: "2023-01-02"},
	}

	// 執行 handler
	handler := ExportSummaryPDF(mockDB)
	handler(c)

	// ✅ 驗證 HTTP status code
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want 200", w.Code)
	}

	// ✅ 驗證 Content-Type
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/pdf") {
		t.Errorf("unexpected content-type: got %s, want application/pdf", ct)
	}

	// ✅ 驗證 Content-Disposition
	if cd := w.Header().Get("Content-Disposition"); !strings.Contains(cd, "attachment") || !strings.Contains(cd, "summary.pdf") {
		t.Errorf("unexpected content-disposition: %s", cd)
	}

	// ✅ 驗證 PDF 內容不是空的
	if len(w.Body.Bytes()) < 1000 {
		t.Errorf("PDF too small or empty, size = %d bytes", len(w.Body.Bytes()))
	}

	// 🧪 可選：輸出 PDF 幫助手動檢查
	// os.WriteFile("test_summary.pdf", w.Body.Bytes(), 0644)
}

func TestFilterAttendanceSummary(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	db.DB = mockDB // 如果你的 repository 使用全域 db

	// 🧪 模擬一名員工
	mock.ExpectQuery(`SELECT \* FROM "employee" WHERE organization_id = \$1`).
		WithArgs("Engineering").
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name", "organization_id"}).
			AddRow("E1", "Alice", "Wang", "Engineering"))

	// 🧪 模擬該員工的 access log
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3 ORDER BY access_time asc`).
		WithArgs("E1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction", "gate_name"}).
			AddRow(time.Date(2023, 1, 2, 8, 30, 0, 0, time.UTC), "IN", "North").
			AddRow(time.Date(2023, 1, 2, 17, 30, 0, 0, time.UTC), "OUT", "South"))

	// 建立 HTTP 測試 context（模擬 query string）
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/?department=Engineering&fromDate=2023-01-01&toDate=2023-01-02", nil)
	c.Request = req

	// 呼叫 handler
	handler := FilterAttendanceSummary(mockDB)
	handler(c)

	// 驗證 HTTP status code
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want 200", w.Code)
	}

	// 驗證 JSON 回傳內容
	var got []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	expected := []map[string]interface{}{
		{
			"date":         "2023-01-02",
			"employeeID":   "E1",
			"name":         "Alice Wang",
			"ClockInTime":  "08:30",
			"ClockOutTime": "17:30",
			"ClockInGate":  "North",
			"ClockOutGate": "South",
			"status":       "On Time",
		},
	}

	gotJSON, _ := json.MarshalIndent(got, "", "  ")
	wantJSON, _ := json.MarshalIndent(expected, "", "  ")
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("response mismatch:\nGot:\n%s\nWant:\n%s", gotJSON, wantJSON)
	}
}

func TestExportAttendanceSummaryCSV(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	db.DB = mockDB // 確保你的 repository 使用這個 mockDB

	// 🧪 模擬一名員工
	mock.ExpectQuery(`SELECT \* FROM "employee" WHERE organization_id = \$1`).
		WithArgs("Engineering").
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name", "organization_id"}).
			AddRow("E1", "Alice", "Wang", "Engineering"))

	// 🧪 模擬打卡資料
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3 ORDER BY access_time asc`).
		WithArgs("E1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction", "gate_name"}).
			AddRow(time.Date(2023, 1, 2, 8, 30, 0, 0, time.UTC), "IN", "North").
			AddRow(time.Date(2023, 1, 2, 17, 30, 0, 0, time.UTC), "OUT", "South"))

	// 建立測試 context（模擬 query string）
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/?department=Engineering&fromDate=2023-01-01&toDate=2023-01-02", nil)
	c.Request = req

	// 執行 handler
	handler := ExportAttendanceSummaryCSV(mockDB)
	handler(c)

	// 驗證 HTTP status code
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want 200", w.Code)
	}

	// 驗證 headers
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/csv") {
		t.Errorf("unexpected content-type: %s", ct)
	}
	if disp := w.Header().Get("Content-Disposition"); !strings.Contains(disp, "attachment") {
		t.Errorf("expected attachment header, got: %s", disp)
	}

	// 解析 CSV
	r := csv.NewReader(bytes.NewReader(w.Body.Bytes()))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse csv: %v", err)
	}

	expected := [][]string{
		{"date", "employee ID", "name", "clock-in time", "clock-in gate", "clock-out time", "clock-out gate", "status"},
		{"2023-01-02", "E1", "Alice Wang", "08:30", "North", "17:30", "South", "On Time"},
	}

	if !reflect.DeepEqual(records, expected) {
		gotJSON, _ := json.MarshalIndent(records, "", "  ")
		wantJSON, _ := json.MarshalIndent(expected, "", "  ")
		t.Errorf("CSV mismatch:\nGot:\n%s\nWant:\n%s", gotJSON, wantJSON)
	}
}

func TestExportAttendanceSummaryPDF(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	db.DB = mockDB // 確保 repository 用到的 db 是 mockDB

	// 🧪 模擬一名員工
	mock.ExpectQuery(`SELECT \* FROM "employee" WHERE organization_id = \$1`).
		WithArgs("Engineering").
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name", "organization_id"}).
			AddRow("E1", "Alice", "Wang", "Engineering"))

	// 🧪 模擬該員工的 access log
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3 ORDER BY access_time asc`).
		WithArgs("E1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction", "gate_name"}).
			AddRow(time.Date(2023, 1, 2, 8, 30, 0, 0, time.UTC), "IN", "North").
			AddRow(time.Date(2023, 1, 2, 17, 30, 0, 0, time.UTC), "OUT", "South"))

	// 建立 HTTP 測試 context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/?department=Engineering&fromDate=2023-01-01&toDate=2023-01-02", nil)
	c.Request = req

	// 呼叫 handler
	handler := ExportAttendanceSummaryPDF(mockDB)
	handler(c)

	// ✅ 驗證 HTTP status
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want 200", w.Code)
	}

	// ✅ 驗證 headers
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/pdf") {
		t.Errorf("unexpected content-type: %s", ct)
	}
	if disp := w.Header().Get("Content-Disposition"); !strings.Contains(disp, "summary.pdf") {
		t.Errorf("expected content-disposition to contain summary.pdf, got: %s", disp)
	}

	// ✅ 驗證回傳 PDF 大小
	if len(w.Body.Bytes()) < 1000 {
		t.Errorf("expected PDF content, got small output (%d bytes)", len(w.Body.Bytes()))
	}

	// ✅ 可選：匯出 PDF 檔案檢查手動開啟
	// os.WriteFile("test_summary.pdf", w.Body.Bytes(), 0644)
}

func TestGetThisMonthTeam(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	db.DB = mockDB

	// 假設今天是 2023-02（測試 2023-02 與 2023-01）
	loc := time.UTC
	// currentStart := time.Date(2023, 2, 1, 0, 0, 0, 0, loc)
	// currentEnd := time.Date(2023, 3, 1, 0, 0, 0, 0, loc)
	// prevStart := time.Date(2023, 1, 1, 0, 0, 0, 0, loc)
	// prevEnd := time.Date(2023, 2, 1, 0, 0, 0, 0, loc)

	// 模擬部門員工
	mock.ExpectQuery(`SELECT \* FROM "employee"`).
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name", "organization_id"}).
			AddRow("E1", "Alice", "Wang", "Engineering"))

	// 模擬當月打卡資料
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3`).
		WithArgs("E1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction"}).
			AddRow(time.Date(2023, 2, 15, 8, 0, 0, 0, loc), "IN").
			AddRow(time.Date(2023, 2, 15, 18, 0, 0, 0, loc), "OUT"))

	mock.ExpectQuery(`SELECT \* FROM "employee"`).
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name", "organization_id"}).
			AddRow("E1", "Alice", "Wang", "Engineering"))

	// 模擬前月打卡資料
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3`).
		WithArgs("E1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction"}).
			AddRow(time.Date(2023, 1, 15, 9, 0, 0, 0, loc), "IN").
			AddRow(time.Date(2023, 1, 15, 19, 0, 0, 0, loc), "OUT"))

	// 建立測試 context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/team-report?month=2023-02", nil)
	c.Request = req
	c.Params = []gin.Param{{Key: "department", Value: "Engineering"}}

	// 執行 handler
	handler := GetThisMonthTeam(mockDB)
	handler(c)

	// 驗證 HTTP status
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want 200", w.Code)
	}

	// 驗證回傳格式
	var got []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 report maps (current & prev), got %d", len(got))
	}

	// 基本格式驗證
	for i, report := range got {
		if _, ok := report["TotalWorkHours"]; !ok {
			t.Errorf("report[%d] missing TotalWorkHours", i)
		}
		if _, ok := report["TotalOTHours"]; !ok {
			t.Errorf("report[%d] missing TotalOTHours", i)
		}
		if _, ok := report["OTHoursPerson"]; !ok {
			t.Errorf("report[%d] missing OTHoursPerson", i)
		}
		if _, ok := report["OTHeadcounts"]; !ok {
			t.Errorf("report[%d] missing OTHeadcounts", i)
		}
	}

	// ✅ 可選：輸出 JSON 幫助檢查
	// out, _ := json.MarshalIndent(got, "", "  ")
	// fmt.Println(string(out))
}

func TestGetThisWeekTeam(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	db.DB = mockDB

	// 🧪 員工清單
	mock.ExpectQuery(`SELECT \* FROM "employee"`).
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name", "organization_id"}).
			AddRow("E1", "Alice", "Wang", "Engineering"))

	// 🧪 本週打卡資料
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3`).
		WithArgs("E1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction"}).
			AddRow(time.Date(2023, 3, 6, 9, 0, 0, 0, time.UTC), "IN").
			AddRow(time.Date(2023, 3, 6, 19, 0, 0, 0, time.UTC), "OUT"))

	// 🧪 上週打卡資料
	mock.ExpectQuery(`SELECT \* FROM "employee"`). // 第二次撈員工
							WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name", "organization_id"}).
								AddRow("E1", "Alice", "Wang", "Engineering"))

	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3`).
		WithArgs("E1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction"}).
			AddRow(time.Date(2023, 2, 27, 8, 0, 0, 0, time.UTC), "IN").
			AddRow(time.Date(2023, 2, 27, 16, 0, 0, 0, time.UTC), "OUT"))

	// 🧪 建立 HTTP 測試 context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/team-report", nil)
	c.Request = req
	c.Params = []gin.Param{{Key: "department", Value: "Engineering"}}

	// 🧪 執行 handler
	handler := GetThisWeekTeam(mockDB)

	// ✅ 代理 time.Now()（可選）：用 monkey patch 或抽象時間來源
	// 此範例假設你直接使用 now（mocked in local scope）

	handler(c)

	// ✅ 驗證回應
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d, want 200", w.Code)
	}

	var got []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 week reports, got %d", len(got))
	}

	for i, report := range got {
		if _, ok := report["TotalWorkHours"]; !ok {
			t.Errorf("report[%d] missing TotalWorkHours", i)
		}
		if _, ok := report["TotalOTHours"]; !ok {
			t.Errorf("report[%d] missing TotalOTHours", i)
		}
		if _, ok := report["OTHoursPerson"]; !ok {
			t.Errorf("report[%d] missing OTHoursPerson", i)
		}
		if _, ok := report["OTHeadcounts"]; !ok {
			t.Errorf("report[%d] missing OTHeadcounts", i)
		}
	}
}

func TestGetCustomPeriodTeam(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	db.DB = mockDB

	// 模擬部門資訊與期間
	department := "Engineering"
	startDate := "2023-03-01"
	endDate := "2023-03-03"

	// 🧪 模擬一名員工
	mock.ExpectQuery(`SELECT \* FROM "employee"`).
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name", "organization_id"}).
			AddRow("E1", "Alice", "Wang", department))

	// 🧪 模擬打卡資料
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3`).
		WithArgs("E1", sqlmock.AnyArg(), sqlmock.AnyArg()). // 跳過時區比對問題
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction"}).
			AddRow(time.Date(2023, 3, 1, 9, 0, 0, 0, time.UTC), "IN").
			AddRow(time.Date(2023, 3, 1, 18, 0, 0, 0, time.UTC), "OUT"))

	// 建立測試 context（使用 path param 模擬路由）
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/team-report", nil)
	c.Request = req
	c.Params = []gin.Param{
		{Key: "department", Value: department},
		{Key: "startDate", Value: startDate},
		{Key: "endDate", Value: endDate},
	}

	// 呼叫 handler
	handler := GetCustomPeriodTeam(mockDB)
	handler(c)

	// 驗證 HTTP status
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want 200", w.Code)
	}

	// 驗證回傳 JSON 結構
	var got map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	expectedKeys := []string{"TotalWorkHours", "TotalOTHours", "OTHoursPerson", "OTHeadcounts"}
	for _, key := range expectedKeys {
		if _, ok := got[key]; !ok {
			t.Errorf("expected key %s missing in response", key)
		}
	}
}
