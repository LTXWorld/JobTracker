// /Users/lutao/GolandProjects/jobView/backend/internal/model/user.go
// 用户模型定义，包含用户基础信息和认证相关的数据结构
// 与认证服务和处理器配合使用，实现完整的用户认证和授权功能

package model

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// User 用户模型
type User struct {
	ID        uint      `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"` // 不在JSON中暴露密码
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	User         *UserProfile `json:"user"`
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    int64        `json:"expires_at"`
}

// UserProfile 用户信息（不包含密码）
type UserProfile struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateUserRequest 更新用户信息请求
type UpdateUserRequest struct {
	Username *string `json:"username,omitempty" binding:"omitempty,min=3,max=50"`
	Email    *string `json:"email,omitempty" binding:"omitempty,email"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// RefreshTokenRequest 刷新token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// TokenClaims JWT token声明
type TokenClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Type     string `json:"type"` // "access" 或 "refresh"
}

// Value 实现 driver.Valuer 接口
func (u User) Value() (driver.Value, error) {
	return u.ID, nil
}

// Scan 实现 sql.Scanner 接口
func (u *User) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case int64:
		u.ID = uint(v)
		return nil
	case uint:
		u.ID = v
		return nil
	}
	return fmt.Errorf("cannot scan %T into User", value)
}

// ToProfile 转换为用户信息（不包含密码）
func (u *User) ToProfile() *UserProfile {
	return &UserProfile{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// ValidatePassword 验证密码强度
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("密码长度不能少于8位")
	}
	
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false
	
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= '!' && char <= '/' || char >= ':' && char <= '@' || char >= '[' && char <= '`' || char >= '{' && char <= '~':
			hasSpecial = true
		}
	}
	
	if !hasUpper {
		return fmt.Errorf("密码必须包含至少一个大写字母")
	}
	if !hasLower {
		return fmt.Errorf("密码必须包含至少一个小写字母")
	}
	if !hasDigit {
		return fmt.Errorf("密码必须包含至少一个数字")
	}
	if !hasSpecial {
		return fmt.Errorf("密码必须包含至少一个特殊字符")
	}
	
	return nil
}