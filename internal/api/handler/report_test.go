// report_handlers.go
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

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
