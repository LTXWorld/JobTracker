package service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"jobView-backend/internal/model"
	"jobView-backend/internal/repository"
)

type fakeExportRepo struct {
	dailyCount   int
	activeCount  int
	countData    int
	applications []model.JobApplication
	createdTasks []*model.ExportTask
	updatedTasks []*model.ExportTask
}

func (f *fakeExportRepo) CreateTask(ctx context.Context, task *model.ExportTask) error {
	clone := *task
	f.createdTasks = append(f.createdTasks, &clone)
	return nil
}

func (f *fakeExportRepo) UpdateTask(ctx context.Context, task *model.ExportTask) error {
	clone := *task
	f.updatedTasks = append(f.updatedTasks, &clone)
	return nil
}

func (f *fakeExportRepo) GetTaskStatus(ctx context.Context, taskID string, userID uint) (*repository.ExportTaskStatusRecord, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f *fakeExportRepo) GetDownloadInfo(ctx context.Context, taskID string, userID uint) (*repository.ExportDownloadRecord, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f *fakeExportRepo) CountHistory(ctx context.Context, userID uint) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (f *fakeExportRepo) ListHistory(ctx context.Context, userID uint, limit, offset int) ([]repository.ExportHistoryRow, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f *fakeExportRepo) CountDailyExports(ctx context.Context, userID uint, day time.Time) (int, error) {
	return f.dailyCount, nil
}

func (f *fakeExportRepo) CountActiveExports(ctx context.Context, userID uint) (int, error) {
	return f.activeCount, nil
}

func (f *fakeExportRepo) CountExportData(ctx context.Context, userID uint, filters *model.ExportFilters) (int, error) {
	return f.countData, nil
}

func (f *fakeExportRepo) FetchExportData(ctx context.Context, userID uint, filters *model.ExportFilters, offset, limit int) ([]model.JobApplication, error) {
	return f.applications, nil
}

func (f *fakeExportRepo) ListExpiredTasks(ctx context.Context) ([]repository.ExportExpiredTask, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f *fakeExportRepo) DeleteTask(ctx context.Context, taskID string) error {
	return fmt.Errorf("not implemented")
}

func TestExportService_StartExport_Sync(t *testing.T) {
	repo := &fakeExportRepo{
		countData: 1,
		applications: []model.JobApplication{
			{
				ID:              1,
				UserID:          10,
				CompanyName:     "测试公司",
				PositionTitle:   "测试岗位",
				ApplicationDate: "2024-01-01",
				Status:          model.StatusApplied,
			},
		},
	}

	svc := NewExportService(repo)
	svc.tempDir = t.TempDir()

	resp, err := svc.StartExport(10, &model.ExportRequest{Format: "xlsx"})
	require.NoError(t, err)
	require.Equal(t, model.TaskStatusCompleted, resp.Status)
	require.NotNil(t, resp.DownloadURL)

	require.Len(t, repo.createdTasks, 1)
	require.Greater(t, len(repo.updatedTasks), 0)
	finalTask := repo.updatedTasks[len(repo.updatedTasks)-1]
	require.NotNil(t, finalTask.FilePath)
	info, err := os.Stat(*finalTask.FilePath)
	require.NoError(t, err)
	require.Greater(t, info.Size(), int64(0))
	t.Cleanup(func() { _ = os.Remove(*finalTask.FilePath) })
}
