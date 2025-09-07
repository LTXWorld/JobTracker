package model

import (
	"database/sql/driver"
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
	ID                int               `json:"id" db:"id"`
	UserID            uint              `json:"user_id" db:"user_id"`
	CompanyName       string            `json:"company_name" db:"company_name"`
	PositionTitle     string            `json:"position_title" db:"position_title"`
	ApplicationDate   string            `json:"application_date" db:"application_date"`
	Status            ApplicationStatus `json:"status" db:"status"`
	JobDescription    *string           `json:"job_description" db:"job_description"`
	SalaryRange       *string           `json:"salary_range" db:"salary_range"`
	WorkLocation      *string           `json:"work_location" db:"work_location"`
	ContactInfo       *string           `json:"contact_info" db:"contact_info"`
	Notes             *string           `json:"notes" db:"notes"`
	InterviewTime     *time.Time        `json:"interview_time" db:"interview_time"`
	ReminderTime      *time.Time        `json:"reminder_time" db:"reminder_time"`
	ReminderEnabled   bool              `json:"reminder_enabled" db:"reminder_enabled"`
	FollowUpDate      *string           `json:"follow_up_date" db:"follow_up_date"`
	HRName            *string           `json:"hr_name" db:"hr_name"`
	HRPhone           *string           `json:"hr_phone" db:"hr_phone"`
	HREmail           *string           `json:"hr_email" db:"hr_email"`
	InterviewLocation *string           `json:"interview_location" db:"interview_location"`
	InterviewType     *string           `json:"interview_type" db:"interview_type"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at" db:"updated_at"`
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