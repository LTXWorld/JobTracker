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
}

type jobAppRepo struct{ db *database.DB }

func NewJobApplicationRepository(db *database.DB) JobApplicationRepository { return &jobAppRepo{db: db} }

func (r *jobAppRepo) Create(userID uint, req *model.CreateJobApplicationRequest) (*model.JobApplication, error) {
    if r.db.ORM == nil { return nil, fmt.Errorf("gorm not initialized") }
    applicationDate := req.ApplicationDate
    if applicationDate == "" { applicationDate = time.Now().Format("2006-01-02") }
    status := req.Status
    if status == "" { status = model.StatusApplied }
    reminderEnabled := false
    if req.ReminderEnabled != nil { reminderEnabled = *req.ReminderEnabled }

    query := `INSERT INTO job_applications (
        user_id, company_name, position_title, application_date, status,
        job_description, salary_range, work_location, contact_info, notes,
        interview_time, reminder_time, reminder_enabled, follow_up_date,
        hr_name, hr_phone, hr_email, interview_location, interview_type
    ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
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
    ).Row()
    if err := row.Scan(&job.ID, &job.CreatedAt, &job.UpdatedAt); err != nil { return nil, fmt.Errorf("failed to create job application: %w", err) }

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

func (r *jobAppRepo) GetByID(userID uint, id int) (*model.JobApplication, error) {
    if r.db.ORM == nil { return nil, fmt.Errorf("gorm not initialized") }
    query := `SELECT id, user_id, company_name, position_title, application_date, status,
        job_description, salary_range, work_location, contact_info, notes,
        interview_time, reminder_time, reminder_enabled, follow_up_date,
        hr_name, hr_phone, hr_email, interview_location, interview_type,
        created_at, updated_at FROM job_applications WHERE id=$1 AND user_id=$2`
    var job model.JobApplication
    row := r.db.ORM.Raw(query, id, userID).Row()
    if err := row.Scan(
        &job.ID,&job.UserID,&job.CompanyName,&job.PositionTitle,&job.ApplicationDate,&job.Status,
        &job.JobDescription,&job.SalaryRange,&job.WorkLocation,&job.ContactInfo,&job.Notes,
        &job.InterviewTime,&job.ReminderTime,&job.ReminderEnabled,&job.FollowUpDate,
        &job.HRName,&job.HRPhone,&job.HREmail,&job.InterviewLocation,&job.InterviewType,
        &job.CreatedAt,&job.UpdatedAt,
    ); err != nil {
        if err == sql.ErrNoRows { return nil, fmt.Errorf("job application not found") }
        return nil, fmt.Errorf("failed to get job application: %w", err)
    }
    return &job, nil
}

func (r *jobAppRepo) GetAll(userID uint) ([]model.JobApplication, error) {
    if r.db.ORM == nil { return nil, fmt.Errorf("gorm not initialized") }
    query := `SELECT id, user_id, company_name, position_title, application_date, status,
        job_description, salary_range, work_location, contact_info, notes,
        interview_time, reminder_time, reminder_enabled, follow_up_date,
        hr_name, hr_phone, hr_email, interview_location, interview_type,
        created_at, updated_at FROM job_applications WHERE user_id = $1
        ORDER BY application_date DESC, created_at DESC LIMIT 500`
    rows, err := r.db.ORM.Raw(query, userID).Rows()
    if err != nil { return nil, fmt.Errorf("failed to get job applications: %w", err) }
    defer rows.Close()
    var list []model.JobApplication
    for rows.Next() {
        var job model.JobApplication
        if err := rows.Scan(&job.ID,&job.UserID,&job.CompanyName,&job.PositionTitle,&job.ApplicationDate,&job.Status,
            &job.JobDescription,&job.SalaryRange,&job.WorkLocation,&job.ContactInfo,&job.Notes,
            &job.InterviewTime,&job.ReminderTime,&job.ReminderEnabled,&job.FollowUpDate,
            &job.HRName,&job.HRPhone,&job.HREmail,&job.InterviewLocation,&job.InterviewType,
            &job.CreatedAt,&job.UpdatedAt); err != nil { return nil, fmt.Errorf("failed to scan job application: %w", err) }
        list = append(list, job)
    }
    return list, nil
}

func (r *jobAppRepo) GetAllPaginated(userID uint, req model.PaginationRequest) (*model.PaginationResponse, error) {
    if r.db.ORM == nil { return nil, fmt.Errorf("gorm not initialized") }
    req.ValidateAndSetDefaults()
    where := "WHERE user_id = $1"
    args := []interface{}{userID}
    idx := 2
    if req.Status != nil { where += fmt.Sprintf(" AND status = $%d", idx); args = append(args, *req.Status); idx++ }

    var total int64
    countSQL := fmt.Sprintf("SELECT COUNT(*) FROM job_applications %s", where)
    if err := r.db.ORM.Raw(countSQL, args...).Row().Scan(&total); err != nil { return nil, fmt.Errorf("failed to count job applications: %w", err) }
    if total == 0 { return &model.PaginationResponse{Data: []model.JobApplication{}, Total: 0, Page: req.Page, PageSize: req.PageSize}, nil }

    allowed := map[string]bool{"application_date":true,"created_at":true,"updated_at":true,"company_name":true,"position_title":true,"status":true}
    if !allowed[req.SortBy] { req.SortBy = "application_date" }
    dataSQL := fmt.Sprintf(`SELECT id, user_id, company_name, position_title, application_date, status,
        job_description, salary_range, work_location, contact_info, notes,
        interview_time, reminder_time, reminder_enabled, follow_up_date,
        hr_name, hr_phone, hr_email, interview_location, interview_type,
        created_at, updated_at FROM job_applications %s ORDER BY %s %s, created_at DESC LIMIT $%d OFFSET $%d`,
        where, req.SortBy, req.SortDir, idx, idx+1)
    args = append(args, req.PageSize, req.GetOffset())
    rows, err := r.db.ORM.Raw(dataSQL, args...).Rows()
    if err != nil { return nil, fmt.Errorf("failed to get job applications: %w", err) }
    defer rows.Close()
    var jobs []model.JobApplication
    for rows.Next() {
        var job model.JobApplication
        if err := rows.Scan(&job.ID,&job.UserID,&job.CompanyName,&job.PositionTitle,&job.ApplicationDate,&job.Status,
            &job.JobDescription,&job.SalaryRange,&job.WorkLocation,&job.ContactInfo,&job.Notes,
            &job.InterviewTime,&job.ReminderTime,&job.ReminderEnabled,&job.FollowUpDate,
            &job.HRName,&job.HRPhone,&job.HREmail,&job.InterviewLocation,&job.InterviewType,
            &job.CreatedAt,&job.UpdatedAt); err != nil { return nil, fmt.Errorf("failed to scan job application: %w", err) }
        jobs = append(jobs, job)
    }
    totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
    return &model.PaginationResponse{Data: jobs, Total: total, Page: req.Page, PageSize: req.PageSize, TotalPages: totalPages, HasNext: req.Page < totalPages, HasPrev: req.Page > 1}, nil
}

func (r *jobAppRepo) Update(userID uint, id int, req *model.UpdateJobApplicationRequest) (*model.JobApplication, error) {
    if r.db.ORM == nil { return nil, fmt.Errorf("gorm not initialized") }
    setParts := []string{}
    args := []interface{}{}
    idx := 1
    if req.CompanyName != nil { setParts = append(setParts, fmt.Sprintf("company_name=$%d", idx)); args = append(args, *req.CompanyName); idx++ }
    if req.PositionTitle != nil { setParts = append(setParts, fmt.Sprintf("position_title=$%d", idx)); args = append(args, *req.PositionTitle); idx++ }
    if req.ApplicationDate != nil { setParts = append(setParts, fmt.Sprintf("application_date=$%d", idx)); args = append(args, *req.ApplicationDate); idx++ }
    if req.Status != nil { setParts = append(setParts, fmt.Sprintf("status=$%d", idx)); args = append(args, *req.Status); idx++ }
    if req.JobDescription != nil { setParts = append(setParts, fmt.Sprintf("job_description=$%d", idx)); args = append(args, *req.JobDescription); idx++ }
    if req.SalaryRange != nil { setParts = append(setParts, fmt.Sprintf("salary_range=$%d", idx)); args = append(args, *req.SalaryRange); idx++ }
    if req.WorkLocation != nil { setParts = append(setParts, fmt.Sprintf("work_location=$%d", idx)); args = append(args, *req.WorkLocation); idx++ }
    if req.ContactInfo != nil { setParts = append(setParts, fmt.Sprintf("contact_info=$%d", idx)); args = append(args, *req.ContactInfo); idx++ }
    if req.Notes != nil { setParts = append(setParts, fmt.Sprintf("notes=$%d", idx)); args = append(args, *req.Notes); idx++ }
    if req.InterviewTime != nil { setParts = append(setParts, fmt.Sprintf("interview_time=$%d", idx)); args = append(args, *req.InterviewTime); idx++ }
    if req.ReminderTime != nil { setParts = append(setParts, fmt.Sprintf("reminder_time=$%d", idx)); args = append(args, *req.ReminderTime); idx++ }
    if req.ReminderEnabled != nil { setParts = append(setParts, fmt.Sprintf("reminder_enabled=$%d", idx)); args = append(args, *req.ReminderEnabled); idx++ }
    if req.FollowUpDate != nil { setParts = append(setParts, fmt.Sprintf("follow_up_date=$%d", idx)); args = append(args, *req.FollowUpDate); idx++ }
    if req.HRName != nil { setParts = append(setParts, fmt.Sprintf("hr_name=$%d", idx)); args = append(args, *req.HRName); idx++ }
    if req.HRPhone != nil { setParts = append(setParts, fmt.Sprintf("hr_phone=$%d", idx)); args = append(args, *req.HRPhone); idx++ }
    if req.HREmail != nil { setParts = append(setParts, fmt.Sprintf("hr_email=$%d", idx)); args = append(args, *req.HREmail); idx++ }
    if req.InterviewLocation != nil { setParts = append(setParts, fmt.Sprintf("interview_location=$%d", idx)); args = append(args, *req.InterviewLocation); idx++ }
    if req.InterviewType != nil { setParts = append(setParts, fmt.Sprintf("interview_type=$%d", idx)); args = append(args, *req.InterviewType); idx++ }
    if len(setParts) == 0 { return r.GetByID(userID, id) }
    setParts = append(setParts, fmt.Sprintf("updated_at=$%d", idx)); args = append(args, time.Now()); idx++
    args = append(args, id, userID)
    query := fmt.Sprintf(`UPDATE job_applications SET %s WHERE id=$%d AND user_id=$%d RETURNING id, user_id, company_name, position_title, application_date, status,
        job_description, salary_range, work_location, contact_info, notes,
        interview_time, reminder_time, reminder_enabled, follow_up_date,
        hr_name, hr_phone, hr_email, interview_location, interview_type,
        created_at, updated_at`, strings.Join(setParts, ", "), idx, idx+1)
    var job model.JobApplication
    row := r.db.ORM.Raw(query, args...).Row()
    if err := row.Scan(&job.ID,&job.UserID,&job.CompanyName,&job.PositionTitle,&job.ApplicationDate,&job.Status,&job.JobDescription,&job.SalaryRange,&job.WorkLocation,&job.ContactInfo,&job.Notes,&job.InterviewTime,&job.ReminderTime,&job.ReminderEnabled,&job.FollowUpDate,&job.HRName,&job.HRPhone,&job.HREmail,&job.InterviewLocation,&job.InterviewType,&job.CreatedAt,&job.UpdatedAt); err != nil {
        if err == sql.ErrNoRows { return nil, fmt.Errorf("job application not found") }
        return nil, fmt.Errorf("failed to update job application: %w", err)
    }
    return &job, nil
}

func (r *jobAppRepo) Delete(userID uint, id int) error {
    if r.db.ORM == nil { return fmt.Errorf("gorm not initialized") }
    res := r.db.ORM.Exec("DELETE FROM job_applications WHERE id = $1 AND user_id = $2", id, userID)
    if res.Error != nil { return fmt.Errorf("failed to delete job application: %w", res.Error) }
    if res.RowsAffected == 0 { return fmt.Errorf("job application not found") }
    return nil
}

