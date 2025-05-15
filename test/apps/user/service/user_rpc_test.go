package rpc_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	user_config "walk/apps/user/config"
	user_model "walk/apps/user/model"
	user_repo "walk/apps/user/repo"
	user_pb "walk/apps/user/rpc"
	user_service "walk/apps/user/service"
	user_utils "walk/apps/user/utils"
	utils "walk/shared/common/utils"
)

var (
	db 			 *gorm.DB
	cfg 		 *user_config.UserConfig
	userRepo 	 user_repo.UserRepo[user_model.User]
	userRedisSct *user_utils.UserRedisSct
)
// 初始化用户服务配置
func init() {
	// 初始化配置信息
	filePath := "/home/pp/programs/program_go/timeTrack/walk/apps/user/config/user_service.yml"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("The file %s does not exist", filePath)
	}

	var err error
	cfg, err = user_config.InitUserConfig(filePath)
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

	// 初始化 userRepo
	userRepo = user_repo.NewUserRepo[user_model.User](db)

	// 初始化userRedisSct
	rdsConfig := &utils.RedisConfig{
		Network:  cfg.Redis.Network,
		Addr:     cfg.Redis.Addr,
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxIdleConns: cfg.Redis.MaxIdleConns,
		MaxRetries: cfg.Redis.MaxRetries,
	}
	rds, err := utils.NewRedisClient(*rdsConfig)
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	userRedisSct = user_utils.NewRedisHandler(rds)
	if userRedisSct == nil {
		log.Fatalf("Failed to create UserRedisSct: %v", err)
	}

}

// 启动grpc服务端
func startTestServer() (*grpc.Server, net.Listener) {
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	//创建用户服务
	userService := user_service.NewUserService(cfg, userRepo, userRedisSct)
	if userService == nil {
		log.Fatalf("Failed to create UserService")
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

// 注册测试
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

// 测试登录
func TestUserServiceLogin(t *testing.T) {
	// 启动测试服务器
	server, listener := startTestServer()
	defer func() {
		server.Stop()
		listener.Close()
	}()

	// 创建 gRPC 客户端连接
	conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	client := user_pb.NewUserServiceClient(conn)

	// 数据库中已有的用户数据
	existingUserEmail := "test@example.com"
	existingUserPassword := "password123" // 假设这是存储在数据库中的明文密码
	// wrongPassword := "wrongpassword"
	// 测试用例1：用户存在且密码正确时的登录成功
	loginReq1 := &user_pb.LoginRequest{
		Email:    existingUserEmail,
		Password: existingUserPassword,
	}
	log.Printf("Testing login with correct credentials: %+v", loginReq1)
	resp1, err := client.Login(context.Background(), loginReq1)
	if err != nil {
		t.Errorf("Login failed: %v", err)
	}
	if resp1 == nil {
		t.Fatalf("Expected non-nil response for successful login, but got nil")
	}
	if resp1.Uid == "" {
		t.Errorf("Expected non-empty UID in response, but got empty")
	}
	if resp1.AccessToken == "" {
		t.Errorf("Expected non-empty access token in response, but got empty")
	}
	if resp1.RefreshToken == "" {
		t.Errorf("Expected non-empty refresh token in response, but got empty")
	}
	if resp1.Email != existingUserEmail {
		t.Errorf("Unexpected email in response: got %s, want %s", resp1.Email, existingUserEmail)
	}
	if resp1.Uname != "testuser1" { // 数据库中 uname 为 testuser1
		t.Errorf("Unexpected username in response: got %s, want %s", resp1.Uname, "testuser1")
	}

	log.Printf("reps1: %+v", resp1)

	fmt.Println("================================================")

	// 从redis获取token
	storedRefreshToken, err := userRedisSct.GetWithTTL(context.Background(), resp1.Uid)
	if err != nil {
		t.Errorf("Failed to get token from Redis: %v", err)
	}
	log.Printf("Stored refresh token: %s", storedRefreshToken)


	// fmt.Println("================================================")

	// // 测试用例2：用户存在但密码错误时的登录失败
	// loginReq2 := &user_pb.LoginRequest{
	// 	Email:    existingUserEmail,
	// 	Password: wrongPassword,
	// }
	// log.Printf("Testing login with incorrect password: %+v", loginReq2)
	// resp2, err := client.Login(context.Background(), loginReq2)
	// if err == nil {
	// 	t.Errorf("Expected error for incorrect password, but got nil")
	// }
	// if status.Code(err) != codes.Unauthenticated {
	// 	t.Errorf("Expected code UNAUTHENTICATED, got %v", status.Code(err))
	// }
	// if resp2 != nil {
	// 	t.Errorf("Expected nil response for incorrect password, but got %+v", resp2)
	// }

	// log.Printf("resp2: %+v", resp2)

	// fmt.Println("================================================")

	// // 解析验证jwt token
	// t.Run("TestToken", func(t *testing.T) {
	// 	if resp1 != nil {
	// 		// 解析access token
	// 		claims, err := user_utils.ParseToken(resp1.AccessToken, &user_config.JWT{
	// 			Secret:           cfg.JWT.Secret,
	// 			Access_token_ttl: 1,
	// 		})

	// 		log.Printf("Parsed claims: %+v", claims)

	// 		if err != nil {
	// 			t.Errorf("Failed to parse access token: %v", err)
	// 		}
	// 		if claims == nil {
	// 			t.Errorf("Parsed claims should not be nil")
	// 		}
	// 		if claims != nil {
	// 			if claims.UUID != resp1.Uid {
	// 				t.Errorf("Expected UUID %s, got %s", resp1.Uid, claims.UUID)
	// 			}
	// 		} else {
	// 			t.Error("Parsed claims is nil")
	// 		}
	// 	} else {
	// 		t.Error("resp1 is nil")
	// 	}
	// })
}

// 测试获取用户信息
func TestUserServiceGetUser(t *testing.T) {
	// 启动测试服务器
	server, listener := startTestServer()
	defer func() {
		server.Stop()
		listener.Close()
	}()

	// 创建 gRPC 客户端连接
	conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()
	client := user_pb.NewUserServiceClient(conn)
	// 测试用例1：模拟登录获取token
	t.Run("TestLogin", func(t *testing.T) {
		loginReq := &user_pb.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		log.Printf("Testing login with correct credentials: %+v", loginReq)
		resp, err := client.Login(context.Background(), loginReq)
		if err != nil {
			t.Errorf("Login failed: %v", err)
		}
		if resp == nil {
			t.Fatalf("Expected non-nil response for successful login, but got nil")
		}
		if resp.Uid == "" {
			t.Errorf("Expected non-empty UID in response, but got empty")
		}
		if resp.AccessToken == "" {
			t.Errorf("Expected non-empty access token in response, but got empty")
		}
		if resp.RefreshToken == "" {
			t.Errorf("Expected non-empty refresh token in response, but got empty")
		}

		// 打印登录响应
		log.Printf("Login response: %+v", resp)

		getUserReq := &user_pb.GetUserRequest{
			Uid:   resp.Uid,
			Token: resp.AccessToken,
		}

		log.Printf("Testing get user info with token: %+v", getUserReq)

		// 根据token获取用户信息
		userResp, err := client.GetUser(context.Background(), getUserReq)
		if err != nil {
			t.Fatalf("GetUser failed: %v", err)
		}
		if userResp == nil {
			t.Fatalf("Expected non-nil response for successful GetUser, but got nil")
		}
		if userResp.Uid != getUserReq.Uid {
			t.Errorf("Expected UID %s, got %s", getUserReq.Uid, userResp.Uid)
		}
		// 打印用户信息
		log.Printf("GetUser response: %+v", userResp)
	})
}

// 更新用户信息测试
func TestUserServiceUpdateUser(t *testing.T) {
	// 启动测试服务器
	server, listener := startTestServer()
	defer func() {
		server.Stop()
		listener.Close()
	}()
	// 创建 gRPC 客户端连接
	conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()
	client := user_pb.NewUserServiceClient(conn)
	// 测试用例1: 获取当前登录用户
	t.Run("TestLogin", func(t *testing.T) {
		loginReq := &user_pb.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		log.Printf("Testing login with correct credentials: %+v", loginReq)
		resp, err := client.Login(context.Background(), loginReq)
		if err != nil {
			t.Errorf("Login failed: %v", err)
		}
		log.Printf("Login response: %+v", resp)

		// 创建前端
	})
}

// 测试逻辑删除用户
func TestUserServiceDeleteUser(t *testing.T) {
	// 启动测试服务器
	server, listener := startTestServer()
	defer func() {
		server.Stop()
		listener.Close()
	}()
	// 创建 gRPC 客户端连接
	conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()
	client := user_pb.NewUserServiceClient(conn)
	// 测试用例1: 获取当前登录用户
	t.Run("TestLogin", func(t *testing.T) {
		loginReq := &user_pb.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		resp, err := client.Login(context.Background(), loginReq)
		if err != nil {
			t.Errorf("Login failed: %v", err)
		}
		log.Printf("Login response: %+v", resp)

		// 逻辑删除用户
		deleteReq := &user_pb.DeleteUserRequest{
			Uid:   resp.Uid,
			Token: resp.AccessToken,
		}
		log.Printf("Testing delete user with token: %+v", deleteReq)
		deleteResp, err := client.DeleteUser(context.Background(), deleteReq)
		if err != nil {
			t.Errorf("DeleteUser failed: %v", err)
		}
		log.Printf("DeleteUser response: %+v", deleteResp)

		// 验证用户是否已逻辑删除
		getUserReq := &user_pb.GetUserRequest{
			Uid:   resp.Uid,
			Token: resp.AccessToken,
		}
		getUserResp, err := client.GetUser(context.Background(), getUserReq)
		if err == nil {
			t.Errorf("Expected error for deleted user, but got nil")
		}
		if status.Code(err) != codes.NotFound { // 添加验证错误码
			t.Errorf("Expected code NOT_FOUND, got %v", status.Code(err))
		}
		// 打印获取用户信息响应
		log.Printf("GetUser response: %+v", getUserResp)

	})
}

// 测试刷新令牌
func TestUserServiceRefreshToken(t *testing.T) {
	// 启动测试服务器
	server, listener := startTestServer()
	defer func() {
		server.Stop()
		listener.Close()
	}()
	// 创建 gRPC 客户端连接
	conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()
	client := user_pb.NewUserServiceClient(conn)

	// 测试正常情况
	t.Run("TestRefreshTokenSuccess", func(t *testing.T) {
		// 1. 先登录获取初始token
		loginReq := &user_pb.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		loginResp, err := client.Login(context.Background(), loginReq)
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}

		// 2. 使用获取到的token进行刷新
		refreshReq := &user_pb.RefreshTokenRequest{
			Uid:   loginResp.Uid,
			Token: loginResp.RefreshToken,
		}

		// 模拟操作延时
		time.Sleep(1 * time.Second)

		refreshResp, err := client.RefreshToken(context.Background(), refreshReq)
		if err != nil {
			t.Fatalf("RefreshToken failed: %v", err)
		}

		// 验证响应
		if refreshResp.AccessToken == "" {
			t.Error("Expected non-empty access token")
		}
		if refreshResp.RefreshToken == "" {
			t.Error("Expected non-empty refresh token")
		}
		if refreshResp.AccessToken == loginResp.AccessToken {
			t.Error("New access token should be different from old one")
		}
		if refreshResp.RefreshToken == loginResp.RefreshToken {
			t.Error("New refresh token should be different from old one")
		}

		log.Printf("Refresh token success: old tokens: %+v, new tokens: %+v",
			map[string]string{"access": loginResp.AccessToken, "refresh": loginResp.RefreshToken},
			map[string]string{"access": refreshResp.AccessToken, "refresh": refreshResp.RefreshToken})
	})

	t.Run("TestRefreshTokenWithMismatchedUID", func(t *testing.T) {
		// 1. 先登录获取有效token
		loginResp, err := client.Login(context.Background(), &user_pb.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}

		// 2. 使用错误的UID
		mismatchReq := &user_pb.RefreshTokenRequest{
			Uid:   "wrong-uid",
			Token: loginResp.RefreshToken,
		}
		_, err = client.RefreshToken(context.Background(), mismatchReq)
		if err == nil {
			t.Error("Expected error with mismatched UID, got nil")
		}
		log.Printf("Expected error received: %v", err)
	})

	// 性能测试
	t.Run("TestRefreshTokenPerformance", func(t *testing.T) {
		// 先登录获取有效token
		loginResp, err := client.Login(context.Background(), &user_pb.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}

		refreshReq := &user_pb.RefreshTokenRequest{
			Uid:   loginResp.Uid,
			Token: loginResp.RefreshToken,
		}

		// 执行性能测试
		iterations := 100
		start := time.Now()

		for i := 0; i < iterations; i++ {
			_, err := client.RefreshToken(context.Background(), refreshReq)
			if err != nil {
				t.Errorf("RefreshToken failed at iteration %d: %v", i, err)
				break
			}
		}

		duration := time.Since(start)
		averageTime := duration.Milliseconds() / int64(iterations)
		log.Printf("Performance test results:\n"+
			"Total iterations: %d\n"+
			"Total time: %v\n"+
			"Average time per request: %dms\n"+
			"Requests per second: %.2f",
			iterations, duration, averageTime,
			float64(iterations)/(float64(duration)/float64(time.Second)))

		// 验证性能指标
		maxAllowedAvgTime := int64(50) // 50ms
		if averageTime > maxAllowedAvgTime {
			t.Errorf("Average response time %dms exceeds maximum allowed time %dms",
				averageTime, maxAllowedAvgTime)
		}
	})
}
