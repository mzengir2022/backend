package models

import (
	"time"
	"gorm.io/gorm"
)

type User struct {
	ID                           uint      `gorm:"primarykey" json:"id"`
	CreatedAt                    time.Time `json:"created_at"`
	UpdatedAt                    time.Time `json:"updated_at"`
	DeletedAt                    gorm.DeletedAt `gorm:"index" json:"-"`
	PhoneNumber                  string    `gorm:"uniqueIndex;not null" json:"phone_number"`
	Email                        string    `gorm:"uniqueIndex;not null" json:"email"`
	Password                     string    `gorm:"not null" json:"-"`
	Role                         string    `gorm:"default:'user'" json:"role"`
	VerificationCode             string    `json:"-"`
	VerificationCodeExpiresAt    time.Time `json:"-"`
	EmailVerificationCode        string    `json:"-"`
	EmailVerificationCodeExpiresAt time.Time `json:"-"`
}
