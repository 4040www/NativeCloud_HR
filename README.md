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
│   ├── 📝 config.yaml      # 設定檔案 (GCP, DB, Kafka, NATS 連線)
│── 📂 internal             # 核心業務邏輯
│   ├── 📂 api              # API 層
│   │   ├── 📝 handlers.go  # HTTP 請求處理
│   │   ├── 📝 middleware.go# 身分驗證、日誌等中介軟體
│   │   ├── 📝 router.go    # 設定 API 路由 (Echo/Gin)
│   ├── 📂 service          # 服務層（業務邏輯）
│   │   ├── 📝 access.go    # 員工刷卡業務邏輯
│   │   ├── 📝 report.go    # 報表分析邏輯
│   │   ├── 📝 auth.go      # 員工登入驗證邏輯
│   ├── 📂 repository       # 資料存取層
│   │   ├── 📝 employee.go  # 員工資料存取
│   │   ├── 📝 attendance.go# 考勤數據存取
│   ├── 📂 db               # 資料庫初始化
│   │   ├── 📝 postgres.go  # PostgreSQL 連線
│   │   ├── 📝 redis.go     # Redis 快取連線
│   ├── 📂 auth             # 身分驗證模組
│   │   ├── 📝 jwt.go       # JWT Token 相關邏輯
│   │   ├── 📝 oauth.go     # OAuth 2.0 整合 (Google Workspace, AD)
│   ├── 📂 utils            # 通用工具
│   │   ├── 📝 logger.go    # 日誌系統
│   │   ├── 📝 response.go  # API 回應格式
│   ├── 📂 messageQueue     # 負責消息隊列操作
│   │   ├── 📝 kafka.go     # Kafka 配置與生產者
│   │   ├── 📝 nats.go      # NATS 配置與生產者
│   │   ├── 📝 producer.go  # 發送消息的生產者
│   │   ├── 📝 consumer.go  # 接收消息的消費者
│   ├── 📂 events           # 事件處理邏輯
│   │   ├── 📝 eventHandler.go # 事件處理邏輯
│   │   ├── 📝 eventProcessor.go # 處理事件順序、優先級、並發限制等
├── 📂 deployments          # 部署相關
│   ├── 📝 Dockerfile       # 容器化設定
│   ├── 📝 docker-compose.yml # 本地測試環境
│   ├── 📝 cloudbuild.yaml  # GCP Cloud Build 設定
│   ├── 📝 k8s.yaml         # Kubernetes 部署設定
│── 📂 scripts              # 運維腳本
│   ├── 📝 migrate.sh       # 資料庫遷移
│   ├── 📝 start.sh         # 啟動指令
│── 📂 docs                 # 文件
│   ├── 📝 API.md           # API 說明文件
│   ├── 📝 architecture.md  # 系統架構說明
│── .env                    # 環境變數
│── go.mod                  # Golang 依賴管理
│── go.sum                  # Golang 依賴鎖定
│── README.md               # 專案說明文件
```

## 2. API 層
**作用**: 提供外部服務和系統的接口，讓用戶（如員工、主管等）與系統互動。

### 運作邏輯:
- **接收請求**: 使用 HTTP 請求處理（如 RESTful API）。 `src/api/`
- **處理請求**: 當接收到 API 請求時，會進行必要的驗證和業務邏輯處理。`src/services/`
- **返回響應**: 根據業務邏輯的結果，返回適當的響應數據。`src/utils/`

## 3. Message Queue 層 (Kafka / NATS)
**作用**: 用來解耦系統內部不同模塊的交互，實現異步處理。相關檔案放置於 `src/messageQueue/` 跟 `src/events/`中。

### 運作邏輯:
- **事件發送**: 當刷卡事件發生時，會將事件資料（如員工 ID、門禁 ID、刷卡結果等）推送到消息隊列（Kafka 或 NATS）。`src/messageQueue/producer.js`
- **事件消費**: 後端服務會訂閱這些消息，接收到消息後執行業務邏輯（如寫入資料庫、發送 WebSocket 通知）。`src/messageQueue/consumer.js`
- **事件處理順序**: 根據事件的嚴重性或處理優先級，可能需要設計不同的事件處理順序或限制並發數量。`src/events/eventProcessor.js`

## 4. 資料庫相關 (PostgreSQL / Redis)
**作用**: 用來儲存系統的基本數據，並對數據進行操作（如讀取、寫入、更新）。相關檔案放置於 `src/database/` 和 `src/sync/` 中。

### 運作邏輯:
- **PostgreSQL**: 主要儲存結構化數據，如員工資訊、刷卡紀錄、考勤數據等，並支援高效的查詢操作。 `src/database/postgresClient.js`
- **Redis**: 作為緩存系統，快速查詢某些高頻資料（如員工的出勤紀錄、門禁狀態等），提高系統效能。`src/database/redisClient.js`
- **資料同步**: 需要確保資料在不同地點或服務之間同步（例如多地點刷卡系統數據的同步）。`src/sync/dataSync.js`

## 5. 系統架構
**作用**: 整體系統的運行架構，涉及多個子模塊和服務的協同工作。相關檔案放置於 `src/deploy/`，`src/docker/` 和 `src/monitoring/` 中。

### 運作邏輯:
- **分布式架構**: 使用 GCP 等雲端服務來部署系統，保證高可用性和可擴展性。`deploy/gcpDeployment.yaml`
- **高可用性**: 通過多區域部署和負載均衡來保證系統高可用，當一個區域的服務宕機時，其他區域能夠繼續提供服務。 `deploy/loadBalancerConfig.js`
- **容器化**: 使用 Docker 和 Kubernetes 進行容器化部署，讓系統具有彈性擴展的能力。 `docker/Dockerfile`
- **系統監控**: 需要有監控系統（如 Prometheus）來監視系統狀態，並在系統發生異常時提供警報。 `monitoring/prometheusConfig.yml`

## 6. 異常處理與報表
### 刷卡異常處理:

- 相關檔案放置於 `src/exceptions/` 中。
- **無刷入記錄但有刷出記錄**: 可能是員工未刷入或設備故障，需設定規則處理或允許主管手動處理。 `cardIssueHandler.js`
- **刷卡時間異常**: 若發現異常頻繁的刷卡行為，系統需自動標記並通知 HR 進行審查。`timeIssueHandler.js`
- **刷卡結果為 Denied**: 這種情況下應立即通知主管或安保，並考慮是否封鎖員工的門禁權限。 `deniedCardHandler.js`

### 報表設計:
- 相關檔案放置於 `src/reports/` 中。
- 設計能夠快速計算出勤、加班等報表，並支持每天、每週、每月的報表查詢。`attendanceReport.js`
- 考慮將計算結果緩存在 Redis 中，提升報表查詢速度，並減少資料庫壓力。 `reportCache.js`
