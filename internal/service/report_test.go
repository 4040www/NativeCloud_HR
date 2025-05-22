package service

import (
	"errors"
	"reflect"
	"testing"

	"github.com/4040www/NativeCloud_HR/internal/model"
	"github.com/DATA-DOG/go-sqlmock"
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

	userID := "emp001"
	expected := &model.AttendanceSummary{
		Date:         "2025-05-22",
		Name:         "John Doe",
		ClockInTime:  "09:00",
		ClockOutTime: "18:00",
		ClockInGate:  "Front Gate",
		ClockOutGate: "Back Gate",
		Status:       "Present",
	}

	t.Run("valid summary", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM "attendance" WHERE employee_id = \$1 AND date = CURRENT_DATE.*`).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{
				"date", "name", "clock_in_time", "clock_out_time", "clock_in_gate", "clock_out_gate", "status",
			}).AddRow(
				expected.Date,
				expected.Name,
				expected.ClockInTime,
				expected.ClockOutTime,
				expected.ClockInGate,
				expected.ClockOutGate,
				expected.Status,
			))

		got, err := GetTodayAttendanceSummary(db, userID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("got = %v, want = %v", got, expected)
		}
	})

	t.Run("no record found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM "attendance" WHERE employee_id = \$1 AND date = CURRENT_DATE.*`).
			WithArgs("emp999").
			WillReturnRows(sqlmock.NewRows([]string{
				"date", "name", "clock_in_time", "clock_out_time", "clock_in_gate", "clock_out_gate", "status",
			})) // 無資料

		got, err := GetTodayAttendanceSummary(db, "emp999")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM "attendance" WHERE employee_id = \$1 AND date = CURRENT_DATE.*`).
			WithArgs("emp500").
			WillReturnError(errors.New("db error"))

		got, err := GetTodayAttendanceSummary(db, "emp500")
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
