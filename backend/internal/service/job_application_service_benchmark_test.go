package service

import (
	"fmt"
	"jobView-backend/internal/config"
	"jobView-backend/internal/database"
	"jobView-backend/internal/model"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"
)

// BenchmarkJobApplicationService 数据库性能基准测试套件
// 运行命令: go test -bench=. -benchmem -count=5 ./internal/service/

// setupBenchmarkService 创建基准测试服务
func setupBenchmarkService(b *testing.B) (*JobApplicationService, func()) {
	// 加载测试配置
	cfg := config.Load()
	cfg.Database.DBName = cfg.Database.DBName + "_test"
	
	// 创建测试数据库连接
	db, err := database.New(&cfg.Database)
	if err != nil {
		b.Fatalf("Failed to connect to test database: %v", err)
	}
	
	service := NewJobApplicationService(db)
	
	// 返回清理函数
	cleanup := func() {
		if db != nil {
			db.Close()
		}
	}
	
	return service, cleanup
}

// setupTestData 创建测试数据
func setupTestData(service *JobApplicationService, userID uint, count int) error {
	applications := make([]model.CreateJobApplicationRequest, count)
	
	companies := []string{
		"阿里巴巴", "腾讯", "字节跳动", "百度", "京东", "美团", "滴滴", "小米",
		"华为", "网易", "新浪", "搜狐", "360", "快手", "拼多多", "蚂蚁金服",
	}
	
	positions := []string{
		"Go后端工程师", "Java后端工程师", "前端工程师", "全栈工程师", "数据工程师",
		"算法工程师", "产品经理", "UI设计师", "测试工程师", "运维工程师",
	}
	
	statuses := []model.ApplicationStatus{
		model.StatusApplied, model.StatusResumeScreening, model.StatusWrittenTest,
		model.StatusFirstInterview, model.StatusSecondInterview, model.StatusHRInterview,
		model.StatusOfferReceived, model.StatusRejected,
	}
	
	for i := 0; i < count; i++ {
		// 随机生成测试数据
		companyIndex := rand.Intn(len(companies))
		positionIndex := rand.Intn(len(positions))
		statusIndex := rand.Intn(len(statuses))
		
		// 生成随机日期 (最近一年内)
		randomDate := time.Now().AddDate(0, 0, -rand.Intn(365))
		
		applications[i] = model.CreateJobApplicationRequest{
			CompanyName:     companies[companyIndex],
			PositionTitle:   positions[positionIndex],
			ApplicationDate: randomDate.Format("2006-01-02"),
			Status:          statuses[statusIndex],
		}
	}
	
	// 批量插入测试数据
	_, err := service.BatchCreate(userID, applications)
	return err
}

// BenchmarkGetAll 基准测试 GetAll 方法
func BenchmarkGetAll(b *testing.B) {
	service, cleanup := setupBenchmarkService(b)
	defer cleanup()
	
	userID := uint(1)
	
	// 创建测试数据
	if err := setupTestData(service, userID, 500); err != nil {
		b.Fatalf("Failed to setup test data: %v", err)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := service.GetAll(userID)
		if err != nil {
			b.Fatalf("GetAll failed: %v", err)
		}
	}
}

// BenchmarkGetAllPaginated 基准测试分页查询方法
func BenchmarkGetAllPaginated(b *testing.B) {
	service, cleanup := setupBenchmarkService(b)
	defer cleanup()
	
	userID := uint(1)
	
	// 创建测试数据
	if err := setupTestData(service, userID, 1000); err != nil {
		b.Fatalf("Failed to setup test data: %v", err)
	}
	
	req := model.PaginationRequest{
		Page:     1,
		PageSize: 20,
		SortBy:   "application_date",
		SortDir:  "DESC",
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := service.GetAllPaginated(userID, req)
		if err != nil {
			b.Fatalf("GetAllPaginated failed: %v", err)
		}
	}
}

// BenchmarkGetStatusStatistics 基准测试状态统计方法
func BenchmarkGetStatusStatistics(b *testing.B) {
	service, cleanup := setupBenchmarkService(b)
	defer cleanup()
	
	userID := uint(1)
	
	// 创建测试数据
	if err := setupTestData(service, userID, 1000); err != nil {
		b.Fatalf("Failed to setup test data: %v", err)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := service.GetStatusStatistics(userID)
		if err != nil {
			b.Fatalf("GetStatusStatistics failed: %v", err)
		}
	}
}

// BenchmarkSearchApplications 基准测试全文搜索方法
func BenchmarkSearchApplications(b *testing.B) {
	service, cleanup := setupBenchmarkService(b)
	defer cleanup()
	
	userID := uint(1)
	
	// 创建测试数据
	if err := setupTestData(service, userID, 1000); err != nil {
		b.Fatalf("Failed to setup test data: %v", err)
	}
	
	req := model.PaginationRequest{
		Page:     1,
		PageSize: 20,
		SortBy:   "application_date",
		SortDir:  "DESC",
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := service.SearchApplications(userID, "阿里巴巴", req)
		if err != nil {
			b.Fatalf("SearchApplications failed: %v", err)
		}
	}
}

// BenchmarkCreate 基准测试单条创建方法
func BenchmarkCreate(b *testing.B) {
	service, cleanup := setupBenchmarkService(b)
	defer cleanup()
	
	userID := uint(1)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		req := &model.CreateJobApplicationRequest{
			CompanyName:     fmt.Sprintf("测试公司%d", i),
			PositionTitle:   "Go后端工程师",
			ApplicationDate: time.Now().Format("2006-01-02"),
			Status:          model.StatusApplied,
		}
		
		_, err := service.Create(userID, req)
		if err != nil {
			b.Fatalf("Create failed: %v", err)
		}
	}
}

// BenchmarkBatchCreate 基准测试批量创建方法
func BenchmarkBatchCreate(b *testing.B) {
	service, cleanup := setupBenchmarkService(b)
	defer cleanup()
	
	userID := uint(1)
	
	// 准备批量数据 (每次插入50条)
	batchSize := 50
	applications := make([]model.CreateJobApplicationRequest, batchSize)
	
	for i := 0; i < batchSize; i++ {
		applications[i] = model.CreateJobApplicationRequest{
			CompanyName:     fmt.Sprintf("批量测试公司%d", i),
			PositionTitle:   "Go后端工程师",
			ApplicationDate: time.Now().Format("2006-01-02"),
			Status:          model.StatusApplied,
		}
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// 为每次基准测试生成不同的公司名称
		for j := range applications {
			applications[j].CompanyName = fmt.Sprintf("批量测试公司%d-%d", i, j)
		}
		
		_, err := service.BatchCreate(userID, applications)
		if err != nil {
			b.Fatalf("BatchCreate failed: %v", err)
		}
	}
}

// BenchmarkUpdate 基准测试更新方法
func BenchmarkUpdate(b *testing.B) {
	service, cleanup := setupBenchmarkService(b)
	defer cleanup()
	
	userID := uint(1)
	
	// 先创建一些测试数据
	if err := setupTestData(service, userID, 100); err != nil {
		b.Fatalf("Failed to setup test data: %v", err)
	}
	
	// 获取要更新的记录ID
	applications, err := service.GetAll(userID)
	if err != nil {
		b.Fatalf("Failed to get applications: %v", err)
	}
	
	if len(applications) == 0 {
		b.Skip("No applications to update")
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// 循环使用已存在的记录ID
		appIndex := i % len(applications)
		appID := applications[appIndex].ID
		
		newStatus := model.StatusFirstInterview
		req := &model.UpdateJobApplicationRequest{
			Status: &newStatus,
		}
		
		_, err := service.Update(userID, appID, req)
		if err != nil {
			b.Fatalf("Update failed: %v", err)
		}
	}
}

// BenchmarkConcurrentOperations 并发操作基准测试
func BenchmarkConcurrentOperations(b *testing.B) {
	service, cleanup := setupBenchmarkService(b)
	defer cleanup()
	
	userID := uint(1)
	
	// 创建基础测试数据
	if err := setupTestData(service, userID, 500); err != nil {
		b.Fatalf("Failed to setup test data: %v", err)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	// 并发执行不同类型的操作
	b.RunParallel(func(pb *testing.PB) {
		req := model.PaginationRequest{
			Page:     1,
			PageSize: 20,
			SortBy:   "application_date",
			SortDir:  "DESC",
		}
		
		for pb.Next() {
			// 随机选择操作类型
			switch rand.Intn(3) {
			case 0:
				// 查询操作
				service.GetAllPaginated(userID, req)
			case 1:
				// 统计操作
				service.GetStatusStatistics(userID)
			case 2:
				// 搜索操作
				service.SearchApplications(userID, "腾讯", req)
			}
		}
	})
}

// 数据库连接池性能测试
func BenchmarkConnectionPool(b *testing.B) {
	service, cleanup := setupBenchmarkService(b)
	defer cleanup()
	
	userID := uint(1)
	
	// 创建测试数据
	if err := setupTestData(service, userID, 100); err != nil {
		b.Fatalf("Failed to setup test data: %v", err)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	// 测试连接池在高并发下的表现
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := service.GetStatusStatistics(userID)
			if err != nil {
				b.Errorf("GetStatusStatistics failed: %v", err)
			}
		}
	})
}

// 测试主函数 - 运行所有基准测试并输出结果
func TestMain(m *testing.M) {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())
	
	// 设置日志
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	// 运行测试
	code := m.Run()
	
	os.Exit(code)
}

// 性能对比测试 - 优化前后对比
func BenchmarkPerformanceComparison(b *testing.B) {
	service, cleanup := setupBenchmarkService(b)
	defer cleanup()
	
	userID := uint(1)
	
	// 创建大量测试数据以观察性能差异
	if err := setupTestData(service, userID, 2000); err != nil {
		b.Fatalf("Failed to setup test data: %v", err)
	}
	
	b.Run("GetAll_WithLimit", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := service.GetAll(userID)
			if err != nil {
				b.Fatalf("GetAll failed: %v", err)
			}
		}
	})
	
	b.Run("GetAllPaginated_WithIndex", func(b *testing.B) {
		req := model.PaginationRequest{
			Page:     1,
			PageSize: 500, // 相同的数据量
			SortBy:   "application_date",
			SortDir:  "DESC",
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := service.GetAllPaginated(userID, req)
			if err != nil {
				b.Fatalf("GetAllPaginated failed: %v", err)
			}
		}
	})
	
	b.Run("StatusStatistics_WithCoveringIndex", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := service.GetStatusStatistics(userID)
			if err != nil {
				b.Fatalf("GetStatusStatistics failed: %v", err)
			}
		}
	})
}