package http

import (
	"github.com/gin-gonic/gin"
	"kentech-project/internal/core/domain/model"
	"kentech-project/internal/core/domain/service"
	"kentech-project/pkg/logger"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
	logger      *logger.Logger
}

func NewAuthHandler(authService *service.AuthService, log *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      log,
	}
}

func (h *AuthHandler) RegisterGin(c *gin.Context) {
	h.logger.Debug("Register endpoint called")

	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body for register")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	h.logger.Infof("Registering user: username=%s", req.Username)

	user, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		if err == model.ErrUserAlreadyExists {
			h.logger.Warnf("User already exists: username=%s", req.Username)
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err == model.ErrWalletUserIDExhausted {
			h.logger.Warn("No more wallet user IDs available for registration")
			c.JSON(http.StatusConflict, gin.H{"error": "No more wallet user IDs available"})
			return
		}
		h.logger.Error("Internal error during register: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	h.logger.Info("User registered successfully")
	c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) LoginGin(c *gin.Context) {
	h.logger.Debug("Login endpoint called")

	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body for login")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	h.logger.Infof("User login attempt: username=%s", req.Username)
	response, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		if err == model.ErrInvalidCredentials {
			h.logger.Warnf("Invalid credentials for username=%s", req.Username)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		h.logger.Error("Internal error during login: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	h.logger.Info("User logged in successfully")
	c.JSON(http.StatusOK, response)
}
