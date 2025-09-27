//go:build integration
// +build integration

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// JobView状态跟踪功能API测试套件
// 测试工程师: 🧪 PACT Tester
// 创建时间: 2025-09-08
// 版本: 1.0

const (
	BaseURL = "http://localhost:8010"
	TestEmail = "api_test@example.com"
	TestPassword = "TestPass123!"
)

// 测试用例结构
type APITestCase struct {
	Name           string
	Method         string
	Endpoint       string
	Headers        map[string]string
	Payload        interface{}
	ExpectedStatus int
	ExpectedFields []string
	Description    string
}

// API响应结构
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 认证响应结构
type AuthResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	} `json:"user"`
}

// 状态更新请求结构
type StatusUpdateRequest struct {
	Status   string            `json:"status"`
	Note     string            `json:"note,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Version  int               `json:"version,omitempty"`
}

// 状态历史响应结构
type StatusHistoryResponse struct {
	History []StatusHistoryEntry `json:"history"`
	Total   int                  `json:"total"`
	Page    int                  `json:"page"`
	HasNext bool                 `json:"has_next"`
}

type StatusHistoryEntry struct {
	ID               int               `json:"id"`
	OldStatus        string            `json:"old_status"`
	NewStatus        string            `json:"new_status"`
	StatusChangedAt  string            `json:"status_changed_at"`
	DurationMinutes  int               `json:"duration_minutes"`
	Metadata         map[string]string `json:"metadata"`
}

// 全局变量
var (
	authToken  string
	testUserID int
	testJobIDs []int
)

// 测试主函数
func TestMain(m *testing.M) {
	fmt.Println("===========================================")
	fmt.Println("开始JobView状态跟踪功能API测试")
	fmt.Println("===========================================")
	
	// 设置测试环境
	if !setupTestEnvironment() {
		fmt.Println("❌ 测试环境设置失败")
		os.Exit(1)
	}
	
	// 执行测试
	code := m.Run()
	
	// 清理测试数据
	cleanupTestData()
	
	fmt.Println("===========================================")
	fmt.Println("JobView状态跟踪功能API测试完成")
	fmt.Println("===========================================")
	
	os.Exit(code)
}

// 设置测试环境
func setupTestEnvironment() bool {
	fmt.Println("📋 设置测试环境...")
	
	// 等待服务启动
	if !waitForServer() {
		return false
	}
	
	// 注册测试用户
	if !registerTestUser() {
		return false
	}
	
	// 用户登录获取认证令牌
	if !loginTestUser() {
		return false
	}
	
	// 创建测试岗位申请
	if !createTestJobApplications() {
		return false
	}
	
	fmt.Println("✅ 测试环境设置完成")
	return true
}

// 等待服务启动
func waitForServer() bool {
	fmt.Println("⏳ 等待服务启动...")
	for i := 0; i < 30; i++ {
		resp, err := http.Get(BaseURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			fmt.Println("✅ 服务已启动")
			return true
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(time.Second)
	}
	fmt.Println("❌ 服务启动超时")
	return false
}

// 注册测试用户
func registerTestUser() bool {
	payload := map[string]string{
		"username": "api_test_user",
		"email":    TestEmail,
		"password": TestPassword,
	}
	
	_, err := makeRequest("POST", "/api/auth/register", nil, payload)
	if err != nil {
		log.Printf("注册用户可能已存在: %v", err)
	}
	
	return true
}

// 用户登录
func loginTestUser() bool {
	payload := map[string]string{
		"email":    TestEmail,
		"password": TestPassword,
	}
	
	resp, err := makeRequest("POST", "/api/auth/login", nil, payload)
	if err != nil {
		fmt.Printf("❌ 用户登录失败: %v\n", err)
		return false
	}
	
	var authResp AuthResponse
	if err := json.Unmarshal(resp, &authResp); err != nil {
		fmt.Printf("❌ 解析登录响应失败: %v\n", err)
		return false
	}
	
	authToken = authResp.Token
	testUserID = authResp.User.ID
	
	fmt.Printf("✅ 用户登录成功，用户ID: %d\n", testUserID)
	return true
}

// 创建测试岗位申请
func createTestJobApplications() bool {
	headers := map[string]string{
		"Authorization": "Bearer " + authToken,
	}
	
	testJobs := []map[string]interface{}{
		{
			"company_name":   "API测试公司A",
			"position_title": "高级Go工程师",
			"status":         "已投递",
		},
		{
			"company_name":   "API测试公司B",
			"position_title": "后端架构师",
			"status":         "简历筛选中",
		},
		{
			"company_name":   "API测试公司C",
			"position_title": "技术负责人",
			"status":         "一面中",
		},
	}
	
	for _, job := range testJobs {
		resp, err := makeRequest("POST", "/api/v1/job-applications", headers, job)
		if err != nil {
			fmt.Printf("❌ 创建测试岗位失败: %v\n", err)
			return false
		}
		
		var result map[string]interface{}
		if err := json.Unmarshal(resp, &result); err != nil {
			continue
		}
		
		if data, ok := result["data"].(map[string]interface{}); ok {
			if id, ok := data["id"].(float64); ok {
				testJobIDs = append(testJobIDs, int(id))
			}
		}
	}
	
	fmt.Printf("✅ 创建了%d个测试岗位申请\n", len(testJobIDs))
	return len(testJobIDs) > 0
}

// 通用请求函数
func makeRequest(method, endpoint string, headers map[string]string, payload interface{}) ([]byte, error) {
	var bodyReader io.Reader
	
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonData)
	}
	
	req, err := http.NewRequest(method, BaseURL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	return body, nil
}

// 测试1: 状态跟踪核心API
func TestStatusTrackingCoreAPIs(t *testing.T) {
	if len(testJobIDs) == 0 {
		t.Skip("跳过测试：没有可用的测试岗位")
	}
	
	jobID := testJobIDs[0]
	headers := map[string]string{
		"Authorization": "Bearer " + authToken,
	}
	
	t.Run("状态更新API测试", func(t *testing.T) {
		payload := StatusUpdateRequest{
			Status:   "简历筛选中",
			Note:     "HR确认收到简历",
			Metadata: map[string]string{
				"source":  "email",
				"hr_name": "张三",
			},
		}
		
		resp, err := makeRequest("POST", fmt.Sprintf("/api/v1/job-applications/%d/status", jobID), headers, payload)
		if err != nil {
			t.Fatalf("状态更新请求失败: %v", err)
		}
		
		var result APIResponse
		if err := json.Unmarshal(resp, &result); err != nil {
			t.Fatalf("响应解析失败: %v", err)
		}
		
		if result.Code != 200 {
			t.Errorf("状态更新失败: code=%d, message=%s", result.Code, result.Message)
		}
		
		t.Logf("✅ 状态更新成功")
	})
	
	t.Run("状态历史查询API测试", func(t *testing.T) {
		resp, err := makeRequest("GET", fmt.Sprintf("/api/v1/job-applications/%d/status-history?page=1&page_size=10", jobID), headers, nil)
		if err != nil {
			t.Fatalf("状态历史查询失败: %v", err)
		}
		
		var result APIResponse
		if err := json.Unmarshal(resp, &result); err != nil {
			t.Fatalf("响应解析失败: %v", err)
		}
		
		if result.Code != 200 {
			t.Errorf("状态历史查询失败: code=%d, message=%s", result.Code, result.Message)
		}
		
		// 验证返回数据结构
		dataBytes, _ := json.Marshal(result.Data)
		var historyResp StatusHistoryResponse
		if err := json.Unmarshal(dataBytes, &historyResp); err == nil {
			if len(historyResp.History) > 0 {
				t.Logf("✅ 状态历史查询成功，找到%d条记录", len(historyResp.History))
			} else {
				t.Log("⚠️  状态历史记录为空")
			}
		}
	})
	
	t.Run("状态时间轴API测试", func(t *testing.T) {
		resp, err := makeRequest("GET", fmt.Sprintf("/api/v1/job-applications/%d/status-timeline", jobID), headers, nil)
		if err != nil {
			t.Fatalf("状态时间轴查询失败: %v", err)
		}
		
		var result APIResponse
		if err := json.Unmarshal(resp, &result); err != nil {
			t.Fatalf("响应解析失败: %v", err)
		}
		
		if result.Code != 200 {
			t.Errorf("状态时间轴查询失败: code=%d, message=%s", result.Code, result.Message)
		}
		
		t.Logf("✅ 状态时间轴查询成功")
	})
}

// 测试2: 批量状态操作API
func TestBatchStatusOperations(t *testing.T) {
	if len(testJobIDs) < 2 {
		t.Skip("跳过测试：测试岗位数量不足")
	}
	
	headers := map[string]string{
		"Authorization": "Bearer " + authToken,
	}
	
	t.Run("批量状态更新API测试", func(t *testing.T) {
		payload := map[string]interface{}{
			"job_ids": testJobIDs[:2],
			"status":  "简历筛选中",
			"note":    "批量更新测试",
		}
		
		resp, err := makeRequest("PUT", "/api/v1/job-applications/status/batch", headers, payload)
		if err != nil {
			t.Fatalf("批量状态更新请求失败: %v", err)
		}
		
		var result APIResponse
		if err := json.Unmarshal(resp, &result); err != nil {
			t.Fatalf("响应解析失败: %v", err)
		}
		
		// 批量操作可能返回部分成功
		if result.Code != 200 && result.Code != 207 {
			t.Errorf("批量状态更新失败: code=%d, message=%s", result.Code, result.Message)
		} else {
			t.Logf("✅ 批量状态更新成功")
		}
	})
}

// 测试3: 状态配置管理API
func TestStatusConfigAPIs(t *testing.T) {
	headers := map[string]string{
		"Authorization": "Bearer " + authToken,
	}
	
	t.Run("状态定义API测试", func(t *testing.T) {
		resp, err := makeRequest("GET", "/api/v1/status-definitions", headers, nil)
		if err != nil {
			t.Fatalf("状态定义查询失败: %v", err)
		}
		
		var result APIResponse
		if err := json.Unmarshal(resp, &result); err != nil {
			t.Fatalf("响应解析失败: %v", err)
		}
		
		if result.Code != 200 {
			t.Errorf("状态定义查询失败: code=%d, message=%s", result.Code, result.Message)
		}
		
		t.Logf("✅ 状态定义查询成功")
	})
	
	t.Run("状态流转模板API测试", func(t *testing.T) {
		resp, err := makeRequest("GET", "/api/v1/status-flow-templates", headers, nil)
		if err != nil {
			t.Fatalf("状态流转模板查询失败: %v", err)
		}
		
		var result APIResponse
		if err := json.Unmarshal(resp, &result); err != nil {
			t.Fatalf("响应解析失败: %v", err)
		}
		
		if result.Code != 200 {
			t.Errorf("状态流转模板查询失败: code=%d, message=%s", result.Code, result.Message)
		}
		
		t.Logf("✅ 状态流转模板查询成功")
	})
	
	t.Run("用户状态偏好API测试", func(t *testing.T) {
		resp, err := makeRequest("GET", "/api/v1/user-status-preferences", headers, nil)
		if err != nil {
			t.Fatalf("用户状态偏好查询失败: %v", err)
		}
		
		var result APIResponse
		if err := json.Unmarshal(resp, &result); err != nil {
			t.Fatalf("响应解析失败: %v", err)
		}
		
		if result.Code != 200 {
			t.Errorf("用户状态偏好查询失败: code=%d, message=%s", result.Code, result.Message)
		}
		
		t.Logf("✅ 用户状态偏好查询成功")
	})
}

// 测试4: 数据分析API
func TestAnalyticsAPIs(t *testing.T) {
	headers := map[string]string{
		"Authorization": "Bearer " + authToken,
	}
	
	t.Run("状态分析API测试", func(t *testing.T) {
		resp, err := makeRequest("GET", "/api/v1/job-applications/status-analytics", headers, nil)
		if err != nil {
			t.Fatalf("状态分析查询失败: %v", err)
		}
		
		var result APIResponse
		if err := json.Unmarshal(resp, &result); err != nil {
			t.Fatalf("响应解析失败: %v", err)
		}
		
		if result.Code != 200 {
			t.Errorf("状态分析查询失败: code=%d, message=%s", result.Code, result.Message)
		}
		
		t.Logf("✅ 状态分析查询成功")
	})
	
	t.Run("状态趋势API测试", func(t *testing.T) {
		resp, err := makeRequest("GET", "/api/v1/job-applications/status-trends?days=30", headers, nil)
		if err != nil {
			t.Fatalf("状态趋势查询失败: %v", err)
		}
		
		var result APIResponse
		if err := json.Unmarshal(resp, &result); err != nil {
			t.Fatalf("响应解析失败: %v", err)
		}
		
		if result.Code != 200 {
			t.Errorf("状态趋势查询失败: code=%d, message=%s", result.Code, result.Message)
		}
		
		t.Logf("✅ 状态趋势查询成功")
	})
}

// 测试5: 错误处理和边界条件
func TestErrorHandling(t *testing.T) {
	headers := map[string]string{
		"Authorization": "Bearer " + authToken,
	}
	
	t.Run("无效状态更新测试", func(t *testing.T) {
		if len(testJobIDs) == 0 {
			t.Skip("跳过测试：没有可用的测试岗位")
		}
		
		payload := StatusUpdateRequest{
			Status: "无效状态",
		}
		
		resp, err := makeRequest("POST", fmt.Sprintf("/api/v1/job-applications/%d/status", testJobIDs[0]), headers, payload)
		if err != nil {
			t.Fatalf("请求发送失败: %v", err)
		}
		
		var result APIResponse
		if err := json.Unmarshal(resp, &result); err != nil {
			t.Fatalf("响应解析失败: %v", err)
		}
		
		if result.Code == 200 {
			t.Errorf("无效状态更新应该失败，但返回成功")
		} else {
			t.Logf("✅ 无效状态更新被正确拒绝: %s", result.Message)
		}
	})
	
	t.Run("不存在的岗位ID测试", func(t *testing.T) {
		resp, err := makeRequest("GET", "/api/v1/job-applications/999999/status-history", headers, nil)
		if err != nil {
			t.Fatalf("请求发送失败: %v", err)
		}
		
		var result APIResponse
		if err := json.Unmarshal(resp, &result); err != nil {
			t.Fatalf("响应解析失败: %v", err)
		}
		
		if result.Code == 200 {
			t.Errorf("不存在的岗位ID应该返回错误")
		} else {
			t.Logf("✅ 不存在的岗位ID被正确处理: %s", result.Message)
		}
	})
	
	t.Run("未授权访问测试", func(t *testing.T) {
		// 不提供认证令牌
		resp, err := makeRequest("GET", "/api/v1/status-definitions", nil, nil)
		if err != nil {
			t.Fatalf("请求发送失败: %v", err)
		}
		
		var result APIResponse
		if err := json.Unmarshal(resp, &result); err != nil {
			t.Fatalf("响应解析失败: %v", err)
		}
		
		if result.Code != 401 {
			t.Errorf("未授权访问应该返回401，实际返回: %d", result.Code)
		} else {
			t.Logf("✅ 未授权访问被正确拒绝")
		}
	})
}

// 测试6: API性能测试
func TestAPIPerformance(t *testing.T) {
	if len(testJobIDs) == 0 {
		t.Skip("跳过测试：没有可用的测试岗位")
	}
	
	headers := map[string]string{
		"Authorization": "Bearer " + authToken,
	}
	
	t.Run("API响应时间测试", func(t *testing.T) {
		testCases := []struct {
			name     string
			endpoint string
			maxTime  time.Duration
		}{
			{"状态历史查询", fmt.Sprintf("/api/v1/job-applications/%d/status-history", testJobIDs[0]), 200 * time.Millisecond},
			{"状态分析查询", "/api/v1/job-applications/status-analytics", 300 * time.Millisecond},
			{"状态定义查询", "/api/v1/status-definitions", 100 * time.Millisecond},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				start := time.Now()
				_, err := makeRequest("GET", tc.endpoint, headers, nil)
				duration := time.Since(start)
				
				if err != nil {
					t.Fatalf("请求失败: %v", err)
				}
				
				if duration > tc.maxTime {
					t.Errorf("响应时间过长: %v (最大允许: %v)", duration, tc.maxTime)
				} else {
					t.Logf("✅ %s响应时间: %v", tc.name, duration)
				}
			})
		}
	})
	
	t.Run("并发请求测试", func(t *testing.T) {
		const concurrency = 10
		const requests = 50
		
		ch := make(chan error, requests)
		
		start := time.Now()
		for i := 0; i < requests; i++ {
			go func() {
				_, err := makeRequest("GET", "/api/v1/status-definitions", headers, nil)
				ch <- err
			}()
		}
		
		// 收集结果
		var errors []error
		for i := 0; i < requests; i++ {
			if err := <-ch; err != nil {
				errors = append(errors, err)
			}
		}
		duration := time.Since(start)
		
		successRate := float64(requests-len(errors)) / float64(requests) * 100
		
		if successRate < 95.0 {
			t.Errorf("并发请求成功率过低: %.1f%%", successRate)
		} else {
			t.Logf("✅ 并发请求测试通过: %.1f%% 成功率, 总耗时: %v", successRate, duration)
		}
	})
}

// 清理测试数据
func cleanupTestData() {
	fmt.Println("🧹 清理测试数据...")
	
	if authToken == "" {
		return
	}
	
	headers := map[string]string{
		"Authorization": "Bearer " + authToken,
	}
	
	// 删除测试岗位申请
	for _, jobID := range testJobIDs {
		makeRequest("DELETE", fmt.Sprintf("/api/v1/job-applications/%d", jobID), headers, nil)
	}
	
	fmt.Println("✅ 测试数据清理完成")
}

// 辅助函数：检查字符串是否包含某个子字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.Contains(s, item) {
			return true
		}
	}
	return false
}
