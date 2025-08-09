package handlers

import (
	"go-corenglish/internal/commons/response"
	"go-corenglish/internal/params"
	"go-corenglish/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type TaskHandler struct {
	taskService services.TaskService
	logger      *logrus.Logger
	validator   *validator.Validate
}

func NewTaskHandler(taskService services.TaskService, logger *logrus.Logger) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
		logger:      logger,
		validator:   validator.New(),
	}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"error":   "unauthorized",
			"message": "User ID not found in context",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"error":   "unauthorized",
			"message": "Invalid user ID format",
		})
		return
	}

	var req params.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to parse create task request")
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Invalid JSON format",
		})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		details := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			details[err.Field()] = getValidationErrorMessage(err)
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Validation failed",
			"errors":  details,
		})
		return
	}

	task, custErr := h.taskService.CreateTask(userUUID, &req)
	if custErr != nil {
		c.AbortWithStatusJSON(custErr.StatusCode, custErr)
		return
	}

	resp := response.CreatedSuccessWithPayload(task)
	c.JSON(resp.StatusCode, resp)
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"error":   "unauthorized",
			"message": "User ID not found in context",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"error":   "unauthorized",
			"message": "Invalid user ID format",
		})
		return
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	tasks, custErr := h.taskService.GetTasks(userUUID, status, page, limit)
	if custErr != nil {
		c.AbortWithStatusJSON(custErr.StatusCode, custErr)
		return
	}

	resp := response.GeneralSuccessCustomMessageAndPayload("Success get tasks", tasks)
	c.JSON(http.StatusOK, resp)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"error":   "unauthorized",
			"message": "User ID not found in context",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"error":   "unauthorized",
			"message": "Invalid user ID format",
		})
		return
	}

	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "invalid_task_id",
			"message": "Invalid task ID format",
		})
		return
	}

	task, custErr := h.taskService.GetTask(taskID, userUUID)
	if custErr != nil {
		c.AbortWithStatusJSON(custErr.StatusCode, custErr)
		return
	}

	response := response.GeneralSuccessCustomMessageAndPayload("Success get task", task)
	c.JSON(http.StatusOK, response)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"error":   "unauthorized",
			"message": "User ID not found in context",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"error":   "unauthorized",
			"message": "Invalid user ID format",
		})
		return
	}

	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "invalid_task_id",
			"message": "Invalid task ID format",
		})
		return
	}

	var req params.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to parse update task request")
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "invalid_request",
			"message": "Invalid JSON format",
		})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		details := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			details[err.Field()] = getValidationErrorMessage(err)
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Validation failed",
			"errors":  details,
		})
		return
	}

	task, custErr := h.taskService.UpdateTask(taskID, userUUID, &req)
	if custErr != nil {
		c.AbortWithStatusJSON(custErr.StatusCode, custErr)
		return

	}

	response := response.GeneralSuccessCustomMessageAndPayload("Success update task", task)
	c.JSON(http.StatusOK, response)
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"error":   "unauthorized",
			"message": "User ID not found in context",
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"error":   "unauthorized",
			"message": "Invalid user ID format",
		})
		return
	}

	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"error":   "invalid_task_id",
			"message": "Invalid task ID format",
		})
		return
	}

	custErr := h.taskService.DeleteTask(taskID, userUUID)
	if custErr != nil {
		c.AbortWithStatusJSON(custErr.StatusCode, custErr)
		return
	}
	response := response.GeneralSuccessCustomMessageAndPayload("Success delete task", nil)
	c.JSON(http.StatusOK, response)
}

func getValidationErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "max":
		return "This field exceeds maximum length of " + err.Param()
	case "min":
		return "This field must be at least " + err.Param() + " characters"
	case "email":
		return "This field must be a valid email"
	case "oneof":
		return "This field must be one of: " + err.Param()
	default:
		return "This field is invalid"
	}
}
