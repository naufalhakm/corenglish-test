package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username  string    `json:"username" gorm:"size:100;uniqueIndex;not null" validate:"required,min=3,max=100"`
	Email     string    `json:"email" gorm:"size:255;uniqueIndex;not null" validate:"required,email,max=255"`
	Password  string    `json:"-" gorm:"size:255;not null" validate:"required,min=6"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`

	// Relationship
	Tasks []Task `json:"-" gorm:"foreignKey:UserID"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
