package main

import (
	"encoding/json"
	"fmt"
	"jobView-backend/internal/auth"
	"jobView-backend/internal/config"
	"jobView-backend/internal/database"
	"jobView-backend/internal/handler"
	"jobView-backend/internal/service"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// 加载配置
	cfg := config.Load()
	
	// 验证配置
	if err := cfg.ValidateConfig(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// 连接数据库
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Successfully connected to database")

	// 运行数据库迁移
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// 初始化服务
	jobService := service.NewJobApplicationService(db)
	authService := service.NewAuthService(db)
	statusTrackingService := service.NewStatusTrackingService(db)
	statusConfigService := service.NewStatusConfigService(db)
	exportService := service.NewExportService(db, jobService)

    // 在创建处理器之前，确保默认模板包含直通规则（幂等补齐）
    if err := statusConfigService.EnsureDirectTransitionsInDefaultTemplate(); err != nil {
        log.Printf("Warning: ensure default flow transitions failed: %v", err)
    }

    // 初始化处理器
	jobHandler := handler.NewJobApplicationHandler(jobService)
	authHandler := handler.NewAuthHandler(authService)
	statusTrackingHandler := handler.NewStatusTrackingHandler(statusTrackingService)
	statusConfigHandler := handler.NewStatusConfigHandler(statusConfigService)
	exportHandler := handler.NewExportHandler(exportService)

	// 设置路由
	router := mux.NewRouter()
	
	// 应用全局中间件
	router.Use(auth.LoggingMiddleware)
	router.Use(auth.SecurityHeadersMiddleware)
	router.Use(auth.CORSMiddleware([]string{"http://localhost:3000", "http://localhost:8010"}))
	
	// 认证相关路由（无需认证）
	authRouter := router.PathPrefix("/api/auth").Subrouter()
	authRouter.Use(auth.RateLimitMiddleware(10, time.Minute)) // 认证接口限流
	
	authRouter.HandleFunc("/register", authHandler.Register).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/login", authHandler.Login).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/refresh", authHandler.RefreshToken).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/health", authHandler.HealthCheck).Methods("GET", "OPTIONS")
	
	// 新增：用户名和邮箱可用性检查
	authRouter.HandleFunc("/check-username", authHandler.CheckUsernameAvailability).Methods("GET", "OPTIONS")
	authRouter.HandleFunc("/check-email", authHandler.CheckEmailAvailability).Methods("GET", "OPTIONS")
	
	// 需要认证的认证相关路由
	protectedAuthRouter := authRouter.PathPrefix("").Subrouter()
	protectedAuthRouter.Use(auth.AuthMiddleware)
	
	protectedAuthRouter.HandleFunc("/profile", authHandler.GetProfile).Methods("GET", "OPTIONS")
	protectedAuthRouter.HandleFunc("/profile", authHandler.UpdateProfile).Methods("PUT", "OPTIONS")
	protectedAuthRouter.HandleFunc("/password", authHandler.ChangePassword).Methods("PUT", "OPTIONS")
	protectedAuthRouter.HandleFunc("/avatar", authHandler.UploadAvatar).Methods("POST", "OPTIONS")
	protectedAuthRouter.HandleFunc("/logout", authHandler.Logout).Methods("POST", "OPTIONS")
	protectedAuthRouter.HandleFunc("/validate", authHandler.ValidateToken).Methods("GET", "OPTIONS")
	protectedAuthRouter.HandleFunc("/stats", authHandler.GetUserStats).Methods("GET", "OPTIONS")

	// API v1 路由（需要认证）
	api := router.PathPrefix("/api/v1").Subrouter()
	
	// 为所有API路径添加OPTIONS处理（不需要认证）
	router.PathPrefix("/api/v1/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// OPTIONS请求已经在CORS中间件中处理，这里只是确保路由匹配
			return
		}
		// 对于非OPTIONS请求，转发给需要认证的处理器
		api.ServeHTTP(w, r)
	}).Methods("OPTIONS")
	
	api.Use(auth.AuthMiddleware) // 所有v1 API都需要认证
	api.Use(auth.RateLimitMiddleware(60, time.Minute)) // API限流

	// 投递记录相关路由
	api.HandleFunc("/applications", jobHandler.Create).Methods("POST")
	api.HandleFunc("/applications", jobHandler.GetJobApplicationsWithFilters).Methods("GET") // 更新为筛选版本
	api.HandleFunc("/applications/statistics", jobHandler.GetStatistics).Methods("GET")
	api.HandleFunc("/applications/search", jobHandler.SearchJobApplications).Methods("GET")
	api.HandleFunc("/applications/dashboard", jobHandler.GetDashboardData).Methods("GET")
	api.HandleFunc("/applications/{id}", jobHandler.GetByID).Methods("GET")
	api.HandleFunc("/applications/{id}", jobHandler.Update).Methods("PUT")
	api.HandleFunc("/applications/{id}", jobHandler.Delete).Methods("DELETE")

	// 状态跟踪相关路由
	api.HandleFunc("/job-applications/{id}/status-history", statusTrackingHandler.GetStatusHistory).Methods("GET")
	api.HandleFunc("/job-applications/{id}/status", statusTrackingHandler.UpdateJobStatus).Methods("POST")
	api.HandleFunc("/job-applications/{id}/status-timeline", statusTrackingHandler.GetStatusTimeline).Methods("GET")
	api.HandleFunc("/job-applications/status/batch", statusTrackingHandler.BatchUpdateStatus).Methods("PUT")
	api.HandleFunc("/job-applications/status-analytics", statusTrackingHandler.GetStatusAnalytics).Methods("GET")
	api.HandleFunc("/job-applications/status-trends", statusTrackingHandler.GetStatusTrends).Methods("GET")
	api.HandleFunc("/job-applications/process-insights", statusTrackingHandler.GetProcessInsights).Methods("GET")

	// 状态配置管理路由
	api.HandleFunc("/status-flow-templates", statusConfigHandler.GetStatusFlowTemplates).Methods("GET")
	api.HandleFunc("/status-flow-templates", statusConfigHandler.CreateStatusFlowTemplate).Methods("POST")
	api.HandleFunc("/status-flow-templates/{id}", statusConfigHandler.UpdateStatusFlowTemplate).Methods("PUT")
	api.HandleFunc("/status-flow-templates/{id}", statusConfigHandler.DeleteStatusFlowTemplate).Methods("DELETE")
	api.HandleFunc("/user-status-preferences", statusConfigHandler.GetUserStatusPreferences).Methods("GET")
	api.HandleFunc("/user-status-preferences", statusConfigHandler.UpdateUserStatusPreferences).Methods("PUT")
	api.HandleFunc("/status-transitions/{status}", statusConfigHandler.GetAvailableStatusTransitions).Methods("GET")
	api.HandleFunc("/status-definitions", statusConfigHandler.GetAllStatusDefinitions).Methods("GET")

	// Excel导出相关路由
	api.HandleFunc("/export/applications", exportHandler.StartExport).Methods("POST")
	api.HandleFunc("/export/status/{task_id}", exportHandler.GetTaskStatus).Methods("GET")
	api.HandleFunc("/export/download/{task_id}", exportHandler.DownloadFile).Methods("GET")
	api.HandleFunc("/export/history", exportHandler.GetExportHistory).Methods("GET")
	api.HandleFunc("/export/cancel/{task_id}", exportHandler.CancelExport).Methods("DELETE")
	api.HandleFunc("/export/formats", exportHandler.GetSupportedFormats).Methods("GET")
	api.HandleFunc("/export/fields", exportHandler.GetExportFields).Methods("GET")
	api.HandleFunc("/export/template", exportHandler.GetExportTemplate).Methods("GET")

	// 健康检查路由（无需认证）

	// 静态文件服务：/static/* -> ./uploads
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./uploads"))))
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"status":      "ok",
			"service":     "jobview-backend",
			"version":     "1.0.0",
			"timestamp":   time.Now().Unix(),
			"environment": cfg.Server.Environment,
		}
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")
	
	// 根路径重定向到健康检查
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/health", http.StatusSeeOther)
	}).Methods("GET")

	// 启动服务器
	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	
	// 打印启动信息
	log.Printf("=== JobView Backend Server Starting ===") 
	log.Printf("Environment: %s", cfg.Server.Environment)
	log.Printf("Server starting on port %s", cfg.Server.Port)
	log.Printf("Health check: http://localhost%s/health", serverAddr)
	log.Printf("Auth endpoints: http://localhost%s/api/auth/*", serverAddr)
	log.Printf("Job Applications: http://localhost%s/api/v1/applications/*", serverAddr)
	log.Printf("Status Tracking: http://localhost%s/api/v1/job-applications/*/status*", serverAddr)
	log.Printf("Status Config: http://localhost%s/api/v1/status-*", serverAddr)
	log.Printf("Excel Export: http://localhost%s/api/v1/export/*", serverAddr)
	log.Printf("=== Ready for connections ===")
	
	// 生产环境启用更多安全特性
	if cfg.IsProduction() {
		log.Println("Production mode: Enhanced security enabled")
	} else {
		log.Println("Development mode: Debug logging enabled")
		// 在开发模式下显示示例用户信息
		log.Println("Default test user: username=testuser, password=TestPass123!")
	}

	// 启动HTTP服务器
	server := &http.Server{
		Addr:           serverAddr,
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

    if err := server.ListenAndServe(); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}
