package handler

import (
	"encoding/json"
	"net/http"
	"jobView-backend/internal/database"
	"time"
)

// DatabaseStatsHandler 数据库性能统计API
type DatabaseStatsHandler struct {
	db *database.DB
}

// NewDatabaseStatsHandler 创建数据库统计处理器
func NewDatabaseStatsHandler(db *database.DB) *DatabaseStatsHandler {
	return &DatabaseStatsHandler{db: db}
}

// GetDatabaseStats 获取数据库性能统计
func (h *DatabaseStatsHandler) GetDatabaseStats(w http.ResponseWriter, r *http.Request) {
	stats := h.db.GetStats()
	
	response := map[string]interface{}{
		"code":    200,
		"message": "success",
		"data": map[string]interface{}{
			"query_performance": map[string]interface{}{
				"total_queries":    stats.QueryStats.TotalQueries,
				"slow_queries":     stats.QueryStats.SlowQueries,
				"average_latency":  stats.QueryStats.AverageLatency.Milliseconds(),
				"slow_query_rate":  stats.QueryStats.SlowQueryRate,
				"top_slow_queries": formatSlowQueries(stats.QueryStats.TopSlowQueries),
			},
			"connection_pool": map[string]interface{}{
				"max_open_connections": stats.ConnectionStats.MaxOpenConnections,
				"open_connections":     stats.ConnectionStats.OpenConnections,
				"in_use":              stats.ConnectionStats.InUse,
				"idle":                stats.ConnectionStats.Idle,
				"wait_count":          stats.ConnectionStats.WaitCount,
				"wait_duration_ms":    stats.ConnectionStats.WaitDuration.Milliseconds(),
				"utilization_rate":    float64(stats.ConnectionStats.InUse) / float64(stats.ConnectionStats.MaxOpenConnections) * 100,
			},
			"health_status": map[string]interface{}{
				"is_healthy":    h.db.IsHealthy(),
				"health_detail": h.db.Health.GetHealthStatus(),
			},
			"timestamp": stats.Timestamp,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetConnectionPoolStats 获取连接池详细统计
func (h *DatabaseStatsHandler) GetConnectionPoolStats(w http.ResponseWriter, r *http.Request) {
	stats := h.db.GetConnectionStats()
	
	response := map[string]interface{}{
		"code":    200,
		"message": "success",
		"data":    stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ResetPerformanceStats 重置性能统计
func (h *DatabaseStatsHandler) ResetPerformanceStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.db.Monitor.ResetStats()
	
	response := map[string]interface{}{
		"code":    200,
		"message": "Performance stats reset successfully",
		"data":    nil,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// formatSlowQueries 格式化慢查询数据
func formatSlowQueries(slowQueries []database.SlowQuery) []map[string]interface{} {
	formatted := make([]map[string]interface{}, 0, len(slowQueries))
	
	// 只返回最近的10条慢查询
	start := 0
	if len(slowQueries) > 10 {
		start = len(slowQueries) - 10
	}
	
	for i := start; i < len(slowQueries); i++ {
		sq := slowQueries[i]
		formatted = append(formatted, map[string]interface{}{
			"sql":           truncateSQL(sq.SQL, 200), // 截断长SQL
			"duration_ms":   sq.Duration.Milliseconds(),
			"timestamp":     sq.Timestamp,
			"relative_time": formatRelativeTime(sq.Timestamp),
		})
	}
	
	return formatted
}

// truncateSQL 截断长SQL语句
func truncateSQL(sql string, maxLength int) string {
	if len(sql) <= maxLength {
		return sql
	}
	return sql[:maxLength] + "..."
}

// formatRelativeTime 格式化相对时间
func formatRelativeTime(t time.Time) string {
	duration := time.Since(t)
	
	switch {
	case duration < time.Minute:
		return "刚刚"
	case duration < time.Hour:
		return string(int(duration.Minutes())) + "分钟前"
	case duration < 24*time.Hour:
		return string(int(duration.Hours())) + "小时前"
	default:
		return string(int(duration.Hours()/24)) + "天前"
	}
}