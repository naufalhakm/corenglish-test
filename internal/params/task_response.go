package params

import (
	"go-corenglish/internal/enum"
	"time"

	"github.com/google/uuid"
)

type TaskResponse struct {
	ID          uuid.UUID       `json:"id"`
	Title       string          `json:"title"`
	Description *string         `json:"description"`
	Status      enum.TaskStatus `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type TasksResponse struct {
	Tasks      []TaskResponse `json:"tasks"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
}
