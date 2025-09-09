package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// ApplicationStatus 投递状态枚举
type ApplicationStatus string

const (
	// 基础状态
	StatusApplied          ApplicationStatus = "已投递"
	StatusResumeScreening  ApplicationStatus = "简历筛选中"
	
	// 笔试状态
	StatusWrittenTest      ApplicationStatus = "笔试中"
	StatusWrittenTestPass  ApplicationStatus = "笔试通过"
	StatusWrittenTestFail  ApplicationStatus = "笔试未通过"
	
	// 一面状态
	StatusFirstInterview   ApplicationStatus = "一面中"
	StatusFirstPass        ApplicationStatus = "一面通过"
	StatusFirstFail        ApplicationStatus = "一面未通过"
	
	// 二面状态
	StatusSecondInterview  ApplicationStatus = "二面中"
	StatusSecondPass       ApplicationStatus = "二面通过"
	StatusSecondFail       ApplicationStatus = "二面未通过"
	
	// 三面状态
	StatusThirdInterview   ApplicationStatus = "三面中"
	StatusThirdPass        ApplicationStatus = "三面通过"
	StatusThirdFail        ApplicationStatus = "三面未通过"
	
	// HR面状态
	StatusHRInterview      ApplicationStatus = "HR面中"
	StatusHRPass           ApplicationStatus = "HR面通过"
	StatusHRFail           ApplicationStatus = "HR面未通过"
	
	// 最终状态
	StatusOfferWaiting     ApplicationStatus = "待发offer"
	StatusRejected         ApplicationStatus = "已拒绝"
	StatusOfferReceived    ApplicationStatus = "已收到offer"
	StatusOfferAccepted    ApplicationStatus = "已接受offer"
	StatusProcessFinished  ApplicationStatus = "流程结束"
	
	// 新增的失败状态
	StatusResumeScreeningFail ApplicationStatus = "简历筛选未通过"
)

// Value 实现 driver.Valuer 接口，用于数据库写入
func (s ApplicationStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// Scan 实现 sql.Scanner 接口，用于数据库读取
func (s *ApplicationStatus) Scan(value interface{}) error {
	if value == nil {
		*s = StatusApplied
		return nil
	}
	switch v := value.(type) {
	case string:
		*s = ApplicationStatus(v)
		return nil
	case []byte:
		*s = ApplicationStatus(string(v))
		return nil
	}
	return fmt.Errorf("cannot scan %T into ApplicationStatus", value)
}

// IsValid 检查状态是否有效
func (s ApplicationStatus) IsValid() bool {
	validStatuses := []ApplicationStatus{
		// 基础状态
		StatusApplied,
		StatusResumeScreening,
		StatusResumeScreeningFail,
		
		// 笔试状态
		StatusWrittenTest,
		StatusWrittenTestPass,
		StatusWrittenTestFail,
		
		// 一面状态
		StatusFirstInterview,
		StatusFirstPass,
		StatusFirstFail,
		
		// 二面状态
		StatusSecondInterview,
		StatusSecondPass,
		StatusSecondFail,
		
		// 三面状态
		StatusThirdInterview,
		StatusThirdPass,
		StatusThirdFail,
		
		// HR面状态
		StatusHRInterview,
		StatusHRPass,
		StatusHRFail,
		
		// 最终状态
		StatusOfferWaiting,
		StatusRejected,
		StatusOfferReceived,
		StatusOfferAccepted,
		StatusProcessFinished,
	}
	
	for _, validStatus := range validStatuses {
		if s == validStatus {
			return true
		}
	}
	return false
}

// IsFailedStatus 检查是否为失败状态
func (s ApplicationStatus) IsFailedStatus() bool {
	failedStatuses := []ApplicationStatus{
		StatusResumeScreeningFail,
		StatusWrittenTestFail,
		StatusFirstFail,
		StatusSecondFail,
		StatusThirdFail,
		StatusHRFail,
		StatusRejected,
	}
	
	for _, failedStatus := range failedStatuses {
		if s == failedStatus {
			return true
		}
	}
	return false
}

// IsInProgressStatus 检查是否为进行中状态
func (s ApplicationStatus) IsInProgressStatus() bool {
	inProgressStatuses := []ApplicationStatus{
		StatusApplied,
		StatusResumeScreening,
		StatusWrittenTest,
		StatusFirstInterview,
		StatusSecondInterview,
		StatusThirdInterview,
		StatusHRInterview,
	}
	
	for _, inProgressStatus := range inProgressStatuses {
		if s == inProgressStatus {
			return true
		}
	}
	return false
}

// IsPassedStatus 检查是否为通过状态
func (s ApplicationStatus) IsPassedStatus() bool {
	passedStatuses := []ApplicationStatus{
		StatusWrittenTestPass,
		StatusFirstPass,
		StatusSecondPass,
		StatusThirdPass,
		StatusHRPass,
		StatusOfferWaiting,
		StatusOfferReceived,
		StatusOfferAccepted,
		StatusProcessFinished,
	}
	
	for _, passedStatus := range passedStatuses {
		if s == passedStatus {
			return true
		}
	}
	return false
}

// JobApplication 投递记录模型
type JobApplication struct {
	ID                   int               `json:"id" db:"id"`
	UserID               uint              `json:"user_id" db:"user_id"`
	CompanyName          string            `json:"company_name" db:"company_name"`
	PositionTitle        string            `json:"position_title" db:"position_title"`
	ApplicationDate      string            `json:"application_date" db:"application_date"`
	Status               ApplicationStatus `json:"status" db:"status"`
	JobDescription       *string           `json:"job_description" db:"job_description"`
	SalaryRange          *string           `json:"salary_range" db:"salary_range"`
	WorkLocation         *string           `json:"work_location" db:"work_location"`
	ContactInfo          *string           `json:"contact_info" db:"contact_info"`
	Notes                *string           `json:"notes" db:"notes"`
	InterviewTime        *time.Time        `json:"interview_time" db:"interview_time"`
	ReminderTime         *time.Time        `json:"reminder_time" db:"reminder_time"`
	ReminderEnabled      bool              `json:"reminder_enabled" db:"reminder_enabled"`
	FollowUpDate         *string           `json:"follow_up_date" db:"follow_up_date"`
	HRName               *string           `json:"hr_name" db:"hr_name"`
	HRPhone              *string           `json:"hr_phone" db:"hr_phone"`
	HREmail              *string           `json:"hr_email" db:"hr_email"`
	InterviewLocation    *string           `json:"interview_location" db:"interview_location"`
	InterviewType        *string           `json:"interview_type" db:"interview_type"`
	CreatedAt            time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at" db:"updated_at"`

	// 新增状态跟踪字段
	StatusHistory        *StatusHistory    `json:"status_history,omitempty" db:"status_history"`
	LastStatusChange     *time.Time        `json:"last_status_change,omitempty" db:"last_status_change"`
	StatusDurationStats  *DurationStats    `json:"status_duration_stats,omitempty" db:"status_duration_stats"`
	StatusVersion        *int              `json:"status_version,omitempty" db:"status_version"`
}

// CreateJobApplicationRequest 创建投递记录请求
type CreateJobApplicationRequest struct {
	CompanyName       string            `json:"company_name" binding:"required"`
	PositionTitle     string            `json:"position_title" binding:"required"`
	ApplicationDate   string            `json:"application_date"`
	Status            ApplicationStatus `json:"status"`
	JobDescription    *string           `json:"job_description"`
	SalaryRange       *string           `json:"salary_range"`
	WorkLocation      *string           `json:"work_location"`
	ContactInfo       *string           `json:"contact_info"`
	Notes             *string           `json:"notes"`
	InterviewTime     *time.Time        `json:"interview_time"`
	ReminderTime      *time.Time        `json:"reminder_time"`
	ReminderEnabled   *bool             `json:"reminder_enabled"`
	FollowUpDate      *string           `json:"follow_up_date"`
	HRName            *string           `json:"hr_name"`
	HRPhone           *string           `json:"hr_phone"`
	HREmail           *string           `json:"hr_email"`
	InterviewLocation *string           `json:"interview_location"`
	InterviewType     *string           `json:"interview_type"`
}

// UpdateJobApplicationRequest 更新投递记录请求
type UpdateJobApplicationRequest struct {
	CompanyName       *string            `json:"company_name"`
	PositionTitle     *string            `json:"position_title"`
	ApplicationDate   *string            `json:"application_date"`
	Status            *ApplicationStatus `json:"status"`
	JobDescription    *string            `json:"job_description"`
	SalaryRange       *string            `json:"salary_range"`
	WorkLocation      *string            `json:"work_location"`
	ContactInfo       *string            `json:"contact_info"`
	Notes             *string            `json:"notes"`
	InterviewTime     *time.Time         `json:"interview_time"`
	ReminderTime      *time.Time         `json:"reminder_time"`
	ReminderEnabled   *bool              `json:"reminder_enabled"`
	FollowUpDate      *string            `json:"follow_up_date"`
	HRName            *string            `json:"hr_name"`
	HRPhone           *string            `json:"hr_phone"`
	HREmail           *string            `json:"hr_email"`
	InterviewLocation *string            `json:"interview_location"`
	InterviewType     *string            `json:"interview_type"`
}

// APIResponse 通用API响应格式
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationRequest 分页请求参数
type PaginationRequest struct {
	Page     int    `json:"page" form:"page"`         // 页码，从1开始
	PageSize int    `json:"page_size" form:"page_size"` // 每页条数，默认20，最大100
	SortBy   string `json:"sort_by" form:"sort_by"`   // 排序字段，默认application_date
	SortDir  string `json:"sort_dir" form:"sort_dir"` // 排序方向，ASC或DESC，默认DESC
	Status   *ApplicationStatus `json:"status" form:"status"` // 状态筛选，可选
}

// PaginationResponse 分页响应结构
type PaginationResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`        // 总记录数
	Page       int         `json:"page"`         // 当前页码
	PageSize   int         `json:"page_size"`    // 每页条数
	TotalPages int         `json:"total_pages"`  // 总页数
	HasNext    bool        `json:"has_next"`     // 是否有下一页
	HasPrev    bool        `json:"has_prev"`     // 是否有上一页
}

// ValidateAndSetDefaults 验证并设置分页参数默认值
func (p *PaginationRequest) ValidateAndSetDefaults() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100 // 限制最大每页条数，避免性能问题
	}
	if p.SortBy == "" {
		p.SortBy = "application_date"
	}
	if p.SortDir == "" || (p.SortDir != "ASC" && p.SortDir != "DESC") {
		p.SortDir = "DESC"
	}
}

// GetOffset 计算偏移量
func (p *PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// Value 实现 JSONB 字段的数据库写入
func (sh StatusHistory) Value() (driver.Value, error) {
	return json.Marshal(sh)
}

// Scan 实现 JSONB 字段的数据库读取
func (sh *StatusHistory) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into StatusHistory", value)
	}
	return json.Unmarshal(bytes, sh)
}

// Value 实现 JSONB 字段的数据库写入
func (ds DurationStats) Value() (driver.Value, error) {
	return json.Marshal(ds)
}

// Scan 实现 JSONB 字段的数据库读取
func (ds *DurationStats) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into DurationStats", value)
	}
	return json.Unmarshal(bytes, ds)
}

// BatchStatusUpdate 批量状态更新结构
type BatchStatusUpdate struct {
	ID     int               `json:"id" binding:"required"`
	Status ApplicationStatus `json:"status" binding:"required"`
}

// BatchCreateRequest 批量创建请求结构
type BatchCreateRequest struct {
	Applications []CreateJobApplicationRequest `json:"applications" binding:"required,min=1,max=50"`
}

// StatusHistory 状态历史结构
type StatusHistory struct {
	History  []StatusHistoryEntry `json:"history"`
	Metadata StatusMetadata      `json:"metadata"`
}

// StatusHistoryEntry 单个状态历史条目
type StatusHistoryEntry struct {
	ID               int64             `json:"id,omitempty" db:"id"`
	JobApplicationID int               `json:"job_application_id,omitempty" db:"job_application_id"`
	UserID           uint              `json:"user_id,omitempty" db:"user_id"`
	OldStatus        *ApplicationStatus `json:"old_status" db:"old_status"`
	NewStatus        ApplicationStatus `json:"new_status" db:"new_status"`
	StatusChangedAt  time.Time         `json:"status_changed_at" db:"status_changed_at"`
	DurationMinutes  *int              `json:"duration_minutes" db:"duration_minutes"`
	Metadata         map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	Note             *string           `json:"note,omitempty"`
	Trigger          string            `json:"trigger,omitempty"`
	CreatedAt        time.Time         `json:"created_at" db:"created_at"`
}

// StatusMetadata 状态元数据
type StatusMetadata struct {
	TotalChanges       int       `json:"total_changes"`
	CurrentStatus      string    `json:"current_status"`
	LastChanged        time.Time `json:"last_changed"`
	TotalDurationMinutes int     `json:"total_duration_minutes"`
	CurrentStage       string    `json:"current_stage,omitempty"`
}

// DurationStats 持续时间统计
type DurationStats struct {
	StatusDurations map[string]StatusDuration `json:"status_durations"`
	Milestones      map[string]time.Time      `json:"milestones,omitempty"`
	Analytics       ProcessAnalytics          `json:"analytics,omitempty"`
}

// StatusDuration 单个状态持续时间
type StatusDuration struct {
	TotalMinutes int     `json:"total_minutes"`
	Percentage   float64 `json:"percentage"`
}

// ProcessAnalytics 流程分析数据
type ProcessAnalytics struct {
	AverageResponseTime  int     `json:"average_response_time"`
	TotalProcessTime     int     `json:"total_process_time"`
	SuccessProbability   float64 `json:"success_probability,omitempty"`
}

// StatusFlowTemplate 状态流转模板
type StatusFlowTemplate struct {
	ID          int                    `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description *string                `json:"description" db:"description"`
	FlowConfig  map[string]interface{} `json:"flow_config" db:"flow_config"`
	IsDefault   bool                   `json:"is_default" db:"is_default"`
	IsActive    bool                   `json:"is_active" db:"is_active"`
	CreatedBy   *uint                  `json:"created_by" db:"created_by"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// UserStatusPreferences 用户状态偏好设置
type UserStatusPreferences struct {
	ID               int                    `json:"id" db:"id"`
	UserID           uint                   `json:"user_id" db:"user_id"`
	PreferenceConfig map[string]interface{} `json:"preference_config" db:"preference_config"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// StatusUpdateRequest 状态更新请求
type StatusUpdateRequest struct {
	Status    ApplicationStatus      `json:"status" binding:"required"`
	Note      *string                `json:"note,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Version   *int                   `json:"version,omitempty"` // 乐观锁版本控制
}

// StatusHistoryResponse 状态历史响应
type StatusHistoryResponse struct {
	History     []StatusHistoryEntry `json:"history"`
	Total       int                  `json:"total"`
	CurrentPage int                  `json:"current_page"`
	PageSize    int                  `json:"page_size"`
}

// StatusAnalyticsResponse 状态分析响应
type StatusAnalyticsResponse struct {
	UserID              uint                         `json:"user_id"`
	TotalApplications   int                          `json:"total_applications"`
	StatusDistribution  map[string]int               `json:"status_distribution"`
	SuccessRate         float64                      `json:"success_rate"`
	AverageDurations    map[string]float64           `json:"average_durations"`
	Trends              []StatusTrend                `json:"trends,omitempty"`
	StageAnalysis       map[string]StageStatistics   `json:"stage_analysis,omitempty"`
}

// StatusTrend 状态趋势数据
type StatusTrend struct {
	Date        string `json:"date"`
	Status      string `json:"status"`
	Count       int    `json:"count"`
	SuccessRate float64 `json:"success_rate,omitempty"`
}

// StageStatistics 阶段统计数据
type StageStatistics struct {
	StageName          string  `json:"stage_name"`
	TotalCount         int     `json:"total_count"`
	SuccessCount       int     `json:"success_count"`
	SuccessRate        float64 `json:"success_rate"`
	AverageDurationDays float64 `json:"average_duration_days"`
}

// StatusTransitionRule 状态转换规则
type StatusTransitionRule struct {
	FromStatus    ApplicationStatus   `json:"from_status"`
	AllowedStates []ApplicationStatus `json:"allowed_states"`
	RequireNote   bool                `json:"require_note,omitempty"`
	AutoTransition *ApplicationStatus `json:"auto_transition,omitempty"`
	TimeLimit     *int                `json:"time_limit,omitempty"`
}

// FlowConfig 流转配置结构
type FlowConfig struct {
	Transitions map[string][]string            `json:"transitions"`
	Rules       map[string]map[string]interface{} `json:"rules"`
}

// PreferenceConfig 偏好配置结构
type PreferenceConfig struct {
	Notifications map[string]bool              `json:"notifications"`
	Display       map[string]interface{}       `json:"display"`
	Automation    map[string]bool              `json:"automation,omitempty"`
}