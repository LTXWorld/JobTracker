package database

import (
	"context"
	"log"
	"sync"
	"time"
)

// DatabaseHealthChecker 数据库健康检查器
type DatabaseHealthChecker struct {
	db           *DB
	interval     time.Duration
	timeout      time.Duration
	isHealthy    bool
	lastCheck    time.Time
	lastError    error
	mu           sync.RWMutex
	stopCh       chan struct{}
	logger       *log.Logger
}

// HealthStatus 健康状态信息
type HealthStatus struct {
	IsHealthy    bool          `json:"is_healthy"`
	LastCheck    time.Time     `json:"last_check"`
	LastError    string        `json:"last_error,omitempty"`
	ResponseTime time.Duration `json:"response_time"`
	Uptime       time.Duration `json:"uptime"`
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(db *DB, checkInterval time.Duration) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{
		db:        db,
		interval:  checkInterval,
		timeout:   5 * time.Second, // 健康检查超时时间
		isHealthy: true,
		stopCh:    make(chan struct{}),
		logger:    log.New(log.Writer(), "[DB-HEALTH] ", log.LstdFlags|log.Lshortfile),
	}
}

// StartHealthCheck 启动健康检查
func (hc *DatabaseHealthChecker) StartHealthCheck() {
	go hc.healthCheckLoop()
}

// StopHealthCheck 停止健康检查
func (hc *DatabaseHealthChecker) StopHealthCheck() {
	close(hc.stopCh)
}

// IsHealthy 检查数据库是否健康
func (hc *DatabaseHealthChecker) IsHealthy() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.isHealthy
}

// GetHealthStatus 获取健康状态详情
func (hc *DatabaseHealthChecker) GetHealthStatus() HealthStatus {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	
	status := HealthStatus{
		IsHealthy: hc.isHealthy,
		LastCheck: hc.lastCheck,
		Uptime:    time.Since(hc.lastCheck),
	}
	
	if hc.lastError != nil {
		status.LastError = hc.lastError.Error()
	}
	
	return status
}

// healthCheckLoop 健康检查循环
func (hc *DatabaseHealthChecker) healthCheckLoop() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()
	
	// 立即执行一次健康检查
	hc.performHealthCheck()
	
	for {
		select {
		case <-ticker.C:
			hc.performHealthCheck()
		case <-hc.stopCh:
			hc.logger.Println("Health check stopped")
			return
		}
	}
}

// performHealthCheck 执行健康检查
func (hc *DatabaseHealthChecker) performHealthCheck() {
	start := time.Now()
	
	ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
	defer cancel()
	
	// 执行简单的数据库查询来检查连接状态
	var result int
	err := hc.db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	
	responseTime := time.Since(start)
	
	hc.mu.Lock()
	hc.lastCheck = time.Now()
	hc.lastError = err
	
	previousHealthy := hc.isHealthy
	hc.isHealthy = (err == nil && result == 1)
	hc.mu.Unlock()
	
	// 记录健康状态变化
	if previousHealthy != hc.isHealthy {
		if hc.isHealthy {
			hc.logger.Printf("Database health recovered (response time: %v)", responseTime)
		} else {
			hc.logger.Printf("Database health check failed: %v (response time: %v)", err, responseTime)
		}
	}
	
	// 记录响应时间警告
	if responseTime > hc.timeout/2 {
		hc.logger.Printf("Slow database response: %v", responseTime)
	}
}

