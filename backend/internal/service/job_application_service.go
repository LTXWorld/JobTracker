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

// GetAll 获取用户的所有投递记录，按日期倒序排列
func (s *JobApplicationService) GetAll(userID uint) ([]model.JobApplication, error) {
	query := `
		SELECT id, user_id, company_name, position_title, application_date, status,
			job_description, salary_range, work_location, contact_info, notes,
			interview_time, reminder_time, reminder_enabled, follow_up_date,
			hr_name, hr_phone, hr_email, interview_location, interview_type,
			created_at, updated_at
		FROM job_applications
		WHERE user_id = $1
		ORDER BY application_date DESC, created_at DESC
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

// Update 更新投递记录（带用户权限检查）
func (s *JobApplicationService) Update(userID uint, id int, req *model.UpdateJobApplicationRequest) (*model.JobApplication, error) {
	// 首先检查记录是否存在且属于该用户
	existingJob, err := s.GetByID(userID, id)
	if err != nil {
		return nil, err
	}

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

	if len(setParts) == 0 {
		return existingJob, nil
	}

	// 添加updated_at更新
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// 添加WHERE条件的ID和用户ID
	args = append(args, id, userID)

	query := fmt.Sprintf(`
		UPDATE job_applications
		SET %s
		WHERE id = $%d AND user_id = $%d
	`, strings.Join(setParts, ", "), argIndex, argIndex+1)

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update job application: %w", err)
	}

	return s.GetByID(userID, id)
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

// GetStatusStatistics 获取用户的状态统计信息
func (s *JobApplicationService) GetStatusStatistics(userID uint) (map[string]interface{}, error) {
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