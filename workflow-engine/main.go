package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"workflow-engine/api"
	"workflow-engine/internal/extension"
	"workflow-engine/internal/model"
	"workflow-engine/internal/pkg/locker"
	"workflow-engine/internal/pkg/memory"
	"workflow-engine/internal/repository"
	"workflow-engine/internal/service"
)

// @title 工作流引擎 API
// @version 1.0
// @description 基于内存存储的工作流引擎 API 文档
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// 创建内存存储组件
	mq := memory.NewQueue()
	defer mq.Close()

	// 创建仓库
	defRepo := repository.NewMemoryProcessDefinitionRepository()
	instanceRepo := repository.NewMemoryProcessInstanceRepository()
	stepRepo := repository.NewMemoryApprovalStepRepository()
	taskRepo := repository.NewMemoryTaskRepository()
	modRepo := repository.NewMemoryApproverListModificationRepository()
	proxyRepo := repository.NewMemoryProxyConfigRepository()
	delegationRepo := repository.NewMemoryDelegationConfigRepository()

	// 创建扩展点客户端（使用模拟客户端，无需外部服务）
	extClient := extension.NewMockClient()

	// 创建锁
	lck := locker.NewMemoryLocker()

	// 创建服务
	processService := service.NewProcessService(
		defRepo, instanceRepo, stepRepo, taskRepo, modRepo,
		proxyRepo, delegationRepo, extClient, lck, mq,
	)
	taskService := service.NewTaskService(
		taskRepo, stepRepo, instanceRepo, defRepo, mq, lck,
	)
	notifyService := service.NewNotificationService(mq)

	// 启动通知服务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	notifyService.Start(ctx)

	// 初始化测试数据
	initTestData(processService)

	// 创建路由
	router := api.NewRouter(processService, taskService)

	// 启动服务器
	go func() {
		fmt.Println("=====================================")
		fmt.Println("工作流引擎已启动")
		fmt.Println("API地址: http://localhost:8080")
		fmt.Println("Swagger文档: http://localhost:8080/swagger/index.html")
		fmt.Println("=====================================")
		if err := router.Run(":8080"); err != nil {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n正在关闭服务器...")
	cancel()
	time.Sleep(time.Second)
	fmt.Println("服务器已关闭")
}

// initTestData 初始化测试数据
func initTestData(processService service.ProcessService) {
	ctx := context.Background()

	// 创建测试流程定义
	_, err := processService.CreateDefinition(ctx, &service.CreateDefinitionRequest{
		Name:          "费用报销",
		NodeTemplates: []model.NodeTemplate{},
		ExtensionPoints: model.ExtensionPointsConfig{
			ApproverResolverURL:    "",
			PermissionValidatorURL: "",
			TimeoutSeconds:         3,
		},
	})
	if err != nil {
		log.Printf("创建测试流程定义失败: %v", err)
		return
	}

	// 激活流程定义
	defs, _ := processService.ListDefinitions(ctx, nil, "费用报销", 1, 10)
	if len(defs.List) > 0 {
		def := defs.List[0]
		processService.ActivateDefinition(ctx, def.ID, 1)
		fmt.Printf("已创建测试流程定义: %s (ID: %s)\n", def.Name, def.ID)
	}
}


