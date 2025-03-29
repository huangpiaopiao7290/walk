// Auth: pp
// Created: 2025/03/30 01:58
// Description: users table operation

package user_repo

import (
	"log"

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
	return nil
}

// create in batches
func (u *UserRepoImpl[T]) CreateInBatches(entities []T, batchSize int) error {

	if err := u.DB.CreateInBatches(entities, batchSize).Error; err != nil {
		log.Println("[UserRepoImpl] CreateInBatches error: ", err)
		return err
	}
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
func (u *UserRepoImpl[T]) GetByFields(fields map[string]interface{}) (*T, error) {

	var entity T
	if err := u.DB.Where(fields).First(&entity).Error; err != nil {
		log.Println("[UserRepoImpl] GetByFields error: ", err)
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
