package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"jobView-backend/internal/database"
	"jobView-backend/internal/model"
)

// JobApplicationRepository 提供 JobApplication 的GORM Raw实现
type JobApplicationRepository interface {
	Create(userID uint, req *model.CreateJobApplicationRequest) (*model.JobApplication, error)
	GetByID(userID uint, id int) (*model.JobApplication, error)
	GetAllPaginated(userID uint, req model.PaginationRequest) (*model.PaginationResponse, error)
	GetAll(userID uint) ([]model.JobApplication, error)
	Update(userID uint, id int, req *model.UpdateJobApplicationRequest) (*model.JobApplication, error)
	Delete(userID uint, id int) error
	GetStatusStatistics(userID uint) (map[string]int, error)
	BatchCreate(userID uint, applications []model.CreateJobApplicationRequest) ([]model.JobApplication, error)
	BatchUpdateStatus(userID uint, updates []model.BatchStatusUpdate) error
	BatchDelete(userID uint, ids []int) error
	Search(userID uint, searchQuery string, req model.PaginationRequest) (*model.PaginationResponse, error)
	ListByDateRange(userID uint, startDate, endDate string, req model.PaginationRequest) (*model.PaginationResponse, error)
	ListWithStatusFilters(userID uint, status *model.ApplicationStatus, stageStatuses []string, req model.PaginationRequest) (*model.PaginationResponse, error)
	ListRecentApplications(userID uint, limit int) ([]map[string]interface{}, error)
	ListUpcomingInterviews(userID uint, limit int) ([]map[string]interface{}, error)
	ListDailyStats(userID uint, days int) ([]map[string]interface{}, error)
}

type jobAppRepo struct{ db *database.DB }

func NewJobApplicationRepository(db *database.DB) JobApplicationRepository {
	return &jobAppRepo{db: db}
}

func (r *jobAppRepo) ensureDB() error {
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return nil
}

func (r *jobAppRepo) queryRow(query string, args ...interface{}) (*sql.Row, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}
	if r.db.ORM != nil {
		return r.db.ORM.Raw(query, args...).Row(), nil
	}
	return r.db.QueryRow(query, args...), nil
}

func (r *jobAppRepo) queryRows(query string, args ...interface{}) (*sql.Rows, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}
	if r.db.ORM != nil {
		return r.db.ORM.Raw(query, args...).Rows()
	}
	return r.db.Query(query, args...)
}

func (r *jobAppRepo) Create(userID uint, req *model.CreateJobApplicationRequest) (*model.JobApplication, error) {
	if r.db.ORM == nil {
		return nil, fmt.Errorf("gorm not initialized")
	}
	applicationDate := req.ApplicationDate
	if applicationDate == "" {
		applicationDate = time.Now().Format("2006-01-02")
	}
	status := req.Status
	if status == "" {
		status = model.StatusApplied
	}
	reminderEnabled := false
	if req.ReminderEnabled != nil {
		reminderEnabled = *req.ReminderEnabled
	}

	query := `INSERT INTO job_applications (
        user_id, company_name, position_title, application_date, status,
        job_description, salary_range, work_location, contact_info, notes,
        interview_time, reminder_time, reminder_enabled, follow_up_date,
        hr_name, hr_phone, hr_email, interview_location, interview_type,
        company_attribute
    ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
    RETURNING id, created_at, updated_at`

	var job model.JobApplication
	row := r.db.ORM.Raw(query,
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
		func() interface{} {
			if strings.TrimSpace(req.CompanyAttribute) == "" {
				return nil
			}
			return req.CompanyAttribute
		}(),
	).Row()
	if err := row.Scan(&job.ID, &job.CreatedAt, &job.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to create job application: %w", err)
	}

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
	if strings.TrimSpace(req.CompanyAttribute) != "" {
		ca := req.CompanyAttribute
		job.CompanyAttribute = &ca
	}
	return &job, nil
}

func (r *jobAppRepo) GetByID(userID uint, id int) (*model.JobApplication, error) {
	if r.db.ORM == nil {
		return nil, fmt.Errorf("gorm not initialized")
	}
	query := `SELECT id, user_id, company_name, position_title, application_date, status,
        job_description, salary_range, work_location, contact_info, notes,
        interview_time, reminder_time, reminder_enabled, follow_up_date,
        hr_name, hr_phone, hr_email, interview_location, interview_type,
        company_attribute,
        created_at, updated_at FROM job_applications WHERE id=$1 AND user_id=$2`
	var job model.JobApplication
	row := r.db.ORM.Raw(query, id, userID).Row()
	if err := row.Scan(
		&job.ID, &job.UserID, &job.CompanyName, &job.PositionTitle, &job.ApplicationDate, &job.Status,
		&job.JobDescription, &job.SalaryRange, &job.WorkLocation, &job.ContactInfo, &job.Notes,
		&job.InterviewTime, &job.ReminderTime, &job.ReminderEnabled, &job.FollowUpDate,
		&job.HRName, &job.HRPhone, &job.HREmail, &job.InterviewLocation, &job.InterviewType,
		&job.CompanyAttribute,
		&job.CreatedAt, &job.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job application not found")
		}
		return nil, fmt.Errorf("failed to get job application: %w", err)
	}
	return &job, nil
}

func (r *jobAppRepo) GetAll(userID uint) ([]model.JobApplication, error) {
	if r.db.ORM == nil {
		return nil, fmt.Errorf("gorm not initialized")
	}
	query := `SELECT id, user_id, company_name, position_title, application_date, status,
        job_description, salary_range, work_location, contact_info, notes,
        interview_time, reminder_time, reminder_enabled, follow_up_date,
        hr_name, hr_phone, hr_email, interview_location, interview_type,
        company_attribute,
        created_at, updated_at FROM job_applications WHERE user_id = $1
        ORDER BY application_date DESC, created_at DESC LIMIT 500`
	rows, err := r.db.ORM.Raw(query, userID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get job applications: %w", err)
	}
	defer rows.Close()
	var list []model.JobApplication
	for rows.Next() {
		var job model.JobApplication
		if err := rows.Scan(&job.ID, &job.UserID, &job.CompanyName, &job.PositionTitle, &job.ApplicationDate, &job.Status,
			&job.JobDescription, &job.SalaryRange, &job.WorkLocation, &job.ContactInfo, &job.Notes,
			&job.InterviewTime, &job.ReminderTime, &job.ReminderEnabled, &job.FollowUpDate,
			&job.HRName, &job.HRPhone, &job.HREmail, &job.InterviewLocation, &job.InterviewType,
			&job.CompanyAttribute,
			&job.CreatedAt, &job.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan job application: %w", err)
		}
		list = append(list, job)
	}
	return list, nil
}

func (r *jobAppRepo) GetAllPaginated(userID uint, req model.PaginationRequest) (*model.PaginationResponse, error) {
	if r.db.ORM == nil {
		return nil, fmt.Errorf("gorm not initialized")
	}
	req.ValidateAndSetDefaults()
	where := "WHERE user_id = $1"
	args := []interface{}{userID}
	idx := 2
	if req.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, *req.Status)
		idx++
	}

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM job_applications %s", where)
	if err := r.db.ORM.Raw(countSQL, args...).Row().Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count job applications: %w", err)
	}
	if total == 0 {
		return &model.PaginationResponse{Data: []model.JobApplication{}, Total: 0, Page: req.Page, PageSize: req.PageSize}, nil
	}

	allowed := map[string]bool{"application_date": true, "created_at": true, "updated_at": true, "company_name": true, "position_title": true, "status": true}
	if !allowed[req.SortBy] {
		req.SortBy = "application_date"
	}
	dataSQL := fmt.Sprintf(`SELECT id, user_id, company_name, position_title, application_date, status,
        job_description, salary_range, work_location, contact_info, notes,
        interview_time, reminder_time, reminder_enabled, follow_up_date,
        hr_name, hr_phone, hr_email, interview_location, interview_type,
        company_attribute,
        created_at, updated_at FROM job_applications %s ORDER BY %s %s, created_at DESC LIMIT $%d OFFSET $%d`,
		where, req.SortBy, req.SortDir, idx, idx+1)
	args = append(args, req.PageSize, req.GetOffset())
	rows, err := r.db.ORM.Raw(dataSQL, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get job applications: %w", err)
	}
	defer rows.Close()
	var jobs []model.JobApplication
	for rows.Next() {
		var job model.JobApplication
		if err := rows.Scan(&job.ID, &job.UserID, &job.CompanyName, &job.PositionTitle, &job.ApplicationDate, &job.Status,
			&job.JobDescription, &job.SalaryRange, &job.WorkLocation, &job.ContactInfo, &job.Notes,
			&job.InterviewTime, &job.ReminderTime, &job.ReminderEnabled, &job.FollowUpDate,
			&job.HRName, &job.HRPhone, &job.HREmail, &job.InterviewLocation, &job.InterviewType,
			&job.CompanyAttribute,
			&job.CreatedAt, &job.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan job application: %w", err)
		}
		jobs = append(jobs, job)
	}
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	return &model.PaginationResponse{Data: jobs, Total: total, Page: req.Page, PageSize: req.PageSize, TotalPages: totalPages, HasNext: req.Page < totalPages, HasPrev: req.Page > 1}, nil
}

func (r *jobAppRepo) Update(userID uint, id int, req *model.UpdateJobApplicationRequest) (*model.JobApplication, error) {
	if r.db.ORM == nil {
		return nil, fmt.Errorf("gorm not initialized")
	}
	setParts := []string{}
	args := []interface{}{}
	idx := 1
	if req.CompanyName != nil {
		setParts = append(setParts, fmt.Sprintf("company_name=$%d", idx))
		args = append(args, *req.CompanyName)
		idx++
	}
	if req.PositionTitle != nil {
		setParts = append(setParts, fmt.Sprintf("position_title=$%d", idx))
		args = append(args, *req.PositionTitle)
		idx++
	}
	if req.ApplicationDate != nil {
		setParts = append(setParts, fmt.Sprintf("application_date=$%d", idx))
		args = append(args, *req.ApplicationDate)
		idx++
	}
	if req.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status=$%d", idx))
		args = append(args, *req.Status)
		idx++
	}
	if req.JobDescription != nil {
		setParts = append(setParts, fmt.Sprintf("job_description=$%d", idx))
		args = append(args, *req.JobDescription)
		idx++
	}
	if req.SalaryRange != nil {
		setParts = append(setParts, fmt.Sprintf("salary_range=$%d", idx))
		args = append(args, *req.SalaryRange)
		idx++
	}
	if req.WorkLocation != nil {
		setParts = append(setParts, fmt.Sprintf("work_location=$%d", idx))
		args = append(args, *req.WorkLocation)
		idx++
	}
	if req.ContactInfo != nil {
		setParts = append(setParts, fmt.Sprintf("contact_info=$%d", idx))
		args = append(args, *req.ContactInfo)
		idx++
	}
	if req.Notes != nil {
		setParts = append(setParts, fmt.Sprintf("notes=$%d", idx))
		args = append(args, *req.Notes)
		idx++
	}
	if req.InterviewTime != nil {
		setParts = append(setParts, fmt.Sprintf("interview_time=$%d", idx))
		args = append(args, *req.InterviewTime)
		idx++
	}
	if req.ReminderTime != nil {
		setParts = append(setParts, fmt.Sprintf("reminder_time=$%d", idx))
		args = append(args, *req.ReminderTime)
		idx++
	}
	if req.ReminderEnabled != nil {
		setParts = append(setParts, fmt.Sprintf("reminder_enabled=$%d", idx))
		args = append(args, *req.ReminderEnabled)
		idx++
	}
	if req.FollowUpDate != nil {
		setParts = append(setParts, fmt.Sprintf("follow_up_date=$%d", idx))
		args = append(args, *req.FollowUpDate)
		idx++
	}
	if req.HRName != nil {
		setParts = append(setParts, fmt.Sprintf("hr_name=$%d", idx))
		args = append(args, *req.HRName)
		idx++
	}
	if req.HRPhone != nil {
		setParts = append(setParts, fmt.Sprintf("hr_phone=$%d", idx))
		args = append(args, *req.HRPhone)
		idx++
	}
	if req.HREmail != nil {
		setParts = append(setParts, fmt.Sprintf("hr_email=$%d", idx))
		args = append(args, *req.HREmail)
		idx++
	}
	if req.InterviewLocation != nil {
		setParts = append(setParts, fmt.Sprintf("interview_location=$%d", idx))
		args = append(args, *req.InterviewLocation)
		idx++
	}
	if req.InterviewType != nil {
		setParts = append(setParts, fmt.Sprintf("interview_type=$%d", idx))
		args = append(args, *req.InterviewType)
		idx++
	}
	if req.CompanyAttribute != nil {
		setParts = append(setParts, fmt.Sprintf("company_attribute=$%d", idx))
		if strings.TrimSpace(*req.CompanyAttribute) == "" {
			args = append(args, nil)
		} else {
			args = append(args, *req.CompanyAttribute)
		}
		idx++
	}
	if len(setParts) == 0 {
		return r.GetByID(userID, id)
	}
	setParts = append(setParts, fmt.Sprintf("updated_at=$%d", idx))
	args = append(args, time.Now())
	idx++
	args = append(args, id, userID)
	query := fmt.Sprintf(`UPDATE job_applications SET %s WHERE id=$%d AND user_id=$%d RETURNING id, user_id, company_name, position_title, application_date, status,
        job_description, salary_range, work_location, contact_info, notes,
        interview_time, reminder_time, reminder_enabled, follow_up_date,
        hr_name, hr_phone, hr_email, interview_location, interview_type,
        company_attribute,
        created_at, updated_at`, strings.Join(setParts, ", "), idx, idx+1)
	var job model.JobApplication
	row := r.db.ORM.Raw(query, args...).Row()
	if err := row.Scan(&job.ID, &job.UserID, &job.CompanyName, &job.PositionTitle, &job.ApplicationDate, &job.Status, &job.JobDescription, &job.SalaryRange, &job.WorkLocation, &job.ContactInfo, &job.Notes, &job.InterviewTime, &job.ReminderTime, &job.ReminderEnabled, &job.FollowUpDate, &job.HRName, &job.HRPhone, &job.HREmail, &job.InterviewLocation, &job.InterviewType, &job.CompanyAttribute, &job.CreatedAt, &job.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job application not found")
		}
		return nil, fmt.Errorf("failed to update job application: %w", err)
	}
	return &job, nil
}

func (r *jobAppRepo) Delete(userID uint, id int) error {
	if r.db.ORM == nil {
		return fmt.Errorf("gorm not initialized")
	}
	res := r.db.ORM.Exec("DELETE FROM job_applications WHERE id = $1 AND user_id = $2", id, userID)
	if res.Error != nil {
		return fmt.Errorf("failed to delete job application: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("job application not found")
	}
	return nil
}

func (r *jobAppRepo) GetStatusStatistics(userID uint) (map[string]int, error) {
	rows, err := r.queryRows("SELECT status, COUNT(*) FROM job_applications WHERE user_id = $1 GROUP BY status", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status statistics: %w", err)
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status statistics: %w", err)
		}
		result[status] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate status statistics: %w", err)
	}
	return result, nil
}

func (r *jobAppRepo) BatchCreate(userID uint, applications []model.CreateJobApplicationRequest) ([]model.JobApplication, error) {
	if len(applications) == 0 {
		return []model.JobApplication{}, nil
	}
	if len(applications) > 50 {
		return nil, fmt.Errorf("batch size too large: maximum 50 applications allowed, got %d", len(applications))
	}
	if err := r.ensureDB(); err != nil {
		return nil, err
	}

	var valueStrings []string
	var valueArgs []interface{}
	argIndex := 1

	for _, req := range applications {
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

		valueStrings = append(valueStrings, fmt.Sprintf(
			"($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			argIndex, argIndex+1, argIndex+2, argIndex+3, argIndex+4,
			argIndex+5, argIndex+6, argIndex+7, argIndex+8, argIndex+9,
			argIndex+10, argIndex+11, argIndex+12, argIndex+13, argIndex+14,
			argIndex+15, argIndex+16, argIndex+17, argIndex+18, argIndex+19,
		))
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
			func() interface{} {
				if strings.TrimSpace(req.CompanyAttribute) == "" {
					return nil
				}
				return req.CompanyAttribute
			}(),
		)
		argIndex += 20
	}

	query := fmt.Sprintf(`
        INSERT INTO job_applications (
            user_id, company_name, position_title, application_date, status,
            job_description, salary_range, work_location, contact_info, notes,
            interview_time, reminder_time, reminder_enabled, follow_up_date,
            hr_name, hr_phone, hr_email, interview_location, interview_type,
            company_attribute
        ) VALUES %s
        RETURNING id, user_id, company_name, position_title, application_date, status,
            job_description, salary_range, work_location, contact_info, notes,
            interview_time, reminder_time, reminder_enabled, follow_up_date,
            hr_name, hr_phone, hr_email, interview_location, interview_type,
            company_attribute,
            created_at, updated_at
    `, strings.Join(valueStrings, ","))

	var rows *sql.Rows
	var err error
	if r.db.ORM != nil {
		rows, err = r.db.ORM.Raw(query, valueArgs...).Rows()
	} else {
		rows, err = r.db.Query(query, valueArgs...)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to batch create job applications: %w", err)
	}
	defer rows.Close()

	var results []model.JobApplication
	for rows.Next() {
		var job model.JobApplication
		if err := rows.Scan(
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
			&job.CompanyAttribute,
			&job.CreatedAt,
			&job.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan batch create result: %w", err)
		}
		results = append(results, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating batch create results: %w", err)
	}

	return results, nil
}

func (r *jobAppRepo) BatchUpdateStatus(userID uint, updates []model.BatchStatusUpdate) error {
	if len(updates) == 0 {
		return nil
	}
	if len(updates) > 100 {
		return fmt.Errorf("batch size too large: maximum 100 updates allowed, got %d", len(updates))
	}
	if err := r.ensureDB(); err != nil {
		return err
	}

	var valueStrings []string
	var valueArgs []interface{}
	argIndex := 1

	for _, update := range updates {
		if !update.Status.IsValid() {
			return fmt.Errorf("invalid status: %s for ID %d", update.Status, update.ID)
		}
		valueStrings = append(valueStrings, fmt.Sprintf("($%d,$%d)", argIndex, argIndex+1))
		valueArgs = append(valueArgs, update.ID, update.Status)
		argIndex += 2
	}

	query := fmt.Sprintf(`
        WITH updates(id, status) AS (VALUES %s)
        UPDATE job_applications
        SET status = updates.status::VARCHAR, updated_at = NOW()
        FROM updates
        WHERE job_applications.id = updates.id AND job_applications.user_id = $%d
    `, strings.Join(valueStrings, ","), argIndex)

	valueArgs = append(valueArgs, userID)

	var res sql.Result
	var err error
	if r.db.ORM != nil {
		exec := r.db.ORM.Exec(query, valueArgs...)
		if exec.Error != nil {
			return fmt.Errorf("failed to batch update status: %w", exec.Error)
		}
		if exec.RowsAffected == 0 {
			return fmt.Errorf("no job applications were updated (check user permissions and record existence)")
		}
		return nil
	}
	res, err = r.db.Exec(query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to batch update status: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no job applications were updated (check user permissions and record existence)")
	}
	return nil
}

func (r *jobAppRepo) BatchDelete(userID uint, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	if len(ids) > 100 {
		return fmt.Errorf("batch size too large: maximum 100 deletions allowed, got %d", len(ids))
	}
	if err := r.ensureDB(); err != nil {
		return err
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, userID)
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, id)
	}

	query := fmt.Sprintf("DELETE FROM job_applications WHERE user_id = $1 AND id IN (%s)", strings.Join(placeholders, ","))

	if r.db.ORM != nil {
		exec := r.db.ORM.Exec(query, args...)
		if exec.Error != nil {
			return fmt.Errorf("failed to batch delete job applications: %w", exec.Error)
		}
		if exec.RowsAffected == 0 {
			return fmt.Errorf("no job applications were deleted (check user permissions and record existence)")
		}
		return nil
	}
	res, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to batch delete job applications: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no job applications were deleted (check user permissions and record existence)")
	}
	return nil
}

func (r *jobAppRepo) Search(userID uint, searchQuery string, req model.PaginationRequest) (*model.PaginationResponse, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}
	req.ValidateAndSetDefaults()
	keyword := "%" + strings.TrimSpace(searchQuery) + "%"

	countQuery := `SELECT COUNT(*) FROM job_applications WHERE user_id = $1 AND (company_name ILIKE $2 OR position_title ILIKE $2 OR notes ILIKE $2)`
	row, err := r.queryRow(countQuery, userID, keyword)
	if err != nil {
		return nil, fmt.Errorf("failed to count search results: %w", err)
	}
	var total int64
	if err := row.Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to scan search count: %w", err)
	}
	if total == 0 {
		return &model.PaginationResponse{Data: []model.JobApplication{}, Total: 0, Page: req.Page, PageSize: req.PageSize}, nil
	}

	allowed := map[string]bool{"application_date": true, "created_at": true, "updated_at": true, "company_name": true, "position_title": true, "status": true}
	if !allowed[req.SortBy] {
		req.SortBy = "application_date"
	}

	dataQuery := fmt.Sprintf(`
        SELECT id, user_id, company_name, position_title, application_date, status,
               job_description, salary_range, work_location, contact_info, notes,
               interview_time, reminder_time, reminder_enabled, follow_up_date,
               hr_name, hr_phone, hr_email, interview_location, interview_type,
               company_attribute,
               created_at, updated_at
        FROM job_applications
        WHERE user_id = $1 AND (company_name ILIKE $2 OR position_title ILIKE $2 OR notes ILIKE $2)
        ORDER BY %s %s, created_at DESC
        LIMIT $3 OFFSET $4
    `, req.SortBy, req.SortDir)

	rows, err := r.queryRows(dataQuery, userID, keyword, req.PageSize, req.GetOffset())
	if err != nil {
		return nil, fmt.Errorf("failed to search job applications: %w", err)
	}
	defer rows.Close()

	var jobs []model.JobApplication
	for rows.Next() {
		var job model.JobApplication
		if err := rows.Scan(
			&job.ID, &job.UserID, &job.CompanyName, &job.PositionTitle, &job.ApplicationDate, &job.Status,
			&job.JobDescription, &job.SalaryRange, &job.WorkLocation, &job.ContactInfo, &job.Notes,
			&job.InterviewTime, &job.ReminderTime, &job.ReminderEnabled, &job.FollowUpDate,
			&job.HRName, &job.HRPhone, &job.HREmail, &job.InterviewLocation, &job.InterviewType,
			&job.CompanyAttribute,
			&job.CreatedAt, &job.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate search results: %w", err)
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	return &model.PaginationResponse{
		Data:       jobs,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1,
	}, nil
}

func (r *jobAppRepo) ListByDateRange(userID uint, startDate, endDate string, req model.PaginationRequest) (*model.PaginationResponse, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}
	req.ValidateAndSetDefaults()

	countQuery := `SELECT COUNT(*) FROM job_applications WHERE user_id = $1 AND application_date BETWEEN $2 AND $3`
	row, err := r.queryRow(countQuery, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to count job applications: %w", err)
	}
	var total int64
	if err := row.Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to scan total: %w", err)
	}
	if total == 0 {
		return &model.PaginationResponse{Data: []model.JobApplication{}, Total: 0, Page: req.Page, PageSize: req.PageSize}, nil
	}

	allowed := map[string]bool{"application_date": true, "created_at": true, "updated_at": true, "company_name": true, "position_title": true, "status": true}
	if !allowed[req.SortBy] {
		req.SortBy = "application_date"
	}

	dataQuery := fmt.Sprintf(`
        SELECT id, user_id, company_name, position_title, application_date, status,
               job_description, salary_range, work_location, contact_info, notes,
               interview_time, reminder_time, reminder_enabled, follow_up_date,
               hr_name, hr_phone, hr_email, interview_location, interview_type,
               company_attribute,
               created_at, updated_at
        FROM job_applications
        WHERE user_id = $1 AND application_date BETWEEN $2 AND $3
        ORDER BY %s %s, created_at DESC
        LIMIT $4 OFFSET $5
    `, req.SortBy, req.SortDir)

	rows, err := r.queryRows(dataQuery, userID, startDate, endDate, req.PageSize, req.GetOffset())
	if err != nil {
		return nil, fmt.Errorf("failed to list job applications: %w", err)
	}
	defer rows.Close()

	var jobs []model.JobApplication
	for rows.Next() {
		var job model.JobApplication
		if err := rows.Scan(
			&job.ID, &job.UserID, &job.CompanyName, &job.PositionTitle, &job.ApplicationDate, &job.Status,
			&job.JobDescription, &job.SalaryRange, &job.WorkLocation, &job.ContactInfo, &job.Notes,
			&job.InterviewTime, &job.ReminderTime, &job.ReminderEnabled, &job.FollowUpDate,
			&job.HRName, &job.HRPhone, &job.HREmail, &job.InterviewLocation, &job.InterviewType,
			&job.CompanyAttribute,
			&job.CreatedAt, &job.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan job application: %w", err)
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate job applications: %w", err)
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	return &model.PaginationResponse{
		Data:       jobs,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1,
	}, nil
}

func (r *jobAppRepo) ListWithStatusFilters(userID uint, status *model.ApplicationStatus, stageStatuses []string, req model.PaginationRequest) (*model.PaginationResponse, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}
	req.ValidateAndSetDefaults()

	where := "WHERE user_id = $1"
	args := []interface{}{userID}
	idx := 2

	if status != nil {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, *status)
		idx++
	}
	if len(stageStatuses) > 0 {
		placeholders := make([]string, len(stageStatuses))
		for i, st := range stageStatuses {
			placeholders[i] = fmt.Sprintf("$%d", idx)
			args = append(args, st)
			idx++
		}
		where += " AND status IN (" + strings.Join(placeholders, ",") + ")"
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM job_applications %s", where)
	row, err := r.queryRow(countQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to count status filtered results: %w", err)
	}
	var total int64
	if err := row.Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to scan total: %w", err)
	}
	if total == 0 {
		return &model.PaginationResponse{Data: []model.JobApplication{}, Total: 0, Page: req.Page, PageSize: req.PageSize}, nil
	}

	allowed := map[string]bool{"application_date": true, "created_at": true, "updated_at": true, "company_name": true, "position_title": true, "status": true}
	if !allowed[req.SortBy] {
		req.SortBy = "application_date"
	}

	dataQuery := fmt.Sprintf(`SELECT id, user_id, company_name, position_title, application_date, status,
            job_description, salary_range, work_location, contact_info, notes,
            interview_time, reminder_time, reminder_enabled, follow_up_date,
            hr_name, hr_phone, hr_email, interview_location, interview_type,
            company_attribute,
            created_at, updated_at
        FROM job_applications %s
        ORDER BY %s %s, created_at DESC
        LIMIT $%d OFFSET $%d`, where, req.SortBy, req.SortDir, idx, idx+1)

	args = append(args, req.PageSize, req.GetOffset())
	rows, err := r.queryRows(dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list job applications with status filters: %w", err)
	}
	defer rows.Close()

	var jobs []model.JobApplication
	for rows.Next() {
		var job model.JobApplication
		if err := rows.Scan(
			&job.ID, &job.UserID, &job.CompanyName, &job.PositionTitle, &job.ApplicationDate, &job.Status,
			&job.JobDescription, &job.SalaryRange, &job.WorkLocation, &job.ContactInfo, &job.Notes,
			&job.InterviewTime, &job.ReminderTime, &job.ReminderEnabled, &job.FollowUpDate,
			&job.HRName, &job.HRPhone, &job.HREmail, &job.InterviewLocation, &job.InterviewType,
			&job.CompanyAttribute,
			&job.CreatedAt, &job.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan job application: %w", err)
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate job applications: %w", err)
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	return &model.PaginationResponse{
		Data:       jobs,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1,
	}, nil
}

func (r *jobAppRepo) ListRecentApplications(userID uint, limit int) ([]map[string]interface{}, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 5
	}

	query := `
        SELECT id, company_name, position_title, status, updated_at
        FROM job_applications
        WHERE user_id = $1
        ORDER BY updated_at DESC
        LIMIT $2`

	rows, err := r.queryRows(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent applications: %w", err)
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var (
			id            int
			companyName   string
			positionTitle string
			status        string
			updatedAt     time.Time
		)
		if err := rows.Scan(&id, &companyName, &positionTitle, &status, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan recent application: %w", err)
		}
		result = append(result, map[string]interface{}{
			"id":             id,
			"company_name":   companyName,
			"position_title": positionTitle,
			"status":         status,
			"updated_at":     updatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate recent applications: %w", err)
	}
	return result, nil
}

func (r *jobAppRepo) ListUpcomingInterviews(userID uint, limit int) ([]map[string]interface{}, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 5
	}

	query := `
        SELECT id, company_name, position_title, interview_time, interview_type
        FROM job_applications
        WHERE user_id = $1 AND interview_time IS NOT NULL AND interview_time > NOW()
        ORDER BY interview_time ASC
        LIMIT $2`

	rows, err := r.queryRows(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming interviews: %w", err)
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var (
			id            int
			companyName   string
			positionTitle string
			interviewTime time.Time
			interviewType sql.NullString
		)
		if err := rows.Scan(&id, &companyName, &positionTitle, &interviewTime, &interviewType); err != nil {
			return nil, fmt.Errorf("failed to scan upcoming interview: %w", err)
		}
		entry := map[string]interface{}{
			"id":             id,
			"company_name":   companyName,
			"position_title": positionTitle,
			"interview_time": interviewTime,
		}
		if interviewType.Valid {
			entry["interview_type"] = interviewType.String
		}
		result = append(result, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate upcoming interviews: %w", err)
	}
	return result, nil
}

func (r *jobAppRepo) ListDailyStats(userID uint, days int) ([]map[string]interface{}, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}
	if days <= 0 {
		days = 30
	}

	query := `
        SELECT DATE(created_at) AS date, COUNT(*) AS count
        FROM job_applications
        WHERE user_id = $1 AND created_at >= CURRENT_DATE - INTERVAL $2
        GROUP BY DATE(created_at)
        ORDER BY date DESC`

	interval := fmt.Sprintf("'%d days'", days)
	rows, err := r.queryRows(query, userID, interval)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily stats: %w", err)
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var date time.Time
		var count int
		if err := rows.Scan(&date, &count); err != nil {
			return nil, fmt.Errorf("failed to scan daily stats: %w", err)
		}
		result = append(result, map[string]interface{}{
			"date":  date.Format("2006-01-02"),
			"count": count,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate daily stats: %w", err)
	}
	return result, nil
}
