// Auth: pp
// Created:
// Description:

package user_service

import (
	"context"
	"log"
	// "fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	user_model "walk/internal/user/model"
	user_repo "walk/internal/user/repo"
	utils "walk/pkg/utils"
	user_pb "walk/rpc/user"
)


type UserService struct {
	UserRepo user_repo.UserRepo[user_model.User]
	user_pb.UnimplementedUserServiceServer

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
	
	// 检测注册邮箱是否存在
	existingUser, getErr := s.UserRepo.GetByFields(map[string]any{"email": req.Email})
	log.Printf("existingUser: %v, getErr: %v", existingUser, getErr)
	if getErr != nil {
		log.Printf("[UserService] Register error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to check email existence: %v", getErr)
	}
	if existingUser != nil {
		return nil, status.Errorf(codes.AlreadyExists, "email already exists")
	}

	// 对用户密码加密
	hashedPwd, hashErr := utils.HashPWD(req.Password)
	if hashErr != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password")
	}	  
	// 定义用户
	newUser := user_model.User{    
		Uname: 		req.Uname,
		Email: 		req.Email,
		Pwd:   		hashedPwd,
		RegisterIP:	req.IpAddress, 
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
		Uid:        newUser.UUID,
		Uname:       newUser.Uname,
		Email:       newUser.Email,
	}, nil
}

// func (s *UserService) Login(ctx context.Context, req *user_pb.LoginRequest) (*user_pb.LoginResponse, error) {

// }

// func (s *UserService) GetUserInfo(ctx context.Context, req *user_pb.GetUserInfoRequest) (*user_pb.GetUserInfoResponse, error) {

// }

