package service

import (
	"errors"
	"fmt"
	"jobView-backend/internal/model"
	"jobView-backend/internal/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockJobApplicationRepository 是轻量级的仓储假实现，用于验证 service 对接口的依赖行为
// （避免引入额外的 mocking 框架，同时仍能检测调用频次与入参）
type mockJobApplicationRepository struct {
	createFunc                       func(uint, *model.CreateJobApplicationRequest) (*model.JobApplication, error)
	getByIDFunc                      func(uint, int) (*model.JobApplication, error)
	getAllFunc                       func(uint) ([]model.JobApplication, error)
	getAllPaginatedFunc              func(uint, model.PaginationRequest) (*model.PaginationResponse, error)
	updateFunc                       func(uint, int, *model.UpdateJobApplicationRequest) (*model.JobApplication, error)
	deleteFunc                       func(uint, int) error
	searchFunc                       func(uint, string, model.PaginationRequest) (*model.PaginationResponse, error)
	batchUpdateStatusFunc            func(uint, []model.BatchStatusUpdate) error
	batchDeleteFunc                  func(uint, []int) error
	batchCreateFunc                  func(uint, []model.CreateJobApplicationRequest) ([]model.JobApplication, error)
	getStatusStatisticsFunc          func(uint) (map[string]int, error)
	getHRPassCountFunc               func(uint) (int, error)
	listByDateRangeFunc              func(uint, string, string, model.PaginationRequest) (*model.PaginationResponse, error)
	listWithStatusFiltersFunc        func(uint, *model.ApplicationStatus, []string, model.PaginationRequest) (*model.PaginationResponse, error)
	listRecentApplicationsFunc       func(uint, int) ([]map[string]interface{}, error)
	listUpcomingInterviewsFunc       func(uint, int) ([]map[string]interface{}, error)
	listDailyStatsFunc               func(uint, int) ([]map[string]interface{}, error)
	getStatusStatisticsCalledWithUID []uint
}

func (m *mockJobApplicationRepository) Create(userID uint, req *model.CreateJobApplicationRequest) (*model.JobApplication, error) {
	if m.createFunc == nil {
		return nil, errors.New("createFunc not implemented")
	}
	return m.createFunc(userID, req)
}

func (m *mockJobApplicationRepository) GetByID(userID uint, id int) (*model.JobApplication, error) {
	if m.getByIDFunc == nil {
		return nil, errors.New("getByIDFunc not implemented")
	}
	return m.getByIDFunc(userID, id)
}

func (m *mockJobApplicationRepository) GetAll(userID uint) ([]model.JobApplication, error) {
	if m.getAllFunc == nil {
		return nil, errors.New("getAllFunc not implemented")
	}
	return m.getAllFunc(userID)
}

func (m *mockJobApplicationRepository) GetAllPaginated(userID uint, req model.PaginationRequest) (*model.PaginationResponse, error) {
	if m.getAllPaginatedFunc == nil {
		return nil, errors.New("getAllPaginatedFunc not implemented")
	}
	return m.getAllPaginatedFunc(userID, req)
}

func (m *mockJobApplicationRepository) Update(userID uint, id int, req *model.UpdateJobApplicationRequest) (*model.JobApplication, error) {
	if m.updateFunc == nil {
		return nil, errors.New("updateFunc not implemented")
	}
	return m.updateFunc(userID, id, req)
}

func (m *mockJobApplicationRepository) Delete(userID uint, id int) error {
	if m.deleteFunc == nil {
		return errors.New("deleteFunc not implemented")
	}
	return m.deleteFunc(userID, id)
}

func (m *mockJobApplicationRepository) Search(userID uint, query string, req model.PaginationRequest) (*model.PaginationResponse, error) {
	if m.searchFunc == nil {
		return nil, errors.New("searchFunc not implemented")
	}
	return m.searchFunc(userID, query, req)
}

func (m *mockJobApplicationRepository) BatchUpdateStatus(userID uint, updates []model.BatchStatusUpdate) error {
	if m.batchUpdateStatusFunc == nil {
		return errors.New("batchUpdateStatusFunc not implemented")
	}
	return m.batchUpdateStatusFunc(userID, updates)
}

func (m *mockJobApplicationRepository) BatchDelete(userID uint, ids []int) error {
	if m.batchDeleteFunc == nil {
		return errors.New("batchDeleteFunc not implemented")
	}
	return m.batchDeleteFunc(userID, ids)
}

func (m *mockJobApplicationRepository) BatchCreate(userID uint, applications []model.CreateJobApplicationRequest) ([]model.JobApplication, error) {
	if m.batchCreateFunc == nil {
		return nil, errors.New("batchCreateFunc not implemented")
	}
	return m.batchCreateFunc(userID, applications)
}

func (m *mockJobApplicationRepository) GetStatusStatistics(userID uint) (map[string]int, error) {
	if m.getStatusStatisticsFunc == nil {
		return nil, errors.New("getStatusStatisticsFunc not implemented")
	}
	m.getStatusStatisticsCalledWithUID = append(m.getStatusStatisticsCalledWithUID, userID)
	return m.getStatusStatisticsFunc(userID)
}

func (m *mockJobApplicationRepository) GetHRPassCount(userID uint) (int, error) {
	if m.getHRPassCountFunc == nil {
		return 0, errors.New("getHRPassCountFunc not implemented")
	}
	return m.getHRPassCountFunc(userID)
}

func (m *mockJobApplicationRepository) ListByDateRange(userID uint, startDate, endDate string, req model.PaginationRequest) (*model.PaginationResponse, error) {
	if m.listByDateRangeFunc == nil {
		return nil, errors.New("listByDateRangeFunc not implemented")
	}
	return m.listByDateRangeFunc(userID, startDate, endDate, req)
}

func (m *mockJobApplicationRepository) ListWithStatusFilters(userID uint, status *model.ApplicationStatus, stageStatuses []string, req model.PaginationRequest) (*model.PaginationResponse, error) {
	if m.listWithStatusFiltersFunc == nil {
		return nil, errors.New("listWithStatusFiltersFunc not implemented")
	}
	return m.listWithStatusFiltersFunc(userID, status, stageStatuses, req)
}

func (m *mockJobApplicationRepository) ListRecentApplications(userID uint, limit int) ([]map[string]interface{}, error) {
	if m.listRecentApplicationsFunc == nil {
		return nil, errors.New("listRecentApplicationsFunc not implemented")
	}
	return m.listRecentApplicationsFunc(userID, limit)
}

func (m *mockJobApplicationRepository) ListUpcomingInterviews(userID uint, limit int) ([]map[string]interface{}, error) {
	if m.listUpcomingInterviewsFunc == nil {
		return nil, errors.New("listUpcomingInterviewsFunc not implemented")
	}
	return m.listUpcomingInterviewsFunc(userID, limit)
}

func (m *mockJobApplicationRepository) ListDailyStats(userID uint, days int) ([]map[string]interface{}, error) {
	if m.listDailyStatsFunc == nil {
		return nil, errors.New("listDailyStatsFunc not implemented")
	}
	return m.listDailyStatsFunc(userID, days)
}

// --------- 单测 ---------

func TestJobApplicationService_Create_DelegatesToRepository(t *testing.T) {
	expectedUserID := uint(42)
	expectedCompany := "测试公司"
	expectedReq := &model.CreateJobApplicationRequest{
		CompanyName:   expectedCompany,
		PositionTitle: "后端工程师",
		Status:        model.StatusApplied,
	}

	mockRepo := &mockJobApplicationRepository{}
	mockRepo.createFunc = func(userID uint, req *model.CreateJobApplicationRequest) (*model.JobApplication, error) {
		require.Equal(t, expectedUserID, userID, "service 应传递用户ID")
		require.Equal(t, expectedReq, req, "service 应传递原始请求指针")
		return &model.JobApplication{ID: 7, UserID: userID, CompanyName: req.CompanyName, Status: req.Status}, nil
	}

	svc := NewJobApplicationService(mockRepo)

	job, err := svc.Create(expectedUserID, expectedReq)

	require.NoError(t, err)
	require.NotNil(t, job)
	assert.Equal(t, 7, job.ID)
	assert.Equal(t, expectedCompany, job.CompanyName)
}

func TestJobApplicationService_Delete_DelegatesToRepository(t *testing.T) {
	expectedUserID := uint(101)
	expectedID := 555
	deleteCalled := 0

	mockRepo := &mockJobApplicationRepository{}
	mockRepo.deleteFunc = func(userID uint, id int) error {
		deleteCalled++
		assert.Equal(t, expectedUserID, userID)
		assert.Equal(t, expectedID, id)
		return nil
	}

	svc := NewJobApplicationService(mockRepo)

	err := svc.Delete(expectedUserID, expectedID)

	require.NoError(t, err)
	assert.Equal(t, 1, deleteCalled, "Delete 应仅调用一次仓储")
}

func TestJobApplicationService_GetAll_DelegatesToRepository(t *testing.T) {
	expectedUserID := uint(77)
	mockRepo := &mockJobApplicationRepository{}
	mockRepo.getAllFunc = func(userID uint) ([]model.JobApplication, error) {
		if userID != expectedUserID {
			return nil, fmt.Errorf("unexpected user: %d", userID)
		}
		return []model.JobApplication{{ID: 1, UserID: userID}, {ID: 2, UserID: userID}}, nil
	}

	svc := NewJobApplicationService(mockRepo)

	jobs, err := svc.GetAll(expectedUserID)

	require.NoError(t, err)
	require.Len(t, jobs, 2)
	assert.Equal(t, expectedUserID, jobs[0].UserID)
	assert.Equal(t, expectedUserID, jobs[1].UserID)
}

func TestJobApplicationService_Search_DelegatesToRepository(t *testing.T) {
	expectedUserID := uint(88)
	expectedQuery := "字节"
	req := model.PaginationRequest{Page: 2, PageSize: 10, SortBy: "created_at", SortDir: "DESC"}

	mockRepo := &mockJobApplicationRepository{}
	mockRepo.searchFunc = func(userID uint, query string, r model.PaginationRequest) (*model.PaginationResponse, error) {
		assert.Equal(t, expectedUserID, userID)
		assert.Equal(t, expectedQuery, query)
		assert.Equal(t, req.Page, r.Page)
		assert.Equal(t, req.PageSize, r.PageSize)
		return &model.PaginationResponse{Data: []model.JobApplication{{ID: 99, UserID: userID}}, Total: 1, Page: r.Page, PageSize: r.PageSize}, nil
	}

	svc := NewJobApplicationService(mockRepo)

	resp, err := svc.SearchApplications(expectedUserID, expectedQuery, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	jobs, ok := resp.Data.([]model.JobApplication)
	require.True(t, ok, "响应数据应为岗位列表")
	require.Len(t, jobs, 1)
	assert.Equal(t, 99, jobs[0].ID)
	assert.Equal(t, expectedUserID, jobs[0].UserID)
}

func TestJobApplicationService_BatchUpdateStatus_DelegatesToRepository(t *testing.T) {
	expectedUserID := uint(66)
	updates := []model.BatchStatusUpdate{{ID: 1, Status: model.StatusFirstInterview}, {ID: 2, Status: model.StatusSecondInterview}}

	mockRepo := &mockJobApplicationRepository{}
	batchCalled := 0
	mockRepo.batchUpdateStatusFunc = func(userID uint, got []model.BatchStatusUpdate) error {
		batchCalled++
		assert.Equal(t, expectedUserID, userID)
		assert.Equal(t, updates, got)
		return nil
	}

	svc := NewJobApplicationService(mockRepo)

	err := svc.BatchUpdateStatus(expectedUserID, updates)

	require.NoError(t, err)
	assert.Equal(t, 1, batchCalled, "BatchUpdateStatus 应调用仓储一次")
}

func TestJobApplicationService_BatchDelete_DelegatesToRepository(t *testing.T) {
	expectedUserID := uint(55)
	ids := []int{11, 22, 33}

	mockRepo := &mockJobApplicationRepository{}
	deleteCalled := 0
	mockRepo.batchDeleteFunc = func(userID uint, got []int) error {
		deleteCalled++
		assert.Equal(t, expectedUserID, userID)
		assert.Equal(t, ids, got)
		return nil
	}

	svc := NewJobApplicationService(mockRepo)

	err := svc.BatchDelete(expectedUserID, ids)

	require.NoError(t, err)
	assert.Equal(t, 1, deleteCalled, "BatchDelete 应调用仓储一次")
}

func TestJobApplicationService_BatchCreate_DelegatesToRepository(t *testing.T) {
	expectedUserID := uint(44)
	batch := []model.CreateJobApplicationRequest{
		{CompanyName: "A公司", PositionTitle: "产品经理"},
		{CompanyName: "B公司", PositionTitle: "测试开发", Status: model.StatusFirstInterview},
	}

	mockRepo := &mockJobApplicationRepository{}
	mockRepo.batchCreateFunc = func(userID uint, reqs []model.CreateJobApplicationRequest) ([]model.JobApplication, error) {
		assert.Equal(t, expectedUserID, userID)
		assert.Equal(t, batch, reqs)
		return []model.JobApplication{{ID: 1, UserID: userID}, {ID: 2, UserID: userID}}, nil
	}

	svc := NewJobApplicationService(mockRepo)

	jobs, err := svc.BatchCreate(expectedUserID, batch)

	require.NoError(t, err)
	require.Len(t, jobs, 2)
	assert.Equal(t, expectedUserID, jobs[0].UserID)
	assert.Equal(t, expectedUserID, jobs[1].UserID)
}

func TestJobApplicationService_GetStatusStatistics_AggregatesFromRepository(t *testing.T) {
	expectedUserID := uint(99)
	mockRepo := &mockJobApplicationRepository{}
	mockRepo.getStatusStatisticsFunc = func(userID uint) (map[string]int, error) {
		assert.Equal(t, expectedUserID, userID)
		return map[string]int{
			string(model.StatusApplied):        3,
			string(model.StatusFirstInterview): 2,
			string(model.StatusHRPass):         1,
			string(model.StatusRejected):       4,
		}, nil
	}
	mockRepo.getHRPassCountFunc = func(userID uint) (int, error) {
		assert.Equal(t, expectedUserID, userID)
		return 6, nil
	}

	svc := NewJobApplicationService(mockRepo)

	stats, err := svc.GetStatusStatistics(expectedUserID)

	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, expectedUserID, stats["user_id"])
	assert.Equal(t, 10, stats["total_applications"])
	assert.Equal(t, 5, stats["in_progress"])
	assert.Equal(t, 5, stats["passed"])
	assert.Equal(t, 0, stats["failed"])
	assert.Equal(t, 6, stats["hr_passed"])
	breakdown, ok := stats["status_breakdown"].(map[string]int)
	require.True(t, ok)
	assert.Equal(t, 3, breakdown[string(model.StatusApplied)])
	assert.Equal(t, "100.0%", stats["pass_rate"])
}

func TestJobApplicationService_GetApplicationsByDateRange_DelegatesToRepository(t *testing.T) {
	expectedUserID := uint(77)
	req := model.PaginationRequest{Page: 1, PageSize: 5, SortBy: "application_date", SortDir: "DESC"}

	mockRepo := &mockJobApplicationRepository{}
	mockRepo.listByDateRangeFunc = func(userID uint, start, end string, r model.PaginationRequest) (*model.PaginationResponse, error) {
		assert.Equal(t, expectedUserID, userID)
		assert.Equal(t, "2024-01-01", start)
		assert.Equal(t, "2024-01-31", end)
		assert.Equal(t, req.Page, r.Page)
		return &model.PaginationResponse{Data: []model.JobApplication{{ID: 1, UserID: userID}}, Total: 1, Page: r.Page, PageSize: r.PageSize}, nil
	}

	svc := NewJobApplicationService(mockRepo)

	resp, err := svc.GetApplicationsByDateRange(expectedUserID, "2024-01-01", "2024-01-31", req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	jobs, ok := resp.Data.([]model.JobApplication)
	require.True(t, ok)
	require.Len(t, jobs, 1)
	assert.Equal(t, expectedUserID, jobs[0].UserID)
}

func TestJobApplicationService_GetJobApplicationsWithStatusFilters_DelegatesToRepository_WithStage(t *testing.T) {
	expectedUserID := uint(88)
	status := model.StatusFirstInterview
	req := model.PaginationRequest{Page: 3, PageSize: 15, SortBy: "updated_at", SortDir: "ASC"}

	mockRepo := &mockJobApplicationRepository{}
	svc := NewJobApplicationService(mockRepo)
	stage := "interviews"
	expectedStageStatuses := svc.getStatusesByStage(stage)

	mockRepo.listWithStatusFiltersFunc = func(userID uint, s *model.ApplicationStatus, stageStatuses []string, r model.PaginationRequest) (*model.PaginationResponse, error) {
		assert.Equal(t, expectedUserID, userID)
		assert.NotNil(t, s)
		assert.Equal(t, status, *s)
		assert.Equal(t, expectedStageStatuses, stageStatuses)
		assert.Equal(t, req.Page, r.Page)
		return &model.PaginationResponse{Data: []model.JobApplication{{ID: 10, UserID: userID}}, Total: 1, Page: r.Page, PageSize: r.PageSize}, nil
	}

	resp, err := svc.GetJobApplicationsWithStatusFilters(expectedUserID, &status, &stage, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	jobs, ok := resp.Data.([]model.JobApplication)
	require.True(t, ok)
	require.Len(t, jobs, 1)
	assert.Equal(t, expectedUserID, jobs[0].UserID)
}

func TestJobApplicationService_GetJobApplicationsWithStatusFilters_DelegatesToRepository(t *testing.T) {
	expectedUserID := uint(88)
	status := model.StatusFirstInterview
	req := model.PaginationRequest{Page: 2, PageSize: 20, SortBy: "created_at", SortDir: "ASC"}

	mockRepo := &mockJobApplicationRepository{}
	svc := NewJobApplicationService(mockRepo)

	mockRepo.listWithStatusFiltersFunc = func(userID uint, s *model.ApplicationStatus, stageStatuses []string, r model.PaginationRequest) (*model.PaginationResponse, error) {
		assert.Equal(t, expectedUserID, userID)
		assert.Equal(t, &status, s)
		assert.Nil(t, stageStatuses)
		assert.Equal(t, req.Page, r.Page)
		return &model.PaginationResponse{Data: []model.JobApplication{{ID: 10, UserID: userID}}, Total: 1, Page: r.Page, PageSize: r.PageSize}, nil
	}

	resp, err := svc.GetJobApplicationsWithStatusFilters(expectedUserID, &status, nil, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	jobs, ok := resp.Data.([]model.JobApplication)
	require.True(t, ok)
	require.Len(t, jobs, 1)
	assert.Equal(t, 10, jobs[0].ID)
}

func TestJobApplicationService_GetDashboardData_DelegatesToRepository(t *testing.T) {
	expectedUserID := uint(50)
	mockRepo := &mockJobApplicationRepository{}
	mockRepo.getStatusStatisticsFunc = func(userID uint) (map[string]int, error) {
		assert.Equal(t, expectedUserID, userID)
		return map[string]int{string(model.StatusApplied): 2}, nil
	}
	mockRepo.getHRPassCountFunc = func(userID uint) (int, error) {
		assert.Equal(t, expectedUserID, userID)
		return 1, nil
	}
	mockRepo.listRecentApplicationsFunc = func(userID uint, limit int) ([]map[string]interface{}, error) {
		assert.Equal(t, expectedUserID, userID)
		assert.Equal(t, 10, limit)
		return []map[string]interface{}{{"id": 1}}, nil
	}
	mockRepo.listUpcomingInterviewsFunc = func(userID uint, limit int) ([]map[string]interface{}, error) {
		assert.Equal(t, expectedUserID, userID)
		assert.Equal(t, 5, limit)
		return []map[string]interface{}{{"id": 2}}, nil
	}
	mockRepo.listDailyStatsFunc = func(userID uint, days int) ([]map[string]interface{}, error) {
		assert.Equal(t, expectedUserID, userID)
		assert.Equal(t, 30, days)
		return []map[string]interface{}{{"date": "2024-01-01", "count": 3}}, nil
	}

	svc := NewJobApplicationService(mockRepo)

	dashboard, err := svc.GetDashboardData(expectedUserID)

	require.NoError(t, err)
	require.NotNil(t, dashboard)
	assert.Contains(t, dashboard, "statistics")
	assert.Contains(t, dashboard, "recent_applications")
	assert.Contains(t, dashboard, "upcoming_interviews")
	assert.Contains(t, dashboard, "daily_stats")
}

// 确保 mock 仓储实现满足接口约束
var _ repository.JobApplicationRepository = (*mockJobApplicationRepository)(nil)
