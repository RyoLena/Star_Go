// Package models internal/models/base_model.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// 通用状态常量
const (
	StatusActive   = 1  // 活跃
	StatusInactive = 0  // 非活跃
	StatusBanned   = -1 // 已禁用
)

// BaseModel 基础模型，包含通用字段
type BaseModel struct {
	ID        uint64         `gorm:"primarykey" json:"id"` // 主键ID
	CreatedAt time.Time      `json:"created_at"`           // 创建时间
	UpdatedAt time.Time      `json:"updated_at"`           // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`       // 删除时间（软删除）
}

// GetCreatedTime 获取创建时间的格式化字符串
func (m *BaseModel) GetCreatedTime() string {
	return m.CreatedAt.Format("2006-01-02 15:04:05")
}

// GetUpdatedTime 获取更新时间的格式化字符串
func (m *BaseModel) GetUpdatedTime() string {
	return m.UpdatedAt.Format("2006-01-02 15:04:05")
}

// IsDeleted 是否已删除
func (m *BaseModel) IsDeleted() bool {
	return !m.DeletedAt.Time.IsZero()
}
