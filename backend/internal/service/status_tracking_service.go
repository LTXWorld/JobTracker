// Location: /Users/lutao/GolandProjects/jobView/backend/internal/service/status_tracking_service.go
// This file implements the core status tracking service for JobView system.
// It handles job application status history, transitions, analytics, and preference management.
// Used by the status tracking handlers to provide business logic and data processing.

package service

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "jobView-backend/internal/database"
    "jobView-backend/internal/model"
    "os"
    "strings"
    "time"
)

type StatusTrackingService struct {
	db *database.DB
}

func NewStatusTrackingService(db *database.DB) *StatusTrackingService {
	return &StatusTrackingService{db: db}
}

// GetStatusHistory 获取岗位状态历史记录
func (s *StatusTrackingService) GetStatusHistory(userID uint, jobApplicationID int, page, pageSize int) (*model.StatusHistoryResponse, error) {
    // GORM path behind flag (Raw/Scan)
    if s.db != nil && s.db.UseGorm && s.db.ORM != nil {
        return s.getStatusHistoryGorm(userID, jobApplicationID, page, pageSize)
    }
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// 验证用户权限
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM job_applications WHERE id = $1 AND user_id = $2)"
	err := s.db.QueryRow(checkQuery, jobApplicationID, userID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to verify job application access: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("job application not found or access denied")
	}

	// 获取总数
	var total int
	countQuery := "SELECT COUNT(*) FROM job_status_history WHERE job_application_id = $1"
	err = s.db.QueryRow(countQuery, jobApplicationID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count status history: %w", err)
	}

	// 获取历史记录
	offset := (page - 1) * pageSize
	historyQuery := `
		SELECT id, job_application_id, user_id, old_status, new_status, 
		       status_changed_at, duration_minutes, metadata, created_at
		FROM job_status_history 
		WHERE job_application_id = $1 
		ORDER BY status_changed_at DESC 
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.Query(historyQuery, jobApplicationID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get status history: %w", err)
	}
	defer rows.Close()

	var history []model.StatusHistoryEntry
	for rows.Next() {
		var entry model.StatusHistoryEntry
		var metadataBytes []byte
		var oldStatusStr sql.NullString

		err := rows.Scan(
			&entry.ID,
			&entry.JobApplicationID,
			&entry.UserID,
			&oldStatusStr,
			&entry.NewStatus,
			&entry.StatusChangedAt,
			&entry.DurationMinutes,
			&metadataBytes,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status history entry: %w", err)
		}

		// 处理旧状态（可能为NULL）
		if oldStatusStr.Valid {
			oldStatus := model.ApplicationStatus(oldStatusStr.String)
			entry.OldStatus = &oldStatus
		}

		// 解析元数据
		if len(metadataBytes) > 0 {
			var metadata map[string]interface{}
			if err := json.Unmarshal(metadataBytes, &metadata); err == nil {
				entry.Metadata = metadata
			}
		}

		history = append(history, entry)
	}

	return &model.StatusHistoryResponse{
		History:     history,
		Total:       total,
		CurrentPage: page,
		PageSize:    pageSize,
	}, nil
}

// getStatusHistoryGorm 使用 GORM 的 Raw/Rows 读取历史
func (s *StatusTrackingService) getStatusHistoryGorm(userID uint, jobApplicationID int, page, pageSize int) (*model.StatusHistoryResponse, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    if page < 1 { page = 1 }
    if pageSize < 1 || pageSize > 100 { pageSize = 50 }

    // 验证用户权限
    var exists bool
    if err := s.db.ORM.WithContext(ctx).Raw("SELECT EXISTS(SELECT 1 FROM job_applications WHERE id = $1 AND user_id = $2)", jobApplicationID, userID).Row().Scan(&exists); err != nil {
        return nil, fmt.Errorf("failed to verify job application access: %w", err)
    }
    if !exists { return nil, fmt.Errorf("job application not found or access denied") }

    // 计数
    var total int
    if err := s.db.ORM.WithContext(ctx).Raw("SELECT COUNT(*) FROM job_status_history WHERE job_application_id = $1", jobApplicationID).Row().Scan(&total); err != nil {
        return nil, fmt.Errorf("failed to count status history: %w", err)
    }

    offset := (page - 1) * pageSize
    historyQuery := `
        SELECT id, job_application_id, user_id, old_status, new_status,
               status_changed_at, duration_minutes, metadata, created_at
        FROM job_status_history
        WHERE job_application_id = $1
        ORDER BY status_changed_at DESC
        LIMIT $2 OFFSET $3`

    rows, err := s.db.ORM.WithContext(ctx).Raw(historyQuery, jobApplicationID, pageSize, offset).Rows()
    if err != nil { return nil, fmt.Errorf("failed to get status history: %w", err) }
    defer rows.Close()

    var history []model.StatusHistoryEntry
    for rows.Next() {
        var entry model.StatusHistoryEntry
        var metadataBytes []byte
        var oldStatusStr sql.NullString
        if err := rows.Scan(
            &entry.ID,
            &entry.JobApplicationID,
            &entry.UserID,
            &oldStatusStr,
            &entry.NewStatus,
            &entry.StatusChangedAt,
            &entry.DurationMinutes,
            &metadataBytes,
            &entry.CreatedAt,
        ); err != nil {
            return nil, fmt.Errorf("failed to scan status history entry: %w", err)
        }
        if oldStatusStr.Valid { os := model.ApplicationStatus(oldStatusStr.String); entry.OldStatus = &os }
        if len(metadataBytes) > 0 {
            var md map[string]interface{}
            if json.Unmarshal(metadataBytes, &md) == nil { entry.Metadata = md }
        }
        history = append(history, entry)
    }

    return &model.StatusHistoryResponse{History: history, Total: total, CurrentPage: page, PageSize: pageSize}, nil
}

// UpdateJobStatus 更新岗位状态并记录历史
func (s *StatusTrackingService) UpdateJobStatus(userID uint, jobApplicationID int, request *model.StatusUpdateRequest) (*model.JobApplication, error) {
    // GORM path behind flag
    if s.db != nil && s.db.UseGorm && s.db.ORM != nil {
        return s.updateJobStatusGorm(userID, jobApplicationID, request)
    }
    // 验证状态有效性
    if !request.Status.IsValid() {
        return nil, fmt.Errorf("invalid status: %s", request.Status)
    }

	// 开始事务
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 获取当前记录（含乐观锁检查）
	var currentJob model.JobApplication
	var currentVersion sql.NullInt32
	var currentStatusHistory sql.NullString
	var currentDurationStats sql.NullString

	getCurrentQuery := `
		SELECT id, user_id, company_name, position_title, application_date, status,
		       job_description, salary_range, work_location, contact_info, notes,
		       interview_time, reminder_time, reminder_enabled, follow_up_date,
		       hr_name, hr_phone, hr_email, interview_location, interview_type,
		       created_at, updated_at, last_status_change, status_version,
		       status_history, status_duration_stats
		FROM job_applications 
		WHERE id = $1 AND user_id = $2
	`

	var lastStatusChange sql.NullTime
	err = tx.QueryRow(getCurrentQuery, jobApplicationID, userID).Scan(
		&currentJob.ID,
		&currentJob.UserID,
		&currentJob.CompanyName,
		&currentJob.PositionTitle,
		&currentJob.ApplicationDate,
		&currentJob.Status,
		&currentJob.JobDescription,
		&currentJob.SalaryRange,
		&currentJob.WorkLocation,
		&currentJob.ContactInfo,
		&currentJob.Notes,
		&currentJob.InterviewTime,
		&currentJob.ReminderTime,
		&currentJob.ReminderEnabled,
		&currentJob.FollowUpDate,
		&currentJob.HRName,
		&currentJob.HRPhone,
		&currentJob.HREmail,
		&currentJob.InterviewLocation,
		&currentJob.InterviewType,
		&currentJob.CreatedAt,
		&currentJob.UpdatedAt,
		&lastStatusChange,
		&currentVersion,
		&currentStatusHistory,
		&currentDurationStats,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job application not found")
		}
		return nil, fmt.Errorf("failed to get current job application: %w", err)
	}

	// 乐观锁检查
	if request.Version != nil && currentVersion.Valid {
		if int32(*request.Version) != currentVersion.Int32 {
			return nil, fmt.Errorf("version conflict: expected %d, got %d", currentVersion.Int32, *request.Version)
		}
	}

	// 检查状态是否有变化
	if currentJob.Status == request.Status {
		return &currentJob, nil // 状态未变化，直接返回
	}

    // 先按模板/内置规则验证；若失败且为回退操作且已确认，则放行
    validateErr := s.validateStatusTransition(userID, currentJob.Status, request.Status)

    isBackward := s.isBackwardTransition(currentJob.Status, request.Status)
    if validateErr != nil {
        if isBackward {
            // 是否允许回退（默认允许，可用环境变量控制）
            allowBackward := true
            if v := os.Getenv("ALLOW_BACKWARD_STATUS"); v != "" {
                allowBackward = strings.EqualFold(v, "true") || v == "1"
            }
            if !allowBackward {
                return nil, fmt.Errorf("BACKWARD_DISABLED")
            }
            // 必须确认
            if request.ConfirmBackward == nil || !*request.ConfirmBackward {
                return nil, fmt.Errorf("BACKWARD_CONFIRM_REQUIRED")
            }
            // 终态回退必须填写备注（流程结束/已拒绝/各阶段未通过）
            if s.isTerminalStatus(currentJob.Status) {
                if request.Note == nil || strings.TrimSpace(*request.Note) == "" {
                    return nil, fmt.Errorf("NOTE_REQUIRED_FOR_BACKWARD")
                }
            }
            // 放行（不再返回 validateErr）
        } else {
            return nil, validateErr
        }
    }

    // 会话级GUC：一律禁用触发器写历史，由应用层统一维护；
    // 回退+确认时额外允许回退。
    _, _ = tx.Exec("SET LOCAL jobview.skip_history = 'on'")
    suppressHistory := isBackward && request.ConfirmBackward != nil && *request.ConfirmBackward
    if suppressHistory {
        _, _ = tx.Exec("SET LOCAL jobview.allow_backward = 'on'")
    }

    // 计算状态持续时间
    now := time.Now()
    var durationMinutes *int
	if lastStatusChange.Valid {
		duration := int(now.Sub(lastStatusChange.Time).Minutes())
		durationMinutes = &duration
	}

    // 是否交由数据库触发器记录历史
    // 注意：回退场景下我们选择不更新历史（suppressHistory=true）
    useDBTriggerForHistory := false

    var statusHistoryBytes []byte
    var durationStatsBytes []byte
    if !useDBTriggerForHistory && !suppressHistory {
        // 创建历史记录（应用层）
	    insertHistoryQuery := `
		    INSERT INTO job_status_history (job_application_id, user_id, old_status, new_status, 
		                                   status_changed_at, duration_minutes, metadata)
		    VALUES ($1, $2, $3, $4, $5, $6, $7)
		    RETURNING id
	    `

	    // 准备metadata，确保是有效的JSON对象
        // 准备metadata，标记回退等信息
        var metadata map[string]interface{}
        if request.Metadata != nil {
            metadata = make(map[string]interface{}, len(request.Metadata)+4)
            for k, v := range request.Metadata {
                metadata[k] = v
            }
        } else {
            metadata = map[string]interface{}{}
        }
        if isBackward {
            metadata["backward"] = true
            metadata["from"] = string(currentJob.Status)
            metadata["to"] = string(request.Status)
            if request.Note != nil && strings.TrimSpace(*request.Note) != "" {
                metadata["note"] = strings.TrimSpace(*request.Note)
            }
        }
        metadataBytes, _ := json.Marshal(metadata)
	    var historyID int64
	    err = tx.QueryRow(insertHistoryQuery, jobApplicationID, userID, currentJob.Status,
		    request.Status, now, durationMinutes, metadataBytes).Scan(&historyID)
	    if err != nil {
		    return nil, fmt.Errorf("failed to insert status history: %w", err)
	    }

	    // 更新状态历史JSON（应用层）
	    statusHistory := s.updateStatusHistoryJSON(currentStatusHistory.String, currentJob.Status, request.Status, now, durationMinutes)
	    statusHistoryBytes, _ = json.Marshal(statusHistory)

	    // 更新持续时间统计（应用层）
	    durationStats := s.updateDurationStats(currentDurationStats.String, currentJob.Status, durationMinutes)
	    durationStatsBytes, _ = json.Marshal(durationStats)
    }

	// 更新主记录
	newVersion := 1
	if currentVersion.Valid {
		newVersion = int(currentVersion.Int32) + 1
	}

	var updatedJob model.JobApplication
    if suppressHistory {
        // 仅更新状态与更新时间，不影响 last_status_change / 版本 / 历史
        updateQuery := `
            UPDATE job_applications 
            SET status = $1, updated_at = $2
            WHERE id = $3 AND user_id = $4
            RETURNING id, user_id, company_name, position_title, application_date, status,
                      job_description, salary_range, work_location, contact_info, notes,
                      interview_time, reminder_time, reminder_enabled, follow_up_date,
                      hr_name, hr_phone, hr_email, interview_location, interview_type,
                      created_at, updated_at
        `
        err = tx.QueryRow(updateQuery, request.Status, now, jobApplicationID, userID).Scan(
            &updatedJob.ID,
            &updatedJob.UserID,
            &updatedJob.CompanyName,
            &updatedJob.PositionTitle,
            &updatedJob.ApplicationDate,
            &updatedJob.Status,
            &updatedJob.JobDescription,
            &updatedJob.SalaryRange,
            &updatedJob.WorkLocation,
            &updatedJob.ContactInfo,
            &updatedJob.Notes,
            &updatedJob.InterviewTime,
            &updatedJob.ReminderTime,
            &updatedJob.ReminderEnabled,
            &updatedJob.FollowUpDate,
            &updatedJob.HRName,
            &updatedJob.HRPhone,
            &updatedJob.HREmail,
            &updatedJob.InterviewLocation,
            &updatedJob.InterviewType,
            &updatedJob.CreatedAt,
            &updatedJob.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to update job application: %w", err)
        }
    } else if useDBTriggerForHistory {
        // 交由触发器维护 last_status_change/status_version/历史
        updateQuery := `
            UPDATE job_applications 
            SET status = $1, updated_at = $2
            WHERE id = $3 AND user_id = $4
			RETURNING id, user_id, company_name, position_title, application_date, status,
			          job_description, salary_range, work_location, contact_info, notes,
			          interview_time, reminder_time, reminder_enabled, follow_up_date,
			          hr_name, hr_phone, hr_email, interview_location, interview_type,
			          created_at, updated_at
		`
		err = tx.QueryRow(updateQuery, request.Status, now, jobApplicationID, userID).Scan(
			&updatedJob.ID,
			&updatedJob.UserID,
			&updatedJob.CompanyName,
			&updatedJob.PositionTitle,
			&updatedJob.ApplicationDate,
			&updatedJob.Status,
			&updatedJob.JobDescription,
			&updatedJob.SalaryRange,
			&updatedJob.WorkLocation,
			&updatedJob.ContactInfo,
			&updatedJob.Notes,
			&updatedJob.InterviewTime,
			&updatedJob.ReminderTime,
			&updatedJob.ReminderEnabled,
			&updatedJob.FollowUpDate,
			&updatedJob.HRName,
			&updatedJob.HRPhone,
			&updatedJob.HREmail,
			&updatedJob.InterviewLocation,
			&updatedJob.InterviewType,
			&updatedJob.CreatedAt,
			&updatedJob.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update job application: %w", err)
		}
    } else {
        updateQuery := `
            UPDATE job_applications 
            SET status = $1, last_status_change = $2, status_version = $3, 
                status_history = $4::jsonb, status_duration_stats = $5::jsonb, updated_at = $6
            WHERE id = $7 AND user_id = $8
            RETURNING id, user_id, company_name, position_title, application_date, status,
                      job_description, salary_range, work_location, contact_info, notes,
                      interview_time, reminder_time, reminder_enabled, follow_up_date,
                      hr_name, hr_phone, hr_email, interview_location, interview_type,
                      created_at, updated_at
        `

		err = tx.QueryRow(updateQuery, request.Status, now, newVersion,
			string(statusHistoryBytes), string(durationStatsBytes), now,
			jobApplicationID, userID).Scan(
		&updatedJob.ID,
		&updatedJob.UserID,
		&updatedJob.CompanyName,
		&updatedJob.PositionTitle,
		&updatedJob.ApplicationDate,
		&updatedJob.Status,
		&updatedJob.JobDescription,
		&updatedJob.SalaryRange,
		&updatedJob.WorkLocation,
		&updatedJob.ContactInfo,
		&updatedJob.Notes,
		&updatedJob.InterviewTime,
		&updatedJob.ReminderTime,
		&updatedJob.ReminderEnabled,
		&updatedJob.FollowUpDate,
		&updatedJob.HRName,
		&updatedJob.HRPhone,
		&updatedJob.HREmail,
		&updatedJob.InterviewLocation,
		&updatedJob.InterviewType,
		&updatedJob.CreatedAt,
		&updatedJob.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to update job application: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &updatedJob, nil
}

// updateJobStatusGorm 使用 GORM 事务执行状态更新（保留与原有逻辑一致的行为）
func (s *StatusTrackingService) updateJobStatusGorm(userID uint, jobApplicationID int, request *model.StatusUpdateRequest) (*model.JobApplication, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    if !request.Status.IsValid() {
        return nil, fmt.Errorf("invalid status: %s", request.Status)
    }

    tx := s.db.ORM.WithContext(ctx).Begin()
    if tx.Error != nil {
        return nil, fmt.Errorf("failed to begin gorm tx: %w", tx.Error)
    }
    defer func() { _ = tx.Rollback().Error }()

    // 查询当前记录
    var currentJob model.JobApplication
    var currentVersion sql.NullInt32
    var currentStatusHistory sql.NullString
    var currentDurationStats sql.NullString
    var lastStatusChange sql.NullTime

    getCurrentQuery := `
        SELECT id, user_id, company_name, position_title, application_date, status,
               job_description, salary_range, work_location, contact_info, notes,
               interview_time, reminder_time, reminder_enabled, follow_up_date,
               hr_name, hr_phone, hr_email, interview_location, interview_type,
               created_at, updated_at, last_status_change, status_version,
               status_history, status_duration_stats
        FROM job_applications 
        WHERE id = $1 AND user_id = $2`

    row := tx.Raw(getCurrentQuery, jobApplicationID, userID).Row()
    if err := row.Scan(
        &currentJob.ID,
        &currentJob.UserID,
        &currentJob.CompanyName,
        &currentJob.PositionTitle,
        &currentJob.ApplicationDate,
        &currentJob.Status,
        &currentJob.JobDescription,
        &currentJob.SalaryRange,
        &currentJob.WorkLocation,
        &currentJob.ContactInfo,
        &currentJob.Notes,
        &currentJob.InterviewTime,
        &currentJob.ReminderTime,
        &currentJob.ReminderEnabled,
        &currentJob.FollowUpDate,
        &currentJob.HRName,
        &currentJob.HRPhone,
        &currentJob.HREmail,
        &currentJob.InterviewLocation,
        &currentJob.InterviewType,
        &currentJob.CreatedAt,
        &currentJob.UpdatedAt,
        &lastStatusChange,
        &currentVersion,
        &currentStatusHistory,
        &currentDurationStats,
    ); err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("job application not found")
        }
        return nil, fmt.Errorf("failed to get current job application: %w", err)
    }

    // 乐观锁
    if request.Version != nil && currentVersion.Valid {
        if int32(*request.Version) != currentVersion.Int32 {
            return nil, fmt.Errorf("version conflict: expected %d, got %d", currentVersion.Int32, *request.Version)
        }
    }

    // 无变化
    if currentJob.Status == request.Status {
        _ = tx.Rollback().Error
        return &currentJob, nil
    }

    // 校验/回退放行
    validateErr := s.validateStatusTransition(userID, currentJob.Status, request.Status)
    isBackward := s.isBackwardTransition(currentJob.Status, request.Status)
    if validateErr != nil {
        if isBackward {
            allowBackward := true
            if v := os.Getenv("ALLOW_BACKWARD_STATUS"); v != "" {
                allowBackward = strings.EqualFold(v, "true") || v == "1"
            }
            if !allowBackward {
                return nil, fmt.Errorf("BACKWARD_DISABLED")
            }
            if request.ConfirmBackward == nil || !*request.ConfirmBackward {
                return nil, fmt.Errorf("BACKWARD_CONFIRM_REQUIRED")
            }
            if s.isTerminalStatus(currentJob.Status) {
                if request.Note == nil || strings.TrimSpace(*request.Note) == "" {
                    return nil, fmt.Errorf("NOTE_REQUIRED_FOR_BACKWARD")
                }
            }
        } else {
            return nil, validateErr
        }
    }

    // 关闭触发器历史，必要时允许回退
    _ = tx.Exec("SET LOCAL jobview.skip_history = 'on'").Error
    suppressHistory := isBackward && request.ConfirmBackward != nil && *request.ConfirmBackward
    if suppressHistory {
        _ = tx.Exec("SET LOCAL jobview.allow_backward = 'on'").Error
    }

    // 持续时间
    now := time.Now()
    var durationMinutes *int
    if lastStatusChange.Valid {
        duration := int(now.Sub(lastStatusChange.Time).Minutes())
        durationMinutes = &duration
    }

    // 应用侧历史/统计（仅非回退）
    var statusHistoryBytes []byte
    var durationStatsBytes []byte
    if !suppressHistory {
        insertHistoryQuery := `
            INSERT INTO job_status_history (job_application_id, user_id, old_status, new_status, 
                                           status_changed_at, duration_minutes, metadata)
            VALUES ($1, $2, $3, $4, $5, $6, $7)
            RETURNING id`
        metadata := map[string]interface{}{}
        if request.Metadata != nil {
            for k, v := range request.Metadata { metadata[k] = v }
        }
        if isBackward {
            metadata["backward"] = true
            metadata["from"] = string(currentJob.Status)
            metadata["to"] = string(request.Status)
            if request.Note != nil && strings.TrimSpace(*request.Note) != "" {
                metadata["note"] = strings.TrimSpace(*request.Note)
            }
        }
        metadataBytes, _ := json.Marshal(metadata)
        var historyID int64
        if err := tx.Raw(insertHistoryQuery, jobApplicationID, userID, currentJob.Status, request.Status, now, durationMinutes, metadataBytes).Row().Scan(&historyID); err != nil {
            return nil, fmt.Errorf("failed to insert status history: %w", err)
        }

        // 计算并序列化 JSON 字段
        statusHistory := s.updateStatusHistoryJSON(currentStatusHistory.String, currentJob.Status, request.Status, now, durationMinutes)
        statusHistoryBytes, _ = json.Marshal(statusHistory)
        durationStats := s.updateDurationStats(currentDurationStats.String, currentJob.Status, durationMinutes)
        durationStatsBytes, _ = json.Marshal(durationStats)
    }

    // 版本号
    newVersion := 1
    if currentVersion.Valid {
        newVersion = int(currentVersion.Int32) + 1
    }

    // 更新主记录
    var updatedJob model.JobApplication
    if suppressHistory {
        updateQuery := `
            UPDATE job_applications 
            SET status = $1, updated_at = $2
            WHERE id = $3 AND user_id = $4
            RETURNING id, user_id, company_name, position_title, application_date, status,
                      job_description, salary_range, work_location, contact_info, notes,
                      interview_time, reminder_time, reminder_enabled, follow_up_date,
                      hr_name, hr_phone, hr_email, interview_location, interview_type,
                      created_at, updated_at`
        if err := tx.Raw(updateQuery, request.Status, now, jobApplicationID, userID).Row().Scan(
            &updatedJob.ID,
            &updatedJob.UserID,
            &updatedJob.CompanyName,
            &updatedJob.PositionTitle,
            &updatedJob.ApplicationDate,
            &updatedJob.Status,
            &updatedJob.JobDescription,
            &updatedJob.SalaryRange,
            &updatedJob.WorkLocation,
            &updatedJob.ContactInfo,
            &updatedJob.Notes,
            &updatedJob.InterviewTime,
            &updatedJob.ReminderTime,
            &updatedJob.ReminderEnabled,
            &updatedJob.FollowUpDate,
            &updatedJob.HRName,
            &updatedJob.HRPhone,
            &updatedJob.HREmail,
            &updatedJob.InterviewLocation,
            &updatedJob.InterviewType,
            &updatedJob.CreatedAt,
            &updatedJob.UpdatedAt,
        ); err != nil {
            return nil, fmt.Errorf("failed to update job application: %w", err)
        }
    } else {
        updateQuery := `
            UPDATE job_applications 
            SET status = $1, last_status_change = $2, status_version = $3, 
                status_history = $4::jsonb, status_duration_stats = $5::jsonb, updated_at = $6
            WHERE id = $7 AND user_id = $8
            RETURNING id, user_id, company_name, position_title, application_date, status,
                      job_description, salary_range, work_location, contact_info, notes,
                      interview_time, reminder_time, reminder_enabled, follow_up_date,
                      hr_name, hr_phone, hr_email, interview_location, interview_type,
                      created_at, updated_at`
        if err := tx.Raw(updateQuery, request.Status, now, newVersion, string(statusHistoryBytes), string(durationStatsBytes), now, jobApplicationID, userID).Row().Scan(
            &updatedJob.ID,
            &updatedJob.UserID,
            &updatedJob.CompanyName,
            &updatedJob.PositionTitle,
            &updatedJob.ApplicationDate,
            &updatedJob.Status,
            &updatedJob.JobDescription,
            &updatedJob.SalaryRange,
            &updatedJob.WorkLocation,
            &updatedJob.ContactInfo,
            &updatedJob.Notes,
            &updatedJob.InterviewTime,
            &updatedJob.ReminderTime,
            &updatedJob.ReminderEnabled,
            &updatedJob.FollowUpDate,
            &updatedJob.HRName,
            &updatedJob.HRPhone,
            &updatedJob.HREmail,
            &updatedJob.InterviewLocation,
            &updatedJob.InterviewType,
            &updatedJob.CreatedAt,
            &updatedJob.UpdatedAt,
        ); err != nil {
            return nil, fmt.Errorf("failed to update job application: %w", err)
        }
    }

    if err := tx.Commit().Error; err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }
    return &updatedJob, nil
}

// GetStatusTimeline 获取岗位状态时间轴视图
func (s *StatusTrackingService) GetStatusTimeline(userID uint, jobApplicationID int) (map[string]interface{}, error) {
    if s.db != nil && s.db.UseGorm && s.db.ORM != nil {
        return s.getStatusTimelineGorm(userID, jobApplicationID)
    }
	// 验证用户权限
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM job_applications WHERE id = $1 AND user_id = $2)"
	err := s.db.QueryRow(checkQuery, jobApplicationID, userID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to verify job application access: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("job application not found or access denied")
	}

	// 获取详细状态历史
	historyQuery := `
		SELECT old_status, new_status, status_changed_at, duration_minutes, metadata
		FROM job_status_history 
		WHERE job_application_id = $1 
		ORDER BY status_changed_at ASC
	`

	rows, err := s.db.Query(historyQuery, jobApplicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status timeline: %w", err)
	}
	defer rows.Close()

	var timeline []map[string]interface{}
	totalDuration := 0

	for rows.Next() {
		var oldStatusStr sql.NullString
		var newStatus string
		var changedAt time.Time
		var durationMinutes sql.NullInt32
		var metadataBytes []byte

		err := rows.Scan(&oldStatusStr, &newStatus, &changedAt, &durationMinutes, &metadataBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to scan timeline entry: %w", err)
		}

		timelineEntry := map[string]interface{}{
			"new_status":        newStatus,
			"status_changed_at": changedAt,
		}

		if oldStatusStr.Valid {
			timelineEntry["old_status"] = oldStatusStr.String
		}

		if durationMinutes.Valid {
			timelineEntry["duration_minutes"] = durationMinutes.Int32
			totalDuration += int(durationMinutes.Int32)
		}

		// 解析元数据
		if len(metadataBytes) > 0 {
			var metadata map[string]interface{}
			if err := json.Unmarshal(metadataBytes, &metadata); err == nil {
				timelineEntry["metadata"] = metadata
			}
		}

		timeline = append(timeline, timelineEntry)
	}

	return map[string]interface{}{
		"job_application_id":    jobApplicationID,
		"timeline":             timeline,
		"total_duration_minutes": totalDuration,
		"total_changes":        len(timeline),
	}, nil
}

func (s *StatusTrackingService) getStatusTimelineGorm(userID uint, jobApplicationID int) (map[string]interface{}, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    var exists bool
    if err := s.db.ORM.WithContext(ctx).Raw("SELECT EXISTS(SELECT 1 FROM job_applications WHERE id = $1 AND user_id = $2)", jobApplicationID, userID).Row().Scan(&exists); err != nil {
        return nil, fmt.Errorf("failed to verify job application access: %w", err)
    }
    if !exists { return nil, fmt.Errorf("job application not found or access denied") }

    historyQuery := `
        SELECT old_status, new_status, status_changed_at, duration_minutes, metadata
        FROM job_status_history
        WHERE job_application_id = $1
        ORDER BY status_changed_at ASC`
    rows, err := s.db.ORM.WithContext(ctx).Raw(historyQuery, jobApplicationID).Rows()
    if err != nil { return nil, fmt.Errorf("failed to get status timeline: %w", err) }
    defer rows.Close()

    var timeline []map[string]interface{}
    totalDuration := 0
    for rows.Next() {
        var oldStatusStr sql.NullString
        var newStatus string
        var changedAt time.Time
        var durationMinutes sql.NullInt32
        var metadataBytes []byte
        if err := rows.Scan(&oldStatusStr, &newStatus, &changedAt, &durationMinutes, &metadataBytes); err != nil {
            return nil, fmt.Errorf("failed to scan timeline entry: %w", err)
        }
        entry := map[string]interface{}{"new_status": newStatus, "status_changed_at": changedAt}
        if oldStatusStr.Valid { entry["old_status"] = oldStatusStr.String }
        if durationMinutes.Valid { entry["duration_minutes"] = durationMinutes.Int32; totalDuration += int(durationMinutes.Int32) }
        if len(metadataBytes) > 0 { var md map[string]interface{}; if json.Unmarshal(metadataBytes, &md) == nil { entry["metadata"] = md } }
        timeline = append(timeline, entry)
    }
    return map[string]interface{}{"job_application_id": jobApplicationID, "timeline": timeline, "total_duration_minutes": totalDuration, "total_changes": len(timeline)}, nil
}

// BatchUpdateStatus 批量状态更新
func (s *StatusTrackingService) BatchUpdateStatus(userID uint, updates []model.BatchStatusUpdate) error {
    if s.db != nil && s.db.UseGorm && s.db.ORM != nil {
        return s.batchUpdateStatusGorm(userID, updates)
    }
	if len(updates) == 0 {
		return nil
	}
	if len(updates) > 100 {
		return fmt.Errorf("batch size too large: maximum 100 updates allowed")
	}

	// 验证所有状态
	for _, update := range updates {
		if !update.Status.IsValid() {
			return fmt.Errorf("invalid status: %s for ID %d", update.Status, update.ID)
		}
	}

	// 开始事务
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()

	for _, update := range updates {
		// 获取当前状态
		var currentStatus model.ApplicationStatus
		var lastStatusChange sql.NullTime
		getCurrentQuery := `
			SELECT status, last_status_change 
			FROM job_applications 
			WHERE id = $1 AND user_id = $2
		`
		err = tx.QueryRow(getCurrentQuery, update.ID, userID).Scan(&currentStatus, &lastStatusChange)
		if err != nil {
			if err == sql.ErrNoRows {
				continue // 跳过不存在或无权限的记录
			}
			return fmt.Errorf("failed to get current status for ID %d: %w", update.ID, err)
		}

		// 跳过相同状态
		if currentStatus == update.Status {
			continue
		}

        // 选项A：批量更新不允许回退
        if s.isBackwardTransition(currentStatus, update.Status) {
            return fmt.Errorf("backward transitions are not allowed in batch updates (ID %d: %s -> %s)", update.ID, currentStatus, update.Status)
        }
        // 验证状态转换（前进方向仍按模板/直通规则校验）
        if err := s.validateStatusTransition(userID, currentStatus, update.Status); err != nil {
            return fmt.Errorf("invalid transition for ID %d: %w", update.ID, err)
        }

		// 计算持续时间
		var durationMinutes *int
		if lastStatusChange.Valid {
			duration := int(now.Sub(lastStatusChange.Time).Minutes())
			durationMinutes = &duration
		}

		// 插入历史记录
		insertHistoryQuery := `
			INSERT INTO job_status_history (job_application_id, user_id, old_status, new_status, 
			                               status_changed_at, duration_minutes, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, '{}')
		`
		_, err = tx.Exec(insertHistoryQuery, update.ID, userID, currentStatus, update.Status, now, durationMinutes)
		if err != nil {
			return fmt.Errorf("failed to insert history for ID %d: %w", update.ID, err)
		}

		// 更新状态
		updateQuery := `
			UPDATE job_applications 
			SET status = $1, last_status_change = $2, updated_at = $3,
			    status_version = COALESCE(status_version, 0) + 1
			WHERE id = $4 AND user_id = $5
		`
		_, err = tx.Exec(updateQuery, update.Status, now, now, update.ID, userID)
		if err != nil {
			return fmt.Errorf("failed to update status for ID %d: %w", update.ID, err)
		}
	}

	return tx.Commit()
}

func (s *StatusTrackingService) batchUpdateStatusGorm(userID uint, updates []model.BatchStatusUpdate) error {
    if len(updates) == 0 { return nil }
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    if len(updates) > 100 { return fmt.Errorf("batch size too large: maximum 100 updates allowed") }
    for _, u := range updates { if !u.Status.IsValid() { return fmt.Errorf("invalid status: %s for ID %d", u.Status, u.ID) } }

    tx := s.db.ORM.WithContext(ctx).Begin()
    if tx.Error != nil { return fmt.Errorf("failed to begin tx: %w", tx.Error) }
    defer func() { _ = tx.Rollback().Error }()

    // 禁用触发器写历史，由应用侧写入
    _ = tx.Exec("SET LOCAL jobview.skip_history = 'on'").Error
    now := time.Now()

    for _, u := range updates {
        var currentStatus model.ApplicationStatus
        var lastStatusChange sql.NullTime
        row := tx.Raw("SELECT status, last_status_change FROM job_applications WHERE id = $1 AND user_id = $2", u.ID, userID).Row()
        if err := row.Scan(&currentStatus, &lastStatusChange); err != nil {
            if err == sql.ErrNoRows { continue }
            return fmt.Errorf("failed to get current status for ID %d: %w", u.ID, err)
        }
        if currentStatus == u.Status { continue }
        if s.isBackwardTransition(currentStatus, u.Status) { return fmt.Errorf("backward transitions are not allowed in batch updates (ID %d: %s -> %s)", u.ID, currentStatus, u.Status) }
        if err := s.validateStatusTransition(userID, currentStatus, u.Status); err != nil { return fmt.Errorf("invalid transition for ID %d: %w", u.ID, err) }

        var durationMinutes *int
        if lastStatusChange.Valid { d := int(now.Sub(lastStatusChange.Time).Minutes()); durationMinutes = &d }

        // 插入历史
        if err := tx.Exec("INSERT INTO job_status_history (job_application_id, user_id, old_status, new_status, status_changed_at, duration_minutes, metadata) VALUES ($1,$2,$3,$4,$5,$6,'{}')", u.ID, userID, currentStatus, u.Status, now, durationMinutes).Error; err != nil {
            return fmt.Errorf("failed to insert history for ID %d: %w", u.ID, err)
        }
        // 更新主表
        if err := tx.Exec("UPDATE job_applications SET status=$1, last_status_change=$2, updated_at=$3, status_version = COALESCE(status_version,0)+1 WHERE id=$4 AND user_id=$5", u.Status, now, now, u.ID, userID).Error; err != nil {
            return fmt.Errorf("failed to update status for ID %d: %w", u.ID, err)
        }
    }
    return tx.Commit().Error
}

// GetStatusAnalytics 获取用户状态分析数据
func (s *StatusTrackingService) GetStatusAnalytics(userID uint) (*model.StatusAnalyticsResponse, error) {
    if s.db != nil && s.db.UseGorm && s.db.ORM != nil {
        return s.getStatusAnalyticsGorm(userID)
    }
	analytics := &model.StatusAnalyticsResponse{
		UserID:             userID,
		StatusDistribution: make(map[string]int),
		AverageDurations:   make(map[string]float64),
		StageAnalysis:      make(map[string]model.StageStatistics),
	}

	// 获取状态分布
	statusQuery := `
		SELECT status, COUNT(*) as count
		FROM job_applications 
		WHERE user_id = $1
		GROUP BY status
		ORDER BY count DESC
	`
	rows, err := s.db.Query(statusQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status distribution: %w", err)
	}
	defer rows.Close()

	totalApplications := 0
	successCount := 0
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status distribution: %w", err)
		}
		analytics.StatusDistribution[status] = count
		totalApplications += count

		// 计算成功率
		appStatus := model.ApplicationStatus(status)
		if appStatus.IsPassedStatus() {
			successCount += count
		}
	}

	analytics.TotalApplications = totalApplications
	if totalApplications > 0 {
		analytics.SuccessRate = float64(successCount) / float64(totalApplications) * 100
	}

	// 获取平均持续时间
	durationQuery := `
		SELECT old_status, AVG(duration_minutes) as avg_duration
		FROM job_status_history 
		WHERE user_id = $1 AND old_status IS NOT NULL AND duration_minutes IS NOT NULL
		GROUP BY old_status
	`
	durationRows, err := s.db.Query(durationQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get average durations: %w", err)
	}
	defer durationRows.Close()

	for durationRows.Next() {
		var status string
		var avgDuration float64
		if err := durationRows.Scan(&status, &avgDuration); err != nil {
			return nil, fmt.Errorf("failed to scan average duration: %w", err)
		}
		analytics.AverageDurations[status] = avgDuration
	}

    // 计算各阶段通过率（基于历史变迁）
    type stageDef struct{ Name, Entry, Next string }
    stages := []stageDef{
        {Name: "written", Entry: string(model.StatusWrittenTest), Next: string(model.StatusFirstInterview)},
        {Name: "first", Entry: string(model.StatusFirstInterview), Next: string(model.StatusSecondInterview)},
        {Name: "second", Entry: string(model.StatusSecondInterview), Next: string(model.StatusThirdInterview)},
        {Name: "third", Entry: string(model.StatusThirdInterview), Next: string(model.StatusHRInterview)},
    }

    for _, st := range stages {
        // reached: 曾经到过该阶段（出现 new_status = Entry）或当前就在该阶段（job_applications.status = Entry）
        reachedQuery := `
            SELECT COUNT(*) FROM (
                SELECT DISTINCT job_application_id FROM job_status_history WHERE user_id = $1 AND new_status = $2
                UNION
                SELECT id AS job_application_id FROM job_applications WHERE user_id = $1 AND status = $2
            ) t`
        var total int
        if err := s.db.QueryRow(reachedQuery, userID, st.Entry).Scan(&total); err != nil {
            return nil, fmt.Errorf("failed to compute stage total for %s: %w", st.Name, err)
        }

        // pass: 出现 old_status = Entry 且 new_status = Next 的转移（直通表达通过）
        passQuery := `
            SELECT COUNT(DISTINCT job_application_id)
            FROM job_status_history
            WHERE user_id = $1 AND old_status = $2 AND new_status = $3`
        var passed int
        if err := s.db.QueryRow(passQuery, userID, st.Entry, st.Next).Scan(&passed); err != nil {
            return nil, fmt.Errorf("failed to compute stage pass for %s: %w", st.Name, err)
        }

        var rate float64
        if total > 0 {
            rate = float64(passed) / float64(total) * 100
        }

        analytics.StageAnalysis[st.Name] = model.StageStatistics{
            StageName:           st.Name,
            TotalCount:          total,
            SuccessCount:        passed,
            SuccessRate:         rate,
            AverageDurationDays: 0,
        }
    }

    return analytics, nil
}

func (s *StatusTrackingService) getStatusAnalyticsGorm(userID uint) (*model.StatusAnalyticsResponse, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    analytics := &model.StatusAnalyticsResponse{UserID: userID, StatusDistribution: make(map[string]int), AverageDurations: make(map[string]float64), StageAnalysis: make(map[string]model.StageStatistics)}
    // 分布
    rows, err := s.db.ORM.WithContext(ctx).Raw("SELECT status, COUNT(*) as count FROM job_applications WHERE user_id = $1 GROUP BY status ORDER BY count DESC", userID).Rows()
    if err != nil { return nil, fmt.Errorf("failed to get status distribution: %w", err) }
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
    if total>0 { analytics.SuccessRate = float64(success)/float64(total)*100 }

    // 平均持续时间
    dRows, err := s.db.ORM.WithContext(ctx).Raw("SELECT old_status, AVG(duration_minutes) as avg_duration FROM job_status_history WHERE user_id = $1 AND old_status IS NOT NULL AND duration_minutes IS NOT NULL GROUP BY old_status", userID).Rows()
    if err != nil { return nil, fmt.Errorf("failed to get average durations: %w", err) }
    defer dRows.Close()
    for dRows.Next() {
        var status string
        var avg float64
        if err := dRows.Scan(&status, &avg); err != nil {
            return nil, fmt.Errorf("failed to scan average duration: %w", err)
        }
        analytics.AverageDurations[status] = avg
    }

    // 阶段通过率（与原逻辑保持一致）
    type stageDef struct{ Name, Entry, Next string }
    stages := []stageDef{{"written", string(model.StatusWrittenTest), string(model.StatusFirstInterview)},{"first", string(model.StatusFirstInterview), string(model.StatusSecondInterview)},{"second", string(model.StatusSecondInterview), string(model.StatusThirdInterview)},{"third", string(model.StatusThirdInterview), string(model.StatusHRInterview)}}
    for _, st := range stages {
        var totalStage int; if err := s.db.ORM.WithContext(ctx).Raw(`SELECT COUNT(*) FROM (SELECT DISTINCT job_application_id FROM job_status_history WHERE user_id = $1 AND new_status = $2 UNION SELECT id AS job_application_id FROM job_applications WHERE user_id = $1 AND status = $2) t`, userID, st.Entry).Row().Scan(&totalStage); err != nil { return nil, fmt.Errorf("failed to compute stage total for %s: %w", st.Name, err) }
        var passed int; if err := s.db.ORM.WithContext(ctx).Raw(`SELECT COUNT(DISTINCT job_application_id) FROM job_status_history WHERE user_id = $1 AND old_status = $2 AND new_status = $3`, userID, st.Entry, st.Next).Row().Scan(&passed); err != nil { return nil, fmt.Errorf("failed to compute stage pass for %s: %w", st.Name, err) }
        var rate float64; if totalStage>0 { rate = float64(passed)/float64(totalStage)*100 }
        analytics.StageAnalysis[st.Name] = model.StageStatistics{StageName: st.Name, TotalCount: totalStage, SuccessCount: passed, SuccessRate: rate, AverageDurationDays: 0}
    }
    return analytics, nil
}

// GetStatusTrends 获取状态趋势数据
func (s *StatusTrackingService) GetStatusTrends(userID uint, days int) ([]model.StatusTrend, error) {
    if s.db != nil && s.db.UseGorm && s.db.ORM != nil {
        return s.getStatusTrendsGorm(userID, days)
    }
	if days <= 0 || days > 365 {
		days = 30
	}

	startDate := time.Now().AddDate(0, 0, -days)

	trendsQuery := `
		SELECT DATE(status_changed_at) as date, new_status, COUNT(*) as count
		FROM job_status_history 
		WHERE user_id = $1 AND status_changed_at >= $2
		GROUP BY DATE(status_changed_at), new_status
		ORDER BY date DESC, count DESC
	`

	rows, err := s.db.Query(trendsQuery, userID, startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get status trends: %w", err)
	}
	defer rows.Close()

	var trends []model.StatusTrend
	for rows.Next() {
		var trend model.StatusTrend
		var date time.Time
		err := rows.Scan(&date, &trend.Status, &trend.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status trend: %w", err)
		}
		trend.Date = date.Format("2006-01-02")
		trends = append(trends, trend)
	}

	return trends, nil
}

func (s *StatusTrackingService) getStatusTrendsGorm(userID uint, days int) ([]model.StatusTrend, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    if days <= 0 || days > 365 { days = 30 }
    startDate := time.Now().AddDate(0,0,-days)
    rows, err := s.db.ORM.WithContext(ctx).Raw("SELECT DATE(status_changed_at) as date, new_status, COUNT(*) as count FROM job_status_history WHERE user_id = $1 AND status_changed_at >= $2 GROUP BY DATE(status_changed_at), new_status ORDER BY date DESC, count DESC", userID, startDate).Rows()
    if err != nil { return nil, fmt.Errorf("failed to get status trends: %w", err) }
    defer rows.Close()
    var trends []model.StatusTrend
    for rows.Next() {
        var date time.Time
        var st string
        var c int
        if err := rows.Scan(&date, &st, &c); err != nil {
            return nil, fmt.Errorf("failed to scan status trend: %w", err)
        }
        trends = append(trends, model.StatusTrend{
            Date:   date.Format("2006-01-02"),
            Status: st,
            Count:  c,
        })
    }
    return trends, nil
}

// validateStatusTransition 验证状态转换合法性
func (s *StatusTrackingService) validateStatusTransition(userID uint, oldStatus, newStatus model.ApplicationStatus) error {
	// 获取用户的流转模板配置
	var flowConfig string
	templateQuery := `
		SELECT COALESCE(sft.flow_config::text, '{"transitions": {}}')
		FROM status_flow_templates sft
		WHERE sft.is_default = true AND sft.is_active = true
		LIMIT 1
	`
	err := s.db.QueryRow(templateQuery).Scan(&flowConfig)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get flow template: %w", err)
	}

	// 如果没有配置模板，允许所有转换
	if err == sql.ErrNoRows {
		return nil
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(flowConfig), &config); err != nil {
		return nil // 配置解析失败，允许转换
	}

    transitions, ok := config["transitions"].(map[string]interface{})
    if !ok {
        // 没有配置，采用内置直通规则放行
        if s.isImplicitDirectTransitionAllowed(oldStatus, newStatus) {
            return nil
        }
        return nil
    }

    allowedStates, ok := transitions[string(oldStatus)].([]interface{})
    if ok {
        // 检查新状态是否在允许列表中
        for _, allowed := range allowedStates {
            if allowedStr, ok := allowed.(string); ok && allowedStr == string(newStatus) {
                return nil
            }
        }
    }

    // 不在模板中时，允许内置直通规则
    if s.isImplicitDirectTransitionAllowed(oldStatus, newStatus) {
        return nil
    }

    return fmt.Errorf("status transition not allowed: %s -> %s", oldStatus, newStatus)
}

// isImplicitDirectTransitionAllowed 允许面试阶段的直接推进：一面中->二面中->三面中->HR面中
func (s *StatusTrackingService) isImplicitDirectTransitionAllowed(oldStatus, newStatus model.ApplicationStatus) bool {
    direct := map[model.ApplicationStatus]model.ApplicationStatus{
        model.StatusWrittenTest:      model.StatusFirstInterview,  // 笔试中 → 一面中（表示笔试通过）
        model.StatusFirstInterview:  model.StatusSecondInterview,
        model.StatusSecondInterview: model.StatusThirdInterview,
        model.StatusThirdInterview:  model.StatusHRInterview,
    }
    if next, ok := direct[oldStatus]; ok {
        return next == newStatus
    }
    return false
}

// isBackwardTransition 判断是否为回退（将状态从后往前调整）
func (s *StatusTrackingService) isBackwardTransition(oldStatus, newStatus model.ApplicationStatus) bool {
    rank := func(st model.ApplicationStatus) int {
        // 主阶段等级：忽略通过/未通过细分，聚类到阶段
        switch st {
        case model.StatusApplied:
            return 0
        case model.StatusResumeScreening, model.StatusResumeScreeningFail:
            return 10
        case model.StatusWrittenTest, model.StatusWrittenTestPass, model.StatusWrittenTestFail:
            return 20
        case model.StatusFirstInterview, model.StatusFirstPass, model.StatusFirstFail:
            return 30
        case model.StatusSecondInterview, model.StatusSecondPass, model.StatusSecondFail:
            return 40
        case model.StatusThirdInterview, model.StatusThirdPass, model.StatusThirdFail:
            return 50
        case model.StatusHRInterview, model.StatusHRPass, model.StatusHRFail:
            return 60
        case model.StatusOfferWaiting:
            return 70
        case model.StatusOfferReceived:
            return 80
        case model.StatusOfferAccepted, model.StatusRejected:
            return 90
        case model.StatusProcessFinished:
            return 100
        default:
            return 0
        }
    }
    return rank(newStatus) < rank(oldStatus)
}

// isTerminalStatus 判断是否为“终态”以用于回退必填备注
// 这里限定为：流程结束、已拒绝、各阶段未通过
func (s *StatusTrackingService) isTerminalStatus(st model.ApplicationStatus) bool {
    if st == model.StatusProcessFinished || st == model.StatusRejected {
        return true
    }
    return st == model.StatusResumeScreeningFail || st == model.StatusWrittenTestFail ||
        st == model.StatusFirstFail || st == model.StatusSecondFail ||
        st == model.StatusThirdFail || st == model.StatusHRFail
}

// updateStatusHistoryJSON 更新状态历史JSON
func (s *StatusTrackingService) updateStatusHistoryJSON(currentHistoryStr string, oldStatus, newStatus model.ApplicationStatus, changedAt time.Time, durationMinutes *int) model.StatusHistory {
	var history model.StatusHistory

	// 解析现有历史
	if currentHistoryStr != "" {
		json.Unmarshal([]byte(currentHistoryStr), &history)
	}

	// 初始化
	if history.History == nil {
		history.History = []model.StatusHistoryEntry{}
	}

	// 添加新记录
	entry := model.StatusHistoryEntry{
		OldStatus:       &oldStatus,
		NewStatus:       newStatus,
		StatusChangedAt: changedAt,
		CreatedAt:       changedAt,
	}
	if durationMinutes != nil {
		entry.DurationMinutes = durationMinutes
	}

	history.History = append(history.History, entry)

	// 更新元数据
	history.Metadata.TotalChanges = len(history.History)
	history.Metadata.CurrentStatus = string(newStatus)
	history.Metadata.LastChanged = changedAt

	// 计算总持续时间
	totalDuration := 0
	for _, h := range history.History {
		if h.DurationMinutes != nil {
			totalDuration += *h.DurationMinutes
		}
	}
	history.Metadata.TotalDurationMinutes = totalDuration

	return history
}

// updateDurationStats 更新持续时间统计
func (s *StatusTrackingService) updateDurationStats(currentStatsStr string, status model.ApplicationStatus, durationMinutes *int) model.DurationStats {
	var stats model.DurationStats

	// 解析现有统计
	if currentStatsStr != "" {
		json.Unmarshal([]byte(currentStatsStr), &stats)
	}

	// 初始化
	if stats.StatusDurations == nil {
		stats.StatusDurations = make(map[string]model.StatusDuration)
	}
	if stats.Milestones == nil {
		stats.Milestones = make(map[string]time.Time)
	}

	// 更新状态持续时间
	if durationMinutes != nil {
		statusStr := string(status)
		if existing, ok := stats.StatusDurations[statusStr]; ok {
			existing.TotalMinutes += *durationMinutes
			stats.StatusDurations[statusStr] = existing
		} else {
			stats.StatusDurations[statusStr] = model.StatusDuration{
				TotalMinutes: *durationMinutes,
			}
		}

		// 更新里程碑
		now := time.Now()
		if status == model.StatusResumeScreening {
			stats.Milestones["first_response"] = now
		} else if status.IsInProgressStatus() && strings.Contains(string(status), "面") {
			if _, exists := stats.Milestones["first_interview"]; !exists {
				stats.Milestones["first_interview"] = now
			}
		}
	}

	// 重新计算百分比
	totalMinutes := 0
	for _, duration := range stats.StatusDurations {
		totalMinutes += duration.TotalMinutes
	}

	if totalMinutes > 0 {
		for statusStr, duration := range stats.StatusDurations {
			duration.Percentage = float64(duration.TotalMinutes) / float64(totalMinutes) * 100
			stats.StatusDurations[statusStr] = duration
		}
	}

	return stats
}
