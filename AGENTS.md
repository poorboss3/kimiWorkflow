# AGENTS.md - 工作流引擎项目指南

> 本文档面向 AI 编码助手，用于快速理解本项目架构和开发规范。

---

## 项目概述

本项目是一个**BPM 工作流引擎系统**，基于 Go 语言实现，采用**纯内存存储**设计，无需安装 PostgreSQL、Redis 或 RabbitMQ 等外部依赖即可本地运行。

### 核心功能

- **流程定义管理** - 创建、激活、归档流程模板
- **流程实例管理** - 提交、撤回、审批流转
- **任务管理** - 待办/已办任务列表、任务操作
- **审批操作** - 通过、驳回、退回、加签、会签
- **代理与委托** - 代提交权限、审批委托
- **扩展点设计** - HTTP 回调支持自定义审批人解析和权限验证

### 设计理念

1. **动态步骤链** - 使用有序步骤列表替代 DAG，支持运行时插入、删除、调整顺序
2. **扩展点解耦** - 引擎通过接口与业务系统交互，零感知业务逻辑
3. **委托透明性** - 委托关系对流程本身透明，保留原始审批人记录用于审计

---

## 技术栈

| 层次 | 技术 | 版本 | 说明 |
|------|------|------|------|
| 后端语言 | Go | 1.21+ | 主开发语言 |
| Web 框架 | Gin | v1.9+ | HTTP 服务和路由 |
| 数据存储 | 内存存储 | - | 纯内存实现，无需外部数据库 |
| 缓存/锁 | 内存锁 | - | 基于 sync.Mutex 的分布式锁模拟 |
| 消息队列 | 内存队列 | - | 基于 channel 的异步通知 |
| 配置管理 | YAML | - | 配置文件位于 `configs/config.yaml` |
| 日志 | 标准库 | - | 使用 Go 标准库 log |
| 文档 | Swagger | v1.16+ | 自动生成 API 文档 |

---

## 项目结构

```
workflow-engine/
├── api/                      # API 层
│   ├── handler/              # HTTP 请求处理器
│   │   ├── process.go        # 流程相关接口（定义、实例）
│   │   └── task.go           # 任务相关接口（待办、已办）
│   ├── middleware/           # Gin 中间件
│   │   ├── auth.go           # 认证中间件（从 Header 读取 X-User-ID）
│   │   └── logger.go         # 日志中间件
│   ├── response/             # 响应封装
│   │   └── response.go       # 统一响应格式
│   └── router.go             # 路由注册
│
├── internal/                 # 内部实现
│   ├── model/                # 领域模型（DTO/实体定义）
│   │   ├── common.go         # 基础类型、状态枚举
│   │   ├── process.go        # 流程定义、流程实例、审批步骤
│   │   ├── task.go           # 任务模型
│   │   └── rule.go           # 审批规则、代理/委托配置
│   │
│   ├── repository/           # 数据访问层（内存实现）
│   │   ├── process.go        # 流程定义/实例存储
│   │   ├── task.go           # 任务存储
│   │   └── config.go         # 代理/委托配置存储
│   │
│   ├── service/              # 业务服务层
│   │   ├── process.go        # 流程服务（创建、提交、撤回）
│   │   ├── task.go           # 任务服务（审批、驳回、退回）
│   │   └── notify.go         # 通知服务（异步处理）
│   │
│   ├── extension/            # 扩展点接口
│   │   ├── client.go         # HTTP 扩展点客户端
│   │   └── types.go          # 扩展点请求/响应类型
│   │
│   └── pkg/                  # 工具包
│       ├── errors/           # 错误码定义
│       ├── locker/           # 分布式锁接口和内存实现
│       ├── memory/           # 内存存储组件（Store + Queue）
│       └── utils/            # 工具函数
│
├── test/                     # 单元测试
│   └── process_service_test.go  # 核心流程测试用例
│
├── docs/                     # Swagger 文档（自动生成）
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
│
├── configs/                  # 配置文件
│   └── config.yaml           # 应用配置（内存存储模式）
│
├── main.go                   # 程序入口
├── go.mod                    # Go 模块定义
├── go.sum                    # 依赖锁定
└── README.md                 # 项目说明
```

---

## 构建和运行

### 安装依赖

```bash
cd workflow-engine
go mod tidy
```

### 运行服务

```bash
go run main.go
```

服务启动后访问：
- API 地址: http://localhost:8080
- Swagger 文档: http://localhost:8080/swagger/index.html
- 健康检查: http://localhost:8080/health

### 构建可执行文件

```bash
go build -o workflow-engine.exe main.go
```

---

## 测试

### 运行所有测试

```bash
go test ./test/... -v
```

### 运行特定测试

```bash
go test ./test/... -v -run TestCreateDefinition
go test ./test/... -v -run TestSubmitProcess
go test ./test/... -v -run TestTaskApprove
```

### 测试覆盖范围

- `TestCreateDefinition` - 创建流程定义
- `TestActivateDefinition` - 激活流程定义
- `TestSubmitProcess` - 提交流程实例
- `TestTaskApprove` - 任务审批通过
- `TestTaskReject` - 任务驳回
- `TestProxySubmission` - 代提交功能
- `TestDelegation` - 委托审批功能

---

## 代码规范

### 命名规范

- **包名** - 全小写，如 `repository`, `service`, `handler`
- **文件名** - 蛇形命名，如 `process.go`, `task.go`
- **类型名** - PascalCase，如 `ProcessService`, `ApprovalStep`
- **接口名** - 动词+Service，如 `ProcessService`, `TaskService`
- **实现名** - 接口名+Impl，如 `ProcessServiceImpl`
- **私有方法** - camelCase，如 `checkStepCompletion`, `applyDelegation`
- **常量/枚举** - PascalCase，如 `ProcessStatusRunning`, `StepStatusPending`

### 代码组织原则

1. **依赖注入** - 通过构造函数注入依赖，避免全局变量
2. **接口隔离** - Service 层定义接口，Handler 层依赖接口而非实现
3. **错误处理** - 使用 `internal/pkg/errors` 包定义错误码，统一返回格式
4. **事务边界** - 当前内存实现无事务概念，生产环境需添加

### 关键类型定义位置

| 类型 | 位置 |
|------|------|
| 状态枚举 | `internal/model/common.go` |
| 流程模型 | `internal/model/process.go` |
| 任务模型 | `internal/model/task.go` |
| 规则/配置模型 | `internal/model/rule.go` |
| 错误码 | `internal/pkg/errors/errors.go` |

---

## API 设计规范

### 路由前缀

- `/api/v1/processes/*` - 流程相关
- `/api/v1/tasks/*` - 任务相关
- `/swagger/*` - API 文档
- `/health` - 健康检查

### 认证方式

所有 API 需要携带 Header `X-User-ID` 标识当前用户：

```
X-User-ID: user_001
```

### 响应格式

统一响应结构：

```json
{
  "code": 0,
  "message": "success",
  "data": { }
}
```

### 主要 API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/processes/definitions | 创建流程定义 |
| POST | /api/v1/processes/definitions/:id/activate | 激活流程定义 |
| POST | /api/v1/processes/instances | 提交流程 |
| GET | /api/v1/tasks/pending | 获取待办任务 |
| POST | /api/v1/tasks/:id/action | 处理任务（通过/驳回/退回） |

---

## 扩展点机制

系统支持通过 HTTP 回调与业务系统集成：

### 扩展点配置

在流程定义中配置：

```json
{
  "extensionPoints": {
    "approverResolverUrl": "https://biz-service/workflow/resolve-approvers",
    "permissionValidatorUrl": "https://biz-service/workflow/validate-permissions",
    "timeoutSeconds": 3
  }
}
```

### 默认行为

- 空 URL 时使用 `MockClient` 返回默认审批人
- MockClient 返回 `manager_001` 作为默认审批人

### 扩展点接口

1. **审批人解析器** - 根据表单数据返回默认审批步骤列表
2. **权限验证器** - 验证最终审批列表中每人的审批权限

---

## 数据模型关键概念

### 流程状态流转

```
draft（草稿） -> active（激活） -> running（运行中） -> completed（完成）
                                    |
                                    -> rejected（驳回）
                                    |
                                    -> withdrawn（撤回）
```

### 步骤状态

- `pending` - 待激活
- `active` - 进行中（已生成任务）
- `completed` - 已完成
- `rejected` - 已驳回
- `returned` - 已退回
- `skipped` - 已跳过

### 任务操作

- `approve` - 通过
- `reject` - 驳回
- `return` - 退回
- `countersign` - 加签
- `delegate` - 委托
- `notify_read` - 通知已读

---

## 安全注意事项

1. **认证缺失** - 当前仅通过 `X-User-ID` Header 识别用户，生产环境需接入真实认证
2. **数据持久化** - 当前为内存存储，重启后数据丢失，生产环境需接入 PostgreSQL
3. **并发控制** - 使用内存锁，分布式部署时需替换为 Redis 锁
4. **扩展点安全** - HTTP 回调需配置超时和重试，防止阻塞

---

## 生产环境迁移建议

当前内存实现适用于本地开发和测试，生产环境迁移步骤：

1. **数据库** - 替换内存 Repository 为 GORM + PostgreSQL
2. **缓存** - 接入 Redis 用于分布式锁和缓存
3. **消息队列** - 接入 RabbitMQ 替代内存队列
4. **扩展点** - 配置真实的业务系统回调 URL
5. **认证** - 集成 SSO/OAuth2 认证

参考 `workflow-engine-api-design.md` 和 `workflow-engine-design.md` 获取完整设计文档。

---

## 常用开发任务

### 添加新的流程操作

1. 在 `internal/model/common.go` 添加操作类型常量
2. 在 `internal/service/task.go` 实现操作逻辑
3. 在 `api/handler/task.go` 添加 HTTP 接口
4. 更新 Swagger 注释
5. 添加单元测试

### 添加新的扩展点

1. 在 `internal/extension/types.go` 定义请求/响应类型
2. 在 `internal/extension/client.go` 实现 HTTP 调用
3. 在 `internal/service/process.go` 中调用扩展点

### 添加新的存储实现

1. 实现 `repository/*_repository.go` 中定义的接口
2. 在 `main.go` 中替换内存 Repository 为新的实现

---

## 相关文档

- `workflow-engine-design.md` - 系统设计文档（BPM 设计理念）
- `workflow-engine-api-design.md` - API 详细设计文档（含完整 Go 代码）
- `workflow-engine/README.md` - 项目快速开始指南
