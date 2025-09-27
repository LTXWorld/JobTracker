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
	"os"
	"strings"
	"time"

	"jobView-backend/internal/model"
	"jobView-backend/internal/repository"
)

type StatusTrackingService struct {
	repo       repository.StatusTrackingRepository
	configRepo repository.StatusConfigRepository
}

func NewStatusTrackingService(repo repository.StatusTrackingRepository, configRepo repository.StatusConfigRepository) *StatusTrackingService {
	return &StatusTrackingService{repo: repo, configRepo: configRepo}
}

func (s *StatusTrackingService) ensureRepo() error {
	if s.repo == nil {
		return fmt.Errorf("status tracking repository not initialized")
	}
	return nil
}

// GetStatusHistory 获取岗位状态历史记录
func (s *StatusTrackingService) GetStatusHistory(userID uint, jobApplicationID int, page, pageSize int) (*model.StatusHistoryResponse, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	total, history, err := s.repo.GetStatusHistory(ctx, userID, jobApplicationID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &model.StatusHistoryResponse{
		History:     history,
		Total:       total,
		CurrentPage: page,
		PageSize:    pageSize,
	}, nil
}

func (s *StatusTrackingService) UpdateJobStatus(userID uint, jobApplicationID int, request *model.StatusUpdateRequest) (*model.JobApplication, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	if !request.Status.IsValid() {
		return nil, fmt.Errorf("invalid status: %s", request.Status)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	snapshot, err := tx.GetJobApplicationForUpdate(userID, jobApplicationID)
	if err != nil {
		return nil, err
	}

	if request.Version != nil && snapshot.StatusVersion != nil {
		if int32(*request.Version) != int32(*snapshot.StatusVersion) {
			return nil, fmt.Errorf("version conflict: expected %d, got %d", *snapshot.StatusVersion, *request.Version)
		}
	}

	if snapshot.Job.Status == request.Status {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}
		return &snapshot.Job, nil
	}

	validateErr := s.validateStatusTransition(userID, snapshot.Job.Status, request.Status)
	isBackward := s.isBackwardTransition(snapshot.Job.Status, request.Status)
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
			if s.isTerminalStatus(snapshot.Job.Status) {
				if request.Note == nil || strings.TrimSpace(*request.Note) == "" {
					return nil, fmt.Errorf("NOTE_REQUIRED_FOR_BACKWARD")
				}
			}
		} else {
			return nil, validateErr
		}
	}

	if err := tx.SetLocalFlag("jobview.skip_history", true); err != nil {
		return nil, fmt.Errorf("failed to set session flag: %w", err)
	}
	suppressHistory := isBackward && request.ConfirmBackward != nil && *request.ConfirmBackward
	if suppressHistory {
		if err := tx.SetLocalFlag("jobview.allow_backward", true); err != nil {
			return nil, fmt.Errorf("failed to enable backward flag: %w", err)
		}
	}

	now := time.Now()
	var durationMinutes *int
	if snapshot.LastStatusChange != nil {
		d := int(now.Sub(*snapshot.LastStatusChange).Minutes())
		durationMinutes = &d
	}

	useDBTriggerForHistory := false
	var statusHistoryBytes []byte
	var durationStatsBytes []byte
	if !useDBTriggerForHistory && !suppressHistory {
		metadata := make(map[string]interface{})
		if request.Metadata != nil {
			for k, v := range request.Metadata {
				metadata[k] = v
			}
		}
		if isBackward {
			metadata["backward"] = true
			metadata["from"] = string(snapshot.Job.Status)
			metadata["to"] = string(request.Status)
			if request.Note != nil && strings.TrimSpace(*request.Note) != "" {
				metadata["note"] = strings.TrimSpace(*request.Note)
			}
		}
		metadataBytes, _ := json.Marshal(metadata)
		_, err = tx.InsertStatusHistory(repository.StatusHistoryInsert{
			JobApplicationID: jobApplicationID,
			UserID:           userID,
			OldStatus:        snapshot.Job.Status,
			NewStatus:        request.Status,
			ChangedAt:        now,
			DurationMinutes:  durationMinutes,
			Metadata:         metadataBytes,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to insert status history: %w", err)
		}

		currentHistory := ""
		if snapshot.StatusHistoryRaw != nil {
			currentHistory = *snapshot.StatusHistoryRaw
		}
		history := s.updateStatusHistoryJSON(currentHistory, snapshot.Job.Status, request.Status, now, durationMinutes)
		statusHistoryBytes, _ = json.Marshal(history)

		currentStats := ""
		if snapshot.DurationStatsRaw != nil {
			currentStats = *snapshot.DurationStatsRaw
		}
		stats := s.updateDurationStats(currentStats, snapshot.Job.Status, durationMinutes)
		durationStatsBytes, _ = json.Marshal(stats)
	}

	newVersion := 1
	if snapshot.StatusVersion != nil {
		newVersion = *snapshot.StatusVersion + 1
	}

	params := repository.UpdateJobApplicationParams{
		JobApplicationID:  jobApplicationID,
		UserID:            userID,
		NewStatus:         request.Status,
		Now:               now,
		LastStatusChange:  &now,
		StatusVersion:     &newVersion,
		StatusHistoryJSON: statusHistoryBytes,
		DurationStatsJSON: durationStatsBytes,
		SuppressHistory:   suppressHistory,
		UseTrigger:        useDBTriggerForHistory,
	}
	if suppressHistory || useDBTriggerForHistory {
		params.LastStatusChange = nil
		params.StatusHistoryJSON = nil
		params.DurationStatsJSON = nil
		params.StatusVersion = nil
	}

	updatedJob, err := tx.UpdateJobApplication(params)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return updatedJob, nil
}

func (s *StatusTrackingService) GetStatusTimeline(userID uint, jobApplicationID int) (map[string]interface{}, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	entries, err := s.repo.GetStatusTimeline(ctx, userID, jobApplicationID)
	if err != nil {
		return nil, err
	}

	totalDuration := 0
	timeline := make([]map[string]interface{}, 0, len(entries))
	for _, entry := range entries {
		row := map[string]interface{}{
			"new_status":        entry.NewStatus,
			"status_changed_at": entry.StatusChangedAt,
		}
		if entry.OldStatus != nil {
			row["old_status"] = string(*entry.OldStatus)
		}
		if entry.DurationMinutes != nil {
			row["duration_minutes"] = *entry.DurationMinutes
			totalDuration += *entry.DurationMinutes
		}
		if entry.Metadata != nil {
			row["metadata"] = entry.Metadata
		}
		timeline = append(timeline, row)
	}

	return map[string]interface{}{
		"job_application_id":     jobApplicationID,
		"timeline":               timeline,
		"total_duration_minutes": totalDuration,
		"total_changes":          len(timeline),
	}, nil
}

func (s *StatusTrackingService) BatchUpdateStatus(userID uint, updates []model.BatchStatusUpdate) error {
	if err := s.ensureRepo(); err != nil {
		return err
	}
	if len(updates) == 0 {
		return nil
	}
	if len(updates) > 100 {
		return fmt.Errorf("batch size too large: maximum 100 updates allowed")
	}
	for _, update := range updates {
		if !update.Status.IsValid() {
			return fmt.Errorf("invalid status: %s for ID %d", update.Status, update.ID)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	for _, update := range updates {
		currentStatus, lastChange, err := tx.GetCurrentStatus(userID, update.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return err
		}
		if currentStatus == update.Status {
			continue
		}
		if s.isBackwardTransition(currentStatus, update.Status) {
			return fmt.Errorf("backward transitions are not allowed in batch updates (ID %d: %s -> %s)", update.ID, currentStatus, update.Status)
		}
		if err := s.validateStatusTransition(userID, currentStatus, update.Status); err != nil {
			return fmt.Errorf("invalid transition for ID %d: %w", update.ID, err)
		}

		var durationMinutes *int
		if lastChange != nil {
			d := int(now.Sub(*lastChange).Minutes())
			durationMinutes = &d
		}

		if _, err := tx.InsertStatusHistory(repository.StatusHistoryInsert{
			JobApplicationID: update.ID,
			UserID:           userID,
			OldStatus:        currentStatus,
			NewStatus:        update.Status,
			ChangedAt:        now,
			DurationMinutes:  durationMinutes,
			Metadata:         []byte("{}"),
		}); err != nil {
			return fmt.Errorf("failed to insert history for ID %d: %w", update.ID, err)
		}

		params := repository.UpdateJobApplicationParams{
			JobApplicationID: update.ID,
			UserID:           userID,
			NewStatus:        update.Status,
			Now:              now,
			LastStatusChange: &now,
			IncrementVersion: true,
		}
		if _, err := tx.UpdateJobApplication(params); err != nil {
			return fmt.Errorf("failed to update status for ID %d: %w", update.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (s *StatusTrackingService) GetStatusAnalytics(userID uint) (*model.StatusAnalyticsResponse, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	analytics, err := s.repo.GetStatusAnalytics(ctx, userID)
	if err != nil {
		return nil, err
	}
	return analytics, nil
}

func (s *StatusTrackingService) GetStatusTrends(userID uint, days int) ([]model.StatusTrend, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	if days <= 0 || days > 365 {
		days = 30
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	trends, err := s.repo.GetStatusTrends(ctx, userID, days)
	if err != nil {
		return nil, err
	}
	return trends, nil
}

func (s *StatusTrackingService) validateStatusTransition(userID uint, oldStatus, newStatus model.ApplicationStatus) error {
	if s.configRepo == nil {
		return nil
	}
	_, flowConfig, err := s.configRepo.GetDefaultFlowTemplate()
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get flow template: %w", err)
	}
	if err == sql.ErrNoRows {
		return nil
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(flowConfig), &config); err != nil {
		return nil
	}
	transitions, ok := config["transitions"].(map[string]interface{})
	if !ok {
		if s.isImplicitDirectTransitionAllowed(oldStatus, newStatus) || s.isImplicitFailTransitionAllowed(oldStatus, newStatus) {
			return nil
		}
		return nil
	}

	if allowedStates, ok := transitions[string(oldStatus)].([]interface{}); ok {
		for _, allowed := range allowedStates {
			if allowedStr, ok := allowed.(string); ok && allowedStr == string(newStatus) {
				return nil
			}
		}
	}

	if s.isImplicitDirectTransitionAllowed(oldStatus, newStatus) || s.isImplicitFailTransitionAllowed(oldStatus, newStatus) {
		return nil
	}
	return fmt.Errorf("transition from %s to %s is not allowed", oldStatus, newStatus)
}

func (s *StatusTrackingService) isImplicitDirectTransitionAllowed(oldStatus, newStatus model.ApplicationStatus) bool {
	direct := map[model.ApplicationStatus]model.ApplicationStatus{
		model.StatusWrittenTest:     model.StatusFirstInterview, // 笔试中 → 一面中（表示笔试通过）
		model.StatusFirstInterview:  model.StatusSecondInterview,
		model.StatusSecondInterview: model.StatusThirdInterview,
		model.StatusThirdInterview:  model.StatusHRInterview,
	}
	if next, ok := direct[oldStatus]; ok {
		return next == newStatus
	}
	return false
}

// isImplicitFailTransitionAllowed 允许同阶段“进行中”到“未通过”的失败流转
func (s *StatusTrackingService) isImplicitFailTransitionAllowed(oldStatus, newStatus model.ApplicationStatus) bool {
	switch oldStatus {
	case model.StatusResumeScreening:
		return newStatus == model.StatusResumeScreeningFail
	case model.StatusWrittenTest:
		return newStatus == model.StatusWrittenTestFail
	case model.StatusFirstInterview:
		return newStatus == model.StatusFirstFail
	case model.StatusSecondInterview:
		return newStatus == model.StatusSecondFail
	case model.StatusThirdInterview:
		return newStatus == model.StatusThirdFail
	case model.StatusHRInterview:
		return newStatus == model.StatusHRFail
	default:
		return false
	}
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
