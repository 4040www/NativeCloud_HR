# API 說明文件

## Auth API 說明文件 `/auth`

提供使用者登入、登出功能，採用 JSON Web Token（JWT）作為身分驗證方式。

---

### 📌 API 一覽

| 方法 | 路徑       | 說明     |
|------|------------|----------|
| POST | `/login`  | 使用者登入，取得 JWT |
| POST | `/logout` | 使用者登出 |

---

### 🟢 POST `/auth/login`

用戶登入並取得 JWT token。

#### 🔸 Request

- Header: `Content-Type: application/json`
- Body:
```json
{
  "email": "user@example.com",
  "password": "yourpassword"
}
```

#### 🔸 Response

- **Status 200 OK**
```json
{
  "message": "Login successful",
  "token": "your.jwt.token",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe"
  }
}
```

- **Status 400 Bad Request**
```json
{
  "error": "Invalid request format"
}
```

- **Status 401 Unauthorized**
```json
{
  "error": "email or password incorrect"
}
```

---

### 🔴 POST `/auth/logout`

模擬登出動作（前端只需刪除 JWT）。如採用 token blacklist，可額外實作伺服器端記錄失效 token。

#### 🔸 Request

- Header: (通常包含 `Authorization: Bearer <token>`)

#### 🔸 Response

- **Status 200 OK**
```json
{
  "message": "Logout successful"
}
```

---

#### 📘 補充說明

- JWT 可存於前端的 `localStorage` 或 `cookie`
- 每次 API 請求應在 Header 加上：

```http
Authorization: Bearer <your.jwt.token>
```

## Clock API 說明文件 `/clock`

用於員工打卡（進出紀錄）。該 API 會從 JWT Token 自動提取使用者身份，產生唯一 access_id，並寫入資料庫中。

---


### 🟢 POST `/clock`

需登入，並在 Authorization header 中提供有效的 JWT Token。

#### 🔸 Request

- Header: `Content-Type: application/json`
- Body:
```json
{
  "access_time": "2025-04-26T09:00:00Z",
  "direction": "in",         // "in" 或 "out"
  "gate_type": "main_gate",  // 例如：main_gate / side_gate（看 db 組）
  "gate_name": "北門"         // 打卡的門口名稱（看 db 組）
}

```

#### 🔸 Response

- **Status 200 OK**
```json
{
  "message": "Check-in successful"
}
```

- **Status 400 Bad Request**
```json
{
  "error": "Invalid request format"
}
```

- **Status 401 Unauthorized**
```json
{
  "error": "User ID not found in token"
}
```


## Report API 說明文件 `/report`

提供出勤紀錄與統計報表功能。

---

### 📌 API 一覽

| 方法 | 路徑 | 說明 |
|------|------|------|
| GET | `/report/myRecords/:userID` | 查詢今日打卡紀錄 |
| GET | `/report/historyRecords/:userID` | 查詢最近 30 天出勤紀錄 |
| GET | `/report/historyRecords/:userID/:startDate/:endDate` | 查詢指定日期範圍出勤紀錄 |
| GET | `/report/thisMonth/:department/:userID` | 查詢本月與前月部門統計報表 |
| GET | `/report/thisWeek/:department/:userID` | 查詢本週與上週部門統計報表 |
| GET | `/report/PeriodTime/:department/:startDate/:endDate/:userID` | 查詢指定時間區間的部門統計 |
| GET | `/report/AlertList/:startDate/:endDate/:userID` | 查詢警示員工名單（加班/遲到） |
| GET | `/report/inChargeDepartment/:userID` | 查詢使用者管理的部門 |
| GET | `/report/summaryExportCSV/:department/:startDate/:endDate/:userID` | 匯出出勤紀錄 CSV |
| GET | `/report/summaryExportPDF/:department/:startDate/:endDate/:userID` | 匯出出勤紀錄 PDF |
| GET | `/report/myDepartments/:userID` | 查詢使用者可檢視的部門 |
| GET | `/report/attendanceSummary?department=...&fromDate=...&toDate=...` | 查詢出勤摘要資料 |
| GET | `/report/attendanceExportCSV?department=...&fromDate=...&toDate=...` | 匯出出勤摘要 CSV |
| GET | `/report/attendanceExportPDF?department=...&fromDate=...&toDate=...` | 匯出出勤摘要 PDF |

---

### 🟢 GET `/report/myRecords/:userID`

查詢使用者今日的出勤記錄。

#### 🔸 Response

```json
{
  "date": "2025-05-08",
  "name": "John Doe",
  "clock_in_time": "09:01",
  "clock_out_time": "18:05",
  "clock_in_gate": "北門",
  "clock_out_gate": "西門",
  "status": "Late"
}
```

---

### 🟢 GET `/report/historyRecords/:userID`

查詢使用者最近 30 天的出勤記錄。

#### 🔸 Response

```json
[
  {
    "date": "2025-05-02",
    "name": "John Doe",
    "clock_in_time": "09:00",
    "clock_out_time": "18:00",
    "clock_in_gate": "北門",
    "clock_out_gate": "北門",
    "status": "On Time"
  }
]
```

---

### 🟢 GET `/report/historyRecords/:userID/:startDate/:endDate`

查詢使用者在指定日期範圍內的出勤記錄。

#### 🔸 路徑參數

- `startDate`: 格式為 `YYYY-MM-DD`
- `endDate`: 格式為 `YYYY-MM-DD`

#### 🔸 Response

與 `/report/historyRecords/:userID` 相同格式。

---

### 🟢 GET `/report/thisMonth/:department/:userID`

查詢部門本月與上月的總工時、加班時數、參與人數等報表。

#### 🔸 Response

```json
[
  {
    "TotalWorkHours": 320,
    "TotalOTHours": 40,
    "OTHoursPerson": 5,
    "OTHeadcounts": 10
  },
  {
    "TotalWorkHours": 310,
    "TotalOTHours": 30,
    "OTHoursPerson": 3,
    "OTHeadcounts": 9
  }
]
```

---

### 🟢 GET `/report/thisWeek/:department/:userID`

查詢部門本週與上週的總體統計資料。

#### 🔸 Response

與 `/report/thisMonth/:department/:userID` 相同格式。

---

### 🟢 GET `/report/PeriodTime/:department/:startDate/:endDate/:userID`

查詢部門指定時間區間的總工時、加班、參與人數統計。

#### 🔸 Response

```json
{
  "TotalWorkHours": 100,
  "TotalOTHours": 12,
  "OTHoursPerson": 2,
  "OTHeadcounts": 3
}
```

---

### 🟢 GET `/report/AlertList/:startDate/:endDate/:userID`

回傳指定期間內，有遲到或加班超過標準的員工。

#### 🔸 Response

```json
[
  {
    "EmployeeID": "d3549701-c2a2-4857-b0d1-c3c7b71aed3d",
    "Name": "John Doe",
    "OTCounts": 4,
    "OTHours": 18,
    "status": "Warning"
  }
]
```

---

### 🟢 GET `/report/inChargeDepartment/:userID`

查詢該使用者所管理的部門（若為主管）。

#### 🔸 Response

```json
[
  "Sales",
  "Engineering",
  "HR"
]
```

---

### 🟢 GET `/report/summaryExportCSV/:department/:startDate/:endDate/:userID`

匯出指定部門與日期的出勤紀錄為 CSV 檔案。

#### 🔸 Response

- Header: `Content-Disposition: attachment; filename=summary.csv`
- Content-Type: `text/csv`
- Response Body 為 CSV 格式的原始資料

---

### 🟢 GET `/report/summaryExportPDF/:department/:startDate/:endDate/:userID`

匯出出勤摘要報表為 PDF 檔案。

#### 🔸 Response

- Header: `Content-Disposition: attachment; filename=summary.pdf`
- Content-Type: `application/pdf`

---

### 🟢 GET `/report/myDepartments/:userID`

取得使用者有權限查看的所有部門列表。

#### 🔸 Response

```json
[
  "Engineering 1",
  "HR",
  "Accounting"
]
```

---

### 🟢 GET `/report/attendanceSummary?department=...&fromDate=...&toDate=...`

查詢某部門特定區間的所有員工出勤紀錄。

#### 🔸 Response

```json
[
  {
    "date": "2025-05-02",
    "employeeID": "abc123",
    "name": "John Doe",
    "ClockInTime": "09:00",
    "ClockOutTime": "17:00",
    "ClockInGate": "北門",
    "ClockOutGate": "西門",
    "status": "On Time"
  }
]
```

---

### 🟢 GET `/report/attendanceExportCSV?department=...&fromDate=...&toDate=...`

匯出出勤摘要為 CSV 檔案。

---

### 🟢 GET `/report/attendanceExportPDF?department=...&fromDate=...&toDate=...`

匯出出勤摘要為 PDF 檔案。

---



## Notify API 說明文件 `/notify`

提供通知功能，協助偵測異常出勤狀況（如加班過多、遲到次數過多）並提醒相關人員。

---

### 📌 API 一覽

| 方法   | 路徑                 | 說明                         |
|--------|----------------------|------------------------------|
| GET    | `/warning`           | 查詢本月異常員工（加班/遲到）     |
| POST   | `/late/:employee_id` | 通知主管員工遲到次數過多         |
| POST   | `/overtime/:employee_id` | 通知 HR 員工加班時數過多     |

---

### 🔵 GET `/notify/warning`

查詢本月有異常情況的員工清單（遲到次數 ≥ 4 次、或加班總時數 > 46 小時）。

#### 🔸 Request

- 無需參數
- Header: (如需驗證可加上 JWT)

#### 🔸 Response

- **Status 200 OK**

```json
[
  {
    "employee_id": "123456",
    "problems": ["TooManyLate", "OvertimeExceeded"]
  },
  {
    "employee_id": "7891011",
    "problems": ["TooManyLate"]
  }
]
```

- **Status 500 Internal Server Error**

```json
{
  "error": "查詢異常員工失敗"
}
```

#### 🔸 前端範例

```js
fetch('/api/notify/warning')
  .then(res => res.json())
  .then(data => console.log(data));
```

---

### 🔵 POST `/notify/late/:employee_id`

通知主管該員工遲到次數過多（≥ 4 次）。

#### 🔸 Request

- Path Param: `employee_id`

#### 🔸 Response

- **Status 200 OK**

```json
{
  "message": "員工 123456 本月遲到超過 4 次，請主管關注。"
}
```

- **Status 404 Not Found**

```json
{
  "message": "員工 123456 遲到次數正常，無需提醒。"
}
```

#### 🔸 前端範例

```js
fetch('/api/notify/late/123456', {
  method: 'POST'
})
.then(res => res.json())
.then(data => console.log(data));
```

---

### 🔵 POST `/notify/overtime/:employee_id`

通知 HR 該員工加班總時數過多（> 46 小時）。

#### 🔸 Request

- Path Param: `employee_id`

#### 🔸 Response

- **Status 200 OK**

```json
{
  "message": "員工 123456 本月加班超過 46 小時，請 HR 檢查。"
}
```

- **Status 404 Not Found**

```json
{
  "message": "員工 123456 加班時數正常，無需提醒。"
}
```

#### 🔸 前端範例

```js
fetch('/api/notify/overtime/123456', {
  method: 'POST'
})
.then(res => res.json())
.then(data => console.log(data));
```

---
