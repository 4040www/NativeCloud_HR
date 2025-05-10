# NativeCloud_HR 部署到 GCP

這篇指南將介紹如何將 `NativeCloud_HR` 應用部署到 Google Cloud Platform (GCP) 上。

## 前置條件

1. 擁有一個 Google Cloud 帳戶。
2. 已安裝並配置 [Google Cloud SDK](https://cloud.google.com/sdk)。
3. 已安裝 [Docker](https://www.docker.com/) 和 [Docker Compose](https://docs.docker.com/compose/)。如果需要，請參考相應文檔進行安裝。

## 步驟

### 1. 設定 GCP 環境

#### 1.1. 登入 GCP 並設置專案
- 登入 [Google Cloud Console](https://console.cloud.google.com/) 並創建一個新的專案。
- 使用以下命令初始化 Google Cloud SDK 並設置專案：
  ```bash
  gcloud init
  ```

#### 1.2. 啟用 GCP 服務
- 啟用以下 GCP 服務：
  - **Compute Engine**（如果使用虛擬機部署）
  - **Cloud SQL**（如果使用 GCP 的資料庫服務）
  - **Cloud Storage**（如果需要存儲）

  這些服務可以通過 Google Cloud Console 啟用。

### 2. 部署應用到 GCP

#### 2.1. 使用 Google Compute Engine 部署

##### 2.1.1. 創建虛擬機實例
- 進入 GCP Console，前往 **Compute Engine > VM instances**，然後點擊 **Create Instance**。
- 選擇你所需的虛擬機配置（例如選擇操作系統、CPU、內存等）。
- 設定防火牆規則，確保允許 HTTP 和 HTTPS 流量。
  
##### 2.1.2. 設定 VM 並 SSH 連接
- 連接到虛擬機：
  ```bash
  gcloud compute ssh <your-instance-name> --zone=<your-zone>
  ```

##### 2.1.3. 安裝 Docker 和 Docker Compose
- 在 VM 上安裝 Docker 和 Docker Compose：
  ```bash
  sudo apt-get update
  sudo apt-get install docker.io docker-compose
  ```

##### 2.1.4. 部署應用
- 將你的專案代碼推送到 VM，然後執行以下命令來啟動應用：
  ```bash
  git clone <your-repo-url>
  cd <your-project-directory>
  docker-compose up -d
  ```

##### 2.1.5. 配置防火牆規則
- 配置防火牆規則來允許流量訪問應用，開放 8080 埠：
  ```bash
  gcloud compute firewall-rules create allow-http --allow tcp:8080
  ```

#### 2.2. 使用 Google Kubernetes Engine (GKE) 部署（可選）

如果你希望使用 Kubernetes 部署應用，請遵循以下步驟：

##### 2.2.1. 創建 GKE 集群
- 在 GCP Console，前往 **Kubernetes Engine > Clusters**，點擊 **Create Cluster**。
- 設定集群配置並創建。

##### 2.2.2. 部署 Docker 容器
- 使用 `kubectl` 部署應用，首先安裝 `kubectl`：
  ```bash
  gcloud components install kubectl
  ```

  創建一個 `deployment.yaml` 文件來部署你的應用：
  ```yaml
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: native-cloud-hr
  spec:
    replicas: 1
    selector:
      matchLabels:
        app: native-cloud-hr
    template:
      metadata:
        labels:
          app: native-cloud-hr
      spec:
        containers:
        - name: native-cloud-hr
          image: <your-docker-image>
          ports:
          - containerPort: 8080
  ```

  然後使用 `kubectl` 部署應用：
  ```bash
  kubectl apply -f deployment.yaml
  ```

##### 2.2.3. 配置負載平衡和外部 IP
- 使用 GKE 自帶的負載平衡來公開你的應用服務，這樣可以從外部訪問。

### 3. 設定資料庫

如果你的應用依賴於資料庫，可以選擇使用 **Cloud SQL** 或者自建 Docker 容器。

#### 3.1. 使用 Cloud SQL 部署 PostgreSQL
- 在 GCP Console 中創建 **Cloud SQL** PostgreSQL 實例，並配置連接。
- 配置應用與 Cloud SQL 的連接，更新應用中的資料庫設定（如資料庫 IP、用戶名和密碼）。

#### 3.2. 自建 PostgreSQL 容器（可選）
如果你希望自己管理 PostgreSQL 容器，可以在 GCP 的 VM 上運行 PostgreSQL 容器。

### 4. 設定持續集成與部署 (CI/CD)

為了實現自動化部署，你可以設置 CI/CD 管道：

- 使用 **Google Cloud Build** 來自動構建並部署應用。
- 配置 **GitHub Actions** 或 **GitLab CI** 來自動觸發 Google Cloud Build 進行部署。

### 5. 網路與安全設置

確保正確設置應用的網路和安全性：

- 配置防火牆規則開放必要的端口。
- 使用適當的身份驗證和授權方法來保護你的應用，例如使用 JWT 或 OAuth。

### 6. 監控與日誌

- 使用 GCP 的 **Stackdriver** 監控服務來跟踪應用的性能。
- 設置 **Google Cloud Logging** 來檢查應用日誌，確保系統運行順利。
