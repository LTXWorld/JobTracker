/*
位置: backend/internal/service/export_service.go
概述: 导出服务，负责处理 Excel 导出的业务逻辑，包括同步和异步导出功能
功能: 数据查询、导出任务管理、文件生成和存储
与其他文件关系: 依赖 JobApplicationService 获取数据，使用 excel.Generator 生成文件，与 export_handler.go 协作
*/

package service

import (
	"database/sql"
	"fmt"
	"jobView-backend/internal/database"
	"jobView-backend/internal/excel"
	"jobView-backend/internal/model"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ExportService 导出服务
type ExportService struct {
	db                      *database.DB
	jobApplicationService   *JobApplicationService
	maxRecordsForSync       int    // 同步导出的最大记录数
	tempDir                 string // 临时文件目录
	fileRetentionHours      int    // 文件保留时间（小时）
	maxConcurrentExports    int    // 最大并发导出数
	maxDailyExportsPerUser  int    // 每用户每日最大导出次数
}

// NewExportService 创建新的导出服务
func NewExportService(db *database.DB, jobApplicationService *JobApplicationService) *ExportService {
	// 确保临时目录存在
	tempDir := os.TempDir() + "/jobview_exports"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		// 如果创建失败，使用系统临时目录
		tempDir = os.TempDir()
	}

	return &ExportService{
		db:                      db,
		jobApplicationService:   jobApplicationService,
		maxRecordsForSync:       1000,  // 超过1000条记录使用异步导出
		tempDir:                 tempDir,
		fileRetentionHours:      24,    // 文件保留24小时
		maxConcurrentExports:    5,     // 最大5个并发导出任务
		maxDailyExportsPerUser:  20,    // 每用户每日最多20次导出
	}
}

// StartExport 开始导出任务
func (s *ExportService) StartExport(userID uint, request *model.ExportRequest) (*model.ExportResponse, error) {
	// 验证导出请求
	if err := request.ValidateExportRequest(); err != nil {
		return nil, fmt.Errorf("导出请求验证失败: %v", err)
	}

	// 检查用户导出限制
	if err := s.checkUserExportLimits(userID); err != nil {
		return nil, err
	}

	// 查询要导出的数据总数
	totalCount, err := s.getExportDataCount(userID, &request.Filters)
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
	if err := s.saveExportTask(task); err != nil {
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
	s.updateExportTask(task)

	// 获取数据
	applications, err := s.getExportData(task.UserID, &request.Filters, 0, *task.TotalRecords)
	if err != nil {
		task.Status = model.TaskStatusFailed
		errorMsg := fmt.Sprintf("获取导出数据失败: %v", err)
		task.ErrorMessage = &errorMsg
		s.updateExportTask(task)
		return nil, fmt.Errorf(errorMsg)
	}

	// 生成文件
	filePath, fileSize, err := s.generateExcelFile(task.TaskID, applications, &request.Options)
	if err != nil {
		task.Status = model.TaskStatusFailed
		errorMsg := fmt.Sprintf("生成Excel文件失败: %v", err)
		task.ErrorMessage = &errorMsg
		s.updateExportTask(task)
		return nil, fmt.Errorf(errorMsg)
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

	s.updateExportTask(task)

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
	s.updateExportTask(task)

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
		applications, err := s.getExportData(task.UserID, &request.Filters, offset, limit)
		if err != nil {
			s.handleExportError(task, fmt.Sprintf("获取第%d批数据失败: %v", offset/batchSize+1, err))
			return
		}

		allApplications = append(allApplications, applications...)

		// 更新进度
		processed := offset + len(applications)
		task.ProcessedRecords = processed
		task.Progress = (processed * 100) / totalRecords
		s.updateExportTask(task)
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

	s.updateExportTask(task)
}

// GetTaskStatus 获取任务状态
func (s *ExportService) GetTaskStatus(taskID string, userID uint) (*model.TaskStatusResponse, error) {
	query := `
		SELECT task_id, status, progress, processed_records, total_records,
			   file_size, expires_at, error_message, created_at, completed_at, filename
		FROM export_tasks 
		WHERE task_id = $1 AND user_id = $2
	`

	var task model.TaskStatusResponse
	var fileSize sql.NullInt64
	var expiresAt, completedAt sql.NullTime
	var errorMessage, filename sql.NullString
	var totalRecords sql.NullInt32

	err := s.db.QueryRow(query, taskID, userID).Scan(
		&task.TaskID,
		&task.Status,
		&task.Progress,
		&task.ProcessedRecords,
		&totalRecords,
		&fileSize,
		&expiresAt,
		&errorMessage,
		&task.CreatedAt,
		&completedAt,
		&filename,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("导出任务不存在或无访问权限")
		}
		return nil, fmt.Errorf("查询任务状态失败: %v", err)
	}

	// 填充可选字段
	if totalRecords.Valid {
		totalRec := int(totalRecords.Int32)
		task.TotalRecords = &totalRec
	}

	if fileSize.Valid {
		size := fileSize.Int64
		formattedSize := formatFileSize(size)
		task.FileSize = &formattedSize
	}

	if expiresAt.Valid {
		task.ExpiresAt = &expiresAt.Time
	}

	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	if errorMessage.Valid {
		task.ErrorMessage = &errorMessage.String
	}

	// 如果任务已完成，生成下载链接
	if task.Status == model.TaskStatusCompleted {
		downloadURL := fmt.Sprintf("/api/v1/export/download/%s", taskID)
		task.DownloadURL = &downloadURL
	}

	return &task, nil
}

// DownloadFile 获取下载文件
func (s *ExportService) DownloadFile(taskID string, userID uint) (string, string, error) {
	query := `
		SELECT file_path, filename, status, expires_at 
		FROM export_tasks 
		WHERE task_id = $1 AND user_id = $2
	`

	var filePath, filename sql.NullString
	var status model.TaskStatus
	var expiresAt sql.NullTime

	err := s.db.QueryRow(query, taskID, userID).Scan(&filePath, &filename, &status, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("文件不存在或无访问权限")
		}
		return "", "", fmt.Errorf("查询文件信息失败: %v", err)
	}

	// 检查任务状态
	if status != model.TaskStatusCompleted {
		return "", "", fmt.Errorf("文件尚未生成完成")
	}

	// 检查文件是否过期
	if expiresAt.Valid && time.Now().After(expiresAt.Time) {
		return "", "", fmt.Errorf("文件已过期")
	}

	if !filePath.Valid || !filename.Valid {
		return "", "", fmt.Errorf("文件路径或文件名无效")
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath.String); os.IsNotExist(err) {
		return "", "", fmt.Errorf("文件不存在")
	}

	return filePath.String, filename.String, nil
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

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM export_tasks WHERE user_id = $1`
	var totalCount int64
	err := s.db.QueryRow(countQuery, userID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("查询总数失败: %v", err)
	}

	// 查询历史记录
	query := `
		SELECT task_id, created_at, status, filename, file_size, total_records, expires_at
		FROM export_tasks 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询导出历史失败: %v", err)
	}
	defer rows.Close()

	var exports []model.ExportHistoryItem
	for rows.Next() {
		var item model.ExportHistoryItem
		var filename sql.NullString
		var fileSize sql.NullInt64
		var totalRecords sql.NullInt32
		var expiresAt sql.NullTime

		err := rows.Scan(
			&item.TaskID,
			&item.CreatedAt,
			&item.Status,
			&filename,
			&fileSize,
			&totalRecords,
			&expiresAt,
		)
		if err != nil {
			continue
		}

		// 填充可选字段
		if filename.Valid {
			item.Filename = &filename.String
		}

		if fileSize.Valid {
			size := formatFileSize(fileSize.Int64)
			item.FileSize = &size
		}

		if totalRecords.Valid {
			count := int(totalRecords.Int32)
			item.RecordCount = &count
		}

		if expiresAt.Valid {
			item.ExpiresAt = &expiresAt.Time
		}

		// 如果任务完成且未过期，生成下载链接
		if item.Status == model.TaskStatusCompleted && 
		   (item.ExpiresAt == nil || time.Now().Before(*item.ExpiresAt)) {
			downloadURL := fmt.Sprintf("/api/v1/export/download/%s", item.TaskID)
			item.DownloadURL = &downloadURL
		}

		exports = append(exports, item)
	}

	// 计算分页信息
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
func (s *ExportService) checkUserExportLimits(userID uint) error {
	// 检查今日导出次数
	today := time.Now().Format("2006-01-02")
	query := `
		SELECT COUNT(*) FROM export_tasks 
		WHERE user_id = $1 AND DATE(created_at) = $2
	`
	
	var dailyCount int
	err := s.db.QueryRow(query, userID, today).Scan(&dailyCount)
	if err != nil {
		return fmt.Errorf("检查日导出次数失败: %v", err)
	}

	if dailyCount >= s.maxDailyExportsPerUser {
		return fmt.Errorf("今日导出次数已达上限 (%d次)", s.maxDailyExportsPerUser)
	}

	// 检查当前并发导出数
	query = `
		SELECT COUNT(*) FROM export_tasks 
		WHERE user_id = $1 AND status = $2
	`
	
	var activeCount int
	err = s.db.QueryRow(query, userID, model.TaskStatusProcessing).Scan(&activeCount)
	if err != nil {
		return fmt.Errorf("检查并发导出数失败: %v", err)
	}

	if activeCount >= s.maxConcurrentExports {
		return fmt.Errorf("当前有太多导出任务在进行中，请稍后再试")
	}

	return nil
}

// getExportDataCount 获取导出数据总数
func (s *ExportService) getExportDataCount(userID uint, filters *model.ExportFilters) (int, error) {
	query, args := s.buildCountQuery(userID, filters)
	
	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询数据总数失败: %v", err)
	}
	
	return count, nil
}

// getExportData 获取导出数据
func (s *ExportService) getExportData(userID uint, filters *model.ExportFilters, offset, limit int) ([]model.JobApplication, error) {
	query, args := s.buildDataQuery(userID, filters, offset, limit)
	
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询导出数据失败: %v", err)
	}
	defer rows.Close()

	var applications []model.JobApplication
	for rows.Next() {
		var app model.JobApplication
		err := rows.Scan(
			&app.ID,
			&app.UserID,
			&app.CompanyName,
			&app.PositionTitle,
			&app.ApplicationDate,
			&app.Status,
			&app.JobDescription,
			&app.SalaryRange,
			&app.WorkLocation,
			&app.ContactInfo,
			&app.Notes,
			&app.InterviewTime,
			&app.ReminderTime,
			&app.ReminderEnabled,
			&app.FollowUpDate,
			&app.HRName,
			&app.HRPhone,
			&app.HREmail,
			&app.InterviewLocation,
			&app.InterviewType,
			&app.CreatedAt,
			&app.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描数据失败: %v", err)
		}
		applications = append(applications, app)
	}

	return applications, nil
}

// buildCountQuery 构建计数查询
func (s *ExportService) buildCountQuery(userID uint, filters *model.ExportFilters) (string, []interface{}) {
	query := "SELECT COUNT(*) FROM job_applications WHERE user_id = $1"
	args := []interface{}{userID}
	argIndex := 2

	// 添加筛选条件
	if len(filters.Status) > 0 {
		statusPlaceholders := make([]string, len(filters.Status))
		for i, status := range filters.Status {
			statusPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, status)
			argIndex++
		}
		query += " AND status IN (" + strings.Join(statusPlaceholders, ",") + ")"
	}

	if filters.DateRange != nil {
		query += fmt.Sprintf(" AND application_date >= $%d AND application_date <= $%d", argIndex, argIndex+1)
		args = append(args, filters.DateRange.Start, filters.DateRange.End)
		argIndex += 2
	}

	if len(filters.CompanyNames) > 0 {
		companyPlaceholders := make([]string, len(filters.CompanyNames))
		for i, company := range filters.CompanyNames {
			companyPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, company)
			argIndex++
		}
		query += " AND company_name IN (" + strings.Join(companyPlaceholders, ",") + ")"
	}

	if filters.Keywords != "" {
		query += fmt.Sprintf(" AND (company_name ILIKE $%d OR position_title ILIKE $%d OR notes ILIKE $%d)", 
							 argIndex, argIndex, argIndex)
		keyword := "%" + filters.Keywords + "%"
		args = append(args, keyword)
	}

	return query, args
}

// buildDataQuery 构建数据查询
func (s *ExportService) buildDataQuery(userID uint, filters *model.ExportFilters, offset, limit int) (string, []interface{}) {
	query := `
		SELECT id, user_id, company_name, position_title, application_date, status,
			   job_description, salary_range, work_location, contact_info, notes,
			   interview_time, reminder_time, reminder_enabled, follow_up_date,
			   hr_name, hr_phone, hr_email, interview_location, interview_type,
			   created_at, updated_at
		FROM job_applications 
		WHERE user_id = $1
	`
	args := []interface{}{userID}
	argIndex := 2

	// 添加筛选条件（与 buildCountQuery 相同的逻辑）
	if len(filters.Status) > 0 {
		statusPlaceholders := make([]string, len(filters.Status))
		for i, status := range filters.Status {
			statusPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, status)
			argIndex++
		}
		query += " AND status IN (" + strings.Join(statusPlaceholders, ",") + ")"
	}

	if filters.DateRange != nil {
		query += fmt.Sprintf(" AND application_date >= $%d AND application_date <= $%d", argIndex, argIndex+1)
		args = append(args, filters.DateRange.Start, filters.DateRange.End)
		argIndex += 2
	}

	if len(filters.CompanyNames) > 0 {
		companyPlaceholders := make([]string, len(filters.CompanyNames))
		for i, company := range filters.CompanyNames {
			companyPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, company)
			argIndex++
		}
		query += " AND company_name IN (" + strings.Join(companyPlaceholders, ",") + ")"
	}

	if filters.Keywords != "" {
		query += fmt.Sprintf(" AND (company_name ILIKE $%d OR position_title ILIKE $%d OR notes ILIKE $%d)", 
							 argIndex, argIndex, argIndex)
		keyword := "%" + filters.Keywords + "%"
		args = append(args, keyword)
	}

	// 添加排序和分页
	query += " ORDER BY application_date DESC, created_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, limit, offset)
	}

	return query, args
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
	s.updateExportTask(task)
}

// saveExportTask 保存导出任务
func (s *ExportService) saveExportTask(task *model.ExportTask) error {
	query := `
		INSERT INTO export_tasks (
			task_id, user_id, status, export_type, total_records, 
			processed_records, progress, filters, options, created_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := s.db.Exec(query,
		task.TaskID,
		task.UserID,
		task.Status,
		task.ExportType,
		task.TotalRecords,
		task.ProcessedRecords,
		task.Progress,
		task.Filters,
		task.Options,
		task.CreatedAt,
		task.ExpiresAt,
	)

	return err
}

// updateExportTask 更新导出任务
func (s *ExportService) updateExportTask(task *model.ExportTask) error {
	query := `
		UPDATE export_tasks SET 
			status = $2, processed_records = $3, progress = $4, 
			file_path = $5, file_size = $6, filename = $7,
			error_message = $8, started_at = $9, completed_at = $10
		WHERE task_id = $1
	`

	_, err := s.db.Exec(query,
		task.TaskID,
		task.Status,
		task.ProcessedRecords,
		task.Progress,
		task.FilePath,
		task.FileSize,
		task.Filename,
		task.ErrorMessage,
		task.StartedAt,
		task.CompletedAt,
	)

	return err
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
	query := `
		SELECT task_id, file_path FROM export_tasks 
		WHERE expires_at < NOW() AND status = 'completed'
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var expiredTasks []struct {
		TaskID   string
		FilePath sql.NullString
	}

	for rows.Next() {
		var task struct {
			TaskID   string
			FilePath sql.NullString
		}
		if err := rows.Scan(&task.TaskID, &task.FilePath); err != nil {
			continue
		}
		expiredTasks = append(expiredTasks, task)
	}

	// 删除过期文件和数据库记录
	for _, task := range expiredTasks {
		// 删除文件
		if task.FilePath.Valid {
			os.Remove(task.FilePath.String)
		}

		// 删除数据库记录
		s.db.Exec("DELETE FROM export_tasks WHERE task_id = $1", task.TaskID)
	}

	return nil
}