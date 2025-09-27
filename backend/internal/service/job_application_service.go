package service

import (
	"fmt"
	"jobView-backend/internal/model"
	"jobView-backend/internal/repository"
	"time"
)

type JobApplicationService struct {
	repo repository.JobApplicationRepository
}

func NewJobApplicationService(repo repository.JobApplicationRepository) *JobApplicationService {
	return &JobApplicationService{repo: repo}
}

// Create 创建新的投递记录
func (s *JobApplicationService) Create(userID uint, req *model.CreateJobApplicationRequest) (*model.JobApplication, error) {
	status := req.Status
	if status == "" {
		status = model.StatusApplied
	}
	if !status.IsValid() {
		return nil, fmt.Errorf("invalid status: %s", status)
	}
	return s.repo.Create(userID, req)
}

// GetByID 根据ID获取投递记录（带用户权限检查）
func (s *JobApplicationService) GetByID(userID uint, id int) (*model.JobApplication, error) {
	return s.repo.GetByID(userID, id)
}

// GetAllPaginated 获取用户的投递记录（分页版）
func (s *JobApplicationService) GetAllPaginated(userID uint, req model.PaginationRequest) (*model.PaginationResponse, error) {
	req.ValidateAndSetDefaults()
	if req.Status != nil && !req.Status.IsValid() {
		return nil, fmt.Errorf("invalid status: %s", *req.Status)
	}
	return s.repo.GetAllPaginated(userID, req)
}

// GetAll 获取用户的全部投递记录（带默认排序限制）
func (s *JobApplicationService) GetAll(userID uint) ([]model.JobApplication, error) {
	return s.repo.GetAll(userID)
}

// Update 更新投递记录
func (s *JobApplicationService) Update(userID uint, id int, req *model.UpdateJobApplicationRequest) (*model.JobApplication, error) {
	if req.Status != nil && !req.Status.IsValid() {
		return nil, fmt.Errorf("invalid status: %s", *req.Status)
	}
	return s.repo.Update(userID, id, req)
}

// Delete 删除投递记录
func (s *JobApplicationService) Delete(userID uint, id int) error {
	return s.repo.Delete(userID, id)
}

// BatchCreate 批量创建投递记录 - 高性能批量插入
func (s *JobApplicationService) BatchCreate(userID uint, applications []model.CreateJobApplicationRequest) ([]model.JobApplication, error) {
	if len(applications) == 0 {
		return []model.JobApplication{}, nil
	}
	if len(applications) > 50 {
		return nil, fmt.Errorf("batch size too large: maximum 50 applications allowed, got %d", len(applications))
	}
	for _, req := range applications {
		status := req.Status
		if status == "" {
			status = model.StatusApplied
		}
		if !status.IsValid() {
			return nil, fmt.Errorf("invalid status: %s", status)
		}
	}
	return s.repo.BatchCreate(userID, applications)
}

func (s *JobApplicationService) BatchUpdateStatus(userID uint, updates []model.BatchStatusUpdate) error {
	if len(updates) == 0 {
		return nil
	}
	if len(updates) > 100 {
		return fmt.Errorf("batch size too large: maximum 100 updates allowed, got %d", len(updates))
	}
	for _, update := range updates {
		if !update.Status.IsValid() {
			return fmt.Errorf("invalid status: %s for ID %d", update.Status, update.ID)
		}
	}
	return s.repo.BatchUpdateStatus(userID, updates)
}

func (s *JobApplicationService) BatchDelete(userID uint, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	if len(ids) > 100 {
		return fmt.Errorf("batch size too large: maximum 100 deletions allowed, got %d", len(ids))
	}
	return s.repo.BatchDelete(userID, ids)
}

func (s *JobApplicationService) SearchApplications(userID uint, searchQuery string, req model.PaginationRequest) (*model.PaginationResponse, error) {
	req.ValidateAndSetDefaults()
	if searchQuery == "" {
		return s.GetAllPaginated(userID, req)
	}
	return s.repo.Search(userID, searchQuery, req)
}

func (s *JobApplicationService) GetApplicationsByDateRange(userID uint, startDate, endDate string, req model.PaginationRequest) (*model.PaginationResponse, error) {
	req.ValidateAndSetDefaults()
	return s.repo.ListByDateRange(userID, startDate, endDate, req)
}

func (s *JobApplicationService) GetJobApplicationsWithStatusFilters(userID uint, status *model.ApplicationStatus, stage *string, req model.PaginationRequest) (*model.PaginationResponse, error) {
	req.ValidateAndSetDefaults()
	if status != nil && !status.IsValid() {
		return nil, fmt.Errorf("invalid status: %s", *status)
	}

	var stageStatuses []string
	if stage != nil {
		stageStatuses = s.getStatusesByStage(*stage)
	}

	return s.repo.ListWithStatusFilters(userID, status, stageStatuses, req)
}

// GetStatusStatistics 获取用户的状态统计信息
func (s *JobApplicationService) GetStatusStatistics(userID uint) (map[string]interface{}, error) {
	counts, err := s.repo.GetStatusStatistics(userID)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"user_id":            userID,
		"total_applications": 0,
		"in_progress":        0,
		"passed":             0,
		"failed":             0,
		"status_breakdown":   counts,
	}

	for status, count := range counts {
		stats["total_applications"] = stats["total_applications"].(int) + count
		appStatus := model.ApplicationStatus(status)
		if appStatus.IsInProgressStatus() {
			stats["in_progress"] = stats["in_progress"].(int) + count
		} else if appStatus.IsPassedStatus() {
			stats["passed"] = stats["passed"].(int) + count
		} else if appStatus.IsFailedStatus() {
			stats["failed"] = stats["failed"].(int) + count
		}
	}

	completed := stats["passed"].(int) + stats["failed"].(int)
	if completed > 0 {
		passRate := float64(stats["passed"].(int)) / float64(completed) * 100
		stats["pass_rate"] = fmt.Sprintf("%.1f%%", passRate)
	} else {
		stats["pass_rate"] = "N/A"
	}

	return stats, nil
}

// getStatusesByStage 根据阶段获取对应的状态列表
func (s *JobApplicationService) getStatusesByStage(stage string) []string {
	stageMap := map[string][]string{
		"application":  {"已投递"},
		"screening":    {"简历筛选中", "简历筛选未通过"},
		"written_test": {"笔试中", "笔试通过", "笔试未通过"},
		"interviews": {
			"一面中", "一面通过", "一面未通过",
			"二面中", "二面通过", "二面未通过",
			"三面中", "三面通过", "三面未通过",
			"HR面中", "HR面通过", "HR面未通过",
		},
		"final": {
			"待发offer", "已收到offer", "已接受offer",
			"已拒绝", "流程结束",
		},
		"in_progress": {"已投递", "简历筛选中", "笔试中", "一面中", "二面中", "三面中", "HR面中"},
		"passed":      {"笔试通过", "一面通过", "二面通过", "三面通过", "HR面通过", "待发offer", "已收到offer", "已接受offer", "流程结束"},
		"failed":      {"简历筛选未通过", "笔试未通过", "一面未通过", "二面未通过", "三面未通过", "HR面未通过", "已拒绝"},
	}

	if statuses, exists := stageMap[stage]; exists {
		return statuses
	}
	return []string{}
}

// GetDashboardData 获取仪表板数据
func (s *JobApplicationService) GetDashboardData(userID uint) (map[string]interface{}, error) {
	// 获取状态统计
	statistics, err := s.GetStatusStatistics(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status statistics: %w", err)
	}

	recentApplications, err := s.repo.ListRecentApplications(userID, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent applications: %w", err)
	}

	upcomingInterviews, err := s.repo.ListUpcomingInterviews(userID, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming interviews: %w", err)
	}

	dailyStats, err := s.repo.ListDailyStats(userID, 30)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily stats: %w", err)
	}

	// 构建仪表板数据
	dashboard := map[string]interface{}{
		"statistics":          statistics,
		"recent_applications": recentApplications,
		"upcoming_interviews": upcomingInterviews,
		"daily_stats":         dailyStats,
		"generated_at":        time.Now(),
	}

	return dashboard, nil
}
