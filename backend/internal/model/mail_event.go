package model

import (
	"encoding/json"
	"time"
)

// MailEvent 记录邮箱解析结果，供后续自动更新或人工确认使用
type MailEvent struct {
	ID             int             `json:"id" db:"id"`
	UserID         uint            `json:"user_id" db:"user_id"`
	MailboxID      int             `json:"mailbox_id" db:"mailbox_id"`
	ApplicationID  *int            `json:"application_id,omitempty" db:"application_id"`
	MessageID      *string         `json:"message_id,omitempty" db:"message_id"`
	MessageUID     *string         `json:"message_uid,omitempty" db:"message_uid"`
	Subject        string          `json:"subject" db:"subject"`
	Sender         string          `json:"sender" db:"sender"`
	ReceivedAt     time.Time       `json:"received_at" db:"received_at"`
	Snippet        *string         `json:"snippet,omitempty" db:"snippet"`
	Classification string          `json:"classification" db:"classification"`
	Confidence     float64         `json:"confidence" db:"confidence"`
	Payload        json.RawMessage `json:"payload" db:"payload"`
	Status         string          `json:"status" db:"status"`
	ErrorMessage   *string         `json:"error_message,omitempty" db:"error_message"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

// MailEventPayload 邮件解析后的结构化数据
type MailEventPayload struct {
	CompanyCandidates  []string   `json:"company_candidates,omitempty"`
	PositionCandidates []string   `json:"position_candidates,omitempty"`
	DetectedTime       *time.Time `json:"detected_time,omitempty"`
	ExamLink           *string    `json:"exam_link,omitempty"`
	MeetingLink        *string    `json:"meeting_link,omitempty"`
	MeetingID          *string    `json:"meeting_id,omitempty"`
	RawLinks           []string   `json:"raw_links,omitempty"`
	Notes              []string   `json:"notes,omitempty"`
	MatchedApplication *int       `json:"matched_application,omitempty"`
}

// MailClassification 邮件分类枚举
const (
	MailClassificationUnknown     = "unknown"
	MailClassificationExam        = "exam"
	MailClassificationInterview   = "interview"
	MailClassificationInformation = "information"
)

// MailEventStatus 处理状态
const (
	MailEventStatusPending         = "pending"
	MailEventStatusProcessed       = "processed"
	MailEventStatusNeedsReview     = "needs_review"
	MailEventStatusProcessingError = "error"
	MailEventStatusDismissed       = "dismissed"
)

// MailEventApplicationSummary 供提醒中心展示的岗位摘要
type MailEventApplicationSummary struct {
	ID              int               `json:"id"`
	CompanyName     string            `json:"company_name"`
	PositionTitle   string            `json:"position_title"`
	Status          ApplicationStatus `json:"status"`
	InterviewTime   *time.Time        `json:"interview_time,omitempty"`
	ReminderTime    *time.Time        `json:"reminder_time,omitempty"`
	ReminderEnabled bool              `json:"reminder_enabled"`
}

// MailEventPendingItem 提供给前端的待确认事件结构
type MailEventPendingItem struct {
	ID             int                          `json:"id"`
	UserID         uint                         `json:"user_id"`
	MailboxID      int                          `json:"mailbox_id"`
	ApplicationID  *int                         `json:"application_id,omitempty"`
	Subject        string                       `json:"subject"`
	Sender         string                       `json:"sender"`
	ReceivedAt     time.Time                    `json:"received_at"`
	Snippet        *string                      `json:"snippet,omitempty"`
	Classification string                       `json:"classification"`
	Confidence     float64                      `json:"confidence"`
	Payload        MailEventPayload             `json:"payload"`
	Status         string                       `json:"status"`
	ErrorMessage   *string                      `json:"error_message,omitempty"`
	Application    *MailEventApplicationSummary `json:"application,omitempty"`
}

// MailEventStatusUpdateRequest 提供前端更新事件状态
type MailEventStatusUpdateRequest struct {
	Status        string  `json:"status"`
	ErrorMessage  *string `json:"error_message,omitempty"`
	ApplicationID *int    `json:"application_id,omitempty"`
}
