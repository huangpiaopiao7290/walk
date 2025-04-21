package utils_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	user_config "walk/apps/user/config"
	utils "walk/shared/common/utils"

	"gorm.io/gorm"
)

func initConfig() *user_config.Database {
	filePath := "/home/pp/programs/program_go/timeTrack/walk/apps/user/config/user-service.yml"
	// Ensure the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("The file %s does not exist", filePath)
		return nil
	}
	cfg, err := user_config.InitUserConfig(filePath) // Initialize user configuration
	if err != nil {
		log.Printf("Failed to initialize user configuration: %v", err)
		return nil
	}

	return &cfg.Database
}

func TestConnDB(t *testing.T) {
	dbConfig := initConfig()
	if dbConfig == nil {
		t.Error("Failed to get database configuration from user configuration yaml file")
		return
	}
	
	// 连接数据库
	db, err := utils.NewDBConnection(&utils.DBConfig{
		Host:   dbConfig.Host,
		Port:   fmt.Sprintf("%d", dbConfig.Port),
		User:   dbConfig.Username,
		Passwd: dbConfig.Password,
		DBName: dbConfig.DBname,
	})
	if err != nil {
		t.Error("Failed to connect to the database")
		return
	}

	// 显示数据库连接信息
	log.Printf("Database connection information: %v", db)
}

func TestConnDBPool(t *testing.T) {
	dbConfig := initConfig()
	if dbConfig == nil {
		t.Error("Failed to get database configuration from user configuration yaml file")
		return
	}

	// 创建数据库连接池
	db, err := utils.NewDBConnection(&utils.DBConfig{
		Host:   dbConfig.Host,
		Port:   fmt.Sprintf("%d", dbConfig.Port),
		User:   dbConfig.Username,
		Passwd: dbConfig.Password,
		DBName: dbConfig.DBname,
	})
	if err != nil {
		t.Error("Failed to connect to the database")
		return
	}	

	// 设置数据库连接池
	err = utils.SetupDBConnectionPool(db, 10, 5, 100)	
	if err != nil {
		t.Errorf("Failed to set up the database connection pool: %v", err)
		return
	}
	
	// 显示数据库连接池状态
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			status := utils.ShowPoolStatus(db)
			if status != nil {
				t.Errorf("Failed to show the database connection pool status: %v", status)
				return
			}
		}
	}()

	// 模拟并发请求
	t.Run("concurrent requests", func(t *testing.T) {
		const (
			concurrentRequests = 50		// 模拟并发请求个数
			testIterations     = 100	// 每次请求执行次数
		)

		var wg sync.WaitGroup
		wg.Add(concurrentRequests)

		for i := range concurrentRequests {
			go func(id int) {
				defer wg.Done()
				for j := range testIterations {
                    // 执行测试查询
                    if err := performTestQuery(db); err != nil {
                        t.Errorf("协程%d-%d查询失败: %v", id, j, err)
                        return
                    }
				}
			}(i)
		}

		wg.Wait()
	})

	// 等待日志输出完成
	time.Sleep(1 * time.Second)

	
	// 验证连接池状态
	sqlDB, err := db.DB()
	if err != nil {
		t.Errorf("Failed to get underlying SQL DB: %v", err)
		return
	}

	stats := sqlDB.Stats()
	if stats.OpenConnections > 10 {
		t.Errorf("Open connections exceed max limit: %d", stats.OpenConnections)
	}
	if stats.WaitCount == 0 {
		t.Errorf("No wait count observed, expected some waits under high load")
	}


}

func performTestQuery(db *gorm.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 执行查询
	var result int
	if err := db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error; err != nil {
		return fmt.Errorf("查询失败: %v", err)
	}
	if result != 1 {
		return fmt.Errorf("查询结果不正确: %d", result)
	}

	// 模拟耗时操作
	time.Sleep(100 * time.Millisecond) // 每次查询后延迟 100ms
		
	return nil
}