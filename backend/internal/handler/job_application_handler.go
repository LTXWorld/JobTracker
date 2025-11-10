package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"jobView-backend/internal/auth"
	"jobView-backend/internal/model"
	"jobView-backend/internal/utils"

	"github.com/gorilla/mux"
)

// JobApplicationUseCase 定义 handler 所依赖的领域服务接口
type JobApplicationUseCase interface {
	Create(userID uint, req *model.CreateJobApplicationRequest) (*model.JobApplication, error)
	GetByID(userID uint, id int) (*model.JobApplication, error)
	GetAll(userID uint) ([]model.JobApplication, error)
	Update(userID uint, id int, req *model.UpdateJobApplicationRequest) (*model.JobApplication, error)
	Delete(userID uint, id int) error
	GetStatusStatistics(userID uint, processType model.ApplicationProcessType) (map[string]interface{}, error)
	GetJobApplicationsWithStatusFilters(userID uint, status *model.ApplicationStatus, stage *string, req model.PaginationRequest) (*model.PaginationResponse, error)
	SearchApplications(userID uint, query string, req model.PaginationRequest) (*model.PaginationResponse, error)
	GetDashboardData(userID uint, processType model.ApplicationProcessType) (map[string]interface{}, error)
}

type JobApplicationHandler struct {
	service JobApplicationUseCase
}

func NewJobApplicationHandler(service JobApplicationUseCase) *JobApplicationHandler {
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

	// 便于排查500，记录关键字段变化（不包含大字段）
	log.Printf("[DEBUG] Update application id=%d by user=%d payload=%v", id, userID, map[string]interface{}{
		"status":            req.Status,
		"company_name":      req.CompanyName,
		"position_title":    req.PositionTitle,
		"application_date":  req.ApplicationDate,
		"company_attribute": req.CompanyAttribute,
	})

	job, err := h.service.Update(userID, id, &req)
	if err != nil {
		// 打印后端实际错误，便于定位500根因
		log.Printf("[ERROR] Update failed for application id=%d user=%d: %v", id, userID, err)
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

	processType := h.resolveProcessTypeParam(r.URL.Query().Get("process_type"))

	statistics, err := h.service.GetStatusStatistics(userID, processType)
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
		response.Data = map[string]string{"error": fmt.Sprintf("%v", err)}
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

	// 企业属性：新建为必填
	if err := utils.ValidateCompanyAttribute(req.CompanyAttribute, true); err != nil {
		return err
	}

	if req.ProcessType != "" && !req.ProcessType.IsValid() {
		return fmt.Errorf("求职周期类型无效")
	}
	if req.ProcessType == "" {
		req.ProcessType = model.ProcessTypeAutumn
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

	if req.CompanyAttribute != nil {
		if err := utils.ValidateCompanyAttribute(*req.CompanyAttribute, false); err != nil {
			return err
		}
	}

	if req.ProcessType != nil && !req.ProcessType.IsValid() {
		return fmt.Errorf("求职周期类型无效")
	}

	return nil
}

func (h *JobApplicationHandler) resolveProcessTypeParam(raw string) model.ApplicationProcessType {
	value := strings.TrimSpace(raw)
	if value == "" {
		return model.ProcessTypeAutumn
	}
	pt := model.ApplicationProcessType(value)
	if !pt.IsValid() {
		return model.ProcessTypeAutumn
	}
	return pt
}

// GetJobApplicationsWithFilters 根据状态和阶段筛选获取岗位申请
// GET /api/v1/applications?status={status}&stage={stage}
func (h *JobApplicationHandler) GetJobApplicationsWithFilters(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 添加调试日志
	log.Printf("[DEBUG] User %d requesting applications with URL: %s", userID, r.URL.String())

	// 解析查询参数
	var req model.PaginationRequest
	req.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	req.PageSize, _ = strconv.Atoi(r.URL.Query().Get("page_size"))
	req.SortBy = r.URL.Query().Get("sort_by")
	req.SortDir = r.URL.Query().Get("sort_dir")

	log.Printf("[DEBUG] Pagination params: page=%d, page_size=%d", req.Page, req.PageSize)

	// 解析状态筛选
	var status *model.ApplicationStatus
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		appStatus := model.ApplicationStatus(statusStr)
		if appStatus.IsValid() {
			status = &appStatus
		}
	}

	// 解析阶段筛选
	var stage *string
	if stageStr := r.URL.Query().Get("stage"); stageStr != "" {
		validStages := map[string]bool{
			"application": true, "screening": true, "written_test": true,
			"interviews": true, "final": true,
			"in_progress": true, "passed": true, "failed": true,
		}
		if validStages[stageStr] {
			stage = &stageStr
		}
	}

	processType := h.resolveProcessTypeParam(r.URL.Query().Get("process_type"))
	req.ProcessType = &processType

	// 调用服务获取筛选结果
	result, err := h.service.GetJobApplicationsWithStatusFilters(userID, status, stage, req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get filtered job applications", err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "filtered job applications retrieved successfully", result)
}

// SearchJobApplications 搜索岗位申请
// GET /api/v1/applications/search?q={query}&filters={filters}
func (h *JobApplicationHandler) SearchJobApplications(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	// 获取搜索查询
	query := r.URL.Query().Get("q")
	if query == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "search query is required", nil)
		return
	}

	// 解析分页参数
	var req model.PaginationRequest
	req.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	req.PageSize, _ = strconv.Atoi(r.URL.Query().Get("page_size"))
	req.SortBy = r.URL.Query().Get("sort_by")
	req.SortDir = r.URL.Query().Get("sort_dir")

	// 解析状态过滤器
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		appStatus := model.ApplicationStatus(statusStr)
		if appStatus.IsValid() {
			req.Status = &appStatus
		}
	}

	processType := h.resolveProcessTypeParam(r.URL.Query().Get("process_type"))
	req.ProcessType = &processType

	// 调用服务进行搜索
	result, err := h.service.SearchApplications(userID, query, req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to search job applications", err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "job applications search completed successfully", result)
}

// GetDashboardData 获取仪表板数据
// GET /api/v1/applications/dashboard
func (h *JobApplicationHandler) GetDashboardData(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}

	processType := h.resolveProcessTypeParam(r.URL.Query().Get("process_type"))

	// 调用服务获取仪表板数据
	dashboard, err := h.service.GetDashboardData(userID, processType)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get dashboard data", err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, "dashboard data retrieved successfully", dashboard)
}
