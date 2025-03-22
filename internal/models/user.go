// internal/models/user.go
package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// 用户角色
const (
	RoleSuperuser = "superuser" // 超级管理员
	RoleAdmin     = "admin"     // 管理员
	RoleUser      = "user"      // 普通用户
	RoleGuest     = "guest"     // 访客
)

// 用户模型，继承基础模型
type User struct {
	BaseModel        // 继承基础模型
	Username  string `gorm:"size:50;uniqueIndex;not null" json:"username"` // 用户名
	Password  string `gorm:"size:100;not null" json:"-"`                   // 密码
	Email     string `gorm:"size:100;uniqueIndex;not null" json:"email"`   // 邮箱
	Phone     string `gorm:"size:20" json:"phone"`                         // 电话
	Nickname  string `gorm:"size:50" json:"nickname"`                      // 昵称
	// Avatar 字段已移除
	RoleID    uint       `gorm:"default:3" json:"role_id"`                // 角色ID，默认为普通用户
	Role      *Role      `gorm:"foreignKey:RoleID" json:"role,omitempty"` // 角色关联
	Status    int        `gorm:"default:1" json:"status"`                 // 状态：1正常 0禁用 -1删除
	LastLogin *time.Time `json:"last_login"`                              // 最后登录时间
}

// 表名
func (User) TableName() string {
	return "star_users"
}

// 设置密码 - 使用bcrypt加密
func (u *User) SetPassword(password string) error {
	if len(password) == 0 {
		return errors.New("密码不能为空")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashedPassword)
	return nil
}

// 检查密码是否正确
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// 检查用户是否活跃
func (u *User) IsActive() bool {
	return u.Status == 1
}

// 检查用户是否是管理员
func (u *User) IsAdmin() bool {
	if u.Role != nil {
		return u.Role.Code == RoleAdmin
	}
	return false
}

// 检查用户是否有指定权限
func (u *User) HasPermission(permission string) bool {
	if u.Role != nil {
		return u.Role.HasPermission(permission)
	}
	return false
}

// 更新最后登录时间
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLogin = &now
}
