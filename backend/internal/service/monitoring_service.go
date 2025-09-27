// 位置: backend/internal/service/monitoring_service.go
// 说明: 封装数据库监控与健康检查逻辑，供 Handler 层依赖，避免直接操作 database.DB。

package service

import (
	"time"

	"jobView-backend/internal/database"
)

// MonitoringService 抽象监控能力，便于后续替换数据源或做单元测试
type MonitoringService interface {
	GetPerformanceStats() database.PerformanceStats
	GetConnectionStats() map[string]interface{}
	IsHealthy() bool
	GetHealthStatus() database.HealthStatus
	ResetPerformanceStats()
	UpdateSlowThreshold(threshold time.Duration)
	CurrentSlowThreshold() time.Duration
}

type monitoringService struct {
	db *database.DB
}

// NewMonitoringService 创建监控服务实例
func NewMonitoringService(db *database.DB) MonitoringService {
	return &monitoringService{db: db}
}

func (s *monitoringService) GetPerformanceStats() database.PerformanceStats {
	if s.db == nil {
		return database.PerformanceStats{}
	}
	return s.db.GetStats()
}

func (s *monitoringService) GetConnectionStats() map[string]interface{} {
	if s.db == nil {
		return map[string]interface{}{}
	}
	return s.db.GetConnectionStats()
}

func (s *monitoringService) IsHealthy() bool {
	if s.db == nil {
		return false
	}
	return s.db.IsHealthy()
}

func (s *monitoringService) GetHealthStatus() database.HealthStatus {
	if s.db == nil || s.db.Health == nil {
		return database.HealthStatus{}
	}
	return s.db.Health.GetHealthStatus()
}

func (s *monitoringService) ResetPerformanceStats() {
	if s.db == nil || s.db.Monitor == nil {
		return
	}
	s.db.Monitor.ResetStats()
}

func (s *monitoringService) UpdateSlowThreshold(threshold time.Duration) {
	if s.db == nil || s.db.Monitor == nil {
		return
	}
	s.db.Monitor.SetSlowThreshold(threshold)
}

func (s *monitoringService) CurrentSlowThreshold() time.Duration {
	if s.db == nil || s.db.Monitor == nil {
		return 0
	}
	return s.db.Monitor.GetSlowThreshold()
}
