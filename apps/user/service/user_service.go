// Auth: pp
// Created:
// Description:

package user_service

import (
	"context"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	user_config "walk/apps/user/config"
	user_model "walk/apps/user/model"
	user_repo "walk/apps/user/repo"
	user_pb "walk/apps/user/rpc"
	user_utils "walk/apps/user/utils"
)

type UserService struct {
	Cfg      		*user_config.UserConfig
	UserRepo 		user_repo.UserRepo[user_model.User]
	UserRedisSct  	*user_utils.UserRedisSct
	user_pb.UnimplementedUserServiceServer
}

var userPool = sync.Pool{
	New: func() any{
		return &user_model.User{}
	},
}

// @brief: 构造函数
// @param cfg *user_config.UserConfig: 用户配置
// @param userRepo user_repo.UserRepo[user_model.User]: 用户仓库
// @param userRedisSct *user_utils.UserRedisSct: 用户redis操作
// @return *UserService: 用户服务
func NewUserService(cfg *user_config.UserConfig, userRepo user_repo.UserRepo[user_model.User], userRedisSct *user_utils.UserRedisSct) *UserService {
	return &UserService{
		Cfg:      		cfg,
		UserRepo: 		userRepo,
		UserRedisSct:   userRedisSct,
	}
}

// register implement
// @Param ctx context.Context: context
// @Param req *user_proto.RegisterRequest: register request
// @Return *user_proto.RegisterResponse: register response
// @Return error: error
func (s *UserService) Register(ctx context.Context, req *user_pb.RegisterRequest) (*user_pb.RegisterResponse, error) {
	if s.UserRepo == nil {
		return nil, status.Errorf(codes.Internal, "UserRepo is not initialized")
	}

	// 开启事务
	tx, err := s.UserRepo.BeginTransaction(ctx)
	if err != nil {
		log.Printf("[UserService] Register error: %v", err)
		return nil, err
	}
	// 回滚事务
	defer func() {
		if err := recover(); err != nil {
			// 捕获panic并回滚
			s.UserRepo.RollbackTransaction(tx)
			log.Printf("[UserService] Register panic recover: %v", err)
		}
		if err != nil { // 如果有错误，回滚事务
			if rollbackErr := s.UserRepo.RollbackTransaction(tx); rollbackErr != nil {
				log.Printf("[UserService] Register error: failed to rollback transaction: %v", rollbackErr)
			}
		}
	}()
	
	// 检测邮箱格式是否正确
	if !user_utils.ValidateEmail(req.Email) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email format")
	}
	// 检测密码强度
	if !user_utils.ValidatePassword(req.Password) {
		return nil, status.Errorf(codes.InvalidArgument, "password must be at least 8 characters")
	}

	// 检测注册邮箱是否存在
	existingUser, getErr := s.UserRepo.GetByFields(map[string]any{"email": req.Email})
	// log.Printf("existingUser: %v, getErr: %v", existingUser, getErr)
	if getErr != nil {
		log.Printf("[UserService] Register error: %v", getErr)
		return nil, status.Errorf(codes.Internal, "failed to check email existence: %v", getErr)
	}
	if existingUser != nil {
		return nil, status.Errorf(codes.AlreadyExists, "email already exists")
	}

	// 对用户密码加密
	hashedPwd, hashErr := user_utils.HashPWD(req.Password)
	if hashErr != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password")
	}
	// 定义用户
	newUser := user_model.User{
		Uname:      req.Uname,
		Email:      req.Email,
		Pwd:        hashedPwd,
		RegisterIP: req.IpAddress,
	}

	// 创建用户
	if err := s.UserRepo.Create(&newUser); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user")
	}

	// 提交事务
	if err := s.UserRepo.CommitTransaction(tx); err != nil {
		log.Printf("[UserService] Register error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
	}

	return &user_pb.RegisterResponse{
		Uid:   newUser.UUID,
		Uname: newUser.Uname,
		Email: newUser.Email,
	}, nil
}

// login implement
// @Param ctx context.Context: context
// @Param req *user_proto.LoginRequest: login request
// @Return *user_proto.LoginResponse: login response
// @Return error: error
func (s *UserService) Login(ctx context.Context, req *user_pb.LoginRequest) (*user_pb.LoginResponse, error) {

	// 检测邮箱是否存在
	existingUser, err := s.UserRepo.GetByFields(map[string]any{"email": req.Email})
	if err != nil {
		log.Printf("[UserService] Login error: failed to get user by email: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch user data: %v", err)
	}
	if existingUser == nil {
		log.Printf("[UserService] Login error: user with email %s not found", req.Email)
		return nil, status.Errorf(codes.NotFound, "user with this email does not exist")
	}

	// log.Printf("existingUser: %v", existingUser)

	// 若用户存在，验证密码
	pwdErr := user_utils.ComparePWD(existingUser.Pwd, req.Password)
	if pwdErr != nil {
		log.Printf("[UserService] Login error: password mismatch for user %s: %v", req.Email, pwdErr)
		return nil, status.Errorf(codes.Unauthenticated, "invalid password")
	}

	// 根据用户id生成jwt
	accessToken, accessTokenErr := user_utils.GenerateAccessToken(&s.Cfg.JWT, existingUser.UUID)
	if accessTokenErr != nil {
		log.Printf("[UserService] Login error: failed to generate access token: %v", accessTokenErr)
		return nil, status.Errorf(codes.Internal, "failed to generate JWT token: %v", accessTokenErr)
	}

	refreshToken, refreshTokenErr := user_utils.GenerateRefreshToken(&s.Cfg.JWT, existingUser.UUID)
	if refreshTokenErr != nil {
		log.Printf("[UserService] Login error: failed to generate refresh token: %v", refreshTokenErr)
		return nil, status.Errorf(codes.Internal, "failed to generate refresh token: %v", refreshTokenErr)
	}

	// 将refreshToken存储到redis并设置过期时间
	err = s.UserRedisSct.SetWithTTL(ctx, existingUser.UUID, refreshToken, time.Duration(s.Cfg.JWT.Refresh_token_ttl)*time.Second)
	if err != nil {
		log.Printf("[UserService] Login error: failed to set refresh token in Redis: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to store refresh token in Redis: %v", err)
	} 
	// 返回登录成功的响应
	return &user_pb.LoginResponse{
		Uid:          existingUser.UUID,
		Uname:        existingUser.Uname,
		Email:        existingUser.Email,
		Avatar:       existingUser.Avatar,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// 退出登录接口实现
// @Param ctx context.Context: context
// @Param req *user_proto.LogoutRequest: logout request
// @Return *user_proto.LogoutResponse: logout response
// @Return error: error
func (s *UserService) Logout(ctx context.Context, req *user_pb.LogoutRequest) (*user_pb.LogoutResponse, error) {
	// 从redis中删除refreshToken
	err := s.UserRedisSct.Delete(ctx, req.Uid)
	if err != nil {
		log.Printf("[UserService] Logout error: failed to delete refresh token from Redis: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to delete refresh token from Redis: %v", err)
	}

	return &user_pb.LogoutResponse{
		Success: true,
	}, nil
}

// 获取用户信息
// @Param ctx context.Context: context
// @Param req *user_proto.GetUserRequest: get user request
// @Return *user_proto.GetUserResponse: get user response
// @Return error: error
func (s *UserService) GetUser(ctx context.Context, req *user_pb.GetUserRequest) (*user_pb.GetUserResponse, error) {

	// 查询用户啊hi否存在
	user, err := s.UserRepo.GetByFields(map[string]any{"uuid": req.Uid})
	if err != nil {
		log.Printf("[UserService] GetUserInfo error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch user data: %v", err)
	}
	if user == nil {
		log.Printf("[UserService] GetUserInfo error: user with UUID %s not found", req.Uid)
		return nil, status.Errorf(codes.NotFound, "user with this UUID does not exist")
	}
	// 返回用户信息
	return &user_pb.GetUserResponse{
		Uid:           user.UUID,
		Uname:         user.Uname,
		Email:         user.Email,
		Avatar:        user.Avatar,
		RegisterTime:  user.CreatedAt.Unix(),
		LastLoginTime: user.UpdatedAt.Unix(),
		LastLoginIp:   user.LastLoginIP,
	}, nil
}

// 更新用户信息
// @Param ctx context.Context: context
// @Param req *user_proto.UpdateUserRequest: update user request
// @Return *user_proto.UpdateUserResponse: update user response
// @Return error: error
func (s *UserService) UpdateUser(ctx context.Context, req *user_pb.UpdateUserRequest) (*user_pb.UpdateUserResponse, error) {
	// 检测用户是否存在
	user, err := s.UserRepo.GetByFields(map[string]any{"uuid": req.Uid})
	if err != nil {
		log.Printf("[UserService] UpdateUser error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch user data: %v", err)
	}
	if user == nil {
		log.Printf("[UserService] UpdateUser error: user with UUID %s not found", req.Uid)
		return nil, status.Errorf(codes.NotFound, "user with this UUID does not exist")
	}

	// 开启事务
	tx, err := s.UserRepo.BeginTransaction(ctx)
	if err != nil {
		log.Printf("[UserService] UpdateUser error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to begin transaction: %v", err)
	}
	defer func() {
		if err := recover(); err != nil {
			// 捕获panic并回滚
			s.UserRepo.RollbackTransaction(tx)
			log.Printf("[UserService] UpdateUser panic recover: %v", err)
		}
	}()

	updateUser := userPool.Get().(*user_model.User) // 从对象池获取对象
	*updateUser = user_model.User{}					// 清空对象
	defer userPool.Put(updateUser)

	// 获取需要更新的字段
	var fields []string
	if req.Uname != "" {
		fields = append(fields, "uname")
		updateUser.Uname = req.Uname
	}
	if req.Password != "" {
		fields = append(fields, "pwd")
		// 对用户密码加密
		hashedPwd, hashErr := user_utils.HashPWD(req.Password)
		if hashErr != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash password")
		}
		updateUser.Pwd = hashedPwd
	}
	if req.Email != "" {
		fields = append(fields, "email")
		updateUser.Email = req.Email
	}
	if req.Avatar != "" {
		fields = append(fields, "avatar")
		updateUser.Avatar = req.Avatar
	}
	// 更新用户信息
	if err := s.UserRepo.UpdateByUid(req.Uid, updateUser, fields); err != nil {
		log.Printf("[UserService] UpdateUser error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update user data: %v", err)
	}
	// 提交事务
	if err := s.UserRepo.CommitTransaction(tx); err != nil {
		log.Printf("[UserService] UpdateUser error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
	}

	return &user_pb.UpdateUserResponse{
		Success: true,
	}, nil
}

// 逻辑删除用户
// @Param ctx context.Context: context
// @Param req *user_proto.DeleteUserRequest: delete user request
// @Return *user_proto.DeleteUserResponse: delete user response
// @Return error: error
func (s *UserService) DeleteUser(ctx context.Context, req *user_pb.DeleteUserRequest) (*user_pb.DeleteUserResponse, error) {
	// 检测用户是否存在
	user, err := s.UserRepo.GetByFields(map[string]any{"uuid": req.Uid})
	if err != nil {
		log.Printf("[UserService] DeleteUser error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch user data: %v", err)
	}
	if user == nil {
		log.Printf("[UserService] DeleteUser error: user with UUID %s not found", req.Uid)
		return nil, status.Errorf(codes.NotFound, "user with this UUID does not exist")
	}

	// 开启事务
	tx, err := s.UserRepo.BeginTransaction(ctx)
	if err != nil {
		log.Printf("[UserService] DeleteUser error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to begin transaction: %v", err)
	}
	defer func() {
		if err := recover(); err != nil {
			// 捕获panic并回滚
			s.UserRepo.RollbackTransaction(tx)
			log.Printf("[UserService] DeleteUser panic recover: %v", err)
		}
	}()

	// 逻辑删除用户
	if err := s.UserRepo.DeleteByUid(req.Uid); err != nil {
		log.Printf("[UserService] DeleteUser error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to delete user data: %v", err)
	}

	// 提交事务	
	if err := s.UserRepo.CommitTransaction(tx); err != nil {
		log.Printf("[UserService] DeleteUser error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
	}

	return &user_pb.DeleteUserResponse{
		Success: true,
	}, nil
}

// 刷新token
// @Param ctx context.Context: context
// @Param req *user_proto.RefreshTokenRequest: refresh token request
// @Return *user_proto.RefreshTokenResponse: refresh token response
// @Return error: error
func (s *UserService) RefreshToken(ctx context.Context, req *user_pb.RefreshTokenRequest) (*user_pb.RefreshTokenResponse, error) {
	// 验证刷新令牌
	claims, err := user_utils.ParseToken(req.Token, &s.Cfg.JWT)
	if err != nil {
		log.Printf("[UserService] RefreshToken error: failed to parse token: %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// 验证用户ID是否匹配
	if claims.UUID != req.Uid {
		log.Printf("[UserService] RefreshToken error: token UUID does not match request UID")
		return nil, status.Errorf(codes.PermissionDenied, "token does not belong to the user")
	}

	// 检查用户是否存在
	user, err := s.UserRepo.GetByFields(map[string]any{"uuid": req.Uid})
	if err != nil {
		log.Printf("[UserService] RefreshToken error: failed to get user: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch user data: %v", err)
	}
	if user == nil {
		log.Printf("[UserService] RefreshToken error: user not found")
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	// 生成新的访问令牌
	accessToken, err := user_utils.GenerateAccessToken(&s.Cfg.JWT, user.UUID)
	if err != nil {
		log.Printf("[UserService] RefreshToken error: failed to generate access token: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to generate access token: %v", err)
	}

	// 生成新的刷新令牌
	refreshToken, err := user_utils.GenerateRefreshToken(&s.Cfg.JWT, user.UUID)
	if err != nil {
		log.Printf("[UserService] RefreshToken error: failed to generate refresh token: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to generate refresh token: %v", err)
	}

	// 更新Redis中的刷新令牌
	err = s.UserRedisSct.SetWithTTL(
		ctx, 
		user.UUID,          // Key: 用户UUID
		refreshToken,       // Value: 新生成的RefreshToken
		time.Duration(s.Cfg.JWT.Refresh_token_ttl)*time.Second, // TTL
	)
	if err != nil {
		log.Printf("[UserService] RefreshToken error: failed to update Redis: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to refresh token")
	}
	
	return &user_pb.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
