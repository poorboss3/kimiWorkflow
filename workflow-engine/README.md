# 工作流引擎 (Workflow Engine)

基于 Go + Gin + 内存存储实现的工作流引擎，无需安装数据库、Redis 或消息队列即可本地运行。

## 特性

- ✅ **纯内存存储** - 无需 PostgreSQL、Redis、RabbitMQ
- ✅ **完整的审批流程** - 提交、审批、驳回、退回、加签、会签
- ✅ **代理与委托** - 支持代提交和审批委托
- ✅ **扩展点设计** - 支持 HTTP 回调扩展审批人解析和权限验证
- ✅ **Swagger 文档** - 自动生成 API 文档
- ✅ **单元测试** - 包含完整的测试用例

## 快速开始

### 1. 安装依赖

```bash
cd workflow-engine
go mod tidy
```

### 2. 运行服务

```bash
go run main.go
```

服务将启动在 `http://localhost:8080`

### 3. 查看 API 文档

打开浏览器访问：http://localhost:8080/swagger/index.html

## API 概览

### 流程定义管理

- `POST /api/v1/processes/definitions` - 创建流程定义
- `GET /api/v1/processes/definitions` - 获取流程定义列表
- `GET /api/v1/processes/definitions/:id` - 获取流程定义详情
- `POST /api/v1/processes/definitions/:id/activate` - 激活流程定义
- `POST /api/v1/processes/definitions/:id/archive` - 归档流程定义

### 流程实例管理

- `POST /api/v1/processes/instances` - 提交流程
- `GET /api/v1/processes/instances` - 获取流程实例列表
- `GET /api/v1/processes/instances/:id` - 获取流程实例详情
- `POST /api/v1/processes/instances/:id/withdraw` - 撤回流程
- `POST /api/v1/processes/instances/:id/urgent` - 标记加急
- `GET /api/v1/processes/instances/:id/history` - 获取审批历史

### 任务管理

- `GET /api/v1/tasks/pending` - 获取待办任务
- `GET /api/v1/tasks/completed` - 获取已办任务
- `GET /api/v1/tasks/statistics` - 获取任务统计
- `GET /api/v1/tasks/:id` - 获取任务详情
- `POST /api/v1/tasks/:id/action` - 处理任务（通过/驳回/退回/加签）

## 测试

```bash
go test ./test/... -v
```

## 示例流程

### 1. 创建流程定义

```bash
curl -X POST http://localhost:8080/api/v1/processes/definitions \
  -H "Content-Type: application/json" \
  -d '{
    "name": "费用报销",
    "extensionPoints": {
      "approverResolverUrl": "",
      "permissionValidatorUrl": "",
      "timeoutSeconds": 3
    }
  }'
```

### 2. 激活流程定义

```bash
curl -X POST http://localhost:8080/api/v1/processes/definitions/{definition_id}/activate \
  -H "Content-Type: application/json" \
  -d '{"version": 1}'
```

### 3. 提交流程

```bash
curl -X POST http://localhost:8080/api/v1/processes/instances \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_001" \
  -d '{
    "definitionId": "{definition_id}",
    "businessKey": "EXP-20240316-001",
    "formData": {
      "amount": 5000,
      "department": "研发部",
      "reason": "差旅费"
    }
  }'
```

### 4. 获取待办任务

```bash
curl http://localhost:8080/api/v1/tasks/pending \
  -H "X-User-ID: manager_001"
```

### 5. 审批任务

```bash
curl -X POST http://localhost:8080/api/v1/tasks/{task_id}/action \
  -H "Content-Type: application/json" \
  -H "X-User-ID: manager_001" \
  -d '{
    "action": "approve",
    "comment": "同意"
  }'
```

## 项目结构

```
workflow-engine/
├── api/
│   ├── handler/          # HTTP 处理器
│   ├── middleware/       # 中间件
│   ├── response/         # 响应封装
│   └── router.go         # 路由
├── internal/
│   ├── extension/        # 扩展点接口
│   ├── model/            # 领域模型
│   ├── pkg/              # 工具包
│   │   ├── errors/       # 错误码
│   │   ├── locker/       # 分布式锁
│   │   ├── memory/       # 内存存储
│   │   └── utils/        # 工具函数
│   ├── repository/       # 数据访问层
│   └── service/          # 业务服务层
├── test/                 # 单元测试
├── docs/                 # Swagger 文档
├── main.go               # 主程序
└── README.md
```

## 配置

配置文件位于 `configs/config.yaml`，目前使用内存存储，无需修改即可运行。

## 扩展点

系统支持两种扩展方式：

1. **HTTP 回调** - 配置 `approverResolverUrl` 和 `permissionValidatorUrl`
2. **Mock 客户端** - 默认使用，无需外部服务

## 许可证

MIT License
