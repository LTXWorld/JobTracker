package handler

import (
	"net/http"
	"time"

	"jobView-backend/internal/database"
	"jobView-backend/internal/model"

	"github.com/gin-gonic/gin"
)

// MonitorHandler 数据库性能监控处理器
type MonitorHandler struct {
	db *database.DB
}

// NewMonitorHandler 创建监控处理器
func NewMonitorHandler(db *database.DB) *MonitorHandler {
	return &MonitorHandler{
		db: db,
	}
}

// GetDatabaseStats 获取数据库性能统计信息
// @Summary 获取数据库性能统计
// @Description 返回数据库查询性能、连接池状态等监控信息
// @Tags 监控
// @Produce json
// @Success 200 {object} model.APIResponse{data=database.PerformanceStats}
// @Router /api/v1/monitor/db-stats [get]
func (h *MonitorHandler) GetDatabaseStats(c *gin.Context) {
	stats := h.db.GetStats()
	
	c.JSON(http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "Database statistics retrieved successfully",
		Data:    stats,
	})
}

// GetConnectionStats 获取连接池统计信息
// @Summary 获取数据库连接池统计
// @Description 返回数据库连接池使用情况和相关指标
// @Tags 监控
// @Produce json
// @Success 200 {object} model.APIResponse{data=map[string]interface{}}
// @Router /api/v1/monitor/connection-stats [get]
func (h *MonitorHandler) GetConnectionStats(c *gin.Context) {
	stats := h.db.GetConnectionStats()
	
	c.JSON(http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "Connection pool statistics retrieved successfully",
		Data:    stats,
	})
}

// GetHealthStatus 获取数据库健康状态
// @Summary 数据库健康检查
// @Description 检查数据库连接和响应状态
// @Tags 监控
// @Produce json
// @Success 200 {object} model.APIResponse{data=map[string]interface{}}
// @Router /api/v1/monitor/health [get]
func (h *MonitorHandler) GetHealthStatus(c *gin.Context) {
	isHealthy := h.db.IsHealthy()
	
	status := map[string]interface{}{
		"healthy":    isHealthy,
		"timestamp":  time.Now(),
		"database":   "postgresql",
	}
	
	if isHealthy {
		c.JSON(http.StatusOK, model.APIResponse{
			Code:    0,
			Message: "Database is healthy",
			Data:    status,
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, model.APIResponse{
			Code:    503,
			Message: "Database is unhealthy",
			Data:    status,
		})
	}
}

// ResetStats 重置监控统计信息
// @Summary 重置数据库监控统计
// @Description 清空查询统计、慢查询记录等监控数据
// @Tags 监控
// @Produce json
// @Success 200 {object} model.APIResponse
// @Router /api/v1/monitor/reset-stats [post]
func (h *MonitorHandler) ResetStats(c *gin.Context) {
	h.db.Monitor.ResetStats()
	
	c.JSON(http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "Database monitoring statistics reset successfully",
	})
}

// SetSlowThreshold 设置慢查询阈值
// @Summary 设置慢查询阈值
// @Description 动态调整慢查询判定阈值（毫秒）
// @Tags 监控
// @Param threshold query int true "慢查询阈值（毫秒）"
// @Produce json
// @Success 200 {object} model.APIResponse
// @Failure 400 {object} model.APIResponse
// @Router /api/v1/monitor/slow-threshold [put]
func (h *MonitorHandler) SetSlowThreshold(c *gin.Context) {
	type ThresholdRequest struct {
		ThresholdMs int `json:"threshold_ms" binding:"required,min=1,max=60000"`
	}
	
	var req ThresholdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Code:    400,
			Message: "Invalid threshold value: " + err.Error(),
		})
		return
	}
	
	threshold := time.Duration(req.ThresholdMs) * time.Millisecond
	h.db.Monitor.SetSlowThreshold(threshold)
	
	c.JSON(http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "Slow query threshold updated successfully",
		Data: map[string]interface{}{
			"new_threshold_ms": req.ThresholdMs,
			"new_threshold":    threshold.String(),
		},
	})
}

// GetSlowThreshold 获取当前慢查询阈值
// @Summary 获取慢查询阈值
// @Description 返回当前的慢查询判定阈值
// @Tags 监控
// @Produce json
// @Success 200 {object} model.APIResponse{data=map[string]interface{}}
// @Router /api/v1/monitor/slow-threshold [get]
func (h *MonitorHandler) GetSlowThreshold(c *gin.Context) {
	threshold := h.db.Monitor.GetSlowThreshold()
	
	c.JSON(http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "Current slow query threshold retrieved",
		Data: map[string]interface{}{
			"threshold_ms": threshold.Milliseconds(),
			"threshold":    threshold.String(),
		},
	})
}

// GetDashboard 获取监控面板数据
// @Summary 获取监控面板数据
// @Description 返回数据库性能监控面板所需的全部信息
// @Tags 监控
// @Produce json
// @Success 200 {object} model.APIResponse{data=map[string]interface{}}
// @Router /api/v1/monitor/dashboard [get]
func (h *MonitorHandler) GetDashboard(c *gin.Context) {
	stats := h.db.GetStats()
	connectionStats := h.db.GetConnectionStats()
	isHealthy := h.db.IsHealthy()
	
	dashboard := map[string]interface{}{
		"health": map[string]interface{}{
			"status":    isHealthy,
			"timestamp": time.Now(),
		},
		"performance":     stats,
		"connection_pool": connectionStats,
		"summary": map[string]interface{}{
			"total_queries":     stats.QueryStats.TotalQueries,
			"slow_queries":      stats.QueryStats.SlowQueries,
			"slow_query_rate":   stats.QueryStats.SlowQueryRate,
			"average_latency":   stats.QueryStats.AverageLatency.String(),
			"connection_usage":  connectionStats["connection_utilization"],
			"healthy":          isHealthy,
		},
	}
	
	c.JSON(http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "Monitoring dashboard data retrieved successfully",
		Data:    dashboard,
	})
}