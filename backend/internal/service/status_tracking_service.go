// Location: /Users/lutao/GolandProjects/jobView/backend/internal/service/status_tracking_service.go
// This file implements the core status tracking service for JobView system.
// It handles job application status history, transitions, analytics, and preference management.
// Used by the status tracking handlers to provide business logic and data processing.

package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"jobView-backend/internal/model"
	"jobView-backend/internal/repository"
	"jobView-backend/internal/utils"
)

type StatusTrackingService struct {
	repo       repository.StatusTrackingRepository
	configRepo repository.StatusConfigRepository
}

var (
	ErrUndoHistoryNotFound = errors.New("UNDO_HISTORY_NOT_FOUND")
	ErrUndoExpired         = errors.New("UNDO_EXPIRED")
	ErrUndoVersionMismatch = errors.New("UNDO_VERSION_CONFLICT")
	ErrUndoStatusMismatch  = errors.New("UNDO_STATUS_MISMATCH")
	ErrUndoInvalidTarget   = errors.New("UNDO_INVALID_TARGET")
)

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

func (s *StatusTrackingService) UpdateJobStatus(userID uint, jobApplicationID int, request *model.StatusUpdateRequest) (*model.StatusUpdateResult, error) {
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
		result := &model.StatusUpdateResult{
			Job: &snapshot.Job,
		}
		if snapshot.StatusVersion != nil {
			result.StatusVersion = snapshot.StatusVersion
			result.Job.StatusVersion = snapshot.StatusVersion
		}
		return result, nil
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
	allowBackwardFlag := isBackward && request.ConfirmBackward != nil && *request.ConfirmBackward
	if allowBackwardFlag {
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
	var insertedHistoryID *int64
	var experienceInsert *repository.InterviewExperienceInsert
	if !useDBTriggerForHistory {
		metadata := make(map[string]interface{})
		if request.Metadata != nil {
			for k, v := range request.Metadata {
				if k == "interview_experience" {
					continue
				}
				metadata[k] = v
			}
		}
		if s.shouldCaptureInterviewExperience(snapshot.Job.Status) {
			interviewMeta, insertPayload, err := s.prepareInterviewExperience(request.Metadata, userID, jobApplicationID, snapshot.Job.Status, request.Status, now)
			if err != nil {
				return nil, err
			}
			if len(interviewMeta) > 0 {
				metadata["interview_experience"] = interviewMeta
			}
			experienceInsert = insertPayload
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
		historyID, insertErr := tx.InsertStatusHistory(repository.StatusHistoryInsert{
			JobApplicationID: jobApplicationID,
			UserID:           userID,
			OldStatus:        snapshot.Job.Status,
			NewStatus:        request.Status,
			ChangedAt:        now,
			DurationMinutes:  durationMinutes,
			Metadata:         metadataBytes,
		})
		if insertErr != nil {
			return nil, fmt.Errorf("failed to insert status history: %w", insertErr)
		}
		insertedHistoryID = &historyID
		if experienceInsert != nil {
			if _, err := tx.InsertInterviewExperience(*experienceInsert); err != nil {
				return nil, fmt.Errorf("failed to insert interview experience: %w", err)
			}
		}

		currentHistory := ""
		if snapshot.StatusHistoryRaw != nil {
			currentHistory = *snapshot.StatusHistoryRaw
		}
		history := s.updateStatusHistoryJSON(currentHistory, snapshot.Job.Status, request.Status, now, durationMinutes, metadata)
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
		SuppressHistory:   false,
		UseTrigger:        useDBTriggerForHistory,
	}
	if useDBTriggerForHistory {
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

	resultVersion := updatedJob.StatusVersion
	if params.StatusVersion != nil {
		versionCopy := newVersion
		resultVersion = &versionCopy
		updatedJob.StatusVersion = resultVersion
	}

	result := &model.StatusUpdateResult{
		Job:           updatedJob,
		StatusVersion: resultVersion,
	}
	if insertedHistoryID != nil {
		result.HistoryID = insertedHistoryID
		deadline := now.Add(s.undoWindowDuration())
		result.UndoAvailableUntil = &deadline
		result.Metadata = map[string]interface{}{
			"undo_window_seconds": int(s.undoWindowDuration().Seconds()),
		}
	}

	return result, nil
}

func (s *StatusTrackingService) UndoJobStatus(userID uint, jobApplicationID int, request *model.StatusUndoRequest) (*model.StatusUndoResult, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	if request == nil {
		return nil, fmt.Errorf("%w: missing request body", ErrUndoInvalidTarget)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	if err := tx.SetLocalFlag("jobview.skip_history", true); err != nil {
		return nil, fmt.Errorf("failed to set session flag: %w", err)
	}
	if err := tx.SetLocalFlag("jobview.allow_backward", true); err != nil {
		return nil, fmt.Errorf("failed to enable backward flag: %w", err)
	}

	snapshot, err := tx.GetJobApplicationForUpdate(userID, jobApplicationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: job application not found", ErrUndoInvalidTarget)
		}
		return nil, fmt.Errorf("failed to load job application: %w", err)
	}

	if snapshot.StatusVersion != nil && request.Version == nil {
		return nil, fmt.Errorf("%w: version required", ErrUndoVersionMismatch)
	}
	if snapshot.StatusVersion != nil && request.Version != nil && int32(*snapshot.StatusVersion) != int32(*request.Version) {
		return nil, fmt.Errorf("%w: expected %d but got %d", ErrUndoVersionMismatch, *snapshot.StatusVersion, *request.Version)
	}

	latest, err := tx.GetLatestHistoryEntry(userID, jobApplicationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUndoHistoryNotFound
		}
		return nil, fmt.Errorf("failed to fetch latest history: %w", err)
	}
	if latest == nil {
		return nil, ErrUndoHistoryNotFound
	}
	if latest.OldStatus == nil {
		return nil, fmt.Errorf("%w: missing previous status", ErrUndoInvalidTarget)
	}
	if latest.NewStatus != snapshot.Job.Status {
		return nil, fmt.Errorf("%w: current status %s does not match history %s", ErrUndoStatusMismatch, snapshot.Job.Status, latest.NewStatus)
	}
	if request.HistoryID != 0 && request.HistoryID != latest.ID {
		return nil, fmt.Errorf("%w: history mismatch", ErrUndoStatusMismatch)
	}

	window := s.undoWindowDuration()
	if time.Since(latest.StatusChangedAt) > window {
		return nil, ErrUndoExpired
	}

	now := time.Now()
	var durationMinutes *int
	if snapshot.LastStatusChange != nil {
		delta := int(now.Sub(*snapshot.LastStatusChange).Minutes())
		if delta < 0 {
			delta = 0
		}
		durationMinutes = &delta
	}

	undoMetadata := map[string]interface{}{
		"undo":          true,
		"undo_by":       userID,
		"undo_source":   latest.ID,
		"undo_at":       now,
		"reverted_to":   string(*latest.OldStatus),
		"undo_window":   int(window.Seconds()),
		"undo_deadline": latest.StatusChangedAt.Add(window),
	}
	undoMetadataBytes, _ := json.Marshal(undoMetadata)
	undoHistoryID, err := tx.InsertStatusHistory(repository.StatusHistoryInsert{
		JobApplicationID: jobApplicationID,
		UserID:           userID,
		OldStatus:        snapshot.Job.Status,
		NewStatus:        *latest.OldStatus,
		ChangedAt:        now,
		DurationMinutes:  durationMinutes,
		Metadata:         undoMetadataBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to insert undo history: %w", err)
	}

	originalMetadata := make(map[string]interface{})
	if latest.Metadata != nil {
		for k, v := range latest.Metadata {
			originalMetadata[k] = v
		}
	}
	originalMetadata["undo_applied"] = true
	originalMetadata["undo_by"] = userID
	originalMetadata["undo_at"] = now
	originalMetadata["undo_history_id"] = undoHistoryID
	originalMetadataBytes, _ := json.Marshal(originalMetadata)
	if err := tx.UpdateHistoryMetadata(latest.ID, originalMetadataBytes); err != nil {
		return nil, fmt.Errorf("failed to update source history metadata: %w", err)
	}

	currentHistory := ""
	if snapshot.StatusHistoryRaw != nil {
		currentHistory = *snapshot.StatusHistoryRaw
	}
	history := s.updateStatusHistoryJSON(currentHistory, snapshot.Job.Status, *latest.OldStatus, now, durationMinutes, undoMetadata)
	statusHistoryBytes, _ := json.Marshal(history)

	currentStats := ""
	if snapshot.DurationStatsRaw != nil {
		currentStats = *snapshot.DurationStatsRaw
	}
	stats := s.updateDurationStats(currentStats, snapshot.Job.Status, durationMinutes)
	durationStatsBytes, _ := json.Marshal(stats)

	var newVersionPtr *int
	if snapshot.StatusVersion != nil {
		v := *snapshot.StatusVersion - 1
		if v < 0 {
			v = 0
		}
		newVersionPtr = &v
	}

	params := repository.UpdateJobApplicationParams{
		JobApplicationID:  jobApplicationID,
		UserID:            userID,
		NewStatus:         *latest.OldStatus,
		Now:               now,
		LastStatusChange:  &now,
		StatusVersion:     newVersionPtr,
		StatusHistoryJSON: statusHistoryBytes,
		DurationStatsJSON: durationStatsBytes,
	}

	updatedJob, err := tx.UpdateJobApplication(params)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit undo transaction: %w", err)
	}

	if newVersionPtr != nil {
		updatedJob.StatusVersion = newVersionPtr
	}

	return &model.StatusUndoResult{
		Job:             updatedJob,
		RevertedTo:      *latest.OldStatus,
		UndoHistoryID:   undoHistoryID,
		SourceHistoryID: latest.ID,
		StatusVersion:   newVersionPtr,
		UndoCompletedAt: now,
	}, nil
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

func (s *StatusTrackingService) GetInterviewExperiences(userID uint, jobApplicationID int) ([]model.InterviewExperience, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	experiences, err := s.repo.GetInterviewExperiences(ctx, userID, jobApplicationID)
	if err != nil {
		return nil, err
	}
	return experiences, nil
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

func (s *StatusTrackingService) GetStatusAnalytics(userID uint, processType model.ApplicationProcessType) (*model.StatusAnalyticsResponse, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	if !processType.IsValid() {
		processType = model.ProcessTypeAutumn
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	pt := processType
	analytics, err := s.repo.GetStatusAnalytics(ctx, userID, &pt)
	if err != nil {
		return nil, err
	}
	return analytics, nil
}

func (s *StatusTrackingService) GetStatusTrends(userID uint, days int, processType model.ApplicationProcessType) ([]model.StatusTrend, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	if days <= 0 || days > 365 {
		days = 30
	}
	if !processType.IsValid() {
		processType = model.ProcessTypeAutumn
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	pt := processType
	trends, err := s.repo.GetStatusTrends(ctx, userID, days, &pt)
	if err != nil {
		return nil, err
	}
	return trends, nil
}

func (s *StatusTrackingService) validateStatusTransition(userID uint, oldStatus, newStatus model.ApplicationStatus) error {
	if s.configRepo == nil {
		return nil
	}

	// 允许从任何状态转换到用户主动拒绝状态
	if newStatus == model.StatusUserRejected {
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
	direct := map[model.ApplicationStatus][]model.ApplicationStatus{
		model.StatusWrittenTest: {
			model.StatusFirstInterview, // 笔试中 → 一面中（表示笔试通过）
		},
		model.StatusFirstInterview: {
			model.StatusSecondInterview,
			model.StatusHRInterview,
		},
		model.StatusSecondInterview: {
			model.StatusThirdInterview,
			model.StatusHRInterview,
		},
		model.StatusThirdInterview: {
			model.StatusHRInterview,
		},
		model.StatusHRInterview: {
			model.StatusHRPass,
		},
		model.StatusHRPass: {
			model.StatusOfferAccepted,
			model.StatusRejected,
		},
	}

	if nextStates, ok := direct[oldStatus]; ok {
		for _, candidate := range nextStates {
			if candidate == newStatus {
				return true
			}
		}
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
		case model.StatusOfferAccepted, model.StatusRejected, model.StatusUserRejected:
			return 70
		default:
			return 0
		}
	}
	return rank(newStatus) < rank(oldStatus)
}

// isTerminalStatus 判断是否为"终态"以用于回退必填备注
// 这里限定为：已接受offer、已拒绝offer、已拒绝、各阶段未通过
func (s *StatusTrackingService) isTerminalStatus(st model.ApplicationStatus) bool {
	if st == model.StatusOfferAccepted || st == model.StatusRejected || st == model.StatusUserRejected {
		return true
	}
	return st == model.StatusResumeScreeningFail || st == model.StatusWrittenTestFail ||
		st == model.StatusFirstFail || st == model.StatusSecondFail ||
		st == model.StatusThirdFail || st == model.StatusHRFail
}

func (s *StatusTrackingService) undoWindowDuration() time.Duration {
	if value := os.Getenv("STATUS_UNDO_WINDOW_SECONDS"); value != "" {
		if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	return 10 * time.Second
}

var interviewStageStatuses = map[model.ApplicationStatus]struct{}{
	model.StatusFirstInterview:  {},
	model.StatusSecondInterview: {},
	model.StatusThirdInterview:  {},
	model.StatusHRInterview:     {},
}

func (s *StatusTrackingService) shouldCaptureInterviewExperience(status model.ApplicationStatus) bool {
	_, ok := interviewStageStatuses[status]
	return ok
}

func (s *StatusTrackingService) prepareInterviewExperience(rawMetadata map[string]interface{}, userID uint, applicationID int, fromStatus, toStatus model.ApplicationStatus, now time.Time) (map[string]interface{}, *repository.InterviewExperienceInsert, error) {
	experience := &repository.InterviewExperienceInsert{
		ApplicationID: applicationID,
		UserID:        userID,
		FromStatus:    fromStatus,
		ToStatus:      toStatus,
		Skip:          true,
		RecordedAt:    now,
	}

	sanitized := map[string]interface{}{
		"skip":        true,
		"recorded_by": userID,
		"recorded_at": now.UTC().Format(time.RFC3339),
		"from_status": string(fromStatus),
		"to_status":   string(toStatus),
	}

	if rawMetadata == nil {
		return sanitized, experience, nil
	}

	rawExperience, ok := rawMetadata["interview_experience"]
	if !ok || rawExperience == nil {
		return sanitized, experience, nil
	}

	experienceMap, ok := rawExperience.(map[string]interface{})
	if !ok {
		return nil, nil, utils.NewValidationError("metadata.interview_experience", "interview_experience 必须是对象")
	}

	skip := false
	if rawSkip, exists := experienceMap["skip"]; exists {
		parsedSkip, err := castToBool(rawSkip)
		if err != nil {
			return nil, nil, utils.NewValidationError("metadata.interview_experience.skip", "skip 字段值无效")
		}
		skip = parsedSkip
	}
	experience.Skip = skip
	sanitized["skip"] = skip

	recordedAt := now
	if rawRecordedAt, exists := experienceMap["recorded_at"]; exists && rawRecordedAt != nil {
		parsedAt, err := parseInterviewRecordedAt(rawRecordedAt)
		if err != nil {
			return nil, nil, utils.NewValidationError("metadata.interview_experience.recorded_at", "recorded_at 需符合 RFC3339 格式")
		}
		recordedAt = parsedAt
	}
	experience.RecordedAt = recordedAt
	sanitized["recorded_at"] = recordedAt.UTC().Format(time.RFC3339)
	sanitized["recorded_by"] = userID
	sanitized["from_status"] = string(fromStatus)
	sanitized["to_status"] = string(toStatus)

	if !skip {
		rawRating, exists := experienceMap["rating"]
		if !exists || rawRating == nil {
			return nil, nil, utils.NewValidationError("metadata.interview_experience.rating", "缺失评价")
		}
		rating, err := extractString(rawRating)
		if err != nil {
			return nil, nil, utils.NewValidationError("metadata.interview_experience.rating", "rating 必须为字符串")
		}
		rating = strings.ToLower(strings.TrimSpace(rating))
		if !isValidInterviewRating(rating) {
			return nil, nil, utils.NewValidationError("metadata.interview_experience.rating", "rating 仅支持 good/average/bad")
		}
		sanitized["rating"] = rating
		ratingCopy := rating
		experience.Rating = &ratingCopy
	} else if rawRating, exists := experienceMap["rating"]; exists && rawRating != nil {
		rating, err := extractString(rawRating)
		if err == nil && rating != "" {
			rating = strings.ToLower(strings.TrimSpace(rating))
			if isValidInterviewRating(rating) {
				sanitized["rating"] = rating
				ratingCopy := rating
				experience.Rating = &ratingCopy
			}
		}
	}

	if rawNote, exists := experienceMap["note"]; exists && rawNote != nil {
		note, err := extractString(rawNote)
		if err != nil {
			return nil, nil, utils.NewValidationError("metadata.interview_experience.note", "note 必须为字符串")
		}
		note = utils.SanitizeInput(note)
		if len([]rune(note)) > 200 {
			return nil, nil, utils.NewValidationError("metadata.interview_experience.note", "note 最长 200 字符")
		}
		if note != "" {
			sanitized["note"] = note
			noteCopy := note
			experience.Note = &noteCopy
		}
	}

	if rawReason, exists := experienceMap["skip_reason"]; exists && rawReason != nil {
		reason, err := extractString(rawReason)
		if err != nil {
			return nil, nil, utils.NewValidationError("metadata.interview_experience.skip_reason", "skip_reason 必须为字符串")
		}
		reason = utils.SanitizeInput(reason)
		if len([]rune(reason)) > 200 {
			return nil, nil, utils.NewValidationError("metadata.interview_experience.skip_reason", "skip_reason 最长 200 字符")
		}
		if reason != "" {
			sanitized["skip_reason"] = reason
			reasonCopy := reason
			experience.SkipReason = &reasonCopy
		}
	}

	return sanitized, experience, nil
}

func castToBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		trimmed := strings.TrimSpace(strings.ToLower(v))
		switch trimmed {
		case "true", "1", "yes", "y":
			return true, nil
		case "false", "0", "no", "n":
			return false, nil
		default:
			return false, fmt.Errorf("invalid boolean string")
		}
	case float64:
		if v == 1 {
			return true, nil
		}
		if v == 0 {
			return false, nil
		}
		return false, fmt.Errorf("invalid boolean number")
	default:
		return false, fmt.Errorf("unsupported boolean type")
	}
}

func extractString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v), nil
	case fmt.Stringer:
		return strings.TrimSpace(v.String()), nil
	default:
		return "", fmt.Errorf("value is not a string")
	}
}

func parseInterviewRecordedAt(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return time.Time{}, fmt.Errorf("empty time")
		}
		if ts, err := time.Parse(time.RFC3339, trimmed); err == nil {
			return ts, nil
		}
		if ts, err := time.Parse(time.RFC3339Nano, trimmed); err == nil {
			return ts, nil
		}
		return time.Time{}, fmt.Errorf("invalid time format")
	case time.Time:
		return v, nil
	default:
		return time.Time{}, fmt.Errorf("unsupported time type")
	}
}

func isValidInterviewRating(rating string) bool {
	switch rating {
	case "good", "average", "bad":
		return true
	default:
		return false
	}
}

func cloneStatusMetadata(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return nil
	}
	cloned := make(map[string]interface{}, len(src))
	for k, v := range src {
		switch nested := v.(type) {
		case map[string]interface{}:
			cloned[k] = cloneStatusMetadata(nested)
		case []interface{}:
			copied := make([]interface{}, len(nested))
			for i, item := range nested {
				if child, ok := item.(map[string]interface{}); ok {
					copied[i] = cloneStatusMetadata(child)
				} else {
					copied[i] = item
				}
			}
			cloned[k] = copied
		default:
			cloned[k] = nested
		}
	}
	return cloned
}

// updateStatusHistoryJSON 更新状态历史JSON
func (s *StatusTrackingService) updateStatusHistoryJSON(currentHistoryStr string, oldStatus, newStatus model.ApplicationStatus, changedAt time.Time, durationMinutes *int, metadata map[string]interface{}) model.StatusHistory {
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
	if metadata != nil && len(metadata) > 0 {
		entry.Metadata = cloneStatusMetadata(metadata)
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
