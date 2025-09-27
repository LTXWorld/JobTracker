package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"jobView-backend/internal/database"
	"jobView-backend/internal/model"
)

type ExportRepository interface {
	CreateTask(ctx context.Context, task *model.ExportTask) error
	UpdateTask(ctx context.Context, task *model.ExportTask) error
	GetTaskStatus(ctx context.Context, taskID string, userID uint) (*ExportTaskStatusRecord, error)
	GetDownloadInfo(ctx context.Context, taskID string, userID uint) (*ExportDownloadRecord, error)
	CountHistory(ctx context.Context, userID uint) (int64, error)
	ListHistory(ctx context.Context, userID uint, limit, offset int) ([]ExportHistoryRow, error)
	CountDailyExports(ctx context.Context, userID uint, day time.Time) (int, error)
	CountActiveExports(ctx context.Context, userID uint) (int, error)
	CountExportData(ctx context.Context, userID uint, filters *model.ExportFilters) (int, error)
	FetchExportData(ctx context.Context, userID uint, filters *model.ExportFilters, offset, limit int) ([]model.JobApplication, error)
	ListExpiredTasks(ctx context.Context) ([]ExportExpiredTask, error)
	DeleteTask(ctx context.Context, taskID string) error
}

type exportRepository struct {
	db *database.DB
}

func NewExportRepository(db *database.DB) ExportRepository {
	return &exportRepository{db: db}
}

type ExportTaskStatusRecord struct {
	TaskID           string
	Status           model.TaskStatus
	Progress         int
	ProcessedRecords int
	TotalRecords     *int
	FileSize         *int64
	ExpiresAt        *time.Time
	ErrorMessage     *string
	CreatedAt        time.Time
	CompletedAt      *time.Time
	Filename         *string
}

type ExportDownloadRecord struct {
	FilePath string
	Filename string
	Status   model.TaskStatus
	Expires  *time.Time
}

type ExportHistoryRow struct {
	TaskID       string
	CreatedAt    time.Time
	Status       model.TaskStatus
	Filename     *string
	FileSize     *int64
	TotalRecords *int
	ExpiresAt    *time.Time
}

type ExportExpiredTask struct {
	TaskID   string
	FilePath *string
}

func (r *exportRepository) baseDB() (*database.DB, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return r.db, nil
}

func (r *exportRepository) CreateTask(ctx context.Context, task *model.ExportTask) error {
	dbWrapper, err := r.baseDB()
	if err != nil {
		return err
	}
	query := `
        INSERT INTO export_tasks (
            task_id, user_id, status, export_type, total_records,
            processed_records, progress, filters, options, created_at, expires_at
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	args := []interface{}{task.TaskID, task.UserID, task.Status, task.ExportType, task.TotalRecords, task.ProcessedRecords, task.Progress, task.Filters, task.Options, task.CreatedAt, task.ExpiresAt}
	if dbWrapper.ORM != nil {
		return dbWrapper.ORM.WithContext(ctx).Exec(query, args...).Error
	}
	_, err = dbWrapper.ExecContext(ctx, query, args...)
	return err
}

func (r *exportRepository) UpdateTask(ctx context.Context, task *model.ExportTask) error {
	dbWrapper, err := r.baseDB()
	if err != nil {
		return err
	}
	query := `
        UPDATE export_tasks SET 
            status = $2, processed_records = $3, progress = $4,
            file_path = $5, file_size = $6, filename = $7,
            error_message = $8, started_at = $9, completed_at = $10
        WHERE task_id = $1`
	args := []interface{}{task.TaskID, task.Status, task.ProcessedRecords, task.Progress, task.FilePath, task.FileSize, task.Filename, task.ErrorMessage, task.StartedAt, task.CompletedAt}
	if dbWrapper.ORM != nil {
		return dbWrapper.ORM.WithContext(ctx).Exec(query, args...).Error
	}
	_, err = dbWrapper.ExecContext(ctx, query, args...)
	return err
}

func (r *exportRepository) GetTaskStatus(ctx context.Context, taskID string, userID uint) (*ExportTaskStatusRecord, error) {
	dbWrapper, err := r.baseDB()
	if err != nil {
		return nil, err
	}
	query := `SELECT task_id, status, progress, processed_records, total_records, file_size, expires_at, error_message, created_at, completed_at, filename FROM export_tasks WHERE task_id=$1 AND user_id=$2`
	var (
		totalRecords sql.NullInt64
		fileSize     sql.NullInt64
		expiresAt    sql.NullTime
		errorMsg     sql.NullString
		completedAt  sql.NullTime
		filename     sql.NullString
		record       ExportTaskStatusRecord
	)
	record.TaskID = taskID
	if dbWrapper.ORM != nil {
		row := dbWrapper.ORM.WithContext(ctx).Raw(query, taskID, userID).Row()
		err = row.Scan(&record.TaskID, &record.Status, &record.Progress, &record.ProcessedRecords, &totalRecords, &fileSize, &expiresAt, &errorMsg, &record.CreatedAt, &completedAt, &filename)
	} else {
		err = dbWrapper.QueryRowContext(ctx, query, taskID, userID).Scan(&record.TaskID, &record.Status, &record.Progress, &record.ProcessedRecords, &totalRecords, &fileSize, &expiresAt, &errorMsg, &record.CreatedAt, &completedAt, &filename)
	}
	if err != nil {
		return nil, err
	}
	if totalRecords.Valid {
		v := int(totalRecords.Int64)
		record.TotalRecords = &v
	}
	if fileSize.Valid {
		v := fileSize.Int64
		record.FileSize = &v
	}
	if expiresAt.Valid {
		t := expiresAt.Time
		record.ExpiresAt = &t
	}
	if errorMsg.Valid {
		msg := errorMsg.String
		record.ErrorMessage = &msg
	}
	if completedAt.Valid {
		t := completedAt.Time
		record.CompletedAt = &t
	}
	if filename.Valid {
		name := filename.String
		record.Filename = &name
	}
	return &record, nil
}

func (r *exportRepository) GetDownloadInfo(ctx context.Context, taskID string, userID uint) (*ExportDownloadRecord, error) {
	dbWrapper, err := r.baseDB()
	if err != nil {
		return nil, err
	}
	query := `SELECT file_path, filename, status, expires_at FROM export_tasks WHERE task_id=$1 AND user_id=$2`
	var filePath, filename sql.NullString
	var status model.TaskStatus
	var expiresAt sql.NullTime
	if dbWrapper.ORM != nil {
		row := dbWrapper.ORM.WithContext(ctx).Raw(query, taskID, userID).Row()
		err = row.Scan(&filePath, &filename, &status, &expiresAt)
	} else {
		err = dbWrapper.QueryRowContext(ctx, query, taskID, userID).Scan(&filePath, &filename, &status, &expiresAt)
	}
	if err != nil {
		return nil, err
	}
	if !filePath.Valid || !filename.Valid {
		return nil, fmt.Errorf("invalid file data")
	}
	record := &ExportDownloadRecord{FilePath: filePath.String, Filename: filename.String, Status: status}
	if expiresAt.Valid {
		t := expiresAt.Time
		record.Expires = &t
	}
	return record, nil
}

func (r *exportRepository) CountHistory(ctx context.Context, userID uint) (int64, error) {
	dbWrapper, err := r.baseDB()
	if err != nil {
		return 0, err
	}
	query := `SELECT COUNT(*) FROM export_tasks WHERE user_id = $1`
	var total int64
	if dbWrapper.ORM != nil {
		err = dbWrapper.ORM.WithContext(ctx).Raw(query, userID).Row().Scan(&total)
	} else {
		err = dbWrapper.QueryRowContext(ctx, query, userID).Scan(&total)
	}
	return total, err
}

func (r *exportRepository) ListHistory(ctx context.Context, userID uint, limit, offset int) ([]ExportHistoryRow, error) {
	dbWrapper, err := r.baseDB()
	if err != nil {
		return nil, err
	}
	query := `
        SELECT task_id, created_at, status, filename, file_size, total_records, expires_at
        FROM export_tasks
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3`
	var rows *sql.Rows
	if dbWrapper.ORM != nil {
		rows, err = dbWrapper.ORM.WithContext(ctx).Raw(query, userID, limit, offset).Rows()
	} else {
		rows, err = dbWrapper.QueryContext(ctx, query, userID, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ExportHistoryRow
	for rows.Next() {
		var (
			rec      ExportHistoryRow
			filename sql.NullString
			fileSize sql.NullInt64
			total    sql.NullInt64
			expires  sql.NullTime
		)
		if err := rows.Scan(&rec.TaskID, &rec.CreatedAt, &rec.Status, &filename, &fileSize, &total, &expires); err != nil {
			return nil, err
		}
		if filename.Valid {
			name := filename.String
			rec.Filename = &name
		}
		if fileSize.Valid {
			size := fileSize.Int64
			rec.FileSize = &size
		}
		if total.Valid {
			v := int(total.Int64)
			rec.TotalRecords = &v
		}
		if expires.Valid {
			t := expires.Time
			rec.ExpiresAt = &t
		}
		result = append(result, rec)
	}
	return result, rows.Err()
}

func (r *exportRepository) CountDailyExports(ctx context.Context, userID uint, day time.Time) (int, error) {
	dbWrapper, err := r.baseDB()
	if err != nil {
		return 0, err
	}
	query := `SELECT COUNT(*) FROM export_tasks WHERE user_id = $1 AND DATE(created_at) = $2`
	formatted := day.Format("2006-01-02")
	var count int
	if dbWrapper.ORM != nil {
		err = dbWrapper.ORM.WithContext(ctx).Raw(query, userID, formatted).Row().Scan(&count)
	} else {
		err = dbWrapper.QueryRowContext(ctx, query, userID, formatted).Scan(&count)
	}
	return count, err
}

func (r *exportRepository) CountActiveExports(ctx context.Context, userID uint) (int, error) {
	dbWrapper, err := r.baseDB()
	if err != nil {
		return 0, err
	}
	query := `SELECT COUNT(*) FROM export_tasks WHERE user_id = $1 AND status = $2`
	var count int
	if dbWrapper.ORM != nil {
		err = dbWrapper.ORM.WithContext(ctx).Raw(query, userID, model.TaskStatusProcessing).Row().Scan(&count)
	} else {
		err = dbWrapper.QueryRowContext(ctx, query, userID, model.TaskStatusProcessing).Scan(&count)
	}
	return count, err
}

func (r *exportRepository) CountExportData(ctx context.Context, userID uint, filters *model.ExportFilters) (int, error) {
	query, args := buildExportCountQuery(userID, filters)
	dbWrapper, err := r.baseDB()
	if err != nil {
		return 0, err
	}
	var count int
	if dbWrapper.ORM != nil {
		err = dbWrapper.ORM.WithContext(ctx).Raw(query, args...).Row().Scan(&count)
	} else {
		err = dbWrapper.QueryRowContext(ctx, query, args...).Scan(&count)
	}
	return count, err
}

func (r *exportRepository) FetchExportData(ctx context.Context, userID uint, filters *model.ExportFilters, offset, limit int) ([]model.JobApplication, error) {
	query, args := buildExportDataQuery(userID, filters, offset, limit)
	dbWrapper, err := r.baseDB()
	if err != nil {
		return nil, err
	}
	var rows *sql.Rows
	if dbWrapper.ORM != nil {
		rows, err = dbWrapper.ORM.WithContext(ctx).Raw(query, args...).Rows()
	} else {
		rows, err = dbWrapper.QueryContext(ctx, query, args...)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.JobApplication
	for rows.Next() {
		var app model.JobApplication
		if err := rows.Scan(
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
			&app.CompanyAttribute,
			&app.CreatedAt,
			&app.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, app)
	}
	return list, rows.Err()
}

func (r *exportRepository) ListExpiredTasks(ctx context.Context) ([]ExportExpiredTask, error) {
	dbWrapper, err := r.baseDB()
	if err != nil {
		return nil, err
	}
	query := `SELECT task_id, file_path FROM export_tasks WHERE expires_at < NOW() AND status = 'completed'`
	var rows *sql.Rows
	if dbWrapper.ORM != nil {
		rows, err = dbWrapper.ORM.WithContext(ctx).Raw(query).Rows()
	} else {
		rows, err = dbWrapper.QueryContext(ctx, query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []ExportExpiredTask
	for rows.Next() {
		var task ExportExpiredTask
		var filePath sql.NullString
		if err := rows.Scan(&task.TaskID, &filePath); err != nil {
			return nil, err
		}
		if filePath.Valid {
			path := filePath.String
			task.FilePath = &path
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (r *exportRepository) DeleteTask(ctx context.Context, taskID string) error {
	dbWrapper, err := r.baseDB()
	if err != nil {
		return err
	}
	query := `DELETE FROM export_tasks WHERE task_id = $1`
	if dbWrapper.ORM != nil {
		return dbWrapper.ORM.WithContext(ctx).Exec(query, taskID).Error
	}
	_, err = dbWrapper.ExecContext(ctx, query, taskID)
	return err
}

func buildExportCountQuery(userID uint, filters *model.ExportFilters) (string, []interface{}) {
	query := "SELECT COUNT(*) FROM job_applications WHERE user_id = $1"
	args := []interface{}{userID}
	argIndex := 2

	if filters != nil {
		if len(filters.Status) > 0 {
			placeholders := make([]string, len(filters.Status))
			for i, status := range filters.Status {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, status)
				argIndex++
			}
			query += " AND status IN (" + strings.Join(placeholders, ",") + ")"
		}
		if filters.DateRange != nil {
			query += fmt.Sprintf(" AND application_date >= $%d AND application_date <= $%d", argIndex, argIndex+1)
			args = append(args, filters.DateRange.Start, filters.DateRange.End)
			argIndex += 2
		}
		if len(filters.CompanyNames) > 0 {
			placeholders := make([]string, len(filters.CompanyNames))
			for i, company := range filters.CompanyNames {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, company)
				argIndex++
			}
			query += " AND company_name IN (" + strings.Join(placeholders, ",") + ")"
		}
		if strings.TrimSpace(filters.Keywords) != "" {
			query += fmt.Sprintf(" AND (company_name ILIKE $%d OR position_title ILIKE $%d OR notes ILIKE $%d)", argIndex, argIndex, argIndex)
			keyword := "%" + filters.Keywords + "%"
			args = append(args, keyword)
		}
	}
	return query, args
}

func buildExportDataQuery(userID uint, filters *model.ExportFilters, offset, limit int) (string, []interface{}) {
	query := `
        SELECT id, user_id, company_name, position_title, application_date, status,
               job_description, salary_range, work_location, contact_info, notes,
               interview_time, reminder_time, reminder_enabled, follow_up_date,
               hr_name, hr_phone, hr_email, interview_location, interview_type,
               company_attribute,
               created_at, updated_at
        FROM job_applications
        WHERE user_id = $1`
	args := []interface{}{userID}
	argIndex := 2

	if filters != nil {
		if len(filters.Status) > 0 {
			placeholders := make([]string, len(filters.Status))
			for i, status := range filters.Status {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, status)
				argIndex++
			}
			query += " AND status IN (" + strings.Join(placeholders, ",") + ")"
		}
		if filters.DateRange != nil {
			query += fmt.Sprintf(" AND application_date >= $%d AND application_date <= $%d", argIndex, argIndex+1)
			args = append(args, filters.DateRange.Start, filters.DateRange.End)
			argIndex += 2
		}
		if len(filters.CompanyNames) > 0 {
			placeholders := make([]string, len(filters.CompanyNames))
			for i, company := range filters.CompanyNames {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, company)
				argIndex++
			}
			query += " AND company_name IN (" + strings.Join(placeholders, ",") + ")"
		}
		if strings.TrimSpace(filters.Keywords) != "" {
			query += fmt.Sprintf(" AND (company_name ILIKE $%d OR position_title ILIKE $%d OR notes ILIKE $%d)", argIndex, argIndex, argIndex)
			keyword := "%" + filters.Keywords + "%"
			args = append(args, keyword)
		}
	}

	query += " ORDER BY application_date DESC, created_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, limit, offset)
	}
	return query, args
}
