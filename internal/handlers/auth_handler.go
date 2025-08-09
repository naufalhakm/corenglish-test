package handlers

import (
	"go-corenglish/internal/commons/response"
	"go-corenglish/internal/params"
	"go-corenglish/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	authService services.AuthService
	logger      *logrus.Logger
	validator   *validator.Validate
}

func NewAuthHandler(authService services.AuthService, logger *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
		validator:   validator.New(),
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req params.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to parse register request")
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

	authResponse, custErr := h.authService.Register(&req)
	if custErr != nil {
		c.AbortWithStatusJSON(custErr.StatusCode, custErr)
		return
	}

	resp := response.CreatedSuccessWithPayload(authResponse)
	c.JSON(resp.StatusCode, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req params.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to parse login request")
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

	authResponse, custErr := h.authService.Login(&req)
	if custErr != nil {
		c.AbortWithStatusJSON(custErr.StatusCode, custErr)
		return
	}

	resp := response.GeneralSuccessCustomMessageAndPayload("Success login user", authResponse)
	c.JSON(http.StatusOK, resp)
}
