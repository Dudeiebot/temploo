package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Id              string
	Name            string
	Password        string
	Email           string
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (u *User) Empty() bool {
	return u.Id == ""
}

func (u *User) EmailVerified() bool {
	if u.EmailVerifiedAt != nil {
		if u.EmailVerifiedAt.IsZero() {
			return false
		}
		return true
	}
	return false
}
