package repository

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

func TestGetEmployeeByID(t *testing.T) {
	db, mock := setupMockDB(t)

	userID := "emp001"
	expected := &model.Employee{
		EmployeeID: userID,
		FirstName:  "John",
		LastName:   "Doe",
	}

	t.Run("valid employee", func(t *testing.T) {
		// 正常查詢
		mock.ExpectQuery(`SELECT .* FROM "employee" WHERE employee_id = \$1.*`).
			WithArgs(userID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name"}).
				AddRow(expected.EmployeeID, expected.FirstName, expected.LastName))

		got, err := GetEmployeeByID(db, userID)
		if err != nil {
			t.Errorf("GetEmployeeByID() unexpected error: %v", err)
		}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("GetEmployeeByID() = %v, want %v", got, expected)
		}
	})

	t.Run("employee not found", func(t *testing.T) {
		nonExistentID := "emp999"
		// 查無此人
		mock.ExpectQuery(`SELECT .* FROM "employee" WHERE employee_id = \$1.*`).
			WithArgs("emp999", 1).
			WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name"})) // 空行

		got, err := GetEmployeeByID(db, nonExistentID)
		if err != nil {
			t.Errorf("GetEmployeeByID() unexpected error: %v", err)
		}
		if got != nil {
			t.Errorf("GetEmployeeByID() = %v, want nil", got)
		}
	})

	t.Run("database error", func(t *testing.T) {
		errorID := "emp500"
		// 資料庫錯誤
		mock.ExpectQuery(`SELECT .* FROM "employee" WHERE employee_id = \$1.*`).
			WithArgs("emp500", 1).
			WillReturnError(errors.New("db connection failed"))

		got, err := GetEmployeeByID(db, errorID)
		if err == nil {
			t.Errorf("GetEmployeeByID() expected error, got nil")
		}
		if got != nil {
			t.Errorf("GetEmployeeByID() = %v, want nil", got)
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
