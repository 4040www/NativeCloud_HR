package service

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/4040www/NativeCloud_HR/internal/db" // ✅ 指的是你 package db 的變數 DB
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

func TestGenerateAlertList(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	db.DB = mockDB

	// 員工資料
	mock.ExpectQuery(`SELECT \* FROM "employee"`).
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name"}).
			AddRow("E1", "Alice", "Normal").
			AddRow("E2", "Bob", "Warning").
			AddRow("E3", "Charlie", "Alert"))

	// E1: Normal（9 小時）
	mock.ExpectQuery(`SELECT \* FROM "?access_log"? WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3`).
		WithArgs("E1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction"}).
			AddRow(time.Date(2023, 1, 1, 8, 30, 0, 0, time.UTC), "IN").
			AddRow(time.Date(2023, 1, 1, 17, 30, 0, 0, time.UTC), "OUT"))

	// E2: Warning（兩天，每天 10.5 小時）
	mock.ExpectQuery(`SELECT \* FROM "?access_log"? WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3`).
		WithArgs("E2", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction"}).
			// Day 1
			AddRow(time.Date(2023, 1, 1, 8, 30, 0, 0, time.UTC), "IN").
			AddRow(time.Date(2023, 1, 1, 19, 30, 0, 0, time.UTC), "OUT").
			// Day 2
			AddRow(time.Date(2023, 1, 2, 8, 30, 0, 0, time.UTC), "IN").
			AddRow(time.Date(2023, 1, 2, 19, 30, 0, 0, time.UTC), "OUT"))

	// E3: Alert（一天 13 小時）
	mock.ExpectQuery(`SELECT \* FROM "?access_log"? WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3`).
		WithArgs("E3", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction"}).
			AddRow(time.Date(2023, 1, 1, 8, 30, 0, 0, time.UTC), "IN").
			AddRow(time.Date(2023, 1, 1, 21, 30, 0, 0, time.UTC), "OUT"))

	got, err := GenerateAlertList(mockDB, "2023-01-01", "2023-01-02")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []map[string]interface{}{
		{
			"EmployeeID": "E1",
			"Name":       "Alice Normal",
			"OTCounts":   1,
			"OTHours":    1,
			"status":     "Normal",
		},
		{
			"EmployeeID": "E2",
			"Name":       "Bob Warning",
			"OTCounts":   2,
			"OTHours":    6,
			"status":     "Warning",
		},
		{
			"EmployeeID": "E3",
			"Name":       "Charlie Alert",
			"OTCounts":   1,
			"OTHours":    5,
			"status":     "Alert",
		},
	}
	gotJSON, _ := json.MarshalIndent(got, "", "  ")
	wantJSON, _ := json.MarshalIndent(want, "", "  ")
	if !bytes.Equal(gotJSON, wantJSON) {
		t.Errorf("GenerateAlertList() mismatch:\nGot:\n%s\nWant:\n%s", gotJSON, wantJSON)
	}

}

func TestGetManagedDepartments(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	db.DB = mockDB // 用你實際 db.DB 全域變數

	// 測試 case 1：無部門（查無資料）
	mock.ExpectQuery(`SELECT distinct organization_id FROM "?employee"? WHERE employee_id = \$1`).
		WithArgs("userA").
		WillReturnRows(sqlmock.NewRows([]string{"organization_id"})) // 無資料

	mock.ExpectQuery(`SELECT distinct organization_id FROM "?employee"? WHERE employee_id = \$1`).
		WithArgs("userB").
		WillReturnRows(sqlmock.NewRows([]string{"organization_id"}).
			AddRow("D001").
			AddRow("D002"))

	tests := []struct {
		name   string
		userID string
		want   []string
	}{
		{
			name:   "無管理部門",
			userID: "userA",
			want:   []string{},
		},
		{
			name:   "有管理部門",
			userID: "userB",
			want:   []string{"D001", "D002"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetManagedDepartments(tt.userID)
			gotJSON, _ := json.MarshalIndent(got, "", "  ")
			wantJSON, _ := json.MarshalIndent(tt.want, "", "  ")
			if !bytes.Equal(gotJSON, wantJSON) {
				t.Errorf("GenerateAlertList() mismatch:\nGot:\n%s\nWant:\n%s", gotJSON, wantJSON)
			}
		})
	}
}

func TestGenerateAttendanceSummaryCSV(t *testing.T) {
	mockDB, mock := setupMockDB(t)

	// 用 mockDB 覆蓋全域 db.DB，這樣 GetEmployeesByDepartment 裡會用到你的 mockDB
	db.DB = mockDB

	// Mock 員工查詢
	mock.ExpectQuery(`SELECT \* FROM "employee" WHERE organization_id = \$1`).
		WithArgs("Engineering").
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name", "organization_id"}).
			AddRow("E1", "Alice", "Wang", "Engineering").
			AddRow("E2", "Bob", "Chen", "Engineering"))

	// Mock access log 查詢（依你 GetAttendanceSummaryForDepartments 裡實作加）
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3`).
		WithArgs("E1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction", "gate_name"}).
			AddRow(time.Date(2023, 1, 2, 8, 30, 0, 0, time.UTC), "IN", "North").
			AddRow(time.Date(2023, 1, 2, 17, 30, 0, 0, time.UTC), "OUT", "South"))

	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3`).
		WithArgs("E2", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction", "gate_name"}).
			AddRow(time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC), "IN", "East").
			AddRow(time.Date(2023, 1, 1, 18, 0, 0, 0, time.UTC), "OUT", "West"))

	got, err := GenerateAttendanceSummaryCSV(mockDB, "Engineering", "2023-01-01", "2023-01-02")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := csv.NewReader(bytes.NewReader(got))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse csv: %v", err)
	}

	want := [][]string{
		{"date", "employee ID", "name", "clock-in time", "clock-in gate", "clock-out time", "clock-out gate", "status"},
		{"2023-01-02", "E1", "Alice Wang", "08:30", "North", "17:30", "South", "On Time"},
		{"2023-01-01", "E2", "Bob Chen", "09:00", "East", "18:00", "West", "Late"},
	}

	if !reflect.DeepEqual(records, want) {
		t.Errorf("CSV output mismatch:\nGot: %#v\nWant: %#v", records, want)
	}
}

func TestGetAttendanceSummaryForDepartments(t *testing.T) {
	mockDB, mock := setupMockDB(t)
	db.DB = mockDB // 如果 GetEmployeesByDepartment 沒參數 db 則需要這行

	// 模擬員工資料
	mock.ExpectQuery(`SELECT \* FROM "employee" WHERE organization_id = \$1`).
		WithArgs("Engineering").
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name", "organization_id"}).
			AddRow("E1", "Alice", "Wang", "Engineering").
			AddRow("E2", "Bob", "Chen", "Engineering"))

	// 模擬 E1 access_log
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3 ORDER BY access_time asc`).
		WithArgs("E1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction", "gate_name"}).
			AddRow(time.Date(2023, 1, 2, 8, 30, 0, 0, time.UTC), "IN", "North").
			AddRow(time.Date(2023, 1, 2, 17, 30, 0, 0, time.UTC), "OUT", "South"))

	// 模擬 E2 access_log
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3 ORDER BY access_time asc`).
		WithArgs("E2", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction", "gate_name"}).
			AddRow(time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC), "IN", "East").
			AddRow(time.Date(2023, 1, 1, 18, 0, 0, 0, time.UTC), "OUT", "West"))

	got, err := GetAttendanceSummaryForDepartments(mockDB, "Engineering", "2023-01-01", "2023-01-02")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []map[string]interface{}{
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
		{
			"date":         "2023-01-01",
			"employeeID":   "E2",
			"name":         "Bob Chen",
			"ClockInTime":  "09:00",
			"ClockOutTime": "18:00",
			"ClockInGate":  "East",
			"ClockOutGate": "West",
			"status":       "Late",
		},
	}

	// 轉換成 JSON 來比對方便閱讀
	gotJSON, _ := json.MarshalIndent(got, "", "  ")
	wantJSON, _ := json.MarshalIndent(want, "", "  ")

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Mismatch in attendance summary\nGot:\n%s\nWant:\n%s", gotJSON, wantJSON)
	}
}

func TestGenerateAttendanceSummaryPDF(t *testing.T) {
	type args struct {
		db    *gorm.DB
		dept  string
		start string
		end   string
	}

	mockDB, mock := setupMockDB(t)
	db.DB = mockDB // 若你的 GetEmployeesByDepartment 依賴全域變數

	// 🧪 Mock 員工資料
	mock.ExpectQuery(`SELECT \* FROM "employee" WHERE organization_id = \$1`).
		WithArgs("Engineering").
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name", "organization_id"}).
			AddRow("E1", "Alice", "Wang", "Engineering").
			AddRow("E2", "Bob", "Chen", "Engineering"))

	// 🧪 Mock access_log for E1
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3 ORDER BY access_time asc`).
		WithArgs("E1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction", "gate_name"}).
			AddRow(time.Date(2023, 1, 2, 8, 30, 0, 0, time.UTC), "IN", "North").
			AddRow(time.Date(2023, 1, 2, 17, 30, 0, 0, time.UTC), "OUT", "South"))

	// 🧪 Mock access_log for E2
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3 ORDER BY access_time asc`).
		WithArgs("E2", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"access_time", "direction", "gate_name"}).
			AddRow(time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC), "IN", "East").
			AddRow(time.Date(2023, 1, 1, 18, 0, 0, 0, time.UTC), "OUT", "West"))

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "PDF generation with 2 employees",
			args: args{
				db:    mockDB,
				dept:  "Engineering",
				start: "2023-01-01",
				end:   "2023-01-02",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateAttendanceSummaryPDF(tt.args.db, tt.args.dept, tt.args.start, tt.args.end)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateAttendanceSummaryPDF() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// ✅ 不做 reflect.DeepEqual，比較 PDF binary 通常沒意義
			if len(got) < 1000 {
				t.Errorf("Generated PDF too small or empty: got %d bytes", len(got))
			}
			// 🧪 可選：輸出成檔案查看
			// os.WriteFile("test_attendance.pdf", got, 0644)
		})
	}
}
