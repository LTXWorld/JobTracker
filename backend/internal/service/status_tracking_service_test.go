package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"jobView-backend/internal/model"
	"jobView-backend/internal/repository"
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

func (f *fakeTrackingRepo) BeginTx(ctx context.Context) (repository.StatusTrackingTx, error) {
	f.used = true
	return f.tx, f.err
}

func (f *fakeTrackingRepo) GetStatusAnalytics(ctx context.Context, userID uint) (*model.StatusAnalyticsResponse, error) {
	panic("not implemented")
}

func (f *fakeTrackingRepo) GetStatusTrends(ctx context.Context, userID uint, days int) ([]model.StatusTrend, error) {
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
	updateCalls           []repository.UpdateJobApplicationParams
	commitCalled          bool
	rollbackCalled        bool
	flags                 map[string]bool
	currentStatusResponse map[int]struct {
		status    model.ApplicationStatus
		changedAt *time.Time
		err       error
	}
}

func newFakeStatusTx() *fakeStatusTx {
	return &fakeStatusTx{
		flags: make(map[string]bool),
		currentStatusResponse: make(map[int]struct {
			status    model.ApplicationStatus
			changedAt *time.Time
			err       error
		}),
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
	require.Equal(t, model.StatusWrittenTest, result.Status)
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
	require.Equal(t, model.StatusHRInterview, result.Status)
	require.True(t, tx.commitCalled)
	require.True(t, tx.rollbackCalled)
	require.Len(t, tx.historyInserts, 1)
	require.Equal(t, model.StatusSecondInterview, tx.historyInserts[0].OldStatus)
	require.Equal(t, model.StatusHRInterview, tx.historyInserts[0].NewStatus)
}

func intPtr(v int) *int {
	return &v
}
