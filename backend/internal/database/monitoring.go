package database

import (
	"database/sql"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// QueryMonitor 查询性能监控器
type QueryMonitor struct {
	slowThreshold time.Duration
	logger        *log.Logger
	stats         *QueryStats
}

// QueryStats 查询统计信息
type QueryStats struct {
	TotalQueries   int64
	SlowQueries    int64
	TotalDuration  int64 // nanoseconds
	slowQueries    []SlowQuery
	mu             sync.RWMutex
}

// SlowQuery 慢查询记录
type SlowQuery struct {
	SQL       string        `json:"sql"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
	Args      []interface{} `json:"args,omitempty"`
}

// PerformanceStats 性能统计
type PerformanceStats struct {
	QueryStats      QueryStatsInfo `json:"query_stats"`
	ConnectionStats ConnectionStats `json:"connection_stats"`
	Timestamp       time.Time      `json:"timestamp"`
}

// QueryStatsInfo 查询统计信息
type QueryStatsInfo struct {
	TotalQueries    int64           `json:"total_queries"`
	SlowQueries     int64           `json:"slow_queries"`
	AverageLatency  time.Duration   `json:"average_latency"`
	SlowQueryRate   float64         `json:"slow_query_rate"`
	TopSlowQueries  []SlowQuery     `json:"top_slow_queries"`
}

// ConnectionStats 连接池统计
type ConnectionStats struct {
	MaxOpenConnections int           `json:"max_open_connections"`
	OpenConnections    int           `json:"open_connections"`
	InUse             int           `json:"in_use"`
	Idle              int           `json:"idle"`
	WaitCount         int64         `json:"wait_count"`
	WaitDuration      time.Duration `json:"wait_duration"`
	MaxIdleClosed     int64         `json:"max_idle_closed"`
	MaxIdleTimeClosed int64         `json:"max_idle_time_closed"`
	MaxLifetimeClosed int64         `json:"max_lifetime_closed"`
}

// NewQueryMonitor 创建查询监控器
func NewQueryMonitor(slowThreshold time.Duration, logger *log.Logger) *QueryMonitor {
	return &QueryMonitor{
		slowThreshold: slowThreshold,
		logger:        logger,
		stats: &QueryStats{
			slowQueries: make([]SlowQuery, 0, 100), // 最多保存100条慢查询记录
		},
	}
}

// WrapDB 包装数据库连接，添加监控功能
func (qm *QueryMonitor) WrapDB(db *sql.DB) *MonitoredDB {
	return &MonitoredDB{
		DB:      db,
		monitor: qm,
	}
}

// RecordQuery 记录查询性能
func (qm *QueryMonitor) recordQuery(query string, duration time.Duration, args []interface{}) {
	// 更新总查询数和总耗时
	atomic.AddInt64(&qm.stats.TotalQueries, 1)
	atomic.AddInt64(&qm.stats.TotalDuration, int64(duration))

	// 检查是否为慢查询
	if duration > qm.slowThreshold {
		atomic.AddInt64(&qm.stats.SlowQueries, 1)
		
		// 记录慢查询详情
		slowQuery := SlowQuery{
			SQL:       query,
			Duration:  duration,
			Timestamp: time.Now(),
			Args:      args,
		}
		
		// 线程安全地添加慢查询记录
		qm.stats.mu.Lock()
		qm.stats.slowQueries = append(qm.stats.slowQueries, slowQuery)
		
		// 限制慢查询记录数量，保留最新的100条
		if len(qm.stats.slowQueries) > 100 {
			qm.stats.slowQueries = qm.stats.slowQueries[len(qm.stats.slowQueries)-100:]
		}
		qm.stats.mu.Unlock()
		
		// 记录慢查询日志
		if qm.logger != nil {
			qm.logger.Printf("SLOW QUERY [%v]: %s", duration, query)
		}
	}
}

// GetStats 获取性能统计信息
func (qm *QueryMonitor) GetStats(dbStats sql.DBStats) PerformanceStats {
	totalQueries := atomic.LoadInt64(&qm.stats.TotalQueries)
	slowQueries := atomic.LoadInt64(&qm.stats.SlowQueries)
	totalDuration := atomic.LoadInt64(&qm.stats.TotalDuration)
	
	var avgLatency time.Duration
	var slowQueryRate float64
	
	if totalQueries > 0 {
		avgLatency = time.Duration(totalDuration / totalQueries)
		slowQueryRate = float64(slowQueries) / float64(totalQueries) * 100
	}
	
	// 获取慢查询记录（线程安全）
	qm.stats.mu.RLock()
	topSlowQueries := make([]SlowQuery, len(qm.stats.slowQueries))
	copy(topSlowQueries, qm.stats.slowQueries)
	qm.stats.mu.RUnlock()
	
	return PerformanceStats{
		QueryStats: QueryStatsInfo{
			TotalQueries:   totalQueries,
			SlowQueries:    slowQueries,
			AverageLatency: avgLatency,
			SlowQueryRate:  slowQueryRate,
			TopSlowQueries: topSlowQueries,
		},
		ConnectionStats: ConnectionStats{
			MaxOpenConnections: dbStats.MaxOpenConnections,
			OpenConnections:    dbStats.OpenConnections,
			InUse:             dbStats.InUse,
			Idle:              dbStats.Idle,
			WaitCount:         dbStats.WaitCount,
			WaitDuration:      dbStats.WaitDuration,
			MaxIdleClosed:     dbStats.MaxIdleClosed,
			MaxIdleTimeClosed: dbStats.MaxIdleTimeClosed,
			MaxLifetimeClosed: dbStats.MaxLifetimeClosed,
		},
		Timestamp: time.Now(),
	}
}

// ResetStats 重置统计信息
func (qm *QueryMonitor) ResetStats() {
	atomic.StoreInt64(&qm.stats.TotalQueries, 0)
	atomic.StoreInt64(&qm.stats.SlowQueries, 0)
	atomic.StoreInt64(&qm.stats.TotalDuration, 0)
	
	qm.stats.mu.Lock()
	qm.stats.slowQueries = qm.stats.slowQueries[:0]
	qm.stats.mu.Unlock()
}

// SetSlowThreshold 设置慢查询阈值
func (qm *QueryMonitor) SetSlowThreshold(threshold time.Duration) {
	qm.slowThreshold = threshold
}

// GetSlowThreshold 获取慢查询阈值
func (qm *QueryMonitor) GetSlowThreshold() time.Duration {
	return qm.slowThreshold
}

// MonitoredDB 带监控的数据库连接
type MonitoredDB struct {
	*sql.DB
	monitor *QueryMonitor
}

// Query 执行查询并记录性能
func (db *MonitoredDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := db.DB.Query(query, args...)
	duration := time.Since(start)
	
	// 记录查询性能
	db.monitor.recordQuery(query, duration, args)
	
	return rows, err
}

// QueryRow 执行单行查询并记录性能
func (db *MonitoredDB) QueryRow(query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := db.DB.QueryRow(query, args...)
	duration := time.Since(start)
	
	// 记录查询性能
	db.monitor.recordQuery(query, duration, args)
	
	return row
}

// Exec 执行SQL并记录性能
func (db *MonitoredDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := db.DB.Exec(query, args...)
	duration := time.Since(start)
	
	// 记录查询性能
	db.monitor.recordQuery(query, duration, args)
	
	return result, err
}

// GetConnectionPoolStats 获取连接池详细统计
func GetConnectionPoolStats(db *sql.DB) map[string]interface{} {
	stats := db.Stats()
	
	return map[string]interface{}{
		"max_open_connections":     stats.MaxOpenConnections,
		"open_connections":         stats.OpenConnections,
		"in_use":                  stats.InUse,
		"idle":                    stats.Idle,
		"wait_count":              stats.WaitCount,
		"wait_duration_ms":        stats.WaitDuration.Milliseconds(),
		"max_idle_closed":         stats.MaxIdleClosed,
		"max_idle_time_closed":    stats.MaxIdleTimeClosed,
		"max_lifetime_closed":     stats.MaxLifetimeClosed,
		"connection_utilization":  float64(stats.InUse) / float64(stats.MaxOpenConnections) * 100,
		"idle_connection_ratio":   float64(stats.Idle) / float64(stats.OpenConnections) * 100,
	}
}