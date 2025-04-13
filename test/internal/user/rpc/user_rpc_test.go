package rpc_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"

	user_config "walk/configs/user"
	user_service "walk/internal/user/service"
	user_model "walk/internal/user/model"
	user_repo "walk/internal/user/repo"
	utils "walk/pkg/utils"
	user_pb "walk/rpc/user"
)

var db *gorm.DB

// 初始化用户服务配置
func init() {
	// 初始化配置信息
	filePath := "/home/pp/programs/program_go/timeTrack/walk/configs/user-service.yml"
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

// 启动grpc服务端
func startTestServer() (*grpc.Server, net.Listener) {
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// 初始化 userRepo
	userRepo := user_repo.NewUserRepo[user_model.User](db)
	userService := &user_service.UserService{
		UserRepo: userRepo,
	}
	s := grpc.NewServer()
	user_pb.RegisterUserServiceServer(s, userService)

	go func() {
		if err := s.Serve(l); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	return s, l
}

func TestUserServiceRegister(t *testing.T) {
	server, listener := startTestServer()
	defer func() {
		server.Stop()
		listener.Close()
	}()

	conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}

	defer conn.Close()

	client := user_pb.NewUserServiceClient(conn)

	// 测试用例1：email未被注册
	req1 := &user_pb.RegisterRequest{
		Uname:     "testuser1",
		Email:     "test@example.com",
		Password:  "password123",
		IpAddress: "127.0.0.1",
	}
	// 打印req1
	log.Printf("req1: %+v", req1)

	resp1, err := client.Register(context.Background(), req1)

	if err != nil {
		t.Errorf("Register failed: %v", err)
	}
	if resp1 == nil {
		t.Errorf("Register response is nil")
	}

	fmt.Println("================================================")

	// 测试用例2：email已被注册
	req2 := &user_pb.RegisterRequest{
		Uname:     "testuser2",
		Email:     "test@example.com",
		Password:  "password123",
		IpAddress: "127.0.0.1",
	}

	// 打印req2
	log.Printf("req2: %+v", req2)

	resp2, err := client.Register(context.Background(), req2)
	if err == nil {
		t.Errorf("Register should fail when email is already registered")
	}
	if resp2 == nil {
		t.Errorf("Register response should be nil when email is already registered")
	}

}
