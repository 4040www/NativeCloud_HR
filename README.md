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

## 2. 專案開啟（使用 Docker）

若要快速啟動 `NativeCloud_HR` 專案的開發環境，建議使用 Docker 來建立本地測試環境。以下是步驟說明。

### 2.1. 安裝 Docker 與 Docker Compose

1. 安裝 Docker：[Docker 安裝指南](https://docs.docker.com/get-docker/)
2. 安裝 Docker Compose：[Docker Compose 安裝指南](https://docs.docker.com/compose/install/)

### 2.2. 克隆專案

首先，將專案代碼克隆到本地端：
```bash
git clone https://github.com/4040www/NativeCloud_HR.git
```

### 2.3. 配置 `.env` 環境變數

在`config`檔案夾下，創建 `.env` 檔案並配置相應的環境變數。你可以參考 `.env.example` 檔案進行配置：

```bash
# DB
DB_HOST = 35.221.151.72
DB_USER = （補）
DB_PASSWORD =（補）
DB_NAME =（補）
DB_PORT = 5432


# JWT
JWT_SECRET=（補）
```

### 2.4. 建立並啟動 Docker 容器

在專案根目錄下，執行以下命令來建立並啟動容器：
```bash
go mod tidy
docker-compose up --build
```

此命令會使用 `docker-compose.yml` 配置文件來構建並啟動容器，並在本地環境中啟動 PostgreSQL 資料庫和 Kafka 消息隊列。

### 2.5. 訪問應用

一旦容器啟動完成，你可以通過以下網址來訪問應用：
- 本地 API 端點：`http://localhost:8080`
- 健康檢查 API：`http://localhost:8080/api/status`

你也可以通過 API 測試工具（例如 Postman）來調試接口，測試相關的身分驗證、打卡、報表生成等功能。

### 2.6. 資料庫遷移

如果需要執行資料庫遷移，可以使用以下腳本來更新資料庫結構：
```bash
docker-compose exec app ./scripts/migrate.sh
```
這個腳本會將資料庫的結構更新到最新版本，並確保應用的資料庫與程式碼同步。

### 2.7. 停止容器

當你完成開發或測試後，可以使用以下命令停止 Docker 容器：
```bash
docker-compose down
```

### 2.8. 日誌查看

若需要查看應用的運行日誌，可以執行：
```bash
docker-compose logs -f
```
這會顯示容器的實時日誌，對於排查錯誤非常有用。
