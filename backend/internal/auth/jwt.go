// /Users/lutao/GolandProjects/jobView/backend/internal/auth/jwt.go
// JWT token 生成、验证和解析工具
// 提供访问token和刷新token的完整管理功能，确保安全的用户认证

package auth

import (
	"errors"
	"jobView-backend/internal/model"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// JWT密钥从环境变量获取，如果未设置则使用默认值（生产环境必须设置）
	jwtSecret     = []byte(getEnvOrDefault("JWT_SECRET", "your-256-bit-secret"))
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token has expired")
)

// TokenDuration token有效期配置
const (
	AccessTokenDuration  = 24 * time.Hour    // 访问token有效期：24小时
	RefreshTokenDuration = 30 * 24 * time.Hour // 刷新token有效期：30天
)

// CustomClaims JWT自定义声明
type CustomClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Type     string `json:"type"` // "access" 或 "refresh"
	jwt.RegisteredClaims
}

// GenerateTokenPair 生成访问token和刷新token对
func GenerateTokenPair(user *model.User) (accessToken, refreshToken string, err error) {
	now := time.Now()
	
	// 生成访问token
	accessClaims := &CustomClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "jobview-backend",
			Subject:   user.Username,
		},
	}
	
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}
	
	// 生成刷新token
	refreshClaims := &CustomClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Type:     "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(RefreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "jobview-backend",
			Subject:   user.Username,
		},
	}
	
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}
	
	return accessToken, refreshToken, nil
}

// ValidateToken 验证token并返回用户信息
func ValidateToken(tokenString string, expectedType string) (*CustomClaims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}
	
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}
	
	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	
	// 验证token类型
	if claims.Type != expectedType {
		return nil, ErrInvalidToken
	}
	
	// JWT库已经处理了过期检查，这里不需要重复检查
	// 移除: if claims.ExpiresAt.Before(time.Now()) { return nil, ErrTokenExpired }
	
	return claims, nil
}

// ValidateAccessToken 验证访问token
func ValidateAccessToken(tokenString string) (*CustomClaims, error) {
	return ValidateToken(tokenString, "access")
}

// ValidateRefreshToken 验证刷新token
func ValidateRefreshToken(tokenString string) (*CustomClaims, error) {
	return ValidateToken(tokenString, "refresh")
}

// ExtractTokenFromHeader 从Authorization header中提取token
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}
	
	// 检查Bearer前缀
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) {
		return "", errors.New("invalid authorization header format")
	}
	
	if authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("authorization header must start with Bearer")
	}
	
	return authHeader[len(bearerPrefix):], nil
}

// GetTokenExpirationTime 获取token过期时间
func GetTokenExpirationTime(tokenType string) time.Duration {
	switch tokenType {
	case "access":
		return AccessTokenDuration
	case "refresh":
		return RefreshTokenDuration
	default:
		return AccessTokenDuration
	}
}

// IsTokenExpired 检查token是否即将过期（30分钟内）
func IsTokenExpired(claims *CustomClaims) bool {
	return claims.ExpiresAt.Before(time.Now().Add(30 * time.Minute))
}

// getEnvOrDefault 获取环境变量或返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}