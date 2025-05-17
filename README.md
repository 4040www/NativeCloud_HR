# NativeCloud_HR 系統架構與運作流程

本專案是針對公司員工考勤與門禁系統設計的後端架構，主要包含以下幾個模組與層次，旨在實現高效、可擴展、並且高可用的系統架構。

## 目錄

1. [後端系統架構說明](#後端系統架構說明)
2. [API 層](#api-層)
3. [Message Queue 層 (Kafka / NATS)](#message-queue-層-kafka--nats)
4. [資料庫相關 (PostgreSQL / Redis)](#資料庫相關-postgresql--redis)
5. [系統架構](#系統架構)
6. [異常處理與報表](#異常處理與報表)

## 1. 後端系統架構說明
```
📦 employee-access-system
│── 📂 cmd                  # 入口點
│   ├── 📝 main.go          # 主要執行程式
│── 📂 config               # 設定檔
│   ├── 📝 config.go        # 設定加載邏輯
│   ├── 📝 config.yaml      # 設定檔案 (GCP, DB, Kafka 連線)
│   ├── 📝 .env             # 環境變數
│── 📂 internal             # 核心業務邏輯
│   ├── 📂 api              # API 層
│   │   ├── 📂 handlers.go  # HTTP 請求處理
│   │   │   ├── 📝 auth.go      # 身分驗證
│   │   │   ├── 📝 clock.go     # 打卡
│   │   │   ├── 📝 notify.go    # 提醒通知
│   │   │   ├── 📝 report.go    # 報表生成
│   │   ├── 📂 routes.go  # HTTP 請求處理
│   │   │   ├── 📝 auth_routes.go      # 身分驗證路徑
│   │   │   ├── 📝 clock_routes.go     # 打卡路徑
│   │   │   ├── 📝 notify_routes.go    # 提醒通知路徑
│   │   │   ├── 📝 report_routes.go    # 報表生成路徑
│   │   ├── 📝 middleware.go# 身分驗證中介軟體
│   │   ├── 📝 router.go    # 設定 API 路由 (Echo/Gin)
│   ├── 📂 db               # 資料庫初始化
│   │   ├── 📝 postgres.go  # PostgreSQL 連線
│   ├── 📂 messageQueue     # 負責消息隊列操作
│   │   ├── 📝 kafka.go     # Kafka 配置與生產者
│   │   ├── 📝 producer.go  # 發送消息的生產者
│   │   ├── 📝 consumer.go  # 接收消息的消費者
│   ├── 📂 repository       # 資料存取層
│   │   │   ├── 📝 user_repo.go      # 身分驗證數據存取
│   │   │   ├── 📝 clock_repo.go     # 打卡數據存取
│   │   │   ├── 📝 notify_repo.go    # 提醒通知數據存取
│   │   │   ├── 📝 report_repo.go    # 報表生成數據存取
│   ├── 📂 service          # 服務層（業務邏輯）
│   │   │   ├── 📝 auth.go      # 身分驗證服務
│   │   │   ├── 📝 clock.go     # 打卡服務
│   │   │   ├── 📝 notify.go    # 提醒通知服務
│   │   │   ├── 📝 report.go    # 報表生成服務
│   ├── 📂 utils            # 通用工具
│   │   ├── 📝 jwt.go       # token 管理
├── 📂 deployments          # 部署相關
│── 📂 scripts              # 運維腳本（未完成）
│   ├── 📝 migrate.sh       # 資料庫遷移
│   ├── 📝 start.sh         # 啟動指令
│── 📂 docs                 # 文件
│   ├── 📝 API.md           # API 說明文件
│   ├── 📝 architecture.md  # 系統架構說明
│   ├── 📝 GCP_deployment.md# 上雲說明
├── Dockerfile       # 容器化設定
├── docker-compose.yml # 本地測試環境
│── go.mod                  # Golang 依賴管理
│── go.sum                  # Golang 依賴鎖定
│── README.md               # 專案說明文件
```

---

## 2. 專案開啟（使用 Docker）

若要快速啟動 `NativeCloud_HR` 專案的開發環境，建議使用 Docker 來建立本地測試環境。以下是步驟說明。

### 2.1. 安裝 Docker 與 Docker Compose

#### macOS / Windows：

若你是使用 macOS 或 Windows，**必須安裝並開啟 Docker Desktop** 才能運行 Docker 容器。

請依照以下步驟進行：

1. 前往 [Docker Desktop 官網](https://www.docker.com/products/docker-desktop/) 下載並安裝對應作業系統的 Docker Desktop。
2. 安裝完成後，**請確認 Docker Desktop 已啟動**，並可正常執行容器。
3. Docker Compose 通常會隨 Docker Desktop 一起安裝，無需額外安裝。

> 💡 **注意**：如未啟動 Docker Desktop，執行 `docker-compose up --build` 可能會出現找不到 Docker daemon 的錯誤。

####  Linux：

本專案已在 Linux 環境中測試通過，**不需要額外安裝 Docker Desktop**，只需安裝 Docker Engine 與 Docker Compose 即可。

請參考以下官方指引安裝：

* 安裝 Docker：[Docker 安裝指南](https://docs.docker.com/engine/install/)
* 安裝 Docker Compose：[Docker Compose 安裝指南](https://docs.docker.com/compose/install/)

### 2.2. 安裝 Go 環境

請確保系統已安裝 [Go](https://go.dev/doc/install) 並設置好環境變數。

⚠️ **建議在「本機終端機」執行 `go mod tidy`，不要在 VS Code 內建終端機執行，避免依賴拉取錯誤。**

```bash
go version   # 確認 Go 已正確安裝
```

### 2.3. 克隆專案

將專案代碼克隆到本地端：

```bash
git clone https://github.com/4040www/NativeCloud_HR.git
cd NativeCloud_HR
```

### 2.4. 配置 `.env` 環境變數

請在 `config/` 資料夾下**直接建立一個 `.env` 檔案**，不可放在 `.env/` 子資料夾中。

你可以參考 `config/.env.example` 來設定：

```bash
# config/.env
DB_HOST = 35.221.151.72
DB_USER =（補）
DB_PASSWORD =（補）
DB_NAME =（補）
DB_PORT = 5432

JWT_SECRET=（補）
```

⚠️ 請確保資料庫連線資訊為**最新版本**，如有更新請依最新提供的設定檔為主。

### 2.5. 下載依賴並啟動 Docker 容器

```bash
go mod tidy
docker-compose up --build
```

此命令會根據 `docker-compose.yml` 配置，建立並啟動容器，包含：

* API server
* PostgreSQL 資料庫
* Kafka message queue（如有）

### 2.6. 訪問應用

* 本地 API 端點：`http://localhost:8080`
* 健康檢查 API：`http://localhost:8080/api/status`

你可使用 Postman 等工具測試 API，包含：

* 身分驗證
* 打卡功能
* 提醒通知
* 報表生成等

### 2.7. 資料庫遷移

```bash
docker-compose exec app ./scripts/migrate.sh
```

執行後會自動更新資料表結構至最新版本。

### 2.8. 停止容器

```bash
docker-compose down
```

### 2.9. 查看日誌

```bash
docker-compose logs -f
```

這將顯示即時的運行紀錄，便於 debug。

