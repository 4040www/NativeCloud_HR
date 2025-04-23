package utils

import (
    "NativeCloud_HR/model"
    "github.com/golang-jwt/jwt/v5"
    "time"
)

var jwtSecret = []byte("your-secret-key") // 可放到 config

func GenerateJWT(user *model.User) (string, error) {
    claims := jwt.MapClaims{
        "user_id": user.ID,
        "email":   user.Email,
        "role":    user.Role,
        "exp":     time.Now().Add(time.Hour * 72).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}
