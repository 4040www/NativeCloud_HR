package repository

import (
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

	// 模擬 GORM 查詢語法
	mock.ExpectQuery(`SELECT .* FROM "employee" WHERE employee_id = \$1.*`).
		WithArgs(userID, 1). // GORM 的 LIMIT 1 是參數
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name"}).
			AddRow(expected.EmployeeID, expected.FirstName, expected.LastName))

	tests := []struct {
		name string
		args struct {
			db *gorm.DB
			id string
		}
		want    *model.Employee
		wantErr bool
	}{
		{
			name: "valid employee",
			args: struct {
				db *gorm.DB
				id string
			}{
				db: db,
				id: userID,
			},
			want:    expected,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEmployeeByID(tt.args.db, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEmployeeByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEmployeeByID() = %v, want %v", got, tt.want)
			}
		})
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
