package service

import (
    "encoding/json"
    "regexp"
    "testing"
    "time"

    sqlmock "github.com/DATA-DOG/go-sqlmock"
    "github.com/stretchr/testify/require"

    "jobView-backend/internal/database"
    "jobView-backend/internal/model"
    "jobView-backend/internal/repository"
)

func TestMailEventServiceIntegrationWithRepository(t *testing.T) {
    mockDB, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer mockDB.Close()

    db := &database.DB{DB: mockDB}
    eventRepo := repository.NewMailEventRepository(db)
    jobRepo := newFakeJobRepo()

    service := NewMailEventService(eventRepo, jobRepo)

    payload := model.MailEventPayload{ExamLink: testPtrString("https://exam.example.com")}
    payloadBytes, _ := json.Marshal(payload)

    eventTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

    jobRepo.jobs[101] = &model.JobApplication{
        ID:              101,
        UserID:          1,
        CompanyName:     "示例公司",
        PositionTitle:   "后端实习生",
        Status:          model.StatusApplied,
        ReminderEnabled: false,
    }

    selectPendingQuery := regexp.QuoteMeta("SELECT id, user_id, mailbox_id, application_id, message_id, message_uid, subject, sender, received_at, snippet,\n\t\tclassification, confidence, payload, status, error_message, created_at, updated_at\n\tFROM mail_events WHERE user_id = $1 AND status IN ('pending', 'needs_review') ORDER BY received_at DESC")
    mock.ExpectQuery(selectPendingQuery).
        WithArgs(1).
        WillReturnRows(sqlmock.NewRows([]string{
            "id", "user_id", "mailbox_id", "application_id", "message_id", "message_uid", "subject", "sender", "received_at", "snippet",
            "classification", "confidence", "payload", "status", "error_message", "created_at", "updated_at",
        }).
            AddRow(1, 1, 9, 101, nil, nil, "笔试通知", "hr@example.com", eventTime, nil, model.MailClassificationExam, 0.9, payloadBytes, model.MailEventStatusNeedsReview, nil, eventTime, eventTime))

    items, err := service.ListPendingEvents(1)
    require.NoError(t, err)
    require.Len(t, items, 1)
    require.Equal(t, "笔试通知", items[0].Subject)
    require.NotNil(t, items[0].Application)
    require.Equal(t, "示例公司", items[0].Application.CompanyName)
    require.Equal(t, "https://exam.example.com", ptrValue(items[0].Payload.ExamLink))

    selectEventQuery := regexp.QuoteMeta("SELECT id, user_id, mailbox_id, application_id, message_id, message_uid, subject, sender, received_at, snippet,\n\t\tclassification, confidence, payload, status, error_message, created_at, updated_at\n\tFROM mail_events WHERE id = $1 AND user_id = $2")
    mock.ExpectQuery(selectEventQuery).
        WithArgs(1, 1).
        WillReturnRows(sqlmock.NewRows([]string{
            "id", "user_id", "mailbox_id", "application_id", "message_id", "message_uid", "subject", "sender", "received_at", "snippet",
            "classification", "confidence", "payload", "status", "error_message", "created_at", "updated_at",
        }).AddRow(1, 1, 9, 101, nil, nil, "笔试通知", "hr@example.com", eventTime, nil, model.MailClassificationExam, 0.9, payloadBytes, model.MailEventStatusNeedsReview, nil, eventTime, eventTime))

    mock.ExpectExec(regexp.QuoteMeta("UPDATE mail_events SET application_id = $1, status = $2, error_message = $3, updated_at = NOW() WHERE id = $4")).
        WithArgs(101, model.MailEventStatusProcessed, sqlmock.AnyArg(), 1).
        WillReturnResult(sqlmock.NewResult(0, 1))

    mock.ExpectQuery(selectEventQuery).
        WithArgs(1, 1).
        WillReturnRows(sqlmock.NewRows([]string{
            "id", "user_id", "mailbox_id", "application_id", "message_id", "message_uid", "subject", "sender", "received_at", "snippet",
            "classification", "confidence", "payload", "status", "error_message", "created_at", "updated_at",
        }).AddRow(1, 1, 9, 101, nil, nil, "笔试通知", "hr@example.com", eventTime, nil, model.MailClassificationExam, 0.9, payloadBytes, model.MailEventStatusProcessed, nil, eventTime, eventTime))

    updated, err := service.UpdateEventStatus(1, 1, &model.MailEventStatusUpdateRequest{Status: model.MailEventStatusProcessed})
    require.NoError(t, err)
    require.Equal(t, model.MailEventStatusProcessed, updated.Status)

    require.NoError(t, mock.ExpectationsWereMet())
}

func ptrValue(v *string) string {
    if v == nil {
        return ""
    }
    return *v
}

func testPtrString(v string) *string { return &v }
