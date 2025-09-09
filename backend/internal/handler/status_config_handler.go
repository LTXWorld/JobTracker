// Location: /Users/lutao/GolandProjects/jobView/backend/internal/handler/status_config_handler.go
// This file implements HTTP handlers for status flow template and user preference management.
// It handles template CRUD operations, user preference settings, and status transition rule management.
// Used by the router to handle configuration related HTTP requests with proper authentication and validation.

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

type StatusConfigHandler struct {
	configService *service.StatusConfigService
}

func NewStatusConfigHandler(configService *service.StatusConfigService) *StatusConfigHandler {
	return &StatusConfigHandler{
		configService: configService,
	}
}

// GetStatusFlowTemplates 获取状态流转模板列表
// GET /api/v1/status-flow-templates
func (h *StatusConfigHandler) GetStatusFlowTemplates(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 调用服务获取模板列表
	templates, err := h.configService.GetStatusFlowTemplates(uint(userID))
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get flow templates", err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "flow templates retrieved successfully", templates)
}

// CreateStatusFlowTemplate 创建自定义状态流转模板
// POST /api/v1/status-flow-templates
func (h *StatusConfigHandler) CreateStatusFlowTemplate(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 解析请求体
	var req struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		FlowConfig  map[string]interface{} `json:"flow_config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// 验证请求
	if err := h.validateCreateTemplateRequest(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 调用服务创建模板
	template, err := h.configService.CreateStatusFlowTemplate(uint(userID), req.Name, req.Description, req.FlowConfig)
	if err != nil {
		if err.Error() == "template name '"+req.Name+"' already exists" {
			h.writeErrorResponse(w, http.StatusConflict, "template name already exists", nil)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to create flow template", err)
		}
		return
	}

	h.writeSuccessResponse(w, http.StatusCreated, "flow template created successfully", template)
}

// UpdateStatusFlowTemplate 更新状态流转模板
// PUT /api/v1/status-flow-templates/{id}
func (h *StatusConfigHandler) UpdateStatusFlowTemplate(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 获取模板ID
	vars := mux.Vars(r)
	templateIDStr, ok := vars["id"]
	if !ok {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing template id", nil)
		return
	}

	templateID, err := strconv.Atoi(templateIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid template id", err)
		return
	}

	// 解析请求体
	var req struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		FlowConfig  map[string]interface{} `json:"flow_config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// 验证请求
	if err := h.validateUpdateTemplateRequest(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 调用服务更新模板
	template, err := h.configService.UpdateStatusFlowTemplate(uint(userID), templateID, req.Name, req.Description, req.FlowConfig)
	if err != nil {
		if err.Error() == "template not found" {
			h.writeErrorResponse(w, http.StatusNotFound, "template not found", nil)
		} else if err.Error() == "cannot modify default template" {
			h.writeErrorResponse(w, http.StatusForbidden, "cannot modify default template", nil)
		} else if err.Error() == "permission denied: can only modify your own templates" {
			h.writeErrorResponse(w, http.StatusForbidden, "permission denied", nil)
		} else if err.Error() == "template name '"+req.Name+"' already exists" {
			h.writeErrorResponse(w, http.StatusConflict, "template name already exists", nil)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to update flow template", err)
		}
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "flow template updated successfully", template)
}

// DeleteStatusFlowTemplate 删除状态流转模板
// DELETE /api/v1/status-flow-templates/{id}
func (h *StatusConfigHandler) DeleteStatusFlowTemplate(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 获取模板ID
	vars := mux.Vars(r)
	templateIDStr, ok := vars["id"]
	if !ok {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing template id", nil)
		return
	}

	templateID, err := strconv.Atoi(templateIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid template id", err)
		return
	}

	// 调用服务删除模板
	err = h.configService.DeleteStatusFlowTemplate(uint(userID), templateID)
	if err != nil {
		if err.Error() == "template not found" {
			h.writeErrorResponse(w, http.StatusNotFound, "template not found", nil)
		} else if err.Error() == "cannot delete default template" {
			h.writeErrorResponse(w, http.StatusForbidden, "cannot delete default template", nil)
		} else if err.Error() == "permission denied: can only delete your own templates" {
			h.writeErrorResponse(w, http.StatusForbidden, "permission denied", nil)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to delete flow template", err)
		}
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "flow template deleted successfully", nil)
}

// GetUserStatusPreferences 获取用户状态偏好设置
// GET /api/v1/user-status-preferences
func (h *StatusConfigHandler) GetUserStatusPreferences(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 调用服务获取用户偏好
	preferences, err := h.configService.GetUserStatusPreferences(uint(userID))
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get user preferences", err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "user preferences retrieved successfully", preferences)
}

// UpdateUserStatusPreferences 更新用户状态偏好设置
// PUT /api/v1/user-status-preferences
func (h *StatusConfigHandler) UpdateUserStatusPreferences(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 解析请求体
	var req struct {
		PreferenceConfig map[string]interface{} `json:"preference_config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// 验证请求
	if err := h.validatePreferenceConfigRequest(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 调用服务更新偏好
	preferences, err := h.configService.UpdateUserStatusPreferences(uint(userID), req.PreferenceConfig)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to update user preferences", err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "user preferences updated successfully", preferences)
}

// GetAvailableStatusTransitions 获取指定状态的可用转换选项
// GET /api/v1/status-transitions/{status}
func (h *StatusConfigHandler) GetAvailableStatusTransitions(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 获取当前状态
	vars := mux.Vars(r)
	statusStr, ok := vars["status"]
	if !ok {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing status parameter", nil)
		return
	}

	currentStatus := model.ApplicationStatus(statusStr)
	if !currentStatus.IsValid() {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid status", nil)
		return
	}

	// 调用服务获取可用转换
	transitions, err := h.configService.GetAvailableStatusTransitions(uint(userID), currentStatus)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get available transitions", err)
		return
	}

	response := map[string]interface{}{
		"current_status":       currentStatus,
		"available_transitions": transitions,
		"transition_count":     len(transitions),
	}

	h.writeSuccessResponse(w, http.StatusOK, "available transitions retrieved successfully", response)
}

// GetAllStatusDefinitions 获取所有状态定义和分类
// GET /api/v1/status-definitions
func (h *StatusConfigHandler) GetAllStatusDefinitions(w http.ResponseWriter, r *http.Request) {
	// 定义状态分类
	statusDefinitions := map[string]interface{}{
		"categories": map[string][]string{
			"application": {"已投递"},
			"screening": {"简历筛选中", "简历筛选未通过"},
			"written_test": {"笔试中", "笔试通过", "笔试未通过"},
			"interviews": {
				"一面中", "一面通过", "一面未通过",
				"二面中", "二面通过", "二面未通过",
				"三面中", "三面通过", "三面未通过",
				"HR面中", "HR面通过", "HR面未通过",
			},
			"final": {
				"待发offer", "已收到offer", "已接受offer",
				"已拒绝", "流程结束",
			},
		},
		"all_statuses": []string{
			"已投递", "简历筛选中", "简历筛选未通过",
			"笔试中", "笔试通过", "笔试未通过",
			"一面中", "一面通过", "一面未通过",
			"二面中", "二面通过", "二面未通过",
			"三面中", "三面通过", "三面未通过",
			"HR面中", "HR面通过", "HR面未通过",
			"待发offer", "已收到offer", "已接受offer",
			"已拒绝", "流程结束",
		},
		"status_types": map[string][]string{
			"in_progress": {"已投递", "简历筛选中", "笔试中", "一面中", "二面中", "三面中", "HR面中"},
			"passed": {"笔试通过", "一面通过", "二面通过", "三面通过", "HR面通过", "待发offer", "已收到offer", "已接受offer", "流程结束"},
			"failed": {"简历筛选未通过", "笔试未通过", "一面未通过", "二面未通过", "三面未通过", "HR面未通过", "已拒绝"},
		},
	}

	h.writeSuccessResponse(w, http.StatusOK, "status definitions retrieved successfully", statusDefinitions)
}

// validateCreateTemplateRequest 验证创建模板请求
func (h *StatusConfigHandler) validateCreateTemplateRequest(req *struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	FlowConfig  map[string]interface{} `json:"flow_config"`
}) error {
	if req.Name == "" {
		return utils.NewValidationError("name", "template name is required")
	}

	if len(req.Name) > 100 {
		return utils.NewValidationError("name", "template name too long (max 100 characters)")
	}

	if len(req.Description) > 500 {
		return utils.NewValidationError("description", "description too long (max 500 characters)")
	}

	if req.FlowConfig == nil || len(req.FlowConfig) == 0 {
		return utils.NewValidationError("flow_config", "flow config is required")
	}

	return nil
}

// validateUpdateTemplateRequest 验证更新模板请求
func (h *StatusConfigHandler) validateUpdateTemplateRequest(req *struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	FlowConfig  map[string]interface{} `json:"flow_config"`
}) error {
	return h.validateCreateTemplateRequest(req)
}

// validatePreferenceConfigRequest 验证偏好配置请求
func (h *StatusConfigHandler) validatePreferenceConfigRequest(req *struct {
	PreferenceConfig map[string]interface{} `json:"preference_config"`
}) error {
	if req.PreferenceConfig == nil || len(req.PreferenceConfig) == 0 {
		return utils.NewValidationError("preference_config", "preference config is required")
	}

	return nil
}

// writeSuccessResponse 写入成功响应
func (h *StatusConfigHandler) writeSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
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
func (h *StatusConfigHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
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