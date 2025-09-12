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

	// 为用户表添加头像相关字段（如果不存在）
	avatarCols := []string{
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_path VARCHAR(255);",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_etag VARCHAR(64);",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_version INTEGER DEFAULT 0;",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_updated_at TIMESTAMP WITH TIME ZONE;",
	}
	for _, stmt := range avatarCols {
		if _, err := db.Exec(stmt); err != nil {
			log.Printf("Warning: Failed to add user avatar column: %v", err)
		}
	}



	// 兼容旧版数据库：如果使用了 PostgreSQL enum 类型 application_status，
	// 确保枚举包含所有最新状态值，避免更新到新状态时报错 500（invalid input value for enum）。
	if err := db.ensureApplicationStatusEnumValues(); err != nil {
		// 不阻断启动流程，仅记录警告，避免在非 PostgreSQL 或未使用该 enum 的环境中失败
		log.Printf("Warning: failed to ensure application_status enum values: %v", err)
	}

	// 检查是否需要创建默认用户（用于开发环境）
	if err := db.createDefaultUserIfNeeded(); err != nil {
		log.Printf("Warning: Failed to create default user: %v", err)
		// 不返回错误，因为这不是关键功能
	}

	// 检查并创建 export_tasks 表
	if err := db.createExportTasksTable(); err != nil {
		return fmt.Errorf("failed to create export_tasks table: %w", err)
	}

	// 创建简历相关表
	if err := db.createResumeTables(); err != nil {
		return fmt.Errorf("failed to create resume tables: %w", err)
	}

    // 确保状态流转校验函数存在且支持应用层放行回退（基于GUC）
    if err := db.ensureStatusTransitionFunctions(); err != nil {
        log.Printf("Warning: failed to ensure status transition functions: %v", err)
    }

    log.Println("Database migrations completed successfully")
    return nil
}

// ensureApplicationStatusEnumValues 在数据库存在 application_status 枚举时，
// 将缺失的状态值补齐（幂等）。
func (db *DB) ensureApplicationStatusEnumValues() error {
    // 检查是否存在名为 application_status 的枚举类型
    var exists bool
    checkEnumSQL := `
        SELECT EXISTS (
            SELECT 1
            FROM pg_type t
            WHERE t.typname = 'application_status'
              AND t.typtype = 'e' -- enum type
        )
    `
    if err := db.QueryRow(checkEnumSQL).Scan(&exists); err != nil {
        return err
    }
    if !exists {
        // 当前数据库不是使用 enum（例如使用 VARCHAR），无需处理
        return nil
    }

    // 需要补齐的新增状态（与后端枚举保持一致）
    values := []string{
        "简历筛选未通过",
        "笔试未通过",
        "一面未通过",
        "二面未通过",
        "三面未通过",
        "HR面未通过",
    }

    // 使用 DO $$ ... $$ + 条件判断，兼容低版本 PG（没有 ADD VALUE IF NOT EXISTS）
    for _, v := range values {
        stmt := fmt.Sprintf(`
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum e
        JOIN pg_type t ON e.enumtypid = t.oid
        WHERE t.typname = 'application_status' AND e.enumlabel = '%s'
    ) THEN
        ALTER TYPE application_status ADD VALUE '%s';
    END IF;
END $$;`, v, v)

        if _, err := db.Exec(stmt); err != nil {
            // 记录警告但不中断，以免影响应用启动
            log.Printf("Warning: failed adding enum value '%s' to application_status: %v", v, err)
        }
    }
    return nil
}

// ensureStatusTransitionFunctions 确保存在 validate_status_transition 函数，
// 并内置基于会话GUC变量 jobview.allow_backward 的回退放行能力。
func (db *DB) ensureStatusTransitionFunctions() error {
    // 检查 job_applications 是否存在
    hasJA, err := db.checkTableExists("job_applications")
    if err != nil {
        return err
    }
    if !hasJA {
        return nil
    }

    // 使用 VARCHAR 版本的函数定义，兼容未创建 enum 的环境
    stmt := `
CREATE OR REPLACE FUNCTION validate_status_transition(
    p_user_id INTEGER,
    p_old_status VARCHAR,
    p_new_status VARCHAR,
    p_flow_template_id INTEGER DEFAULT NULL
) RETURNS BOOLEAN AS $$
DECLARE
    v_allowed_transitions JSONB;
    v_flow_config JSONB;
    v_allow TEXT;
BEGIN
    -- 应用层放行：当会话设置 jobview.allow_backward = 'on' 时，直接允许
    v_allow := current_setting('jobview.allow_backward', true);
    IF COALESCE(v_allow, '') = 'on' THEN
        RETURN TRUE;
    END IF;

    -- 初始状态允许
    IF p_old_status IS NULL THEN
        RETURN TRUE;
    END IF;

    -- 相同状态不允许
    IF p_old_status = p_new_status THEN
        RETURN FALSE;
    END IF;

    -- 读取默认模板
    SELECT flow_config::jsonb INTO v_flow_config
    FROM status_flow_templates 
    WHERE (p_flow_template_id IS NOT NULL AND id = p_flow_template_id)
       OR (p_flow_template_id IS NULL AND is_default = TRUE)
    LIMIT 1;

    -- 无配置则放行
    IF v_flow_config IS NULL THEN
        RETURN TRUE;
    END IF;

    -- 检查转换列表（按旧状态键取出允许的目标数组）
    v_allowed_transitions := v_flow_config->'transitions'->(p_old_status);
    IF v_allowed_transitions IS NULL THEN
        RETURN TRUE;
    END IF;

    IF v_allowed_transitions ? p_new_status THEN
        RETURN TRUE;
    END IF;

    -- 内置直通规则补充：笔试中->一面中->二面中->三面中->HR面中
    IF p_old_status = '笔试中' AND p_new_status = '一面中' THEN
        RETURN TRUE;
    ELSIF p_old_status = '一面中' AND p_new_status = '二面中' THEN
        RETURN TRUE;
    ELSIF p_old_status = '二面中' AND p_new_status = '三面中' THEN
        RETURN TRUE;
    ELSIF p_old_status = '三面中' AND p_new_status = 'HR面中' THEN
        RETURN TRUE;
    END IF;

    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;
`

    if _, err := db.Exec(stmt); err != nil {
        return err
    }

    // 如果存在 application_status 枚举类型，则再创建一个重载版本以匹配触发器定义
    var hasEnum bool
    if err := db.QueryRow(`SELECT EXISTS (
        SELECT 1 FROM pg_type t WHERE t.typname = 'application_status' AND t.typtype = 'e'
    )`).Scan(&hasEnum); err == nil && hasEnum {
        stmtEnum := `
CREATE OR REPLACE FUNCTION validate_status_transition(
    p_user_id INTEGER,
    p_old_status application_status,
    p_new_status application_status,
    p_flow_template_id INTEGER DEFAULT NULL
) RETURNS BOOLEAN AS $$
DECLARE
    v_allowed_transitions JSONB;
    v_flow_config JSONB;
    v_allow TEXT;
BEGIN
    v_allow := current_setting('jobview.allow_backward', true);
    IF COALESCE(v_allow, '') = 'on' THEN
        RETURN TRUE;
    END IF;

    IF p_old_status IS NULL THEN
        RETURN TRUE;
    END IF;

    IF p_old_status = p_new_status THEN
        RETURN FALSE;
    END IF;

    SELECT flow_config::jsonb INTO v_flow_config
    FROM status_flow_templates 
    WHERE (p_flow_template_id IS NOT NULL AND id = p_flow_template_id)
       OR (p_flow_template_id IS NULL AND is_default = TRUE)
    LIMIT 1;

    IF v_flow_config IS NULL THEN
        RETURN TRUE;
    END IF;

    v_allowed_transitions := v_flow_config->'transitions'->(p_old_status::text);
    IF v_allowed_transitions IS NOT NULL AND v_allowed_transitions ? p_new_status::text THEN
        RETURN TRUE;
    END IF;

    -- 内置直通规则补充
    IF p_old_status::text = '笔试中' AND p_new_status::text = '一面中' THEN
        RETURN TRUE;
    ELSIF p_old_status::text = '一面中' AND p_new_status::text = '二面中' THEN
        RETURN TRUE;
    ELSIF p_old_status::text = '二面中' AND p_new_status::text = '三面中' THEN
        RETURN TRUE;
    ELSIF p_old_status::text = '三面中' AND p_new_status::text = 'HR面中' THEN
        RETURN TRUE;
    END IF;

    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;`
        if _, err := db.Exec(stmtEnum); err != nil {
            // 不阻断启动
            log.Printf("Warning: failed to create enum overload for validate_status_transition: %v", err)
        }
    }

    // 覆盖触发器函数：支持基于 GUC 跳过写历史（jobview.skip_history='on'）
    triggerFn := `
CREATE OR REPLACE FUNCTION trigger_job_status_change() 
RETURNS TRIGGER AS $$
DECLARE
    v_old_status VARCHAR;
    v_duration_minutes INTEGER;
    v_status_history JSONB;
    v_history_entry JSONB;
    v_skip TEXT;
BEGIN
    v_old_status := OLD.status;

    -- 跳过：当设置跳过历史时，不做任何改动（不更新 last_status_change/version/历史）
    v_skip := current_setting('jobview.skip_history', true);
    IF COALESCE(v_skip, '') = 'on' THEN
        RETURN NEW;
    END IF;

    IF NEW.status = OLD.status THEN
        RETURN NEW;
    END IF;

    -- 若未通过 validate_status_transition 放行则拒绝
    IF NOT validate_status_transition(NEW.user_id, v_old_status, NEW.status) THEN
        RAISE EXCEPTION '不允许的状态转换: % -> %', v_old_status, NEW.status;
    END IF;

    -- 计算停留时长
    IF OLD.last_status_change IS NOT NULL THEN
        v_duration_minutes := EXTRACT(EPOCH FROM (NOW() - OLD.last_status_change)) / 60;
    ELSE
        v_duration_minutes := EXTRACT(EPOCH FROM (NOW() - OLD.created_at)) / 60;
    END IF;

    -- 更新 last_status_change 与版本
    NEW.last_status_change := NOW();
    NEW.status_version := COALESCE(OLD.status_version, 0) + 1;

    -- 记录历史
    INSERT INTO job_status_history (
        job_application_id, user_id, old_status, new_status, status_changed_at, duration_minutes, metadata
    ) VALUES (
        NEW.id, NEW.user_id, v_old_status, NEW.status, NOW(), v_duration_minutes,
        COALESCE(NEW.status_history->'current_metadata', '{}')
    );

    -- 更新 status_history JSONB 摘要
    v_status_history := COALESCE(NEW.status_history, '{"history": [], "summary": {}}'::jsonb);
    v_history_entry := jsonb_build_object(
        'timestamp', extract(epoch from NOW()),
        'old_status', v_old_status,
        'new_status', NEW.status,
        'duration_minutes', v_duration_minutes,
        'changed_at', NOW()::text
    );
    v_status_history := jsonb_set(v_status_history, '{history}', (v_status_history->'history') || v_history_entry);
    v_status_history := jsonb_set(
        v_status_history,
        '{summary}',
        jsonb_build_object(
            'total_changes', jsonb_array_length(v_status_history->'history'),
            'current_status', NEW.status,
            'last_changed', NOW()::text,
            'total_duration_minutes', COALESCE((v_status_history->'summary'->>'total_duration_minutes')::INTEGER, 0) + v_duration_minutes
        )
    );
    NEW.status_history := v_status_history;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;`

    if _, err := db.Exec(triggerFn); err != nil {
        log.Printf("Warning: failed to ensure trigger_job_status_change(): %v", err)
    }
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

// createExportTasksTable 创建导出任务表
func (db *DB) createExportTasksTable() error {
	// 检查 export_tasks 表是否存在
	hasTable, err := db.checkTableExists("export_tasks")
	if err != nil {
		return fmt.Errorf("failed to check export_tasks table: %w", err)
	}
	
	if hasTable {
		return nil // 表已存在，跳过创建
	}

	// 创建 export_tasks 表的 SQL
	createTableSQL := `
		CREATE TABLE export_tasks (
			id SERIAL PRIMARY KEY,
			task_id VARCHAR(100) UNIQUE NOT NULL,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			export_type VARCHAR(20) NOT NULL DEFAULT 'xlsx',
			filename VARCHAR(255),
			file_path VARCHAR(500),
			file_size BIGINT,
			total_records INTEGER,
			processed_records INTEGER DEFAULT 0,
			progress INTEGER DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
			filters JSONB,
			options JSONB,
			error_message TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			expires_at TIMESTAMP,
			CONSTRAINT valid_status CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled', 'expired')),
			CONSTRAINT valid_export_type CHECK (export_type IN ('xlsx', 'csv')),
			CONSTRAINT valid_progress_range CHECK (progress >= 0 AND progress <= 100),
			CONSTRAINT valid_records_count CHECK (
				(total_records IS NULL OR total_records >= 0) AND
				(processed_records >= 0) AND
				(total_records IS NULL OR processed_records <= total_records)
			)
		);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		return fmt.Errorf("failed to create export_tasks table: %w", err)
	}

	// 创建索引
	indexes := []string{
		"CREATE INDEX idx_export_tasks_user_id ON export_tasks(user_id);",
		"CREATE INDEX idx_export_tasks_status ON export_tasks(status);",
		"CREATE INDEX idx_export_tasks_created_at ON export_tasks(created_at);",
		"CREATE INDEX idx_export_tasks_task_id ON export_tasks(task_id);",
		"CREATE INDEX idx_export_tasks_expires_at ON export_tasks(expires_at) WHERE expires_at IS NOT NULL;",
		"CREATE INDEX idx_export_tasks_user_status ON export_tasks(user_id, status);",
		"CREATE INDEX IF NOT EXISTS idx_export_tasks_filters_gin ON export_tasks USING GIN (filters);",
		"CREATE INDEX IF NOT EXISTS idx_export_tasks_options_gin ON export_tasks USING GIN (options);",
	}

	for _, indexSQL := range indexes {
		if _, err := db.Exec(indexSQL); err != nil {
			log.Printf("Warning: Failed to create export_tasks index: %v", err)
		}
	}

	log.Println("Created export_tasks table and indexes")
	return nil
}

// createResumeTables 创建简历相关表
func (db *DB) createResumeTables() error {
    // resumes
    hasResumes, err := db.checkTableExists("resumes")
    if err != nil { return fmt.Errorf("failed to check resumes table: %w", err) }
    if !hasResumes {
        create := `
            CREATE TABLE resumes (
                id SERIAL PRIMARY KEY,
                user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                title VARCHAR(100) DEFAULT '默认简历',
                summary TEXT,
                privacy VARCHAR(20) DEFAULT 'private',
                current_version INTEGER DEFAULT 1,
                is_completed BOOLEAN DEFAULT FALSE,
                completeness INTEGER DEFAULT 0,
                created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
            );
        `
        if _, err := db.Exec(create); err != nil {
            return fmt.Errorf("failed to create resumes table: %w", err)
        }
        log.Println("Created table resumes")
    }

    // resume_sections
    hasSections, err := db.checkTableExists("resume_sections")
    if err != nil { return fmt.Errorf("failed to check resume_sections table: %w", err) }
    if !hasSections {
        create := `
            CREATE TABLE resume_sections (
                id SERIAL PRIMARY KEY,
                resume_id INTEGER NOT NULL REFERENCES resumes(id) ON DELETE CASCADE,
                type VARCHAR(30) NOT NULL,
                sort_order INTEGER DEFAULT 0,
                content JSONB NOT NULL,
                created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
            );
        `
        if _, err := db.Exec(create); err != nil {
            return fmt.Errorf("failed to create resume_sections table: %w", err)
        }
        log.Println("Created table resume_sections")
    }

    // resume_attachments
    hasAtt, err := db.checkTableExists("resume_attachments")
    if err != nil { return fmt.Errorf("failed to check resume_attachments table: %w", err) }
    if !hasAtt {
        create := `
            CREATE TABLE resume_attachments (
                id SERIAL PRIMARY KEY,
                resume_id INTEGER NOT NULL REFERENCES resumes(id) ON DELETE CASCADE,
                file_name VARCHAR(255) NOT NULL,
                file_path VARCHAR(500) NOT NULL,
                mime_type VARCHAR(100),
                file_size BIGINT,
                etag VARCHAR(64),
                created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
            );
        `
        if _, err := db.Exec(create); err != nil {
            return fmt.Errorf("failed to create resume_attachments table: %w", err)
        }
        log.Println("Created table resume_attachments")
    }

    // indexes
    idx := []string{
        "CREATE INDEX IF NOT EXISTS idx_resumes_user_id ON resumes(user_id)",
        "CREATE INDEX IF NOT EXISTS idx_sections_resume_type ON resume_sections(resume_id, type)",
    }
    for _, s := range idx {
        if _, err := db.Exec(s); err != nil {
            log.Printf("Warning: failed to create resume index: %v", err)
        }
    }
    return nil
}
