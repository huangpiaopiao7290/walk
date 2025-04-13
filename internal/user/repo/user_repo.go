// Auth: pp
// Created: 2025/03/30 01:58
// Description: users table operation

package user_repo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"
)

// 定义用户服务通用仓库接口
type UserRepo[T any] interface {
	Create(entity *T) error
	CreateInBatches(entities []T, batchSize int) error
	GetByID(id uint) (*T, error)
	GetByFields(fields map[string]any) (*T, error)
	Update(entity *T) error
	Delete(id uint) error
	BeginTransaction(ctx context.Context) (*gorm.DB, error)
	CommitTransaction(tx *gorm.DB) error
	RollbackTransaction(tx *gorm.DB) error
}

// 定义用户服务通用仓库实现
type UserRepoImpl[T any] struct {
	DB *gorm.DB
}

// new UserRepoImpl
func NewUserRepo[T any](db *gorm.DB) UserRepo[T] {
	return &UserRepoImpl[T]{DB: db}
}

// create
func (u *UserRepoImpl[T]) Create(entity *T) error {

	if err := u.DB.Create(entity).Error; err != nil {
		log.Println("[UserRepoImpl] Create error: ", err)
		return err
	}
	log.Printf("[UserRepoImpl] Create success: %v", entity)
	return nil
}

// create in batches
func (u *UserRepoImpl[T]) CreateInBatches(entities []T, batchSize int) error {

	if err := u.DB.CreateInBatches(entities, batchSize).Error; err != nil {
		log.Println("[UserRepoImpl] CreateInBatches error: ", err)
		return err
	}
	log.Printf("[UserRepoImpl] CreateInBatches success: %v", entities)
	return nil
}


// get by id
func (u *UserRepoImpl[T]) GetByID(id uint) (*T, error) {

	var entity T
	if err := u.DB.Where("id = ?", id).First(&entity).Error; err != nil {
		log.Println("[UserRepoImpl] GetByID error: ", err)
		return nil, err
	}
	return &entity, nil
}

// get by fields
func (u *UserRepoImpl[T]) GetByFields(fields map[string]any) (*T, error) {

	var entity T
	err := u.DB.Where(fields).First(&entity).Error
	// 捕获record not found错误
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("[UserRepoImpl] GetByFields error: %v", err)
		return nil, nil			// 返回nil表示未找到但不报错
	}

	// 其他错误
	if err != nil {
		log.Printf("[UserRepoImpl] GetByFields error: %v", err)
		return nil, err
	}

	return &entity, nil
}

// update
func (u *UserRepoImpl[T]) Update(entity *T) error {

	if err := u.DB.Save(entity).Error; err != nil {
		log.Println("[UserRepoImpl] Update error: ", err)
		return err
	}
	return nil
}

// delete
func (u *UserRepoImpl[T]) Delete(id uint) error {

	var entity T
	if err := u.DB.Where("id = ?", id).Delete(&entity).Error; err != nil {
		log.Println("[UserRepoImpl] Delete error: ", err)
		return err
	}
	return nil
}

// 开启事务
func (u *UserRepoImpl[T]) BeginTransaction(ctx context.Context) (*gorm.DB, error) {
	tx := u.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		log.Printf("[UserRepoImpl] BeginTransaction error: %v", tx.Error)
		return nil, tx.Error
	}

	log.Printf("[UserRepoImpl] BeginTransaction success: %v", tx)

	return tx, nil
}

// 提交事务
func (u *UserRepoImpl[T]) CommitTransaction(tx *gorm.DB) error {
	if tx == nil {
		log.Printf("[UserRepoImpl] CommitTransaction error: transaction is nil")
		return fmt.Errorf("transaction is nil")
	}

	if commitErr := tx.Commit().Error; commitErr != nil {
		// 如果事务已经提交或回滚，GORM 会返回 "sql: transaction has already been committed or rolled back"
		if !strings.Contains(commitErr.Error(), "transaction has already been committed or rolled back") {
			log.Printf("[UserRepoImpl] CommitTransaction error: %v", commitErr)
			return commitErr
		}
	}

	log.Printf("[UserRepoImpl] CommitTransaction success")

	return nil
}

// 回滚事务
func (u *UserRepoImpl[T]) RollbackTransaction(tx *gorm.DB) error {
	if tx == nil {
		log.Printf("[UserRepoImpl] RollbackTransaction error: transaction is nil")
		return fmt.Errorf("transaction is nil")
	}

	if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
		// 如果事务已经提交或回滚，GORM 会返回 "sql: transaction has already been committed or rolled back"
		if !strings.Contains(rollbackErr.Error(), "transaction has already been committed or rolled back") {
			log.Printf("[UserRepoImpl] RollbackTransaction error: %v", rollbackErr)
			return rollbackErr
		}
	}

	log.Printf("[UserRepoImpl] RollbackTransaction success")

	return nil
}
