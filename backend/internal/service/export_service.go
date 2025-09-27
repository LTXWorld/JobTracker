/*
位置: backend/internal/service/export_service.go
概述: 导出服务，负责处理 Excel 导出的业务逻辑，包括同步和异步导出功能
功能: 数据查询、导出任务管理、文件生成和存储
与其他文件关系: 依赖 JobApplicationService 获取数据，使用 excel.Generator 生成文件，与 export_handler.go 协作
*/

package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"jobView-backend/internal/excel"
	"jobView-backend/internal/model"
	"jobView-backend/internal/repository"
	"os"
	"path/filepath"
	"time"
)

// ExportService 导出服务
type ExportService struct {
	repo                   repository.ExportRepository
	maxRecordsForSync      int    // 同步导出的最大记录数
	tempDir                string // 临时文件目录
	fileRetentionHours     int    // 文件保留时间（小时）
	maxConcurrentExports   int    // 最大并发导出数
	maxDailyExportsPerUser int    // 每用户每日最大导出次数
}

// NewExportService 创建新的导出服务
func NewExportService(repo repository.ExportRepository) *ExportService {
	// 确保临时目录存在
	tempDir := os.TempDir() + "/jobview_exports"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		// 如果创建失败，使用系统临时目录
		tempDir = os.TempDir()
	}

	return &ExportService{
		repo:                   repo,
		maxRecordsForSync:      1000, // 超过1000条记录使用异步导出
		tempDir:                tempDir,
		fileRetentionHours:     24, // 文件保留24小时
		maxConcurrentExports:   5,  // 最大5个并发导出任务
		maxDailyExportsPerUser: 20, // 每用户每日最多20次导出
	}
}

// StartExport 开始导出任务
func (s *ExportService) StartExport(userID uint, request *model.ExportRequest) (*model.ExportResponse, error) {
	// 验证导出请求
	if err := request.ValidateExportRequest(); err != nil {
		return nil, fmt.Errorf("导出请求验证失败: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 检查用户导出限制
	if err := s.checkUserExportLimits(ctx, userID); err != nil {
		return nil, err
	}

	// 查询要导出的数据总数
	totalCount, err := s.repo.CountExportData(ctx, userID, &request.Filters)
	if err != nil {
		return nil, fmt.Errorf("查询数据总数失败: %v", err)
	}

	if totalCount == 0 {
		return nil, fmt.Errorf("没有符合条件的数据可导出")
	}

	// 生成任务ID
	taskID := model.GenerateTaskID(userID)

	// 创建导出任务记录
	task := &model.ExportTask{
		TaskID:           taskID,
		UserID:           userID,
		Status:           model.TaskStatusPending,
		ExportType:       request.Format,
		TotalRecords:     &totalCount,
		ProcessedRecords: 0,
		Progress:         0,
		Filters:          &request.Filters,
		Options:          &request.Options,
		CreatedAt:        time.Now(),
	}

	// 设置过期时间
	expiresAt := time.Now().Add(time.Duration(s.fileRetentionHours) * time.Hour)
	task.ExpiresAt = &expiresAt

	// 保存任务到数据库
	if err := s.createTask(task); err != nil {
		return nil, fmt.Errorf("保存导出任务失败: %v", err)
	}

	// 根据数据量决定使用同步还是异步处理
	if totalCount <= s.maxRecordsForSync {
		// 同步处理小数据量
		return s.processSyncExport(task, request)
	} else {
		// 异步处理大数据量
		go s.processAsyncExport(task, request)

		// 返回任务状态
		estimatedTime := s.estimateProcessingTime(totalCount)
		return &model.ExportResponse{
			TaskID:        taskID,
			Status:        model.TaskStatusProcessing,
			Progress:      0,
			TotalRecords:  &totalCount,
			EstimatedTime: &estimatedTime,
			Message:       "导出任务已启动，正在后台处理",
		}, nil
	}
}

// processSyncExport 同步处理导出
func (s *ExportService) processSyncExport(task *model.ExportTask, request *model.ExportRequest) (*model.ExportResponse, error) {
	// 更新任务状态为处理中
	task.Status = model.TaskStatusProcessing
	startTime := time.Now()
	task.StartedAt = &startTime
	_ = s.updateTask(task)

	// 获取数据
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	applications, err := s.repo.FetchExportData(ctx, task.UserID, &request.Filters, 0, *task.TotalRecords)
	if err != nil {
		task.Status = model.TaskStatusFailed
		errorMsg := fmt.Sprintf("获取导出数据失败: %v", err)
		task.ErrorMessage = &errorMsg
		_ = s.updateTask(task)
		return nil, errors.New(errorMsg)
	}

	// 生成文件
	filePath, fileSize, err := s.generateExcelFile(task.TaskID, applications, &request.Options)
	if err != nil {
		task.Status = model.TaskStatusFailed
		errorMsg := fmt.Sprintf("生成Excel文件失败: %v", err)
		task.ErrorMessage = &errorMsg
		_ = s.updateTask(task)
		return nil, errors.New(errorMsg)
	}

	// 更新任务状态为完成
	task.Status = model.TaskStatusCompleted
	task.FilePath = &filePath
	task.FileSize = &fileSize
	task.ProcessedRecords = *task.TotalRecords
	task.Progress = 100
	completedTime := time.Now()
	task.CompletedAt = &completedTime

	// 生成文件名
	filename := s.generateFilename(task.UserID, &request.Options)
	task.Filename = &filename

	_ = s.updateTask(task)

	// 生成下载URL
	downloadURL := fmt.Sprintf("/api/v1/export/download/%s", task.TaskID)

	return &model.ExportResponse{
		TaskID:      task.TaskID,
		Status:      model.TaskStatusCompleted,
		Progress:    100,
		DownloadURL: &downloadURL,
		FileSize:    func() *string { s := task.GetFormattedFileSize(); return &s }(),
		Message:     "导出完成",
	}, nil
}

// processAsyncExport 异步处理导出
func (s *ExportService) processAsyncExport(task *model.ExportTask, request *model.ExportRequest) {
	// 更新任务状态为处理中
	task.Status = model.TaskStatusProcessing
	startTime := time.Now()
	task.StartedAt = &startTime
	_ = s.updateTask(task)

	// 分批处理数据
	batchSize := 1000
	totalRecords := *task.TotalRecords

	// 创建临时文件
	generator := excel.NewGenerator()
	defer generator.Close()

	if err := generator.InitializeWorkbook(); err != nil {
		s.handleExportError(task, fmt.Sprintf("初始化Excel工作簿失败: %v", err))
		return
	}

	var allApplications []model.JobApplication

	// 分批获取数据并处理
	for offset := 0; offset < totalRecords; offset += batchSize {
		limit := batchSize
		if offset+batchSize > totalRecords {
			limit = totalRecords - offset
		}

		// 获取批次数据
		ctx, cancelBatch := context.WithTimeout(context.Background(), 10*time.Second)
		applications, err := s.repo.FetchExportData(ctx, task.UserID, &request.Filters, offset, limit)
		cancelBatch()
		if err != nil {
			s.handleExportError(task, fmt.Sprintf("获取第%d批数据失败: %v", offset/batchSize+1, err))
			return
		}

		allApplications = append(allApplications, applications...)

		// 更新进度
		processed := offset + len(applications)
		task.ProcessedRecords = processed
		task.Progress = (processed * 100) / totalRecords
		_ = s.updateTask(task)
	}

	// 写入所有数据到Excel
	if err := generator.WriteJobApplications(allApplications); err != nil {
		s.handleExportError(task, fmt.Sprintf("写入Excel数据失败: %v", err))
		return
	}

	// 如果需要包含统计信息
	if request.Options.IncludeStatistics {
		stats := s.generateStatistics(allApplications)
		if err := generator.AddStatisticsSheet(stats); err != nil {
			// 统计信息生成失败不影响主要导出
			fmt.Printf("生成统计信息失败: %v\n", err)
		}
	}

	// 保存文件
	filePath := filepath.Join(s.tempDir, fmt.Sprintf("%s.xlsx", task.TaskID))
	if err := generator.SaveToFile(filePath); err != nil {
		s.handleExportError(task, fmt.Sprintf("保存Excel文件失败: %v", err))
		return
	}

	// 获取文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		s.handleExportError(task, fmt.Sprintf("获取文件信息失败: %v", err))
		return
	}
	fileSize := fileInfo.Size()

	// 更新任务状态为完成
	task.Status = model.TaskStatusCompleted
	task.FilePath = &filePath
	task.FileSize = &fileSize
	task.ProcessedRecords = totalRecords
	task.Progress = 100
	completedTime := time.Now()
	task.CompletedAt = &completedTime

	// 生成文件名
	filename := s.generateFilename(task.UserID, &request.Options)
	task.Filename = &filename

	_ = s.updateTask(task)
}

// GetTaskStatus 获取任务状态
func (s *ExportService) GetTaskStatus(taskID string, userID uint) (*model.TaskStatusResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	record, err := s.repo.GetTaskStatus(ctx, taskID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("导出任务不存在或无访问权限")
		}
		return nil, fmt.Errorf("查询任务状态失败: %v", err)
	}

	resp := &model.TaskStatusResponse{
		TaskID:           record.TaskID,
		Status:           record.Status,
		Progress:         record.Progress,
		ProcessedRecords: record.ProcessedRecords,
		CreatedAt:        record.CreatedAt,
		CompletedAt:      record.CompletedAt,
		TotalRecords:     record.TotalRecords,
		ExpiresAt:        record.ExpiresAt,
		ErrorMessage:     record.ErrorMessage,
	}

	if record.FileSize != nil {
		size := formatFileSize(*record.FileSize)
		resp.FileSize = &size
	}
	if resp.Status == model.TaskStatusCompleted {
		downloadURL := fmt.Sprintf("/api/v1/export/download/%s", taskID)
		resp.DownloadURL = &downloadURL
	}

	return resp, nil
}

// DownloadFile 获取下载文件
func (s *ExportService) DownloadFile(taskID string, userID uint) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	record, err := s.repo.GetDownloadInfo(ctx, taskID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", fmt.Errorf("文件不存在或无访问权限")
		}
		return "", "", fmt.Errorf("查询文件信息失败: %v", err)
	}

	if record.Status != model.TaskStatusCompleted {
		return "", "", fmt.Errorf("文件尚未生成完成")
	}
	if record.Expires != nil && time.Now().After(*record.Expires) {
		return "", "", fmt.Errorf("文件已过期")
	}
	if _, err := os.Stat(record.FilePath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("文件不存在")
	}

	return record.FilePath, record.Filename, nil
}

// GetExportHistory 获取导出历史
func (s *ExportService) GetExportHistory(userID uint, page, limit int) (*model.ExportHistoryResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	totalCount, err := s.repo.CountHistory(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("查询总数失败: %v", err)
	}

	rows, err := s.repo.ListHistory(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询导出历史失败: %v", err)
	}

	exports := make([]model.ExportHistoryItem, 0, len(rows))
	for _, rec := range rows {
		item := model.ExportHistoryItem{
			TaskID:    rec.TaskID,
			CreatedAt: rec.CreatedAt,
			Status:    rec.Status,
			ExpiresAt: rec.ExpiresAt,
		}
		if rec.Filename != nil {
			item.Filename = rec.Filename
		}
		if rec.FileSize != nil {
			size := formatFileSize(*rec.FileSize)
			item.FileSize = &size
		}
		if rec.TotalRecords != nil {
			item.RecordCount = rec.TotalRecords
		}
		if item.Status == model.TaskStatusCompleted && (item.ExpiresAt == nil || time.Now().Before(*item.ExpiresAt)) {
			downloadURL := fmt.Sprintf("/api/v1/export/download/%s", item.TaskID)
			item.DownloadURL = &downloadURL
		}
		exports = append(exports, item)
	}

	totalPages := int((totalCount + int64(limit) - 1) / int64(limit))
	pagination := model.PaginationResponse{
		Data:       exports,
		Total:      totalCount,
		Page:       page,
		PageSize:   limit,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	return &model.ExportHistoryResponse{
		Exports:    exports,
		Pagination: pagination,
	}, nil
}

// 内部辅助方法

// checkUserExportLimits 检查用户导出限制
func (s *ExportService) checkUserExportLimits(ctx context.Context, userID uint) error {
	dailyCount, err := s.repo.CountDailyExports(ctx, userID, time.Now())
	if err != nil {
		return fmt.Errorf("检查日导出次数失败: %v", err)
	}
	if dailyCount >= s.maxDailyExportsPerUser {
		return fmt.Errorf("今日导出次数已达上限 (%d次)", s.maxDailyExportsPerUser)
	}

	activeCount, err := s.repo.CountActiveExports(ctx, userID)
	if err != nil {
		return fmt.Errorf("检查并发导出数失败: %v", err)
	}
	if activeCount >= s.maxConcurrentExports {
		return fmt.Errorf("当前有太多导出任务在进行中，请稍后再试")
	}
	return nil
}

// generateExcelFile 生成Excel文件
func (s *ExportService) generateExcelFile(taskID string, applications []model.JobApplication, options *model.ExportOptions) (string, int64, error) {
	generator := excel.NewGenerator()
	defer generator.Close()

	if err := generator.InitializeWorkbook(); err != nil {
		return "", 0, fmt.Errorf("初始化工作簿失败: %v", err)
	}

	if err := generator.WriteJobApplications(applications); err != nil {
		return "", 0, fmt.Errorf("写入数据失败: %v", err)
	}

	// 如果需要统计信息
	if options.IncludeStatistics {
		stats := s.generateStatistics(applications)
		if err := generator.AddStatisticsSheet(stats); err != nil {
			// 统计信息失败不影响主要功能
			fmt.Printf("添加统计工作表失败: %v\n", err)
		}
	}

	// 保存文件
	filePath := filepath.Join(s.tempDir, fmt.Sprintf("%s.xlsx", taskID))
	if err := generator.SaveToFile(filePath); err != nil {
		return "", 0, fmt.Errorf("保存文件失败: %v", err)
	}

	// 获取文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("获取文件信息失败: %v", err)
	}

	return filePath, fileInfo.Size(), nil
}

// generateStatistics 生成统计信息
func (s *ExportService) generateStatistics(applications []model.JobApplication) map[string]interface{} {
	stats := make(map[string]interface{})
	statusDistribution := make(map[string]int)

	for _, app := range applications {
		statusDistribution[string(app.Status)]++
	}

	stats["statusDistribution"] = statusDistribution
	stats["totalCount"] = len(applications)

	return stats
}

// generateFilename 生成文件名
func (s *ExportService) generateFilename(userID uint, options *model.ExportOptions) string {
	if options.Filename != "" {
		return options.Filename + ".xlsx"
	}

	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("求职投递记录_用户%d_%s.xlsx", userID, timestamp)
}

// estimateProcessingTime 估算处理时间（秒）
func (s *ExportService) estimateProcessingTime(recordCount int) int {
	// 大约每1000条记录需要5秒
	return (recordCount / 1000) * 5
}

// handleExportError 处理导出错误
func (s *ExportService) handleExportError(task *model.ExportTask, errorMsg string) {
	task.Status = model.TaskStatusFailed
	task.ErrorMessage = &errorMsg
	_ = s.updateTask(task)
}

// formatFileSize 格式化文件大小
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// CleanupExpiredTasks 清理过期任务（可以通过定时任务调用）
func (s *ExportService) CleanupExpiredTasks() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tasks, err := s.repo.ListExpiredTasks(ctx)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if task.FilePath != nil {
			_ = os.Remove(*task.FilePath)
		}
		deleteCtx, cancelDelete := context.WithTimeout(context.Background(), 3*time.Second)
		_ = s.repo.DeleteTask(deleteCtx, task.TaskID)
		cancelDelete()
	}
	return nil
}

func (s *ExportService) createTask(task *model.ExportTask) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.repo.CreateTask(ctx, task)
}

func (s *ExportService) updateTask(task *model.ExportTask) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.repo.UpdateTask(ctx, task)
}
