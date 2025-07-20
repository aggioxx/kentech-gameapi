package server

import (
	"context"
	"database/sql"
	httpSwagger "github.com/swaggo/http-swagger"
	"kentech-project/internal/adapters/repository/wallet"
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
	router        *gin.Engine
	logger        *logger.Logger
	authHandler   *httpHandlers.AuthHandler
	playerHandler *httpHandlers.PlayerHandler
	txHandler     *httpHandlers.TransactionHandler
	jwtService    *auth.JWTService
}

func NewServer(cfg *config.Config, db *sql.DB, log *logger.Logger) *Server {
	userRepo := postgres.NewUserRepository(db, log)
	txRepo := postgres.NewTransactionRepository(db, log)
	walletClient := wallet.NewWalletClient(cfg.WalletURL, log, cfg.WalletAPIKey)
	jwtService := auth.NewJWTService(cfg.JWTSecret, log)

	authService := service2.NewAuthService(userRepo, jwtService, log)
	playerService := service2.NewPlayerService(userRepo, txRepo, log)
	txService := service2.NewTransactionService(userRepo, txRepo, walletClient, db, log)

	authHandler := httpHandlers.NewAuthHandler(authService, log)
	playerHandler := httpHandlers.NewPlayerHandler(playerService, log)
	txHandler := httpHandlers.NewTransactionHandler(txService, log)

	router := gin.Default()
	log.Debug("Gin router initialized")

	router.Use(goGinOtel.Middleware("kentech-project"))
	log.Info("OpenTelemetry middleware registered")

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"*"}
	router.Use(cors.New(corsConfig))
	log.Info("CORS middleware registered")

	server := &Server{
		router:        router,
		logger:        log,
		authHandler:   authHandler,
		playerHandler: playerHandler,
		txHandler:     txHandler,
		jwtService:    jwtService,
	}
	server.registerRoutes()
	log.Info("Server initialization complete")
	return server
}

func (s *Server) registerRoutes() {
	s.router.POST("/api/auth/register", func(c *gin.Context) {
		s.logger.Info("Register endpoint called")
		s.authHandler.RegisterGin(c)
	})
	s.router.POST("/api/auth/login", func(c *gin.Context) {
		s.logger.Info("Login endpoint called")
		s.authHandler.LoginGin(c)
	})

	authMiddleware := NewAuthMiddleware(s.jwtService, s.logger)
	api := s.router.Group("/api")
	api.Use(authMiddleware.MiddlewareGin)

	api.GET("/player/profile", func(c *gin.Context) {
		s.logger.Debug("GetProfile endpoint called")
		s.playerHandler.GetProfileGin(c)
	})
	api.GET("/player/balance", func(c *gin.Context) {
		s.logger.Debug("GetBalance endpoint called")
		s.playerHandler.GetBalanceGin(c)
	})
	api.GET("/player/transactions", func(c *gin.Context) {
		s.logger.Debug("GetTransactionHistory endpoint called")
		s.playerHandler.GetTransactionHistoryGin(c)
	})

	api.POST("/transactions/deposit", func(c *gin.Context) {
		s.logger.Info("Deposit endpoint called")
		s.txHandler.DepositGin(c)
	})
	api.POST("/transactions/withdraw", func(c *gin.Context) {
		s.logger.Info("Withdraw endpoint called")
		s.txHandler.WithdrawGin(c)
	})
	api.POST("/transactions/:id/cancel", func(c *gin.Context) {
		s.logger.Info("CancelTransaction endpoint called")
		s.txHandler.CancelGin(c)
	})

	s.router.GET("/health", func(c *gin.Context) {
		s.logger.Debug("Health check endpoint called")
		c.String(http.StatusOK, "OK")
	})

	s.router.GET("/swagger/*any", gin.WrapH(httpSwagger.WrapHandler))
}

func (s *Server) Handler() http.Handler {
	return s.router
}

type AuthMiddleware struct {
	jwtService *auth.JWTService
	logger     *logger.Logger
}

func NewAuthMiddleware(jwtService *auth.JWTService, log *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{jwtService: jwtService, logger: log}
}

func (m *AuthMiddleware) MiddlewareGin(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		m.logger.Warn("Missing Authorization header")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		m.logger.Warn("Bearer token missing in Authorization header")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
		return
	}

	claims, err := m.jwtService.ValidateToken(tokenString)
	if err != nil {
		m.logger.Error("Invalid token: " + err.Error())
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	m.logger.Debug("Token validated for user: " + claims.UserID.String())
	ctx := context.WithValue(c.Request.Context(), "user_id", claims.UserID)
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}
