package database

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "runtime"
    "time"
    "jobView-backend/internal/config"

    _ "github.com/lib/pq"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    glogger "gorm.io/gorm/logger"
    "gorm.io/gorm/schema"
)

type DB struct {
    *sql.DB
    Monitor *QueryMonitor
    Health  *DatabaseHealthChecker
    ORM     *gorm.DB
    UseGorm bool
}

func New(cfg *config.DatabaseConfig) (*DB, error) {
	// 使用PostgreSQL URL格式确保正确连接
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 优化连接池配置
	optimizeConnectionPool(db, cfg)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database with DSN %s: %w", dsn, err)
	}

	// 创建监控器（慢查询阈值：100ms）
	logger := log.New(os.Stdout, "[DB-MONITOR] ", log.LstdFlags|log.Lshortfile)
	monitor := NewQueryMonitor(100*time.Millisecond, logger)
	
    // 初始包装对象（为健康检查准备）
    tmp := &DB{DB: db}
    // 创建健康检查器
    healthChecker := NewHealthChecker(tmp, 30*time.Second)
    healthChecker.StartHealthCheck()

    wrapper := &DB{
        DB:      db,
        Monitor: monitor,
        Health:  healthChecker,
        UseGorm: cfg.UseGorm,
    }
    // 回填到健康检查器
    tmp.Monitor = monitor
    tmp.Health = healthChecker

    // 初始化 GORM（可选）
    if cfg.UseGorm {
        orm, err := initGorm(db, cfg)
        if err != nil {
            return nil, fmt.Errorf("failed to init gorm: %w", err)
        }
        wrapper.ORM = orm
    }
    return wrapper, nil
}

// GetMonitoredDB 获取带监控的数据库连接
func (db *DB) GetMonitoredDB() *MonitoredDB {
	return db.Monitor.WrapDB(db.DB)
}

// GetStats 获取数据库性能统计
func (db *DB) GetStats() PerformanceStats {
	return db.Monitor.GetStats(db.DB.Stats())
}

// GetConnectionStats 获取连接池统计信息
func (db *DB) GetConnectionStats() map[string]interface{} {
	return GetConnectionPoolStats(db.DB)
}

// IsHealthy 检查数据库健康状态
func (db *DB) IsHealthy() bool {
	return db.Health.IsHealthy()
}

// optimizeConnectionPool 优化数据库连接池配置 - 高级调优版本
func optimizeConnectionPool(db *sql.DB, cfg *config.DatabaseConfig) {
	// 根据环境和负载情况调整连接池参数
	
	// 1. 计算合适的最大连接数
	// 生产环境：CPU核数 * 4，开发环境：CPU核数 * 2
	cpuCores := runtime.NumCPU()
	var maxOpenConns int
	
	if cfg.MaxOpenConns > 0 {
		// 如果配置中指定了最大连接数，使用配置值
		maxOpenConns = cfg.MaxOpenConns
	} else {
		// 自动计算最佳连接数
		if os.Getenv("ENVIRONMENT") == "production" {
			maxOpenConns = cpuCores * 4
		} else {
			maxOpenConns = cpuCores * 2
		}
		
		// 确保连接数在合理范围内
		if maxOpenConns < 10 {
			maxOpenConns = 10 // 最小10个连接
		}
		if maxOpenConns > 100 {
			maxOpenConns = 100 // 最大100个连接，避免数据库过载
		}
	}
	
	// 2. 计算空闲连接数
	// 空闲连接数 = 最大连接数的 25-50%，确保有足够的预热连接
	var maxIdleConns int
	if cfg.MaxIdleConns > 0 {
		maxIdleConns = cfg.MaxIdleConns
	} else {
		maxIdleConns = maxOpenConns / 3 // 约33%的空闲连接
		if maxIdleConns < 5 {
			maxIdleConns = 5 // 最小5个空闲连接
		}
	}
	
	// 3. 设置连接生命周期
	// 连接最大生命周期：避免长期连接导致的资源泄漏和数据库连接超时
	connMaxLifetime := 30 * time.Minute
	if os.Getenv("ENVIRONMENT") == "production" {
		connMaxLifetime = 60 * time.Minute // 生产环境连接生命周期更长
	}
	
	// 4. 设置连接空闲时间
	// 连接最大空闲时间：及时释放空闲连接，减少资源占用
	connMaxIdleTime := 15 * time.Minute
	if os.Getenv("ENVIRONMENT") == "production" {
		connMaxIdleTime = 30 * time.Minute // 生产环境空闲时间更长
	}
	
	// 应用连接池配置
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)
	
	// 记录连接池配置信息
	fmt.Printf("[DB-POOL] Connection pool optimized:\n")
	fmt.Printf("  - MaxOpenConns: %d\n", maxOpenConns)
	fmt.Printf("  - MaxIdleConns: %d\n", maxIdleConns)
	fmt.Printf("  - ConnMaxLifetime: %v\n", connMaxLifetime)
	fmt.Printf("  - ConnMaxIdleTime: %v\n", connMaxIdleTime)
	fmt.Printf("  - CPU Cores: %d\n", cpuCores)
	fmt.Printf("  - Environment: %s\n", os.Getenv("ENVIRONMENT"))
}

// initGorm 使用现有 *sql.DB 初始化 GORM，避免重复连接池
func initGorm(std *sql.DB, cfg *config.DatabaseConfig) (*gorm.DB, error) {
    // GORM 日志配置：默认慢查询 200ms
    logger := glogger.New(
        log.New(os.Stdout, "[GORM] ", log.LstdFlags),
        glogger.Config{
            SlowThreshold:             200 * time.Millisecond,
            LogLevel:                  glogger.Warn,
            IgnoreRecordNotFoundError: true,
            Colorful:                  false,
        },
    )

    dialector := postgres.New(postgres.Config{Conn: std})
    orm, err := gorm.Open(dialector, &gorm.Config{
        Logger: logger,
        NamingStrategy: schema.NamingStrategy{
            TablePrefix:   "",    // 保持现有表名
            SingularTable: true,   // 禁止复数
            NoLowerCase:   false,  // 使用 snake_case
        },
        DisableAutomaticPing: false,
        SkipDefaultTransaction: false,
    })
    if err != nil {
        return nil, err
    }
    return orm, nil
}
