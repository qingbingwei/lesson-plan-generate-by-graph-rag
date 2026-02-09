package handler

import (
	"lesson-plan/backend/internal/middleware"
	"lesson-plan/backend/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// Router 路由管理器
type Router struct {
	authHandler       *AuthHandler
	userHandler       *UserHandler
	lessonHandler     *LessonHandler
	generationHandler *GenerationHandler
	knowledgeHandler  *KnowledgeHandler
	jwtManager        *jwt.Manager
}

// NewRouter 创建路由管理器
func NewRouter(
	authHandler *AuthHandler,
	userHandler *UserHandler,
	lessonHandler *LessonHandler,
	generationHandler *GenerationHandler,
	knowledgeHandler *KnowledgeHandler,
	jwtManager *jwt.Manager,
) *Router {
	return &Router{
		authHandler:       authHandler,
		userHandler:       userHandler,
		lessonHandler:     lessonHandler,
		generationHandler: generationHandler,
		knowledgeHandler:  knowledgeHandler,
		jwtManager:        jwtManager,
	}
}

// Setup 配置路由
func (r *Router) Setup(engine *gin.Engine) {
	// 中间件
	engine.Use(middleware.LoggerMiddleware())
	engine.Use(middleware.RecoveryMiddleware())
	engine.Use(middleware.NewCORSMiddleware([]string{"*"}, true))
	engine.Use(middleware.NewRateLimitMiddleware(100, 200))

	// 健康检查
	engine.GET("/health", HealthCheck)

	// API v1
	v1 := engine.Group("/api/v1")
	{
		// 认证路由
		auth := v1.Group("/auth")
		{
			auth.POST("/register", r.authHandler.Register)
			auth.POST("/login", r.authHandler.Login)
			auth.POST("/refresh", r.authHandler.RefreshToken)
			auth.POST("/logout", middleware.AuthMiddleware(r.jwtManager), r.authHandler.Logout)
			auth.POST("/change-password", middleware.AuthMiddleware(r.jwtManager), r.authHandler.ChangePassword)
			auth.GET("/me", middleware.AuthMiddleware(r.jwtManager), r.authHandler.GetCurrentUser)
		}

		// 用户路由
		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(r.jwtManager))
		{
			users.GET("/profile", r.userHandler.GetProfile)
			users.PUT("/profile", r.userHandler.UpdateProfile)
			users.POST("/avatar", r.userHandler.UploadAvatar)
		}

		// 教案路由
		lessons := v1.Group("/lessons")
		{
			lessons.GET("", middleware.OptionalAuthMiddleware(r.jwtManager), r.lessonHandler.List)
			lessons.GET("/search", r.lessonHandler.Search)
			lessons.GET("/:id", middleware.OptionalAuthMiddleware(r.jwtManager), r.lessonHandler.GetByID)
			lessons.GET("/:id/comments", r.lessonHandler.ListComments)
			lessons.GET("/:id/export", middleware.OptionalAuthMiddleware(r.jwtManager), r.lessonHandler.Export)

			// 需要认证的教案路由
			lessonsAuth := lessons.Group("")
			lessonsAuth.Use(middleware.AuthMiddleware(r.jwtManager))
			{
				lessonsAuth.POST("", r.lessonHandler.Create)
				lessonsAuth.PUT("/:id", r.lessonHandler.Update)
				lessonsAuth.DELETE("/:id", r.lessonHandler.Delete)
				lessonsAuth.POST("/:id/publish", r.lessonHandler.Publish)
				lessonsAuth.GET("/:id/versions", r.lessonHandler.ListVersions)
				lessonsAuth.GET("/:id/versions/:version", r.lessonHandler.GetVersion)
				lessonsAuth.POST("/:id/versions/:version/rollback", r.lessonHandler.RollbackToVersion)
				lessonsAuth.POST("/:id/favorite", r.lessonHandler.AddFavorite)
				lessonsAuth.DELETE("/:id/favorite", r.lessonHandler.RemoveFavorite)
				lessonsAuth.POST("/:id/like", r.lessonHandler.Like)
				lessonsAuth.DELETE("/:id/like", r.lessonHandler.Unlike)
				lessonsAuth.POST("/:id/comments", r.lessonHandler.CreateComment)
				lessonsAuth.DELETE("/:id/comments/:commentId", r.lessonHandler.DeleteComment)
			}
		}

		// 我的教案
		my := v1.Group("/my")
		my.Use(middleware.AuthMiddleware(r.jwtManager))
		{
			my.GET("/lessons", r.lessonHandler.MyLessons)
			my.GET("/favorites", r.lessonHandler.MyFavorites)
		}

		// 生成路由
		generate := v1.Group("/generate")
		generate.Use(middleware.AuthMiddleware(r.jwtManager))
		{
			generate.POST("", r.generationHandler.Generate)
			generate.GET("/history", r.generationHandler.ListGenerations)
			generate.GET("/history/:id", r.generationHandler.GetGeneration)
			generate.GET("/stats", r.generationHandler.GetStats)
		}

		// 知识图谱路由
		knowledge := v1.Group("/knowledge")
		{
			knowledge.GET("/search", r.generationHandler.SearchKnowledge)

			// 需要认证的知识图谱路由
			knowledgeAuth := knowledge.Group("")
			knowledgeAuth.Use(middleware.AuthMiddleware(r.jwtManager))
			{
				// 获取用户的知识图谱
				knowledgeAuth.GET("/graph", r.generationHandler.GetKnowledgeGraph)
			}

			// 文档管理 (需要认证)
			documents := knowledge.Group("/documents")
			documents.Use(middleware.AuthMiddleware(r.jwtManager))
			{
				documents.POST("", r.knowledgeHandler.UploadDocument)
				documents.GET("", r.knowledgeHandler.ListDocuments)
				documents.GET("/:id", r.knowledgeHandler.GetDocument)
				documents.DELETE("/:id", r.knowledgeHandler.DeleteDocument)
				documents.GET("/:id/status", r.knowledgeHandler.GetDocumentStatus)
			}
		}
	}
}
