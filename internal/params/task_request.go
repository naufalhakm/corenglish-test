package params

import "go-corenglish/internal/enum"

type CreateTaskRequest struct {
	Title       string  `json:"title" validate:"required,max=255"`
	Description *string `json:"description"`
}

type UpdateTaskRequest struct {
	Title       *string          `json:"title" validate:"omitempty,max=255"`
	Description *string          `json:"description"`
	Status      *enum.TaskStatus `json:"status" validate:"omitempty,oneof=TO_DO IN_PROGRESS DONE"`
}
