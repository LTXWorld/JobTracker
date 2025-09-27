package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"jobView-backend/internal/database"
	"jobView-backend/internal/model"

	"gorm.io/gorm"
)

type ResumeRepository interface {
	GetLatestResume(ctx context.Context, userID uint) (*model.Resume, error)
	CreateResume(ctx context.Context, userID uint, title string, privacy string, now time.Time) (*model.Resume, error)
	GetResumeByID(ctx context.Context, id int, userID uint) (*model.Resume, error)
	UpdateResumeMetadata(ctx context.Context, id int, userID uint, fields map[string]interface{}) error
	ListSections(ctx context.Context, resumeID int, userID uint) ([]model.ResumeSection, error)
	UpsertSection(ctx context.Context, resumeID int, sectionType string, content json.RawMessage, now time.Time) (*model.ResumeSection, error)
	CheckResumeOwnership(ctx context.Context, resumeID int, userID uint) (bool, error)
	InsertAttachment(ctx context.Context, resumeID int, filename, path, mime string, now time.Time) (*model.ResumeAttachment, error)
	ListAttachments(ctx context.Context, resumeID int, userID uint) ([]model.ResumeAttachment, error)
	GetSectionsMap(ctx context.Context, resumeID int) (map[string]json.RawMessage, error)
	UpdateResumeCompleteness(ctx context.Context, resumeID int, completeness int, completed bool, now time.Time) error
}

type resumeRepository struct {
	db  *database.DB
	orm *gorm.DB
}

func NewResumeRepository(db *database.DB) ResumeRepository {
	var orm *gorm.DB
	if db != nil {
		orm = db.ORM
	}
	return &resumeRepository{db: db, orm: orm}
}

func (r *resumeRepository) ormWithContext(ctx context.Context) (*gorm.DB, error) {
	if r.orm == nil {
		return nil, fmt.Errorf("gorm instance not initialized")
	}
	return r.orm.WithContext(ctx), nil
}

func (r *resumeRepository) GetLatestResume(ctx context.Context, userID uint) (*model.Resume, error) {
	orm, err := r.ormWithContext(ctx)
	if err != nil {
		return nil, err
	}
	var resume model.Resume
	if err := orm.Order("updated_at DESC").Where("user_id = ?", userID).First(&resume).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &resume, nil
}

func (r *resumeRepository) CreateResume(ctx context.Context, userID uint, title string, privacy string, now time.Time) (*model.Resume, error) {
	orm, err := r.ormWithContext(ctx)
	if err != nil {
		return nil, err
	}
	resume := &model.Resume{
		UserID:         userID,
		Title:          title,
		Privacy:        privacy,
		CurrentVersion: 1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := orm.Create(resume).Error; err != nil {
		return nil, err
	}
	return resume, nil
}

func (r *resumeRepository) GetResumeByID(ctx context.Context, id int, userID uint) (*model.Resume, error) {
	orm, err := r.ormWithContext(ctx)
	if err != nil {
		return nil, err
	}
	var resume model.Resume
	if err := orm.Where("id = ? AND user_id = ?", id, userID).First(&resume).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &resume, nil
}

func (r *resumeRepository) UpdateResumeMetadata(ctx context.Context, id int, userID uint, fields map[string]interface{}) error {
	orm, err := r.ormWithContext(ctx)
	if err != nil {
		return err
	}
	if err := orm.Model(&model.Resume{}).Where("id = ? AND user_id = ?", id, userID).Updates(fields).Error; err != nil {
		return err
	}
	return nil
}

func (r *resumeRepository) ListSections(ctx context.Context, resumeID int, userID uint) ([]model.ResumeSection, error) {
	orm, err := r.ormWithContext(ctx)
	if err != nil {
		return nil, err
	}
	var sections []model.ResumeSection
	if err := orm.
		Where("resume_id = ?", resumeID).
		Order("sort_order, id").
		Find(&sections).Error; err != nil {
		return nil, err
	}
	return sections, nil
}

func (r *resumeRepository) UpsertSection(ctx context.Context, resumeID int, sectionType string, content json.RawMessage, now time.Time) (*model.ResumeSection, error) {
	orm, err := r.ormWithContext(ctx)
	if err != nil {
		return nil, err
	}
	section := model.ResumeSection{}
	err = orm.Where("resume_id = ? AND type = ?", resumeID, sectionType).First(&section).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			section = model.ResumeSection{
				ResumeID:  resumeID,
				Type:      sectionType,
				SortOrder: 0,
				Content:   content,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := orm.Create(&section).Error; err != nil {
				return nil, err
			}
			return &section, nil
		}
		return nil, err
	}
	section.Content = content
	section.UpdatedAt = now
	if err := orm.Save(&section).Error; err != nil {
		return nil, err
	}
	return &section, nil
}

func (r *resumeRepository) CheckResumeOwnership(ctx context.Context, resumeID int, userID uint) (bool, error) {
	orm, err := r.ormWithContext(ctx)
	if err != nil {
		return false, err
	}
	var exists bool
	if err := orm.Raw("SELECT EXISTS(SELECT 1 FROM resumes WHERE id = ? AND user_id = ?)", resumeID, userID).Scan(&exists).Error; err != nil {
		return false, err
	}
	return exists, nil
}

func (r *resumeRepository) InsertAttachment(ctx context.Context, resumeID int, filename, path, mime string, now time.Time) (*model.ResumeAttachment, error) {
	orm, err := r.ormWithContext(ctx)
	if err != nil {
		return nil, err
	}
	att := &model.ResumeAttachment{
		ResumeID:  resumeID,
		FileName:  filename,
		FilePath:  path,
		CreatedAt: now,
	}
	if mime != "" {
		att.MimeType = &mime
	}
	if err := orm.Create(att).Error; err != nil {
		return nil, err
	}
	return att, nil
}

func (r *resumeRepository) ListAttachments(ctx context.Context, resumeID int, userID uint) ([]model.ResumeAttachment, error) {
	orm, err := r.ormWithContext(ctx)
	if err != nil {
		return nil, err
	}
	var list []model.ResumeAttachment
	if err := orm.Where("resume_id = ?", resumeID).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *resumeRepository) GetSectionsMap(ctx context.Context, resumeID int) (map[string]json.RawMessage, error) {
	orm, err := r.ormWithContext(ctx)
	if err != nil {
		return nil, err
	}
	var sections []model.ResumeSection
	if err := orm.Select("type", "content").Where("resume_id = ?", resumeID).Find(&sections).Error; err != nil {
		return nil, err
	}
	result := make(map[string]json.RawMessage, len(sections))
	for _, section := range sections {
		if section.Content == nil {
			continue
		}
		// 拷贝底层字节，避免后续被 GORM 复用缓冲区污染
		copied := make([]byte, len(section.Content))
		copy(copied, section.Content)
		result[section.Type] = json.RawMessage(copied)
	}
	return result, nil
}

func (r *resumeRepository) UpdateResumeCompleteness(ctx context.Context, resumeID int, completeness int, completed bool, now time.Time) error {
	orm, err := r.ormWithContext(ctx)
	if err != nil {
		return err
	}
	updates := map[string]interface{}{
		"completeness": completeness,
		"is_completed": completed,
		"updated_at":   now,
	}
	return orm.Model(&model.Resume{}).Where("id = ?", resumeID).Updates(updates).Error
}
