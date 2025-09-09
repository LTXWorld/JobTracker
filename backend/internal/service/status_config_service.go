// Location: /Users/lutao/GolandProjects/jobView/backend/internal/service/status_config_service.go
// This file implements status flow template and user preference management service.
// It handles creation, updating, and retrieval of status transition rules and user customizations.
// Used by configuration management handlers to provide business logic for status flow management.

package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"jobView-backend/internal/database"
	"jobView-backend/internal/model"
	"time"
)

type StatusConfigService struct {
	db *database.DB
}

func NewStatusConfigService(db *database.DB) *StatusConfigService {
	return &StatusConfigService{db: db}
}

// GetStatusFlowTemplates 获取状态流转模板列表
func (s *StatusConfigService) GetStatusFlowTemplates(userID uint) ([]model.StatusFlowTemplate, error) {
	query := `
		SELECT id, name, description, flow_config, is_default, is_active, 
		       created_by, created_at, updated_at
		FROM status_flow_templates 
		WHERE is_active = true AND (created_by IS NULL OR created_by = $1)
		ORDER BY is_default DESC, name ASC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow templates: %w", err)
	}
	defer rows.Close()

	var templates []model.StatusFlowTemplate
	for rows.Next() {
		var template model.StatusFlowTemplate
		var flowConfigBytes []byte
		var description sql.NullString
		var createdBy sql.NullInt64

		err := rows.Scan(
			&template.ID,
			&template.Name,
			&description,
			&flowConfigBytes,
			&template.IsDefault,
			&template.IsActive,
			&createdBy,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan flow template: %w", err)
		}

		// 处理可选字段
		if description.Valid {
			template.Description = &description.String
		}
		if createdBy.Valid {
			userIDUint := uint(createdBy.Int64)
			template.CreatedBy = &userIDUint
		}

		// 解析flow_config
		if len(flowConfigBytes) > 0 {
			var flowConfig map[string]interface{}
			if err := json.Unmarshal(flowConfigBytes, &flowConfig); err == nil {
				template.FlowConfig = flowConfig
			}
		}

		templates = append(templates, template)
	}

	return templates, nil
}

// CreateStatusFlowTemplate 创建自定义状态流转模板
func (s *StatusConfigService) CreateStatusFlowTemplate(userID uint, name, description string, flowConfig map[string]interface{}) (*model.StatusFlowTemplate, error) {
	// 验证名称唯一性
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM status_flow_templates WHERE name = $1 AND is_active = true)"
	err := s.db.QueryRow(checkQuery, name).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check template name uniqueness: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("template name '%s' already exists", name)
	}

	// 验证流转配置格式
	if err := s.validateFlowConfig(flowConfig); err != nil {
		return nil, fmt.Errorf("invalid flow config: %w", err)
	}

	flowConfigBytes, err := json.Marshal(flowConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal flow config: %w", err)
	}

	insertQuery := `
		INSERT INTO status_flow_templates (name, description, flow_config, created_by, is_active)
		VALUES ($1, $2, $3, $4, true)
		RETURNING id, created_at, updated_at
	`

	var template model.StatusFlowTemplate
	var desc *string
	if description != "" {
		desc = &description
	}

	err = s.db.QueryRow(insertQuery, name, desc, flowConfigBytes, userID).Scan(
		&template.ID,
		&template.CreatedAt,
		&template.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create flow template: %w", err)
	}

	// 填充返回数据
	template.Name = name
	template.Description = desc
	template.FlowConfig = flowConfig
	template.IsDefault = false
	template.IsActive = true
	template.CreatedBy = &userID

	return &template, nil
}

// UpdateStatusFlowTemplate 更新状态流转模板
func (s *StatusConfigService) UpdateStatusFlowTemplate(userID uint, templateID int, name, description string, flowConfig map[string]interface{}) (*model.StatusFlowTemplate, error) {
	// 检查权限 - 只能更新自己创建的模板
	var createdBy sql.NullInt64
	var isDefault bool
	checkQuery := "SELECT created_by, is_default FROM status_flow_templates WHERE id = $1 AND is_active = true"
	err := s.db.QueryRow(checkQuery, templateID).Scan(&createdBy, &isDefault)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("failed to check template permissions: %w", err)
	}

	if isDefault {
		return nil, fmt.Errorf("cannot modify default template")
	}

	if !createdBy.Valid || uint(createdBy.Int64) != userID {
		return nil, fmt.Errorf("permission denied: can only modify your own templates")
	}

	// 验证名称唯一性（排除当前模板）
	var exists bool
	checkNameQuery := "SELECT EXISTS(SELECT 1 FROM status_flow_templates WHERE name = $1 AND id != $2 AND is_active = true)"
	err = s.db.QueryRow(checkNameQuery, name, templateID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check template name uniqueness: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("template name '%s' already exists", name)
	}

	// 验证流转配置格式
	if err := s.validateFlowConfig(flowConfig); err != nil {
		return nil, fmt.Errorf("invalid flow config: %w", err)
	}

	flowConfigBytes, err := json.Marshal(flowConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal flow config: %w", err)
	}

	updateQuery := `
		UPDATE status_flow_templates 
		SET name = $1, description = $2, flow_config = $3, updated_at = $4
		WHERE id = $5 AND created_by = $6
		RETURNING id, name, description, flow_config, is_default, is_active, 
		          created_by, created_at, updated_at
	`

	var template model.StatusFlowTemplate
	var desc sql.NullString
	var flowConfigBytesResult []byte
	var createdByResult sql.NullInt64

	var descParam *string
	if description != "" {
		descParam = &description
	}

	err = s.db.QueryRow(updateQuery, name, descParam, flowConfigBytes, time.Now(), templateID, userID).Scan(
		&template.ID,
		&template.Name,
		&desc,
		&flowConfigBytesResult,
		&template.IsDefault,
		&template.IsActive,
		&createdByResult,
		&template.CreatedAt,
		&template.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update flow template: %w", err)
	}

	// 处理可选字段
	if desc.Valid {
		template.Description = &desc.String
	}
	if createdByResult.Valid {
		userIDUint := uint(createdByResult.Int64)
		template.CreatedBy = &userIDUint
	}

	// 解析flow_config
	if len(flowConfigBytesResult) > 0 {
		json.Unmarshal(flowConfigBytesResult, &template.FlowConfig)
	}

	return &template, nil
}

// DeleteStatusFlowTemplate 删除状态流转模板（软删除）
func (s *StatusConfigService) DeleteStatusFlowTemplate(userID uint, templateID int) error {
	// 检查权限
	var createdBy sql.NullInt64
	var isDefault bool
	checkQuery := "SELECT created_by, is_default FROM status_flow_templates WHERE id = $1 AND is_active = true"
	err := s.db.QueryRow(checkQuery, templateID).Scan(&createdBy, &isDefault)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("template not found")
		}
		return fmt.Errorf("failed to check template permissions: %w", err)
	}

	if isDefault {
		return fmt.Errorf("cannot delete default template")
	}

	if !createdBy.Valid || uint(createdBy.Int64) != userID {
		return fmt.Errorf("permission denied: can only delete your own templates")
	}

	// 软删除
	deleteQuery := "UPDATE status_flow_templates SET is_active = false, updated_at = $1 WHERE id = $2 AND created_by = $3"
	result, err := s.db.Exec(deleteQuery, time.Now(), templateID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("template not found or permission denied")
	}

	return nil
}

// GetUserStatusPreferences 获取用户状态偏好设置
func (s *StatusConfigService) GetUserStatusPreferences(userID uint) (*model.UserStatusPreferences, error) {
	query := `
		SELECT id, user_id, preference_config, created_at, updated_at
		FROM user_status_preferences 
		WHERE user_id = $1
	`

	var preferences model.UserStatusPreferences
	var preferenceConfigBytes []byte

	err := s.db.QueryRow(query, userID).Scan(
		&preferences.ID,
		&preferences.UserID,
		&preferenceConfigBytes,
		&preferences.CreatedAt,
		&preferences.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// 返回默认配置
			return &model.UserStatusPreferences{
				UserID: userID,
				PreferenceConfig: s.getDefaultPreferenceConfig(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	// 解析preference_config
	if len(preferenceConfigBytes) > 0 {
		var preferenceConfig map[string]interface{}
		if err := json.Unmarshal(preferenceConfigBytes, &preferenceConfig); err == nil {
			preferences.PreferenceConfig = preferenceConfig
		} else {
			preferences.PreferenceConfig = s.getDefaultPreferenceConfig()
		}
	} else {
		preferences.PreferenceConfig = s.getDefaultPreferenceConfig()
	}

	return &preferences, nil
}

// UpdateUserStatusPreferences 更新用户状态偏好设置
func (s *StatusConfigService) UpdateUserStatusPreferences(userID uint, preferenceConfig map[string]interface{}) (*model.UserStatusPreferences, error) {
	// 验证配置格式
	if err := s.validatePreferenceConfig(preferenceConfig); err != nil {
		return nil, fmt.Errorf("invalid preference config: %w", err)
	}

	preferenceConfigBytes, err := json.Marshal(preferenceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal preference config: %w", err)
	}

	// 使用UPSERT语法
	upsertQuery := `
		INSERT INTO user_status_preferences (user_id, preference_config, created_at, updated_at)
		VALUES ($1, $2, $3, $3)
		ON CONFLICT (user_id) 
		DO UPDATE SET preference_config = EXCLUDED.preference_config, updated_at = EXCLUDED.updated_at
		RETURNING id, user_id, preference_config, created_at, updated_at
	`

	var preferences model.UserStatusPreferences
	var preferenceConfigBytesResult []byte

	err = s.db.QueryRow(upsertQuery, userID, preferenceConfigBytes, time.Now()).Scan(
		&preferences.ID,
		&preferences.UserID,
		&preferenceConfigBytesResult,
		&preferences.CreatedAt,
		&preferences.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update user preferences: %w", err)
	}

	// 解析返回的preference_config
	if len(preferenceConfigBytesResult) > 0 {
		json.Unmarshal(preferenceConfigBytesResult, &preferences.PreferenceConfig)
	}

	return &preferences, nil
}

// GetAvailableStatusTransitions 获取指定状态的可用转换选项
func (s *StatusConfigService) GetAvailableStatusTransitions(userID uint, currentStatus model.ApplicationStatus) ([]model.ApplicationStatus, error) {
	// 获取用户使用的模板配置
	var flowConfig string
	templateQuery := `
		SELECT COALESCE(sft.flow_config::text, '{"transitions": {}}')
		FROM status_flow_templates sft
		WHERE sft.is_default = true AND sft.is_active = true
		LIMIT 1
	`
	err := s.db.QueryRow(templateQuery).Scan(&flowConfig)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get flow template: %w", err)
	}

	var transitions []model.ApplicationStatus

	if err == sql.ErrNoRows {
		// 没有配置模板，返回所有有效状态
		allStatuses := []model.ApplicationStatus{
			model.StatusApplied,
			model.StatusResumeScreening,
			model.StatusResumeScreeningFail,
			model.StatusWrittenTest,
			model.StatusWrittenTestPass,
			model.StatusWrittenTestFail,
			model.StatusFirstInterview,
			model.StatusFirstPass,
			model.StatusFirstFail,
			model.StatusSecondInterview,
			model.StatusSecondPass,
			model.StatusSecondFail,
			model.StatusThirdInterview,
			model.StatusThirdPass,
			model.StatusThirdFail,
			model.StatusHRInterview,
			model.StatusHRPass,
			model.StatusHRFail,
			model.StatusOfferWaiting,
			model.StatusRejected,
			model.StatusOfferReceived,
			model.StatusOfferAccepted,
			model.StatusProcessFinished,
		}

		// 排除当前状态
		for _, status := range allStatuses {
			if status != currentStatus {
				transitions = append(transitions, status)
			}
		}
		return transitions, nil
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(flowConfig), &config); err != nil {
		return transitions, nil // 配置解析失败，返回空列表
	}

	transitionsMap, ok := config["transitions"].(map[string]interface{})
	if !ok {
		return transitions, nil
	}

	allowedStates, ok := transitionsMap[string(currentStatus)].([]interface{})
	if !ok {
		return transitions, nil
	}

	// 转换为ApplicationStatus类型
	for _, allowed := range allowedStates {
		if allowedStr, ok := allowed.(string); ok {
			status := model.ApplicationStatus(allowedStr)
			if status.IsValid() {
				transitions = append(transitions, status)
			}
		}
	}

	return transitions, nil
}

// validateFlowConfig 验证流转配置格式
func (s *StatusConfigService) validateFlowConfig(flowConfig map[string]interface{}) error {
	// 检查必需的字段
	transitions, exists := flowConfig["transitions"]
	if !exists {
		return fmt.Errorf("missing 'transitions' field")
	}

	transitionsMap, ok := transitions.(map[string]interface{})
	if !ok {
		return fmt.Errorf("'transitions' must be an object")
	}

	// 验证每个状态转换
	for fromStatus, toStates := range transitionsMap {
		// 验证源状态
		if !model.ApplicationStatus(fromStatus).IsValid() {
			return fmt.Errorf("invalid source status: %s", fromStatus)
		}

		// 验证目标状态列表
		toStatesList, ok := toStates.([]interface{})
		if !ok {
			return fmt.Errorf("transitions for status '%s' must be an array", fromStatus)
		}

		for _, toState := range toStatesList {
			toStateStr, ok := toState.(string)
			if !ok {
				return fmt.Errorf("transition target must be string")
			}
			if !model.ApplicationStatus(toStateStr).IsValid() {
				return fmt.Errorf("invalid target status: %s", toStateStr)
			}
		}
	}

	return nil
}

// validatePreferenceConfig 验证偏好配置格式
func (s *StatusConfigService) validatePreferenceConfig(preferenceConfig map[string]interface{}) error {
	// 验证通知设置
	if notifications, exists := preferenceConfig["notifications"]; exists {
		notificationsMap, ok := notifications.(map[string]interface{})
		if !ok {
			return fmt.Errorf("'notifications' must be an object")
		}

		// 验证通知类型
		validNotificationTypes := map[string]bool{
			"status_change":    true,
			"reminder_alerts":  true,
			"weekly_summary":   true,
		}

		for key, value := range notificationsMap {
			if !validNotificationTypes[key] {
				return fmt.Errorf("invalid notification type: %s", key)
			}
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("notification value for '%s' must be boolean", key)
			}
		}
	}

	// 验证显示设置
	if display, exists := preferenceConfig["display"]; exists {
		displayMap, ok := display.(map[string]interface{})
		if !ok {
			return fmt.Errorf("'display' must be an object")
		}

		// 验证时间线视图设置
		if timelineView, exists := displayMap["timeline_view"]; exists {
			if timelineViewStr, ok := timelineView.(string); ok {
				validViews := map[string]bool{
					"chronological": true,
					"stage":         true,
					"compact":       true,
				}
				if !validViews[timelineViewStr] {
					return fmt.Errorf("invalid timeline_view: %s", timelineViewStr)
				}
			}
		}

		// 验证状态颜色设置
		if statusColors, exists := displayMap["status_colors"]; exists {
			if colorsMap, ok := statusColors.(map[string]interface{}); ok {
				for status, color := range colorsMap {
					if !model.ApplicationStatus(status).IsValid() {
						return fmt.Errorf("invalid status in color config: %s", status)
					}
					if _, ok := color.(string); !ok {
						return fmt.Errorf("color value must be string for status: %s", status)
					}
				}
			}
		}
	}

	return nil
}

// getDefaultPreferenceConfig 获取默认偏好配置
func (s *StatusConfigService) getDefaultPreferenceConfig() map[string]interface{} {
	return map[string]interface{}{
		"notifications": map[string]bool{
			"status_change":   true,
			"reminder_alerts": true,
			"weekly_summary":  false,
		},
		"display": map[string]interface{}{
			"timeline_view": "chronological",
			"status_colors": map[string]string{
				"已投递":        "#6366f1",
				"简历筛选中":      "#f59e0b",
				"简历筛选未通过":    "#ef4444",
				"笔试中":        "#8b5cf6",
				"笔试通过":       "#059669",
				"笔试未通过":      "#ef4444",
				"一面中":        "#3b82f6",
				"一面通过":       "#10b981",
				"一面未通过":      "#ef4444",
				"二面中":        "#3b82f6",
				"二面通过":       "#10b981",
				"二面未通过":      "#ef4444",
				"三面中":        "#3b82f6",
				"三面通过":       "#10b981",
				"三面未通过":      "#ef4444",
				"HR面中":       "#8b5cf6",
				"HR面通过":      "#10b981",
				"HR面未通过":     "#ef4444",
				"待发offer":    "#f59e0b",
				"已收到offer":  "#059669",
				"已接受offer":  "#10b981",
				"已拒绝":        "#ef4444",
				"流程结束":       "#6b7280",
			},
			"show_duration": true,
		},
		"automation": map[string]bool{
			"auto_reminders":    true,
			"smart_suggestions": true,
		},
	}
}