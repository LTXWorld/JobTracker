package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"jobView-backend/internal/database"
	"jobView-backend/internal/model"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type StatusTrackingRepository interface {
	GetStatusHistory(ctx context.Context, userID uint, jobApplicationID, page, pageSize int) (int, []model.StatusHistoryEntry, error)
	GetStatusTimeline(ctx context.Context, userID uint, jobApplicationID int) ([]model.StatusHistoryEntry, error)
	BeginTx(ctx context.Context) (StatusTrackingTx, error)
	GetStatusAnalytics(ctx context.Context, userID uint) (*model.StatusAnalyticsResponse, error)
	GetStatusTrends(ctx context.Context, userID uint, days int) ([]model.StatusTrend, error)
}

type StatusTrackingTx interface {
	GetJobApplicationForUpdate(userID uint, jobApplicationID int) (*JobStatusSnapshot, error)
	SetLocalFlag(flag string, enabled bool) error
	InsertStatusHistory(entry StatusHistoryInsert) (int64, error)
	UpdateJobApplication(params UpdateJobApplicationParams) (*model.JobApplication, error)
	GetCurrentStatus(userID uint, jobApplicationID int) (model.ApplicationStatus, *time.Time, error)
	Commit() error
	Rollback() error
}

type JobStatusSnapshot struct {
	Job              model.JobApplication
	StatusVersion    *int
	StatusHistoryRaw *string
	DurationStatsRaw *string
	LastStatusChange *time.Time
}

type StatusHistoryInsert struct {
	JobApplicationID int
	UserID           uint
	OldStatus        model.ApplicationStatus
	NewStatus        model.ApplicationStatus
	ChangedAt        time.Time
	DurationMinutes  *int
	Metadata         []byte
}

type UpdateJobApplicationParams struct {
	JobApplicationID  int
	UserID            uint
	NewStatus         model.ApplicationStatus
	Now               time.Time
	LastStatusChange  *time.Time
	StatusVersion     *int
	IncrementVersion  bool
	StatusHistoryJSON []byte
	DurationStatsJSON []byte
	SuppressHistory   bool
	UseTrigger        bool
}

type statusTrackingRepo struct {
	db  *database.DB
	orm *gorm.DB
}

func NewStatusTrackingRepository(db *database.DB) StatusTrackingRepository {
	var orm *gorm.DB
	if db != nil {
		orm = db.ORM
	}
	return &statusTrackingRepo{db: db, orm: orm}
}

func (r *statusTrackingRepo) ormWithContext(ctx context.Context) (*gorm.DB, error) {
	if r.orm == nil {
		return nil, fmt.Errorf("gorm instance not initialized")
	}
	return r.orm.WithContext(ctx), nil
}

// ---- Query implementations ----

func (r *statusTrackingRepo) GetStatusHistory(ctx context.Context, userID uint, jobApplicationID, page, pageSize int) (int, []model.StatusHistoryEntry, error) {
	orm, err := r.ormWithContext(ctx)
	if err != nil {
		return 0, nil, err
	}
	var exists bool
	if err := orm.Raw("SELECT EXISTS(SELECT 1 FROM job_applications WHERE id = $1 AND user_id = $2)", jobApplicationID, userID).Row().Scan(&exists); err != nil {
		return 0, nil, fmt.Errorf("failed to verify job application access: %w", err)
	}
	if !exists {
		return 0, nil, fmt.Errorf("job application not found or access denied")
	}

	var total int
	if err := orm.Raw("SELECT COUNT(*) FROM job_status_history WHERE job_application_id = $1", jobApplicationID).Row().Scan(&total); err != nil {
		return 0, nil, fmt.Errorf("failed to count status history: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `SELECT id, job_application_id, user_id, old_status, new_status, status_changed_at, duration_minutes, metadata, created_at
              FROM job_status_history
              WHERE job_application_id = $1
              ORDER BY status_changed_at DESC
              LIMIT $2 OFFSET $3`
	rows, err := orm.Raw(query, jobApplicationID, pageSize, offset).Rows()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get status history: %w", err)
	}
	defer rows.Close()

	var history []model.StatusHistoryEntry
	for rows.Next() {
		var entry model.StatusHistoryEntry
		var metadataBytes []byte
		var oldStatus sql.NullString
		var duration sql.NullInt32
		if err := rows.Scan(
			&entry.ID,
			&entry.JobApplicationID,
			&entry.UserID,
			&oldStatus,
			&entry.NewStatus,
			&entry.StatusChangedAt,
			&duration,
			&metadataBytes,
			&entry.CreatedAt,
		); err != nil {
			return 0, nil, fmt.Errorf("failed to scan status history: %w", err)
		}
		if oldStatus.Valid {
			value := model.ApplicationStatus(oldStatus.String)
			entry.OldStatus = &value
		}
		if duration.Valid {
			d := int(duration.Int32)
			entry.DurationMinutes = &d
		}
		if len(metadataBytes) > 0 {
			_ = json.Unmarshal(metadataBytes, &entry.Metadata)
		}
		history = append(history, entry)
	}

	return total, history, nil
}

func (r *statusTrackingRepo) GetStatusTimeline(ctx context.Context, userID uint, jobApplicationID int) ([]model.StatusHistoryEntry, error) {
	orm, err := r.ormWithContext(ctx)
	if err != nil {
		return nil, err
	}
	var exists bool
	if err := orm.Raw("SELECT EXISTS(SELECT 1 FROM job_applications WHERE id = $1 AND user_id = $2)", jobApplicationID, userID).Row().Scan(&exists); err != nil {
		return nil, fmt.Errorf("failed to verify job application access: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("job application not found or access denied")
	}

	query := `SELECT old_status, new_status, status_changed_at, duration_minutes, metadata
              FROM job_status_history
              WHERE job_application_id = $1
              ORDER BY status_changed_at ASC`
	rows, err := orm.Raw(query, jobApplicationID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get status timeline: %w", err)
	}
	defer rows.Close()

	var timeline []model.StatusHistoryEntry
	for rows.Next() {
		var entry model.StatusHistoryEntry
		var oldStatus sql.NullString
		var duration sql.NullInt32
		var metadataBytes []byte
		if err := rows.Scan(&oldStatus, &entry.NewStatus, &entry.StatusChangedAt, &duration, &metadataBytes); err != nil {
			return nil, fmt.Errorf("failed to scan timeline entry: %w", err)
		}
		entry.JobApplicationID = jobApplicationID
		entry.UserID = userID
		if oldStatus.Valid {
			value := model.ApplicationStatus(oldStatus.String)
			entry.OldStatus = &value
		}
		if duration.Valid {
			d := int(duration.Int32)
			entry.DurationMinutes = &d
		}
		if len(metadataBytes) > 0 {
			_ = json.Unmarshal(metadataBytes, &entry.Metadata)
		}
		timeline = append(timeline, entry)
	}
	return timeline, nil
}

func (r *statusTrackingRepo) BeginTx(ctx context.Context) (StatusTrackingTx, error) {
	orm, err := r.ormWithContext(ctx)
	if err != nil {
		return nil, err
	}
	tx := orm.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &statusTrackingTx{tx: tx}, nil
}

func (r *statusTrackingRepo) GetStatusAnalytics(ctx context.Context, userID uint) (*model.StatusAnalyticsResponse, error) {
	orm, err := r.ormWithContext(ctx)
	if err != nil {
		return nil, err
	}
	analytics := &model.StatusAnalyticsResponse{
		UserID:             userID,
		StatusDistribution: make(map[string]int),
		AverageDurations:   make(map[string]float64),
		StageAnalysis:      make(map[string]model.StageStatistics),
	}

	rows, err := orm.Raw("SELECT status, COUNT(*) FROM job_applications WHERE user_id = $1 GROUP BY status", userID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get status distribution: %w", err)
	}
	defer rows.Close()

	total := 0
	success := 0
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status distribution: %w", err)
		}
		analytics.StatusDistribution[status] = count
		total += count
		if model.ApplicationStatus(status).IsPassedStatus() {
			success += count
		}
	}
	analytics.TotalApplications = total
	if total > 0 {
		analytics.SuccessRate = float64(success) / float64(total) * 100
	}

	durationRows, err := orm.Raw("SELECT old_status, AVG(duration_minutes) FROM job_status_history WHERE user_id = $1 AND old_status IS NOT NULL AND duration_minutes IS NOT NULL GROUP BY old_status", userID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get average durations: %w", err)
	}
	defer durationRows.Close()

	for durationRows.Next() {
		var status string
		var avg float64
		if err := durationRows.Scan(&status, &avg); err != nil {
			return nil, fmt.Errorf("failed to scan average duration: %w", err)
		}
		analytics.AverageDurations[status] = avg
	}

	type stageDef struct {
		Name         string
		Entry        []string // 该阶段及之后的所有状态（用于统计总数）
		PassStatuses []string // 通过该阶段后的所有状态（用于统计通过数）
	}
	stages := []stageDef{
		{
			Name: "written",
			Entry: []string{
				string(model.StatusWrittenTest),
				string(model.StatusWrittenTestPass), string(model.StatusWrittenTestFail),
			},
			PassStatuses: []string{
				string(model.StatusWrittenTestPass),
				string(model.StatusFirstInterview), string(model.StatusFirstPass), string(model.StatusFirstFail),
				string(model.StatusSecondInterview), string(model.StatusSecondPass), string(model.StatusSecondFail),
				string(model.StatusThirdInterview), string(model.StatusThirdPass), string(model.StatusThirdFail),
				string(model.StatusHRInterview), string(model.StatusHRPass), string(model.StatusHRFail),
				string(model.StatusOfferAccepted), string(model.StatusRejected),
			},
		},
		{
			Name: "first",
			Entry: []string{
				string(model.StatusFirstInterview), string(model.StatusFirstPass), string(model.StatusFirstFail),
			},
			PassStatuses: []string{
				string(model.StatusFirstPass),
				string(model.StatusSecondInterview), string(model.StatusSecondPass), string(model.StatusSecondFail),
				string(model.StatusThirdInterview), string(model.StatusThirdPass), string(model.StatusThirdFail),
				string(model.StatusHRInterview), string(model.StatusHRPass), string(model.StatusHRFail),
				string(model.StatusOfferAccepted), string(model.StatusRejected),
			},
		},
		{
			Name: "second",
			Entry: []string{
				string(model.StatusSecondInterview), string(model.StatusSecondPass), string(model.StatusSecondFail),
			},
			PassStatuses: []string{
				string(model.StatusSecondPass),
				string(model.StatusThirdInterview), string(model.StatusThirdPass), string(model.StatusThirdFail),
				string(model.StatusHRInterview), string(model.StatusHRPass), string(model.StatusHRFail),
				string(model.StatusOfferAccepted), string(model.StatusRejected),
			},
		},
		{
			Name: "third",
			Entry: []string{
				string(model.StatusThirdInterview), string(model.StatusThirdPass), string(model.StatusThirdFail),
			},
			PassStatuses: []string{
				string(model.StatusThirdPass),
				string(model.StatusHRInterview), string(model.StatusHRPass), string(model.StatusHRFail),
				string(model.StatusOfferAccepted), string(model.StatusRejected),
			},
		},
		{
			Name: "hr",
			Entry: []string{
				string(model.StatusHRInterview), string(model.StatusHRPass), string(model.StatusHRFail),
			},
			PassStatuses: []string{
				string(model.StatusHRPass),
				string(model.StatusOfferAccepted), string(model.StatusRejected),
			},
		},
	}
	// 检查是否存在历史数据，如果没有则回退到基于当前状态的估算
	hasHistory := false
	{
		var marker int
		err := orm.Raw(`SELECT 1 FROM job_status_history WHERE user_id = $1 LIMIT 1`, userID).Row().Scan(&marker)
		if err == nil {
			hasHistory = true
		} else if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to check history availability: %w", err)
		}
	}

	countFromHistory := func(stageName string, statuses []string) (int, error) {
		if len(statuses) == 0 {
			return 0, nil
		}
		var count int
		query := `
			SELECT COUNT(DISTINCT job_application_id)
			FROM job_status_history
			WHERE user_id = $1 AND new_status = ANY($2)
		`
		if err := orm.Raw(query, userID, pq.Array(statuses)).Row().Scan(&count); err != nil {
			return 0, fmt.Errorf("failed to compute history count for %s: %w", stageName, err)
		}
		return count, nil
	}

	countFromCurrentStatus := func(stageName string, statuses []string) (int, error) {
		if len(statuses) == 0 {
			return 0, nil
		}
		var count int
		query := `
			SELECT COUNT(*) FROM job_applications
			WHERE user_id = $1 AND status = ANY($2)
		`
		if err := orm.Raw(query, userID, pq.Array(statuses)).Row().Scan(&count); err != nil {
			return 0, fmt.Errorf("failed to compute current status count for %s: %w", stageName, err)
		}
		return count, nil
	}

	for _, st := range stages {
		var totalStage, passed int
		var err error
		if hasHistory {
			totalStage, err = countFromHistory(st.Name, st.Entry)
			if err != nil {
				return nil, err
			}
			passed, err = countFromHistory(st.Name, st.PassStatuses)
			if err != nil {
				return nil, err
			}
		} else {
			totalStage, err = countFromCurrentStatus(st.Name, st.Entry)
			if err != nil {
				return nil, err
			}
			passed, err = countFromCurrentStatus(st.Name, st.PassStatuses)
			if err != nil {
				return nil, err
			}
		}

		var rate float64
		if totalStage > 0 {
			rate = float64(passed) / float64(totalStage) * 100
		}
		analytics.StageAnalysis[st.Name] = model.StageStatistics{
			StageName:    st.Name,
			TotalCount:   totalStage,
			SuccessCount: passed,
			SuccessRate:  rate,
		}
	}
	return analytics, nil
}

func (r *statusTrackingRepo) GetStatusTrends(ctx context.Context, userID uint, days int) ([]model.StatusTrend, error) {
	orm, err := r.ormWithContext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := orm.Raw("SELECT DATE(status_changed_at) as date, new_status, COUNT(*) FROM job_status_history WHERE user_id = $1 AND status_changed_at >= $2 GROUP BY DATE(status_changed_at), new_status ORDER BY date DESC, count DESC", userID, time.Now().AddDate(0, 0, -days)).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get status trends: %w", err)
	}
	defer rows.Close()

	var trends []model.StatusTrend
	for rows.Next() {
		var date time.Time
		var status string
		var count int
		if err := rows.Scan(&date, &status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status trend: %w", err)
		}
		trends = append(trends, model.StatusTrend{
			Date:   date.Format("2006-01-02"),
			Status: status,
			Count:  count,
		})
	}
	return trends, nil
}

// ---- Transaction implementation ----

type statusTrackingTx struct {
	tx *gorm.DB
}

func (t *statusTrackingTx) GetJobApplicationForUpdate(userID uint, jobApplicationID int) (*JobStatusSnapshot, error) {
	query := `SELECT id, user_id, company_name, position_title, application_date, status,
                     job_description, salary_range, work_location, contact_info, notes,
                     interview_time, reminder_time, reminder_enabled, reminder_category, follow_up_date,
                     hr_name, hr_phone, hr_email, interview_location, interview_type,
                     created_at, updated_at, last_status_change, status_version,
                     status_history, status_duration_stats
              FROM job_applications
              WHERE id = $1 AND user_id = $2`
	var record model.JobApplication
	var lastChange sql.NullTime
	var version sql.NullInt32
	var history sql.NullString
	var duration sql.NullString
	var reminderCategory sql.NullString
	if err := t.tx.Raw(query, jobApplicationID, userID).Row().Scan(
		&record.ID,
		&record.UserID,
		&record.CompanyName,
		&record.PositionTitle,
		&record.ApplicationDate,
		&record.Status,
		&record.JobDescription,
		&record.SalaryRange,
		&record.WorkLocation,
		&record.ContactInfo,
		&record.Notes,
		&record.InterviewTime,
		&record.ReminderTime,
		&record.ReminderEnabled,
		&reminderCategory,
		&record.FollowUpDate,
		&record.HRName,
		&record.HRPhone,
		&record.HREmail,
		&record.InterviewLocation,
		&record.InterviewType,
		&record.CreatedAt,
		&record.UpdatedAt,
		&lastChange,
		&version,
		&history,
		&duration,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job application not found")
		}
		return nil, fmt.Errorf("failed to fetch job application: %w", err)
	}

	if reminderCategory.Valid {
		rc := reminderCategory.String
		record.ReminderCategory = &rc
	}
	snapshot := &JobStatusSnapshot{Job: record}
	if lastChange.Valid {
		tm := lastChange.Time
		snapshot.LastStatusChange = &tm
	}
	if version.Valid {
		v := int(version.Int32)
		snapshot.StatusVersion = &v
	}
	if history.Valid {
		val := history.String
		snapshot.StatusHistoryRaw = &val
	}
	if duration.Valid {
		val := duration.String
		snapshot.DurationStatsRaw = &val
	}
	return snapshot, nil
}

func (t *statusTrackingTx) SetLocalFlag(flag string, enabled bool) error {
	value := "off"
	if enabled {
		value = "on"
	}
	switch flag {
	case "jobview.skip_history", "jobview.allow_backward":
		return t.tx.Exec(fmt.Sprintf("SET LOCAL %s = '%s'", flag, value)).Error
	default:
		return fmt.Errorf("unsupported session flag: %s", flag)
	}
}

func (t *statusTrackingTx) InsertStatusHistory(entry StatusHistoryInsert) (int64, error) {
	query := `INSERT INTO job_status_history (job_application_id, user_id, old_status, new_status, status_changed_at, duration_minutes, metadata)
              VALUES ($1,$2,$3,$4,$5,$6,$7)
              RETURNING id`
	var id int64
	if err := t.tx.Raw(query, entry.JobApplicationID, entry.UserID, entry.OldStatus, entry.NewStatus, entry.ChangedAt, entry.DurationMinutes, entry.Metadata).Row().Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (t *statusTrackingTx) UpdateJobApplication(params UpdateJobApplicationParams) (*model.JobApplication, error) {
	if params.SuppressHistory || params.UseTrigger {
		query := `UPDATE job_applications SET status = $1, updated_at = $2 WHERE id = $3 AND user_id = $4
                  RETURNING id, user_id, company_name, position_title, application_date, status,
                            job_description, salary_range, work_location, contact_info, notes,
                            interview_time, reminder_time, reminder_enabled, reminder_category, follow_up_date,
                            hr_name, hr_phone, hr_email, interview_location, interview_type,
                            created_at, updated_at`
		var job model.JobApplication
		var reminderCategory sql.NullString
		if err := t.tx.Raw(query, params.NewStatus, params.Now, params.JobApplicationID, params.UserID).Row().Scan(
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
			&reminderCategory,
			&job.FollowUpDate,
			&job.HRName,
			&job.HRPhone,
			&job.HREmail,
			&job.InterviewLocation,
			&job.InterviewType,
			&job.CreatedAt,
			&job.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to update job application: %w", err)
		}
		if reminderCategory.Valid {
			rc := reminderCategory.String
			job.ReminderCategory = &rc
		} else {
			job.ReminderCategory = nil
		}
		return &job, nil
	}

	query := "UPDATE job_applications SET status = $1"
	args := []interface{}{params.NewStatus}
	idx := 2

	if params.LastStatusChange != nil {
		query += fmt.Sprintf(", last_status_change = $%d", idx)
		args = append(args, *params.LastStatusChange)
		idx++
	}
	query += fmt.Sprintf(", updated_at = $%d", idx)
	args = append(args, params.Now)
	idx++

	if params.StatusVersion != nil {
		query += fmt.Sprintf(", status_version = $%d", idx)
		args = append(args, *params.StatusVersion)
		idx++
	} else if params.IncrementVersion {
		query += ", status_version = COALESCE(status_version, 0) + 1"
	}
	if params.StatusHistoryJSON != nil {
		query += fmt.Sprintf(", status_history = $%d::jsonb", idx)
		args = append(args, string(params.StatusHistoryJSON))
		idx++
	}
	if params.DurationStatsJSON != nil {
		query += fmt.Sprintf(", status_duration_stats = $%d::jsonb", idx)
		args = append(args, string(params.DurationStatsJSON))
		idx++
	}

	query += fmt.Sprintf(" WHERE id = $%d AND user_id = $%d RETURNING ", idx, idx+1)
	args = append(args, params.JobApplicationID, params.UserID)

	query += `id, user_id, company_name, position_title, application_date, status,
              job_description, salary_range, work_location, contact_info, notes,
              interview_time, reminder_time, reminder_enabled, reminder_category, follow_up_date,
              hr_name, hr_phone, hr_email, interview_location, interview_type,
              created_at, updated_at`

	var job model.JobApplication
	var reminderCategory sql.NullString
	if err := t.tx.Raw(query, args...).Row().Scan(
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
		&reminderCategory,
		&job.FollowUpDate,
		&job.HRName,
		&job.HRPhone,
		&job.HREmail,
		&job.InterviewLocation,
		&job.InterviewType,
		&job.CreatedAt,
		&job.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("failed to update job application: %w", err)
	}
	if reminderCategory.Valid {
		rc := reminderCategory.String
		job.ReminderCategory = &rc
	} else {
		job.ReminderCategory = nil
	}
	return &job, nil
}

func (t *statusTrackingTx) GetCurrentStatus(userID uint, jobApplicationID int) (model.ApplicationStatus, *time.Time, error) {
	query := "SELECT status, last_status_change FROM job_applications WHERE id = $1 AND user_id = $2"
	var status model.ApplicationStatus
	var last sql.NullTime
	if err := t.tx.Raw(query, jobApplicationID, userID).Row().Scan(&status, &last); err != nil {
		if err == sql.ErrNoRows {
			return "", nil, err
		}
		return "", nil, fmt.Errorf("failed to get current status: %w", err)
	}
	var ts *time.Time
	if last.Valid {
		tm := last.Time
		ts = &tm
	}
	return status, ts, nil
}

func (t *statusTrackingTx) Commit() error {
	return t.tx.Commit().Error
}

func (t *statusTrackingTx) Rollback() error {
	return t.tx.Rollback().Error
}
