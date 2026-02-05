package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lesson-plan/backend/internal/config"
	"lesson-plan/backend/internal/handler"
	"lesson-plan/backend/internal/repository"
	"lesson-plan/backend/internal/service"
	"lesson-plan/backend/pkg/database"
	"lesson-plan/backend/pkg/jwt"
	"lesson-plan/backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logCfg := &logger.Config{
		Level:      cfg.Log.Level,
		Format:     cfg.Log.Format,
		Output:     cfg.Log.Output,
		FilePath:   cfg.Log.FilePath,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
	}
	if err := logger.Init(logCfg); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Starting lesson-plan server...")

	// 初始化PostgreSQL
	db, err := database.InitPostgres(&cfg.Database.Postgres)
	if err != nil {
		logger.Fatal("Failed to init postgres: " + err.Error())
	}
	_ = db
	logger.Info("PostgreSQL connected")

	// 初始化Neo4j
	neo4jDriver, err := database.InitNeo4j(&cfg.Database.Neo4j)
	if err != nil {
		logger.Fatal("Failed to init neo4j: " + err.Error())
	}
	defer neo4jDriver.Close(context.Background())
	logger.Info("Neo4j connected")

	// 初始化Redis
	redisClient, err := database.InitRedis(&cfg.Database.Redis)
	if err != nil {
		logger.Fatal("Failed to init redis: " + err.Error())
	}
	defer redisClient.Close()
	logger.Info("Redis connected")

	// 初始化JWT管理器
	jwtManager := jwt.NewManager(
		cfg.JWT.Secret,
		cfg.JWT.ExpiryDuration(),
		cfg.JWT.RefreshExpiryDuration(),
		cfg.JWT.Issuer,
	)

	// 初始化Repository
	gormDB := database.GetDB()
	userRepo := repository.NewUserRepository(gormDB)
	lessonRepo := repository.NewLessonRepository(gormDB)
	commentRepo := repository.NewCommentRepository(gormDB)
	favoriteRepo := repository.NewFavoriteRepository(gormDB)
	likeRepo := repository.NewLikeRepository(gormDB)
	generationRepo := repository.NewGenerationRepository(gormDB)
	knowledgeRepo := repository.NewKnowledgeRepository(neo4jDriver, &cfg.Database.Neo4j)
	documentRepo := repository.NewDocumentRepository(gormDB)

	// 初始化Service
	authService := service.NewAuthService(userRepo, jwtManager)
	userService := service.NewUserService(userRepo, lessonRepo, favoriteRepo)
	lessonService := service.NewLessonService(lessonRepo, favoriteRepo, likeRepo)
	commentService := service.NewCommentService(commentRepo, lessonRepo)
	favoriteService := service.NewFavoriteService(favoriteRepo, lessonRepo)
	likeService := service.NewLikeService(likeRepo, lessonRepo)
	generationService := service.NewGenerationService(generationRepo, lessonRepo, &cfg.Agent)
	knowledgeService := service.NewKnowledgeService(knowledgeRepo, &cfg.Agent)
	documentService := service.NewDocumentService(documentRepo, &cfg.Agent)

	// 初始化Handler
	authHandler := handler.NewAuthHandler(authService, userService)
	userHandler := handler.NewUserHandler(userService)
	lessonHandler := handler.NewLessonHandler(lessonService, favoriteService, likeService, commentService)
	generationHandler := handler.NewGenerationHandler(generationService, knowledgeService)
	knowledgeHandler := handler.NewKnowledgeHandler(documentService)

	// 初始化路由
	router := handler.NewRouter(authHandler, userHandler, lessonHandler, generationHandler, knowledgeHandler, jwtManager)

	// 设置Gin模式
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	engine := gin.New()
	router.Setup(engine)

	// 启动服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      engine,
		ReadTimeout:  600 * time.Second, // 10分钟
		WriteTimeout: 600 * time.Second, // 10分钟
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server: " + err.Error())
		}
	}()

	logger.Info(fmt.Sprintf("Server started on port %d", cfg.App.Port))

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown: " + err.Error())
	}

	logger.Info("Server exited properly")
}
