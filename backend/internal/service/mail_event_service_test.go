package service

import (
	"encoding/json"
	"testing"
	"time"

	"jobView-backend/internal/model"
)

type fakeMailEventRepo struct {
	events map[int]*model.MailEvent
}

func newFakeMailEventRepo() *fakeMailEventRepo {
	return &fakeMailEventRepo{events: make(map[int]*model.MailEvent)}
}

func (f *fakeMailEventRepo) Create(event *model.MailEvent) (*model.MailEvent, error) {
	return event, nil
}

func (f *fakeMailEventRepo) ListPendingByUser(userID uint) ([]model.MailEvent, error) {
	result := make([]model.MailEvent, 0, len(f.events))
	for _, evt := range f.events {
		result = append(result, *evt)
	}
	return result, nil
}

func (f *fakeMailEventRepo) MarkStatus(id int, status string, errMsg *string) error {
	if evt, ok := f.events[id]; ok {
		evt.Status = status
		evt.ErrorMessage = errMsg
	}
	return nil
}

func (f *fakeMailEventRepo) UpdateProcessingResult(id int, applicationID *int, status string, errMsg *string) error {
	if evt, ok := f.events[id]; ok {
		evt.Status = status
		evt.ErrorMessage = errMsg
		evt.ApplicationID = applicationID
	}
	return nil
}

func (f *fakeMailEventRepo) GetByIDForUser(userID uint, id int) (*model.MailEvent, error) {
	if evt, ok := f.events[id]; ok {
		return evt, nil
	}
	return nil, nil
}

type fakeJobRepo struct {
	jobs map[int]*model.JobApplication
}

func newFakeJobRepo() *fakeJobRepo {
	return &fakeJobRepo{jobs: make(map[int]*model.JobApplication)}
}

func (f *fakeJobRepo) Create(userID uint, req *model.CreateJobApplicationRequest) (*model.JobApplication, error) {
	return nil, nil
}

func (f *fakeJobRepo) GetByID(userID uint, id int) (*model.JobApplication, error) {
	if job, ok := f.jobs[id]; ok {
		return job, nil
	}
	return nil, nil
}

func (f *fakeJobRepo) GetAllPaginated(userID uint, req model.PaginationRequest) (*model.PaginationResponse, error) {
	return nil, nil
}

func (f *fakeJobRepo) GetAll(userID uint) ([]model.JobApplication, error) {
	return nil, nil
}

func (f *fakeJobRepo) Update(userID uint, id int, req *model.UpdateJobApplicationRequest) (*model.JobApplication, error) {
	return nil, nil
}

func (f *fakeJobRepo) Delete(userID uint, id int) error                        { return nil }
func (f *fakeJobRepo) GetStatusStatistics(userID uint) (map[string]int, error) { return nil, nil }
func (f *fakeJobRepo) GetHRPassCount(userID uint) (int, error)                 { return 0, nil }
func (f *fakeJobRepo) BatchCreate(userID uint, applications []model.CreateJobApplicationRequest) ([]model.JobApplication, error) {
	return nil, nil
}
func (f *fakeJobRepo) BatchUpdateStatus(userID uint, updates []model.BatchStatusUpdate) error {
	return nil
}
func (f *fakeJobRepo) BatchDelete(userID uint, ids []int) error { return nil }
func (f *fakeJobRepo) Search(userID uint, searchQuery string, req model.PaginationRequest) (*model.PaginationResponse, error) {
	return nil, nil
}
func (f *fakeJobRepo) ListByDateRange(userID uint, startDate, endDate string, req model.PaginationRequest) (*model.PaginationResponse, error) {
	return nil, nil
}
func (f *fakeJobRepo) ListWithStatusFilters(userID uint, status *model.ApplicationStatus, stageStatuses []string, req model.PaginationRequest) (*model.PaginationResponse, error) {
	return nil, nil
}
func (f *fakeJobRepo) ListRecentApplications(userID uint, limit int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (f *fakeJobRepo) ListUpcomingInterviews(userID uint, limit int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (f *fakeJobRepo) ListDailyStats(userID uint, days int) ([]map[string]interface{}, error) {
	return nil, nil
}

func TestMailEventService_ListPendingEvents(t *testing.T) {
	eventRepo := newFakeMailEventRepo()
	jobRepo := newFakeJobRepo()
	service := NewMailEventService(eventRepo, jobRepo)

	payload := model.MailEventPayload{}
	link := "https://example.com/exam"
	payload.ExamLink = &link
	payloadBytes, _ := json.Marshal(payload)

	jobRepo.jobs[101] = &model.JobApplication{
		ID:              101,
		CompanyName:     "示例公司",
		PositionTitle:   "后端工程师",
		Status:          model.StatusApplied,
		InterviewTime:   nil,
		ReminderTime:    nil,
		ReminderEnabled: true,
	}

	eventRepo.events[1] = &model.MailEvent{
		ID:             1,
		UserID:         1,
		MailboxID:      10,
		ApplicationID:  func() *int { i := 101; return &i }(),
		Subject:        "笔试通知",
		Sender:         "hr@example.com",
		ReceivedAt:     time.Now(),
		Classification: model.MailClassificationExam,
		Confidence:     0.9,
		Payload:        payloadBytes,
		Status:         model.MailEventStatusNeedsReview,
	}

	items, err := service.ListPendingEvents(1)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("期望返回1条事件，实际为 %d", len(items))
	}
	if items[0].Payload.ExamLink == nil || *items[0].Payload.ExamLink != link {
		t.Fatalf("期望解析出笔试链接，实际为 %+v", items[0].Payload)
	}
	if items[0].Application == nil || items[0].Application.CompanyName != "示例公司" {
		t.Fatalf("期望带回岗位摘要，实际为 %+v", items[0].Application)
	}
}

func TestMailEventService_UpdateEventStatusValidation(t *testing.T) {
	service := NewMailEventService(newFakeMailEventRepo(), newFakeJobRepo())

	if _, err := service.UpdateEventStatus(1, 1, nil); err != ErrMailEventEmptyRequest {
		t.Fatalf("期望 ErrMailEventEmptyRequest，实际得到 %v", err)
	}

	if _, err := service.UpdateEventStatus(1, 1, &model.MailEventStatusUpdateRequest{}); err != ErrMailEventStatusEmpty {
		t.Fatalf("期望 ErrMailEventStatusEmpty，实际得到 %v", err)
	}

	req := &model.MailEventStatusUpdateRequest{Status: "unknown"}
	if _, err := service.UpdateEventStatus(1, 1, req); err != ErrMailEventStatusInvalid {
		t.Fatalf("期望 ErrMailEventStatusInvalid，实际得到 %v", err)
	}
}

func TestMailEventService_UpdateEventStatusSuccess(t *testing.T) {
	eventRepo := newFakeMailEventRepo()
	jobRepo := newFakeJobRepo()
	service := NewMailEventService(eventRepo, jobRepo)

	eventRepo.events[1] = &model.MailEvent{
		ID:             1,
		UserID:         1,
		MailboxID:      10,
		Subject:        "面试通知",
		Sender:         "hr@example.com",
		ReceivedAt:     time.Now(),
		Status:         model.MailEventStatusNeedsReview,
		Classification: model.MailClassificationInterview,
	}

	req := &model.MailEventStatusUpdateRequest{Status: model.MailEventStatusProcessed}
	item, err := service.UpdateEventStatus(1, 1, req)
	if err != nil {
		t.Fatalf("期望更新成功，实际得到 %v", err)
	}
	if item.Status != model.MailEventStatusProcessed {
		t.Fatalf("期望状态已更新为 processed，实际 %s", item.Status)
	}
}
