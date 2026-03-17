# Workflow Engine UI

Workflow Engine 的前端测试平台，基于 Vue 3 + Element Plus 构建。

## 功能特性

- 📋 **流程定义管理** - 创建、编辑、激活、归档流程模板
- 📁 **流程实例管理** - 提交、撤回、查看流程实例
- ✅ **任务中心** - 待办/已办任务列表，支持多种审批操作
- 🔀 **用户切换** - 支持切换不同用户模拟多角色操作
- ⚙️ **代理配置** - 配置代提交权限
- 📝 **委托配置** - 配置审批委托关系

## 技术栈

- Vue 3.4+
- Element Plus 2.6+
- Vue Router 4
- Pinia
- Axios
- Vite

## 快速开始

### 安装依赖

```bash
cd workflow-engine-ui
npm install
```

### 启动开发服务器

```bash
npm run dev
```

访问 http://localhost:3000

### 构建生产版本

```bash
npm run build
```

## 预定义测试用户

| 用户ID | 姓名 | 角色 |
|--------|------|------|
| user_001 | 张三 | 员工 |
| user_002 | 李四 | 经理 |
| user_003 | 王五 | 总监 |
| user_004 | 赵六 | HR |
| user_005 | 管理员 | 管理员 |

点击右上角用户头像可切换不同用户。

## API 代理配置

开发服务器已配置代理，所有 `/api` 请求会转发到 `http://localhost:8080`（后端服务地址）。

如需修改，请编辑 `vite.config.js`：

```javascript
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true
    }
  }
}
```

## 页面说明

### 首页
- 展示系统统计信息
- 快捷操作入口
- 使用说明

### 流程定义
- 创建新的流程定义
- 配置审批步骤（支持审批/会签/通知）
- 激活/归档流程定义
- 直接提交流程

### 流程实例
- 查看我发起的流程
- 提交流流程（支持代提交）
- 撤回流程
- 标记加急
- 查看审批历史

### 任务中心
- **待办任务**: 处理审批任务
  - 通过/驳回/退回
  - 加签: 在当前步骤后添加额外审批人
  - 委托: 将任务委托给他人处理
- **已办任务**: 查看已处理的任务记录

### 代理配置
- 配置 A 用户可以代替 B 用户提交流程
- 支持时间范围限制
- 支持流程类型限制

### 委托配置
- 配置审批委托关系
- A 用户的审批任务自动转给 B 用户处理
- 记录原始审批人用于审计

## 后端 API 接口

前端对接的后端 API 接口定义在 `workflow-engine` 项目中：

- `GET /api/v1/processes/definitions` - 流程定义列表
- `POST /api/v1/processes/definitions` - 创建流程定义
- `POST /api/v1/processes/instances` - 提交流程
- `GET /api/v1/tasks/pending` - 待办任务
- `POST /api/v1/tasks/:id/action` - 处理任务

完整 API 文档请参考后端 Swagger 文档：`http://localhost:8080/swagger/index.html`

## 目录结构

```
src/
├── api/              # API 接口封装
├── components/       # 公共组件
├── router/           # 路由配置
├── stores/           # Pinia 状态管理
├── utils/            # 工具函数
├── views/            # 页面视图
├── App.vue           # 根组件
└── main.js           # 入口文件
```

## 开发注意事项

1. **用户认证**: 系统通过 `X-User-ID` Header 识别当前用户，由前端自动注入
2. **代理/委托**: 代理配置和委托配置目前使用 localStorage 存储，实际项目应接入后端 API
3. **数据持久化**: 后端使用内存存储，重启后数据会丢失
