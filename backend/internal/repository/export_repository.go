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

// ExportRepository 导出相关的数据访问
type ExportRepository interface {
    SaveTask(task *model.ExportTask) error
    UpdateTask(task *model.ExportTask) error
    GetTaskStatus(taskID string, userID uint) (*model.TaskStatusResponse, error)
    GetFileMeta(taskID string, userID uint) (filePath string, filename string, status model.TaskStatus, expiresAt *time.Time, err error)
    GetExportHistory(userID uint, page, limit int) ([]model.ExportHistoryItem, int64, error)
    CleanupExpiredTasks() ([]struct{ TaskID string; FilePath sql.NullString }, error)
    DeleteTaskByID(taskID string) error

    GetExportDataCount(userID uint, filters *model.ExportFilters) (int, error)
    GetExportData(userID uint, filters *model.ExportFilters, offset, limit int) ([]model.JobApplication, error)
}

type exportRepo struct{ db *database.DB }

func NewExportRepository(db *database.DB) ExportRepository { return &exportRepo{db: db} }

// --- Export tasks ---

func (r *exportRepo) SaveTask(task *model.ExportTask) error {
    if r.db == nil || r.db.ORM == nil { return fmt.Errorf("gorm not initialized") }
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    q := `INSERT INTO export_tasks (
            task_id, user_id, status, export_type, total_records,
            processed_records, progress, filters, options, created_at, expires_at
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
    res := r.db.ORM.WithContext(ctx).Exec(q, task.TaskID, task.UserID, task.Status, task.ExportType, task.TotalRecords, task.ProcessedRecords, task.Progress, task.Filters, task.Options, task.CreatedAt, task.ExpiresAt)
    return res.Error
}

func (r *exportRepo) UpdateTask(task *model.ExportTask) error {
    if r.db == nil || r.db.ORM == nil { return fmt.Errorf("gorm not initialized") }
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    q := `UPDATE export_tasks SET 
            status=$2, processed_records=$3, progress=$4,
            file_path=$5, file_size=$6, filename=$7,
            error_message=$8, started_at=$9, completed_at=$10
          WHERE task_id=$1`
    return r.db.ORM.WithContext(ctx).Exec(q, task.TaskID, task.Status, task.ProcessedRecords, task.Progress, task.FilePath, task.FileSize, task.Filename, task.ErrorMessage, task.StartedAt, task.CompletedAt).Error
}

func (r *exportRepo) GetTaskStatus(taskID string, userID uint) (*model.TaskStatusResponse, error) {
    if r.db == nil || r.db.ORM == nil { return nil, fmt.Errorf("gorm not initialized") }
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    q := `SELECT task_id, status, progress, processed_records, total_records,
                 file_size, expires_at, error_message, created_at, completed_at, filename
          FROM export_tasks WHERE task_id=$1 AND user_id=$2`
    var resp model.TaskStatusResponse
    var fileSize sql.NullInt64
    var expiresAt, completedAt sql.NullTime
    var errorMessage, filename sql.NullString
    var totalRecords sql.NullInt32
    row := r.db.ORM.WithContext(ctx).Raw(q, taskID, userID).Row()
    if err := row.Scan(&resp.TaskID, &resp.Status, &resp.Progress, &resp.ProcessedRecords, &totalRecords, &fileSize, &expiresAt, &errorMessage, &resp.CreatedAt, &completedAt, &filename); err != nil {
        if err == sql.ErrNoRows { return nil, fmt.Errorf("导出任务不存在或无访问权限") }
        return nil, err
    }
    if totalRecords.Valid { tr := int(totalRecords.Int32); resp.TotalRecords=&tr }
    if fileSize.Valid { fs := formatFileSize(fileSize.Int64); resp.FileSize=&fs }
    if expiresAt.Valid { resp.ExpiresAt=&expiresAt.Time }
    if completedAt.Valid { resp.CompletedAt=&completedAt.Time }
    if errorMessage.Valid { resp.ErrorMessage=&errorMessage.String }
    if resp.Status == model.TaskStatusCompleted { d := fmt.Sprintf("/api/v1/export/download/%s", taskID); resp.DownloadURL=&d }
    return &resp, nil
}

func (r *exportRepo) GetFileMeta(taskID string, userID uint) (string, string, model.TaskStatus, *time.Time, error) {
    if r.db == nil || r.db.ORM == nil { return "", "", "", nil, fmt.Errorf("gorm not initialized") }
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    q := `SELECT file_path, filename, status, expires_at FROM export_tasks WHERE task_id=$1 AND user_id=$2`
    var filePath, filename sql.NullString
    var status model.TaskStatus
    var expiresAt sql.NullTime
    row := r.db.ORM.WithContext(ctx).Raw(q, taskID, userID).Row()
    if err := row.Scan(&filePath, &filename, &status, &expiresAt); err != nil {
        if err == sql.ErrNoRows { return "", "", "", nil, fmt.Errorf("文件不存在或无访问权限") }
        return "", "", "", nil, err
    }
    var exp *time.Time
    if expiresAt.Valid { exp = &expiresAt.Time }
    return filePath.String, filename.String, status, exp, nil
}

func (r *exportRepo) GetExportHistory(userID uint, page, limit int) ([]model.ExportHistoryItem, int64, error) {
    if r.db == nil || r.db.ORM == nil { return nil, 0, fmt.Errorf("gorm not initialized") }
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    if limit <= 0 { limit = 10 }
    if limit > 50 { limit = 50 }
    if page <= 0 { page = 1 }
    offset := (page - 1) * limit

    var total int64
    if err := r.db.ORM.WithContext(ctx).Raw("SELECT COUNT(*) FROM export_tasks WHERE user_id=$1", userID).Row().Scan(&total); err != nil { return nil, 0, err }

    q := `SELECT task_id, created_at, status, filename, file_size, total_records, expires_at
          FROM export_tasks WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
    rows, err := r.db.ORM.WithContext(ctx).Raw(q, userID, limit, offset).Rows()
    if err != nil { return nil, 0, err }
    defer rows.Close()
    var items []model.ExportHistoryItem
    for rows.Next() {
        var it model.ExportHistoryItem
        var filename sql.NullString
        var fileSize sql.NullInt64
        var totalRecords sql.NullInt32
        var expiresAt sql.NullTime
        if err := rows.Scan(&it.TaskID, &it.CreatedAt, &it.Status, &filename, &fileSize, &totalRecords, &expiresAt); err != nil { continue }
        if filename.Valid { it.Filename=&filename.String }
        if fileSize.Valid { fs := formatFileSize(fileSize.Int64); it.FileSize=&fs }
        if totalRecords.Valid { c := int(totalRecords.Int32); it.RecordCount=&c }
        if expiresAt.Valid { it.ExpiresAt=&expiresAt.Time }
        if it.Status == model.TaskStatusCompleted && (it.ExpiresAt==nil || time.Now().Before(*it.ExpiresAt)) { d := fmt.Sprintf("/api/v1/export/download/%s", it.TaskID); it.DownloadURL=&d }
        items = append(items, it)
    }
    return items, total, nil
}

func (r *exportRepo) CleanupExpiredTasks() ([]struct{ TaskID string; FilePath sql.NullString }, error) {
    if r.db == nil || r.db.ORM == nil { return nil, fmt.Errorf("gorm not initialized") }
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    q := `SELECT task_id, file_path FROM export_tasks WHERE expires_at < NOW() AND status = 'completed'`
    rows, err := r.db.ORM.WithContext(ctx).Raw(q).Rows()
    if err != nil { return nil, err }
    defer rows.Close()
    var list []struct{ TaskID string; FilePath sql.NullString }
    for rows.Next() {
        var t struct{ TaskID string; FilePath sql.NullString }
        if err := rows.Scan(&t.TaskID, &t.FilePath); err != nil { continue }
        list = append(list, t)
    }
    return list, nil
}

func (r *exportRepo) DeleteTaskByID(taskID string) error {
    if r.db == nil || r.db.ORM == nil { return fmt.Errorf("gorm not initialized") }
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    return r.db.ORM.WithContext(ctx).Exec("DELETE FROM export_tasks WHERE task_id=$1", taskID).Error
}

// --- Export data queries ---

func (r *exportRepo) GetExportDataCount(userID uint, filters *model.ExportFilters) (int, error) {
    if r.db == nil || r.db.ORM == nil { return 0, fmt.Errorf("gorm not initialized") }
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    query, args := buildCountQueryInternal(userID, filters)
    var count int
    if err := r.db.ORM.WithContext(ctx).Raw(query, args...).Row().Scan(&count); err != nil { return 0, err }
    return count, nil
}

func (r *exportRepo) GetExportData(userID uint, filters *model.ExportFilters, offset, limit int) ([]model.JobApplication, error) {
    if r.db == nil || r.db.ORM == nil { return nil, fmt.Errorf("gorm not initialized") }
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    query, args := buildDataQueryInternal(userID, filters, offset, limit)
    rows, err := r.db.ORM.WithContext(ctx).Raw(query, args...).Rows()
    if err != nil { return nil, err }
    defer rows.Close()
    var apps []model.JobApplication
    for rows.Next() {
        var a model.JobApplication
        if err := rows.Scan(&a.ID,&a.UserID,&a.CompanyName,&a.PositionTitle,&a.ApplicationDate,&a.Status,&a.JobDescription,&a.SalaryRange,&a.WorkLocation,&a.ContactInfo,&a.Notes,&a.InterviewTime,&a.ReminderTime,&a.ReminderEnabled,&a.FollowUpDate,&a.HRName,&a.HRPhone,&a.HREmail,&a.InterviewLocation,&a.InterviewType,&a.CreatedAt,&a.UpdatedAt); err != nil { return nil, err }
        apps = append(apps, a)
    }
    return apps, nil
}

// 内部构造 SQL 与参数（与服务层保持一致逻辑）
func buildCountQueryInternal(userID uint, filters *model.ExportFilters) (string, []interface{}) {
    query := "SELECT COUNT(*) FROM job_applications WHERE user_id = $1"
    args := []interface{}{userID}
    argIndex := 2
    if len(filters.Status) > 0 {
        placeholders := make([]string, len(filters.Status))
        for i, st := range filters.Status { placeholders[i] = fmt.Sprintf("$%d", argIndex); args = append(args, st); argIndex++ }
        query += " AND status IN (" + strings.Join(placeholders, ",") + ")"
    }
    if filters.DateRange != nil {
        query += fmt.Sprintf(" AND application_date >= $%d AND application_date <= $%d", argIndex, argIndex+1)
        args = append(args, filters.DateRange.Start, filters.DateRange.End)
        argIndex += 2
    }
    if len(filters.CompanyNames) > 0 {
        placeholders := make([]string, len(filters.CompanyNames))
        for i, c := range filters.CompanyNames { placeholders[i] = fmt.Sprintf("$%d", argIndex); args = append(args, c); argIndex++ }
        query += " AND company_name IN (" + strings.Join(placeholders, ",") + ")"
    }
    if filters.Keywords != "" {
        query += fmt.Sprintf(" AND (company_name ILIKE $%d OR position_title ILIKE $%d OR notes ILIKE $%d)", argIndex, argIndex, argIndex)
        kw := "%" + filters.Keywords + "%"; args = append(args, kw)
    }
    return query, args
}

func buildDataQueryInternal(userID uint, filters *model.ExportFilters, offset, limit int) (string, []interface{}) {
    query := "SELECT id, user_id, company_name, position_title, application_date, status, job_description, salary_range, work_location, contact_info, notes, interview_time, reminder_time, reminder_enabled, follow_up_date, hr_name, hr_phone, hr_email, interview_location, interview_type, created_at, updated_at FROM job_applications WHERE user_id = $1"
    args := []interface{}{userID}
    argIndex := 2
    if len(filters.Status) > 0 {
        placeholders := make([]string, len(filters.Status))
        for i, st := range filters.Status { placeholders[i] = fmt.Sprintf("$%d", argIndex); args = append(args, st); argIndex++ }
        query += " AND status IN (" + strings.Join(placeholders, ",") + ")"
    }
    if filters.DateRange != nil {
        query += fmt.Sprintf(" AND application_date >= $%d AND application_date <= $%d", argIndex, argIndex+1)
        args = append(args, filters.DateRange.Start, filters.DateRange.End)
        argIndex += 2
    }
    if len(filters.CompanyNames) > 0 {
        placeholders := make([]string, len(filters.CompanyNames))
        for i, c := range filters.CompanyNames { placeholders[i] = fmt.Sprintf("$%d", argIndex); args = append(args, c); argIndex++ }
        query += " AND company_name IN (" + strings.Join(placeholders, ",") + ")"
    }
    if filters.Keywords != "" {
        query += fmt.Sprintf(" AND (company_name ILIKE $%d OR position_title ILIKE $%d OR notes ILIKE $%d)", argIndex, argIndex, argIndex)
        kw := "%" + filters.Keywords + "%"; args = append(args, kw)
    }
    query += " ORDER BY application_date DESC, created_at DESC"
    if limit > 0 {
        query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
        args = append(args, limit, offset)
    }
    return query, args
}

// 复用格式化（轻量）
func formatFileSize(bytes int64) string {
    const unit = 1024
    if bytes < unit { return fmt.Sprintf("%d B", bytes) }
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit { div *= unit; exp++ }
    return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
