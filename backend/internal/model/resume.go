package model

import (
    "encoding/json"
    "time"
)

// Resume 主简历表
type Resume struct {
    ID             int       `json:"id" db:"id"`
    UserID         uint      `json:"user_id" db:"user_id"`
    Title          string    `json:"title" db:"title"`
    Summary        *string   `json:"summary,omitempty" db:"summary"`
    Privacy        string    `json:"privacy" db:"privacy"`
    CurrentVersion int       `json:"current_version" db:"current_version"`
    IsCompleted    bool      `json:"is_completed" db:"is_completed"`
    Completeness   int       `json:"completeness" db:"completeness"`
    CreatedAt      time.Time `json:"created_at" db:"created_at"`
    UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// ResumeSection 每个区块一条记录
type ResumeSection struct {
    ID        int             `json:"id" db:"id"`
    ResumeID  int             `json:"resume_id" db:"resume_id"`
    Type      string          `json:"type" db:"type"`
    SortOrder int             `json:"sort_order" db:"sort_order"`
    Content   json.RawMessage `json:"content" db:"content"`
    CreatedAt time.Time       `json:"created_at" db:"created_at"`
    UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

// ResumeAttachment 附件
type ResumeAttachment struct {
    ID        int       `json:"id" db:"id"`
    ResumeID  int       `json:"resume_id" db:"resume_id"`
    FileName  string    `json:"file_name" db:"file_name"`
    FilePath  string    `json:"file_path" db:"file_path"`
    MimeType  *string   `json:"mime_type,omitempty" db:"mime_type"`
    FileSize  *int64    `json:"file_size,omitempty" db:"file_size"`
    ETag      *string   `json:"etag,omitempty" db:"etag"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ResumeAggregate 返回聚合视图
type ResumeAggregate struct {
    Resume   Resume                     `json:"resume"`
    Sections map[string]json.RawMessage `json:"sections"`
}

// ResumeSummary 当前用户简历与完成度
type ResumeSummary struct {
    Resume        Resume   `json:"resume"`
    SectionTypes  []string `json:"sections"`
    MissingFields []string `json:"missing_required,omitempty"`
}

