<template>
  <div class="home-page">
    <!-- 欢迎区域 -->
    <el-card class="welcome-card">
      <div class="welcome-content">
        <div class="welcome-text">
          <h1>欢迎使用 Workflow Engine</h1>
          <p>BPM 工作流引擎测试平台 - Vue 3 + Element Plus</p>
        </div>
        <div class="current-user-info">
          <el-avatar :size="64" class="user-avatar">
            {{ currentUser?.avatar }}
          </el-avatar>
          <div class="user-detail">
            <h3>{{ currentUser?.name }}</h3>
            <p>{{ getRoleLabel(currentUser?.role) }}</p>
          </div>
        </div>
      </div>
    </el-card>
    
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stat-row">
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon" style="background: #ecf5ff; color: #409eff;">
            <el-icon :size="24"><Document /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.definitions }}</div>
            <div class="stat-label">流程定义</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon" style="background: #f0f9eb; color: #67c23a;">
            <el-icon :size="24"><List /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.instances }}</div>
            <div class="stat-label">我的流程</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon" style="background: #fdf6ec; color: #e6a23c;">
            <el-icon :size="24"><Bell /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.pendingTasks }}</div>
            <div class="stat-label">待办任务</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon" style="background: #f4f4f5; color: #909399;">
            <el-icon :size="24"><Finished /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.completedTasks }}</div>
            <div class="stat-label">已办任务</div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <!-- 快捷操作 -->
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>快捷操作</span>
            </div>
          </template>
          <div class="quick-actions">
            <el-button type="primary" @click="$router.push('/definitions')">
              <el-icon><Plus /></el-icon>
              创建流程定义
            </el-button>
            <el-button type="success" @click="$router.push('/instances')">
              <el-icon><Promotion /></el-icon>
              提交流程
            </el-button>
            <el-button type="warning" @click="$router.push('/tasks')">
              <el-icon><View /></el-icon>
              查看待办
            </el-button>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>系统状态</span>
            </div>
          </template>
          <div class="system-status">
            <div class="status-item">
              <span class="status-label">后端服务</span>
              <el-tag :type="apiConnected ? 'success' : 'danger'">
                {{ apiConnected ? '运行中' : '未连接' }}
              </el-tag>
            </div>
            <div class="status-item">
              <span class="status-label">存储模式</span>
              <el-tag type="info">内存存储</el-tag>
            </div>
            <div class="status-item">
              <span class="status-label">API 版本</span>
              <el-tag type="info">v1</el-tag>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <!-- 使用说明 -->
    <el-card class="help-card">
      <template #header>
        <div class="card-header">
          <span>使用说明</span>
        </div>
      </template>
      <el-collapse>
        <el-collapse-item title="1. 流程定义管理" name="1">
          <p>在"流程定义"页面可以创建、编辑、激活和归档流程模板。</p>
          <p>流程定义包含多个审批步骤，支持审批、会签、通知三种步骤类型。</p>
        </el-collapse-item>
        <el-collapse-item title="2. 流程实例管理" name="2">
          <p>在"流程实例"页面可以提交新的流程、查看我的流程列表。</p>
          <p>支持撤回尚未处理的流程，以及标记流程为加急。</p>
        </el-collapse-item>
        <el-collapse-item title="3. 任务处理" name="3">
          <p>在"任务中心"页面可以查看待办和已办任务。</p>
          <p>支持通过、驳回、退回、加签、委托等多种审批操作。</p>
        </el-collapse-item>
        <el-collapse-item title="4. 用户切换" name="4">
          <p>点击右上角用户头像可以切换不同用户，模拟不同角色的操作。</p>
          <p>预定义用户：张三(员工)、李四(经理)、王五(总监)、赵六(HR)、管理员</p>
        </el-collapse-item>
      </el-collapse>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { listDefinitions } from '@/api/process'
import { listInstances } from '@/api/process'
import { getPendingTasks, getCompletedTasks } from '@/api/task'
import axios from 'axios'

const userStore = useUserStore()
const currentUser = computed(() => userStore.currentUser)

const apiConnected = ref(false)
const stats = ref({
  definitions: 0,
  instances: 0,
  pendingTasks: 0,
  completedTasks: 0
})

const roleLabels = {
  employee: '员工',
  manager: '经理',
  director: '总监',
  hr: 'HR',
  admin: '管理员'
}

function getRoleLabel(role) {
  return roleLabels[role] || role
}

async function loadStats() {
  try {
    // 并行加载统计数据
    const [defs, insts, pending, completed] = await Promise.allSettled([
      listDefinitions({ page: 1, size: 1 }),
      listInstances({ page: 1, size: 1 }),
      getPendingTasks({ page: 1, size: 1 }),
      getCompletedTasks({ page: 1, size: 1 })
    ])
    
    if (defs.status === 'fulfilled') {
      stats.value.definitions = defs.value?.total || 0
    }
    if (insts.status === 'fulfilled') {
      stats.value.instances = insts.value?.total || 0
    }
    if (pending.status === 'fulfilled') {
      stats.value.pendingTasks = pending.value?.total || 0
    }
    if (completed.status === 'fulfilled') {
      stats.value.completedTasks = completed.value?.total || 0
    }
  } catch (error) {
    console.error('Failed to load stats:', error)
  }
}

onMounted(async () => {
  try {
    await axios.get('/api/v1/health', { timeout: 3000 })
    apiConnected.value = true
    loadStats()
  } catch {
    apiConnected.value = false
  }
})
</script>

<style scoped>
.home-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.welcome-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
}

.welcome-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.welcome-text h1 {
  margin: 0 0 8px 0;
  font-size: 24px;
}

.welcome-text p {
  margin: 0;
  opacity: 0.9;
}

.current-user-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-avatar {
  background: #fff;
  font-size: 32px;
}

.user-detail h3 {
  margin: 0 0 4px 0;
  font-size: 18px;
}

.user-detail p {
  margin: 0;
  opacity: 0.9;
  font-size: 14px;
}

.stat-row {
  margin-bottom: 0;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 10px;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  line-height: 1;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.card-header {
  font-weight: 600;
}

.quick-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.system-status {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-label {
  color: var(--el-text-color-secondary);
}

.help-card {
  margin-top: 10px;
}
</style>
