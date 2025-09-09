package service

import (
	"database/sql"
	"fmt"
	"jobView-backend/internal/database"
	"jobView-backend/internal/model"
	"strings"
	"time"
)

type JobApplicationService struct {
	db *database.DB
}

func NewJobApplicationService(db *database.DB) *JobApplicationService {
	return &JobApplicationService{db: db}
}

// Create 创建新的投递记录
func (s *JobApplicationService) Create(userID uint, req *model.CreateJobApplicationRequest) (*model.JobApplication, error) {
	// 如果没有提供日期，使用当前日期
	applicationDate := req.ApplicationDate
	if applicationDate == "" {
		applicationDate = time.Now().Format("2006-01-02")
	}

	// 如果没有提供状态，使用默认状态
	status := req.Status
	if status == "" {
		status = model.StatusApplied
	}
	
	// 验证状态是否有效
	if !status.IsValid() {
		return nil, fmt.Errorf("invalid status: %s", status)
	}

	// 设置默认提醒启用状态
	reminderEnabled := false
	if req.ReminderEnabled != nil {
		reminderEnabled = *req.ReminderEnabled
	}

	query := `
		INSERT INTO job_applications (
			user_id, company_name, position_title, application_date, status, 
			job_description, salary_range, work_location, contact_info, notes,
			interview_time, reminder_time, reminder_enabled, follow_up_date,
			hr_name, hr_phone, hr_email, interview_location, interview_type
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING id, created_at, updated_at
	`

	var job model.JobApplication
	err := s.db.QueryRow(query,
		userID,
		req.CompanyName,
		req.PositionTitle,
		applicationDate,
		status,
		req.JobDescription,
		req.SalaryRange,
		req.WorkLocation,
		req.ContactInfo,
		req.Notes,
		req.InterviewTime,
		req.ReminderTime,
		reminderEnabled,
		req.FollowUpDate,
		req.HRName,
		req.HRPhone,
		req.HREmail,
		req.InterviewLocation,
		req.InterviewType,
	).Scan(&job.ID, &job.CreatedAt, &job.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create job application: %w", err)
	}

	// 填充返回的数据
	job.UserID = userID
	job.CompanyName = req.CompanyName
	job.PositionTitle = req.PositionTitle
	job.ApplicationDate = applicationDate
	job.Status = status
	job.JobDescription = req.JobDescription
	job.SalaryRange = req.SalaryRange
	job.WorkLocation = req.WorkLocation
	job.ContactInfo = req.ContactInfo
	job.Notes = req.Notes
	job.InterviewTime = req.InterviewTime
	job.ReminderTime = req.ReminderTime
	job.ReminderEnabled = reminderEnabled
	job.FollowUpDate = req.FollowUpDate
	job.HRName = req.HRName
	job.HRPhone = req.HRPhone
	job.HREmail = req.HREmail
	job.InterviewLocation = req.InterviewLocation
	job.InterviewType = req.InterviewType

	return &job, nil
}

// GetByID 根据ID获取投递记录（带用户权限检查）
func (s *JobApplicationService) GetByID(userID uint, id int) (*model.JobApplication, error) {
	query := `
		SELECT id, user_id, company_name, position_title, application_date, status,
			job_description, salary_range, work_location, contact_info, notes,
			interview_time, reminder_time, reminder_enabled, follow_up_date,
			hr_name, hr_phone, hr_email, interview_location, interview_type,
			created_at, updated_at
		FROM job_applications
		WHERE id = $1 AND user_id = $2
	`

	var job model.JobApplication
	err := s.db.QueryRow(query, id, userID).Scan(
		&job.ID,
		&job.UserID,
		&job.CompanyName,
		&job.PositionTitle,
		&job.ApplicationDate,
		&job.Status,
		&job.JobDescription,
		&job.SalaryRange,
		&job.WorkLocation,
		&job.ContactInfo,
		&job.Notes,
		&job.InterviewTime,
		&job.ReminderTime,
		&job.ReminderEnabled,
		&job.FollowUpDate,
		&job.HRName,
		&job.HRPhone,
		&job.HREmail,
		&job.InterviewLocation,
		&job.InterviewType,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job application not found")
		}
		return nil, fmt.Errorf("failed to get job application: %w", err)
	}

	return &job, nil
}

// GetAllPaginated 获取用户的投递记录（分页版）
func (s *JobApplicationService) GetAllPaginated(userID uint, req model.PaginationRequest) (*model.PaginationResponse, error) {
	// 验证并设置默认值
	req.ValidateAndSetDefaults()

	// 构建WHERE条件
	whereClause := "WHERE user_id = $1"
	args := []interface{}{userID}
	argIndex := 2

	// 添加状态筛选
	if req.Status != nil {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *req.Status)
		argIndex++
	}

	// 1. 计数查询（使用索引优化）
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM job_applications %s", whereClause)
	var total int64
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count job applications: %w", err)
	}

	// 如果没有数据，直接返回空结果
	if total == 0 {
		return &model.PaginationResponse{
			Data:       []model.JobApplication{},
			Total:      0,
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalPages: 0,
			HasNext:    false,
			HasPrev:    false,
		}, nil
	}

	// 2. 验证排序字段安全性（防止SQL注入）
	allowedSortFields := map[string]bool{
		"application_date": true,
		"created_at":       true,
		"updated_at":       true,
		"company_name":     true,
		"position_title":   true,
		"status":           true,
	}
	if !allowedSortFields[req.SortBy] {
		req.SortBy = "application_date" // 默认排序字段
	}

	// 3. 数据查询（使用复合索引）
	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, company_name, position_title, application_date, status,
			job_description, salary_range, work_location, contact_info, notes,
			interview_time, reminder_time, reminder_enabled, follow_up_date,
			hr_name, hr_phone, hr_email, interview_location, interview_type,
			created_at, updated_at
		FROM job_applications 
		%s 
		ORDER BY %s %s, created_at DESC 
		LIMIT $%d OFFSET $%d
	`, whereClause, req.SortBy, req.SortDir, argIndex, argIndex+1)

	// 添加LIMIT和OFFSET参数
	args = append(args, req.PageSize, req.GetOffset())

	// 执行查询
	rows, err := s.db.Query(dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get job applications: %w", err)
	}
	defer rows.Close()

	// 扫描结果
	var jobs []model.JobApplication
	for rows.Next() {
		var job model.JobApplication
		err := rows.Scan(
			&job.ID,
			&job.UserID,
			&job.CompanyName,
			&job.PositionTitle,
			&job.ApplicationDate,
			&job.Status,
			&job.JobDescription,
			&job.SalaryRange,
			&job.WorkLocation,
			&job.ContactInfo,
			&job.Notes,
			&job.InterviewTime,
			&job.ReminderTime,
			&job.ReminderEnabled,
			&job.FollowUpDate,
			&job.HRName,
			&job.HRPhone,
			&job.HREmail,
			&job.InterviewLocation,
			&job.InterviewType,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job application: %w", err)
		}
		jobs = append(jobs, job)
	}

	// 4. 计算分页信息
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	hasNext := req.Page < totalPages
	hasPrev := req.Page > 1

	return &model.PaginationResponse{
		Data:       jobs,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}, nil
}

// GetAll 获取用户的所有投递记录，按日期倒序排列（保持兼容性）
// 注意：这个方法保留以保持后向兼容，但建议使用 GetAllPaginated
// 优化：使用复合索引 idx_job_applications_user_date 提升查询性能
func (s *JobApplicationService) GetAll(userID uint) ([]model.JobApplication, error) {
	// 优化查询：显式使用复合索引，限制返回数量避免大数据集性能问题
	query := `
		SELECT id, user_id, company_name, position_title, application_date, status,
			job_description, salary_range, work_location, contact_info, notes,
			interview_time, reminder_time, reminder_enabled, follow_up_date,
			hr_name, hr_phone, hr_email, interview_location, interview_type,
			created_at, updated_at
		FROM job_applications
		WHERE user_id = $1
		ORDER BY application_date DESC, created_at DESC
		LIMIT 500
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job applications: %w", err)
	}
	defer rows.Close()

	var jobs []model.JobApplication
	for rows.Next() {
		var job model.JobApplication
		err := rows.Scan(
			&job.ID,
			&job.UserID,
			&job.CompanyName,
			&job.PositionTitle,
			&job.ApplicationDate,
			&job.Status,
			&job.JobDescription,
			&job.SalaryRange,
			&job.WorkLocation,
			&job.ContactInfo,
			&job.Notes,
			&job.InterviewTime,
			&job.ReminderTime,
			&job.ReminderEnabled,
			&job.FollowUpDate,
			&job.HRName,
			&job.HRPhone,
			&job.HREmail,
			&job.InterviewLocation,
			&job.InterviewType,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job application: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// Update 更新投递记录（带用户权限检查）- 优化版，避免N+1查询问题
func (s *JobApplicationService) Update(userID uint, id int, req *model.UpdateJobApplicationRequest) (*model.JobApplication, error) {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.CompanyName != nil {
		setParts = append(setParts, fmt.Sprintf("company_name = $%d", argIndex))
		args = append(args, *req.CompanyName)
		argIndex++
	}

	if req.PositionTitle != nil {
		setParts = append(setParts, fmt.Sprintf("position_title = $%d", argIndex))
		args = append(args, *req.PositionTitle)
		argIndex++
	}

	if req.ApplicationDate != nil {
		setParts = append(setParts, fmt.Sprintf("application_date = $%d", argIndex))
		args = append(args, *req.ApplicationDate)
		argIndex++
	}

	if req.Status != nil {
		// 验证状态是否有效
		if !req.Status.IsValid() {
			return nil, fmt.Errorf("invalid status: %s", *req.Status)
		}
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *req.Status)
		argIndex++
	}

	if req.JobDescription != nil {
		setParts = append(setParts, fmt.Sprintf("job_description = $%d", argIndex))
		args = append(args, *req.JobDescription)
		argIndex++
	}

	if req.SalaryRange != nil {
		setParts = append(setParts, fmt.Sprintf("salary_range = $%d", argIndex))
		args = append(args, *req.SalaryRange)
		argIndex++
	}

	if req.WorkLocation != nil {
		setParts = append(setParts, fmt.Sprintf("work_location = $%d", argIndex))
		args = append(args, *req.WorkLocation)
		argIndex++
	}

	if req.ContactInfo != nil {
		setParts = append(setParts, fmt.Sprintf("contact_info = $%d", argIndex))
		args = append(args, *req.ContactInfo)
		argIndex++
	}

	if req.Notes != nil {
		setParts = append(setParts, fmt.Sprintf("notes = $%d", argIndex))
		args = append(args, *req.Notes)
		argIndex++
	}

	if req.InterviewTime != nil {
		setParts = append(setParts, fmt.Sprintf("interview_time = $%d", argIndex))
		args = append(args, *req.InterviewTime)
		argIndex++
	}

	if req.ReminderTime != nil {
		setParts = append(setParts, fmt.Sprintf("reminder_time = $%d", argIndex))
		args = append(args, *req.ReminderTime)
		argIndex++
	}

	if req.ReminderEnabled != nil {
		setParts = append(setParts, fmt.Sprintf("reminder_enabled = $%d", argIndex))
		args = append(args, *req.ReminderEnabled)
		argIndex++
	}

	if req.FollowUpDate != nil {
		setParts = append(setParts, fmt.Sprintf("follow_up_date = $%d", argIndex))
		args = append(args, *req.FollowUpDate)
		argIndex++
	}

	if req.HRName != nil {
		setParts = append(setParts, fmt.Sprintf("hr_name = $%d", argIndex))
		args = append(args, *req.HRName)
		argIndex++
	}

	if req.HRPhone != nil {
		setParts = append(setParts, fmt.Sprintf("hr_phone = $%d", argIndex))
		args = append(args, *req.HRPhone)
		argIndex++
	}

	if req.HREmail != nil {
		setParts = append(setParts, fmt.Sprintf("hr_email = $%d", argIndex))
		args = append(args, *req.HREmail)
		argIndex++
	}

	if req.InterviewLocation != nil {
		setParts = append(setParts, fmt.Sprintf("interview_location = $%d", argIndex))
		args = append(args, *req.InterviewLocation)
		argIndex++
	}

	if req.InterviewType != nil {
		setParts = append(setParts, fmt.Sprintf("interview_type = $%d", argIndex))
		args = append(args, *req.InterviewType)
		argIndex++
	}

	// 如果没有需要更新的字段，直接返回现有记录
	if len(setParts) == 0 {
		return s.GetByID(userID, id)
	}

	// 添加updated_at更新
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// 添加WHERE条件的ID和用户ID
	args = append(args, id, userID)

	// 优化：使用UPDATE ... RETURNING避免额外查询，一次SQL完成更新并返回结果
	query := fmt.Sprintf(`
		UPDATE job_applications
		SET %s
		WHERE id = $%d AND user_id = $%d
		RETURNING id, user_id, company_name, position_title, application_date, status,
			job_description, salary_range, work_location, contact_info, notes,
			interview_time, reminder_time, reminder_enabled, follow_up_date,
			hr_name, hr_phone, hr_email, interview_location, interview_type,
			created_at, updated_at
	`, strings.Join(setParts, ", "), argIndex, argIndex+1)

	var job model.JobApplication
	err := s.db.QueryRow(query, args...).Scan(
		&job.ID,
		&job.UserID,
		&job.CompanyName,
		&job.PositionTitle,
		&job.ApplicationDate,
		&job.Status,
		&job.JobDescription,
		&job.SalaryRange,
		&job.WorkLocation,
		&job.ContactInfo,
		&job.Notes,
		&job.InterviewTime,
		&job.ReminderTime,
		&job.ReminderEnabled,
		&job.FollowUpDate,
		&job.HRName,
		&job.HRPhone,
		&job.HREmail,
		&job.InterviewLocation,
		&job.InterviewType,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job application not found")
		}
		return nil, fmt.Errorf("failed to update job application: %w", err)
	}

	return &job, nil
}

// Delete 删除投递记录（带用户权限检查）
func (s *JobApplicationService) Delete(userID uint, id int) error {
	query := "DELETE FROM job_applications WHERE id = $1 AND user_id = $2"
	result, err := s.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete job application: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("job application not found")
	}

	return nil
}

// GetStatusStatistics 获取用户的状态统计信息 - 高度优化版本
// 使用覆盖索引 idx_job_applications_status_stats 避免回表查询
func (s *JobApplicationService) GetStatusStatistics(userID uint) (map[string]interface{}, error) {
	// 高度优化的查询：使用覆盖索引，只访问索引页面，不需要访问表数据
	query := `
		SELECT status, COUNT(*) as count
		FROM job_applications
		WHERE user_id = $1
		GROUP BY status
		ORDER BY count DESC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status statistics: %w", err)
	}
	defer rows.Close()

	statusCounts := make(map[string]int)
	totalCount := 0
	inProgressCount := 0
	passedCount := 0
	failedCount := 0

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status statistics: %w", err)
		}

		statusCounts[status] = count
		totalCount += count

		// 分类统计
		appStatus := model.ApplicationStatus(status)
		if appStatus.IsInProgressStatus() {
			inProgressCount += count
		} else if appStatus.IsPassedStatus() {
			passedCount += count
		} else if appStatus.IsFailedStatus() {
			failedCount += count
		}
	}

	// 检查rows.Next()过程中是否有错误
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating status statistics: %w", err)
	}

	statistics := map[string]interface{}{
		"user_id":            userID,
		"total_applications": totalCount,
		"in_progress":       inProgressCount,
		"passed":           passedCount,
		"failed":           failedCount,
		"status_breakdown": statusCounts,
	}

	// 计算通过率（如果有完成的申请）
	completedCount := passedCount + failedCount
	if completedCount > 0 {
		passRate := float64(passedCount) / float64(completedCount) * 100
		statistics["pass_rate"] = fmt.Sprintf("%.1f%%", passRate)
	} else {
		statistics["pass_rate"] = "N/A"
	}

	return statistics, nil
}

// BatchCreate 批量创建投递记录 - 高性能批量插入
func (s *JobApplicationService) BatchCreate(userID uint, applications []model.CreateJobApplicationRequest) ([]model.JobApplication, error) {
	if len(applications) == 0 {
		return []model.JobApplication{}, nil
	}

	// 限制批量操作的数量，避免性能问题
	if len(applications) > 50 {
		return nil, fmt.Errorf("batch size too large: maximum 50 applications allowed, got %d", len(applications))
	}

	// 构建批量插入SQL
	var valueStrings []string
	var valueArgs []interface{}
	argIndex := 1

	for _, req := range applications {
		// 验证和设置默认值
		applicationDate := req.ApplicationDate
		if applicationDate == "" {
			applicationDate = time.Now().Format("2006-01-02")
		}

		status := req.Status
		if status == "" {
			status = model.StatusApplied
		}

		if !status.IsValid() {
			return nil, fmt.Errorf("invalid status: %s", status)
		}

		reminderEnabled := false
		if req.ReminderEnabled != nil {
			reminderEnabled = *req.ReminderEnabled
		}

		// 构建单个记录的值占位符
		valueStrings = append(valueStrings, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			argIndex, argIndex+1, argIndex+2, argIndex+3, argIndex+4, argIndex+5, argIndex+6, argIndex+7, argIndex+8,
			argIndex+9, argIndex+10, argIndex+11, argIndex+12, argIndex+13, argIndex+14, argIndex+15, argIndex+16, argIndex+17, argIndex+18,
		))

		// 添加参数值
		valueArgs = append(valueArgs,
			userID,
			req.CompanyName,
			req.PositionTitle,
			applicationDate,
			status,
			req.JobDescription,
			req.SalaryRange,
			req.WorkLocation,
			req.ContactInfo,
			req.Notes,
			req.InterviewTime,
			req.ReminderTime,
			reminderEnabled,
			req.FollowUpDate,
			req.HRName,
			req.HRPhone,
			req.HREmail,
			req.InterviewLocation,
			req.InterviewType,
		)

		argIndex += 19
	}

	// 执行批量插入
	query := fmt.Sprintf(`
		INSERT INTO job_applications (
			user_id, company_name, position_title, application_date, status, 
			job_description, salary_range, work_location, contact_info, notes,
			interview_time, reminder_time, reminder_enabled, follow_up_date,
			hr_name, hr_phone, hr_email, interview_location, interview_type
		) VALUES %s
		RETURNING id, created_at, updated_at
	`, strings.Join(valueStrings, ", "))

	rows, err := s.db.Query(query, valueArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to batch create job applications: %w", err)
	}
	defer rows.Close()

	// 收集返回的ID和时间戳
	var results []model.JobApplication
	i := 0
	for rows.Next() {
		if i >= len(applications) {
			return nil, fmt.Errorf("unexpected number of returned rows")
		}

		var job model.JobApplication
		var id int
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan batch create result: %w", err)
		}

		// 填充完整的应用数据
		req := applications[i]
		job.ID = id
		job.UserID = userID
		job.CompanyName = req.CompanyName
		job.PositionTitle = req.PositionTitle
		job.ApplicationDate = req.ApplicationDate
		if job.ApplicationDate == "" {
			job.ApplicationDate = time.Now().Format("2006-01-02")
		}
		job.Status = req.Status
		if job.Status == "" {
			job.Status = model.StatusApplied
		}
		job.JobDescription = req.JobDescription
		job.SalaryRange = req.SalaryRange
		job.WorkLocation = req.WorkLocation
		job.ContactInfo = req.ContactInfo
		job.Notes = req.Notes
		job.InterviewTime = req.InterviewTime
		job.ReminderTime = req.ReminderTime
		job.ReminderEnabled = req.ReminderEnabled != nil && *req.ReminderEnabled
		job.FollowUpDate = req.FollowUpDate
		job.HRName = req.HRName
		job.HRPhone = req.HRPhone
		job.HREmail = req.HREmail
		job.InterviewLocation = req.InterviewLocation
		job.InterviewType = req.InterviewType
		job.CreatedAt = createdAt
		job.UpdatedAt = updatedAt

		results = append(results, job)
		i++
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating batch create results: %w", err)
	}

	return results, nil
}

// BatchUpdateStatus 批量更新状态 - 高性能批量更新
func (s *JobApplicationService) BatchUpdateStatus(userID uint, updates []model.BatchStatusUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	// 限制批量操作的数量
	if len(updates) > 100 {
		return fmt.Errorf("batch size too large: maximum 100 updates allowed, got %d", len(updates))
	}

	// 验证所有状态
	for _, update := range updates {
		if !update.Status.IsValid() {
			return fmt.Errorf("invalid status: %s for ID %d", update.Status, update.ID)
		}
	}

	// 构建临时表的值
	var valueStrings []string
	var valueArgs []interface{}
	argIndex := 1

	for _, update := range updates {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", argIndex, argIndex+1))
		valueArgs = append(valueArgs, update.ID, update.Status)
		argIndex += 2
	}

	// 使用 CTE (Common Table Expression) 进行批量更新
	query := fmt.Sprintf(`
		WITH updates(id, status) AS (VALUES %s)
		UPDATE job_applications 
		SET status = updates.status::VARCHAR, updated_at = NOW()
		FROM updates 
		WHERE job_applications.id = updates.id 
		AND job_applications.user_id = $%d
	`, strings.Join(valueStrings, ", "), argIndex)

	// 添加 userID 参数
	valueArgs = append(valueArgs, userID)

	result, err := s.db.Exec(query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to batch update status: %w", err)
	}

	// 检查更新的行数
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no job applications were updated (check user permissions and record existence)")
	}

	return nil
}

// BatchDelete 批量删除记录 - 高性能批量删除
func (s *JobApplicationService) BatchDelete(userID uint, ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	// 限制批量操作的数量
	if len(ids) > 100 {
		return fmt.Errorf("batch size too large: maximum 100 deletions allowed, got %d", len(ids))
	}

	// 构建 IN 子句的占位符
	var placeholders []string
	var args []interface{}
	
	for i, id := range ids {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+2)) // +2 因为 $1 是 userID
		args = append(args, id)
	}

	// 构建删除查询
	query := fmt.Sprintf(`
		DELETE FROM job_applications 
		WHERE user_id = $1 AND id IN (%s)
	`, strings.Join(placeholders, ", "))

	// userID 作为第一个参数
	allArgs := append([]interface{}{userID}, args...)

	result, err := s.db.Exec(query, allArgs...)
	if err != nil {
		return fmt.Errorf("failed to batch delete job applications: %w", err)
	}

	// 检查删除的行数
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no job applications were deleted (check user permissions and record existence)")
	}

	return nil
}

// SearchApplications 全文搜索投递记录 - 使用 GIN 索引优化
func (s *JobApplicationService) SearchApplications(userID uint, searchQuery string, req model.PaginationRequest) (*model.PaginationResponse, error) {
	// 验证并设置默认值
	req.ValidateAndSetDefaults()
	
	if searchQuery == "" {
		return s.GetAllPaginated(userID, req)
	}

	// 构建全文搜索查询
	whereClause := "WHERE user_id = $1 AND to_tsvector('simple', COALESCE(company_name, '') || ' ' || COALESCE(position_title, '')) @@ plainto_tsquery('simple', $2)"
	args := []interface{}{userID, searchQuery}
	argIndex := 3

	// 添加状态筛选
	if req.Status != nil {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *req.Status)
		argIndex++
	}

	// 1. 计数查询
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM job_applications %s", whereClause)
	var total int64
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	// 如果没有数据，直接返回空结果
	if total == 0 {
		return &model.PaginationResponse{
			Data:       []model.JobApplication{},
			Total:      0,
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalPages: 0,
			HasNext:    false,
			HasPrev:    false,
		}, nil
	}

	// 2. 验证排序字段安全性
	allowedSortFields := map[string]bool{
		"application_date": true,
		"created_at":       true,
		"updated_at":       true,
		"company_name":     true,
		"position_title":   true,
		"status":           true,
	}
	if !allowedSortFields[req.SortBy] {
		req.SortBy = "application_date"
	}

	// 3. 数据查询
	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, company_name, position_title, application_date, status,
			job_description, salary_range, work_location, contact_info, notes,
			interview_time, reminder_time, reminder_enabled, follow_up_date,
			hr_name, hr_phone, hr_email, interview_location, interview_type,
			created_at, updated_at,
			ts_rank_cd(to_tsvector('simple', COALESCE(company_name, '') || ' ' || COALESCE(position_title, '')), 
					  plainto_tsquery('simple', $2)) as rank
		FROM job_applications 
		%s 
		ORDER BY rank DESC, %s %s, created_at DESC 
		LIMIT $%d OFFSET $%d
	`, whereClause, req.SortBy, req.SortDir, argIndex, argIndex+1)

	// 添加LIMIT和OFFSET参数
	args = append(args, req.PageSize, req.GetOffset())

	// 执行查询
	rows, err := s.db.Query(dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search applications: %w", err)
	}
	defer rows.Close()

	// 扫描结果
	var jobs []model.JobApplication
	for rows.Next() {
		var job model.JobApplication
		var rank float64
		err := rows.Scan(
			&job.ID,
			&job.UserID,
			&job.CompanyName,
			&job.PositionTitle,
			&job.ApplicationDate,
			&job.Status,
			&job.JobDescription,
			&job.SalaryRange,
			&job.WorkLocation,
			&job.ContactInfo,
			&job.Notes,
			&job.InterviewTime,
			&job.ReminderTime,
			&job.ReminderEnabled,
			&job.FollowUpDate,
			&job.HRName,
			&job.HRPhone,
			&job.HREmail,
			&job.InterviewLocation,
			&job.InterviewType,
			&job.CreatedAt,
			&job.UpdatedAt,
			&rank, // 搜索相关度分数
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		jobs = append(jobs, job)
	}

	// 4. 计算分页信息
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	hasNext := req.Page < totalPages
	hasPrev := req.Page > 1

	return &model.PaginationResponse{
		Data:       jobs,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}, nil
}

// GetApplicationsByDateRange 按日期范围获取投递记录 - 使用日期范围索引优化
func (s *JobApplicationService) GetApplicationsByDateRange(userID uint, startDate, endDate string, req model.PaginationRequest) (*model.PaginationResponse, error) {
	// 验证并设置默认值
	req.ValidateAndSetDefaults()
	
	// 验证日期格式
	if startDate != "" && !isValidDate(startDate) {
		return nil, fmt.Errorf("invalid start date format: %s", startDate)
	}
	if endDate != "" && !isValidDate(endDate) {
		return nil, fmt.Errorf("invalid end date format: %s", endDate)
	}

	// 构建WHERE条件
	whereClause := "WHERE user_id = $1"
	args := []interface{}{userID}
	argIndex := 2

	if startDate != "" && endDate != "" {
		whereClause += fmt.Sprintf(" AND application_date BETWEEN $%d AND $%d", argIndex, argIndex+1)
		args = append(args, startDate, endDate)
		argIndex += 2
	} else if startDate != "" {
		whereClause += fmt.Sprintf(" AND application_date >= $%d", argIndex)
		args = append(args, startDate)
		argIndex++
	} else if endDate != "" {
		whereClause += fmt.Sprintf(" AND application_date <= $%d", argIndex)
		args = append(args, endDate)
		argIndex++
	}

	// 添加状态筛选
	if req.Status != nil {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *req.Status)
		argIndex++
	}

	// 1. 计数查询
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM job_applications %s", whereClause)
	var total int64
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count applications by date range: %w", err)
	}

	// 如果没有数据，直接返回空结果
	if total == 0 {
		return &model.PaginationResponse{
			Data:       []model.JobApplication{},
			Total:      0,
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalPages: 0,
			HasNext:    false,
			HasPrev:    false,
		}, nil
	}

	// 2. 验证排序字段安全性
	allowedSortFields := map[string]bool{
		"application_date": true,
		"created_at":       true,
		"updated_at":       true,
		"company_name":     true,
		"position_title":   true,
		"status":           true,
	}
	if !allowedSortFields[req.SortBy] {
		req.SortBy = "application_date"
	}

	// 3. 数据查询 - 使用日期范围索引
	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, company_name, position_title, application_date, status,
			job_description, salary_range, work_location, contact_info, notes,
			interview_time, reminder_time, reminder_enabled, follow_up_date,
			hr_name, hr_phone, hr_email, interview_location, interview_type,
			created_at, updated_at
		FROM job_applications 
		%s 
		ORDER BY %s %s, created_at DESC 
		LIMIT $%d OFFSET $%d
	`, whereClause, req.SortBy, req.SortDir, argIndex, argIndex+1)

	// 添加LIMIT和OFFSET参数
	args = append(args, req.PageSize, req.GetOffset())

	// 执行查询
	rows, err := s.db.Query(dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get applications by date range: %w", err)
	}
	defer rows.Close()

	// 扫描结果
	var jobs []model.JobApplication
	for rows.Next() {
		var job model.JobApplication
		err := rows.Scan(
			&job.ID,
			&job.UserID,
			&job.CompanyName,
			&job.PositionTitle,
			&job.ApplicationDate,
			&job.Status,
			&job.JobDescription,
			&job.SalaryRange,
			&job.WorkLocation,
			&job.ContactInfo,
			&job.Notes,
			&job.InterviewTime,
			&job.ReminderTime,
			&job.ReminderEnabled,
			&job.FollowUpDate,
			&job.HRName,
			&job.HRPhone,
			&job.HREmail,
			&job.InterviewLocation,
			&job.InterviewType,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan application by date range: %w", err)
		}
		jobs = append(jobs, job)
	}

	// 4. 计算分页信息
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	hasNext := req.Page < totalPages
	hasPrev := req.Page > 1

	return &model.PaginationResponse{
		Data:       jobs,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}, nil
}

// isValidDate 验证日期格式是否为 YYYY-MM-DD
func isValidDate(date string) bool {
	if len(date) != 10 {
		return false
	}
	for i, char := range date {
		if i == 4 || i == 7 {
			if char != '-' {
				return false
			}
		} else {
			if char < '0' || char > '9' {
				return false
			}
		}
	}
	return true
}

// GetJobApplicationsWithStatusFilters 根据状态和阶段筛选岗位申请
func (s *JobApplicationService) GetJobApplicationsWithStatusFilters(userID uint, status *model.ApplicationStatus, stage *string, req model.PaginationRequest) (*model.PaginationResponse, error) {
	// 验证并设置默认值
	req.ValidateAndSetDefaults()

	// 构建WHERE条件
	whereClause := "WHERE user_id = $1"
	args := []interface{}{userID}
	argIndex := 2

	// 添加状态筛选
	if status != nil {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *status)
		argIndex++
	}

	// 添加阶段筛选（基于状态分类）
	if stage != nil {
		stageStatuses := s.getStatusesByStage(*stage)
		if len(stageStatuses) > 0 {
			placeholders := make([]string, len(stageStatuses))
			for i, stageStatus := range stageStatuses {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, stageStatus)
				argIndex++
			}
			whereClause += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ", "))
		}
	}

	// 1. 计数查询
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM job_applications %s", whereClause)
	var total int64
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count filtered job applications: %w", err)
	}

	// 如果没有数据，直接返回空结果
	if total == 0 {
		return &model.PaginationResponse{
			Data:       []model.JobApplication{},
			Total:      0,
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalPages: 0,
			HasNext:    false,
			HasPrev:    false,
		}, nil
	}

	// 2. 验证排序字段安全性
	allowedSortFields := map[string]bool{
		"application_date": true,
		"created_at":       true,
		"updated_at":       true,
		"company_name":     true,
		"position_title":   true,
		"status":           true,
	}
	if !allowedSortFields[req.SortBy] {
		req.SortBy = "application_date"
	}

	// 3. 数据查询
	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, company_name, position_title, application_date, status,
			job_description, salary_range, work_location, contact_info, notes,
			interview_time, reminder_time, reminder_enabled, follow_up_date,
			hr_name, hr_phone, hr_email, interview_location, interview_type,
			created_at, updated_at
		FROM job_applications 
		%s 
		ORDER BY %s %s, created_at DESC 
		LIMIT $%d OFFSET $%d
	`, whereClause, req.SortBy, req.SortDir, argIndex, argIndex+1)

	// 添加LIMIT和OFFSET参数
	args = append(args, req.PageSize, req.GetOffset())

	// 执行查询
	rows, err := s.db.Query(dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered job applications: %w", err)
	}
	defer rows.Close()

	// 扫描结果
	var jobs []model.JobApplication
	for rows.Next() {
		var job model.JobApplication
		err := rows.Scan(
			&job.ID,
			&job.UserID,
			&job.CompanyName,
			&job.PositionTitle,
			&job.ApplicationDate,
			&job.Status,
			&job.JobDescription,
			&job.SalaryRange,
			&job.WorkLocation,
			&job.ContactInfo,
			&job.Notes,
			&job.InterviewTime,
			&job.ReminderTime,
			&job.ReminderEnabled,
			&job.FollowUpDate,
			&job.HRName,
			&job.HRPhone,
			&job.HREmail,
			&job.InterviewLocation,
			&job.InterviewType,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan filtered job application: %w", err)
		}
		jobs = append(jobs, job)
	}

	// 4. 计算分页信息
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	hasNext := req.Page < totalPages
	hasPrev := req.Page > 1

	return &model.PaginationResponse{
		Data:       jobs,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}, nil
}

// getStatusesByStage 根据阶段获取对应的状态列表
func (s *JobApplicationService) getStatusesByStage(stage string) []string {
	stageMap := map[string][]string{
		"application": {"已投递"},
		"screening": {"简历筛选中", "简历筛选未通过"},
		"written_test": {"笔试中", "笔试通过", "笔试未通过"},
		"interviews": {
			"一面中", "一面通过", "一面未通过",
			"二面中", "二面通过", "二面未通过",
			"三面中", "三面通过", "三面未通过",
			"HR面中", "HR面通过", "HR面未通过",
		},
		"final": {
			"待发offer", "已收到offer", "已接受offer",
			"已拒绝", "流程结束",
		},
		"in_progress": {"已投递", "简历筛选中", "笔试中", "一面中", "二面中", "三面中", "HR面中"},
		"passed": {"笔试通过", "一面通过", "二面通过", "三面通过", "HR面通过", "待发offer", "已收到offer", "已接受offer", "流程结束"},
		"failed": {"简历筛选未通过", "笔试未通过", "一面未通过", "二面未通过", "三面未通过", "HR面未通过", "已拒绝"},
	}

	if statuses, exists := stageMap[stage]; exists {
		return statuses
	}
	return []string{}
}

// GetDashboardData 获取仪表板数据
func (s *JobApplicationService) GetDashboardData(userID uint) (map[string]interface{}, error) {
	// 获取状态统计
	statistics, err := s.GetStatusStatistics(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status statistics: %w", err)
	}

	// 获取最近更新的申请
	recentQuery := `
		SELECT id, company_name, position_title, status, updated_at
		FROM job_applications
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT 10
	`

	rows, err := s.db.Query(recentQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent applications: %w", err)
	}
	defer rows.Close()

	var recentApplications []map[string]interface{}
	for rows.Next() {
		var id int
		var companyName, positionTitle, status string
		var updatedAt time.Time

		err := rows.Scan(&id, &companyName, &positionTitle, &status, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recent application: %w", err)
		}

		recentApplications = append(recentApplications, map[string]interface{}{
			"id":             id,
			"company_name":   companyName,
			"position_title": positionTitle,
			"status":         status,
			"updated_at":     updatedAt,
		})
	}

	// 获取即将到来的面试
	upcomingQuery := `
		SELECT id, company_name, position_title, interview_time, interview_type
		FROM job_applications
		WHERE user_id = $1 AND interview_time > NOW() AND interview_time <= NOW() + INTERVAL '7 days'
		ORDER BY interview_time ASC
		LIMIT 5
	`

	upcomingRows, err := s.db.Query(upcomingQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming interviews: %w", err)
	}
	defer upcomingRows.Close()

	var upcomingInterviews []map[string]interface{}
	for upcomingRows.Next() {
		var id int
		var companyName, positionTitle string
		var interviewTime time.Time
		var interviewType sql.NullString

		err := upcomingRows.Scan(&id, &companyName, &positionTitle, &interviewTime, &interviewType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan upcoming interview: %w", err)
		}

		interview := map[string]interface{}{
			"id":             id,
			"company_name":   companyName,
			"position_title": positionTitle,
			"interview_time": interviewTime,
		}

		if interviewType.Valid {
			interview["interview_type"] = interviewType.String
		}

		upcomingInterviews = append(upcomingInterviews, interview)
	}

	// 获取每日申请数据（最近30天）
	dailyStatsQuery := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM job_applications
		WHERE user_id = $1 AND created_at >= CURRENT_DATE - INTERVAL '30 days'
		GROUP BY DATE(created_at)
		ORDER BY date DESC
	`

	dailyRows, err := s.db.Query(dailyStatsQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily stats: %w", err)
	}
	defer dailyRows.Close()

	var dailyStats []map[string]interface{}
	for dailyRows.Next() {
		var date time.Time
		var count int

		err := dailyRows.Scan(&date, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily stats: %w", err)
		}

		dailyStats = append(dailyStats, map[string]interface{}{
			"date":  date.Format("2006-01-02"),
			"count": count,
		})
	}

	// 构建仪表板数据
	dashboard := map[string]interface{}{
		"statistics":         statistics,
		"recent_applications": recentApplications,
		"upcoming_interviews": upcomingInterviews,
		"daily_stats":        dailyStats,
		"generated_at":       time.Now(),
	}

	return dashboard, nil
}