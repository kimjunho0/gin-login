package models

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	Id              int            `json:"id" gorm:"primaryKey;autoIncrement"`
	PhoneNumber     string         `json:"phone_number" gorm:"not null;unique;type:varchar(255)"`
	Password        string         `json:"-" gorm:"not null;type:varchar(255)"`
	RefreshToken    string         `json:"-" gorm:"not null;type:varchar(255)"`
	NumPasswordFail int            `json:"-" gorm:"default:0;not null"`
	Name            string         `json:"name" gorm:"not null;type:varchar(255)"`
	CreatedAt       time.Time      `json:"-"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}
