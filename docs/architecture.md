# 架構概述

本文檔提供了後端 API 的架構概述，詳細描述了應用的結構、組件及其運作流程。旨在讓開發者了解整體設計，以便更好地參與專案開發。

## 1. 開發環境

### 1.1 使用 VS Code 擴展 Go 開發（請自行下載）

我們使用 **Visual Studio Code (VS Code)** 的 **Go** 擴展來開發，這提供了極好的開發體驗。該擴展的一個關鍵特性是自動代碼格式化。在你按下 **Ctrl + S** 儲存文件時，會自動格式化代碼，遵循 Go 標準的代碼風格，這樣可以保持代碼整潔且一致。當下載完成後，可以執行：
```go
go mod tidy
```
安裝套件並確保你專案需要的依賴、版本檔案都是正確的。

## 2. 後端路由與邏輯

後端 API 遵循乾淨且模組化的架構模式。以下是路由與邏輯流程的高層次概述：

### 2.1 `router.go` 和 `routes`

- **router.go**: `router.go` 文件中定義了所有的 API 路由。每個路由對應於一個 HTTP 方法（如 GET、POST 等），並映射到相應的處理函數（handler）。
  
- **routes**: 路由的組織和註冊通常會在這個文件中進行。這裡會指定每個端點的路徑，並將其與相應的處理函數綁定，例如 `/login`、`/logout`、`/users/{id}` 等。

範例：

```go
func setupRouter() *gin.Engine {
    r := gin.Default()

    // 公開路由
    r.POST("/login", handler.LoginHandler)
    r.POST("/logout", handler.LogoutHandler)

    // 需要認證的路由
    authorized := r.Group("/")
    authorized.Use(middleware.JWTMiddleware())
    authorized.GET("/profile", handler.ProfileHandler)

    return r
}
```

### 2.2 `handlers`

- **Handlers** 負責處理進來的 HTTP 請求。它們解析請求，進行驗證，然後將數據傳遞給服務層處理業務邏輯。
  
- Handler 通常處理的任務包括驗證 JSON 請求、從請求的標頭或 URL 參數中提取數據，以及調用服務層進行業務邏輯處理。

範例：

```go
func LoginHandler(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }

    user, token, err := service.AuthenticateUser(req.Email, req.Password)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Login successful",
        "token":   token,
        "user":    user,
    })
}
```

### 2.3 `services`

- **Services** 包含應用的核心業務邏輯。它們接收來自 handler 的數據，根據應用的業務規則進行處理，並將結果返回給 handler。

- 服務層會與資料庫存取層（repository）互動，從資料庫中檢索或保存數據。

範例：

```go
func AuthenticateUser(email, password string) (*model.User, string, error) {
    user, err := repository.FindUserByEmail(email)
    if err != nil {
        return nil, "", err
    }

    if !utils.CheckPasswordHash(password, user.Password) {
        return nil, "", errors.New("invalid password")
    }

    token, err := utils.GenerateJWT(user)
    if err != nil {
        return nil, "", err
    }

    return user, token, nil
}
```

### 2.4 `repository`

- **Repository** 作為資料存取層，直接與資料庫進行交互。它包含了檢索、插入、更新和刪除資料的功能。

- 我們使用 **GORM** 這個 ORM（物件關聯映射）庫來抽象 SQL 查詢，讓開發者能夠通過 Go 函數調用來操作資料庫，而不需要手寫 SQL 語句。

範例：

```go
func FindUserByEmail(email string) (*model.User, error) {
    var user model.User
    if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
        return nil, err
    }
    return &user, nil
}
```

## 3. 資料庫：使用 GORM

應用程式使用 **GORM** 作為物件關聯映射（ORM）工具來與資料庫進行交互。GORM 允許我們定義 Go 結構體來映射資料庫中的表格，並提供一個簡單易用的 API 來查詢和修改資料。

- **Model**: Model 代表資料庫中的一個實體，例如，`User` model 代表資料庫中的使用者。

- **GORM** 抽象了 SQL 查詢，使開發者可以操作 Go 結構體，而不需要直接寫 SQL 字串。GORM 自動處理表格創建、關聯以及複雜查詢。

範例模型：

```go
type User struct {
    ID       uint   `gorm:"primaryKey"`
    Email    string `gorm:"unique"`
    Password string
    Role     string
}
```

### 3.1 連接資料庫

資料庫連接通過 GORM 在 `db` 包中進行初始化。`db.GetDB()` 函數提供了與資料庫互動的 GORM DB 實例。

資料庫連接設置範例：

```go
package db

import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "log"
)

var DB *gorm.DB

func InitDB() {
    var err error
    DB, err = gorm.Open(mysql.Open("your-database-connection-string"), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to the database: %v", err)
    }
}
```

## 4. 配置：加載 `.env` 和 `config.yaml`

應用的配置設定，例如資料庫連接資訊和其他環境變數，都是從 `.env` 和 `config.yaml` 檔案中讀取的。

### 4.1 `.env` 檔案

`.env` 檔案儲存環境特定的變數，如資料庫憑證、JWT 秘密金鑰和其他敏感資訊。

`.env` 檔案：

```env
# DB
DB_HOST=your-cloudsql-ip
DB_USER=myuser
DB_PASSWORD=mypass
DB_NAME=mydb
DB_PORT=5432


# JWT
JWT_SECRET=sdgrlgrwi875e54reg54rwr45gwrg
```

### 4.2 `config.yaml` 檔案

`config.yaml` 檔案包含應用運行所需的設定，例如服務器端口、資料庫連接資訊等。

`config.yaml` 檔案：

```yaml
sserver:
  port: 8080

database:
  host: "localhost"
  port: 5432
  user: "admin"
  password: "password"
  name: "employee_db"
  
messageQueue:
  type: "kafka"
  brokers:
    - "localhost:9092"
```

### 4.3 加載配置

`config.LoadConfig()` 函數會加載 `.env` 和 `config.yaml` 中的設定，並將它們提供給應用程式使用。

配置加載範例：

```go
package config

import (
    "github.com/spf13/viper"
    "log"
    "os"
)

var Config struct {
    ServerPort int    `mapstructure:"server_port"`
    DBHost     string `mapstructure:"db_host"`
    DBUser     string `mapstructure:"db_user"`
    DBPassword string `mapstructure:"db_password"`
    DBName     string `mapstructure:"db_name"`
}

func LoadConfig() error {
    viper.SetConfigFile(".env")
    viper.AddConfigPath(".")
    viper.ReadInConfig()

    if err := viper.Unmarshal(&Config); err != nil {
        log.Fatalf("Error unmarshalling config: %v", err)
        return err
    }

    return nil
}
```

## 5. 錯誤處理

在應用的每一層（handler、service、repository）中進行錯誤處理，確保 API 能返回適當的響應代碼和錯誤訊息。

- **400 Bad Request**: 請求格式錯誤或資料無效。
- **401 Unauthorized**: JWT 無效或缺少授權。
- **500 Internal Server Error**: 伺服器內部錯誤。

上次更改日期: 2025 / 04 / 23