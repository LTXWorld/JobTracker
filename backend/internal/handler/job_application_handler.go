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

type JobApplicationHandler struct {
	service *service.JobApplicationService
}

func NewJobApplicationHandler(service *service.JobApplicationService) *JobApplicationHandler {
	return &JobApplicationHandler{service: service}
}

// Create 创建投递记录
func (h *JobApplicationHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	var req model.CreateJobApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// 验证输入
	if err := h.validateCreateRequest(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	job, err := h.service.Create(userID, &req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to create job application", err)
		return
	}

	h.writeSuccessResponse(w, http.StatusCreated, "job application created successfully", job)
}

// GetByID 获取单个投递记录
func (h *JobApplicationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing id parameter", nil)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid id parameter", err)
		return
	}

	job, err := h.service.GetByID(userID, id)
	if err != nil {
		if err.Error() == "job application not found" {
			h.writeErrorResponse(w, http.StatusNotFound, "job application not found", nil)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get job application", err)
		}
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "job application retrieved successfully", job)
}

// GetAll 获取所有投递记录
func (h *JobApplicationHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	jobs, err := h.service.GetAll(userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get job applications", err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "job applications retrieved successfully", jobs)
}

// Update 更新投递记录
func (h *JobApplicationHandler) Update(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing id parameter", nil)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid id parameter", err)
		return
	}

	var req model.UpdateJobApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// 验证输入
	if err := h.validateUpdateRequest(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	job, err := h.service.Update(userID, id, &req)
	if err != nil {
		if err.Error() == "job application not found" {
			h.writeErrorResponse(w, http.StatusNotFound, "job application not found", nil)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to update job application", err)
		}
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "job application updated successfully", job)
}

// Delete 删除投递记录
func (h *JobApplicationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing id parameter", nil)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid id parameter", err)
		return
	}

	if err := h.service.Delete(userID, id); err != nil {
		if err.Error() == "job application not found" {
			h.writeErrorResponse(w, http.StatusNotFound, "job application not found", nil)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to delete job application", err)
		}
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "job application deleted successfully", nil)
}

// GetStatistics 获取状态统计信息
func (h *JobApplicationHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	statistics, err := h.service.GetStatusStatistics(userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get statistics", err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "statistics retrieved successfully", statistics)
}

// writeSuccessResponse 写入成功响应
func (h *JobApplicationHandler) writeSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
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
func (h *JobApplicationHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
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

// validateCreateRequest 验证创建请求
func (h *JobApplicationHandler) validateCreateRequest(req *model.CreateJobApplicationRequest) error {
	if err := utils.ValidateCompanyName(req.CompanyName); err != nil {
		return err
	}
	
	if err := utils.ValidatePositionTitle(req.PositionTitle); err != nil {
		return err
	}
	
	if req.ApplicationDate != "" {
		if err := utils.ValidateDate(req.ApplicationDate); err != nil {
			return err
		}
	}
	
	if req.SalaryRange != nil {
		if err := utils.ValidateSalaryRange(*req.SalaryRange); err != nil {
			return err
		}
	}
	
	if req.WorkLocation != nil {
		if err := utils.ValidateWorkLocation(*req.WorkLocation); err != nil {
			return err
		}
	}
	
	if req.Notes != nil {
		if err := utils.ValidateNotes(*req.Notes); err != nil {
			return err
		}
	}
	
	if req.ContactInfo != nil {
		if err := utils.ValidateContactInfo(*req.ContactInfo); err != nil {
			return err
		}
	}
	
	return nil
}

// validateUpdateRequest 验证更新请求
func (h *JobApplicationHandler) validateUpdateRequest(req *model.UpdateJobApplicationRequest) error {
	if req.CompanyName != nil {
		if err := utils.ValidateCompanyName(*req.CompanyName); err != nil {
			return err
		}
	}
	
	if req.PositionTitle != nil {
		if err := utils.ValidatePositionTitle(*req.PositionTitle); err != nil {
			return err
		}
	}
	
	if req.ApplicationDate != nil {
		if err := utils.ValidateDate(*req.ApplicationDate); err != nil {
			return err
		}
	}
	
	if req.SalaryRange != nil {
		if err := utils.ValidateSalaryRange(*req.SalaryRange); err != nil {
			return err
		}
	}
	
	if req.WorkLocation != nil {
		if err := utils.ValidateWorkLocation(*req.WorkLocation); err != nil {
			return err
		}
	}
	
	if req.Notes != nil {
		if err := utils.ValidateNotes(*req.Notes); err != nil {
			return err
		}
	}
	
	if req.ContactInfo != nil {
		if err := utils.ValidateContactInfo(*req.ContactInfo); err != nil {
			return err
		}
	}
	
	return nil
}
