package domain

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email    string `gorm:"uniqueIndex;size:255"`
	Password string // Hashed password
	RoleCode string `gorm:"size:50"`
}
