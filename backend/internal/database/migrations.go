package database

import (
	"fmt"
	"log"
)

// RunMigrations 运行数据库迁移
func (db *DB) RunMigrations() error {
	log.Println("Running database migrations...")

	// 首先创建用户表
	createUsersTable := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`

	if _, err := db.Exec(createUsersTable); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// 检查job_applications表是否存在
	hasTable, err := db.checkTableExists("job_applications")
	if err != nil {
		return fmt.Errorf("failed to check job_applications table: %w", err)
	}
	
	if !hasTable {
		// 表不存在，创建完整的表
		createJobApplicationsTable := `
			CREATE TABLE job_applications (
				id SERIAL PRIMARY KEY,
				user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				company_name VARCHAR(255) NOT NULL,
				position_title VARCHAR(255) NOT NULL,
				application_date VARCHAR(10) NOT NULL,
				status VARCHAR(50) NOT NULL DEFAULT '已投递',
				job_description TEXT,
				salary_range VARCHAR(100),
				work_location VARCHAR(255),
				contact_info VARCHAR(500),
				notes TEXT,
				interview_time TIMESTAMP WITH TIME ZONE,
				reminder_time TIMESTAMP WITH TIME ZONE,
				reminder_enabled BOOLEAN DEFAULT FALSE,
				follow_up_date VARCHAR(10),
				hr_name VARCHAR(255),
				hr_phone VARCHAR(255),
				hr_email VARCHAR(255),
				interview_location VARCHAR(255),
				interview_type VARCHAR(255),
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			);
		`

		if _, err := db.Exec(createJobApplicationsTable); err != nil {
			return fmt.Errorf("failed to create job_applications table: %w", err)
		}
	} else {
		// 表存在，检查是否有user_id列
		hasUserIDColumn, err := db.checkColumnExists("job_applications", "user_id")
		if err != nil {
			return fmt.Errorf("failed to check user_id column: %w", err)
		}
		
		if !hasUserIDColumn {
			// 添加user_id列
			if _, err := db.Exec("ALTER TABLE job_applications ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;"); err != nil {
				return fmt.Errorf("failed to add user_id column: %w", err)
			}
		}
		
		// 添加其他可能缺失的字段
		alterTableSQL := []string{
			"ALTER TABLE job_applications ADD COLUMN IF NOT EXISTS interview_time TIMESTAMP WITH TIME ZONE;",
			"ALTER TABLE job_applications ADD COLUMN IF NOT EXISTS reminder_time TIMESTAMP WITH TIME ZONE;",
			"ALTER TABLE job_applications ADD COLUMN IF NOT EXISTS reminder_enabled BOOLEAN DEFAULT FALSE;",
			"ALTER TABLE job_applications ADD COLUMN IF NOT EXISTS follow_up_date VARCHAR(10);",
			"ALTER TABLE job_applications ADD COLUMN IF NOT EXISTS hr_name VARCHAR(255);",
			"ALTER TABLE job_applications ADD COLUMN IF NOT EXISTS hr_phone VARCHAR(255);",
			"ALTER TABLE job_applications ADD COLUMN IF NOT EXISTS hr_email VARCHAR(255);",
			"ALTER TABLE job_applications ADD COLUMN IF NOT EXISTS interview_location VARCHAR(255);",
			"ALTER TABLE job_applications ADD COLUMN IF NOT EXISTS interview_type VARCHAR(255);",
		}

		for _, alterSQL := range alterTableSQL {
			if _, err := db.Exec(alterSQL); err != nil {
				log.Printf("Warning: Failed to alter table (may already exist): %v", err)
			}
		}
	}

	// 创建索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);",
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);",
		"CREATE INDEX IF NOT EXISTS idx_job_applications_user_id ON job_applications(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_job_applications_application_date ON job_applications(application_date);",
		"CREATE INDEX IF NOT EXISTS idx_job_applications_status ON job_applications(status);",
		"CREATE INDEX IF NOT EXISTS idx_job_applications_company_name ON job_applications(company_name);",
		"CREATE INDEX IF NOT EXISTS idx_job_applications_user_status ON job_applications(user_id, status);",
		"CREATE INDEX IF NOT EXISTS idx_job_applications_reminder_time ON job_applications(reminder_time) WHERE reminder_enabled = TRUE;",
	}

	for _, indexSQL := range indexes {
		if _, err := db.Exec(indexSQL); err != nil {
			log.Printf("Warning: Failed to create index (may already exist): %v", err)
			// 继续执行，因为索引可能已经存在
		}
	}

	// 检查是否需要创建默认用户（用于开发环境）
	if err := db.createDefaultUserIfNeeded(); err != nil {
		log.Printf("Warning: Failed to create default user: %v", err)
		// 不返回错误，因为这不是关键功能
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// checkTableExists 检查表是否存在
func (db *DB) checkTableExists(tableName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_name = $1
	`
	
	var count int
	err := db.QueryRow(query, tableName).Scan(&count)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// checkColumnExists 检查表中是否存在指定列
func (db *DB) checkColumnExists(tableName, columnName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.columns 
		WHERE table_name = $1 AND column_name = $2
	`
	
	var count int
	err := db.QueryRow(query, tableName, columnName).Scan(&count)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// createDefaultUserIfNeeded 在开发环境创建默认用户
func (db *DB) createDefaultUserIfNeeded() error {
	// 检查是否已有用户
	var userCount int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		return err
	}
	
	// 如果没有用户，创建一个默认的测试用户
	if userCount == 0 {
		// 使用bcrypt加密默认密码
		hashedPassword := "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewqCQx.Lk5FYR7.G" // 密码: TestPass123!
		
		query := `
			INSERT INTO users (username, email, password, created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW())
		`
		
		_, err = db.Exec(query, "testuser", "test@example.com", hashedPassword)
		if err != nil {
			return err
		}
		
		log.Println("Created default test user: username=testuser, password=TestPass123!")
	}
	
	return nil
}