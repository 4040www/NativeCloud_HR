package service

import (
	"testing"
	"time"

	"github.com/4040www/NativeCloud_HR/internal/model"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() failed: %v", err)
	}
	dialector := postgres.New(postgres.Config{
		Conn: mockDB,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() failed: %v", err)
	}
	return db, mock
}

func TestGetTodayAttendanceSummary(t *testing.T) {
	db, mock := setupMockDB(t)

	loc, _ := time.LoadLocation("Asia/Taipei")
	today := time.Date(2025, 5, 23, 0, 0, 0, 0, loc)

	userID := "user-123"

	// 模擬 access_log 查詢
	accessLogs := sqlmock.NewRows([]string{"employee_id", "direction", "access_time", "gate_name"}).
		AddRow(userID, "IN", today.Add(9*time.Hour), "A1").
		AddRow(userID, "OUT", today.Add(18*time.Hour), "B2")

	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3 ORDER BY access_time asc`).
		WithArgs(userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(accessLogs)

	// 模擬 employees 查詢
	employeeRows := sqlmock.NewRows([]string{"id", "first_name", "last_name"}).
		AddRow(userID, "Test", "User")

	mock.ExpectQuery(`SELECT.*FROM.*employee.*WHERE.*employee_id.*\$1.*`).
		WithArgs(userID, 1).
		WillReturnRows(employeeRows)

	summary, err := GetTodayAttendanceSummary(db, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := &model.AttendanceSummary{
		Date:         "2025-05-23",
		Name:         "Test User",
		ClockInTime:  "09:00",
		ClockOutTime: "18:00",
		ClockInGate:  "A1",
		ClockOutGate: "B2",
		Status:       "Late",
	}

	if diff := cmp.Diff(want, summary); diff != "" {
		t.Errorf("GetTodayAttendanceSummary() mismatch (-want +got):\n%s", diff)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
