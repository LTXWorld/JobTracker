package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"jobView-backend/internal/database"
	"jobView-backend/internal/model"
)

// MailEventRepository 负责存储邮件解析结果
type MailEventRepository interface {
	Create(event *model.MailEvent) (*model.MailEvent, error)
	ListPendingByUser(userID uint) ([]model.MailEvent, error)
	MarkStatus(id int, status string, errMsg *string) error
	UpdateProcessingResult(id int, applicationID *int, status string, errMsg *string) error
	GetByIDForUser(userID uint, id int) (*model.MailEvent, error)
}

type mailEventRepo struct{ db *database.DB }

func NewMailEventRepository(db *database.DB) MailEventRepository {
	return &mailEventRepo{db: db}
}

func (r *mailEventRepo) ensureDB() error {
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return nil
}

func (r *mailEventRepo) Create(event *model.MailEvent) (*model.MailEvent, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event cannot be nil")
	}
	query := `INSERT INTO mail_events (
		user_id, mailbox_id, application_id, message_id, message_uid, subject, sender, received_at, snippet,
		classification, confidence, payload, status, error_message, created_at, updated_at
	) VALUES (
		$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,NOW(),NOW()
	) RETURNING id, created_at, updated_at`

	payload := event.Payload
	if payload == nil {
		payload = json.RawMessage([]byte("{}"))
	}
	var createdAt, updatedAt time.Time
	if err := r.db.QueryRow(query,
		event.UserID,
		event.MailboxID,
		event.ApplicationID,
		event.MessageID,
		event.MessageUID,
		event.Subject,
		event.Sender,
		event.ReceivedAt,
		event.Snippet,
		event.Classification,
		event.Confidence,
		payload,
		event.Status,
		event.ErrorMessage,
	).Scan(&event.ID, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	event.CreatedAt = createdAt
	event.UpdatedAt = updatedAt
	return event, nil
}

func (r *mailEventRepo) ListPendingByUser(userID uint) ([]model.MailEvent, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}
	query := `SELECT id, user_id, mailbox_id, application_id, message_id, message_uid, subject, sender, received_at, snippet,
		classification, confidence, payload, status, error_message, created_at, updated_at
	FROM mail_events WHERE user_id = $1 AND status IN ('pending', 'needs_review') ORDER BY received_at DESC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []model.MailEvent
	for rows.Next() {
		var event model.MailEvent
		if err := rows.Scan(
			&event.ID,
			&event.UserID,
			&event.MailboxID,
			&event.ApplicationID,
			&event.MessageID,
			&event.MessageUID,
			&event.Subject,
			&event.Sender,
			&event.ReceivedAt,
			&event.Snippet,
			&event.Classification,
			&event.Confidence,
			&event.Payload,
			&event.Status,
			&event.ErrorMessage,
			&event.CreatedAt,
			&event.UpdatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func (r *mailEventRepo) MarkStatus(id int, status string, errMsg *string) error {
	if err := r.ensureDB(); err != nil {
		return err
	}
	query := `UPDATE mail_events SET status = $1, error_message = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.Exec(query, status, errMsg, id)
	return err
}

// Helper interface to reuse scan logic
type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanMailEvent(row rowScanner) (*model.MailEvent, error) {
	var event model.MailEvent
	if err := row.Scan(
		&event.ID,
		&event.UserID,
		&event.MailboxID,
		&event.ApplicationID,
		&event.MessageID,
		&event.MessageUID,
		&event.Subject,
		&event.Sender,
		&event.ReceivedAt,
		&event.Snippet,
		&event.Classification,
		&event.Confidence,
		&event.Payload,
		&event.Status,
		&event.ErrorMessage,
		&event.CreatedAt,
		&event.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &event, nil
}

// UpdateProcessingResult 更新事件处理结果（应用匹配和状态）
func (r *mailEventRepo) UpdateProcessingResult(id int, applicationID *int, status string, errMsg *string) error {
	if err := r.ensureDB(); err != nil {
		return err
	}
	query := `UPDATE mail_events SET application_id = $1, status = $2, error_message = $3, updated_at = NOW() WHERE id = $4`
	_, err := r.db.Exec(query, applicationID, status, errMsg, id)
	return err
}

func (r *mailEventRepo) GetByIDForUser(userID uint, id int) (*model.MailEvent, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}
	query := `SELECT id, user_id, mailbox_id, application_id, message_id, message_uid, subject, sender, received_at, snippet,
		classification, confidence, payload, status, error_message, created_at, updated_at
	FROM mail_events WHERE id = $1 AND user_id = $2`
	row := r.db.QueryRow(query, id, userID)
	return scanMailEvent(row)
}
