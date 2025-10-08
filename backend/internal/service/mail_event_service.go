package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"jobView-backend/internal/model"
	"jobView-backend/internal/repository"
)

var (
	// ErrMailEventEmptyRequest 表示请求体为空
	ErrMailEventEmptyRequest = errors.New("mail event request is empty")
	// ErrMailEventStatusEmpty 表示状态字段为空
	ErrMailEventStatusEmpty = errors.New("mail event status is empty")
	// ErrMailEventStatusInvalid 表示状态非法
	ErrMailEventStatusInvalid = errors.New("mail event status invalid")
	// ErrMailEventNotFound 表示事件不存在或无操作权限
	ErrMailEventNotFound = errors.New("mail event not found")
)

// MailEventService 负责邮件事件聚合与人工处理
type MailEventService struct {
	eventRepo repository.MailEventRepository
	jobRepo   repository.JobApplicationRepository
}

func NewMailEventService(eventRepo repository.MailEventRepository, jobRepo repository.JobApplicationRepository) *MailEventService {
	return &MailEventService{eventRepo: eventRepo, jobRepo: jobRepo}
}

// ListPendingEvents 返回用户待确认的邮件事件
func (s *MailEventService) ListPendingEvents(userID uint) ([]model.MailEventPendingItem, error) {
	events, err := s.eventRepo.ListPendingByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("查询邮件事件失败: %w", err)
	}

	results := make([]model.MailEventPendingItem, 0, len(events))
	for i := range events {
		item, err := s.toPendingItem(userID, &events[i])
		if err != nil {
			return nil, err
		}
		results = append(results, *item)
	}
	return results, nil
}

// UpdateEventStatus 更新事件状态并返回最新数据
func (s *MailEventService) UpdateEventStatus(userID uint, id int, req *model.MailEventStatusUpdateRequest) (*model.MailEventPendingItem, error) {
	if req == nil {
		return nil, ErrMailEventEmptyRequest
	}
	if req.Status == "" {
		return nil, ErrMailEventStatusEmpty
	}

	allowed := map[string]struct{}{
		model.MailEventStatusPending:         {},
		model.MailEventStatusNeedsReview:     {},
		model.MailEventStatusProcessed:       {},
		model.MailEventStatusProcessingError: {},
		model.MailEventStatusDismissed:       {},
	}
	if _, ok := allowed[req.Status]; !ok {
		return nil, ErrMailEventStatusInvalid
	}

	event, err := s.eventRepo.GetByIDForUser(userID, id)
	if err != nil || event == nil {
		return nil, ErrMailEventNotFound
	}

	targetAppID := event.ApplicationID
	if req.ApplicationID != nil {
		targetAppID = req.ApplicationID
	}

	if err := s.eventRepo.UpdateProcessingResult(id, targetAppID, req.Status, req.ErrorMessage); err != nil {
		return nil, fmt.Errorf("更新事件状态失败: %w", err)
	}

	updated, err := s.eventRepo.GetByIDForUser(userID, id)
	if err != nil || updated == nil {
		return nil, fmt.Errorf("获取更新后的事件失败: %w", err)
	}

	return s.toPendingItem(userID, updated)
}

func (s *MailEventService) toPendingItem(userID uint, event *model.MailEvent) (*model.MailEventPendingItem, error) {
	if event == nil {
		return nil, fmt.Errorf("事件数据为空")
	}

	payload := model.MailEventPayload{}
	if len(event.Payload) > 0 {
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			log.Printf("[MailEventService] payload 解析失败 event=%d: %v", event.ID, err)
		}
	}

	item := &model.MailEventPendingItem{
		ID:             event.ID,
		UserID:         event.UserID,
		MailboxID:      event.MailboxID,
		ApplicationID:  event.ApplicationID,
		Subject:        event.Subject,
		Sender:         event.Sender,
		ReceivedAt:     event.ReceivedAt,
		Snippet:        event.Snippet,
		Classification: event.Classification,
		Confidence:     event.Confidence,
		Payload:        payload,
		Status:         event.Status,
		ErrorMessage:   event.ErrorMessage,
	}

	if event.ApplicationID != nil {
		app, err := s.jobRepo.GetByID(userID, *event.ApplicationID)
		if err != nil {
			log.Printf("[MailEventService] 获取关联岗位失败 event=%d app=%d: %v", event.ID, *event.ApplicationID, err)
		} else if app != nil {
			item.Application = &model.MailEventApplicationSummary{
				ID:              app.ID,
				CompanyName:     app.CompanyName,
				PositionTitle:   app.PositionTitle,
				Status:          app.Status,
				InterviewTime:   app.InterviewTime,
				ReminderTime:    app.ReminderTime,
				ReminderEnabled: app.ReminderEnabled,
			}
		}
	}

	return item, nil
}
