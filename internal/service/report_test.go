package service

import (
	"reflect"
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

func TestGetAttendanceWithEmployee(t *testing.T) {
	db, mock := setupMockDB(t)

	userID := "user-123"
	loc, _ := time.LoadLocation("Asia/Taipei")
	start := time.Date(2025, 5, 20, 0, 0, 0, 0, loc)
	end := time.Date(2025, 5, 25, 23, 59, 59, 0, loc)

	// 模擬 access_log 查詢 (7欄，符合 AccessLog 結構)
	accessLogs := sqlmock.NewRows([]string{
		"access_id", "employee_id", "access_time", "direction", "gate_type", "gate_name", "access_result",
	}).AddRow(
		"access-001", userID, start.Add(9*time.Hour), "IN", "TypeA", "Gate1", "Success",
	).AddRow(
		"access-002", userID, start.Add(18*time.Hour), "OUT", "TypeB", "Gate2", "Success",
	).AddRow(
		"access-003", userID, start.AddDate(0, 0, 1).Add(9*time.Hour), "IN", "TypeA", "Gate1", "Success",
	).AddRow(
		"access-004", userID, start.AddDate(0, 0, 1).Add(18*time.Hour), "OUT", "TypeB", "Gate2", "Success",
	)

	// 模擬 employees 查詢
	employeeRows := sqlmock.NewRows([]string{"employee_id", "first_name", "last_name"}).
		AddRow(userID, "Test", "User")

	// 預期 access_log 查詢，注意 end 加一天
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3 ORDER BY access_time asc`).
		WithArgs(userID, start, end.Add(24*time.Hour)).
		WillReturnRows(accessLogs)

	// 預期 employees 查詢
	mock.ExpectQuery(`SELECT.*FROM.*employee.*WHERE.*employee_id.*\$1.*`).
		WithArgs(userID, 1). // 加上 LIMIT 1 的參數
		WillReturnRows(employeeRows)

	got, err := GetAttendanceWithEmployee(db, userID, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []model.AttendanceSummary{
		{
			Date:         start.Format("2006-01-02"),
			Name:         "Test User",
			ClockInTime:  "09:00",
			ClockOutTime: "18:00",
			ClockInGate:  "Gate1",
			ClockOutGate: "Gate2",
			Status:       "Late", // 9點入場超過8:30晚班準時判斷邏輯
		},
		{
			Date:         start.AddDate(0, 0, 1).Format("2006-01-02"),
			Name:         "Test User",
			ClockInTime:  "09:00",
			ClockOutTime: "18:00",
			ClockInGate:  "Gate1",
			ClockOutGate: "Gate2",
			Status:       "Late",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetAttendanceWithEmployee() = %+v, want %+v", got, want)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestFetchMonthlyTeamReport(t *testing.T) {
	db, mock := setupMockDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	mock.ExpectQuery(`SELECT \* FROM "employee"`).
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name", "organization_id"}).
			AddRow("E001", "John", "Doe", "D001"))

	// 假設這位員工在該月內有兩筆打卡記錄
	start := time.Date(2024, 5, 1, 0, 0, 0, 0, time.Local)
	// end := start.AddDate(0, 1, 0)

	accessRows := sqlmock.NewRows([]string{"employee_id", "direction", "access_time"}).
		AddRow("E001", "IN", start.Add(9*time.Hour)).                 // 09:00
		AddRow("E001", "OUT", start.Add(18*time.Hour+30*time.Minute)) // 18:30

	mock.ExpectQuery(`SELECT \* FROM "access_log"`).
		WithArgs("E001", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(accessRows)

	// 測試目標
	tests := []struct {
		name string
		args struct {
			db           *gorm.DB
			departmentID string
			month        string
		}
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "Valid report for department D001",
			args: struct {
				db           *gorm.DB
				departmentID string
				month        string
			}{
				db:           db,
				departmentID: "D001",
				month:        "2024-05",
			},
			want: map[string]interface{}{
				"TotalWorkHours": 9.5,
				"TotalOTHours":   1.5,
				"OTHoursPerson":  1,
				"OTHeadcounts":   1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FetchMonthlyTeamReport(tt.args.db, tt.args.departmentID, tt.args.month)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchMonthlyTeamReport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchMonthlyTeamReport() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchCustomPeriodTeamReport(t *testing.T) {
	db, mock := setupMockDB(t) // 你要實作這個 helper

	startDate := "2025-05-01"
	endDate := "2025-05-10"
	deptID := "D001"

	// 1. mock employee 資料
	mock.ExpectQuery(`SELECT \* FROM "employee"`).
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "first_name", "last_name", "organization_id"}).
			AddRow("E001", "John", "Doe", "D001").
			AddRow("E002", "Jane", "Smith", "D002")) // D002 的員工不應被處理

	// 2. mock access_log for E001（只 mock D001 部門的員工）
	mock.ExpectQuery(`SELECT \* FROM "access_log" WHERE employee_id = \$1 AND access_time BETWEEN \$2 AND \$3 ORDER BY access_time asc`).
		WithArgs("E001", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"access_id", "employee_id", "direction", "access_time", "gate_type", "gate_name", "access_result",
		}).
			AddRow("1", "E001", "in", time.Date(2025, 5, 1, 9, 0, 0, 0, time.UTC), "", "", "").
			AddRow("2", "E001", "out", time.Date(2025, 5, 1, 19, 0, 0, 0, time.UTC), "", "", "")) // 10 小時，OT 2

	// ❌ 移除這段，因為 E002 是 D002，不會被查到 access_log，所以不需要 mock
	// mock.ExpectQuery(`SELECT \* FROM "access_log" ...`)

	// 呼叫被測 function
	result, err := FetchCustomPeriodTeamReport(db, deptID, startDate, endDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[string]interface{}{
		"TotalWorkHours": 10.0,
		"TotalOTHours":   2.0,
		"OTHoursPerson":  1,
		"OTHeadcounts":   1,
	}

	for k, v := range want {
		if result[k] != v {
			t.Errorf("expected %s = %v, got %v", k, v, result[k])
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}
