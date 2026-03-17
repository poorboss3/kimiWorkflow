<template>
  <div class="task-list-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <h2>任务中心</h2>
      <div class="header-tabs">
        <el-radio-group v-model="activeTab" size="large">
          <el-radio-button label="pending">
            <el-icon><Bell /></el-icon>
            待办任务
            <el-badge v-if="stats.pending > 0" :value="stats.pending" class="tab-badge" />
          </el-radio-button>
          <el-radio-button label="completed">
            <el-icon><Finished /></el-icon>
            已办任务
          </el-radio-button>
        </el-radio-group>
      </div>
    </div>
    
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stat-row" v-if="activeTab === 'pending'">
      <el-col :span="8">
        <el-card class="stat-card urgent-stat" shadow="hover" @click="handleShowUrgent">
          <div class="stat-content">
            <el-icon :size="32" color="#f56c6c"><AlarmClock /></el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.urgent }}</div>
              <div class="stat-label">加急待办</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <el-icon :size="32" color="#e6a23c"><Timer /></el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.pending }}</div>
              <div class="stat-label">普通待办</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <el-icon :size="32" color="#67c23a"><CircleCheck /></el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.completed }}</div>
              <div class="stat-label">本周已办</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <!-- 任务列表 -->
    <el-card>
      <el-table 
        :data="taskList" 
        v-loading="loading"
        stripe
        border
      >
        <el-table-column type="index" width="50" align="center" />
        <el-table-column prop="businessKey" label="业务单号" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="business-key">
              <el-icon v-if="row.isUrgent" color="#f56c6c" class="urgent-icon"><AlarmClock /></el-icon>
              {{ row.businessKey }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="流程名称" min-width="130">
          <template #default="{ row }">
            {{ row.processName || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="当前步骤" width="110" align="center">
          <template #default="{ row }">
            <el-tag type="info" effect="plain" size="small">
              步骤 {{ Math.floor(row.stepIndex) + 1 }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="步骤类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStepType(row.stepType)" size="small">
              {{ getStepTypeLabel(row.stepType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="initiatorName" label="发起人" width="100" />
        <el-table-column prop="submittedAt" label="提交时间" width="170">
          <template #default="{ row }">
            {{ formatTime(row.submittedAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <template v-if="activeTab === 'pending'">
              <el-button type="success" size="small" @click="handleAction(row, 'approve')">
                通过
              </el-button>
              <el-button type="danger" size="small" @click="handleAction(row, 'reject')">
                驳回
              </el-button>
              <el-button type="warning" size="small" @click="handleAction(row, 'return')">
                退回
              </el-button>
              <el-dropdown trigger="click" @command="cmd => handleMoreAction(row, cmd)">
                <el-button type="primary" size="small" link>
                  更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="countersign">
                      <el-icon><Plus /></el-icon> 加签
                    </el-dropdown-item>
                    <el-dropdown-item command="delegate">
                      <el-icon><User /></el-icon> 委托
                    </el-dropdown-item>
                    <el-dropdown-item command="view">
                      <el-icon><View /></el-icon> 查看详情
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </template>
            <template v-else>
              <el-tag :type="getActionType(row.action)" size="small">
                {{ getActionLabel(row.action) }}
              </el-tag>
              <span v-if="row.comment" class="action-comment">
                "{{ row.comment }}"
              </span>
            </template>
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :total="pagination.total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
    
    <!-- 任务操作对话框 -->
    <task-action-dialog
      ref="actionDialogRef"
      :task-id="currentTask?.id"
      :action="currentAction"
      :steps="currentInstanceSteps"
      @success="handleActionSuccess"
    />
    
    <!-- 任务详情对话框 -->
    <el-dialog v-model="detailVisible" title="任务详情" width="700px">
      <el-descriptions :column="2" border v-if="currentTask">
        <el-descriptions-item label="业务单号" :span="2">
          {{ currentTask.businessKey }}
        </el-descriptions-item>
        <el-descriptions-item label="流程名称">
          {{ currentTask.processName }}
        </el-descriptions-item>
        <el-descriptions-item label="当前步骤">
          步骤 {{ Math.floor(currentTask.stepIndex) + 1 }}
        </el-descriptions-item>
        <el-descriptions-item label="步骤类型">
          {{ getStepTypeLabel(currentTask.stepType) }}
        </el-descriptions-item>
        <el-descriptions-item label="发起人">
          {{ currentTask.initiatorName }}
        </el-descriptions-item>
        <el-descriptions-item label="提交时间">
          {{ formatTime(currentTask.submittedAt) }}
        </el-descriptions-item>
      </el-descriptions>
      
      <h4 style="margin: 20px 0 10px;">审批流程</h4>
      <step-timeline :steps="currentInstanceSteps" />
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getPendingTasks, getCompletedTasks, getTaskStatistics, getTaskDetail } from '@/api/task'
import TaskActionDialog from '@/components/TaskActionDialog.vue'
import StepTimeline from '@/components/StepTimeline.vue'
import dayjs from 'dayjs'

const activeTab = ref('pending')
const loading = ref(false)
const taskList = ref([])
const currentTask = ref(null)
const currentAction = ref('')
const currentInstanceSteps = ref([])
const actionDialogRef = ref(null)
const detailVisible = ref(false)

const stats = reactive({
  pending: 0,
  urgent: 0,
  completed: 0
})

const pagination = reactive({
  page: 1,
  size: 10,
  total: 0
})

const stepTypeMap = {
  approval: { label: '审批', type: 'primary' },
  joint_sign: { label: '会签', type: 'warning' },
  notify: { label: '通知', type: 'info' }
}

const actionMap = {
  approve: { label: '通过', type: 'success' },
  reject: { label: '驳回', type: 'danger' },
  return: { label: '退回', type: 'warning' },
  delegate: { label: '委托', type: 'info' },
  countersign: { label: '加签', type: 'primary' },
  notify_read: { label: '已读', type: 'info' }
}

function getStepTypeLabel(type) {
  return stepTypeMap[type]?.label || type
}

function getStepType(type) {
  return stepTypeMap[type]?.type || 'info'
}

function getActionLabel(action) {
  return actionMap[action]?.label || action
}

function getActionType(action) {
  return actionMap[action]?.type || 'info'
}

function formatTime(time) {
  if (!time) return '-'
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

async function loadStats() {
  try {
    const result = await getTaskStatistics()
    stats.pending = result.pendingCount || 0
    stats.urgent = result.urgentCount || 0
    stats.completed = result.completedThisWeek || 0
  } catch (error) {
    console.error('Failed to load stats:', error)
  }
}

async function loadData() {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.size
    }
    
    let result
    if (activeTab.value === 'pending') {
      result = await getPendingTasks(params)
    } else {
      result = await getCompletedTasks(params)
    }
    
    taskList.value = result.list || []
    pagination.total = result.total || 0
  } catch (error) {
    console.error('Failed to load tasks:', error)
  } finally {
    loading.value = false
  }
}

function handleSizeChange(size) {
  pagination.size = size
  loadData()
}

function handlePageChange(page) {
  pagination.page = page
  loadData()
}

async function handleAction(row, action) {
  currentTask.value = row
  currentAction.value = action
  
  // 获取流程步骤信息
  try {
    const detail = await getTaskDetail(row.id)
    currentInstanceSteps.value = detail.instance?.steps || []
  } catch {
    currentInstanceSteps.value = []
  }
  
  actionDialogRef.value?.open()
}

async function handleMoreAction(row, command) {
  if (command === 'view') {
    currentTask.value = row
    try {
      const detail = await getTaskDetail(row.id)
      currentInstanceSteps.value = detail.instance?.steps || []
      detailVisible.value = true
    } catch (error) {
      console.error('Failed to get task detail:', error)
    }
  } else {
    handleAction(row, command)
  }
}

function handleActionSuccess() {
  loadData()
  loadStats()
}

function handleShowUrgent() {
  // 可以添加筛选逻辑
  ElMessage.info('显示加急任务')
}

watch(activeTab, () => {
  pagination.page = 1
  loadData()
})

onMounted(() => {
  loadData()
  loadStats()
})
</script>

<style scoped>
.task-list-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-header h2 {
  margin: 0;
}

.tab-badge {
  margin-left: 4px;
}

.stat-row {
  margin-bottom: 0;
}

.stat-card {
  cursor: pointer;
  transition: transform 0.2s;
}

.stat-card:hover {
  transform: translateY(-2px);
}

.urgent-stat {
  border: 1px solid #fde2e2;
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  line-height: 1;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.business-key {
  display: flex;
  align-items: center;
  gap: 6px;
}

.urgent-icon {
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}

.action-comment {
  margin-left: 8px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
</style>
