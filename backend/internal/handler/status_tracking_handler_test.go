package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"jobView-backend/internal/auth"
	"jobView-backend/internal/model"
	"jobView-backend/internal/repository"
	"jobView-backend/internal/service"
)

type stubInterviewRepo struct {
	experiences []model.InterviewExperience
	err         error
}

func (s *stubInterviewRepo) GetStatusHistory(ctx context.Context, userID uint, jobApplicationID, page, pageSize int) (int, []model.StatusHistoryEntry, error) {
	panic("not implemented")
}

func (s *stubInterviewRepo) GetStatusTimeline(ctx context.Context, userID uint, jobApplicationID int) ([]model.StatusHistoryEntry, error) {
	panic("not implemented")
}

func (s *stubInterviewRepo) GetInterviewExperiences(ctx context.Context, userID uint, jobApplicationID int) ([]model.InterviewExperience, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.experiences, nil
}

func (s *stubInterviewRepo) BeginTx(ctx context.Context) (repository.StatusTrackingTx, error) {
	panic("not implemented")
}

func (s *stubInterviewRepo) GetStatusAnalytics(ctx context.Context, userID uint) (*model.StatusAnalyticsResponse, error) {
	panic("not implemented")
}

func (s *stubInterviewRepo) GetStatusTrends(ctx context.Context, userID uint, days int) ([]model.StatusTrend, error) {
	panic("not implemented")
}

type validationErrorRepo struct {
	tx repository.StatusTrackingTx
}

func (r *validationErrorRepo) GetStatusHistory(ctx context.Context, userID uint, jobApplicationID, page, pageSize int) (int, []model.StatusHistoryEntry, error) {
	panic("not implemented")
}

func (r *validationErrorRepo) GetStatusTimeline(ctx context.Context, userID uint, jobApplicationID int) ([]model.StatusHistoryEntry, error) {
	panic("not implemented")
}

func (r *validationErrorRepo) GetInterviewExperiences(ctx context.Context, userID uint, jobApplicationID int) ([]model.InterviewExperience, error) {
	panic("not implemented")
}

func (r *validationErrorRepo) BeginTx(ctx context.Context) (repository.StatusTrackingTx, error) {
	return r.tx, nil
}

func (r *validationErrorRepo) GetStatusAnalytics(ctx context.Context, userID uint) (*model.StatusAnalyticsResponse, error) {
	panic("not implemented")
}

func (r *validationErrorRepo) GetStatusTrends(ctx context.Context, userID uint, days int) ([]model.StatusTrend, error) {
	panic("not implemented")
}

type validationErrorTx struct{}

func (t *validationErrorTx) GetJobApplicationForUpdate(userID uint, jobApplicationID int) (*repository.JobStatusSnapshot, error) {
	now := time.Now()
	version := 1
	return &repository.JobStatusSnapshot{
		Job: model.JobApplication{
			ID:     jobApplicationID,
			UserID: userID,
			Status: model.StatusFirstInterview,
		},
		StatusVersion:    &version,
		LastStatusChange: &now,
	}, nil
}

func (t *validationErrorTx) SetLocalFlag(flag string, enabled bool) error {
	return nil
}

func (t *validationErrorTx) InsertStatusHistory(entry repository.StatusHistoryInsert) (int64, error) {
	panic("not implemented")
}

func (t *validationErrorTx) InsertInterviewExperience(entry repository.InterviewExperienceInsert) (int64, error) {
	panic("not implemented")
}

func (t *validationErrorTx) UpdateJobApplication(params repository.UpdateJobApplicationParams) (*model.JobApplication, error) {
	panic("not implemented")
}

func (t *validationErrorTx) GetCurrentStatus(userID uint, jobApplicationID int) (model.ApplicationStatus, *time.Time, error) {
	panic("not implemented")
}

func (t *validationErrorTx) GetLatestHistoryEntry(userID uint, jobApplicationID int) (*model.StatusHistoryEntry, error) {
	panic("not implemented")
}

func (t *validationErrorTx) UpdateHistoryMetadata(historyID int64, metadata []byte) error {
	panic("not implemented")
}

func (t *validationErrorTx) Commit() error {
	return nil
}

func (t *validationErrorTx) Rollback() error {
	return nil
}

func TestStatusTrackingHandler_GetInterviewExperiences_Success(t *testing.T) {
	repo := &stubInterviewRepo{
		experiences: []model.InterviewExperience{
			{
				ID:            1,
				ApplicationID: 501,
				FromStatus:    model.StatusFirstInterview,
				ToStatus:      model.StatusSecondInterview,
				Rating:        strPtr("good"),
				Note:          strPtr("顺利通过"),
				Skip:          false,
				RecordedBy:    99,
				RecordedAt:    time.Now(),
				CreatedAt:     time.Now(),
			},
		},
	}
	svc := service.NewStatusTrackingService(repo, nil)
	handler := NewStatusTrackingHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/applications/501/interview-experiences", nil)
	ctx := context.WithValue(req.Context(), auth.UserIDContextKey, uint(99))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": "501"})

	rr := httptest.NewRecorder()
	handler.GetInterviewExperiences(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Code    int                         `json:"code"`
		Message string                      `json:"message"`
		Data    []model.InterviewExperience `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, 200, resp.Code)
	require.Equal(t, "interview experiences retrieved successfully", resp.Message)
	require.Len(t, resp.Data, 1)
	require.Equal(t, "good", *resp.Data[0].Rating)
	require.Equal(t, false, resp.Data[0].Skip)
}

func TestStatusTrackingHandler_GetInterviewExperiences_NotFound(t *testing.T) {
	repo := &stubInterviewRepo{
		err: fmt.Errorf("job application not found or access denied"),
	}
	svc := service.NewStatusTrackingService(repo, nil)
	handler := NewStatusTrackingHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/applications/999/interview-experiences", nil)
	ctx := context.WithValue(req.Context(), auth.UserIDContextKey, uint(42))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})

	rr := httptest.NewRecorder()
	handler.GetInterviewExperiences(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, 404, resp.Code)
	require.Equal(t, "job application not found", resp.Message)
}

func TestStatusTrackingHandler_UpdateJobStatus_MissingInterviewRating(t *testing.T) {
	repo := &validationErrorRepo{tx: &validationErrorTx{}}
	svc := service.NewStatusTrackingService(repo, nil)
	handler := NewStatusTrackingHandler(svc)

	payload := map[string]interface{}{
		"status": model.StatusSecondInterview,
		"metadata": map[string]interface{}{
			"interview_experience": map[string]interface{}{
				"skip": false,
			},
		},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/job-applications/123/status", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "123"})
	ctx := context.WithValue(req.Context(), auth.UserIDContextKey, uint(88))
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UpdateJobStatus(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, 400, resp.Code)
	require.Equal(t, "缺失评价", resp.Message)
}

func strPtr(v string) *string {
	return &v
}
