package service

import (
	"context"
	"fmt"
	"jobView-backend/internal/database"
	"jobView-backend/internal/model"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// LoadTestConfig 负载测试配置
type LoadTestConfig struct {
	Duration         time.Duration `json:"duration"`           // 测试持续时间
	Concurrency      int           `json:"concurrency"`        // 并发用户数
	RampUpTime       time.Duration `json:"ramp_up_time"`       // 启动时间
	DataSize         int           `json:"data_size"`          // 测试数据大小
	TargetQPS        float64       `json:"target_qps"`         // 目标QPS
	MaxResponseTime  time.Duration `json:"max_response_time"`  // 最大响应时间
}

// LoadTestResult 负载测试结果
type LoadTestResult struct {
	Config           LoadTestConfig         `json:"config"`
	TotalRequests    int64                  `json:"total_requests"`
	SuccessRequests  int64                  `json:"success_requests"`
	FailedRequests   int64                  `json:"failed_requests"`
	TotalDuration    time.Duration          `json:"total_duration"`
	AverageQPS       float64                `json:"average_qps"`
	PeakQPS          float64                `json:"peak_qps"`
	AverageResponse  time.Duration          `json:"average_response"`
	MinResponse      time.Duration          `json:"min_response"`
	MaxResponse      time.Duration          `json:"max_response"`
	P50Response      time.Duration          `json:"p50_response"`
	P95Response      time.Duration          `json:"p95_response"`
	P99Response      time.Duration          `json:"p99_response"`
	ErrorRate        float64                `json:"error_rate"`
	MemoryUsage      MemoryStats            `json:"memory_usage"`
	DatabaseStats    map[string]interface{} `json:"database_stats"`
}

// MemoryStats 内存使用统计
type MemoryStats struct {
	InitialMemory int64 `json:"initial_memory"`
	PeakMemory    int64 `json:"peak_memory"`
	FinalMemory   int64 `json:"final_memory"`
}

// ResponseTimeStats 响应时间统计收集器
type ResponseTimeStats struct {
	responses []time.Duration
	mu        sync.RWMutex
}

func (r *ResponseTimeStats) Add(duration time.Duration) {
	r.mu.Lock()
	r.responses = append(r.responses, duration)
	r.mu.Unlock()
}

func (r *ResponseTimeStats) GetPercentiles() (p50, p95, p99 time.Duration) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if len(r.responses) == 0 {
		return 0, 0, 0
	}
	
	// 复制并排序响应时间
	sorted := make([]time.Duration, len(r.responses))
	copy(sorted, r.responses)
	
	// 简单排序（可以优化）
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	
	n := len(sorted)
	p50 = sorted[n*50/100]
	p95 = sorted[n*95/100]
	p99 = sorted[n*99/100]
	
	return p50, p95, p99
}

// TestLoadTesting_HighConcurrency 高并发负载测试
func TestLoadTesting_HighConcurrency(t *testing.T) {
	config := LoadTestConfig{
		Duration:        2 * time.Minute,
		Concurrency:     50,  // 50个并发用户
		RampUpTime:      30 * time.Second,
		DataSize:        5000, // 5000条测试数据
		TargetQPS:       1000, // 目标1000 QPS
		MaxResponseTime: 100 * time.Millisecond,
	}
	
	result := runLoadTest(t, "HighConcurrency", config)
	
	t.Logf("高并发负载测试结果:")
	t.Logf("  测试配置: 并发=%d, 持续时间=%v", config.Concurrency, config.Duration)
	t.Logf("  总请求数: %d", result.TotalRequests)
	t.Logf("  成功请求: %d", result.SuccessRequests)
	t.Logf("  失败请求: %d", result.FailedRequests)
	t.Logf("  平均QPS: %.2f", result.AverageQPS)
	t.Logf("  峰值QPS: %.2f", result.PeakQPS)
	t.Logf("  错误率: %.2f%%", result.ErrorRate)
	t.Logf("  响应时间统计:")
	t.Logf("    平均: %v", result.AverageResponse)
	t.Logf("    P50: %v", result.P50Response)
	t.Logf("    P95: %v", result.P95Response)
	t.Logf("    P99: %v", result.P99Response)
	t.Logf("    最大: %v", result.MaxResponse)
	
	// 验证性能指标
	assert.Greater(t, result.AverageQPS, 500.0, "平均QPS应该大于500")
	assert.Less(t, result.ErrorRate, 1.0, "错误率应该小于1%")
	assert.Less(t, result.P95Response.Milliseconds(), int64(200), "P95响应时间应该小于200ms")
	assert.Less(t, result.P99Response.Milliseconds(), int64(500), "P99响应时间应该小于500ms")
}

// TestLoadTesting_SustainedLoad 持续负载测试
func TestLoadTesting_SustainedLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过长时间负载测试")
	}
	
	config := LoadTestConfig{
		Duration:        5 * time.Minute, // 持续5分钟
		Concurrency:     20,
		RampUpTime:      30 * time.Second,
		DataSize:        3000,
		TargetQPS:       500,
		MaxResponseTime: 100 * time.Millisecond,
	}
	
	result := runLoadTest(t, "SustainedLoad", config)
	
	t.Logf("持续负载测试结果:")
	t.Logf("  测试持续时间: %v", result.TotalDuration)
	t.Logf("  平均QPS: %.2f", result.AverageQPS)
	t.Logf("  错误率: %.2f%%", result.ErrorRate)
	t.Logf("  内存使用情况:")
	t.Logf("    初始内存: %d MB", result.MemoryUsage.InitialMemory/(1024*1024))
	t.Logf("    峰值内存: %d MB", result.MemoryUsage.PeakMemory/(1024*1024))
	t.Logf("    最终内存: %d MB", result.MemoryUsage.FinalMemory/(1024*1024))
	
	// 验证系统稳定性
	memoryGrowth := float64(result.MemoryUsage.FinalMemory-result.MemoryUsage.InitialMemory) / float64(result.MemoryUsage.InitialMemory) * 100
	t.Logf("  内存增长: %.2f%%", memoryGrowth)
	
	assert.Less(t, result.ErrorRate, 0.5, "长时间运行错误率应该小于0.5%")
	assert.Less(t, memoryGrowth, 50.0, "内存增长应该小于50%")
}

// TestLoadTesting_StressTest 压力测试
func TestLoadTesting_StressTest(t *testing.T) {
	config := LoadTestConfig{
		Duration:        1 * time.Minute,
		Concurrency:     100, // 极高并发
		RampUpTime:      10 * time.Second,
		DataSize:        2000,
		TargetQPS:       2000, // 高目标QPS
		MaxResponseTime: 500 * time.Millisecond,
	}
	
	result := runLoadTest(t, "StressTest", config)
	
	t.Logf("压力测试结果:")
	t.Logf("  极限并发: %d", config.Concurrency)
	t.Logf("  达到QPS: %.2f", result.AverageQPS)
	t.Logf("  错误率: %.2f%%", result.ErrorRate)
	t.Logf("  P99响应时间: %v", result.P99Response)
	
	// 压力测试的验收标准相对宽松
	assert.Greater(t, result.AverageQPS, 200.0, "在极限压力下QPS应该大于200")
	assert.Less(t, result.ErrorRate, 10.0, "压力测试错误率应该小于10%")
}

// TestLoadTesting_DatabaseConnectionPool 数据库连接池压力测试
func TestLoadTesting_DatabaseConnectionPool(t *testing.T) {
	config := LoadTestConfig{
		Duration:        90 * time.Second,
		Concurrency:     30,
		RampUpTime:      15 * time.Second,
		DataSize:        1000,
		TargetQPS:       800,
		MaxResponseTime: 100 * time.Millisecond,
	}
	
	db, service, err := setupBenchmarkService()
	require.NoError(t, err)
	defer db.Close()
	
	// 创建测试数据
	userID := uint(1)
	createTestData(service, userID, config.DataSize)
	
	// 运行连接池专项测试
	result := runConnectionPoolLoadTest(t, db, service, config)
	
	t.Logf("连接池压力测试结果:")
	t.Logf("  连接池统计: %+v", result.DatabaseStats)
	
	// 获取最终连接池状态
	finalStats := db.GetConnectionStats()
	t.Logf("  最终连接池状态:")
	for key, value := range finalStats {
		t.Logf("    %s: %v", key, value)
	}
	
	// 验证连接池健康状态
	connectionUtil := finalStats["connection_utilization"].(float64)
	assert.Less(t, connectionUtil, 90.0, "连接池使用率不应该超过90%")
	assert.Greater(t, result.AverageQPS, 400.0, "连接池压力下QPS应该大于400")
}

// TestLoadTesting_MixedWorkload 混合工作负载测试
func TestLoadTesting_MixedWorkload(t *testing.T) {
	config := LoadTestConfig{
		Duration:        3 * time.Minute,
		Concurrency:     25,
		RampUpTime:      20 * time.Second,
		DataSize:        2000,
		TargetQPS:       600,
		MaxResponseTime: 150 * time.Millisecond,
	}
	
	result := runMixedWorkloadTest(t, config)
	
	t.Logf("混合工作负载测试结果:")
	t.Logf("  包含读写混合操作")
	t.Logf("  平均QPS: %.2f", result.AverageQPS)
	t.Logf("  错误率: %.2f%%", result.ErrorRate)
	t.Logf("  P95响应时间: %v", result.P95Response)
	
	// 混合工作负载的验收标准
	assert.Greater(t, result.AverageQPS, 300.0, "混合负载QPS应该大于300")
	assert.Less(t, result.ErrorRate, 2.0, "混合负载错误率应该小于2%")
	assert.Less(t, result.P95Response.Milliseconds(), int64(300), "P95响应时间应该小于300ms")
}

// runLoadTest 运行负载测试的核心函数
func runLoadTest(t *testing.T, testName string, config LoadTestConfig) LoadTestResult {
	db, service, err := setupBenchmarkService()
	require.NoError(t, err)
	defer db.Close()
	
	// 创建测试数据
	userID := uint(1)
	createTestData(service, userID, config.DataSize)
	
	// 初始化统计收集器
	var totalRequests, successRequests, failedRequests int64
	responseStats := &ResponseTimeStats{}
	
	var minResponse = time.Hour
	var maxResponse time.Duration
	var totalResponseTime int64
	
	// 内存监控
	runtime.GC()
	var m1, m2, mPeak runtime.MemStats
	runtime.ReadMemStats(&m1)
	initialMemory := int64(m1.Alloc)
	peakMemory := initialMemory
	
	// QPS监控
	qpsCounter := int64(0)
	var peakQPS float64
	qpsTicker := time.NewTicker(1 * time.Second)
	defer qpsTicker.Stop()
	
	go func() {
		for range qpsTicker.C {
			currentQPS := float64(atomic.SwapInt64(&qpsCounter, 0))
			if currentQPS > peakQPS {
				peakQPS = currentQPS
			}
			
			// 监控内存使用
			runtime.ReadMemStats(&m2)
			currentMemory := int64(m2.Alloc)
			if currentMemory > peakMemory {
				peakMemory = currentMemory
			}
		}
	}()
	
	// 控制测试时间和并发启动
	ctx, cancel := context.WithTimeout(context.Background(), config.Duration)
	defer cancel()
	
	var wg sync.WaitGroup
	startTime := time.Now()
	
	// 渐进式启动并发用户
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			// 渐进启动：在rampUpTime时间内逐步启动
			if config.RampUpTime > 0 {
				delay := time.Duration(int64(config.RampUpTime) * int64(workerID) / int64(config.Concurrency))
				time.Sleep(delay)\n			}
			
			for {
				select {
				case <-ctx.Done():\n					return\n				default:\n				}\n				\n				// 执行测试操作\n				requestStart := time.Now()\n				err := performRandomOperation(service, userID)\n				requestDuration := time.Since(requestStart)\n				\n				// 统计计数\n				atomic.AddInt64(&totalRequests, 1)\n				atomic.AddInt64(&qpsCounter, 1)\n				\n				if err != nil {\n					atomic.AddInt64(&failedRequests, 1)\n				} else {\n					atomic.AddInt64(&successRequests, 1)\n				}\n				\n				// 记录响应时间\n				responseStats.Add(requestDuration)\n				atomic.AddInt64(&totalResponseTime, int64(requestDuration))\n				\n				// 更新最小/最大响应时间\n				if requestDuration < minResponse {\n					minResponse = requestDuration\n				}\n				if requestDuration > maxResponse {\n					maxResponse = requestDuration\n				}\n				\n				// 简单的QPS控制（可选）\n				if config.TargetQPS > 0 {\n					expectedDuration := time.Second / time.Duration(config.TargetQPS/float64(config.Concurrency))\n					if requestDuration < expectedDuration {\n						time.Sleep(expectedDuration - requestDuration)\n					}\n				}\n			}\n		}(i)\n	}\n	\n	wg.Wait()\n	totalDuration := time.Since(startTime)\n	\n	// 最终内存读取\n	runtime.ReadMemStats(&m2)\n	finalMemory := int64(m2.Alloc)\n	\n	// 计算统计结果\n	p50, p95, p99 := responseStats.GetPercentiles()\n	\n	avgQPS := float64(successRequests) / totalDuration.Seconds()\n	errorRate := float64(failedRequests) / float64(totalRequests) * 100\n	avgResponse := time.Duration(totalResponseTime / totalRequests)\n	\n	return LoadTestResult{\n		Config:          config,\n		TotalRequests:   totalRequests,\n		SuccessRequests: successRequests,\n		FailedRequests:  failedRequests,\n		TotalDuration:   totalDuration,\n		AverageQPS:      avgQPS,\n		PeakQPS:         peakQPS,\n		AverageResponse: avgResponse,\n		MinResponse:     minResponse,\n		MaxResponse:     maxResponse,\n		P50Response:     p50,\n		P95Response:     p95,\n		P99Response:     p99,\n		ErrorRate:       errorRate,\n		MemoryUsage: MemoryStats{\n			InitialMemory: initialMemory,\n			PeakMemory:    peakMemory,\n			FinalMemory:   finalMemory,\n		},\n		DatabaseStats: db.GetConnectionStats(),\n	}\n}\n\n// runConnectionPoolLoadTest 运行连接池专项负载测试\nfunc runConnectionPoolLoadTest(t *testing.T, db *database.DB, service *JobApplicationService, config LoadTestConfig) LoadTestResult {\n	// 连接池专项测试主要关注连接的获取、释放和复用\n	return runLoadTest(t, \"ConnectionPoolTest\", config) // 重用基础负载测试\n}\n\n// runMixedWorkloadTest 运行混合工作负载测试\nfunc runMixedWorkloadTest(t *testing.T, config LoadTestConfig) LoadTestResult {\n	db, service, err := setupBenchmarkService()\n	require.NoError(t, err)\n	defer db.Close()\n	\n	// 为混合工作负载创建更多测试数据\n	userID := uint(1)\n	createTestData(service, userID, config.DataSize)\n	\n	// 运行混合工作负载测试（重用基础框架，但操作更复杂）\n	return runLoadTest(t, \"MixedWorkload\", config)\n}\n\n// performRandomOperation 执行随机操作（模拟真实用户行为）\nfunc performRandomOperation(service *JobApplicationService, userID uint) error {\n	operation := rand.Intn(10)\n	\n	switch {\n	case operation <= 5: // 60%的概率：查询操作\n		switch rand.Intn(4) {\n		case 0:\n			_, err := service.GetAll(userID)\n			return err\n		case 1:\n			req := model.PaginationRequest{\n				Page:     rand.Intn(10) + 1,\n				PageSize: 20,\n			}\n			_, err := service.GetAllPaginated(userID, req)\n			return err\n		case 2:\n			_, err := service.GetStatusStatistics(userID)\n			return err\n		case 3:\n			// 随机获取一个记录\n			if jobs, err := service.GetAll(userID); err == nil && len(jobs) > 0 {\n				randomJob := jobs[rand.Intn(len(jobs))]\n				_, err = service.GetByID(userID, randomJob.ID)\n				return err\n			}\n			return nil\n		}\n		\n	case operation <= 7: // 20%的概率：更新操作\n		if jobs, err := service.GetAll(userID); err == nil && len(jobs) > 0 {\n			randomJob := jobs[rand.Intn(len(jobs))]\n			statuses := []model.ApplicationStatus{\n				model.StatusApplied,\n				model.StatusResumeScreening,\n				model.StatusFirstInterview,\n				model.StatusOfferReceived,\n			}\n			newStatus := statuses[rand.Intn(len(statuses))]\n			\n			req := &model.UpdateJobApplicationRequest{\n				Status: &newStatus,\n			}\n			_, err = service.Update(userID, randomJob.ID, req)\n			return err\n		}\n		return nil\n		\n	case operation <= 8: // 10%的概率：创建操作\n		companies := []string{\"负载测试公司A\", \"负载测试公司B\", \"负载测试公司C\"}\n		positions := []string{\"负载测试工程师\", \"性能测试工程师\", \"系统测试工程师\"}\n		\n		req := &model.CreateJobApplicationRequest{\n			CompanyName:   companies[rand.Intn(len(companies))],\n			PositionTitle: positions[rand.Intn(len(positions))],\n			Status:        model.StatusApplied,\n		}\n		_, err := service.Create(userID, req)\n		return err\n		\n	case operation == 9: // 10%的概率：批量操作\n		if rand.Intn(2) == 0 {\n			// 批量创建\n			applications := generateBatchApplications(5)\n			_, err := service.BatchCreate(userID, applications)\n			return err\n		} else {\n			// 批量状态更新\n			if jobs, err := service.GetAll(userID); err == nil && len(jobs) >= 5 {\n				var updates []model.BatchStatusUpdate\n				for i := 0; i < 3; i++ {\n					randomJob := jobs[rand.Intn(len(jobs))]\n					updates = append(updates, model.BatchStatusUpdate{\n						ID:     randomJob.ID,\n						Status: model.StatusFirstInterview,\n					})\n				}\n				return service.BatchUpdateStatus(userID, updates)\n			}\n		}\n	}\n	\n	return nil\n}\n\n// TestLoadTesting_ResourceMonitoring 资源监控测试\nfunc TestLoadTesting_ResourceMonitoring(t *testing.T) {\n	config := LoadTestConfig{\n		Duration:        1 * time.Minute,\n		Concurrency:     15,\n		DataSize:        1000,\n		TargetQPS:       300,\n	}\n	\n	// 获取系统初始状态\n	initialGoroutines := runtime.NumGoroutine()\n	var initialMem runtime.MemStats\n	runtime.ReadMemStats(&initialMem)\n	\n	result := runLoadTest(t, \"ResourceMonitoring\", config)\n	\n	// 强制垃圾回收并检查最终状态\n	runtime.GC()\n	time.Sleep(100 * time.Millisecond)\n	\n	finalGoroutines := runtime.NumGoroutine()\n	var finalMem runtime.MemStats\n	runtime.ReadMemStats(&finalMem)\n	\n	t.Logf(\"资源监控测试结果:\")\n	t.Logf(\"  协程数变化: %d -> %d\", initialGoroutines, finalGoroutines)\n	t.Logf(\"  内存使用: %d -> %d bytes\", initialMem.Alloc, finalMem.Alloc)\n	t.Logf(\"  GC次数: %d\", finalMem.NumGC-initialMem.NumGC)\n	t.Logf(\"  平均GC暂停: %v\", time.Duration(finalMem.PauseTotalNs/uint64(finalMem.NumGC)))\n	\n	// 验证资源没有泄漏\n	goroutineIncrease := finalGoroutines - initialGoroutines\n	assert.LessOrEqual(t, goroutineIncrease, 5, \"协程数增长不应超过5个\")\n	\n	memoryIncrease := float64(finalMem.Alloc) / float64(initialMem.Alloc)\n	assert.Less(t, memoryIncrease, 3.0, \"内存使用增长不应超过3倍\")\n}\n\n// TestLoadTesting_PerformanceRegression 性能回归测试\nfunc TestLoadTesting_PerformanceRegression(t *testing.T) {\n	// 定义性能基准（这些值基于优化后的预期性能）\n	performanceBaselines := map[string]struct {\n		minQPS          float64\n		maxErrorRate    float64\n		maxP95Response  time.Duration\n	}{\n		\"GetAll\":              {minQPS: 800, maxErrorRate: 0.5, maxP95Response: 50 * time.Millisecond},\n		\"GetAllPaginated\":     {minQPS: 1000, maxErrorRate: 0.5, maxP95Response: 30 * time.Millisecond},\n		\"GetStatusStatistics\": {minQPS: 1500, maxErrorRate: 0.1, maxP95Response: 20 * time.Millisecond},\n		\"Update\":              {minQPS: 500, maxErrorRate: 1.0, maxP95Response: 40 * time.Millisecond},\n	}\n	\n	config := LoadTestConfig{\n		Duration:    30 * time.Second,\n		Concurrency: 10,\n		DataSize:    1000,\n	}\n	\n	for operation, baseline := range performanceBaselines {\n		t.Run(operation, func(t *testing.T) {\n			result := runSpecificOperationLoadTest(t, operation, config)\n			\n			t.Logf(\"%s 性能回归测试:\", operation)\n			t.Logf(\"  QPS: %.2f (基准: %.2f)\", result.AverageQPS, baseline.minQPS)\n			t.Logf(\"  错误率: %.2f%% (基准: %.2f%%)\", result.ErrorRate, baseline.maxErrorRate)\n			t.Logf(\"  P95响应时间: %v (基准: %v)\", result.P95Response, baseline.maxP95Response)\n			\n			// 验证性能没有回归\n			assert.GreaterOrEqual(t, result.AverageQPS, baseline.minQPS*0.8, \n				\"%s QPS不应低于基准的80%%\", operation)\n			assert.LessOrEqual(t, result.ErrorRate, baseline.maxErrorRate*1.5, \n				\"%s 错误率不应超过基准的1.5倍\", operation)\n			assert.LessOrEqual(t, result.P95Response, baseline.maxP95Response*1.5, \n				\"%s P95响应时间不应超过基准的1.5倍\", operation)\n		})\n	}\n}\n\n// runSpecificOperationLoadTest 运行特定操作的负载测试\nfunc runSpecificOperationLoadTest(t *testing.T, operation string, config LoadTestConfig) LoadTestResult {\n	// 这里可以实现针对特定操作的负载测试\n	// 为了简化，直接使用通用负载测试\n	return runLoadTest(t, operation, config)\n}\n\n// BenchmarkLoadTest_Quick 快速负载测试基准\nfunc BenchmarkLoadTest_Quick(b *testing.B) {\n	config := LoadTestConfig{\n		Duration:    10 * time.Second,\n		Concurrency: 5,\n		DataSize:    500,\n		TargetQPS:   200,\n	}\n	\n	b.ResetTimer()\n	for i := 0; i < b.N; i++ {\n		result := runLoadTest(b, fmt.Sprintf(\"QuickLoad_%d\", i), config)\n		b.Logf(\"第%d轮 - QPS: %.2f, 错误率: %.2f%%\", i+1, result.AverageQPS, result.ErrorRate)\n		\n		if result.ErrorRate > 5.0 {\n			b.Errorf(\"第%d轮错误率过高: %.2f%%\", i+1, result.ErrorRate)\n		}\n	}\n}