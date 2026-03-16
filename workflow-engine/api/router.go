package api

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"workflow-engine/api/handler"
	"workflow-engine/api/middleware"
	_ "workflow-engine/docs"
	"workflow-engine/internal/service"
)

// Router API路由
type Router struct {
	engine          *gin.Engine
	processHandler  *handler.ProcessHandler
	taskHandler     *handler.TaskHandler
}

// NewRouter 创建路由
func NewRouter(processService service.ProcessService, taskService service.TaskService) *Router {
	r := &Router{
		engine:         gin.Default(),
		processHandler: handler.NewProcessHandler(processService),
		taskHandler:    handler.NewTaskHandler(taskService),
	}
	r.setup()
	return r
}

// setup 设置路由
func (r *Router) setup() {
	// 中间件
	r.engine.Use(middleware.Logger())
	r.engine.Use(middleware.Auth())

	// Swagger
	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := r.engine.Group("/api/v1")
	{
		r.processHandler.RegisterRoutes(v1)
		r.taskHandler.RegisterRoutes(v1)
	}
}

// Run 启动服务
func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}

// Engine 获取gin引擎（用于测试）
func (r *Router) Engine() *gin.Engine {
	return r.engine
}
