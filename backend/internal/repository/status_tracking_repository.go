package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"jobView-backend/internal/database"
	"jobView-backend/internal/model"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type StatusTrackingRepository interface {
	GetStatusHistory(ctx context.Context, userID uint, jobApplicationID, page, pageSize int) (int, []model.StatusHistoryEntry, error)
	GetStatusTimeline(ctx context.Context, userID uint, jobApplicationID int) ([]model.StatusHistoryEntry, error)
	GetInterviewExperiences(ctx context.Context, userID uint, jobApplicationID int) ([]model.InterviewExperience, error)
	BeginTx(ctx context.Context) (StatusTrackingTx, error)
	GetStatusAnalytics(ctx context.Context, userID uint) (*model.StatusAnalyticsResponse, error)
	GetStatusTrends(ctx context.Context, userID uint, days int) ([]model.StatusTrend, error)
}

type StatusTrackingTx interface {
	GetJobApplicationForUpdate(userID uint, jobApplicationID int) (*JobStatusSnapshot, error)
	SetLocalFlag(flag string, enabled bool) error
	InsertStatusHistory(entry StatusHistoryInsert) (int64, error)
	InsertInterviewExperience(entry InterviewExperienceInsert) (int64, error)
	UpdateJobApplication(params UpdateJobApplicationParams) (*model.JobApplication, error)
	GetCurrentStatus(userID uint, jobApplicationID int) (model.ApplicationStatus, *time.Time, error)
	GetLatestHistoryEntry(userID uint, jobApplicationID int) (*model.StatusHistoryEntry, error)
	UpdateHistoryMetadata(historyID int64, metadata []byte) error
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

type InterviewExperienceInsert struct {
	ApplicationID int
	UserID        uint
	FromStatus    model.ApplicationStatus
	ToStatus      model.ApplicationStatus
	Rating        *string
	Note          *string
	Skip          bool
	SkipReason    *string
	RecordedAt    time.Time
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

func (r *statusTrackingRepo) GetInterviewExperiences(ctx context.Context, userID uint, jobApplicationID int) ([]model.InterviewExperience, error) {
	orm, err := r.ormWithContext(ctx)
	if err != nil {
		return nil, err
	}

	var exists bool
	if err := orm.
		Raw("SELECT EXISTS(SELECT 1 FROM job_applications WHERE id = $1 AND user_id = $2)", jobApplicationID, userID).
		Row().
		Scan(&exists); err != nil {
		return nil, fmt.Errorf("failed to verify job application access: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("job application not found or access denied")
	}

	query := `SELECT id, application_id, from_status, to_status, rating, note, skip, skip_reason, recorded_by, recorded_at, created_at
	          FROM interview_experiences
	          WHERE application_id = $1
	          ORDER BY recorded_at DESC, id DESC`
	rows, err := orm.Raw(query, jobApplicationID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query interview experiences: %w", err)
	}
	defer rows.Close()

	var experiences []model.InterviewExperience
	for rows.Next() {
		var (
			experience model.InterviewExperience
			rating     sql.NullString
			note       sql.NullString
			skipReason sql.NullString
		)
		if err := rows.Scan(
			&experience.ID,
			&experience.ApplicationID,
			&experience.FromStatus,
			&experience.ToStatus,
			&rating,
			&note,
			&experience.Skip,
			&skipReason,
			&experience.RecordedBy,
			&experience.RecordedAt,
			&experience.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan interview experience: %w", err)
		}
		if rating.Valid {
			r := rating.String
			experience.Rating = &r
		}
		if note.Valid {
			n := note.String
			experience.Note = &n
		}
		if skipReason.Valid {
			reason := skipReason.String
			experience.SkipReason = &reason
		}
		experiences = append(experiences, experience)
	}

	return experiences, nil
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
	legacyAliases := map[string][]string{
		string(model.StatusResumeScreeningFail): {"简历挂"},
		string(model.StatusWrittenTestFail):     {"笔试挂"},
		string(model.StatusFirstFail):           {"一面挂"},
		string(model.StatusSecondFail):          {"二面挂"},
		string(model.StatusThirdFail):           {"三面挂"},
		string(model.StatusHRFail):              {"HR面挂"},
		string(model.StatusHRPass):              {"待发offer", "已收到offer"},
		string(model.StatusRejected):            {"被拒绝", "已拒绝"},
		string(model.StatusOfferAccepted):       {"已入职"},
	}
	expandStatuses := func(statuses []string) (canon []string, aliases []string) {
		if len(statuses) == 0 {
			return nil, nil
		}
		seen := make(map[string]struct{}, len(statuses))
		seenCanon := make(map[string]struct{})
		seenAlias := make(map[string]struct{})

		queue := make([]string, 0, len(statuses))
		for _, status := range statuses {
			queue = append(queue, status)
			if extra, ok := legacyAliases[status]; ok {
				queue = append(queue, extra...)
			}
		}
		for _, status := range queue {
			if _, ok := seen[status]; ok {
				continue
			}
			seen[status] = struct{}{}
			if model.ApplicationStatus(status).IsValid() {
				if _, ok := seenCanon[status]; !ok {
					seenCanon[status] = struct{}{}
					canon = append(canon, status)
				}
			} else {
				if _, ok := seenAlias[status]; !ok {
					seenAlias[status] = struct{}{}
					aliases = append(aliases, status)
				}
			}
		}
		return canon, aliases
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
		Entry        []string // 真正进入该阶段的状态(只统计实际经历过该阶段的岗位)
		PassStatuses []string // 通过该阶段后的所有状态(用于判断是否通过该阶段)
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
			},
		},
		{
			Name: "hr",
			Entry: []string{
				string(model.StatusHRInterview), string(model.StatusHRPass), string(model.StatusHRFail),
			},
			PassStatuses: []string{
				string(model.StatusHRPass),
			},
		},
	}
	fetchIDsFromHistory := func(statuses []string) (map[int]struct{}, error) {
		canonical, aliasOnly := expandStatuses(statuses)
		if len(canonical) == 0 && len(aliasOnly) == 0 {
			return map[int]struct{}{}, nil
		}
		conditions := make([]string, 0, 2)
		args := []interface{}{userID}

		if len(canonical) > 0 {
			args = append(args, pq.Array(canonical))
			placeholder := len(args)
			conditions = append(conditions, fmt.Sprintf("new_status = ANY($%d)", placeholder))
		}
		if len(aliasOnly) > 0 {
			args = append(args, pq.Array(aliasOnly))
			placeholder := len(args)
			conditions = append(conditions, fmt.Sprintf("new_status::text = ANY($%d)", placeholder))
		}

		query := fmt.Sprintf(`
			SELECT DISTINCT job_application_id
			FROM job_status_history
			WHERE user_id = $1 AND (%s)
		`, strings.Join(conditions, " OR "))

		rows, err := orm.Raw(query, args...).Rows()
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		result := make(map[int]struct{})
		for rows.Next() {
			var id int
			if scanErr := rows.Scan(&id); scanErr != nil {
				return nil, scanErr
			}
			result[id] = struct{}{}
		}
		return result, nil
	}

	for _, st := range stages {
		entryIDs, entryErr := fetchIDsFromHistory(st.Entry)
		if entryErr != nil {
			return nil, fmt.Errorf("failed to compute history entry for %s: %w", st.Name, entryErr)
		}
		passIDs, passErr := fetchIDsFromHistory(st.PassStatuses)
		if passErr != nil {
			return nil, fmt.Errorf("failed to compute history pass for %s: %w", st.Name, passErr)
		}

		totalStage := len(entryIDs)
		passed := 0
		if totalStage > 0 && len(passIDs) > 0 {
			for id := range passIDs {
				if _, ok := entryIDs[id]; ok {
					passed++
				}
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

func (t *statusTrackingTx) InsertInterviewExperience(entry InterviewExperienceInsert) (int64, error) {
	query := `INSERT INTO interview_experiences (
	              application_id,
	              from_status,
	              to_status,
	              rating,
	              note,
	              skip,
	              skip_reason,
	              recorded_by,
	              recorded_at
	          ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	          RETURNING id`

	var (
		ratingVal     interface{}
		noteVal       interface{}
		skipReasonVal interface{}
	)
	if entry.Rating != nil {
		ratingVal = *entry.Rating
	}
	if entry.Note != nil {
		noteVal = *entry.Note
	}
	if entry.SkipReason != nil {
		skipReasonVal = *entry.SkipReason
	}

	var id int64
	if err := t.tx.Raw(
		query,
		entry.ApplicationID,
		entry.FromStatus,
		entry.ToStatus,
		ratingVal,
		noteVal,
		entry.Skip,
		skipReasonVal,
		int(entry.UserID),
		entry.RecordedAt,
	).Row().Scan(&id); err != nil {
		return 0, fmt.Errorf("failed to insert interview experience: %w", err)
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
		                    created_at, updated_at, status_version`
		var job model.JobApplication
		var reminderCategory sql.NullString
		var statusVersion sql.NullInt32
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
			&statusVersion,
		); err != nil {
			return nil, fmt.Errorf("failed to update job application: %w", err)
		}
		if reminderCategory.Valid {
			rc := reminderCategory.String
			job.ReminderCategory = &rc
		} else {
			job.ReminderCategory = nil
		}
		if statusVersion.Valid {
			v := int(statusVersion.Int32)
			job.StatusVersion = &v
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
	          created_at, updated_at, status_version`

	var job model.JobApplication
	var reminderCategory sql.NullString
	var statusVersion sql.NullInt32
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
		&statusVersion,
	); err != nil {
		return nil, fmt.Errorf("failed to update job application: %w", err)
	}
	if reminderCategory.Valid {
		rc := reminderCategory.String
		job.ReminderCategory = &rc
	} else {
		job.ReminderCategory = nil
	}
	if statusVersion.Valid {
		v := int(statusVersion.Int32)
		job.StatusVersion = &v
	} else {
		job.StatusVersion = nil
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

func (t *statusTrackingTx) GetLatestHistoryEntry(userID uint, jobApplicationID int) (*model.StatusHistoryEntry, error) {
	query := `SELECT id, job_application_id, user_id, old_status, new_status, status_changed_at, duration_minutes, metadata, created_at
              FROM job_status_history
              WHERE job_application_id = $1 AND user_id = $2
              ORDER BY status_changed_at DESC, id DESC
              LIMIT 1
              FOR UPDATE`
	row := t.tx.Raw(query, jobApplicationID, userID).Row()
	var entry model.StatusHistoryEntry
	var oldStatus sql.NullString
	var duration sql.NullInt32
	var metadataBytes []byte
	if err := row.Scan(
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
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get latest history entry: %w", err)
	}
	if oldStatus.Valid {
		value := model.ApplicationStatus(oldStatus.String)
		entry.OldStatus = &value
	}
	if duration.Valid {
		val := int(duration.Int32)
		entry.DurationMinutes = &val
	}
	if len(metadataBytes) > 0 {
		if err := json.Unmarshal(metadataBytes, &entry.Metadata); err != nil {
			return nil, fmt.Errorf("failed to decode history metadata: %w", err)
		}
	}
	return &entry, nil
}

func (t *statusTrackingTx) UpdateHistoryMetadata(historyID int64, metadata []byte) error {
	if len(metadata) == 0 {
		metadata = []byte("{}")
	}
	if err := t.tx.Exec("UPDATE job_status_history SET metadata = $1 WHERE id = $2", metadata, historyID).Error; err != nil {
		return fmt.Errorf("failed to update history metadata: %w", err)
	}
	return nil
}

func (t *statusTrackingTx) Commit() error {
	return t.tx.Commit().Error
}

func (t *statusTrackingTx) Rollback() error {
	return t.tx.Rollback().Error
}
