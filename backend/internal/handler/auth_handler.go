// /Users/lutao/GolandProjects/jobView/backend/internal/handler/auth_handler.go
// 认证处理器，负责处理用户认证相关的HTTP请求
// 包含注册、登录、刷新token、获取和更新用户信息等接口处理逻辑

package handler

import (
	"encoding/json"
	"jobView-backend/internal/auth"
	"jobView-backend/internal/model"
	"jobView-backend/internal/service"
	"log"
	"net/http"
	"regexp"
	"time"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register 用户注册
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[AUTH] Register JSON decode error: %v", err)
		h.writeErrorResponse(w, http.StatusBadRequest, "请求参数格式错误", err)
		return
	}
	
	// 记录注册尝试
	log.Printf("[AUTH] Registration attempt for username: %s, email: %s", 
		req.Username, req.Email)
	
	response, err := h.service.Register(&req)
	if err != nil {
		log.Printf("[AUTH] Registration failed for username: %s - %v", req.Username, err)
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	
	h.writeSuccessResponse(w, http.StatusCreated, "注册成功", response)
}

// Login 用户登录
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "请求参数格式错误", err)
		return
	}
	
	// 记录登录尝试（不记录密码）
	log.Printf("[AUTH] Login attempt for username: %s", req.Username)
	
	response, err := h.service.Login(&req)
	if err != nil {
		// 登录失败不暴露具体错误信息给客户端
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户名或密码错误", nil)
		return
	}
	
	h.writeSuccessResponse(w, http.StatusOK, "登录成功", response)
}

// RefreshToken 刷新token
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req model.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "请求参数格式错误", err)
		return
	}
	
	response, err := h.service.RefreshToken(&req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "刷新token失败", nil)
		return
	}
	
	h.writeSuccessResponse(w, http.StatusOK, "刷新token成功", response)
}

// GetProfile 获取用户信息
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}
	
	profile, err := h.service.GetProfile(userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "获取用户信息失败", err)
		return
	}
	
	h.writeSuccessResponse(w, http.StatusOK, "获取用户信息成功", profile)
}

// UpdateProfile 更新用户信息
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}
	
	var req model.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "请求参数格式错误", err)
		return
	}
	
	profile, err := h.service.UpdateProfile(userID, &req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	
	h.writeSuccessResponse(w, http.StatusOK, "更新用户信息成功", profile)
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}
	
	var req model.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "请求参数格式错误", err)
		return
	}
	
	err := h.service.ChangePassword(userID, &req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	
	h.writeSuccessResponse(w, http.StatusOK, "密码修改成功", nil)
}

// Logout 用户登出（主要是清除客户端token）
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if ok {
		log.Printf("[AUTH] User logged out: ID=%d", userID)
	}
	
	// 在实际项目中，可以在这里将token加入黑名单
	// 或者记录登出时间等操作
	
	h.writeSuccessResponse(w, http.StatusOK, "登出成功", nil)
}

// ValidateToken 验证token有效性（用于前端检查）
func (h *AuthHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "token无效", nil)
		return
	}
	
	// 检查token是否即将过期
	isExpiring := auth.IsTokenExpired(user)
	
	response := map[string]interface{}{
		"valid":      true,
		"user_id":    user.UserID,
		"username":   user.Username,
		"expires_at": user.ExpiresAt.Time,
		"expiring":   isExpiring, // 如果在30分钟内过期则为true
	}
	
	h.writeSuccessResponse(w, http.StatusOK, "token有效", response)
}

// GetUserStats 获取用户统计信息（可选功能）
func (h *AuthHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "用户未认证", nil)
		return
	}
	
	// 这里可以扩展获取用户的求职统计信息
	// 比如投递数量、面试次数等
	stats := map[string]interface{}{
		"user_id":         userID,
		"last_login":      time.Now().Format("2006-01-02 15:04:05"),
		"total_applications": 0, // 这里可以查询实际数据
		"active_processes": 0,   // 这里可以查询实际数据
	}
	
	h.writeSuccessResponse(w, http.StatusOK, "获取用户统计信息成功", stats)
}

// HealthCheck 健康检查（不需要认证）
func (h *AuthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "ok",
		"service":   "auth",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	}
	
	h.writeSuccessResponse(w, http.StatusOK, "服务正常", health)
}

// writeSuccessResponse 写入成功响应
func (h *AuthHandler) writeSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := model.APIResponse{
		Code:    statusCode,
		Message: message,
		Data:    data,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ERROR] Failed to encode response: %v", err)
	}
}

// writeErrorResponse 写入错误响应
func (h *AuthHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := model.APIResponse{
		Code:    statusCode,
		Message: message,
	}
	
	// 只在开发环境或内部服务器错误时显示详细错误信息
	if err != nil && statusCode >= 500 {
		response.Data = map[string]string{"error": err.Error()}
		log.Printf("[ERROR] Internal server error: %v", err)
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ERROR] Failed to encode error response: %v", err)
	}
}

// RateLimitedHandler 带速率限制的处理器包装器
func (h *AuthHandler) RateLimitedHandler(handler http.HandlerFunc, requests int, window time.Duration) http.HandlerFunc {
	middleware := auth.RateLimitMiddleware(requests, window)
	return middleware(handler).ServeHTTP
}

// Helper方法：从请求中安全地提取IP地址
func (h *AuthHandler) getClientIP(r *http.Request) string {
	// 优先从X-Forwarded-For头获取
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	
	// 从X-Real-IP头获取
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// 最后从RemoteAddr获取
	return r.RemoteAddr
}

// CheckUsernameAvailability 检查用户名可用性
func (h *AuthHandler) CheckUsernameAvailability(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "用户名参数缺失", nil)
		return
	}

	// 验证用户名格式
	if len(username) < 3 || len(username) > 20 {
		h.writeErrorResponse(w, http.StatusBadRequest, "用户名长度必须在3-20个字符之间", nil)
		return
	}

	// 检查用户名格式（只允许字母、数字和下划线）
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(username) {
		h.writeErrorResponse(w, http.StatusBadRequest, "用户名只能包含字母、数字和下划线", nil)
		return
	}

	// 记录检查请求
	log.Printf("[AUTH] Username availability check for: %s", username)

	// 检查用户名是否已存在
	available, err := h.service.IsUsernameAvailable(username)
	if err != nil {
		log.Printf("[ERROR] Failed to check username availability: %v", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "检查用户名可用性失败", err)
		return
	}

	response := map[string]interface{}{
		"available": available,
		"message":   getAvailabilityMessage(available, "用户名"),
	}

	h.writeSuccessResponse(w, http.StatusOK, "检查完成", response)
}

// CheckEmailAvailability 检查邮箱可用性
func (h *AuthHandler) CheckEmailAvailability(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "邮箱参数缺失", nil)
		return
	}

	// 验证邮箱格式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		h.writeErrorResponse(w, http.StatusBadRequest, "邮箱格式不正确", nil)
		return
	}

	// 记录检查请求
	log.Printf("[AUTH] Email availability check for: %s", email)

	// 检查邮箱是否已存在
	available, err := h.service.IsEmailAvailable(email)
	if err != nil {
		log.Printf("[ERROR] Failed to check email availability: %v", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "检查邮箱可用性失败", err)
		return
	}

	response := map[string]interface{}{
		"available": available,
		"message":   getAvailabilityMessage(available, "邮箱"),
	}

	h.writeSuccessResponse(w, http.StatusOK, "检查完成", response)
}

// getAvailabilityMessage 辅助函数：生成可用性消息
func getAvailabilityMessage(available bool, resourceType string) string {
	if available {
		return resourceType + "可用"
	}
	return resourceType + "已被使用"
}