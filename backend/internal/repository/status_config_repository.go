package repository

import (
    "database/sql"
    "fmt"
    "time"

    "jobView-backend/internal/database"
    "jobView-backend/internal/model"
    "encoding/json"
    "context"
)

// StatusConfigRepository 封装状态模板与用户偏好的数据访问（基于 GORM Raw/Exec）
type StatusConfigRepository interface {
    // 模板
    GetFlowTemplates(userID uint) ([]model.StatusFlowTemplate, error)
    CheckTemplateNameExists(name string, excludeID *int) (bool, error)
    GetTemplatePermissions(templateID int) (createdBy sql.NullInt64, isDefault bool, err error)
    CreateFlowTemplate(userID uint, name string, desc *string, flowConfigBytes []byte) (*model.StatusFlowTemplate, error)
    UpdateFlowTemplate(userID uint, templateID int, name string, desc *string, flowConfigBytes []byte) (*model.StatusFlowTemplate, error)
    DeleteFlowTemplate(userID uint, templateID int) error

    // 默认模板读写（用于直通规则补齐与可用转换）
    GetDefaultFlowTemplate() (id int, flowConfigText string, err error)
    UpdateFlowConfigByID(id int, flowConfigText string) error

    // 用户偏好
    GetPreferences(userID uint) (*model.UserStatusPreferences, error)
    UpsertPreferences(userID uint, preferenceBytes []byte, now time.Time) (*model.UserStatusPreferences, error)
}

type statusConfigRepo struct { db *database.DB }

func NewStatusConfigRepository(db *database.DB) StatusConfigRepository { return &statusConfigRepo{db: db} }

func (r *statusConfigRepo) GetFlowTemplates(userID uint) ([]model.StatusFlowTemplate, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    q := `SELECT id, name, description, flow_config, is_default, is_active, created_by, created_at, updated_at
          FROM status_flow_templates WHERE is_active=true AND (created_by IS NULL OR created_by=$1)
          ORDER BY is_default DESC, name ASC`
    rows, err := r.db.ORM.WithContext(ctx).Raw(q, userID).Rows()
    if err != nil { return nil, fmt.Errorf("failed to get flow templates: %w", err) }
    defer rows.Close()
    var list []model.StatusFlowTemplate
    for rows.Next() {
        var t model.StatusFlowTemplate
        var desc sql.NullString
        var cfg []byte
        var createdBy sql.NullInt64
        if err := rows.Scan(&t.ID, &t.Name, &desc, &cfg, &t.IsDefault, &t.IsActive, &createdBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
            return nil, err
        }
        if desc.Valid { t.Description=&desc.String }
        if createdBy.Valid { v:=uint(createdBy.Int64); t.CreatedBy=&v }
        if len(cfg)>0 { var m map[string]interface{}; _ = jsonUnmarshal(cfg, &m); t.FlowConfig=m }
        list = append(list, t)
    }
    return list, nil
}

func (r *statusConfigRepo) CheckTemplateNameExists(name string, excludeID *int) (bool, error) {
    var exists bool
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    if excludeID != nil {
        if err := r.db.ORM.WithContext(ctx).Raw("SELECT EXISTS(SELECT 1 FROM status_flow_templates WHERE name=$1 AND id<>$2 AND is_active=true)", name, *excludeID).Row().Scan(&exists); err != nil { return false, err }
    } else {
        if err := r.db.ORM.WithContext(ctx).Raw("SELECT EXISTS(SELECT 1 FROM status_flow_templates WHERE name=$1 AND is_active=true)", name).Row().Scan(&exists); err != nil { return false, err }
    }
    return exists, nil
}

func (r *statusConfigRepo) GetTemplatePermissions(templateID int) (sql.NullInt64, bool, error) {
    var createdBy sql.NullInt64
    var isDefault bool
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    if err := r.db.ORM.WithContext(ctx).Raw("SELECT created_by, is_default FROM status_flow_templates WHERE id=$1 AND is_active=true", templateID).Row().Scan(&createdBy, &isDefault); err != nil {
        return sql.NullInt64{}, false, err
    }
    return createdBy, isDefault, nil
}

func (r *statusConfigRepo) CreateFlowTemplate(userID uint, name string, desc *string, flowConfigBytes []byte) (*model.StatusFlowTemplate, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    q := `INSERT INTO status_flow_templates (name, description, flow_config, created_by, is_active)
          VALUES ($1,$2,$3,$4,true) RETURNING id, created_at, updated_at`
    var t model.StatusFlowTemplate
    if err := r.db.ORM.WithContext(ctx).Raw(q, name, desc, flowConfigBytes, userID).Row().Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt); err != nil { return nil, err }
    t.Name=name; t.Description=desc; t.FlowConfig=nil; t.IsDefault=false; t.IsActive=true; t.CreatedBy=&userID
    return &t, nil
}

func (r *statusConfigRepo) UpdateFlowTemplate(userID uint, templateID int, name string, desc *string, flowConfigBytes []byte) (*model.StatusFlowTemplate, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    q := `UPDATE status_flow_templates SET name=$1, description=$2, flow_config=$3, updated_at=$4
          WHERE id=$5 AND created_by=$6
          RETURNING id, name, description, flow_config, is_default, is_active, created_by, created_at, updated_at`
    var t model.StatusFlowTemplate
    var d sql.NullString
    var cfg []byte
    var createdBy sql.NullInt64
    row := r.db.ORM.WithContext(ctx).Raw(q, name, desc, flowConfigBytes, time.Now(), templateID, userID).Row()
    if err := row.Scan(&t.ID,&t.Name,&d,&cfg,&t.IsDefault,&t.IsActive,&createdBy,&t.CreatedAt,&t.UpdatedAt); err != nil { return nil, err }
    if d.Valid { t.Description=&d.String }
    if createdBy.Valid { v:=uint(createdBy.Int64); t.CreatedBy=&v }
    if len(cfg)>0 { var m map[string]interface{}; _ = jsonUnmarshal(cfg, &m); t.FlowConfig=m }
    return &t, nil
}

func (r *statusConfigRepo) DeleteFlowTemplate(userID uint, templateID int) error {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    res := r.db.ORM.WithContext(ctx).Exec("UPDATE status_flow_templates SET is_active=false, updated_at=$1 WHERE id=$2 AND created_by=$3", time.Now(), templateID, userID)
    return res.Error
}

func (r *statusConfigRepo) GetDefaultFlowTemplate() (int, string, error) {
    var id int
    var text string
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    q := `SELECT id, COALESCE(flow_config::text, '{"transitions": {}, "rules": {}}') FROM status_flow_templates WHERE is_default=true AND is_active=true LIMIT 1`
    if err := r.db.ORM.WithContext(ctx).Raw(q).Row().Scan(&id, &text); err != nil { return 0, "", err }
    return id, text, nil
}

func (r *statusConfigRepo) UpdateFlowConfigByID(id int, flowConfigText string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    return r.db.ORM.WithContext(ctx).Exec("UPDATE status_flow_templates SET flow_config=$1, updated_at=$2 WHERE id=$3", flowConfigText, time.Now(), id).Error
}

func (r *statusConfigRepo) GetPreferences(userID uint) (*model.UserStatusPreferences, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    q := `SELECT id, user_id, preference_config, created_at, updated_at FROM user_status_preferences WHERE user_id=$1`
    var pref model.UserStatusPreferences
    var cfg []byte
    row := r.db.ORM.WithContext(ctx).Raw(q, userID).Row()
    if err := row.Scan(&pref.ID,&pref.UserID,&cfg,&pref.CreatedAt,&pref.UpdatedAt); err != nil { return nil, err }
    if len(cfg)>0 { var m map[string]interface{}; _ = jsonUnmarshal(cfg, &m); pref.PreferenceConfig=m }
    return &pref, nil
}

func (r *statusConfigRepo) UpsertPreferences(userID uint, preferenceBytes []byte, now time.Time) (*model.UserStatusPreferences, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    q := `INSERT INTO user_status_preferences (user_id, preference_config, created_at, updated_at)
          VALUES ($1,$2,$3,$3)
          ON CONFLICT (user_id)
          DO UPDATE SET preference_config = EXCLUDED.preference_config, updated_at = EXCLUDED.updated_at
          RETURNING id, user_id, preference_config, created_at, updated_at`
    var pref model.UserStatusPreferences
    var out []byte
    row := r.db.ORM.WithContext(ctx).Raw(q, userID, preferenceBytes, now).Row()
    if err := row.Scan(&pref.ID,&pref.UserID,&out,&pref.CreatedAt,&pref.UpdatedAt); err != nil { return nil, err }
    if len(out)>0 { var m map[string]interface{}; _ = jsonUnmarshal(out, &m); pref.PreferenceConfig=m }
    return &pref, nil
}

// 轻量 JSON 解析避免循环依赖
func jsonUnmarshal(b []byte, v interface{}) error { return json.Unmarshal(b, v) }
