package services

import (
	"context"
	"encoding/json"
	"fmt"
	"go-corenglish/internal/commons/response"
	"go-corenglish/internal/enum"
	"go-corenglish/internal/models"
	"go-corenglish/internal/params"
	"go-corenglish/internal/repositories"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const cacheTTL = 60 * time.Second

type TaskService interface {
	CreateTask(userID uuid.UUID, req *params.CreateTaskRequest) (*params.TaskResponse, *response.CustomError)
	GetTask(taskID uuid.UUID, userID uuid.UUID) (*params.TaskResponse, *response.CustomError)
	GetTasks(userID uuid.UUID, status string, page, limit int) (*params.TasksResponse, *response.CustomError)
	UpdateTask(taskID uuid.UUID, userID uuid.UUID, req *params.UpdateTaskRequest) (*params.TaskResponse, *response.CustomError)
	DeleteTask(taskID uuid.UUID, userID uuid.UUID) *response.CustomError
}

type taskService struct {
	taskRepo repositories.TaskRepository
	logger   *logrus.Logger
	cache    *redis.Client
}

func NewTaskService(taskRepo repositories.TaskRepository, logger *logrus.Logger, cache *redis.Client) TaskService {
	return &taskService{
		taskRepo: taskRepo,
		logger:   logger,
		cache:    cache,
	}
}

func (s *taskService) CreateTask(userID uuid.UUID, req *params.CreateTaskRequest) (*params.TaskResponse, *response.CustomError) {
	task := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      enum.StatusToDo,
		UserID:      userID,
	}

	if err := s.taskRepo.Create(task); err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to create task")
		return nil, response.RepositoryError("failed to create task")
	}

	s.invalidateUserTasksCache(userID)

	s.logger.WithFields(logrus.Fields{
		"task_id": task.ID,
		"user_id": userID,
		"title":   task.Title,
	}).Info("Task created successfully")

	return &params.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}, nil
}

func (s *taskService) GetTask(taskID uuid.UUID, userID uuid.UUID) (*params.TaskResponse, *response.CustomError) {
	task, err := s.taskRepo.GetByID(taskID, userID)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"task_id": taskID,
			"user_id": userID,
		}).Error("Failed to get task")
		return nil, response.RepositoryError("failed to get task")
	}

	return &params.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}, nil
}

func (s *taskService) GetTasks(userID uuid.UUID, status string, page, limit int) (*params.TasksResponse, *response.CustomError) {
	if status != "" {
		if !enum.TaskStatus(status).IsValid() {
			return nil, response.BadRequestError(fmt.Sprintf("invalid status: %s", status))
		}
	}

	ctx := context.Background()
	key := s.cacheKeyTasks(userID, status, page, limit)

	if val, err := s.cache.Get(ctx, key).Result(); err == nil {
		var cached params.TasksResponse
		if json.Unmarshal([]byte(val), &cached) == nil {
			s.logger.WithField("cache_key", key).Info("Cache hit for tasks list")
			return &cached, nil
		}
	}

	tasks, total, err := s.taskRepo.GetAll(userID, status, page, limit)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to get tasks")
		return nil, response.RepositoryError("failed to get tasks")
	}

	taskResponses := make([]params.TaskResponse, len(tasks))
	for i, task := range tasks {
		taskResponses[i] = params.TaskResponse{
			ID:          task.ID,
			Title:       task.Title,
			Description: task.Description,
			Status:      task.Status,
			CreatedAt:   task.CreatedAt,
			UpdatedAt:   task.UpdatedAt,
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	response := &params.TasksResponse{
		Tasks:      taskResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	if data, err := json.Marshal(response); err == nil {
		_ = s.cache.Set(ctx, key, data, cacheTTL).Err()
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"status":      status,
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
	}).Info("Tasks retrieved successfully")

	return response, nil
}

func (s *taskService) UpdateTask(taskID uuid.UUID, userID uuid.UUID, req *params.UpdateTaskRequest) (*params.TaskResponse, *response.CustomError) {
	task, err := s.taskRepo.GetByID(taskID, userID)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"task_id": taskID,
			"user_id": userID,
		}).Error("Failed to get task for update")
		return nil, response.RepositoryError("failed to get task for update")
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = req.Description
	}
	if req.Status != nil {
		if !req.Status.IsValid() {
			return nil, response.BadRequestError(fmt.Sprintf("invalid status: %s", *req.Status))
		}
		task.Status = *req.Status
	}

	if err := s.taskRepo.Update(task); err != nil {
		s.logger.WithError(err).WithField("task_id", taskID).Error("Failed to update task")
		return nil, response.RepositoryError("failed to update task")
	}

	s.invalidateUserTasksCache(userID)

	s.logger.WithFields(logrus.Fields{
		"task_id": taskID,
		"user_id": userID,
		"title":   task.Title,
		"status":  task.Status,
	}).Info("Task updated successfully")

	return &params.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}, nil
}

func (s *taskService) DeleteTask(taskID uuid.UUID, userID uuid.UUID) *response.CustomError {
	if err := s.taskRepo.Delete(taskID, userID); err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"task_id": taskID,
			"user_id": userID,
		}).Error("Failed to delete task")
		return response.RepositoryError("failed to delete task")
	}

	s.invalidateUserTasksCache(userID)

	s.logger.WithFields(logrus.Fields{
		"task_id": taskID,
		"user_id": userID,
	}).Info("Task deleted successfully")

	return nil
}

func (s *taskService) cacheKeyTasks(userID uuid.UUID, status string, page, limit int) string {
	return fmt.Sprintf("tasks:%s:%s:%d:%d", userID.String(), status, page, limit)
}

func (s *taskService) invalidateUserTasksCache(userID uuid.UUID) {
	ctx := context.Background()
	pattern := fmt.Sprintf("tasks:%s:*", userID.String())
	iter := s.cache.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		_ = s.cache.Del(ctx, iter.Val()).Err()
	}
}
