package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"jobView-backend/internal/database"
	"jobView-backend/internal/model"

	"gorm.io/gorm"
)

type AuthRepository interface {
	UsernameExists(username string) (bool, error)
	UsernameExistsInsensitive(username string) (bool, error)
	EmailExists(email string) (bool, error)
	EmailExistsInsensitive(email string) (bool, error)
	CreateUser(username, email, hashedPassword string, now time.Time) (*model.User, error)
	GetUserByID(id uint) (*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
	UpdateUser(userID uint, fields map[string]interface{}) error
	UpdatePassword(userID uint, hashedPassword string, updatedAt time.Time) error
	GetAvatarVersion(userID uint) (int, error)
	UpdateAvatar(userID uint, relPath string, version int, updatedAt time.Time) error
	ListRecentApplications(userID uint, limit int) ([]map[string]interface{}, error)
	ListUpcomingInterviews(userID uint, limit int) ([]map[string]interface{}, error)
	ListDailyStats(userID uint, days int) ([]map[string]interface{}, error)
}

type authRepository struct {
	db  *database.DB
	orm *gorm.DB
}

func NewAuthRepository(db *database.DB) AuthRepository {
	var orm *gorm.DB
	if db != nil {
		orm = db.ORM
	}
	return &authRepository{db: db, orm: orm}
}

func (r *authRepository) ormWithContext() (*gorm.DB, error) {
	if r.orm == nil {
		return nil, fmt.Errorf("gorm instance not initialized")
	}
	return r.orm.WithContext(context.Background()), nil
}

func (r *authRepository) UsernameExists(username string) (bool, error) {
	orm, err := r.ormWithContext()
	if err != nil {
		return false, err
	}
	var exists bool
	if err := orm.Raw("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", username).Scan(&exists).Error; err != nil {
		return false, err
	}
	return exists, nil
}

func (r *authRepository) UsernameExistsInsensitive(username string) (bool, error) {
	orm, err := r.ormWithContext()
	if err != nil {
		return false, err
	}
	var exists bool
	if err := orm.Raw("SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(username) = LOWER(?))", username).Scan(&exists).Error; err != nil {
		return false, err
	}
	return exists, nil
}

func (r *authRepository) EmailExists(email string) (bool, error) {
	orm, err := r.ormWithContext()
	if err != nil {
		return false, err
	}
	var exists bool
	if err := orm.Raw("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email).Scan(&exists).Error; err != nil {
		return false, err
	}
	return exists, nil
}

func (r *authRepository) EmailExistsInsensitive(email string) (bool, error) {
	orm, err := r.ormWithContext()
	if err != nil {
		return false, err
	}
	var exists bool
	if err := orm.Raw("SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(email) = LOWER(?))", email).Scan(&exists).Error; err != nil {
		return false, err
	}
	return exists, nil
}

func (r *authRepository) CreateUser(username, email, hashedPassword string, now time.Time) (*model.User, error) {
	orm, err := r.ormWithContext()
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Username:  username,
		Email:     email,
		Password:  hashedPassword,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := orm.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *authRepository) GetUserByID(id uint) (*model.User, error) {
	orm, err := r.ormWithContext()
	if err != nil {
		return nil, err
	}
	var user model.User
	if err := orm.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) GetUserByUsername(username string) (*model.User, error) {
	orm, err := r.ormWithContext()
	if err != nil {
		return nil, err
	}
	var user model.User
	if err := orm.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) UpdateUser(userID uint, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}
	orm, err := r.ormWithContext()
	if err != nil {
		return err
	}
	if err := orm.Model(&model.User{}).Where("id = ?", userID).Updates(fields).Error; err != nil {
		return err
	}
	return nil
}

func (r *authRepository) UpdatePassword(userID uint, hashedPassword string, updatedAt time.Time) error {
	orm, err := r.ormWithContext()
	if err != nil {
		return err
	}
	updates := map[string]interface{}{
		"password":   hashedPassword,
		"updated_at": updatedAt,
	}
	return orm.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (r *authRepository) GetAvatarVersion(userID uint) (int, error) {
	orm, err := r.ormWithContext()
	if err != nil {
		return 0, err
	}
	var version sql.NullInt64
	if err := orm.Raw("SELECT avatar_version FROM users WHERE id = ?", userID).Scan(&version).Error; err != nil {
		return 0, err
	}
	if !version.Valid {
		return 0, nil
	}
	return int(version.Int64), nil
}

func (r *authRepository) UpdateAvatar(userID uint, relPath string, version int, updatedAt time.Time) error {
	orm, err := r.ormWithContext()
	if err != nil {
		return err
	}
	updates := map[string]interface{}{
		"avatar_path":       relPath,
		"avatar_version":    version,
		"avatar_updated_at": updatedAt,
		"updated_at":        updatedAt,
	}
	return orm.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (r *authRepository) ListRecentApplications(userID uint, limit int) ([]map[string]interface{}, error) {
	orm, err := r.ormWithContext()
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 10
	}
	rows, err := orm.Raw(`SELECT id, company_name, position_title, status, updated_at FROM job_applications WHERE user_id = ? ORDER BY updated_at DESC LIMIT ?`, userID, limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var id int
		var companyName, positionTitle, status string
		var updatedAt time.Time
		if err := rows.Scan(&id, &companyName, &positionTitle, &status, &updatedAt); err != nil {
			return nil, err
		}
		result = append(result, map[string]interface{}{
			"id":             id,
			"company_name":   companyName,
			"position_title": positionTitle,
			"status":         status,
			"updated_at":     updatedAt,
		})
	}
	return result, nil
}

func (r *authRepository) ListUpcomingInterviews(userID uint, limit int) ([]map[string]interface{}, error) {
	orm, err := r.ormWithContext()
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 5
	}
	rows, err := orm.Raw(`SELECT id, company_name, position_title, interview_time, interview_type FROM job_applications WHERE user_id = ? AND interview_time > NOW() AND interview_time <= NOW() + INTERVAL '7 days' ORDER BY interview_time ASC LIMIT ?`, userID, limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var id int
		var companyName, positionTitle string
		var interviewTime time.Time
		var interviewType sql.NullString
		if err := rows.Scan(&id, &companyName, &positionTitle, &interviewTime, &interviewType); err != nil {
			return nil, err
		}
		entry := map[string]interface{}{
			"id":             id,
			"company_name":   companyName,
			"position_title": positionTitle,
			"interview_time": interviewTime,
		}
		if interviewType.Valid {
			entry["interview_type"] = interviewType.String
		}
		result = append(result, entry)
	}
	return result, nil
}

func (r *authRepository) ListDailyStats(userID uint, days int) ([]map[string]interface{}, error) {
	orm, err := r.ormWithContext()
	if err != nil {
		return nil, err
	}
	if days <= 0 {
		days = 30
	}
	rows, err := orm.Raw(`SELECT DATE(created_at) as date, COUNT(*) as count FROM job_applications WHERE user_id = ? AND created_at >= CURRENT_DATE - INTERVAL '1 day' * ? GROUP BY DATE(created_at) ORDER BY date DESC`, userID, days).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var date time.Time
		var count int
		if err := rows.Scan(&date, &count); err != nil {
			return nil, err
		}
		result = append(result, map[string]interface{}{
			"date":  date.Format("2006-01-02"),
			"count": count,
		})
	}
	return result, nil
}
