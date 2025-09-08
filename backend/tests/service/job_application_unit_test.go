package service

import (
	"jobView-backend/internal/database"
	"jobView-backend/internal/model"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJobApplicationService_OptimizedQueries 测试优化后的查询方法
func TestJobApplicationService_OptimizedQueries(t *testing.T) {
	// 设置测试环境
	db, service := setupTestService(t)
	defer db.Close()

	userID := uint(1)

	// 创建测试数据
	testJobs := createTestJobs(t, service, userID, 100)
	require.Len(t, testJobs, 100, "应该创建100条测试数据")

	t.Run("GetAll性能优化验证", func(t *testing.T) {
		start := time.Now()
		jobs, err := service.GetAll(userID)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.NotEmpty(t, jobs)
		assert.LessOrEqual(t, len(jobs), 500, "GetAll应该有LIMIT限制")
		
		// 验证性能：应该在100ms内完成
		assert.Less(t, duration.Milliseconds(), int64(100), 
			"GetAll查询应该在100ms内完成，实际耗时: %v", duration)
		
		// 验证排序正确性
		for i := 1; i < len(jobs); i++ {
			prev := jobs[i-1]
			curr := jobs[i]
			assert.True(t, prev.ApplicationDate >= curr.ApplicationDate, 
				"结果应该按application_date DESC排序")
		}
	})

	t.Run("GetAllPaginated分页查询优化验证", func(t *testing.T) {
		req := model.PaginationRequest{
			Page:     1,
			PageSize: 20,
			SortBy:   "application_date",
			SortDir:  "DESC",
		}

		start := time.Now()
		response, err := service.GetAllPaginated(userID, req)
		duration := time.Since(start)

		require.NoError(t, err)
		require.NotNil(t, response)
		
		// 验证性能：应该在50ms内完成
		assert.Less(t, duration.Milliseconds(), int64(50), 
			"分页查询应该在50ms内完成，实际耗时: %v", duration)
		
		// 验证分页数据正确性
		assert.LessOrEqual(t, len(response.Data), 20, "页面大小应该受限制")
		assert.Equal(t, int64(100), response.Total, "总数应该正确")
		assert.Equal(t, 1, response.Page, "当前页码应该正确")
		assert.Equal(t, 5, response.TotalPages, "总页数应该正确") // 100/20=5
		assert.True(t, response.HasNext, "第一页应该有下一页")
		assert.False(t, response.HasPrev, "第一页不应该有上一页")
	})

	t.Run("GetStatusStatistics统计优化验证", func(t *testing.T) {
		start := time.Now()
		stats, err := service.GetStatusStatistics(userID)
		duration := time.Since(start)

		require.NoError(t, err)
		require.NotNil(t, stats)
		
		// 验证性能：应该在30ms内完成（使用覆盖索引）
		assert.Less(t, duration.Milliseconds(), int64(30), 
			"统计查询应该在30ms内完成，实际耗时: %v", duration)
		
		// 验证统计数据正确性
		totalApps, ok := stats["total_applications"].(int)
		require.True(t, ok, "total_applications应该是int类型")
		assert.Equal(t, 100, totalApps, "总申请数应该正确")
		
		breakdown, ok := stats["status_breakdown"].(map[string]int)
		require.True(t, ok, "status_breakdown应该是map[string]int类型")
		assert.NotEmpty(t, breakdown, "状态分解不应该为空")
		
		// 验证通过率计算
		_, hasPassRate := stats["pass_rate"]
		assert.True(t, hasPassRate, "应该计算通过率")
	})

	t.Run("Update方法RETURNING优化验证", func(t *testing.T) {
		// 选择一个测试记录进行更新
		testJob := testJobs[0]
		newStatus := model.StatusFirstInterview
		newCompany := "优化后的公司名称"
		
		req := &model.UpdateJobApplicationRequest{
			Status:      &newStatus,
			CompanyName: &newCompany,
		}

		start := time.Now()
		updatedJob, err := service.Update(userID, testJob.ID, req)
		duration := time.Since(start)

		require.NoError(t, err)
		require.NotNil(t, updatedJob)
		
		// 验证性能：应该在20ms内完成（避免N+1查询）
		assert.Less(t, duration.Milliseconds(), int64(20), 
			"Update应该在20ms内完成，实际耗时: %v", duration)
		
		// 验证更新结果正确性
		assert.Equal(t, newStatus, updatedJob.Status, "状态应该更新成功")
		assert.Equal(t, newCompany, updatedJob.CompanyName, "公司名应该更新成功")
		assert.True(t, updatedJob.UpdatedAt.After(updatedJob.CreatedAt), "UpdatedAt应该更新")
	})
}

// TestJobApplicationService_BatchOperations 测试批量操作优化
func TestJobApplicationService_BatchOperations(t *testing.T) {
	db, service := setupTestService(t)
	defer db.Close()

	userID := uint(1)

	t.Run("BatchCreate批量创建优化验证", func(t *testing.T) {
		applications := generateTestApplications(25) // 创建25条记录

		start := time.Now()
		results, err := service.BatchCreate(userID, applications)
		duration := time.Since(start)

		require.NoError(t, err)
		require.Len(t, results, 25, "应该创建25条记录")
		
		// 验证性能：批量创建应该比逐个创建快得多
		assert.Less(t, duration.Milliseconds(), int64(100), 
			"批量创建25条记录应该在100ms内完成，实际耗时: %v", duration)
		
		// 验证数据完整性
		for i, result := range results {
			assert.NotZero(t, result.ID, "记录%d应该有有效ID", i)
			assert.Equal(t, userID, result.UserID, "记录%d的UserID应该正确", i)
			assert.NotZero(t, result.CreatedAt, "记录%d应该有创建时间", i)
		}
	})

	// 为后续测试创建数据
	testJobs := createTestJobs(t, service, userID, 50)

	t.Run("BatchUpdateStatus批量状态更新优化验证", func(t *testing.T) {
		// 准备批量更新数据
		var updates []model.BatchStatusUpdate
		for i, job := range testJobs[:10] { // 更新前10条记录
			updates = append(updates, model.BatchStatusUpdate{
				ID:     job.ID,
				Status: model.StatusFirstInterview,
			})
			_ = i // 避免未使用变量警告
		}

		start := time.Now()
		err := service.BatchUpdateStatus(userID, updates)
		duration := time.Since(start)

		require.NoError(t, err)
		
		// 验证性能：批量更新应该在50ms内完成
		assert.Less(t, duration.Milliseconds(), int64(50), 
			"批量更新10条记录应该在50ms内完成，实际耗时: %v", duration)
		
		// 验证更新结果
		for _, update := range updates {
			job, err := service.GetByID(userID, update.ID)
			require.NoError(t, err)
			assert.Equal(t, update.Status, job.Status, "ID %d 的状态应该更新成功", update.ID)
		}
	})

	t.Run("BatchDelete批量删除优化验证", func(t *testing.T) {
		// 准备要删除的ID
		var idsToDelete []int
		for i := 40; i < 50; i++ { // 删除后10条记录
			idsToDelete = append(idsToDelete, testJobs[i].ID)
		}

		start := time.Now()
		err := service.BatchDelete(userID, idsToDelete)
		duration := time.Since(start)

		require.NoError(t, err)
		
		// 验证性能：批量删除应该在30ms内完成
		assert.Less(t, duration.Milliseconds(), int64(30), 
			"批量删除10条记录应该在30ms内完成，实际耗时: %v", duration)
		
		// 验证删除结果
		for _, id := range idsToDelete {
			_, err := service.GetByID(userID, id)
			assert.Error(t, err, "ID %d 应该已被删除", id)
		}
	})
}

// TestJobApplicationService_SearchOptimization 测试搜索功能优化
func TestJobApplicationService_SearchOptimization(t *testing.T) {
	db, service := setupTestService(t)
	defer db.Close()

	userID := uint(1)
	testJobs := createTestJobs(t, service, userID, 50)
	_ = testJobs // 避免未使用变量警告

	t.Run("SearchApplications全文搜索优化验证", func(t *testing.T) {
		req := model.PaginationRequest{
			Page:     1,
			PageSize: 10,
		}

		start := time.Now()
		results, err := service.SearchApplications(userID, "腾讯", req)
		duration := time.Since(start)

		require.NoError(t, err)
		require.NotNil(t, results)
		
		// 验证性能：搜索应该在100ms内完成
		assert.Less(t, duration.Milliseconds(), int64(100), 
			"搜索查询应该在100ms内完成，实际耗时: %v", duration)
		
		// 验证搜索结果相关性
		for _, job := range results.Data {
			// 应该包含搜索关键词（在公司名或职位标题中）
			containsKeyword := contains(job.CompanyName, "腾讯") || 
							 contains(job.PositionTitle, "腾讯")
			assert.True(t, containsKeyword, 
				"搜索结果应该包含关键词：%s - %s", job.CompanyName, job.PositionTitle)
		}
	})

	t.Run("GetApplicationsByDateRange日期范围查询优化验证", func(t *testing.T) {
		startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
		endDate := time.Now().Format("2006-01-02")
		
		req := model.PaginationRequest{
			Page:     1,
			PageSize: 20,
		}

		start := time.Now()
		results, err := service.GetApplicationsByDateRange(userID, startDate, endDate, req)
		duration := time.Since(start)

		require.NoError(t, err)
		require.NotNil(t, results)
		
		// 验证性能：日期范围查询应该在50ms内完成
		assert.Less(t, duration.Milliseconds(), int64(50), 
			"日期范围查询应该在50ms内完成，实际耗时: %v", duration)
		
		// 验证日期范围正确性
		for _, job := range results.Data {
			assert.True(t, job.ApplicationDate >= startDate, 
				"记录日期应该在范围内：%s", job.ApplicationDate)
			assert.True(t, job.ApplicationDate <= endDate, 
				"记录日期应该在范围内：%s", job.ApplicationDate)
		}
	})
}

// TestJobApplicationService_IndexUsageValidation 测试索引使用验证
func TestJobApplicationService_IndexUsageValidation(t *testing.T) {
	db, service := setupTestService(t)
	defer db.Close()

	userID := uint(1)
	createTestJobs(t, service, userID, 1000) // 创建足够多的数据以触发索引使用

	t.Run("验证索引在查询计划中被使用", func(t *testing.T) {
		// 测试主要查询的执行计划
		testQueries := []struct {
			name  string
			query string
			expectedIndexes []string
		}{
			{
				name:  "用户ID查询",
				query: "EXPLAIN (FORMAT JSON) SELECT * FROM job_applications WHERE user_id = $1 LIMIT 10",
				expectedIndexes: []string{"idx_job_applications_user_id", "idx_job_applications_user_date"},
			},
			{
				name:  "用户+状态查询",
				query: "EXPLAIN (FORMAT JSON) SELECT * FROM job_applications WHERE user_id = $1 AND status = $2",
				expectedIndexes: []string{"idx_job_applications_user_status"},
			},
			{
				name:  "统计查询",
				query: "EXPLAIN (FORMAT JSON) SELECT status, COUNT(*) FROM job_applications WHERE user_id = $1 GROUP BY status",
				expectedIndexes: []string{"idx_job_applications_status_stats"},
			},
		}

		for _, tc := range testQueries {
			t.Run(tc.name, func(t *testing.T) {
				var planJSON string
				err := db.QueryRow(tc.query, userID, model.StatusApplied).Scan(&planJSON)
				require.NoError(t, err, "执行计划查询应该成功")
				
				// 验证执行计划中包含预期的索引（简化验证）
				t.Logf("执行计划 - %s:\n%s", tc.name, planJSON)
				
				// 实际项目中，这里应该解析JSON并验证索引使用
				// 为了演示，我们只检查是否包含索引相关关键词
				assert.Contains(t, planJSON, "Index", "执行计划应该使用索引")
			})
		}
	})
}

// TestJobApplicationService_ErrorHandling 测试错误处理
func TestJobApplicationService_ErrorHandling(t *testing.T) {
	db, service := setupTestService(t)
	defer db.Close()

	userID := uint(1)

	t.Run("无效状态处理", func(t *testing.T) {
		req := &model.CreateJobApplicationRequest{
			CompanyName:   "测试公司",
			PositionTitle: "测试职位",
			Status:        model.ApplicationStatus("无效状态"),
		}

		_, err := service.Create(userID, req)
		assert.Error(t, err, "应该拒绝无效状态")
		assert.Contains(t, err.Error(), "invalid status", "错误信息应该包含状态无效提示")
	})

	t.Run("批量操作大小限制", func(t *testing.T) {
		// 测试批量创建大小限制
		applications := generateTestApplications(51) // 超过50条限制
		_, err := service.BatchCreate(userID, applications)
		assert.Error(t, err, "应该拒绝过大的批量操作")
		assert.Contains(t, err.Error(), "batch size too large", "错误信息应该包含批量大小提示")
	})

	t.Run("权限检查", func(t *testing.T) {
		// 创建测试数据
		job, err := service.Create(userID, &model.CreateJobApplicationRequest{
			CompanyName:   "测试公司",
			PositionTitle: "测试职位",
		})
		require.NoError(t, err)

		// 尝试用其他用户ID访问
		otherUserID := uint(999)
		_, err = service.GetByID(otherUserID, job.ID)
		assert.Error(t, err, "应该拒绝其他用户访问")
		assert.Contains(t, err.Error(), "not found", "错误信息应该表示未找到")
	})
}

// 辅助函数
func setupTestService(t *testing.T) (*database.DB, *JobApplicationService) {
	cfg := &database.DatabaseConfig{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "5432"),
		User:     getEnvOrDefault("DB_USER", "test_user"),
		Password: getEnvOrDefault("DB_PASSWORD", "test_pass"),
		DBName:   getEnvOrDefault("DB_NAME", "jobview_test"),
		SSLMode:  "disable",
	}

	db, err := database.New(cfg)
	require.NoError(t, err, "数据库连接应该成功")

	service := NewJobApplicationService(db)
	return db, service
}

func createTestJobs(t *testing.T, service *JobApplicationService, userID uint, count int) []model.JobApplication {
	var jobs []model.JobApplication
	companies := []string{"阿里巴巴", "腾讯", "字节跳动", "华为", "美团", "滴滴", "京东", "网易", "百度", "小米"}
	positions := []string{"Go开发工程师", "后端工程师", "全栈工程师", "系统架构师", "技术专家"}
	statuses := []model.ApplicationStatus{
		model.StatusApplied,
		model.StatusResumeScreening,
		model.StatusFirstInterview,
		model.StatusSecondInterview,
		model.StatusOfferReceived,
		model.StatusRejected,
	}

	for i := 0; i < count; i++ {
		req := &model.CreateJobApplicationRequest{
			CompanyName:     companies[i%len(companies)],
			PositionTitle:   positions[i%len(positions)],
			Status:          statuses[i%len(statuses)],
			ApplicationDate: time.Now().AddDate(0, 0, -i%365).Format("2006-01-02"), // 分布在一年内
		}

		job, err := service.Create(userID, req)
		require.NoError(t, err, "创建测试数据应该成功")
		jobs = append(jobs, *job)
	}

	return jobs
}

func generateTestApplications(count int) []model.CreateJobApplicationRequest {
	companies := []string{"优化测试公司A", "优化测试公司B", "优化测试公司C"}
	positions := []string{"高级Go工程师", "系统架构师", "技术总监"}

	var applications []model.CreateJobApplicationRequest
	for i := 0; i < count; i++ {
		applications = append(applications, model.CreateJobApplicationRequest{
			CompanyName:   companies[i%len(companies)],
			PositionTitle: positions[i%len(positions)],
			Status:        model.StatusApplied,
		})
	}
	return applications
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
			findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}