package repository

import (
	"errors"
	"reflect"
	"testing"
	"time"

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

func TestGetAccessLogsByEmployeeBetween(t *testing.T) {
	db, mock := setupMockDB(t)

	employeeID := "emp-001"
	start := time.Date(2025, 5, 24, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 5, 25, 0, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"access_id", "employee_id", "access_time", "direction", "gate_type", "gate_name", "access_result",
	}).AddRow(
		"access-001", employeeID, time.Date(2025, 5, 24, 9, 0, 0, 0, time.UTC),
		"IN", "TypeA", "Gate1", "Success",
	).AddRow(
		"access-002", employeeID, time.Date(2025, 5, 24, 18, 0, 0, 0, time.UTC),
		"OUT", "TypeB", "Gate2", "Success",
	)

	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3 ORDER BY access_time asc`).
		WithArgs(employeeID, start, end).
		WillReturnRows(rows)

	got, err := GetAccessLogsByEmployeeBetween(db, employeeID, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []model.AccessLog{
		{
			AccessID:     "access-001",
			EmployeeID:   employeeID,
			AccessTime:   time.Date(2025, 5, 24, 9, 0, 0, 0, time.UTC),
			Direction:    "IN",
			GateType:     "TypeA",
			GateName:     "Gate1",
			AccessResult: "Success",
		},
		{
			AccessID:     "access-002",
			EmployeeID:   employeeID,
			AccessTime:   time.Date(2025, 5, 24, 18, 0, 0, 0, time.UTC),
			Direction:    "OUT",
			GateType:     "TypeB",
			GateName:     "Gate2",
			AccessResult: "Success",
		},
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d results, got %d", len(want), len(got))
	}

	for i := range got {
		if got[i].AccessID != want[i].AccessID ||
			got[i].EmployeeID != want[i].EmployeeID ||
			!got[i].AccessTime.Equal(want[i].AccessTime) ||
			got[i].Direction != want[i].Direction ||
			got[i].GateType != want[i].GateType ||
			got[i].GateName != want[i].GateName ||
			got[i].AccessResult != want[i].AccessResult {
			t.Errorf("mismatch at index %d: got %+v, want %+v", i, got[i], want[i])
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetAllEmployees(t *testing.T) {
	db, mock := setupMockDB(t)

	// 模擬資料庫查詢，符合 GetAllEmployees 會執行的 SQL
	mock.ExpectQuery(`SELECT \* FROM "employee"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"employee_id", "first_name", "last_name", "is_manager", "password", "email", "organization_id",
		}).AddRow(
			"E001", "John", "Doe", false, "hashedpwd", "john@example.com", "D001",
		).AddRow(
			"E002", "Jane", "Smith", true, "hashedpwd2", "jane@example.com", "D002",
		))

	want := []model.Employee{
		{
			EmployeeID:     "E001",
			FirstName:      "John",
			LastName:       "Doe",
			IsManager:      false,
			Password:       "hashedpwd",
			Email:          "john@example.com",
			OrganizationID: "D001",
		},
		{
			EmployeeID:     "E002",
			FirstName:      "Jane",
			LastName:       "Smith",
			IsManager:      true,
			Password:       "hashedpwd2",
			Email:          "jane@example.com",
			OrganizationID: "D002",
		},
	}

	type args struct {
		db *gorm.DB
	}
	tests := []struct {
		name    string
		args    args
		want    []model.Employee
		wantErr bool
	}{
		{
			name:    "return all employees successfully",
			args:    args{db: db},
			want:    want,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAllEmployees(tt.args.db)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllEmployees() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllEmployees() = %v, want %v", got, tt.want)
			}
		})
	}

	// 確認所有預期操作都被使用
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
