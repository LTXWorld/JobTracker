package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"jobView-backend/internal/auth"
	"jobView-backend/internal/model"
	"jobView-backend/internal/service"
)

// MailEventHandler 负责提醒中心邮件事件相关接口
type MailEventHandler struct {
	service *service.MailEventService
}

func NewMailEventHandler(service *service.MailEventService) *MailEventHandler {
	return &MailEventHandler{service: service}
}

// ListPending 返回待确认事件
func (h *MailEventHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "用户未认证")
		return
	}

	events, err := h.service.ListPendingEvents(userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeSuccess(w, http.StatusOK, "获取待确认事件成功", events)
}

// UpdateStatus 更新事件状态
func (h *MailEventHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "用户未认证")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "事件ID格式不正确")
		return
	}

	var req model.MailEventStatusUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "请求体格式不正确")
		return
	}

	updated, err := h.service.UpdateEventStatus(userID, id, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMailEventEmptyRequest),
			errors.Is(err, service.ErrMailEventStatusEmpty),
			errors.Is(err, service.ErrMailEventStatusInvalid):
			h.writeError(w, http.StatusBadRequest, err.Error())
			return
		case errors.Is(err, service.ErrMailEventNotFound):
			h.writeError(w, http.StatusNotFound, "事件不存在或无权访问")
			return
		default:
			h.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	h.writeSuccess(w, http.StatusOK, "事件状态更新成功", updated)
}

func (h *MailEventHandler) writeSuccess(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := model.APIResponse{Code: status, Message: message, Data: data}
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *MailEventHandler) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := model.APIResponse{Code: status, Message: message}
	_ = json.NewEncoder(w).Encode(resp)
}
