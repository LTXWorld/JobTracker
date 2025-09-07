package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"jobView-backend/internal/auth"
	"jobView-backend/internal/handler"
	"jobView-backend/internal/model"
	"jobView-backend/internal/service"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAuthService 认证服务的Mock
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(req *model.RegisterRequest) (*model.LoginResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*model.LoginResponse), args.Error(1)
}

func (m *MockAuthService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*model.LoginResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(req *model.RefreshTokenRequest) (*model.LoginResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*model.LoginResponse), args.Error(1)
}

func (m *MockAuthService) GetProfile(userID uint) (*model.UserProfile, error) {
	args := m.Called(userID)
	return args.Get(0).(*model.UserProfile), args.Error(1)
}

func (m *MockAuthService) UpdateProfile(userID uint, req *model.UpdateUserRequest) (*model.UserProfile, error) {
	args := m.Called(userID, req)
	return args.Get(0).(*model.UserProfile), args.Error(1)
}

func (m *MockAuthService) ChangePassword(userID uint, req *model.ChangePasswordRequest) error {
	args := m.Called(userID, req)
	return args.Error(0)
}

func (m *MockAuthService) IsUsernameAvailable(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthService) IsEmailAvailable(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func setupAuthHandler() (*handler.AuthHandler, *MockAuthService) {
	// 设置测试环境变量
	os.Setenv("JWT_SECRET", "test-secret-key-for-handler-testing")
	
	mockService := &MockAuthService{}
	h := handler.NewAuthHandler(mockService)
	return h, mockService
}

func setupTestUser() *model.User {
	return &model.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}
}

func TestAuthHandler_Register(t *testing.T) {
	h, mockService := setupAuthHandler()

	t.Run("SuccessfulRegistration", func(t *testing.T) {
		req := &model.RegisterRequest{
			Username: "newuser",
			Email:    "newuser@example.com",
			Password: "password123",
		}

		expectedResponse := &model.LoginResponse{
			User: &model.UserProfile{
				ID:       1,
				Username: "newuser",
				Email:    "newuser@example.com",
			},
			Token:        "access-token",
			RefreshToken: "refresh-token",
			ExpiresAt:    1234567890,
		}

		mockService.On("Register", req).Return(expectedResponse, nil)

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Register(w, httpReq)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "注册成功")
		assert.Contains(t, w.Body.String(), "newuser")
		mockService.AssertExpectations(t)
	})

	t.Run("InvalidJSONRequest", func(t *testing.T) {
		httpReq := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBufferString("invalid json"))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Register(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "请求参数格式错误")
	})

	t.Run("ServiceError", func(t *testing.T) {
		req := &model.RegisterRequest{
			Username: "existinguser",
			Email:    "existing@example.com",
			Password: "password123",
		}

		mockService.On("Register", req).Return((*model.LoginResponse)(nil), errors.New("用户名已存在"))

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Register(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "用户名已存在")
		mockService.AssertExpectations(t)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	h, mockService := setupAuthHandler()

	t.Run("SuccessfulLogin", func(t *testing.T) {
		req := &model.LoginRequest{
			Username: "testuser",
			Password: "password123",
		}

		expectedResponse := &model.LoginResponse{
			User: &model.UserProfile{
				ID:       1,
				Username: "testuser",
				Email:    "test@example.com",
			},
			Token:        "access-token",
			RefreshToken: "refresh-token",
			ExpiresAt:    1234567890,
		}

		mockService.On("Login", req).Return(expectedResponse, nil)

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Login(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "登录成功")
		assert.Contains(t, w.Body.String(), "testuser")
		mockService.AssertExpectations(t)
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		req := &model.LoginRequest{
			Username: "testuser",
			Password: "wrongpassword",
		}

		mockService.On("Login", req).Return((*model.LoginResponse)(nil), errors.New("invalid credentials"))

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Login(w, httpReq)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "用户名或密码错误")
		mockService.AssertExpectations(t)
	})
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	h, mockService := setupAuthHandler()

	t.Run("SuccessfulRefresh", func(t *testing.T) {
		req := &model.RefreshTokenRequest{
			RefreshToken: "valid-refresh-token",
		}

		expectedResponse := &model.LoginResponse{
			Token:        "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresAt:    1234567890,
		}

		mockService.On("RefreshToken", req).Return(expectedResponse, nil)

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.RefreshToken(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "刷新token成功")
		assert.Contains(t, w.Body.String(), "new-access-token")
		mockService.AssertExpectations(t)
	})

	t.Run("InvalidRefreshToken", func(t *testing.T) {
		req := &model.RefreshTokenRequest{
			RefreshToken: "invalid-refresh-token",
		}

		mockService.On("RefreshToken", req).Return((*model.LoginResponse)(nil), errors.New("invalid refresh token"))

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.RefreshToken(w, httpReq)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "刷新token失败")
		mockService.AssertExpectations(t)
	})
}

func TestAuthHandler_GetProfile(t *testing.T) {
	h, mockService := setupAuthHandler()

	t.Run("SuccessfulGetProfile", func(t *testing.T) {
		userID := uint(1)
		expectedProfile := &model.UserProfile{
			ID:       1,
			Username: "testuser",
			Email:    "test@example.com",
		}

		mockService.On("GetProfile", userID).Return(expectedProfile, nil)

		httpReq := httptest.NewRequest("GET", "/api/auth/profile", nil)
		// 模拟认证中间件设置的上下文
		ctx := httpReq.Context()
		ctx = context.WithValue(ctx, auth.UserIDContextKey, userID)
		httpReq = httpReq.WithContext(ctx)
		w := httptest.NewRecorder()

		h.GetProfile(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "获取用户信息成功")
		assert.Contains(t, w.Body.String(), "testuser")
		mockService.AssertExpectations(t)
	})

	t.Run("NoUserInContext", func(t *testing.T) {
		httpReq := httptest.NewRequest("GET", "/api/auth/profile", nil)
		w := httptest.NewRecorder()

		h.GetProfile(w, httpReq)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "用户未认证")
	})
}

func TestAuthHandler_UpdateProfile(t *testing.T) {
	h, mockService := setupAuthHandler()

	t.Run("SuccessfulUpdate", func(t *testing.T) {
		userID := uint(1)
		req := &model.UpdateUserRequest{
			Username: "updateduser",
			Email:    "updated@example.com",
		}

		expectedProfile := &model.UserProfile{
			ID:       1,
			Username: "updateduser",
			Email:    "updated@example.com",
		}

		mockService.On("UpdateProfile", userID, req).Return(expectedProfile, nil)

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/api/auth/profile", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		// 模拟认证中间件设置的上下文
		ctx := httpReq.Context()
		ctx = context.WithValue(ctx, auth.UserIDContextKey, userID)
		httpReq = httpReq.WithContext(ctx)
		w := httptest.NewRecorder()

		h.UpdateProfile(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "更新用户信息成功")
		assert.Contains(t, w.Body.String(), "updateduser")
		mockService.AssertExpectations(t)
	})
}

func TestAuthHandler_ChangePassword(t *testing.T) {
	h, mockService := setupAuthHandler()

	t.Run("SuccessfulPasswordChange", func(t *testing.T) {
		userID := uint(1)
		req := &model.ChangePasswordRequest{
			CurrentPassword: "oldpassword",
			NewPassword:     "newpassword123",
		}

		mockService.On("ChangePassword", userID, req).Return(nil)

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/api/auth/change-password", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		// 模拟认证中间件设置的上下文
		ctx := httpReq.Context()
		ctx = context.WithValue(ctx, auth.UserIDContextKey, userID)
		httpReq = httpReq.WithContext(ctx)
		w := httptest.NewRecorder()

		h.ChangePassword(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "密码修改成功")
		mockService.AssertExpectations(t)
	})

	t.Run("InvalidCurrentPassword", func(t *testing.T) {
		userID := uint(1)
		req := &model.ChangePasswordRequest{
			CurrentPassword: "wrongpassword",
			NewPassword:     "newpassword123",
		}

		mockService.On("ChangePassword", userID, req).Return(errors.New("当前密码错误"))

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/api/auth/change-password", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		// 模拟认证中间件设置的上下文
		ctx := httpReq.Context()
		ctx = context.WithValue(ctx, auth.UserIDContextKey, userID)
		httpReq = httpReq.WithContext(ctx)
		w := httptest.NewRecorder()

		h.ChangePassword(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "当前密码错误")
		mockService.AssertExpectations(t)
	})
}

func TestAuthHandler_ValidateToken(t *testing.T) {
	h, _ := setupAuthHandler()

	t.Run("ValidToken", func(t *testing.T) {
		user := setupTestUser()
		accessToken, _, err := auth.GenerateTokenPair(user)
		require.NoError(t, err)

		claims, err := auth.ValidateAccessToken(accessToken)
		require.NoError(t, err)

		httpReq := httptest.NewRequest("GET", "/api/auth/validate", nil)
		// 模拟认证中间件设置的上下文
		ctx := httpReq.Context()
		ctx = context.WithValue(ctx, auth.UserContextKey, claims)
		httpReq = httpReq.WithContext(ctx)
		w := httptest.NewRecorder()

		h.ValidateToken(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "token有效")
		assert.Contains(t, w.Body.String(), "testuser")
	})

	t.Run("NoUserInContext", func(t *testing.T) {
		httpReq := httptest.NewRequest("GET", "/api/auth/validate", nil)
		w := httptest.NewRecorder()

		h.ValidateToken(w, httpReq)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "token无效")
	})
}

func TestAuthHandler_CheckUsernameAvailability(t *testing.T) {
	h, mockService := setupAuthHandler()

	t.Run("UsernameAvailable", func(t *testing.T) {
		mockService.On("IsUsernameAvailable", "newuser").Return(true, nil)

		httpReq := httptest.NewRequest("GET", "/api/auth/check-username?username=newuser", nil)
		w := httptest.NewRecorder()

		h.CheckUsernameAvailability(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "用户名可用")
		mockService.AssertExpectations(t)
	})

	t.Run("UsernameNotAvailable", func(t *testing.T) {
		mockService.On("IsUsernameAvailable", "existinguser").Return(false, nil)

		httpReq := httptest.NewRequest("GET", "/api/auth/check-username?username=existinguser", nil)
		w := httptest.NewRecorder()

		h.CheckUsernameAvailability(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "用户名已被使用")
		mockService.AssertExpectations(t)
	})

	t.Run("InvalidUsernameLength", func(t *testing.T) {
		httpReq := httptest.NewRequest("GET", "/api/auth/check-username?username=ab", nil) // 太短
		w := httptest.NewRecorder()

		h.CheckUsernameAvailability(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "用户名长度必须在3-20个字符之间")
	})

	t.Run("InvalidUsernameFormat", func(t *testing.T) {
		httpReq := httptest.NewRequest("GET", "/api/auth/check-username?username=user@name", nil) // 含特殊字符
		w := httptest.NewRecorder()

		h.CheckUsernameAvailability(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "用户名只能包含字母、数字和下划线")
	})
}

func TestAuthHandler_CheckEmailAvailability(t *testing.T) {
	h, mockService := setupAuthHandler()

	t.Run("EmailAvailable", func(t *testing.T) {
		mockService.On("IsEmailAvailable", "new@example.com").Return(true, nil)

		httpReq := httptest.NewRequest("GET", "/api/auth/check-email?email=new@example.com", nil)
		w := httptest.NewRecorder()

		h.CheckEmailAvailability(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "邮箱可用")
		mockService.AssertExpectations(t)
	})

	t.Run("InvalidEmailFormat", func(t *testing.T) {
		httpReq := httptest.NewRequest("GET", "/api/auth/check-email?email=invalid-email", nil)
		w := httptest.NewRecorder()

		h.CheckEmailAvailability(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "邮箱格式不正确")
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	h, _ := setupAuthHandler()

	t.Run("SuccessfulLogout", func(t *testing.T) {
		userID := uint(1)

		httpReq := httptest.NewRequest("POST", "/api/auth/logout", nil)
		// 模拟认证中间件设置的上下文
		ctx := httpReq.Context()
		ctx = context.WithValue(ctx, auth.UserIDContextKey, userID)
		httpReq = httpReq.WithContext(ctx)
		w := httptest.NewRecorder()

		h.Logout(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "登出成功")
	})

	t.Run("LogoutWithoutAuth", func(t *testing.T) {
		httpReq := httptest.NewRequest("POST", "/api/auth/logout", nil)
		w := httptest.NewRecorder()

		h.Logout(w, httpReq)

		// 即使没有认证也应该成功（客户端清理）
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "登出成功")
	})
}

func TestAuthHandler_HealthCheck(t *testing.T) {
	h, _ := setupAuthHandler()

	t.Run("HealthCheckReturnsOK", func(t *testing.T) {
		httpReq := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		h.HealthCheck(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "服务正常")
		assert.Contains(t, w.Body.String(), "auth")
		assert.Contains(t, w.Body.String(), "1.0.0")
	})
}

func TestAuthHandler_GetUserStats(t *testing.T) {
	h, _ := setupAuthHandler()

	t.Run("SuccessfulGetStats", func(t *testing.T) {
		userID := uint(1)

		httpReq := httptest.NewRequest("GET", "/api/auth/stats", nil)
		// 模拟认证中间件设置的上下文
		ctx := httpReq.Context()
		ctx = context.WithValue(ctx, auth.UserIDContextKey, userID)
		httpReq = httpReq.WithContext(ctx)
		w := httptest.NewRecorder()

		h.GetUserStats(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "获取用户统计信息成功")
		assert.Contains(t, w.Body.String(), "user_id")
		assert.Contains(t, w.Body.String(), "last_login")
	})
}

// Integration test for multiple handlers
func TestAuthHandler_IntegrationFlow(t *testing.T) {
	h, mockService := setupAuthHandler()

	// 注册用户
	registerReq := &model.RegisterRequest{
		Username: "integrationuser",
		Email:    "integration@example.com",
		Password: "password123",
	}

	registerResponse := &model.LoginResponse{
		User: &model.UserProfile{
			ID:       1,
			Username: "integrationuser",
			Email:    "integration@example.com",
		},
		Token:        "access-token",
		RefreshToken: "refresh-token",
		ExpiresAt:    1234567890,
	}

	mockService.On("Register", registerReq).Return(registerResponse, nil)

	// 测试注册
	reqBody, _ := json.Marshal(registerReq)
	httpReq := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, httpReq)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "注册成功")

	// 验证所有 mock 调用都被执行
	mockService.AssertExpectations(t)
}

// Error handling tests
func TestAuthHandler_ErrorHandling(t *testing.T) {
	h, mockService := setupAuthHandler()

	t.Run("ServiceUnavailableError", func(t *testing.T) {
		req := &model.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		mockService.On("Register", req).Return((*model.LoginResponse)(nil), errors.New("数据库连接失败"))

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Register(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "数据库连接失败")
		mockService.AssertExpectations(t)
	})
}

// Performance test for auth handler
func BenchmarkAuthHandler_Register(b *testing.B) {
	h, mockService := setupAuthHandler()

	req := &model.RegisterRequest{
		Username: "benchuser",
		Email:    "bench@example.com",
		Password: "password123",
	}

	response := &model.LoginResponse{
		User: &model.UserProfile{
			ID:       1,
			Username: "benchuser",
			Email:    "bench@example.com",
		},
		Token:        "access-token",
		RefreshToken: "refresh-token",
		ExpiresAt:    1234567890,
	}

	// 设置 mock 期望，允许多次调用
	mockService.On("Register", req).Return(response, nil).Maybe()

	reqBody, _ := json.Marshal(req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		httpReq := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Register(w, httpReq)
	}
}

func TestMain(m *testing.M) {
	// 设置测试环境变量
	os.Setenv("JWT_SECRET", "test-secret-key-for-handler-testing-main")
	
	// 运行测试
	code := m.Run()
	
	// 清理
	os.Unsetenv("JWT_SECRET")
	os.Exit(code)
}
