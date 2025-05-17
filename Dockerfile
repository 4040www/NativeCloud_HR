# 第一階段: 建置 Go 二進位檔
FROM golang:1.21 AS builder

# 設定工作目錄
WORKDIR /app

# 複製 go.mod 和 go.sum 
COPY go.mod go.sum ./

# 確認 go.mod 和 go.sum 是否成功複製
RUN ls -l

# 複製所有程式碼
COPY . .

# 確認專案程式碼是否成功複製
RUN ls -R /app

# 下載依賴
RUN go mod download
RUN go mod tidy
# 編譯 Go 可執行檔（靜態編譯，適用於 GCP）
RUN CGO_ENABLED=0 GOOS=linux go build -o native-cloud-hr ./cmd/main.go


# 第二階段: 建立最小化的運行環境
FROM gcr.io/distroless/base-debian11

# 設定工作目錄
WORKDIR /root/

# 複製已編譯的 Go 執行檔
COPY --from=builder /app/native-cloud-hr /app/native-cloud-hr

# 設定環境變數（確保應用程式讀取 GCP 服務帳戶）
ENV GOOGLE_APPLICATION_CREDENTIALS="/root/gcp-service-account.json"

# 開放必要的 Port（假設 API 運行於 8080）
EXPOSE 8080

# 執行應用程式
CMD ["./native-cloud-hr"]
