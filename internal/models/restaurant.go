package models

import (
	"gorm.io/gorm"
	"time"
)

// Restaurant represents a restaurant in the system.
type Restaurant struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Name        string         `json:"name" binding:"required"`
	Address     string         `json:"address" binding:"required"`
	UserID      uint           `json:"user_id"`
	User        User           `json:"-"` // Avoid circular dependency in JSON responses
	Menus       []Menu         `json:"menus,omitempty"`
	DailyMenuID *uint          `gorm:"default:null" json:"daily_menu_id,omitempty"`
}

// Menu represents a menu for a restaurant.
type Menu struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Name         string         `json:"name" binding:"required"`
	RestaurantID uint           `json:"restaurant_id"`
	Items        []MenuItem     `json:"items,omitempty"`
}

// MenuItem represents an item on a menu.
type MenuItem struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Name        string         `json:"name" binding:"required"`
	Description string         `json:"description"`
	Price       float64        `json:"price" binding:"required"`
	MenuID      uint           `json:"menu_id"`
}
