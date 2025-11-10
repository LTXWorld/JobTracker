package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"jobView-backend/internal/model"
	"jobView-backend/internal/repository"
	"jobView-backend/internal/utils"
)

type fakeTrackingRepo struct {
	tx   repository.StatusTrackingTx
	err  error
	used bool
}

func (f *fakeTrackingRepo) GetStatusHistory(ctx context.Context, userID uint, jobApplicationID, page, pageSize int) (int, []model.StatusHistoryEntry, error) {
	panic("not implemented")
}

func (f *fakeTrackingRepo) GetStatusTimeline(ctx context.Context, userID uint, jobApplicationID int) ([]model.StatusHistoryEntry, error) {
	panic("not implemented")
}

func (f *fakeTrackingRepo) GetInterviewExperiences(ctx context.Context, userID uint, jobApplicationID int) ([]model.InterviewExperience, error) {
	panic("not implemented")
}

func (f *fakeTrackingRepo) BeginTx(ctx context.Context) (repository.StatusTrackingTx, error) {
	f.used = true
	return f.tx, f.err
}

func (f *fakeTrackingRepo) GetStatusAnalytics(ctx context.Context, userID uint, processType *model.ApplicationProcessType) (*model.StatusAnalyticsResponse, error) {
	panic("not implemented")
}

func (f *fakeTrackingRepo) GetStatusTrends(ctx context.Context, userID uint, days int, processType *model.ApplicationProcessType) ([]model.StatusTrend, error) {
	panic("not implemented")
}

type fakeConfigRepo struct {
	flowID  int
	flowCfg string
	err     error
}

func (f *fakeConfigRepo) GetFlowTemplates(userID uint) ([]model.StatusFlowTemplate, error) {
	panic("not implemented")
}

func (f *fakeConfigRepo) CheckTemplateNameExists(name string, excludeID *int) (bool, error) {
	panic("not implemented")
}

func (f *fakeConfigRepo) GetTemplatePermissions(templateID int) (sql.NullInt64, bool, error) {
	panic("not implemented")
}

func (f *fakeConfigRepo) CreateFlowTemplate(userID uint, name string, desc *string, flowConfigBytes []byte) (*model.StatusFlowTemplate, error) {
	panic("not implemented")
}

func (f *fakeConfigRepo) UpdateFlowTemplate(userID uint, templateID int, name string, desc *string, flowConfigBytes []byte) (*model.StatusFlowTemplate, error) {
	panic("not implemented")
}

func (f *fakeConfigRepo) DeleteFlowTemplate(userID uint, templateID int) error {
	panic("not implemented")
}

func (f *fakeConfigRepo) GetDefaultFlowTemplate() (int, string, error) {
	return f.flowID, f.flowCfg, f.err
}

func (f *fakeConfigRepo) UpdateFlowConfigByID(id int, flowConfigText string) error {
	panic("not implemented")
}

func (f *fakeConfigRepo) GetPreferences(userID uint) (*model.UserStatusPreferences, error) {
	panic("not implemented")
}

func (f *fakeConfigRepo) UpsertPreferences(userID uint, preferenceBytes []byte, now time.Time) (*model.UserStatusPreferences, error) {
	panic("not implemented")
}

type fakeStatusTx struct {
	snapshot              *repository.JobStatusSnapshot
	historyInserts        []repository.StatusHistoryInsert
	interviewInserts      []repository.InterviewExperienceInsert
	updateCalls           []repository.UpdateJobApplicationParams
	commitCalled          bool
	rollbackCalled        bool
	flags                 map[string]bool
	currentStatusResponse map[int]struct {
		status    model.ApplicationStatus
		changedAt *time.Time
		err       error
	}
	latestHistoryEntry *model.StatusHistoryEntry
	latestHistoryErr   error
	metadataUpdates    map[int64][]byte
}

func newFakeStatusTx() *fakeStatusTx {
	return &fakeStatusTx{
		flags: make(map[string]bool),
		currentStatusResponse: make(map[int]struct {
			status    model.ApplicationStatus
			changedAt *time.Time
			err       error
		}),
		metadataUpdates: make(map[int64][]byte),
	}
}

func (f *fakeStatusTx) GetJobApplicationForUpdate(userID uint, jobApplicationID int) (*repository.JobStatusSnapshot, error) {
	if f.snapshot == nil {
		return nil, sql.ErrNoRows
	}
	return f.snapshot, nil
}

func (f *fakeStatusTx) SetLocalFlag(flag string, enabled bool) error {
	f.flags[flag] = enabled
	return nil
}

func (f *fakeStatusTx) InsertStatusHistory(entry repository.StatusHistoryInsert) (int64, error) {
	f.historyInserts = append(f.historyInserts, entry)
	return int64(len(f.historyInserts)), nil
}

func (f *fakeStatusTx) InsertInterviewExperience(entry repository.InterviewExperienceInsert) (int64, error) {
	f.interviewInserts = append(f.interviewInserts, entry)
	return int64(len(f.interviewInserts)), nil
}

func (f *fakeStatusTx) UpdateJobApplication(params repository.UpdateJobApplicationParams) (*model.JobApplication, error) {
	f.updateCalls = append(f.updateCalls, params)
	job := &model.JobApplication{ID: params.JobApplicationID, UserID: params.UserID, Status: params.NewStatus}
	return job, nil
}

func (f *fakeStatusTx) GetCurrentStatus(userID uint, jobApplicationID int) (model.ApplicationStatus, *time.Time, error) {
	resp, ok := f.currentStatusResponse[jobApplicationID]
	if !ok {
		return "", nil, sql.ErrNoRows
	}
	return resp.status, resp.changedAt, resp.err
}

func (f *fakeStatusTx) GetLatestHistoryEntry(userID uint, jobApplicationID int) (*model.StatusHistoryEntry, error) {
	if f.latestHistoryErr != nil {
		return nil, f.latestHistoryErr
	}
	if f.latestHistoryEntry == nil {
		return nil, sql.ErrNoRows
	}
	return f.latestHistoryEntry, nil
}

func (f *fakeStatusTx) UpdateHistoryMetadata(historyID int64, metadata []byte) error {
	f.metadataUpdates[historyID] = metadata
	return nil
}

func (f *fakeStatusTx) Commit() error {
	f.commitCalled = true
	return nil
}

func (f *fakeStatusTx) Rollback() error {
	f.rollbackCalled = true
	return nil
}

type fakeServiceRepo struct {
	repo repository.StatusTrackingRepository
}

func TestStatusTrackingService_UpdateJobStatus_ForwardTransition(t *testing.T) {
	tx := newFakeStatusTx()
	now := time.Now().Add(-2 * time.Hour)
	historyJSON := `{"entries":[]}`
	durationJSON := `{"durations":{}}`
	tx.snapshot = &repository.JobStatusSnapshot{
		Job: model.JobApplication{
			ID:     101,
			UserID: 33,
			Status: model.StatusResumeScreening,
		},
		StatusVersion:    intPtr(5),
		StatusHistoryRaw: &historyJSON,
		DurationStatsRaw: &durationJSON,
		LastStatusChange: &now,
	}

	repo := &fakeTrackingRepo{tx: tx}
	cfg := &fakeConfigRepo{
		flowID:  1,
		flowCfg: `{"transitions": {"` + string(model.StatusResumeScreening) + `": ["` + string(model.StatusWrittenTest) + `"]}}`,
	}

	service := &StatusTrackingService{repo: repo, configRepo: cfg}

	result, err := service.UpdateJobStatus(33, 101, &model.StatusUpdateRequest{
		Status:  model.StatusWrittenTest,
		Version: intPtr(5),
		Metadata: map[string]interface{}{
			"note": "准备笔试",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Job)
	require.Equal(t, model.StatusWrittenTest, result.Job.Status)
	require.NotNil(t, result.HistoryID)
	require.NotNil(t, result.StatusVersion)
	require.NotNil(t, result.UndoAvailableUntil)
	require.True(t, tx.commitCalled)
	require.True(t, tx.rollbackCalled)
	require.Contains(t, tx.flags, "jobview.skip_history")
	require.False(t, tx.flags["jobview.allow_backward"])
	require.Len(t, tx.historyInserts, 1)
	require.Equal(t, model.StatusResumeScreening, tx.historyInserts[0].OldStatus)
	require.Equal(t, model.StatusWrittenTest, tx.historyInserts[0].NewStatus)
	require.Len(t, tx.updateCalls, 1)
	require.NotNil(t, tx.updateCalls[0].StatusHistoryJSON)
	require.NotNil(t, tx.updateCalls[0].DurationStatsJSON)
	require.NotNil(t, tx.updateCalls[0].StatusVersion)
	require.False(t, tx.updateCalls[0].SuppressHistory)
}

func TestStatusTrackingService_UpdateJobStatus_CapturesInterviewExperience(t *testing.T) {
	tx := newFakeStatusTx()
	lastChange := time.Now().Add(-15 * time.Minute)
	historyJSON := `{"history":[],"metadata":{"total_changes":0}}`
	durationJSON := `{"status_durations":{}}`
	tx.snapshot = &repository.JobStatusSnapshot{
		Job: model.JobApplication{
			ID:     201,
			UserID: 18,
			Status: model.StatusFirstInterview,
		},
		StatusHistoryRaw: &historyJSON,
		DurationStatsRaw: &durationJSON,
		LastStatusChange: &lastChange,
	}

	repo := &fakeTrackingRepo{tx: tx}
	cfg := &fakeConfigRepo{
		flowID:  1,
		flowCfg: `{"transitions": {"` + string(model.StatusFirstInterview) + `": ["` + string(model.StatusSecondInterview) + `"]}}`,
	}

	service := &StatusTrackingService{repo: repo, configRepo: cfg}
	recordedAt := time.Now().Add(-5 * time.Minute).UTC().Format(time.RFC3339)

	result, err := service.UpdateJobStatus(18, 201, &model.StatusUpdateRequest{
		Status: model.StatusSecondInterview,
		Metadata: map[string]interface{}{
			"interview_experience": map[string]interface{}{
				"rating":      "GOOD",
				"note":        " 表现不错 ",
				"recorded_at": recordedAt,
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, tx.interviewInserts, 1)
	insert := tx.interviewInserts[0]
	require.Equal(t, 201, insert.ApplicationID)
	require.EqualValues(t, 18, insert.UserID)
	require.False(t, insert.Skip)
	require.NotNil(t, insert.Rating)
	require.Equal(t, "good", *insert.Rating)
	require.NotNil(t, insert.Note)
	require.Equal(t, "表现不错", *insert.Note)
	parsedRecordedAt, _ := time.Parse(time.RFC3339, recordedAt)
	require.WithinDuration(t, parsedRecordedAt, insert.RecordedAt, time.Second)
	require.Nil(t, insert.SkipReason)

	require.Len(t, tx.historyInserts, 1)
	var stored map[string]interface{}
	require.NoError(t, json.Unmarshal(tx.historyInserts[0].Metadata, &stored))
	meta, ok := stored["interview_experience"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "good", meta["rating"])
	require.Equal(t, false, meta["skip"])
	require.Equal(t, "表现不错", meta["note"])
	require.Equal(t, float64(18), meta["recorded_by"])
	require.Equal(t, string(model.StatusFirstInterview), meta["from_status"])
	require.Equal(t, string(model.StatusSecondInterview), meta["to_status"])
	require.Equal(t, recordedAt, meta["recorded_at"])

	require.Len(t, tx.updateCalls, 1)
	var updatedHistory model.StatusHistory
	require.NoError(t, json.Unmarshal(tx.updateCalls[0].StatusHistoryJSON, &updatedHistory))
	require.NotEmpty(t, updatedHistory.History)
	lastEntry := updatedHistory.History[len(updatedHistory.History)-1]
	require.NotNil(t, lastEntry.Metadata)
	hMeta, ok := lastEntry.Metadata["interview_experience"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "good", hMeta["rating"])
}

func TestStatusTrackingService_UpdateJobStatus_DefaultSkipInterviewExperience(t *testing.T) {
	tx := newFakeStatusTx()
	lastChange := time.Now().Add(-20 * time.Minute)
	tx.snapshot = &repository.JobStatusSnapshot{
		Job: model.JobApplication{
			ID:     302,
			UserID: 27,
			Status: model.StatusSecondInterview,
		},
		LastStatusChange: &lastChange,
	}

	repo := &fakeTrackingRepo{tx: tx}
	cfg := &fakeConfigRepo{
		flowID:  1,
		flowCfg: `{"transitions": {"` + string(model.StatusSecondInterview) + `": ["` + string(model.StatusThirdInterview) + `"]}}`,
	}

	service := &StatusTrackingService{repo: repo, configRepo: cfg}

	result, err := service.UpdateJobStatus(27, 302, &model.StatusUpdateRequest{
		Status: model.StatusThirdInterview,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, tx.interviewInserts, 1)
	insert := tx.interviewInserts[0]
	require.True(t, insert.Skip)
	require.Nil(t, insert.Rating)
	require.Nil(t, insert.Note)
	require.Nil(t, insert.SkipReason)

	var stored map[string]interface{}
	require.NoError(t, json.Unmarshal(tx.historyInserts[0].Metadata, &stored))
	meta, ok := stored["interview_experience"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, true, meta["skip"])
	require.NotEmpty(t, meta["recorded_at"])
}

func TestStatusTrackingService_UpdateJobStatus_MissingInterviewRating(t *testing.T) {
	tx := newFakeStatusTx()
	lastChange := time.Now().Add(-25 * time.Minute)
	tx.snapshot = &repository.JobStatusSnapshot{
		Job: model.JobApplication{
			ID:     410,
			UserID: 35,
			Status: model.StatusFirstInterview,
		},
		LastStatusChange: &lastChange,
	}

	repo := &fakeTrackingRepo{tx: tx}
	cfg := &fakeConfigRepo{
		flowID:  1,
		flowCfg: `{"transitions": {"` + string(model.StatusFirstInterview) + `": ["` + string(model.StatusSecondInterview) + `"]}}`,
	}

	service := &StatusTrackingService{repo: repo, configRepo: cfg}

	_, err := service.UpdateJobStatus(35, 410, &model.StatusUpdateRequest{
		Status: model.StatusSecondInterview,
		Metadata: map[string]interface{}{
			"interview_experience": map[string]interface{}{
				"note": "缺少评分",
			},
		},
	})
	require.Error(t, err)
	var validationErr utils.ValidationError
	require.True(t, errors.As(err, &validationErr))
	require.Contains(t, validationErr.Error(), "缺失评价")
	require.Empty(t, tx.interviewInserts)
	require.False(t, tx.commitCalled)
}

func TestStatusTrackingService_UpdateJobStatus_BackwardRequiresConfirm(t *testing.T) {
	tx := newFakeStatusTx()
	now := time.Now()
	tx.snapshot = &repository.JobStatusSnapshot{
		Job: model.JobApplication{
			ID:     45,
			UserID: 9,
			Status: model.StatusSecondInterview,
		},
		StatusVersion:    intPtr(2),
		LastStatusChange: &now,
	}
	repo := &fakeTrackingRepo{tx: tx}
	cfg := &fakeConfigRepo{
		flowID:  1,
		flowCfg: `{"transitions": {}}`,
	}

	service := &StatusTrackingService{repo: repo, configRepo: cfg}

	_, err := service.UpdateJobStatus(9, 45, &model.StatusUpdateRequest{
		Status: model.StatusFirstInterview,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "BACKWARD_CONFIRM_REQUIRED")
	require.False(t, tx.commitCalled)
	require.True(t, tx.rollbackCalled)
	require.Empty(t, tx.historyInserts)
}

func TestStatusTrackingService_UpdateJobStatus_BackwardConfirmCreatesHistory(t *testing.T) {
	tx := newFakeStatusTx()
	now := time.Now().Add(-30 * time.Minute)
	historyJSON := `{"entries":[]}`
	durationJSON := `{"durations":{}}`
	tx.snapshot = &repository.JobStatusSnapshot{
		Job: model.JobApplication{
			ID:     200,
			UserID: 11,
			Status: model.StatusSecondInterview,
		},
		StatusVersion:    intPtr(7),
		StatusHistoryRaw: &historyJSON,
		DurationStatsRaw: &durationJSON,
		LastStatusChange: &now,
	}

	repo := &fakeTrackingRepo{tx: tx}
	cfg := &fakeConfigRepo{
		flowID:  1,
		flowCfg: `{"transitions": {}}`,
	}

	service := &StatusTrackingService{repo: repo, configRepo: cfg}
	confirm := true

	result, err := service.UpdateJobStatus(11, 200, &model.StatusUpdateRequest{
		Status:          model.StatusFirstInterview,
		ConfirmBackward: &confirm,
		Note:            strPtr("重新安排面试"),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Job)
	require.Equal(t, model.StatusFirstInterview, result.Job.Status)
	require.NotNil(t, result.HistoryID)
	require.NotNil(t, result.StatusVersion)
	require.True(t, len(tx.historyInserts[0].Metadata) > 0)
	require.True(t, tx.commitCalled)
	require.True(t, tx.rollbackCalled)
	require.Len(t, tx.historyInserts, 1)
	require.Equal(t, model.StatusSecondInterview, tx.historyInserts[0].OldStatus)
	require.Equal(t, model.StatusFirstInterview, tx.historyInserts[0].NewStatus)
	require.True(t, tx.historyInserts[0].Metadata != nil)
}

func TestStatusTrackingService_BatchUpdateStatus(t *testing.T) {
	tx := newFakeStatusTx()
	now := time.Now().Add(-1 * time.Hour)
	tx.currentStatusResponse[11] = struct {
		status    model.ApplicationStatus
		changedAt *time.Time
		err       error
	}{status: model.StatusApplied, changedAt: &now}
	tx.currentStatusResponse[12] = struct {
		status    model.ApplicationStatus
		changedAt *time.Time
		err       error
	}{status: model.StatusResumeScreening, changedAt: &now}

	repo := &fakeTrackingRepo{tx: tx}
	cfg := &fakeConfigRepo{
		flowID:  1,
		flowCfg: `{"transitions": {"` + string(model.StatusApplied) + `": ["` + string(model.StatusResumeScreening) + `"], "` + string(model.StatusResumeScreening) + `": ["` + string(model.StatusWrittenTest) + `"]}}`,
	}

	service := &StatusTrackingService{repo: repo, configRepo: cfg}

	err := service.BatchUpdateStatus(55, []model.BatchStatusUpdate{
		{ID: 11, Status: model.StatusResumeScreening},
		{ID: 12, Status: model.StatusWrittenTest},
	})
	require.NoError(t, err)
	require.True(t, tx.commitCalled)
	require.True(t, tx.rollbackCalled)
	require.Len(t, tx.historyInserts, 2)
	require.Len(t, tx.updateCalls, 2)
	require.True(t, tx.updateCalls[0].IncrementVersion)
	require.NotNil(t, tx.updateCalls[0].LastStatusChange)
}

func TestStatusTrackingService_UpdateJobStatus_SecondInterviewDirectToHR(t *testing.T) {
	tx := newFakeStatusTx()
	now := time.Now().Add(-45 * time.Minute)
	historyJSON := `{"entries":[]}`
	durationJSON := `{"durations":{}}`
	tx.snapshot = &repository.JobStatusSnapshot{
		Job: model.JobApplication{
			ID:     205,
			UserID: 77,
			Status: model.StatusSecondInterview,
		},
		StatusHistoryRaw: &historyJSON,
		DurationStatsRaw: &durationJSON,
		LastStatusChange: &now,
	}

	repo := &fakeTrackingRepo{tx: tx}
	cfg := &fakeConfigRepo{
		flowID: 1,
		// 模拟用户模板只允许二面进入三面，校验应通过隐式规则允许直达HR面
		flowCfg: `{"transitions": {"` + string(model.StatusSecondInterview) + `": ["` + string(model.StatusThirdInterview) + `"]}}`,
	}

	service := &StatusTrackingService{repo: repo, configRepo: cfg}

	result, err := service.UpdateJobStatus(77, 205, &model.StatusUpdateRequest{
		Status: model.StatusHRInterview,
	})
	require.NoError(t, err)
	require.NotNil(t, result.Job)
	require.Equal(t, model.StatusHRInterview, result.Job.Status)
	require.True(t, tx.commitCalled)
	require.True(t, tx.rollbackCalled)
	require.Len(t, tx.historyInserts, 1)
	require.Equal(t, model.StatusSecondInterview, tx.historyInserts[0].OldStatus)
	require.Equal(t, model.StatusHRInterview, tx.historyInserts[0].NewStatus)
}

func TestStatusTrackingService_UpdateJobStatus_FirstInterviewDirectToHR(t *testing.T) {
	tx := newFakeStatusTx()
	now := time.Now().Add(-30 * time.Minute)
	historyJSON := `{"entries":[]}`
	durationJSON := `{"durations":{}}`
	tx.snapshot = &repository.JobStatusSnapshot{
		Job: model.JobApplication{
			ID:     306,
			UserID: 52,
			Status: model.StatusFirstInterview,
		},
		StatusHistoryRaw: &historyJSON,
		DurationStatsRaw: &durationJSON,
		LastStatusChange: &now,
	}

	repo := &fakeTrackingRepo{tx: tx}
	cfg := &fakeConfigRepo{
		flowID: 1,
		// 模拟模板仅允许进入二面，校验应通过隐式规则允许直达HR面
		flowCfg: `{"transitions": {"` + string(model.StatusFirstInterview) + `": ["` + string(model.StatusSecondInterview) + `"]}}`,
	}

	service := &StatusTrackingService{repo: repo, configRepo: cfg}

	result, err := service.UpdateJobStatus(52, 306, &model.StatusUpdateRequest{
		Status: model.StatusHRInterview,
	})
	require.NoError(t, err)
	require.NotNil(t, result.Job)
	require.Equal(t, model.StatusHRInterview, result.Job.Status)
	require.True(t, tx.commitCalled)
	require.True(t, tx.rollbackCalled)
	require.Len(t, tx.historyInserts, 1)
	require.Equal(t, model.StatusFirstInterview, tx.historyInserts[0].OldStatus)
	require.Equal(t, model.StatusHRInterview, tx.historyInserts[0].NewStatus)
}

func TestStatusTrackingService_UpdateJobStatus_HRPassToAccepted(t *testing.T) {
	tx := newFakeStatusTx()
	now := time.Now().Add(-10 * time.Minute)
	historyJSON := `{"entries":[]}`
	durationJSON := `{"durations":{}}`
	tx.snapshot = &repository.JobStatusSnapshot{
		Job: model.JobApplication{
			ID:     412,
			UserID: 23,
			Status: model.StatusHRPass,
		},
		StatusHistoryRaw: &historyJSON,
		DurationStatsRaw: &durationJSON,
		LastStatusChange: &now,
	}

	repo := &fakeTrackingRepo{tx: tx}
	cfg := &fakeConfigRepo{
		flowID:  1,
		flowCfg: `{"transitions": {"` + string(model.StatusHRPass) + `": []}}`,
	}

	service := &StatusTrackingService{repo: repo, configRepo: cfg}

	result, err := service.UpdateJobStatus(23, 412, &model.StatusUpdateRequest{
		Status: model.StatusOfferAccepted,
	})
	require.NoError(t, err)
	require.NotNil(t, result.Job)
	require.Equal(t, model.StatusOfferAccepted, result.Job.Status)
	require.True(t, tx.commitCalled)
	require.True(t, tx.rollbackCalled)
	require.Len(t, tx.historyInserts, 1)
	require.Equal(t, model.StatusHRPass, tx.historyInserts[0].OldStatus)
	require.Equal(t, model.StatusOfferAccepted, tx.historyInserts[0].NewStatus)
}

func TestStatusTrackingService_UndoJobStatus_Success(t *testing.T) {
	tx := newFakeStatusTx()
	oldStatus := model.StatusFirstInterview
	changeTime := time.Now().Add(-5 * time.Second)
	tx.latestHistoryEntry = &model.StatusHistoryEntry{
		ID:               88,
		JobApplicationID: 900,
		UserID:           501,
		OldStatus:        &oldStatus,
		NewStatus:        model.StatusSecondInterview,
		StatusChangedAt:  changeTime,
		Metadata:         map[string]interface{}{"source": "test"},
	}
	tx.snapshot = &repository.JobStatusSnapshot{
		Job: model.JobApplication{
			ID:     900,
			UserID: 501,
			Status: model.StatusSecondInterview,
		},
		StatusVersion:    intPtr(3),
		LastStatusChange: &changeTime,
	}

	repo := &fakeTrackingRepo{tx: tx}
	service := &StatusTrackingService{repo: repo}

	result, err := service.UndoJobStatus(501, 900, &model.StatusUndoRequest{
		HistoryID: tx.latestHistoryEntry.ID,
		Version:   intPtr(3),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Job)
	require.Equal(t, oldStatus, result.Job.Status)
	require.Equal(t, oldStatus, result.RevertedTo)
	require.NotZero(t, result.UndoHistoryID)
	require.Equal(t, tx.latestHistoryEntry.ID, result.SourceHistoryID)
	require.NotNil(t, result.StatusVersion)
	require.Equal(t, 2, *result.StatusVersion)
	require.NotNil(t, result.Job.StatusVersion)
	require.Equal(t, *result.StatusVersion, *result.Job.StatusVersion)
	require.True(t, tx.flags["jobview.skip_history"])
	require.True(t, tx.flags["jobview.allow_backward"])
	require.True(t, tx.commitCalled)
	require.True(t, tx.rollbackCalled)
	require.Len(t, tx.historyInserts, 1)
	require.Equal(t, model.StatusSecondInterview, tx.historyInserts[0].OldStatus)
	require.Equal(t, oldStatus, tx.historyInserts[0].NewStatus)
	require.Contains(t, tx.metadataUpdates, tx.latestHistoryEntry.ID)
}

func TestStatusTrackingService_UndoJobStatus_Expired(t *testing.T) {
	tx := newFakeStatusTx()
	oldStatus := model.StatusApplied
	past := time.Now().Add(-2 * time.Minute)
	tx.latestHistoryEntry = &model.StatusHistoryEntry{
		ID:               55,
		JobApplicationID: 300,
		UserID:           77,
		OldStatus:        &oldStatus,
		NewStatus:        model.StatusResumeScreening,
		StatusChangedAt:  past,
	}
	tx.snapshot = &repository.JobStatusSnapshot{
		Job: model.JobApplication{
			ID:     300,
			UserID: 77,
			Status: model.StatusResumeScreening,
		},
		StatusVersion:    intPtr(4),
		LastStatusChange: &past,
	}

	repo := &fakeTrackingRepo{tx: tx}
	service := &StatusTrackingService{repo: repo}

	_, err := service.UndoJobStatus(77, 300, &model.StatusUndoRequest{
		HistoryID: tx.latestHistoryEntry.ID,
		Version:   intPtr(4),
	})
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrUndoExpired))
	require.False(t, tx.commitCalled)
	require.True(t, tx.rollbackCalled)
	require.Empty(t, tx.historyInserts)
}

func TestStatusTrackingService_UpdateJobStatus_HRPassToRejected(t *testing.T) {
	tx := newFakeStatusTx()
	now := time.Now().Add(-15 * time.Minute)
	historyJSON := `{"entries":[]}`
	durationJSON := `{"durations":{}}`
	tx.snapshot = &repository.JobStatusSnapshot{
		Job: model.JobApplication{
			ID:     413,
			UserID: 23,
			Status: model.StatusHRPass,
		},
		StatusHistoryRaw: &historyJSON,
		DurationStatsRaw: &durationJSON,
		LastStatusChange: &now,
	}

	repo := &fakeTrackingRepo{tx: tx}
	cfg := &fakeConfigRepo{
		flowID:  1,
		flowCfg: `{"transitions": {"` + string(model.StatusHRPass) + `": []}}`,
	}

	service := &StatusTrackingService{repo: repo, configRepo: cfg}

	result, err := service.UpdateJobStatus(23, 413, &model.StatusUpdateRequest{
		Status: model.StatusRejected,
	})
	require.NoError(t, err)
	require.NotNil(t, result.Job)
	require.Equal(t, model.StatusRejected, result.Job.Status)
	require.True(t, tx.commitCalled)
	require.True(t, tx.rollbackCalled)
	require.Len(t, tx.historyInserts, 1)
	require.Equal(t, model.StatusHRPass, tx.historyInserts[0].OldStatus)
	require.Equal(t, model.StatusRejected, tx.historyInserts[0].NewStatus)
}

func intPtr(v int) *int {
	return &v
}

func strPtr(v string) *string {
	return &v
}
