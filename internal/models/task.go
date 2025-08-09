package models

import (
	"go-corenglish/internal/enum"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Task struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title       string          `json:"title" gorm:"size:255;not null" validate:"required,max=255"`
	Description *string         `json:"description" gorm:"type:text"`
	Status      enum.TaskStatus `json:"status" gorm:"type:varchar(20);not null;default:'TO_DO'" validate:"required,oneof=TO_DO IN_PROGRESS DONE"`
	UserID      uuid.UUID       `json:"user_id" gorm:"type:uuid;not null"`
	CreatedAt   time.Time       `json:"created_at" gorm:"not null"`
	UpdatedAt   time.Time       `json:"updated_at" gorm:"not null"`

	User User `json:"-" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
