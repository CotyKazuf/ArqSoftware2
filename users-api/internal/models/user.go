package models

import "time"

const (
	RoleNormal = "normal"
	RoleAdmin  = "admin"
)

// User represents a persisted user in MySQL.
type User struct {
	ID           uint      `gorm:"primaryKey"`
	Name         string    `gorm:"size:255;not null"`
	Email        string    `gorm:"size:255;not null;uniqueIndex"`
	PasswordHash string    `gorm:"size:255;not null"`
	Role         string    `gorm:"size:50;not null;default:normal"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}
