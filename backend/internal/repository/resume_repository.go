package repository

import (
    "database/sql"
    "context"
    "encoding/json"
    "fmt"
    "time"

    "jobView-backend/internal/database"
    "jobView-backend/internal/model"
)

// ResumeRepository 定义简历持久化接口（骨架版）
type ResumeRepository interface {
    CheckResumeOwnership(ctx context.Context, resumeID int, userID uint) (bool, error)
    GetSectionsMap(ctx context.Context, resumeID int) (map[string]json.RawMessage, error)
    UpsertSection(ctx context.Context, resumeID int, typ string, content json.RawMessage, now time.Time) (*model.ResumeSection, error)
    UpdateResumeCompleteness(ctx context.Context, resumeID int, completeness int, isCompleted bool, now time.Time) error
    InsertAttachment(ctx context.Context, resumeID int, headerFileName, relPath, mime string, createdAt time.Time) (*model.ResumeAttachment, error)
    ListAttachments(ctx context.Context, resumeID int, userID uint) ([]model.ResumeAttachment, error)
}

type resumeRepo struct { db *database.DB }

func NewResumeRepository(db *database.DB) ResumeRepository { return &resumeRepo{db: db} }

func (r *resumeRepo) CheckResumeOwnership(ctx context.Context, resumeID int, userID uint) (bool, error) {
    var exists int
    if err := r.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM resumes WHERE id=$1 AND user_id=$2", resumeID, userID).Scan(&exists); err != nil {
        return false, err
    }
    return exists > 0, nil
}

func (r *resumeRepo) GetSectionsMap(ctx context.Context, resumeID int) (map[string]json.RawMessage, error) {
    rows, err := r.db.QueryContext(ctx, "SELECT type, content FROM resume_sections WHERE resume_id=$1", resumeID)
    if err != nil { return nil, err }
    defer rows.Close()
    sections := make(map[string]json.RawMessage)
    for rows.Next() {
        var t string; var c json.RawMessage
        if err := rows.Scan(&t, &c); err != nil { return nil, err }
        sections[t] = c
    }
    return sections, nil
}

func (r *resumeRepo) UpsertSection(ctx context.Context, resumeID int, typ string, content json.RawMessage, now time.Time) (*model.ResumeSection, error) {
    var sectionID int
    err := r.db.QueryRowContext(ctx, "SELECT id FROM resume_sections WHERE resume_id=$1 AND type=$2", resumeID, typ).Scan(&sectionID)
    if err == nil {
        if _, err := r.db.ExecContext(ctx, "UPDATE resume_sections SET content=$1, updated_at=$2 WHERE id=$3", content, now, sectionID); err != nil {
            return nil, fmt.Errorf("update section: %w", err)
        }
    } else if err == sql.ErrNoRows {
        if err := r.db.QueryRowContext(ctx, "INSERT INTO resume_sections (resume_id,type,content,created_at,updated_at) VALUES ($1,$2,$3,$4,$4) RETURNING id", resumeID, typ, content, now).Scan(&sectionID); err != nil {
            return nil, fmt.Errorf("insert section: %w", err)
        }
    } else {
        return nil, fmt.Errorf("query section: %w", err)
    }
    var sct model.ResumeSection
    if err := r.db.QueryRowContext(ctx, "SELECT id,resume_id,type,sort_order,content,created_at,updated_at FROM resume_sections WHERE id=$1", sectionID).Scan(&sct.ID,&sct.ResumeID,&sct.Type,&sct.SortOrder,&sct.Content,&sct.CreatedAt,&sct.UpdatedAt); err != nil {
        return nil, err
    }
    return &sct, nil
}

func (r *resumeRepo) UpdateResumeCompleteness(ctx context.Context, resumeID int, completeness int, isCompleted bool, now time.Time) error {
    _, err := r.db.ExecContext(ctx, "UPDATE resumes SET completeness=$1, is_completed=$2, updated_at=$3 WHERE id=$4", completeness, isCompleted, now, resumeID)
    return err
}

func (r *resumeRepo) InsertAttachment(ctx context.Context, resumeID int, headerFileName, relPath, mime string, createdAt time.Time) (*model.ResumeAttachment, error) {
    var att model.ResumeAttachment
    if err := r.db.QueryRowContext(ctx, "INSERT INTO resume_attachments (resume_id,file_name,file_path,mime_type,created_at) VALUES ($1,$2,$3,$4,$5) RETURNING id,resume_id,file_name,file_path,mime_type,created_at", resumeID, headerFileName, relPath, mime, createdAt).Scan(&att.ID,&att.ResumeID,&att.FileName,&att.FilePath,&att.MimeType,&att.CreatedAt); err != nil {
        return nil, err
    }
    return &att, nil
}

func (r *resumeRepo) ListAttachments(ctx context.Context, resumeID int, userID uint) ([]model.ResumeAttachment, error) {
    // ownership assumed validated by caller, but keep a join for safety
    rows, err := r.db.QueryContext(ctx, "SELECT a.id,a.resume_id,a.file_name,a.file_path,a.mime_type,a.file_size,a.etag,a.created_at FROM resume_attachments a JOIN resumes r ON a.resume_id=r.id WHERE a.resume_id=$1 AND r.user_id=$2 ORDER BY a.created_at DESC, a.id DESC", resumeID, userID)
    if err != nil { return nil, err }
    defer rows.Close()
    var list []model.ResumeAttachment
    for rows.Next() {
        var a model.ResumeAttachment
        if err := rows.Scan(&a.ID,&a.ResumeID,&a.FileName,&a.FilePath,&a.MimeType,&a.FileSize,&a.ETag,&a.CreatedAt); err != nil { return nil, err }
        list = append(list, a)
    }
    return list, nil
}
