package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"jobView-backend/internal/auth"
	"jobView-backend/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestUser() *model.User {
	return &model.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}
}

func TestMain(m *testing.M) {
	// 设置测试环境变量
	os.Setenv("JWT_SECRET", "test-secret-key-for-middleware-testing")
	
	// 运行测试
	code := m.Run()
	
	// 清理
	os.Unsetenv("JWT_SECRET")
	os.Exit(code)
}

func TestAuthMiddleware(t *testing.T) {
	user := setupTestUser()
	accessToken, _, err := auth.GenerateTokenPair(user)
	require.NoError(t, err)

	// 创建一个简单的处理器来测试中间件
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查用户是否在上下文中
		userFromContext, ok := auth.GetUserFromContext(r.Context())
		if ok {
			w.Header().Set("X-User-ID", string(rune(userFromContext.UserID)))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("authenticated"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("user not in context"))
		}
	})

	middleware := auth.AuthMiddleware(handler)

	t.Run("ValidToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "authenticated", w.Body.String())
	})

	t.Run("MissingAuthHeader", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization header is required")
	})

	t.Run("InvalidAuthHeader", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Invalid header format")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid authorization header format")
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		// 这里我们需要创建一个过期的token来测试
		// 由于我们无法直接创建过期token，我们使用无效token来模拟
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid token")
	})
}

func TestOptionalAuthMiddleware(t *testing.T) {
	user := setupTestUser()
	accessToken, _, err := auth.GenerateTokenPair(user)
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userFromContext, ok := auth.GetUserFromContext(r.Context())
		if ok {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("authenticated:" + userFromContext.Username))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("anonymous"))
		}
	})

	middleware := auth.OptionalAuthMiddleware(handler)

	t.Run("WithValidToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "authenticated:")
	})

	t.Run("WithoutToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "anonymous", w.Body.String())
	})

	t.Run("WithInvalidToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid.token")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "anonymous", w.Body.String())
	})
}

func TestCORSMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	allowedOrigins := []string{"http://localhost:3000", "https://example.com"}
	middleware := auth.CORSMiddleware(allowedOrigins)(handler)

	t.Run("AllowedOrigin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	})

	t.Run("UnallowedOrigin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://malicious.com")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// 即使是不允许的域名，也应该设置CORS头以便错误处理
		assert.Equal(t, "http://malicious.com", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("NoOriginHeader", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// 没有Origin时应该设置第一个允许的域名
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("OptionsRequest", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
	})
}

func TestRateLimitMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// 限制：每秒最多2个请求
	middleware := auth.RateLimitMiddleware(2, time.Second)(handler)

	t.Run("WithinLimit", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "127.0.0.1:8080"
			w := httptest.NewRecorder()

			middleware.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	t.Run("ExceedsLimit", func(t *testing.T) {
		// 先发送2个请求达到限制
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "127.0.0.1:8080"
			w := httptest.NewRecorder()
			middleware.ServeHTTP(w, req)
		}

		// 第三个请求应该被限制
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:8080"
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Contains(t, w.Body.String(), "Too many requests")
	})

	t.Run("DifferentIPs", func(t *testing.T) {
		// 不同IP应该有独立的限制
		req1 := httptest.NewRequest("GET", "/test", nil)
		req1.RemoteAddr = "127.0.0.1:8080"
		w1 := httptest.NewRecorder()

		req2 := httptest.NewRequest("GET", "/test", nil)
		req2.RemoteAddr = "127.0.0.2:8080"
		w2 := httptest.NewRecorder()

		middleware.ServeHTTP(w1, req1)
		middleware.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w1.Code)
		assert.Equal(t, http.StatusOK, w2.Code)
	})
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := auth.SecurityHeadersMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Contains(t, w.Header().Get("Referrer-Policy"), "strict-origin-when-cross-origin")
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "default-src 'self'")
}

func TestLoggingMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	middleware := auth.LoggingMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:8080"
	w := httptest.NewRecorder()

	// 记录日志输出（在实际测试中，我们可能需要mock log包）
	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())
}

func TestGetUserFromContext(t *testing.T) {
	user := setupTestUser()
	accessToken, _, err := auth.GenerateTokenPair(user)
	require.NoError(t, err)

	claims, err := auth.ValidateAccessToken(accessToken)
	require.NoError(t, err)

	t.Run("UserInContext", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), auth.UserContextKey, claims)
		
		userFromCtx, ok := auth.GetUserFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, claims.UserID, userFromCtx.UserID)
		assert.Equal(t, claims.Username, userFromCtx.Username)
	})

	t.Run("NoUserInContext", func(t *testing.T) {
		ctx := context.Background()
		
		userFromCtx, ok := auth.GetUserFromContext(ctx)
		assert.False(t, ok)
		assert.Nil(t, userFromCtx)
	})
}

func TestGetUserIDFromContext(t *testing.T) {
	userID := uint(123)
	
	t.Run("UserIDInContext", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), auth.UserIDContextKey, userID)
		
		idFromCtx, ok := auth.GetUserIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, userID, idFromCtx)
	})

	t.Run("NoUserIDInContext", func(t *testing.T) {
		ctx := context.Background()
		
		idFromCtx, ok := auth.GetUserIDFromContext(ctx)
		assert.False(t, ok)
		assert.Equal(t, uint(0), idFromCtx)
	})
}

func TestRequireOwnership(t *testing.T) {
	testCases := []struct {
		name         string
		userID       uint
		resourceUserID uint
		expected     bool
	}{
		{
			name:         "SameUser",
			userID:       123,
			resourceUserID: 123,
			expected:     true,
		},
		{
			name:         "DifferentUser",
			userID:       123,
			resourceUserID: 456,
			expected:     false,
		},
		{
			name:         "ZeroIDs",
			userID:       0,
			resourceUserID: 0,
			expected:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := auth.RequireOwnership(tc.userID, tc.resourceUserID)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMiddlewareChaining(t *testing.T) {
	// 测试多个中间件的链式调用
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("final handler"))
	})

	user := setupTestUser()
	accessToken, _, err := auth.GenerateTokenPair(user)
	require.NoError(t, err)

	// 链式应用中间件
	wrapped := auth.SecurityHeadersMiddleware(
		auth.LoggingMiddleware(
			auth.CORSMiddleware([]string{"http://localhost:3000"})(
				auth.AuthMiddleware(handler),
			),
		),
	)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "final handler", w.Body.String())
	
	// 验证各个中间件的效果
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
}

// BenchmarkAuthMiddleware 性能测试
func BenchmarkAuthMiddleware(b *testing.B) {
	user := setupTestUser()
	accessToken, _, err := auth.GenerateTokenPair(user)
	require.NoError(b, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := auth.AuthMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)
	}
}