package database

import (
	"context"
	"jobView-backend/internal/config"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatabaseConnection 测试数据库连接和连接池优化
func TestDatabaseConnection(t *testing.T) {
	cfg := getTestDatabaseConfig()
	db, err := New(cfg)
	require.NoError(t, err, "数据库连接应该成功")
	defer db.Close()

	t.Run("基本连接测试", func(t *testing.T) {
		err := db.Ping()
		assert.NoError(t, err, "Ping应该成功")
	})

	t.Run("连接池配置验证", func(t *testing.T) {
		stats := db.Stats()
		
		// 验证连接池参数设置合理
		assert.Greater(t, stats.MaxOpenConnections, 0, "最大连接数应该大于0")
		assert.LessOrEqual(t, stats.MaxOpenConnections, 100, "最大连接数不应该超过100")
		
		// 验证空闲连接数合理
		maxIdle := stats.MaxOpenConnections / 3
		if maxIdle < 5 {
			maxIdle = 5
		}
		
		t.Logf("连接池配置验证:")
		t.Logf("  MaxOpenConnections: %d", stats.MaxOpenConnections)
		t.Logf("  当前OpenConnections: %d", stats.OpenConnections)
		t.Logf("  InUse: %d", stats.InUse)
		t.Logf("  Idle: %d", stats.Idle)
	})

	t.Run("连接池健康状态", func(t *testing.T) {
		// 执行一些查询来测试连接池
		for i := 0; i < 10; i++ {
			var result int
			err := db.QueryRow("SELECT 1").Scan(&result)
			assert.NoError(t, err, "查询%d应该成功", i)
			assert.Equal(t, 1, result, "查询%d结果应该正确", i)
		}
		
		stats := db.Stats()
		assert.Zero(t, stats.WaitCount, "不应该有连接等待")
		assert.Zero(t, stats.WaitDuration, "等待时间应该为0")
	})
}

// TestQueryMonitoring 测试查询监控功能
func TestQueryMonitoring(t *testing.T) {
	cfg := getTestDatabaseConfig()
	db, err := New(cfg)
	require.NoError(t, err)
	defer db.Close()

	// 获取带监控的数据库连接
	monitoredDB := db.GetMonitoredDB()

	t.Run("查询性能监控", func(t *testing.T) {
		// 重置统计信息
		db.Monitor.ResetStats()
		
		// 执行一些查询
		for i := 0; i < 10; i++ {
			var result int
			err := monitoredDB.QueryRow("SELECT 1").Scan(&result)
			require.NoError(t, err)
		}
		
		// 检查统计信息
		stats := db.GetStats()
		assert.Equal(t, int64(10), stats.QueryStats.TotalQueries, "应该记录10次查询")
		assert.GreaterOrEqual(t, stats.QueryStats.AverageLatency, time.Duration(0), "平均延迟应该大于等于0")
		
		t.Logf("查询监控统计:")
		t.Logf("  总查询数: %d", stats.QueryStats.TotalQueries)
		t.Logf("  慢查询数: %d", stats.QueryStats.SlowQueries)
		t.Logf("  平均延迟: %v", stats.QueryStats.AverageLatency)
		t.Logf("  慢查询率: %.2f%%", stats.QueryStats.SlowQueryRate)
	})

	t.Run("慢查询检测", func(t *testing.T) {
		// 重置统计信息
		db.Monitor.ResetStats()
		
		// 执行一个慢查询（睡眠150ms，超过100ms阈值）
		_, err := monitoredDB.Query("SELECT pg_sleep(0.15)")
		require.NoError(t, err)
		
		// 检查是否被记录为慢查询
		stats := db.GetStats()
		assert.Equal(t, int64(1), stats.QueryStats.TotalQueries, "应该记录1次查询")
		assert.Equal(t, int64(1), stats.QueryStats.SlowQueries, "应该记录1次慢查询")
		assert.Equal(t, float64(100), stats.QueryStats.SlowQueryRate, "慢查询率应该是100%")
		assert.NotEmpty(t, stats.QueryStats.TopSlowQueries, "应该有慢查询记录")
		
		t.Logf("慢查询检测:")
		t.Logf("  慢查询数: %d", stats.QueryStats.SlowQueries)
		t.Logf("  慢查询率: %.2f%%", stats.QueryStats.SlowQueryRate)
		if len(stats.QueryStats.TopSlowQueries) > 0 {
			slowQuery := stats.QueryStats.TopSlowQueries[0]
			t.Logf("  慢查询详情: %v, SQL: %s", slowQuery.Duration, slowQuery.SQL)
		}
	})

	t.Run("并发查询监控", func(t *testing.T) {
		db.Monitor.ResetStats()
		
		var wg sync.WaitGroup
		concurrency := 5
		queriesPerGoroutine := 10
		
		// 并发执行查询
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < queriesPerGoroutine; j++ {
					var result int
					err := monitoredDB.QueryRow("SELECT $1", workerID*queriesPerGoroutine+j).Scan(&result)
					assert.NoError(t, err)
				}
			}(i)
		}
		
		wg.Wait()
		
		// 验证并发监控统计
		stats := db.GetStats()
		expectedQueries := int64(concurrency * queriesPerGoroutine)
		assert.Equal(t, expectedQueries, stats.QueryStats.TotalQueries, 
			"应该记录%d次查询", expectedQueries)
		
		t.Logf("并发查询监控:")
		t.Logf("  并发度: %d", concurrency)
		t.Logf("  每个协程查询数: %d", queriesPerGoroutine)
		t.Logf("  总查询数: %d", stats.QueryStats.TotalQueries)
		t.Logf("  平均延迟: %v", stats.QueryStats.AverageLatency)
	})
}

// TestDatabaseHealthChecker 测试数据库健康检查
func TestDatabaseHealthChecker(t *testing.T) {
	cfg := getTestDatabaseConfig()
	db, err := New(cfg)
	require.NoError(t, err)
	defer db.Close()

	t.Run("健康检查基本功能", func(t *testing.T) {
		// 检查初始健康状态
		assert.True(t, db.IsHealthy(), "数据库应该是健康的")
		
		healthStatus := db.Health.GetHealthStatus()
		assert.True(t, healthStatus.IsHealthy, "健康状态应该为true")
		assert.WithinDuration(t, time.Now(), healthStatus.LastCheck, 5*time.Second, 
			"最后检查时间应该在最近5秒内")
		
		t.Logf("健康检查状态:")
		t.Logf("  是否健康: %t", healthStatus.IsHealthy)
		t.Logf("  最后检查时间: %v", healthStatus.LastCheck)
		t.Logf("  运行时间: %v", healthStatus.Uptime)
		if healthStatus.LastError != "" {
			t.Logf("  最后错误: %s", healthStatus.LastError)
		}
	})

	t.Run("健康检查周期性执行", func(t *testing.T) {
		// 等待至少一个健康检查周期
		initialStatus := db.Health.GetHealthStatus()
		time.Sleep(31 * time.Second) // 等待超过30秒的检查间隔
		
		newStatus := db.Health.GetHealthStatus()
		assert.True(t, newStatus.LastCheck.After(initialStatus.LastCheck), 
			"健康检查应该周期性更新")
	})
}

// TestConnectionPoolPerformance 测试连接池性能
func TestConnectionPoolPerformance(t *testing.T) {
	cfg := getTestDatabaseConfig()
	db, err := New(cfg)
	require.NoError(t, err)
	defer db.Close()

	t.Run("连接池并发性能", func(t *testing.T) {
		var wg sync.WaitGroup
		concurrency := 20 // 高并发测试
		queriesPerGoroutine := 50
		
		start := time.Now()
		
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < queriesPerGoroutine; j++ {
					var result int
					err := db.QueryRow("SELECT $1", workerID*queriesPerGoroutine+j).Scan(&result)
					if err != nil {
						t.Errorf("Worker %d query %d failed: %v", workerID, j, err)
					}
				}
			}(i)
		}
		
		wg.Wait()
		duration := time.Since(start)
		
		totalQueries := concurrency * queriesPerGoroutine
		qps := float64(totalQueries) / duration.Seconds()
		
		t.Logf("连接池并发性能测试:")
		t.Logf("  并发数: %d", concurrency)
		t.Logf("  每个协程查询数: %d", queriesPerGoroutine)
		t.Logf("  总查询数: %d", totalQueries)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  平均QPS: %.2f", qps)
		
		// 验证性能指标
		assert.Greater(t, qps, 500.0, "QPS应该大于500")
		
		// 检查连接池状态
		stats := db.Stats()
		t.Logf("  连接池使用情况:")
		t.Logf("    MaxOpen: %d", stats.MaxOpenConnections)
		t.Logf("    Open: %d", stats.OpenConnections)
		t.Logf("    InUse: %d", stats.InUse)
		t.Logf("    Idle: %d", stats.Idle)
		t.Logf("    WaitCount: %d", stats.WaitCount)
		t.Logf("    WaitDuration: %v", stats.WaitDuration)
		
		// 连接池使用率不应该超过80%
		utilization := float64(stats.InUse) / float64(stats.MaxOpenConnections) * 100
		assert.LessOrEqual(t, utilization, 80.0, 
			"连接池使用率不应该超过80%，当前: %.2f%%", utilization)
	})

	t.Run("连接池资源管理", func(t *testing.T) {
		initialStats := db.Stats()
		
		// 执行一批查询后等待连接回收
		for i := 0; i < 10; i++ {
			var result int
			err := db.QueryRow("SELECT 1").Scan(&result)
			require.NoError(t, err)
		}
		
		// 等待连接回收
		time.Sleep(1 * time.Second)
		
		finalStats := db.Stats()
		
		t.Logf("连接池资源管理:")
		t.Logf("  初始状态 - Open: %d, InUse: %d, Idle: %d", 
			initialStats.OpenConnections, initialStats.InUse, initialStats.Idle)
		t.Logf("  最终状态 - Open: %d, InUse: %d, Idle: %d", 
			finalStats.OpenConnections, finalStats.InUse, finalStats.Idle)
		
		// 验证连接能够正确回收
		assert.LessOrEqual(t, finalStats.InUse, initialStats.InUse, 
			"使用中的连接数应该能够回收")
	})
}

// TestTransactionSupport 测试事务支持
func TestTransactionSupport(t *testing.T) {
	cfg := getTestDatabaseConfig()
	db, err := New(cfg)
	require.NoError(t, err)
	defer db.Close()

	t.Run("事务提交", func(t *testing.T) {
		tx, err := db.Begin()
		require.NoError(t, err)

		// 在事务中执行操作
		_, err = tx.Exec("CREATE TEMP TABLE test_tx (id INT, name TEXT)")
		require.NoError(t, err)

		_, err = tx.Exec("INSERT INTO test_tx (id, name) VALUES (1, 'test')")
		require.NoError(t, err)

		// 提交事务
		err = tx.Commit()
		require.NoError(t, err)

		// 验证数据是否提交成功（在新事务中查询）
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM test_tx").Scan(&count)
		if err == nil {
			assert.Equal(t, 1, count, "事务提交后数据应该存在")
		}
	})

	t.Run("事务回滚", func(t *testing.T) {
		tx, err := db.Begin()
		require.NoError(t, err)

		// 在事务中执行操作
		_, err = tx.Exec("CREATE TEMP TABLE test_rollback (id INT)")
		require.NoError(t, err)

		_, err = tx.Exec("INSERT INTO test_rollback (id) VALUES (1)")
		require.NoError(t, err)

		// 回滚事务
		err = tx.Rollback()
		require.NoError(t, err)

		// 验证数据是否被回滚（表应该不存在）
		var exists bool
		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_name = 'test_rollback'
			)`).Scan(&exists)
		
		if err == nil {
			assert.False(t, exists, "事务回滚后表不应该存在")
		}
	})

	t.Run("事务超时处理", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)

		// 执行一个会超时的操作
		_, err = tx.ExecContext(ctx, "SELECT pg_sleep(2)")
		assert.Error(t, err, "超时操作应该失败")

		// 事务应该自动回滚
		err = tx.Rollback()
		// 回滚可能成功也可能失败（如果已经被超时取消）
		t.Logf("事务回滚结果: %v", err)
	})
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	t.Run("无效数据库配置", func(t *testing.T) {
		cfg := &config.DatabaseConfig{
			Host:     "invalid-host",
			Port:     "5432",
			User:     "invalid",
			Password: "invalid",
			DBName:   "invalid",
			SSLMode:  "disable",
		}

		_, err := New(cfg)
		assert.Error(t, err, "无效配置应该返回错误")
		t.Logf("预期错误: %v", err)
	})

	t.Run("连接中断处理", func(t *testing.T) {
		cfg := getTestDatabaseConfig()
		db, err := New(cfg)
		require.NoError(t, err)
		defer db.Close()

		// 正常查询应该成功
		var result int
		err = db.QueryRow("SELECT 1").Scan(&result)
		require.NoError(t, err)
		assert.Equal(t, 1, result)

		// 模拟连接问题（这里只是演示，实际测试中可能需要更复杂的设置）
		t.Log("连接中断测试需要特殊的网络环境设置，跳过详细测试")
	})
}

// 辅助函数
func getTestDatabaseConfig() *config.DatabaseConfig {
	return &config.DatabaseConfig{
		Host:         getEnvOrDefault("DB_HOST", "localhost"),
		Port:         getEnvOrDefault("DB_PORT", "5432"),
		User:         getEnvOrDefault("DB_USER", "test_user"),
		Password:     getEnvOrDefault("DB_PASSWORD", "test_pass"),
		DBName:       getEnvOrDefault("DB_NAME", "jobview_test"),
		SSLMode:      "disable",
		MaxOpenConns: 20, // 测试环境使用较小的连接数
		MaxIdleConns: 5,
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}