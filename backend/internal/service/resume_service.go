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

    "jobView-backend/internal/database"
    "jobView-backend/internal/model"
    "jobView-backend/internal/repository"
)

type ResumeService struct {
    db   *database.DB
    repo repository.ResumeRepository
}

func NewResumeService(db *database.DB) *ResumeService { return &ResumeService{db: db, repo: repository.NewResumeRepository(db)} }

// EnsureUserResume 获取或创建用户默认简历
func (s *ResumeService) EnsureUserResume(ctx context.Context, userID uint) (*model.Resume, error) {
    var r model.Resume
    // 优先返回最近更新的简历，避免误拿到历史最早的一份
    err := s.db.QueryRowContext(ctx, "SELECT id,user_id,title,summary,privacy,current_version,is_completed,completeness,created_at,updated_at FROM resumes WHERE user_id=$1 ORDER BY updated_at DESC LIMIT 1", userID).Scan(
        &r.ID,&r.UserID,&r.Title,&r.Summary,&r.Privacy,&r.CurrentVersion,&r.IsCompleted,&r.Completeness,&r.CreatedAt,&r.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        // create
        now := time.Now()
        err = s.db.QueryRowContext(ctx, "INSERT INTO resumes (user_id,title,privacy,created_at,updated_at) VALUES ($1,'默认简历','private',$2,$2) RETURNING id,created_at,updated_at", userID, now).Scan(&r.ID,&r.CreatedAt,&r.UpdatedAt)
        if err != nil { return nil, fmt.Errorf("create resume failed: %w", err) }
        r.UserID = userID; r.Title = "默认简历"; r.Privacy = "private"; r.CurrentVersion=1
        return &r,nil
    }
    if err != nil { return nil, fmt.Errorf("get resume failed: %w", err) }
    return &r,nil
}

// GetResumeAggregate 返回简历及所有分区
func (s *ResumeService) GetResumeAggregate(ctx context.Context, userID uint, id int) (*model.ResumeAggregate, error) {
    var r model.Resume
    err := s.db.QueryRowContext(ctx, "SELECT id,user_id,title,summary,privacy,current_version,is_completed,completeness,created_at,updated_at FROM resumes WHERE id=$1 AND user_id=$2", id, userID).Scan(
        &r.ID,&r.UserID,&r.Title,&r.Summary,&r.Privacy,&r.CurrentVersion,&r.IsCompleted,&r.Completeness,&r.CreatedAt,&r.UpdatedAt,
    )
    if err != nil { if err==sql.ErrNoRows {return nil, fmt.Errorf("resume not found")}; return nil, fmt.Errorf("get resume: %w", err) }

    rows, err := s.db.QueryContext(ctx, "SELECT type,content FROM resume_sections WHERE resume_id=$1 ORDER BY sort_order,id", id)
    if err != nil { return nil, fmt.Errorf("list sections: %w", err) }
    defer rows.Close()
    sections := make(map[string]json.RawMessage)
    for rows.Next() {
        var t string; var c json.RawMessage
        if err := rows.Scan(&t,&c); err != nil { return nil, err }
        sections[t] = c
    }
    return &model.ResumeAggregate{Resume:r, Sections:sections}, nil
}

// UpdateMetadata 更新简历元信息
func (s *ResumeService) UpdateMetadata(ctx context.Context, userID uint, id int, title, privacy *string) (*model.Resume, error) {
    set := []string{}; args := []interface{}{}; i:=1
    if title!=nil { set = append(set, fmt.Sprintf("title=$%d", i)); args=append(args,*title); i++ }
    if privacy!=nil { set = append(set, fmt.Sprintf("privacy=$%d", i)); args=append(args,*privacy); i++ }
    set = append(set, fmt.Sprintf("updated_at=$%d", i)); args=append(args,time.Now()); i++
    args = append(args, id, userID)
    q := fmt.Sprintf("UPDATE resumes SET %s WHERE id=$%d AND user_id=$%d", strings.Join(set,","), i, i+1)
    if _, err := s.db.ExecContext(ctx, q, args...); err != nil { return nil, fmt.Errorf("update resume: %w", err) }
    return s.EnsureUserResume(ctx, userID) // 返回最新
}

// ListSections 返回分区列表
func (s *ResumeService) ListSections(ctx context.Context, userID uint, id int) ([]model.ResumeSection, error) {
    rows, err := s.db.QueryContext(ctx, "SELECT s.id,s.resume_id,s.type,s.sort_order,s.content,s.created_at,s.updated_at FROM resume_sections s JOIN resumes r ON s.resume_id=r.id WHERE r.id=$1 AND r.user_id=$2 ORDER BY sort_order,id", id, userID)
    if err != nil { return nil, fmt.Errorf("list sections: %w", err) }
    defer rows.Close()
    var list []model.ResumeSection
    for rows.Next() {
        var sct model.ResumeSection
        if err := rows.Scan(&sct.ID,&sct.ResumeID,&sct.Type,&sct.SortOrder,&sct.Content,&sct.CreatedAt,&sct.UpdatedAt); err!=nil {return nil, err}
        list = append(list, sct)
    }
    return list,nil
}

// UpsertSection upsert 指定分区
func (s *ResumeService) UpsertSection(ctx context.Context, userID uint, id int, typ string, content json.RawMessage) (*model.ResumeSection, error) {
    // 验证归属
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
    // 验证归属
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
