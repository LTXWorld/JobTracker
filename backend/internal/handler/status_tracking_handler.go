// Location: /Users/lutao/GolandProjects/jobView/backend/internal/handler/status_tracking_handler.go
// This file implements HTTP handlers for job application status tracking API endpoints.
// It handles status history retrieval, status updates, timeline views, and batch operations.
// Used by the router to handle status tracking related HTTP requests with proper authentication and validation.

package handler

import (
	"encoding/json"
	"jobView-backend/internal/auth"
	"jobView-backend/internal/model"
	"jobView-backend/internal/service"
	"jobView-backend/internal/utils"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type StatusTrackingHandler struct {
	statusService *service.StatusTrackingService
}

func NewStatusTrackingHandler(statusService *service.StatusTrackingService) *StatusTrackingHandler {
	return &StatusTrackingHandler{
		statusService: statusService,
	}
}

// GetStatusHistory 获取岗位状态历史记录
// GET /api/v1/job-applications/{id}/status-history
func (h *StatusTrackingHandler) GetStatusHistory(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 获取岗位申请ID
	vars := mux.Vars(r)
	jobIDStr, ok := vars["id"]
	if !ok {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing job application id", nil)
		return
	}

	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid job application id", err)
		return
	}

	// 获取分页参数
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 50
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// 调用服务获取状态历史
	history, err := h.statusService.GetStatusHistory(uint(userID), jobID, page, pageSize)
	if err != nil {
		if err.Error() == "job application not found or access denied" {
			h.writeErrorResponse(w, http.StatusNotFound, "job application not found", nil)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get status history", err)
		}
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "status history retrieved successfully", history)
}

// UpdateJobStatus 更新岗位状态
// POST /api/v1/job-applications/{id}/status
func (h *StatusTrackingHandler) UpdateJobStatus(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 获取岗位申请ID
	vars := mux.Vars(r)
	jobIDStr, ok := vars["id"]
	if !ok {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing job application id", nil)
		return
	}

	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid job application id", err)
		return
	}

	// 解析请求体
	var req model.StatusUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// 验证请求
	if err := h.validateStatusUpdateRequest(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 调用服务更新状态
	updatedJob, err := h.statusService.UpdateJobStatus(uint(userID), jobID, &req)
	if err != nil {
		if err.Error() == "job application not found" {
			h.writeErrorResponse(w, http.StatusNotFound, "job application not found", nil)
		} else if err.Error() == "version conflict" {
			h.writeErrorResponse(w, http.StatusConflict, "version conflict, please refresh and try again", nil)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to update job status", err)
		}
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "job status updated successfully", updatedJob)
}

// GetStatusTimeline 获取岗位状态时间轴视图
// GET /api/v1/job-applications/{id}/status-timeline
func (h *StatusTrackingHandler) GetStatusTimeline(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 获取岗位申请ID
	vars := mux.Vars(r)
	jobIDStr, ok := vars["id"]
	if !ok {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing job application id", nil)
		return
	}

	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid job application id", err)
		return
	}

	// 调用服务获取时间轴数据
	timeline, err := h.statusService.GetStatusTimeline(uint(userID), jobID)
	if err != nil {
		if err.Error() == "job application not found or access denied" {
			h.writeErrorResponse(w, http.StatusNotFound, "job application not found", nil)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get status timeline", err)
		}
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "status timeline retrieved successfully", timeline)
}

// BatchUpdateStatus 批量状态更新
// PUT /api/v1/job-applications/status/batch
func (h *StatusTrackingHandler) BatchUpdateStatus(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 解析请求体
	var updates []model.BatchStatusUpdate
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// 验证批量更新请求
	if err := h.validateBatchStatusUpdateRequest(updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 调用服务进行批量更新
	err := h.statusService.BatchUpdateStatus(uint(userID), updates)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to batch update status", err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "batch status update completed successfully", map[string]interface{}{
		"updated_count": len(updates),
	})
}

// GetStatusAnalytics 获取用户状态分析数据
// GET /api/v1/job-applications/status-analytics
func (h *StatusTrackingHandler) GetStatusAnalytics(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 调用服务获取分析数据
	analytics, err := h.statusService.GetStatusAnalytics(uint(userID))
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get status analytics", err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "status analytics retrieved successfully", analytics)
}

// GetStatusTrends 获取状态趋势数据
// GET /api/v1/job-applications/status-trends
func (h *StatusTrackingHandler) GetStatusTrends(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 获取天数参数
	days := 30 // 默认30天
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	// 调用服务获取趋势数据
	trends, err := h.statusService.GetStatusTrends(uint(userID), days)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get status trends", err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "status trends retrieved successfully", map[string]interface{}{
		"days":   days,
		"trends": trends,
	})
}

// GetProcessInsights 获取流程洞察数据
// GET /api/v1/job-applications/process-insights
func (h *StatusTrackingHandler) GetProcessInsights(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 获取分析数据
	analytics, err := h.statusService.GetStatusAnalytics(uint(userID))
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get process insights", err)
		return
	}

	// 获取趋势数据
	trends, err := h.statusService.GetStatusTrends(uint(userID), 90) // 最近90天
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get trends for insights", err)
		return
	}

	// 构建洞察数据
	insights := map[string]interface{}{
		"summary": map[string]interface{}{
			"total_applications": analytics.TotalApplications,
			"success_rate":      analytics.SuccessRate,
			"active_applications": h.calculateActiveApplications(analytics.StatusDistribution),
		},
		"performance": map[string]interface{}{
			"average_durations": analytics.AverageDurations,
			"stage_analysis":    analytics.StageAnalysis,
		},
		"trends": map[string]interface{}{
			"recent_activity": trends,
		},
		"recommendations": h.generateRecommendations(analytics),
	}

	h.writeSuccessResponse(w, http.StatusOK, "process insights retrieved successfully", insights)
}

// validateStatusUpdateRequest 验证状态更新请求
func (h *StatusTrackingHandler) validateStatusUpdateRequest(req *model.StatusUpdateRequest) error {
	if !req.Status.IsValid() {
		return utils.NewValidationError("status", "invalid status value")
	}

	if req.Note != nil && len(*req.Note) > 1000 {
		return utils.NewValidationError("note", "note too long (max 1000 characters)")
	}

	return nil
}

// validateBatchStatusUpdateRequest 验证批量状态更新请求
func (h *StatusTrackingHandler) validateBatchStatusUpdateRequest(updates []model.BatchStatusUpdate) error {
	if len(updates) == 0 {
		return utils.NewValidationError("updates", "no updates provided")
	}

	if len(updates) > 100 {
		return utils.NewValidationError("updates", "too many updates (max 100)")
	}

	// 验证每个更新项
	for i, update := range updates {
		if update.ID <= 0 {
			return utils.NewValidationError("updates", "invalid id at index %d", i)
		}
		if !update.Status.IsValid() {
			return utils.NewValidationError("updates", "invalid status at index %d", i)
		}
	}

	return nil
}

// calculateActiveApplications 计算活跃申请数量
func (h *StatusTrackingHandler) calculateActiveApplications(statusDistribution map[string]int) int {
	activeCount := 0
	for status, count := range statusDistribution {
		appStatus := model.ApplicationStatus(status)
		if appStatus.IsInProgressStatus() {
			activeCount += count
		}
	}
	return activeCount
}

// generateRecommendations 生成建议
func (h *StatusTrackingHandler) generateRecommendations(analytics *model.StatusAnalyticsResponse) []string {
	var recommendations []string

	// 基于成功率的建议
	if analytics.SuccessRate < 10 {
		recommendations = append(recommendations, "建议优化简历和投递策略，当前成功率较低")
	} else if analytics.SuccessRate > 50 {
		recommendations = append(recommendations, "投递策略很好，继续保持")
	}

	// 基于申请数量的建议
	if analytics.TotalApplications < 10 {
		recommendations = append(recommendations, "建议增加投递数量，扩大求职机会")
	}

	// 基于持续时间的建议
	if avgDuration, ok := analytics.AverageDurations["简历筛选中"]; ok && avgDuration > 7*24*60 { // 超过7天
		recommendations = append(recommendations, "简历筛选时间较长，可考虑主动跟进或优化简历")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "继续保持良好的求职进展")
	}

	return recommendations
}

// writeSuccessResponse 写入成功响应
func (h *StatusTrackingHandler) writeSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := model.APIResponse{
		Code:    statusCode,
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

// writeErrorResponse 写入错误响应
func (h *StatusTrackingHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := model.APIResponse{
		Code:    statusCode,
		Message: message,
	}

	if err != nil && statusCode >= 500 {
		// 只在服务器内部错误时显示详细错误信息
		response.Data = map[string]string{"error": err.Error()}
	}

	json.NewEncoder(w).Encode(response)
}