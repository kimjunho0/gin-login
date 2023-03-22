package models

import "time"

type User struct {
	Id              int        `json:"id" gorm:"primaryKey;autoIncrement"`
	PhoneNumber     string     `json:"phone_number" gorm:"not null;unique;type:varchar(255)"`
	Password        string     `json:"-" gorm:"not null;type:varchar(255)"`
	RefreshToken    string     `json:"-" gorm:"not null;type:varchar(255)"`
	NumPasswordFail int        `json:"num_password_fail" gorm:"default:0;not null"`
	Name            string     `json:"name" gorm:"not null;type:varchar(255)"`
	CreatedAt       time.Time  `json:"-"`
	DeletedAt       *time.Time `json:"-" gorm:"index"`
}
