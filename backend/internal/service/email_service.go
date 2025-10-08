package service

import (
	"fmt"
	"strings"
	"time"

	"jobView-backend/internal/model"
	"jobView-backend/internal/repository"
	"jobView-backend/internal/utils"
)

// EmailIntegrationService 处理用户邮箱绑定与凭据管理
type EmailIntegrationService struct {
	repo          repository.EmailIntegrationRepository
	encryptionKey string
}

func NewEmailIntegrationService(repo repository.EmailIntegrationRepository, encryptionKey string) *EmailIntegrationService {
	return &EmailIntegrationService{repo: repo, encryptionKey: encryptionKey}
}

// BindMailbox 绑定或更新用户邮箱授权
func (s *EmailIntegrationService) BindMailbox(userID uint, req *model.MailboxBindRequest) (*model.MailboxResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("请求不能为空")
	}

	if err := utils.ValidateEmail(req.EmailAddress); err != nil {
		return nil, err
	}
	if err := utils.ValidateMailboxProtocol(req.Protocol); err != nil {
		return nil, err
	}
	if err := utils.ValidateHost(req.Host); err != nil {
		return nil, err
	}
	if err := utils.ValidatePort(req.Port); err != nil {
		return nil, err
	}

	authorization := strings.TrimSpace(req.AuthorizationCode)
	if authorization == "" {
		return nil, utils.ValidationError{Field: "authorization_code", Message: "授权码不能为空"}
	}

	encrypted, err := utils.EncryptString(s.encryptionKey, authorization)
	if err != nil {
		return nil, fmt.Errorf("加密授权码失败: %w", err)
	}

	emailLower := strings.ToLower(strings.TrimSpace(req.EmailAddress))
	protocol := strings.ToLower(strings.TrimSpace(req.Protocol))
	host := strings.TrimSpace(req.Host)

	mailbox := &model.UserMailbox{
		UserID:          userID,
		EmailAddress:    emailLower,
		Provider:        req.Provider,
		Protocol:        protocol,
		Host:            host,
		Port:            req.Port,
		UseSSL:          req.UseSSL,
		EncryptedSecret: encrypted,
		Status:          "active",
		UpdatedAt:       time.Now(),
	}

	saved, err := s.repo.UpsertMailbox(mailbox)
	if err != nil {
		return nil, fmt.Errorf("保存邮箱配置失败: %w", err)
	}

	return saved.ToResponse(), nil
}

// GetMailbox 查询用户已绑定邮箱
func (s *EmailIntegrationService) GetMailbox(userID uint) (*model.MailboxResponse, error) {
	mailbox, err := s.repo.GetMailboxByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("查询邮箱配置失败: %w", err)
	}
	return mailbox.ToResponse(), nil
}

// RemoveMailbox 解除绑定
func (s *EmailIntegrationService) RemoveMailbox(userID uint) error {
	if err := s.repo.DeleteMailbox(userID); err != nil {
		return fmt.Errorf("删除邮箱配置失败: %w", err)
	}
	return nil
}

// UpdateSyncResult 更新同步状态（供调度任务使用）
func (s *EmailIntegrationService) UpdateSyncResult(userID uint, uid *string, syncedAt time.Time, status string, errMsg *string) error {
	if err := s.repo.UpdateSyncState(userID, uid, syncedAt, status, errMsg); err != nil {
		return fmt.Errorf("更新同步状态失败: %w", err)
	}
	return nil
}
