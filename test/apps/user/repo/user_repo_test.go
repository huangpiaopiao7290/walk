package repo_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"gorm.io/gorm"
	user_config "walk/apps/user/config"
	user_model "walk/apps/user/model"
	user_repo "walk/apps/user/repo"
	utils "walk/shared/common/utils"
)

// var dbConfig *user_config.Database

var db *gorm.DB
var repo user_repo.UserRepo[user_model.User]

func init() {
	// 初始化配置信息
	filePath := "/home/pp/programs/program_go/timeTrack/walk/apps/user/config/user-service.yml"
	// Ensure the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("The file %s does not exist", filePath)
	}
	var err error
	cfg, err := user_config.InitUserConfig(filePath)
	if err != nil {
		log.Printf("Failed to initialize user configuration: %v", err)
	}

	db, err = utils.NewDBConnection(&utils.DBConfig{
		Host:   cfg.Database.Host,
		Port:   fmt.Sprintf("%d", cfg.Database.Port),
		User:   cfg.Database.Username,
		Passwd: cfg.Database.Password,
		DBName: cfg.Database.DBname,
	})

	if err != nil {
		log.Printf("Failed to create database connection: %v", err)
	}
	log.Printf("Database connection information: %v", db)

	// 配置连接池
	err = utils.SetupDBConnectionPool(db, 10, 5, 60)
	if err != nil {
		log.Fatalf("Failed to configure connection pool: %v", err)
	}

}

// 测试create方法
func TestCreate(t *testing.T) {
	// 开启事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	defer func() {
		// 这里还可以使用docker启动临时数据库, 测试完自动销毁
		if !t.Failed() {
			db.Exec("ALTER TABLE users AUTO_INCREMENT = 1")
		}
	}()
	// 初始化repo接口
	repo = user_repo.NewUserRepo[user_model.User](tx)

	// 模拟user映射
	user := user_model.User{
		Uname:       "test1",
		Email:       "test1@test1.com",
		Pwd:         "test1",
		PhoneNumber: "1234567890",
		RegisterIP:  "127.0.0.1",
		LastLoginIP: "127.0.0.1",
		Avatar:      "https://www.google.com",
	}

	// insert into
	err := repo.Create(&user)
	if err != nil {
		t.Errorf("Create error: %v", err)
	}

	// 验证插入的数据
	if user.ID == 0 {
		t.Errorf("Create error: ID is not set")
	}
	if user.CreatedAt.IsZero() || user.UpdatedAt.IsZero() {
		t.Errorf("Create error: CreatedAt is not set")
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		t.Errorf("Failed to commit transaction: %v", err)
	}
}

// 测试createInBatches方法
func TestCreateInBatches(t *testing.T) {
	// 开启事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	defer func() {
		// 这里还可以使用docker启动临时数据库, 测试完自动销毁
		if !t.Failed() {
			db.Exec("ALTER TABLE users AUTO_INCREMENT = 1")
		}
	}()
	// 初始化repo接口
	repo = user_repo.NewUserRepo[user_model.User](tx)

	// 模拟批量用户
	users := []user_model.User{
		{
			Uname:       "batch1",
			Email:       "batch1@test.com",
			Pwd:         "password",
			PhoneNumber: "1234567890",
			RegisterIP:  "127.0.0.1",
			LastLoginIP: "127.0.0.1",
			Avatar:      "https://example.com/avatar1.jpg",
		},
		{
			Uname:       "batch2",
			Email:       "batch2@test.com",
			Pwd:         "password",
			PhoneNumber: "0987654321",
			RegisterIP:  "127.0.0.1",
			LastLoginIP: "127.0.0.1",
			Avatar:      "https://example.com/avatar2.jpg",
		},
	}

	// 批量插入数据
	err := repo.CreateInBatches(users, 2)
	if err != nil {
		t.Errorf("CreateInBatches error: %v", err)
	}

	// 验证插入的数据
	var count int64
	if err := tx.Model(&user_model.User{}).Count(&count).Error; err != nil {
		t.Errorf("Failed to count users: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 users, got %d", count)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		t.Errorf("Failed to commit transaction: %v", err)
	}

	// 打印操作数
	log.Printf("Number of rows affected: %d", count)
}

// 测试getByID方法
func TestGetByID(t *testing.T) {
	// 开启事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	defer func() {
		// 这里还可以使用docker启动临时数据库, 测试完自动销毁
		if !t.Failed() {
			db.Exec("ALTER TABLE users AUTO_INCREMENT = 1")
		}
	}()

	// 初始化repo接口
	repo = user_repo.NewUserRepo[user_model.User](tx)

	// 插入一条记录
	user := user_model.User{
		Uname:       "Bob",
		Email:       "bob@example.com",
		Pwd:         "password",
		PhoneNumber: "0987654321",
		RegisterIP:  "127.0.0.1",
		LastLoginIP: "127.0.0.1",
		Avatar:      "https://example.com/avatar.jpg",
	}
	err := repo.Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 查询记录
	result, err := repo.GetByID(user.ID)
	if err != nil {
		t.Errorf("GetByID error: %v", err)
	}
	if result.Uname != user.Uname {
		t.Errorf("Expected username '%s', got '%s'", user.Uname, result.Uname)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		t.Errorf("Failed to commit transaction: %v", err)
	}

	// 打印结果
	log.Printf("Result: %+v", result)
}

// 测试getByFields方法
func TestGetByFields(t *testing.T) {
	// 开启事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	defer func() {
		// 这里还可以使用docker启动临时数据库, 测试完自动销毁
		if !t.Failed() {
			db.Exec("ALTER TABLE users AUTO_INCREMENT = 1")
		}
	}()

	// 初始化repo接口
	repo = user_repo.NewUserRepo[user_model.User](tx)

	// 插入一条记录
	user := user_model.User{
		Uname:       "ccdog",
		Email:       "ccdog@example.com",
		Pwd:         "password",
		PhoneNumber: "0987654321",
		RegisterIP:  "127.0.0.1",
		LastLoginIP: "127.0.0.1",
		Avatar:      "https://example.com/avatar.jpg",
	}
	err := repo.Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 查询记录
	fields := map[string]interface{}{
		"uname": user.Uname,
		"email": user.Email,
	}
	result, err := repo.GetByFields(fields)
	if err != nil {
		t.Errorf("GetByFields error: %v", err)
	}
	if result.Uname != user.Uname {
		t.Errorf("Expected username '%s', got '%s'", user.Uname, result.Uname)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		t.Errorf("Failed to commit transaction: %v", err)
	}

	log.Printf("Result: %+v", result)
}

// 测试 Update 方法
func TestUpdate(t *testing.T) {
	// 开启事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	defer func() {
		// 这里还可以使用docker启动临时数据库, 测试完自动销毁
		if !t.Failed() {
			db.Exec("ALTER TABLE users AUTO_INCREMENT = 1")
		}
	}()

	// 初始化repo接口
	repo = user_repo.NewUserRepo[user_model.User](tx)

	// 插入一条记录
	user := user_model.User{
		Uname:       "Charlie",
		Email:       "charlie@example.com",
		Pwd:         "password",
		PhoneNumber: "1234567890",
		RegisterIP:  "127.0.0.1",
		LastLoginIP: "127.0.0.1",
		Avatar:      "https://example.com/avatar.jpg",
	}
	err := repo.Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 指定要更新的字段
	fields := []string{"uname"}
	user.Uname = "Charlie Updated"
	// 更新记录
	err = repo.UpdateByUid(user.UUID, &user, fields)
	if err != nil {
		t.Errorf("Update error: %v", err)
	}

	// 查询更新后的记录
	result, err := repo.GetByID(user.ID)
	if err != nil {
		t.Errorf("Failed to get user by ID: %v", err)
	}
	if result.Uname != "Charlie Updated" {
		t.Errorf("Expected username 'Charlie Updated', got '%s'", result.Uname)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		t.Errorf("Failed to commit transaction: %v", err)
	}

	log.Printf("Updated user: %+v", result)
}

// 测试 Delete 方法
func TestDelete(t *testing.T) {
	// 开启事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	defer func() {
		// 这里还可以使用docker启动临时数据库, 测试完自动销毁
		if !t.Failed() {
			db.Exec("ALTER TABLE users AUTO_INCREMENT = 1")
		}
	}()

	// 插入一条记录
	user := user_model.User{
		Uname:       "David",
		Email:       "david@example.com",
		Pwd:         "password",
		PhoneNumber: "0987654321",
		RegisterIP:  "127.0.0.1",
		LastLoginIP: "127.0.0.1",
		Avatar:      "https://example.com/avatar.jpg",
	}
	err := repo.Create(&user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 删除记录
	err = repo.DeleteByUid(user.UUID)
	if err != nil {
		t.Errorf("Delete error: %v", err)
	}

	// 验证删除结果
	_, err = repo.GetByID(user.ID)
	if err == nil {
		t.Errorf("Expected error when fetching deleted user, but got nil")
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		t.Errorf("Failed to commit transaction: %v", err)
	}

	log.Printf("Deleted user ID: %d", user.ID)
}
