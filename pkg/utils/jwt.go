// Package utils pkg/utils/jwt.go
package utils

import (
	"errors"
	"star-go/pkg/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims 自定义JWT Claims
type JWTClaims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateAccessToken 生成访问令牌
func GenerateAccessToken(userID uint64) (string, error) {
	cfg := config.GetConfig().JWT

	// 设置JWT声明
	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.AccessTokenExp) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.TokenIssuer,
		},
	}

	// 创建JWT令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString([]byte(cfg.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateRefreshToken 生成刷新令牌
func GenerateRefreshToken(userID uint64) (string, error) {
	cfg := config.GetConfig().JWT

	// 设置JWT声明 - 刷新令牌只包含最小必要信息
	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.RefreshTokenExp) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.TokenIssuer,
		},
	}

	// 创建JWT令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString([]byte(cfg.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseAccessToken 解析访问令牌
func ParseAccessToken(tokenString string) (*JWTClaims, error) {
	return parseToken(tokenString)
}

// ParseRefreshToken 解析刷新令牌
func ParseRefreshToken(tokenString string) (*JWTClaims, error) {
	return parseToken(tokenString)
}

// 解析JWT令牌
func parseToken(tokenString string) (*JWTClaims, error) {
	cfg := config.GetConfig().JWT

	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return []byte(cfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证令牌有效性并转换为自定义Claims
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的令牌")
}

// ValidateToken 验证令牌是否有效
func ValidateToken(tokenString string) bool {
	_, err := ParseAccessToken(tokenString)
	return err == nil
}
