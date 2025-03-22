// Package models internal/models/role.go
package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// 权限常量
const (
	// 用户相关权限
	PermUserView   = "user:view"   // 查看用户
	PermUserCreate = "user:create" // 创建用户
	PermUserEdit   = "user:edit"   // 编辑用户
	PermUserDelete = "user:delete" // 删除用户

	// 内容相关权限
	PermContentView   = "content:view"   // 查看内容
	PermContentCreate = "content:create" // 创建内容
	PermContentEdit   = "content:edit"   // 编辑内容
	PermContentDelete = "content:delete" // 删除内容

	// 系统相关权限
	PermSystemConfig = "system:config" // 系统配置
	PermSystemLog    = "system:log"    // 系统日志
	PermSystemBackup = "system:backup" // 系统备份

	// 全部权限
	PermAll = "*" // 所有权限
)

// 根据角色编码获取预定义角色
func GetPredefinedRole(code string) *Role {
	switch code {
	case RoleAdmin:
		return AdminRole
	case RoleUser:
		return UserRole
	case RoleGuest:
		return GuestRole
	default:
		return nil
	}
}

// Permissions 权限类型 - 使用字符串切片存储权限
type Permissions []string

// Scan 实现Scanner接口，用于从数据库读取JSON
func (p *Permissions) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("类型断言为[]byte失败")
	}

	return json.Unmarshal(bytes, p)
}

// Value 实现Valuer接口，用于将权限写入数据库
func (p Permissions) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return json.Marshal(p)
}

// HasPermission 检查是否包含指定权限
func (p Permissions) HasPermission(permission string) bool {
	for _, perm := range p {
		if perm == permission || perm == "*" {
			return true
		}
	}
	return false
}

// 添加权限
func (p *Permissions) AddPermission(permission string) {
	// 检查是否已存在该权限
	for _, perm := range *p {
		if perm == permission {
			return
		}
	}
	*p = append(*p, permission)
}

// 删除权限
func (p *Permissions) RemovePermission(permission string) {
	for i, perm := range *p {
		if perm == permission {
			*p = append((*p)[:i], (*p)[i+1:]...)
			return
		}
	}
}

// 角色模型
type Role struct {
	BaseModel
	Name        string      `gorm:"size:50;uniqueIndex;not null" json:"name"` // 角色名称
	Code        string      `gorm:"size:50;uniqueIndex;not null" json:"code"` // 角色编码
	Description string      `gorm:"size:200" json:"description"`              // 角色描述
	Permissions Permissions `gorm:"type:json" json:"permissions"`             // 角色权限
}

// 表名
func (Role) TableName() string {
	return "star_roles"
}

// 检查是否有指定权限
func (r *Role) HasPermission(permission string) bool {
	return r.Permissions.HasPermission(permission)
}

// 添加权限
func (r *Role) AddPermission(permission string) {
	r.Permissions.AddPermission(permission)
}

// 删除权限
func (r *Role) RemovePermission(permission string) {
	r.Permissions.RemovePermission(permission)
}

// 预定义角色
var (
	// 管理员角色 - 拥有所有权限
	AdminRole = &Role{
		Name:        "管理员",
		Code:        RoleAdmin,
		Description: "系统管理员，拥有所有权限",
		Permissions: Permissions{"*"},
	}

	// 普通用户角色 - 拥有基本权限
	UserRole = &Role{
		Name:        "普通用户",
		Code:        RoleUser,
		Description: "普通用户，拥有基本权限",
		Permissions: Permissions{
			"user:view",
			"user:edit",
			"content:view",
			"content:create",
			"content:edit",
		},
	}

	// 访客角色 - 仅拥有查看权限
	GuestRole = &Role{
		Name:        "访客",
		Code:        RoleGuest,
		Description: "访客，仅拥有查看权限",
		Permissions: Permissions{
			"user:view",
			"content:view",
		},
	}
)


