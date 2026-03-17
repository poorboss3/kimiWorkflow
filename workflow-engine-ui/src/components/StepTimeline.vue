<template>
  <div class="step-timeline">
    <el-timeline>
      <el-timeline-item
        v-for="(step, index) in steps"
        :key="step.id || index"
        :type="getStepType(step)"
        :color="getStepColor(step)"
        :icon="getStepIcon(step)"
        :timestamp="formatTime(step.completedAt)"
      >
        <div class="step-card">
          <div class="step-header">
            <span class="step-title">步骤 {{ getStepIndex(step.stepIndex) }}</span>
            <el-tag :type="getStatusType(step.status)" size="small">
              {{ getStatusLabel(step.status) }}
            </el-tag>
          </div>
          
          <div class="step-content">
            <div class="step-type">
              <el-tag effect="plain" size="small">
                {{ getStepTypeLabel(step.type) }}
              </el-tag>
              <el-tag 
                v-if="step.jointSignPolicy" 
                effect="plain" 
                size="small"
                type="info"
              >
                {{ getJointSignLabel(step.jointSignPolicy) }}
              </el-tag>
            </div>
            
            <div class="assignees">
              <span class="label">审批人:</span>
              <el-space wrap>
                <el-tag 
                  v-for="assignee in getAssignees(step.assignees)" 
                  :key="assignee.value"
                  size="small"
                  :type="assignee.type === 'user' ? 'primary' : 'info'"
                >
                  {{ assignee.name || assignee.value }}
                </el-tag>
              </el-space>
            </div>
            
            <!-- 任务列表 -->
            <div v-if="step.tasks && step.tasks.length > 0" class="tasks">
              <div 
                v-for="task in step.tasks" 
                :key="task.id"
                class="task-item"
                :class="{ 'task-completed': task.status !== 'pending' }"
              >
                <el-icon :size="14">
                  <CircleCheck v-if="task.status === 'completed'" />
                  <CircleClose v-else-if="task.status === 'rejected'" />
                  <Timer v-else />
                </el-icon>
                <span class="task-assignee">{{ task.assigneeId }}</span>
                <span v-if="task.originalAssigneeId" class="delegated-from">
                  (原: {{ task.originalAssigneeId }})
                </span>
                <el-tag v-if="task.isDelegated" type="warning" size="small">委托</el-tag>
                <span v-if="task.comment" class="task-comment">"{{ task.comment }}"</span>
              </div>
            </div>
            
            <!-- 来源标记 -->
            <div v-if="step.source && step.source !== 'original'" class="step-source">
              <el-tag type="info" size="small" effect="plain">
                {{ getSourceLabel(step.source) }}
                <span v-if="step.addedByUserId">by {{ step.addedByUserId }}</span>
              </el-tag>
            </div>
          </div>
        </div>
      </el-timeline-item>
    </el-timeline>
  </div>
</template>

<script setup>
import dayjs from 'dayjs'

const props = defineProps({
  steps: {
    type: Array,
    default: () => []
  }
})

const statusMap = {
  pending: { label: '待处理', type: 'info' },
  active: { label: '进行中', type: 'warning' },
  completed: { label: '已完成', type: 'success' },
  rejected: { label: '已驳回', type: 'danger' },
  returned: { label: '已退回', type: 'danger' },
  skipped: { label: '已跳过', type: 'info' }
}

const stepTypeMap = {
  approval: '审批',
  joint_sign: '会签',
  notify: '通知'
}

const sourceMap = {
  original: '原始',
  countersign: '加签',
  dynamic_added: '动态添加'
}

const jointSignMap = {
  ALL_PASS: '全部通过',
  ANY_ONE: '任一人通过',
  MAJORITY: '多数通过'
}

function getStepIndex(index) {
  return Math.floor(index) + 1
}

function getStatusLabel(status) {
  return statusMap[status]?.label || status
}

function getStatusType(status) {
  return statusMap[status]?.type || 'info'
}

function getStepTypeLabel(type) {
  return stepTypeMap[type] || type
}

function getSourceLabel(source) {
  return sourceMap[source] || source
}

function getJointSignLabel(policy) {
  return jointSignMap[policy] || policy
}

function getStepType(step) {
  if (step.status === 'completed') return 'success'
  if (step.status === 'active') return 'primary'
  if (step.status === 'rejected' || step.status === 'returned') return 'danger'
  return 'info'
}

function getStepColor(step) {
  const colors = {
    completed: '#67C23A',
    active: '#409EFF',
    rejected: '#F56C6C',
    returned: '#E6A23C',
    pending: '#909399'
  }
  return colors[step.status]
}

function getStepIcon(step) {
  if (step.status === 'completed') return 'CircleCheck'
  if (step.status === 'active') return 'Loading'
  if (step.status === 'rejected') return 'CircleClose'
  return null
}

function getAssignees(assignees) {
  if (!assignees) return []
  if (typeof assignees === 'string') {
    try {
      return JSON.parse(assignees)
    } catch {
      return []
    }
  }
  return assignees
}

function formatTime(time) {
  if (!time) return ''
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}
</script>

<style scoped>
.step-timeline {
  padding: 16px;
}

.step-card {
  background: var(--el-fill-color-light);
  padding: 12px;
  border-radius: 8px;
  margin-bottom: 8px;
}

.step-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.step-title {
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.step-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.step-type {
  display: flex;
  gap: 8px;
}

.assignees {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.assignees .label {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.tasks {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 8px;
  background: var(--el-bg-color);
  border-radius: 4px;
}

.task-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--el-text-color-regular);
}

.task-item.task-completed {
  color: var(--el-text-color-secondary);
}

.task-assignee {
  font-weight: 500;
}

.delegated-from {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.task-comment {
  color: var(--el-text-color-secondary);
  font-style: italic;
}

.step-source {
  margin-top: 4px;
}
</style>
