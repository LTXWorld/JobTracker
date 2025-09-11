/*
位置: backend/internal/handler/export_handler.go
概述: Excel导出功能的HTTP处理器，提供导出API端点
功能: 处理导出请求、任务状态查询、文件下载和导出历史
与其他文件关系: 依赖 export_service.go 和 JWT 认证中间件，被主路由器调用
*/

package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"jobView-backend/internal/auth"
	"jobView-backend/internal/model"
	"jobView-backend/internal/service"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// ExportHandler 导出处理器
type ExportHandler struct {
	exportService *service.ExportService
}

// NewExportHandler 创建导出处理器
func NewExportHandler(exportService *service.ExportService) *ExportHandler {
	return &ExportHandler{
		exportService: exportService,
	}
}

// StartExport 启动导出任务
// POST /api/v1/export/applications
func (h *ExportHandler) StartExport(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 解析请求体
	var req model.ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "请求体格式错误", err)
		return
	}

	// 验证导出请求
	if err := req.ValidateExportRequest(); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 开始导出任务
	response, err := h.exportService.StartExport(userID, &req)
	if err != nil {
		// 根据错误类型返回不同的HTTP状态码
		if strings.Contains(err.Error(), "上限") || strings.Contains(err.Error(), "超限") {
			h.writeErrorResponse(w, http.StatusTooManyRequests, err.Error(), nil)
		} else if strings.Contains(err.Error(), "没有符合条件的数据") {
			h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "启动导出任务失败", err)
		}
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "导出任务已启动", response)
}

// GetTaskStatus 获取任务状态
// GET /api/v1/export/status/{task_id}
func (h *ExportHandler) GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 获取任务ID
	vars := mux.Vars(r)
	taskID, ok := vars["task_id"]
	if !ok {
		h.writeErrorResponse(w, http.StatusBadRequest, "缺少任务ID参数", nil)
		return
	}

	// 查询任务状态
	status, err := h.exportService.GetTaskStatus(taskID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") || strings.Contains(err.Error(), "无访问权限") {
			h.writeErrorResponse(w, http.StatusNotFound, err.Error(), nil)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "查询任务状态失败", err)
		}
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "查询成功", status)
}

// DownloadFile 下载导出文件
// GET /api/v1/export/download/{task_id}
func (h *ExportHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 获取任务ID
	vars := mux.Vars(r)
	taskID, ok := vars["task_id"]
	if !ok {
		h.writeErrorResponse(w, http.StatusBadRequest, "缺少任务ID参数", nil)
		return
	}

	// 获取文件路径和文件名
	filePath, filename, err := h.exportService.DownloadFile(taskID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") || strings.Contains(err.Error(), "无访问权限") {
			h.writeErrorResponse(w, http.StatusNotFound, err.Error(), nil)
		} else if strings.Contains(err.Error(), "过期") {
			h.writeErrorResponse(w, http.StatusGone, err.Error(), nil)
		} else if strings.Contains(err.Error(), "尚未生成完成") {
			h.writeErrorResponse(w, http.StatusAccepted, err.Error(), nil)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "获取下载文件失败", err)
		}
		return
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "打开文件失败", err)
		return
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "获取文件信息失败", err)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// 发送文件内容
	_, err = io.Copy(w, file)
	if err != nil {
		// 如果已经开始发送响应，就不能再设置错误响应了
		fmt.Printf("发送文件内容时出错: %v\n", err)
		return
	}
}

// GetExportHistory 获取导出历史
// GET /api/v1/export/history
func (h *ExportHandler) GetExportHistory(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 获取分页参数
	page := 1
	limit := 10

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// 查询导出历史
	history, err := h.exportService.GetExportHistory(userID, page, limit)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "查询导出历史失败", err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "查询成功", history)
}

// CancelExport 取消导出任务
// DELETE /api/v1/export/cancel/{task_id}
func (h *ExportHandler) CancelExport(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	_, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 获取任务ID
	vars := mux.Vars(r)
	_, ok = vars["task_id"]
	if !ok {
		h.writeErrorResponse(w, http.StatusBadRequest, "缺少任务ID参数", nil)
		return
	}

	// TODO: 实现取消导出功能
	// 这里可以添加取消正在进行的导出任务的逻辑
	h.writeErrorResponse(w, http.StatusNotImplemented, "取消导出功能暂未实现", nil)
}

// GetSupportedFormats 获取支持的导出格式
// GET /api/v1/export/formats
func (h *ExportHandler) GetSupportedFormats(w http.ResponseWriter, r *http.Request) {
	formats := map[string]interface{}{
		"formats": []map[string]string{
			{
				"value":       "xlsx",
				"label":       "Excel 文件 (.xlsx)",
				"description": "Microsoft Excel 2007+ 格式，支持样式和多工作表",
			},
			{
				"value":       "csv",
				"label":       "CSV 文件 (.csv)",
				"description": "逗号分隔值格式，通用性强但不支持样式",
			},
		},
		"defaultFormat": "xlsx",
	}

	h.writeSuccessResponse(w, http.StatusOK, "查询成功", formats)
}

// GetExportFields 获取可导出的字段
// GET /api/v1/export/fields
func (h *ExportHandler) GetExportFields(w http.ResponseWriter, r *http.Request) {
	fields := map[string]interface{}{
		"fields": []map[string]interface{}{
			{
				"value":       "company_name",
				"label":       "公司名称",
				"required":    true,
				"description": "投递的公司名称",
			},
			{
				"value":       "position_title",
				"label":       "职位标题",
				"required":    true,
				"description": "申请的职位名称",
			},
			{
				"value":       "application_date",
				"label":       "投递日期",
				"required":    false,
				"description": "投递简历的日期",
			},
			{
				"value":       "status",
				"label":       "当前状态",
				"required":    false,
				"description": "当前投递状态",
			},
			{
				"value":       "salary_range",
				"label":       "薪资范围",
				"required":    false,
				"description": "期望或提供的薪资范围",
			},
			{
				"value":       "work_location",
				"label":       "工作地点",
				"required":    false,
				"description": "工作城市或地点",
			},
			{
				"value":       "interview_time",
				"label":       "面试时间",
				"required":    false,
				"description": "面试安排时间",
			},
			{
				"value":       "interview_location",
				"label":       "面试地点",
				"required":    false,
				"description": "面试地点或方式",
			},
			{
				"value":       "interview_type",
				"label":       "面试类型",
				"required":    false,
				"description": "面试形式（线上/线下/电话等）",
			},
			{
				"value":       "hr_name",
				"label":       "HR姓名",
				"required":    false,
				"description": "负责HR的姓名",
			},
			{
				"value":       "hr_phone",
				"label":       "HR电话",
				"required":    false,
				"description": "HR联系电话",
			},
			{
				"value":       "hr_email",
				"label":       "HR邮箱",
				"required":    false,
				"description": "HR联系邮箱",
			},
			{
				"value":       "reminder_time",
				"label":       "提醒时间",
				"required":    false,
				"description": "设置的提醒时间",
			},
			{
				"value":       "follow_up_date",
				"label":       "跟进日期",
				"required":    false,
				"description": "计划跟进的日期",
			},
			{
				"value":       "notes",
				"label":       "备注",
				"required":    false,
				"description": "相关备注信息",
			},
			{
				"value":       "created_at",
				"label":       "创建时间",
				"required":    false,
				"description": "记录创建时间",
			},
			{
				"value":       "updated_at",
				"label":       "更新时间",
				"required":    false,
				"description": "记录最后更新时间",
			},
		},
		"defaultFields": []string{
			"company_name",
			"position_title",
			"application_date",
			"status",
			"salary_range",
			"work_location",
			"interview_time",
			"notes",
		},
	}

	h.writeSuccessResponse(w, http.StatusOK, "查询成功", fields)
}

// 辅助方法

// writeSuccessResponse 写入成功响应
func (h *ExportHandler) writeSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	response := model.APIResponse{
		Code:    statusCode,
		Message: message,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("编码响应失败: %v\n", err)
	}
}

// writeErrorResponse 写入错误响应
func (h *ExportHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	response := model.APIResponse{
		Code:    statusCode,
		Message: message,
	}

	// 在开发环境下可以包含详细错误信息
	if err != nil {
		// 这里可以根据环境变量决定是否包含详细错误
		// response.Data = map[string]string{"error": err.Error()}
		fmt.Printf("错误详情: %v\n", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if jsonErr := json.NewEncoder(w).Encode(response); jsonErr != nil {
		fmt.Printf("编码错误响应失败: %v\n", jsonErr)
	}
}

// ValidateExportPermission 验证导出权限（可扩展）
func (h *ExportHandler) ValidateExportPermission(userID uint, request *model.ExportRequest) error {
	// 这里可以添加更复杂的权限验证逻辑
	// 例如：检查用户角色、导出数据量限制、时间段限制等
	
	// 基本验证：检查用户是否有导出权限
	// TODO: 可以从数据库查询用户权限配置
	
	return nil
}

// GetExportTemplate 获取导出模板（用于前端展示）
// GET /api/v1/export/template
func (h *ExportHandler) GetExportTemplate(w http.ResponseWriter, r *http.Request) {
	template := map[string]interface{}{
		"template": map[string]interface{}{
			"name":        "求职投递记录导出",
			"description": "导出您的求职投递记录到Excel文件",
			"options": map[string]interface{}{
				"formats": []string{"xlsx", "csv"},
				"maxRecords": 10000,
				"supportedFilters": []string{
					"status", "dateRange", "companyNames", "keywords",
				},
			},
		},
		"examples": map[string]interface{}{
			"basicExport": map[string]interface{}{
				"format": "xlsx",
				"fields": []string{
					"company_name", "position_title", "application_date", "status",
				},
			},
			"fullExport": map[string]interface{}{
				"format": "xlsx",
				"fields": []string{
					"company_name", "position_title", "application_date", "status",
					"salary_range", "work_location", "interview_time", "notes",
				},
				"options": map[string]interface{}{
					"includeStatistics": true,
					"filename":         "我的求职记录",
				},
			},
		},
	}

	h.writeSuccessResponse(w, http.StatusOK, "查询成功", template)
}