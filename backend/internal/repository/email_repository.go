package repository

import (
	"database/sql"
	"fmt"
	"time"

	"jobView-backend/internal/database"
	"jobView-backend/internal/model"
)

// EmailIntegrationRepository 定义邮箱绑定数据操作
type EmailIntegrationRepository interface {
	UpsertMailbox(mailbox *model.UserMailbox) (*model.UserMailbox, error)
	GetMailboxByUser(userID uint) (*model.UserMailbox, error)
	DeleteMailbox(userID uint) error
	UpdateSyncState(userID uint, uid *string, syncedAt time.Time, status string, errMsg *string) error
	ListActiveMailboxes() ([]*model.UserMailbox, error)
}

type scanner interface {
	Scan(dest ...interface{}) error
}

type emailRepo struct{ db *database.DB }

func NewEmailIntegrationRepository(db *database.DB) EmailIntegrationRepository {
	return &emailRepo{db: db}
}

func (r *emailRepo) ensureDB() error {
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return nil
}

func (r *emailRepo) UpsertMailbox(mailbox *model.UserMailbox) (*model.UserMailbox, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}

	query := `INSERT INTO user_mailboxes (
		user_id, email_address, provider, protocol, host, port, use_ssl, encrypted_password, status, last_message_uid, last_synced_at, error_message, created_at, updated_at
	) VALUES (
		$1,$2,$3,$4,$5,$6,$7,$8,'active',NULL,NULL,NULL,NOW(),NOW()
	) ON CONFLICT (user_id)
	DO UPDATE SET 
		email_address = EXCLUDED.email_address,
		provider = EXCLUDED.provider,
		protocol = EXCLUDED.protocol,
		host = EXCLUDED.host,
		port = EXCLUDED.port,
		use_ssl = EXCLUDED.use_ssl,
		encrypted_password = EXCLUDED.encrypted_password,
		status = 'active',
		error_message = NULL,
		updated_at = NOW()
	RETURNING id, user_id, email_address, provider, protocol, host, port, use_ssl, encrypted_password, status, last_message_uid, last_synced_at, error_message, created_at, updated_at`

	row := r.db.QueryRow(query,
		mailbox.UserID,
		mailbox.EmailAddress,
		mailbox.Provider,
		mailbox.Protocol,
		mailbox.Host,
		mailbox.Port,
		mailbox.UseSSL,
		mailbox.EncryptedSecret,
	)
	return scanMailbox(row)
}

func (r *emailRepo) GetMailboxByUser(userID uint) (*model.UserMailbox, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}
	query := `SELECT id, user_id, email_address, provider, protocol, host, port, use_ssl, encrypted_password, status, last_message_uid, last_synced_at, error_message, created_at, updated_at
	FROM user_mailboxes WHERE user_id = $1`
	row := r.db.QueryRow(query, userID)
	return scanMailbox(row)
}

func (r *emailRepo) DeleteMailbox(userID uint) error {
	if err := r.ensureDB(); err != nil {
		return err
	}
	_, err := r.db.Exec(`DELETE FROM user_mailboxes WHERE user_id = $1`, userID)
	return err
}

func (r *emailRepo) UpdateSyncState(userID uint, uid *string, syncedAt time.Time, status string, errMsg *string) error {
	if err := r.ensureDB(); err != nil {
		return err
	}
	query := `UPDATE user_mailboxes SET last_message_uid = $1, last_synced_at = $2, status = $3, error_message = $4, updated_at = NOW() WHERE user_id = $5`
	_, err := r.db.Exec(query, uid, syncedAt, status, errMsg, userID)
	return err
}

func (r *emailRepo) ListActiveMailboxes() ([]*model.UserMailbox, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}
	query := `SELECT id, user_id, email_address, provider, protocol, host, port, use_ssl, encrypted_password, status, last_message_uid, last_synced_at, error_message, created_at, updated_at
	FROM user_mailboxes WHERE status = 'active' OR status = 'needs_review'`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mailboxes []*model.UserMailbox
	for rows.Next() {
		mailbox, err := scanMailbox(rows)
		if err != nil {
			return nil, err
		}
		mailboxes = append(mailboxes, mailbox)
	}
	return mailboxes, nil
}

func scanMailbox(row scanner) (*model.UserMailbox, error) {
	var mailbox model.UserMailbox
	if err := row.Scan(
		&mailbox.ID,
		&mailbox.UserID,
		&mailbox.EmailAddress,
		&mailbox.Provider,
		&mailbox.Protocol,
		&mailbox.Host,
		&mailbox.Port,
		&mailbox.UseSSL,
		&mailbox.EncryptedSecret,
		&mailbox.Status,
		&mailbox.LastMessageUID,
		&mailbox.LastSyncedAt,
		&mailbox.ErrorMessage,
		&mailbox.CreatedAt,
		&mailbox.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &mailbox, nil
}
