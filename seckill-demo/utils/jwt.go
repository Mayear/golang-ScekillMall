package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 签名密钥，实际项目中应该写在配置文件或环境变量里
var jwtSecret = []byte("seckill_super_secret_key_2026")

// MyClaims 自定义声明结构体
type MyClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT
func GenerateToken(userID int) (string, error) {
	// 设置 Token 过期时间（例如 24 小时）
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := MyClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "seckill-demo",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken 解析并校验 JWT
func ParseToken(tokenString string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
