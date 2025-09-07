// /Users/lutao/GolandProjects/jobView/backend/internal/auth/middleware.go  
// 认证和授权中间件，负责保护API端点，确保只有认证用户可以访问
// 提供JWT token验证、用户身份提取和权限控制功能

package auth

import (
	"context"
	"encoding/json"
	"errors"
	"jobView-backend/internal/model"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ContextKey 上下文键类型
type ContextKey string

const (
	UserContextKey ContextKey = "user"
	UserIDContextKey ContextKey = "user_id"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 从请求头获取Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeErrorResponse(w, http.StatusUnauthorized, "Authorization header is required", nil)
			return
		}
		
		// 提取token
		token, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			writeErrorResponse(w, http.StatusUnauthorized, "Invalid authorization header format", err)
			return
		}
		
		// 验证token
		claims, err := ValidateAccessToken(token)
		if err != nil {
			if err == ErrTokenExpired {
				writeErrorResponse(w, http.StatusUnauthorized, "Token has expired", err)
			} else {
				writeErrorResponse(w, http.StatusUnauthorized, "Invalid token", err)
			}
			return
		}
		
		// 将用户信息添加到请求上下文
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)
		
		// 记录认证日志
		log.Printf("[AUTH] User %d (%s) accessing %s %s", 
			claims.UserID, claims.Username, r.Method, r.URL.Path)
		
		// 继续处理请求
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthMiddleware 可选认证中间件（对于某些可以匿名访问的端点）
func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// 没有认证信息，继续处理请求
			next.ServeHTTP(w, r)
			return
		}
		
		token, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			// token格式错误，继续处理请求但不设置用户信息
			next.ServeHTTP(w, r)
			return
		}
		
		claims, err := ValidateAccessToken(token)
		if err != nil {
			// token验证失败，继续处理请求但不设置用户信息
			next.ServeHTTP(w, r)
			return
		}
		
		// 将用户信息添加到请求上下文
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RateLimitMiddleware 简单的速率限制中间件
func RateLimitMiddleware(requests int, window time.Duration) func(http.Handler) http.Handler {
	// 简单的内存限流实现（生产环境建议使用Redis）
	clients := make(map[string][]time.Time)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 使用IP作为限流键
			ip := getClientIP(r)
			now := time.Now()
			
			// 清理过期记录
			if records, exists := clients[ip]; exists {
				validRecords := []time.Time{}
				for _, record := range records {
					if now.Sub(record) <= window {
						validRecords = append(validRecords, record)
					}
				}
				clients[ip] = validRecords
			}
			
			// 检查请求数量
			if len(clients[ip]) >= requests {
				writeErrorResponse(w, http.StatusTooManyRequests, "Too many requests", nil)
				return
			}
			
			// 记录当前请求
			clients[ip] = append(clients[ip], now)
			
			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware CORS中间件（增强版）
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// 检查允许的域名
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin || allowedOrigin == "*" {
					allowed = true
					break
				}
			}
			
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if origin != "" {
				// 即使不允许的域名也设置基本的 CORS 头以便错误处理
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				// 如果没有Origin头，设置默认允许的第一个域名
				if len(allowedOrigins) > 0 {
					w.Header().Set("Access-Control-Allow-Origin", allowedOrigins[0])
				}
			}
			
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Accept-Encoding, Accept-Language")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24小时预检缓存
			
			// 处理预检请求
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersMiddleware 安全响应头中间件
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置安全响应头
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		
		// HSTS（HTTPS严格传输安全）
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware 请求日志中间件
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// 包装ResponseWriter以捕获状态码
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapped, r)
		
		duration := time.Since(start)
		log.Printf("[HTTP] %s %s - %d - %v - %s", 
			r.Method, r.URL.Path, wrapped.statusCode, duration, getClientIP(r))
	})
}

// responseWriter 包装http.ResponseWriter以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// GetUserFromContext 从上下文中获取用户信息
func GetUserFromContext(ctx context.Context) (*CustomClaims, bool) {
	user, ok := ctx.Value(UserContextKey).(*CustomClaims)
	return user, ok
}

// GetUserIDFromContext 从上下文中获取用户ID
func GetUserIDFromContext(ctx context.Context) (uint, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(uint)
	return userID, ok
}

// RequireOwnership 检查用户是否拥有资源的权限（用于资源访问控制）
func RequireOwnership(userID uint, resourceUserID uint) bool {
	return userID == resourceUserID
}

// getClientIP 获取客户端真实IP
func getClientIP(r *http.Request) string {
	// 优先从X-Forwarded-For获取
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// 从X-Real-IP获取
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	
	// 从RemoteAddr获取
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	
	return ip
}

// writeErrorResponse 写入错误响应
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
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

// ParseIDFromURL 从URL路径中解析ID参数（辅助函数）
func ParseIDFromURL(r *http.Request, param string) (int, error) {
	// 这里假设使用gorilla/mux，需要根据实际路由器调整
	// 这是一个辅助函数，在handler中使用
	vars := make(map[string]string) // 占位符，实际实现需要从路由器获取
	idStr, exists := vars[param]
	if !exists {
		return 0, errors.New("missing " + param + " parameter")
	}
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, errors.New("invalid " + param + " parameter")
	}
	
	return id, nil
}