package handler

import (
	"encoding/json"
	"net/http"

	"jobView-backend/internal/auth"
	"jobView-backend/internal/model"
	"jobView-backend/internal/service"
)

// EmailIntegrationHandler 处理邮箱绑定相关请求
type EmailIntegrationHandler struct {
	service *service.EmailIntegrationService
}

func NewEmailIntegrationHandler(s *service.EmailIntegrationService) *EmailIntegrationHandler {
	return &EmailIntegrationHandler{service: s}
}

// BindMailbox 创建或更新邮箱授权
func (h *EmailIntegrationHandler) BindMailbox(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}
	var req model.MailboxBindRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求体格式不正确", err)
		return
	}
	result, err := h.service.BindMailbox(userID, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	writeSuccess(w, http.StatusOK, "邮箱授权保存成功", result)
}

// GetMailbox 获取当前绑定信息
func (h *EmailIntegrationHandler) GetMailbox(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}
	result, err := h.service.GetMailbox(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if result == nil {
		writeSuccess(w, http.StatusOK, "邮箱尚未绑定", nil)
		return
	}
	writeSuccess(w, http.StatusOK, "获取邮箱绑定信息成功", result)
}

// RemoveMailbox 解除绑定
func (h *EmailIntegrationHandler) RemoveMailbox(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}
	if err := h.service.RemoveMailbox(userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	writeSuccess(w, http.StatusOK, "邮箱已解除绑定", nil)
}

func writeSuccess(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := model.APIResponse{Code: status, Message: message, Data: data}
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := model.APIResponse{Code: status, Message: message}
	if err != nil && status >= 500 {
		resp.Data = map[string]string{"error": err.Error()}
	}
	json.NewEncoder(w).Encode(resp)
}
