<template>
  <el-dialog
    v-model="visible"
    :title="dialogTitle"
    width="500px"
    destroy-on-close
  >
    <el-form :model="form" label-width="80px">
      <!-- 评论 -->
      <el-form-item label="审批意见">
        <el-input
          v-model="form.comment"
          type="textarea"
          :rows="3"
          placeholder="请输入审批意见（可选）"
          maxlength="500"
          show-word-limit
        />
      </el-form-item>
      
      <!-- 退回步骤选择 -->
      <el-form-item v-if="action === 'return'" label="退回至">
        <el-select v-model="form.returnToStep" placeholder="选择退回步骤">
          <el-option
            v-for="step in availableSteps"
            :key="step.stepIndex"
            :label="`步骤 ${getStepIndex(step.stepIndex)}`"
            :value="step.stepIndex"
          />
        </el-select>
      </el-form-item>
      
      <!-- 加签配置 -->
      <template v-if="action === 'countersign'">
        <el-form-item label="加签类型">
          <el-radio-group v-model="form.countersignData.type">
            <el-radio label="approval">审批</el-radio>
            <el-radio label="joint_sign">会签</el-radio>
          </el-radio-group>
        </el-form-item>
        
        <el-form-item 
          v-if="form.countersignData.type === 'joint_sign'" 
          label="会签策略"
        >
          <el-radio-group v-model="form.countersignData.jointSignPolicy">
            <el-radio label="ALL_PASS">全部通过</el-radio>
            <el-radio label="ANY_ONE">任一人通过</el-radio>
            <el-radio label="MAJORITY">多数通过</el-radio>
          </el-radio-group>
        </el-form-item>
        
        <el-form-item label="加签人员">
          <el-select
            v-model="selectedAssignees"
            multiple
            filterable
            placeholder="选择加签人员"
            style="width: 100%"
          >
            <el-option
              v-for="user in availableUsers"
              :key="user.id"
              :label="user.name"
              :value="user.id"
            >
              <span>{{ user.avatar }} {{ user.name }}</span>
            </el-option>
          </el-select>
        </el-form-item>
      </template>
      
      <!-- 委托配置 -->
      <template v-if="action === 'delegate'">
        <el-form-item label="委托人">
          <el-select
            v-model="form.delegateeId"
            placeholder="选择委托人"
            style="width: 100%"
          >
            <el-option
              v-for="user in availableUsers"
              :key="user.id"
              :label="user.name"
              :value="user.id"
            >
              <span>{{ user.avatar }} {{ user.name }}</span>
            </el-option>
          </el-select>
        </el-form-item>
      </template>
    </el-form>
    
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button 
        type="primary" 
        @click="handleSubmit"
        :loading="loading"
      >
        确认{{ actionLabels[action] }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch, defineExpose } from 'vue'
import { ElMessage } from 'element-plus'
import { processTask } from '@/api/task'
import { useUserStore, TEST_USERS } from '@/stores/user'

const props = defineProps({
  taskId: {
    type: String,
    default: ''
  },
  action: {
    type: String,
    default: ''
  },
  steps: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['success', 'update:modelValue'])

const userStore = useUserStore()
const visible = ref(false)
const loading = ref(false)

const actionLabels = {
  approve: '通过',
  reject: '驳回',
  return: '退回',
  countersign: '加签',
  delegate: '委托'
}

const dialogTitle = computed(() => {
  return `确认${actionLabels[props.action] || '操作'}`
})

// 可用用户（排除当前用户）
const availableUsers = computed(() => {
  return TEST_USERS.filter(u => u.id !== userStore.currentUser?.id)
})

// 可退回的步骤（已完成或当前步骤之前的）
const availableSteps = computed(() => {
  return props.steps.filter(s => 
    s.status === 'completed' || s.status === 'active'
  )
})

const selectedAssignees = ref([])

const form = ref({
  comment: '',
  returnToStep: null,
  countersignData: {
    type: 'approval',
    jointSignPolicy: 'ALL_PASS',
    assignees: []
  },
  delegateeId: ''
})

function getStepIndex(index) {
  return Math.floor(index) + 1
}

function open() {
  visible.value = true
  resetForm()
}

function resetForm() {
  form.value = {
    comment: '',
    returnToStep: null,
    countersignData: {
      type: 'approval',
      jointSignPolicy: 'ALL_PASS',
      assignees: []
    },
    delegateeId: ''
  }
  selectedAssignees.value = []
}

async function handleSubmit() {
  if (props.action === 'delegate' && !form.value.delegateeId) {
    ElMessage.warning('请选择委托人')
    return
  }
  
  if (props.action === 'countersign' && selectedAssignees.value.length === 0) {
    ElMessage.warning('请选择加签人员')
    return
  }
  
  loading.value = true
  
  try {
    const data = {
      action: props.action,
      comment: form.value.comment
    }
    
    if (props.action === 'return') {
      data.returnToStep = form.value.returnToStep
    }
    
    if (props.action === 'countersign') {
      data.countersignData = {
        ...form.value.countersignData,
        assignees: selectedAssignees.value.map(id => ({
          type: 'user',
          value: id
        }))
      }
    }
    
    if (props.action === 'delegate') {
      data.delegateeId = form.value.delegateeId
    }
    
    await processTask(props.taskId, data)
    ElMessage.success('操作成功')
    visible.value = false
    emit('success')
  } catch (error) {
    console.error('Task action failed:', error)
  } finally {
    loading.value = false
  }
}

watch(() => props.taskId, () => {
  if (!props.taskId) {
    visible.value = false
  }
})

defineExpose({ open })
</script>
