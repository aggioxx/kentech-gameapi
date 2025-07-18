package server

import (
	"context"
	"database/sql"
	httpSwagger "github.com/swaggo/http-swagger"
	service2 "kentech-project/internal/core/domain/service"
	"net/http"
	"strings"

	"kentech-project/internal/adapters/auth"
	httpHandlers "kentech-project/internal/adapters/http"
	"kentech-project/internal/adapters/repository/postgres"
	"kentech-project/pkg/config"
	"kentech-project/pkg/logger"

	goGinOtel "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	router *gin.Engine
	logger *logger.Logger
}

func New(cfg *config.Config, db *sql.DB, log *logger.Logger) http.Handler {
	userRepo := postgres.NewUserRepository(db)
	txRepo := postgres.NewTransactionRepository(db)

	// todo uncomment this when transaction service is implemented
	//walletClient := wallet.NewWalletClient(cfg.WalletURL)

	jwtService := auth.NewJWTService(cfg.JWTSecret)

	authService := service2.NewAuthService(userRepo, jwtService)
	playerService := service2.NewPlayerService(userRepo, txRepo)
	//txService := service2.NewTransactionService(userRepo, txRepo, walletClient, db)

	authHandler := httpHandlers.NewAuthHandler(authService)
	playerHandler := httpHandlers.NewPlayerHandler(playerService)
	//txHandler := httpHandlers.NewTransactionHandler(txService)

	router := gin.Default()

	router.Use(goGinOtel.Middleware("kentech-project"))

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"*"}
	router.Use(cors.New(corsConfig))

	router.POST("/api/auth/register", authHandler.RegisterGin)
	router.POST("/api/auth/login", authHandler.LoginGin)

	authMiddleware := NewAuthMiddleware(jwtService)
	api := router.Group("/api")
	api.Use(authMiddleware.MiddlewareGin)

	api.GET("/player/profile", playerHandler.GetProfileGin)
	api.GET("/player/balance", playerHandler.GetBalanceGin)
	api.GET("/player/transactions", playerHandler.GetTransactionHistoryGin)

	// TODO implement the following to enable transaction endpoints
	//api.POST("/transactions/deposit", txHandler.DepositGin)
	//api.POST("/transactions/withdraw", txHandler.WithdrawGin)
	//api.POST("/transactions/:id/cancel", txHandler.CancelGin)

	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	router.GET("/swagger/*any", gin.WrapH(httpSwagger.WrapHandler))

	return router
}

type AuthMiddleware struct {
	jwtService *auth.JWTService
}

func NewAuthMiddleware(jwtService *auth.JWTService) *AuthMiddleware {
	return &AuthMiddleware{jwtService: jwtService}
}

func (m *AuthMiddleware) MiddlewareGin(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
		return
	}

	claims, err := m.jwtService.ValidateToken(tokenString)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "user_id", claims.UserID)
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}
