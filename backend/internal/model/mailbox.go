package model

import "time"

// UserMailbox 保存用户邮箱授权与同步信息
type UserMailbox struct {
	ID              int        `json:"id" db:"id"`
	UserID          uint       `json:"user_id" db:"user_id"`
	EmailAddress    string     `json:"email_address" db:"email_address"`
	Provider        *string    `json:"provider,omitempty" db:"provider"`
	Protocol        string     `json:"protocol" db:"protocol"`
	Host            string     `json:"host" db:"host"`
	Port            int        `json:"port" db:"port"`
	UseSSL          bool       `json:"use_ssl" db:"use_ssl"`
	EncryptedSecret string     `json:"-" db:"encrypted_password"`
	LastMessageUID  *string    `json:"last_message_uid,omitempty" db:"last_message_uid"`
	LastSyncedAt    *time.Time `json:"last_synced_at,omitempty" db:"last_synced_at"`
	Status          string     `json:"status" db:"status"`
	ErrorMessage    *string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// MailboxBindRequest 用户邮箱绑定请求体
type MailboxBindRequest struct {
	EmailAddress      string  `json:"email_address"`
	Provider          *string `json:"provider"`
	Protocol          string  `json:"protocol"`
	Host              string  `json:"host"`
	Port              int     `json:"port"`
	UseSSL            bool    `json:"use_ssl"`
	AuthorizationCode string  `json:"authorization_code"`
}

// MailboxResponse 返回给前端的邮箱绑定信息
type MailboxResponse struct {
	EmailAddress      string     `json:"email_address"`
	Provider          *string    `json:"provider,omitempty"`
	Protocol          string     `json:"protocol"`
	Host              string     `json:"host"`
	Port              int        `json:"port"`
	UseSSL            bool       `json:"use_ssl"`
	Status            string     `json:"status"`
	LastSyncedAt      *time.Time `json:"last_synced_at,omitempty"`
	RequiresAttention bool       `json:"requires_attention"`
	ErrorMessage      *string    `json:"error_message,omitempty"`
}

// ToResponse 转换为前端展示结构
func (mb *UserMailbox) ToResponse() *MailboxResponse {
	if mb == nil {
		return nil
	}
	return &MailboxResponse{
		EmailAddress:      mb.EmailAddress,
		Provider:          mb.Provider,
		Protocol:          mb.Protocol,
		Host:              mb.Host,
		Port:              mb.Port,
		UseSSL:            mb.UseSSL,
		Status:            mb.Status,
		LastSyncedAt:      mb.LastSyncedAt,
		RequiresAttention: mb.Status == "error" || mb.Status == "pending_review",
		ErrorMessage:      mb.ErrorMessage,
	}
}
