package model

import "gorm.io/gorm"

type Content struct {
	gorm.Model
}

func (m *Content) TableName() string {
    return "content"
}
