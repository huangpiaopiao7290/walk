// Auth: pp
// Created: 2025-03-21 0:17
// Description: the defination of user model related to database

package user_model

import (
	utils "walk/shared/common/utils"

	"gorm.io/gorm"
)

// User represents the `users` table in the database
type User struct {
	gorm.Model
	UUID        string `gorm:"type:varchar(36);comment:用户唯一标识符" json:"uuid"`
	Uname       string `gorm:"type:varchar(50);not null;comment:用户名" json:"uname"`
	Email       string `gorm:"type:varchar(100);unique;not null;comment:邮箱" json:"email"`
	Pwd         string `gorm:"type:varchar(255);not null;comment:密码（加密存储）" json:"pwd"`
	PhoneNumber string `gorm:"type:varchar(20);default:null;comment:手机号" json:"phone_number,omitempty"` // 可为空(后期需要调整)
	Avatar      string `gorm:"type:varchar(255);default:null;comment:头像URL" json:"avatar,omitempty"`    // 可为空
	RegisterIP  string `gorm:"type:varchar(45);not null;comment:注册IP地址" json:"register_ip"`
	LastLoginIP string `gorm:"type:varchar(45);default:null;comment:最后登录IP地址" json:"last_login_ip,omitempty"` // 可为空
}

// 在创建用户前生成uuid
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	// 生成UUID
	uuid, err := utils.GenerateUUID()
	if err != nil {
		return err
	}
	u.UUID = uuid
	return nil
}
