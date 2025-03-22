package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UUID版 基础模型

type BaseModelUUID struct {
	ID        string         `gorm:"primarykey" json:"id"` // 主键ID
	CreatedAt time.Time      `json:"created_at"`           // 创建时间
	UpdatedAt time.Time      `json:"updated_at"`           // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`       // 删除时间（软删除）
}

// 在创建对象时生成 UUID
func (m *BaseModelUUID) BeforeCreate(tx *gorm.DB) (err error) {
	m.ID = uuid.New().String()
	return nil
}

// 获取创建时间的格式化字符串
func (m *BaseModelUUID) GetCreatedTime() string {
	return m.CreatedAt.Format("2006-01-02 15:04:05")
}

// 获取更新时间的格式化字符串
func (m *BaseModelUUID) GetUpdatedTime() string {
	return m.UpdatedAt.Format("2006-01-02 15:04:05")
}

// 是否已删除
func (m *BaseModelUUID) IsDeleted() bool {
	return !m.DeletedAt.Time.IsZero()
}
