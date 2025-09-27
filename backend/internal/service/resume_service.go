package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"jobView-backend/internal/model"
	"jobView-backend/internal/repository"
)

type ResumeService struct {
    repo repository.ResumeRepository
}

func NewResumeService(repo repository.ResumeRepository) *ResumeService { return &ResumeService{repo: repo} }

// EnsureUserResume 获取或创建用户默认简历
func (s *ResumeService) EnsureUserResume(ctx context.Context, userID uint) (*model.Resume, error) {
    resume, err := s.repo.GetLatestResume(ctx, userID)
    if err != nil {
        if err == sql.ErrNoRows {
            now := time.Now()
            resume, err = s.repo.CreateResume(ctx, userID, "默认简历", "private", now)
            if err != nil {
                return nil, fmt.Errorf("create resume failed: %w", err)
            }
            return resume, nil
        }
        return nil, fmt.Errorf("get resume failed: %w", err)
    }
    return resume, nil
}

// GetResumeAggregate 返回简历及所有分区
func (s *ResumeService) GetResumeAggregate(ctx context.Context, userID uint, id int) (*model.ResumeAggregate, error) {
    resume, err := s.repo.GetResumeByID(ctx, id, userID)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("resume not found")
        }
        return nil, fmt.Errorf("get resume: %w", err)
    }

    sectionsList, err := s.repo.ListSections(ctx, id, userID)
    if err != nil {
        return nil, fmt.Errorf("list sections: %w", err)
    }
    sections := make(map[string]json.RawMessage)
    for _, section := range sectionsList {
        sections[section.Type] = section.Content
    }

    return &model.ResumeAggregate{Resume: *resume, Sections: sections}, nil
}

// UpdateMetadata 更新简历元信息
func (s *ResumeService) UpdateMetadata(ctx context.Context, userID uint, id int, title, privacy *string) (*model.Resume, error) {
    updates := make(map[string]interface{})
    if title != nil {
        updates["title"] = *title
    }
    if privacy != nil {
        updates["privacy"] = *privacy
    }
    updates["updated_at"] = time.Now()
    if err := s.repo.UpdateResumeMetadata(ctx, id, userID, updates); err != nil {
        return nil, fmt.Errorf("update resume: %w", err)
    }
    return s.EnsureUserResume(ctx, userID)
}

// ListSections 返回分区列表
func (s *ResumeService) ListSections(ctx context.Context, userID uint, id int) ([]model.ResumeSection, error) {
    return s.repo.ListSections(ctx, id, userID)
}

// UpsertSection upsert 指定分区
func (s *ResumeService) UpsertSection(ctx context.Context, userID uint, id int, typ string, content json.RawMessage) (*model.ResumeSection, error) {
    owned, err := s.repo.CheckResumeOwnership(ctx, id, userID)
    if err != nil || !owned { return nil, fmt.Errorf("resume not found") }
    // 验证分区类型合法
    if !model.IsValidSectionType(typ) {
        return nil, fmt.Errorf("invalid section type: %s", typ)
    }
    sct, err := s.repo.UpsertSection(ctx, id, typ, content, time.Now())
    if err != nil { return nil, err }
    // 更新简历完成度（改进版）：任意分区更新后重算
    _ = s.recalcCompleteness(ctx, id)
    return sct, nil
}

// UploadAttachment 保存附件
func (s *ResumeService) UploadAttachment(ctx context.Context, userID uint, id int, file multipartFile, headerFileName, mime string) (*model.ResumeAttachment, string, error) {
    owned, err := s.repo.CheckResumeOwnership(ctx, id, userID)
    if err != nil || !owned { return nil, "", fmt.Errorf("resume not found") }
    // 路径与保存
    base := "./uploads"
    relDir := filepath.Join("resumes", fmt.Sprintf("%d", userID), fmt.Sprintf("%d", id))
    if err := os.MkdirAll(filepath.Join(base, relDir), 0o755); err != nil { return nil, "", err }
    ext := strings.ToLower(filepath.Ext(headerFileName)); if ext=="" { ext = ".pdf" }
    name := fmt.Sprintf("cv_%d%s", time.Now().Unix(), ext)
    abs := filepath.Join(base, relDir, name)
    out, err := os.Create(abs); if err!=nil { return nil, "", err }
    if _, err := io.Copy(out, file); err!=nil { out.Close(); os.Remove(abs); return nil, "", err }
    out.Close()
    rel := filepath.ToSlash(filepath.Join(relDir, name))
    // 记录数据库
    att, err := s.repo.InsertAttachment(ctx, id, headerFileName, rel, mime, time.Now())
    if err != nil { return nil, "", err }
    url := "/static/"+rel
    return att, url, nil
}

// ListAttachments 列出简历附件（按创建时间倒序）
func (s *ResumeService) ListAttachments(ctx context.Context, userID uint, id int) ([]model.ResumeAttachment, error) {
    owned, err := s.repo.CheckResumeOwnership(ctx, id, userID)
    if err != nil || !owned { return nil, fmt.Errorf("resume not found") }
    return s.repo.ListAttachments(ctx, id, userID)
}

// multipartFile 接口便于测试
type multipartFile interface{ io.Reader }
