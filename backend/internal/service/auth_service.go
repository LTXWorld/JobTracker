// /Users/lutao/GolandProjects/jobView/backend/internal/service/auth_service.go
// 认证服务，负责用户注册、登录、密码验证等核心认证业务逻辑
// 与数据库交互，处理用户认证相关的所有业务规则和安全控制

package service

import (
    "database/sql"
    "fmt"
    "jobView-backend/internal/auth"
    "jobView-backend/internal/database"
    "jobView-backend/internal/model"
    "jobView-backend/internal/utils"
    "log"
    "io"
    "mime/multipart"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

    "golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db *database.DB
}

func NewAuthService(db *database.DB) *AuthService {
	return &AuthService{db: db}
}

// Register 用户注册
func (s *AuthService) Register(req *model.RegisterRequest) (*model.LoginResponse, error) {
	// 验证输入
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, err
	}
	
	// 检查用户名是否已存在
	if exists, err := s.usernameExists(req.Username); err != nil {
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	} else if exists {
		return nil, fmt.Errorf("用户名已存在")
	}
	
	// 检查邮箱是否已存在
	if exists, err := s.emailExists(req.Email); err != nil {
		return nil, fmt.Errorf("检查邮箱失败: %w", err)
	} else if exists {
		return nil, fmt.Errorf("邮箱已被注册")
	}
	
	// 加密密码
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}
	
	// 创建用户
	user, err := s.createUser(req.Username, req.Email, hashedPassword)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}
	
	// 生成token
	accessToken, refreshToken, err := auth.GenerateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %w", err)
	}
	
	// 记录注册日志
	log.Printf("[AUTH] New user registered: ID=%d, Username=%s, Email=%s", 
		user.ID, user.Username, user.Email)
	
	return &model.LoginResponse{
		User:         user.ToProfile(),
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(auth.AccessTokenDuration).Unix(),
	}, nil
}

// Login 用户登录
func (s *AuthService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	// 验证输入
	if err := s.validateLoginRequest(req); err != nil {
		return nil, err
	}
	
	// 查找用户
	user, err := s.getUserByUsername(req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("用户名或密码错误")
		}
		return nil, fmt.Errorf("查找用户失败: %w", err)
	}
	
	// 验证密码
	if err := s.verifyPassword(req.Password, user.Password); err != nil {
		// 记录登录失败日志
		log.Printf("[AUTH] Login failed for user: %s - incorrect password", req.Username)
		return nil, fmt.Errorf("用户名或密码错误")
	}
	
	// 生成token
	accessToken, refreshToken, err := auth.GenerateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %w", err)
	}
	
	// 记录登录成功日志
	log.Printf("[AUTH] User logged in successfully: ID=%d, Username=%s", 
		user.ID, user.Username)
	
	return &model.LoginResponse{
		User:         user.ToProfile(),
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(auth.AccessTokenDuration).Unix(),
	}, nil
}

// RefreshToken 刷新token
func (s *AuthService) RefreshToken(req *model.RefreshTokenRequest) (*model.LoginResponse, error) {
	// 验证刷新token
	claims, err := auth.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("无效的刷新token: %w", err)
	}
	
	// 获取用户信息
	user, err := s.getUserByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}
	
	// 生成新的token对
	accessToken, refreshToken, err := auth.GenerateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %w", err)
	}
	
	log.Printf("[AUTH] Token refreshed for user: ID=%d, Username=%s", 
		user.ID, user.Username)
	
	return &model.LoginResponse{
		User:         user.ToProfile(),
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(auth.AccessTokenDuration).Unix(),
	}, nil
}

// GetProfile 获取用户信息
func (s *AuthService) GetProfile(userID uint) (*model.UserProfile, error) {
	user, err := s.getUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	
	return user.ToProfile(), nil
}

// UpdateProfile 更新用户信息
func (s *AuthService) UpdateProfile(userID uint, req *model.UpdateUserRequest) (*model.UserProfile, error) {
	// 验证输入
	if err := s.validateUpdateUserRequest(req); err != nil {
		return nil, err
	}
	
	// 检查用户是否存在
	user, err := s.getUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}
	
	// 构建更新SQL
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1
	
	if req.Username != nil && *req.Username != user.Username {
		// 检查新用户名是否已存在
		if exists, err := s.usernameExists(*req.Username); err != nil {
			return nil, fmt.Errorf("检查用户名失败: %w", err)
		} else if exists {
			return nil, fmt.Errorf("用户名已存在")
		}
		
		setParts = append(setParts, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, *req.Username)
		argIndex++
	}
	
	if req.Email != nil && *req.Email != user.Email {
		// 检查新邮箱是否已存在
		if exists, err := s.emailExists(*req.Email); err != nil {
			return nil, fmt.Errorf("检查邮箱失败: %w", err)
		} else if exists {
			return nil, fmt.Errorf("邮箱已被注册")
		}
		
		setParts = append(setParts, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, *req.Email)
		argIndex++
	}
	
	if len(setParts) == 0 {
		return user.ToProfile(), nil
	}
	
	// 添加更新时间和用户ID
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++
	
	args = append(args, userID)
	
	query := fmt.Sprintf(`
		UPDATE users 
		SET %s 
		WHERE id = $%d
	`, fmt.Sprintf("%s", setParts[0]), argIndex)
	
	for i := 1; i < len(setParts); i++ {
		query = fmt.Sprintf("%s, %s", query[:len(query)-len(fmt.Sprintf("WHERE id = $%d", argIndex))], setParts[i]) + 
			fmt.Sprintf(" WHERE id = $%d", argIndex)
	}
	
	// 重构查询字符串
	query = fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", 
		fmt.Sprintf("%s", setParts[0]), argIndex)
	if len(setParts) > 1 {
		for i := 1; i < len(setParts); i++ {
			query = fmt.Sprintf("UPDATE users SET %s, %s WHERE id = $%d", 
				fmt.Sprintf("%s", setParts[0]), setParts[i], argIndex)
		}
	}
	
	// 执行更新
	_, err = s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("更新用户信息失败: %w", err)
	}
	
	log.Printf("[AUTH] User profile updated: ID=%d", userID)
	
	// 返回更新后的用户信息
	updatedUser, err := s.getUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取更新后的用户信息失败: %w", err)
	}
	
	return updatedUser.ToProfile(), nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userID uint, req *model.ChangePasswordRequest) error {
	// 验证输入
	if err := utils.ValidatePassword(req.NewPassword); err != nil {
		return err
	}
	
	// 获取用户信息
	user, err := s.getUserByID(userID)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}
	
	// 验证当前密码
	if err := s.verifyPassword(req.CurrentPassword, user.Password); err != nil {
		return fmt.Errorf("当前密码错误")
	}
	
	// 检查新密码是否与当前密码相同
	if err := s.verifyPassword(req.NewPassword, user.Password); err == nil {
		return fmt.Errorf("新密码不能与当前密码相同")
	}
	
	// 加密新密码
	hashedPassword, err := s.hashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}
	
	// 更新密码
	query := "UPDATE users SET password = $1, updated_at = $2 WHERE id = $3"
	_, err = s.db.Exec(query, hashedPassword, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}
	
	log.Printf("[AUTH] Password changed for user: ID=%d", userID)
	
	return nil
}

// UpdateAvatar 保存用户头像并更新数据库记录
// 返回可访问的URL（/static 前缀）与版本号
func (s *AuthService) UpdateAvatar(userID uint, file multipart.File, header *multipart.FileHeader) (string, int, error) {
    // 探测类型
    var buf [512]byte
    n, _ := file.Read(buf[:])
    contentType := http.DetectContentType(buf[:n])
    // 支持常见图片
    allowed := map[string]string{
        "image/jpeg": ".jpg",
        "image/png":  ".png",
        "image/webp": ".webp",
        "image/gif":  ".gif",
    }
    ext, ok := allowed[contentType]
    if !ok {
        // 从文件名后缀兜底
        ext = strings.ToLower(filepath.Ext(header.Filename))
        if ext == "" {
            ext = ".jpg"
        }
    }
    // 复位读指针
    if _, err := file.Seek(0, 0); err != nil {
        return "", 0, fmt.Errorf("failed to seek: %w", err)
    }

    // 读取当前版本
    var currentVersion int
    _ = s.db.QueryRow("SELECT COALESCE(avatar_version,0) FROM users WHERE id=$1", userID).Scan(&currentVersion)
    newVersion := currentVersion + 1

    // 目录与文件名
    baseDir := "./uploads"
    relDir := filepath.Join("avatars", fmt.Sprintf("%d", userID))
    if err := os.MkdirAll(filepath.Join(baseDir, relDir), 0o755); err != nil {
        return "", 0, fmt.Errorf("failed to create dir: %w", err)
    }
    filename := fmt.Sprintf("avatar_v%d%s", newVersion, ext)
    absPath := filepath.Join(baseDir, relDir, filename)

    // 原子写入：先写临时文件，再重命名
    tmp := absPath + ".tmp"
    out, err := os.Create(tmp)
    if err != nil {
        return "", 0, fmt.Errorf("failed to create file: %w", err)
    }
    if _, err := io.Copy(out, file); err != nil {
        out.Close()
        os.Remove(tmp)
        return "", 0, fmt.Errorf("failed to write file: %w", err)
    }
    out.Close()
    if err := os.Rename(tmp, absPath); err != nil {
        os.Remove(tmp)
        return "", 0, fmt.Errorf("failed to rename file: %w", err)
    }

    relPath := filepath.ToSlash(filepath.Join(relDir, filename))

    // 更新数据库
    _, err = s.db.Exec(`UPDATE users SET avatar_path=$1, avatar_version=$2, avatar_updated_at=NOW(), updated_at=NOW() WHERE id=$3`, relPath, newVersion, userID)
    if err != nil {
        return "", 0, fmt.Errorf("failed to update user avatar: %w", err)
    }

    url := "/static/" + relPath + fmt.Sprintf("?v=%d", newVersion)
    return url, newVersion, nil
}

// validateRegisterRequest 验证注册请求
func (s *AuthService) validateRegisterRequest(req *model.RegisterRequest) error {
	if err := utils.ValidateUsername(req.Username); err != nil {
		return err
	}
	
	if err := utils.ValidateEmail(req.Email); err != nil {
		return err
	}
	
	if err := utils.ValidatePassword(req.Password); err != nil {
		return err
	}
	
	return nil
}

// validateLoginRequest 验证登录请求
func (s *AuthService) validateLoginRequest(req *model.LoginRequest) error {
	if req.Username == "" {
		return fmt.Errorf("用户名不能为空")
	}
	
	if req.Password == "" {
		return fmt.Errorf("密码不能为空")
	}
	
	return nil
}

// validateUpdateUserRequest 验证更新用户请求
func (s *AuthService) validateUpdateUserRequest(req *model.UpdateUserRequest) error {
	if req.Username != nil {
		if err := utils.ValidateUsername(*req.Username); err != nil {
			return err
		}
	}
	
	if req.Email != nil {
		if err := utils.ValidateEmail(*req.Email); err != nil {
			return err
		}
	}
	
	return nil
}

// usernameExists 检查用户名是否存在
func (s *AuthService) usernameExists(username string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)"
	var exists bool
	err := s.db.QueryRow(query, username).Scan(&exists)
	return exists, err
}

// emailExists 检查邮箱是否存在
func (s *AuthService) emailExists(email string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"
	var exists bool
	err := s.db.QueryRow(query, email).Scan(&exists)
	return exists, err
}

// hashPassword 加密密码
func (s *AuthService) hashPassword(password string) (string, error) {
	// 使用bcrypt加密，cost=12提供更高安全性
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// verifyPassword 验证密码
func (s *AuthService) verifyPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// createUser 创建新用户
func (s *AuthService) createUser(username, email, hashedPassword string) (*model.User, error) {
	query := `
		INSERT INTO users (username, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	
	user := &model.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
	}
	
	err := s.db.QueryRow(query, username, email, hashedPassword, time.Now(), time.Now()).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

// getUserByID 根据ID获取用户
func (s *AuthService) getUserByID(id uint) (*model.User, error) {
	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	
	user := &model.User{}
	err := s.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.CreatedAt, &user.UpdatedAt,
	)
	
	return user, err
}

// getUserByUsername 根据用户名获取用户
func (s *AuthService) getUserByUsername(username string) (*model.User, error) {
	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	
	user := &model.User{}
	err := s.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.CreatedAt, &user.UpdatedAt,
	)
	
	return user, err
}

// IsUsernameAvailable 检查用户名是否可用
func (s *AuthService) IsUsernameAvailable(username string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(username) = LOWER($1))`
	
	err := s.db.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("database query failed: %w", err)
	}
	
	return !exists, nil
}

// IsEmailAvailable 检查邮箱是否可用
func (s *AuthService) IsEmailAvailable(email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(email) = LOWER($1))`
	
	err := s.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("database query failed: %w", err)
	}
	
	return !exists, nil
}
