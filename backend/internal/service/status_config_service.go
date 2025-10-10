// Location: /Users/lutao/GolandProjects/jobView/backend/internal/service/status_config_service.go
// This file implements status flow template and user preference management service.
// It handles creation, updating, and retrieval of status transition rules and user customizations.
// Used by configuration management handlers to provide business logic for status flow management.

package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"jobView-backend/internal/model"
	"jobView-backend/internal/repository"
)

type StatusConfigService struct {
	repo repository.StatusConfigRepository
}

func NewStatusConfigService(repo repository.StatusConfigRepository) *StatusConfigService {
	return &StatusConfigService{repo: repo}
}

func (s *StatusConfigService) ensureRepo() error {
	if s.repo == nil {
		return fmt.Errorf("status config repository not initialized")
	}
	return nil
}

// EnsureDirectTransitionsInDefaultTemplate 确保默认模板包含面试阶段的直通转移规则
// 若模板不存在则忽略（由外部迁移负责创建）；若存在则在不改变其他配置的前提下补充：
// 一面中 -> 二面中，二面中 -> (三面中/HR面中)，三面中 -> HR面中
// EnsureDirectTransitionsInDefaultTemplate
// 1) 幂等补齐面试阶段直通规则：笔试中→一面中→二面中→(三面中/HR面中)→HR面中
// 2) 幂等补齐基础默认规则：已投递→(简历筛选中/简历筛选未通过/已拒绝)，简历筛选中→(笔试中/简历筛选未通过)
func (s *StatusConfigService) EnsureDirectTransitionsInDefaultTemplate() error {
	if err := s.ensureRepo(); err != nil {
		return err
	}

	id, cfgText, err := s.repo.GetDefaultFlowTemplate()
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read default flow template: %w", err)
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(cfgText), &cfg); err != nil {
		cfg = map[string]interface{}{
			"transitions": map[string]interface{}{},
			"rules":       map[string]interface{}{},
		}
	}

	transitionsMap, ok := cfg["transitions"].(map[string]interface{})
	if !ok || transitionsMap == nil {
		transitionsMap = map[string]interface{}{}
		cfg["transitions"] = transitionsMap
	}

	direct := map[string][]string{
		string(model.StatusWrittenTest): {
			string(model.StatusFirstInterview),
		},
		string(model.StatusFirstInterview): {
			string(model.StatusSecondInterview),
		},
		string(model.StatusSecondInterview): {
			string(model.StatusThirdInterview),
			string(model.StatusHRInterview),
		},
		string(model.StatusThirdInterview): {
			string(model.StatusHRInterview),
		},
	}
	baseline := map[string][]string{
		string(model.StatusApplied):         {string(model.StatusResumeScreening), "简历筛选未通过", string(model.StatusRejected)},
		string(model.StatusResumeScreening): {string(model.StatusWrittenTest), "简历筛选未通过"},
	}

	changed := false
	for from, targets := range direct {
		arrAny, _ := transitionsMap[from].([]interface{})
		if arrAny == nil {
			arrAny = []interface{}{}
		}
		arr := arrAny
		for _, to := range targets {
			exists := false
			for _, v := range arr {
				if sv, ok := v.(string); ok && sv == to {
					exists = true
					break
				}
			}
			if !exists {
				arr = append(arr, to)
				changed = true
			}
		}
		if len(arr) > 0 {
			transitionsMap[from] = arr
		}
	}

	for from, list := range baseline {
		arr, _ := transitionsMap[from].([]interface{})
		for _, to := range list {
			exists := false
			for _, v := range arr {
				if sv, ok := v.(string); ok && sv == to {
					exists = true
					break
				}
			}
			if !exists {
				arr = append(arr, to)
				changed = true
			}
		}
		if len(arr) > 0 {
			transitionsMap[from] = arr
		}
	}

	if !changed {
		return nil
	}

	newBytes, _ := json.Marshal(cfg)
	if err := s.repo.UpdateFlowConfigByID(id, string(newBytes)); err != nil {
		return fmt.Errorf("failed to update default flow template: %w", err)
	}
	return nil
}

// GetStatusFlowTemplates 获取状态流转模板列表
func (s *StatusConfigService) GetStatusFlowTemplates(userID uint) ([]model.StatusFlowTemplate, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	return s.repo.GetFlowTemplates(userID)
}

// CreateStatusFlowTemplate 创建自定义状态流转模板
func (s *StatusConfigService) CreateStatusFlowTemplate(userID uint, name, description string, flowConfig map[string]interface{}) (*model.StatusFlowTemplate, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}

	exists, err := s.repo.CheckTemplateNameExists(name, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check template name uniqueness: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("template name '%s' already exists", name)
	}
	if err := s.validateFlowConfig(flowConfig); err != nil {
		return nil, fmt.Errorf("invalid flow config: %w", err)
	}

	bytes, err := json.Marshal(flowConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal flow config: %w", err)
	}

	var desc *string
	if description != "" {
		desc = &description
	}

	tpl, err := s.repo.CreateFlowTemplate(userID, name, desc, bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create flow template: %w", err)
	}
	tpl.FlowConfig = flowConfig
	return tpl, nil
}

// UpdateStatusFlowTemplate 更新状态流转模板
func (s *StatusConfigService) UpdateStatusFlowTemplate(userID uint, templateID int, name, description string, flowConfig map[string]interface{}) (*model.StatusFlowTemplate, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}

	createdBy, isDefault, err := s.repo.GetTemplatePermissions(templateID)
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

	exists, err := s.repo.CheckTemplateNameExists(name, &templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to check template name uniqueness: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("template name '%s' already exists", name)
	}
	if err := s.validateFlowConfig(flowConfig); err != nil {
		return nil, fmt.Errorf("invalid flow config: %w", err)
	}

	bytes, err := json.Marshal(flowConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal flow config: %w", err)
	}
	var desc *string
	if description != "" {
		desc = &description
	}

	tpl, err := s.repo.UpdateFlowTemplate(userID, templateID, name, desc, bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to update flow template: %w", err)
	}
	tpl.FlowConfig = flowConfig
	return tpl, nil
}

// DeleteStatusFlowTemplate 删除状态流转模板（软删除）
func (s *StatusConfigService) DeleteStatusFlowTemplate(userID uint, templateID int) error {
	if err := s.ensureRepo(); err != nil {
		return err
	}

	createdBy, isDefault, err := s.repo.GetTemplatePermissions(templateID)
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
	if err := s.repo.DeleteFlowTemplate(userID, templateID); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}
	return nil
}

// GetUserStatusPreferences 获取用户状态偏好设置
func (s *StatusConfigService) GetUserStatusPreferences(userID uint) (*model.UserStatusPreferences, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}

	pref, err := s.repo.GetPreferences(userID)
	if err == sql.ErrNoRows {
		now := time.Now()
		return &model.UserStatusPreferences{UserID: userID, PreferenceConfig: s.getDefaultPreferenceConfig(), CreatedAt: now, UpdatedAt: now}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}
	if pref.PreferenceConfig == nil {
		pref.PreferenceConfig = s.getDefaultPreferenceConfig()
	}
	return pref, nil
}

// UpdateUserStatusPreferences 更新用户状态偏好设置
func (s *StatusConfigService) UpdateUserStatusPreferences(userID uint, preferenceConfig map[string]interface{}) (*model.UserStatusPreferences, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}

	if err := s.validatePreferenceConfig(preferenceConfig); err != nil {
		return nil, fmt.Errorf("invalid preference config: %w", err)
	}
	bytes, err := json.Marshal(preferenceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal preference config: %w", err)
	}
	pref, err := s.repo.UpsertPreferences(userID, bytes, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to update user preferences: %w", err)
	}
	pref.PreferenceConfig = preferenceConfig
	return pref, nil
}

// GetAvailableStatusTransitions 获取指定状态的可用转换选项
func (s *StatusConfigService) GetAvailableStatusTransitions(userID uint, currentStatus model.ApplicationStatus) ([]model.ApplicationStatus, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}

	_, flowConfigText, err := s.repo.GetDefaultFlowTemplate()
	transitionsSet := make(map[model.ApplicationStatus]bool)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get flow template: %w", err)
	}

	if err == sql.ErrNoRows {
		all := []model.ApplicationStatus{
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
		for _, st := range all {
			if st != currentStatus {
				transitionsSet[st] = true
			}
		}
		s.addImplicitDirectTransitions(currentStatus, transitionsSet)
		return setToSlice(transitionsSet), nil
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(flowConfigText), &cfg); err != nil {
		s.addBaselineDefaultTransitions(currentStatus, transitionsSet)
		s.addImplicitDirectTransitions(currentStatus, transitionsSet)
		return setToSlice(transitionsSet), nil
	}

	if transitions, ok := cfg["transitions"].(map[string]interface{}); ok {
		if allowed, ok := transitions[string(currentStatus)].([]interface{}); ok {
			for _, item := range allowed {
				if as, ok := item.(string); ok {
					st := model.ApplicationStatus(as)
					if st.IsValid() && st != currentStatus {
						transitionsSet[st] = true
					}
				}
			}
		} else {
			s.addBaselineDefaultTransitions(currentStatus, transitionsSet)
		}
	} else {
		s.addBaselineDefaultTransitions(currentStatus, transitionsSet)
	}

	s.addImplicitDirectTransitions(currentStatus, transitionsSet)
	return setToSlice(transitionsSet), nil
}
func (s *StatusConfigService) addImplicitDirectTransitions(currentStatus model.ApplicationStatus, set map[model.ApplicationStatus]bool) {
	direct := map[model.ApplicationStatus]model.ApplicationStatus{
		model.StatusWrittenTest:     model.StatusFirstInterview,
		model.StatusFirstInterview:  model.StatusSecondInterview,
		model.StatusSecondInterview: model.StatusThirdInterview,
		model.StatusThirdInterview:  model.StatusHRInterview,
	}
	if next, ok := direct[currentStatus]; ok {
		set[next] = true
	}
}

// addBaselineDefaultTransitions 在未配置或配置缺失时追加基础可用的转移规则
// 目的：避免前端拖拽列表为空（例如：已投递→简历筛选中）
func (s *StatusConfigService) addBaselineDefaultTransitions(currentStatus model.ApplicationStatus, set map[model.ApplicationStatus]bool) {
	switch currentStatus {
	case model.StatusApplied:
		set[model.StatusResumeScreening] = true
		set[model.ApplicationStatus("简历筛选未通过")] = true
		set[model.StatusRejected] = true
	case model.StatusResumeScreening:
		set[model.StatusWrittenTest] = true
		set[model.ApplicationStatus("简历筛选未通过")] = true
	}
}

// setToSlice 将状态集合转换为去重后的切片（稳定顺序不强制）
func setToSlice(set map[model.ApplicationStatus]bool) []model.ApplicationStatus {
	res := make([]model.ApplicationStatus, 0, len(set))
	for k := range set {
		res = append(res, k)
	}
	return res
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
			"status_change":   true,
			"reminder_alerts": true,
			"weekly_summary":  true,
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
				"已投递":      "#6366f1",
				"简历筛选中":    "#f59e0b",
				"简历筛选未通过":  "#ef4444",
				"笔试中":      "#8b5cf6",
				"笔试通过":     "#059669",
				"笔试未通过":    "#ef4444",
				"一面中":      "#3b82f6",
				"一面通过":     "#10b981",
				"一面未通过":    "#ef4444",
				"二面中":      "#3b82f6",
				"二面通过":     "#10b981",
				"二面未通过":    "#ef4444",
				"三面中":      "#3b82f6",
				"三面通过":     "#10b981",
				"三面未通过":    "#ef4444",
				"HR面中":     "#8b5cf6",
				"HR面通过":    "#10b981",
				"HR面未通过":   "#ef4444",
				"待发offer":  "#f59e0b",
				"已收到offer": "#059669",
				"已接受offer": "#10b981",
				"已拒绝":      "#ef4444",
				"流程结束":     "#6b7280",
			},
			"show_duration": true,
		},
		"automation": map[string]bool{
			"auto_reminders":    true,
			"smart_suggestions": true,
		},
	}
}
