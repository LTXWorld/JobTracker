package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"jobView-backend/internal/model"
	"jobView-backend/internal/repository"
)

// EmailEventProcessor 将邮件事件与投递记录关联并触发自动更新
type EmailEventProcessor struct {
	jobRepo      repository.JobApplicationRepository
	eventRepo    repository.MailEventRepository
	statusOffset time.Duration
}

// NewEmailEventProcessor 初始化
func NewEmailEventProcessor(jobRepo repository.JobApplicationRepository, eventRepo repository.MailEventRepository) *EmailEventProcessor {
	return &EmailEventProcessor{
		jobRepo:      jobRepo,
		eventRepo:    eventRepo,
		statusOffset: 30 * time.Minute,
	}
}

// Process 尝试自动处理邮件事件
func (p *EmailEventProcessor) Process(event *model.MailEvent) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}
	var payload model.MailEventPayload
	if len(event.Payload) > 0 {
		_ = json.Unmarshal(event.Payload, &payload)
	}

	match, score := p.matchApplication(event.UserID, &payload)
	if match == nil || score < 0.4 {
		return p.eventRepo.UpdateProcessingResult(event.ID, nil, model.MailEventStatusNeedsReview, ptrString("未找到高置信度的关联投递记录"))
	}

	switch event.Classification {
	case model.MailClassificationExam:
		return p.handleExam(event, match, &payload)
	case model.MailClassificationInterview:
		return p.handleInterview(event, match, &payload)
	default:
		return p.eventRepo.UpdateProcessingResult(event.ID, ptrInt(match.ID), model.MailEventStatusNeedsReview, ptrString("暂不支持的邮件类型"))
	}
}

func (p *EmailEventProcessor) handleExam(event *model.MailEvent, app *model.JobApplication, payload *model.MailEventPayload) error {
	if payload.DetectedTime == nil {
		return p.eventRepo.UpdateProcessingResult(event.ID, ptrInt(app.ID), model.MailEventStatusNeedsReview, ptrString("笔试时间解析失败"))
	}
	eventTime := payload.DetectedTime.In(time.Local)
	reminderTime := eventTime.Add(-p.statusOffset)
	status := model.StatusWrittenTest
	category := "written"
	interviewType := "笔试"
	enabled := true

	updateReq := &model.UpdateJobApplicationRequest{
		Status:           &status,
		InterviewTime:    &eventTime,
		ReminderTime:     &reminderTime,
		ReminderEnabled:  &enabled,
		ReminderCategory: &category,
		InterviewType:    &interviewType,
	}

	if payload.ExamLink != nil && *payload.ExamLink != "" {
		notes := mergeNotes(app.Notes, fmt.Sprintf("【自动识别】笔试链接：%s", *payload.ExamLink))
		updateReq.Notes = &notes
	}

	if _, err := p.jobRepo.Update(event.UserID, app.ID, updateReq); err != nil {
		errMsg := fmt.Sprintf("更新笔试信息失败: %v", err)
		return p.eventRepo.UpdateProcessingResult(event.ID, ptrInt(app.ID), model.MailEventStatusProcessingError, &errMsg)
	}

	return p.eventRepo.UpdateProcessingResult(event.ID, ptrInt(app.ID), model.MailEventStatusProcessed, nil)
}

func (p *EmailEventProcessor) handleInterview(event *model.MailEvent, app *model.JobApplication, payload *model.MailEventPayload) error {
	if payload.DetectedTime == nil {
		return p.eventRepo.UpdateProcessingResult(event.ID, ptrInt(app.ID), model.MailEventStatusNeedsReview, ptrString("面试时间解析失败"))
	}
	eventTime := payload.DetectedTime.In(time.Local)
	reminderTime := eventTime.Add(-p.statusOffset)
	status := nextInterviewStatus(app.Status)
	category := "interview"
	interviewType := "面试"
	enabled := true

	updateReq := &model.UpdateJobApplicationRequest{
		Status:           &status,
		InterviewTime:    &eventTime,
		ReminderTime:     &reminderTime,
		ReminderEnabled:  &enabled,
		ReminderCategory: &category,
		InterviewType:    &interviewType,
	}

	var notesToAdd []string
	if payload.MeetingLink != nil && *payload.MeetingLink != "" {
		notesToAdd = append(notesToAdd, "会议链接："+*payload.MeetingLink)
	}
	if payload.MeetingID != nil && *payload.MeetingID != "" {
		notesToAdd = append(notesToAdd, "会议号："+*payload.MeetingID)
	}
	if len(notesToAdd) > 0 {
		notes := mergeNotes(app.Notes, "【自动识别】"+strings.Join(notesToAdd, "，"))
		updateReq.Notes = &notes
	}

	if _, err := p.jobRepo.Update(event.UserID, app.ID, updateReq); err != nil {
		errMsg := fmt.Sprintf("更新面试信息失败: %v", err)
		return p.eventRepo.UpdateProcessingResult(event.ID, ptrInt(app.ID), model.MailEventStatusProcessingError, &errMsg)
	}

	return p.eventRepo.UpdateProcessingResult(event.ID, ptrInt(app.ID), model.MailEventStatusProcessed, nil)
}

func (p *EmailEventProcessor) matchApplication(userID uint, payload *model.MailEventPayload) (*model.JobApplication, float64) {
	apps, err := p.jobRepo.GetAll(userID)
	if err != nil {
		return nil, 0
	}
	var best *model.JobApplication
	bestScore := 0.0
	for i := range apps {
		score := similarityScore(&apps[i], payload)
		if score > bestScore {
			best = &apps[i]
			bestScore = score
		}
	}
	return best, bestScore
}

func similarityScore(app *model.JobApplication, payload *model.MailEventPayload) float64 {
	if app == nil {
		return 0
	}
	score := 0.0
	name := strings.ToLower(app.CompanyName)
	position := strings.ToLower(app.PositionTitle)

	for _, candidate := range payload.CompanyCandidates {
		candidate = strings.ToLower(candidate)
		if candidate == "" {
			continue
		}
		if strings.Contains(name, candidate) || strings.Contains(candidate, name) {
			score += 0.4
		}
	}

	for _, candidate := range payload.PositionCandidates {
		candidate = strings.ToLower(candidate)
		if candidate == "" {
			continue
		}
		if strings.Contains(position, candidate) || strings.Contains(candidate, position) {
			score += 0.4
		}
	}

	return score
}

func mergeNotes(existing *string, addition string) string {
	if existing == nil || strings.TrimSpace(*existing) == "" {
		return addition
	}
	if strings.Contains(*existing, addition) {
		return *existing
	}
	return strings.TrimSpace(*existing) + "\n" + addition
}

func ptrInt(v int) *int { return &v }

func ptrString(v string) *string { return &v }

func nextInterviewStatus(current model.ApplicationStatus) model.ApplicationStatus {
	switch current {
	case model.StatusFirstInterview, model.StatusFirstPass:
		return model.StatusSecondInterview
	case model.StatusSecondInterview, model.StatusSecondPass:
		return model.StatusThirdInterview
	case model.StatusThirdInterview, model.StatusThirdPass:
		return model.StatusHRInterview
	default:
		return model.StatusFirstInterview
	}
}
