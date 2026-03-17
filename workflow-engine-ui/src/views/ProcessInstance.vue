<template>
  <div class="process-instance-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <h2>流程实例管理</h2>
      <el-button type="success" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        提交流程
      </el-button>
    </div>
    
    <!-- 搜索栏 -->
    <el-card class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="状态">
          <el-select v-model="queryForm.status" placeholder="全部状态" clearable>
            <el-option label="进行中" value="running" />
            <el-option label="已完成" value="completed" />
            <el-option label="已驳回" value="rejected" />
            <el-option label="已撤回" value="withdrawn" />
          </el-select>
        </el-form-item>
        <el-form-item label="业务单号">
          <el-input 
            v-model="queryForm.businessKey" 
            placeholder="搜索业务单号"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
    
    <!-- 数据表格 -->
    <el-card>
      <el-table 
        :data="instanceList" 
        v-loading="loading"
        stripe
        border
      >
        <el-table-column prop="businessKey" label="业务单号" min-width="150" show-overflow-tooltip />
        <el-table-column label="流程名称" min-width="150">
          <template #default="{ row }">
            {{ getProcessName(row.definitionId) }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="加急" width="80" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.isUrgent" type="danger" effect="dark">加急</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="submittedBy" label="发起人" width="120" />
        <el-table-column prop="currentStepIndex" label="当前步骤" width="100" align="center">
          <template #default="{ row }">
            <el-tag type="info" effect="plain">
              第 {{ Math.floor(row.currentStepIndex) + 1 }} 步
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="提交时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleView(row)">
              详情
            </el-button>
            <el-button 
              v-if="row.status === 'running' && row.submittedBy === currentUser?.id"
              type="warning" 
              link 
              @click="handleWithdraw(row)"
            >
              撤回
            </el-button>
            <el-button 
              v-if="row.status === 'running' && row.submittedBy === currentUser?.id && !row.isUrgent"
              type="danger" 
              link 
              @click="handleMarkUrgent(row)"
            >
              加急
            </el-button>
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
    
    <!-- 提交流程对话框 -->
    <el-dialog v-model="submitVisible" title="提交流程" width="600px">
      <el-form :model="submitForm" label-width="100px" ref="submitFormRef">
        <el-form-item label="流程定义" prop="definitionId" required>
          <el-select 
            v-model="submitForm.definitionId" 
            placeholder="选择流程定义"
            style="width: 100%"
            filterable
          >
            <el-option
              v-for="def in activeDefinitions"
              :key="def.id"
              :label="def.name"
              :value="def.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="业务单号" prop="businessKey" required>
          <el-input v-model="submitForm.businessKey" placeholder="请输入业务单号" />
        </el-form-item>
        <el-form-item label="表单数据">
          <el-input
            v-model="submitForm.formDataJson"
            type="textarea"
            :rows="6"
            placeholder='{"amount": 1000, "reason": "测试申请"}'
          />
        </el-form-item>
        <el-form-item label="代提交">
          <el-select 
            v-model="submitForm.onBehalfOf" 
            placeholder="选择代提交人（可选）"
            clearable
            style="width: 100%"
          >
            <el-option
              v-for="user in TEST_USERS"
              :key="user.id"
              :label="user.name"
              :value="user.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="加急">
          <el-switch v-model="submitForm.isUrgent" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="submitVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmSubmit" :loading="submitting">
          提交
        </el-button>
      </template>
    </el-dialog>
    
    <!-- 流程详情对话框 -->
    <el-dialog v-model="detailVisible" title="流程详情" width="800px">
      <el-descriptions :column="2" border v-if="currentInstance">
        <el-descriptions-item label="业务单号" :span="2">
          {{ currentInstance.businessKey }}
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(currentInstance.status)">
            {{ getStatusLabel(currentInstance.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="加急">
          <el-tag v-if="currentInstance.isUrgent" type="danger">是</el-tag>
          <span v-else>否</span>
        </el-descriptions-item>
        <el-descriptions-item label="发起人">
          {{ currentInstance.submittedBy }}
        </el-descriptions-item>
        <el-descriptions-item label="提交时间">
          {{ formatTime(currentInstance.createdAt) }}
        </el-descriptions-item>
      </el-descriptions>
      
      <h4 style="margin: 20px 0 10px;">审批步骤</h4>
      <step-timeline :steps="currentInstance?.steps || []" />
    </el-dialog>
    

  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  listInstances, 
  getInstance, 
  submitProcess, 
  withdrawInstance,
  markUrgent 
} from '@/api/process'
import { listDefinitions } from '@/api/process'
import StepTimeline from '@/components/StepTimeline.vue'
import { useUserStore, TEST_USERS } from '@/stores/user'
import dayjs from 'dayjs'

const userStore = useUserStore()
const currentUser = computed(() => userStore.currentUser)

const loading = ref(false)
const instanceList = ref([])
const currentInstance = ref(null)

const activeDefinitions = ref([])

const submitVisible = ref(false)
const detailVisible = ref(false)

const submitting = ref(false)
const submitFormRef = ref(null)

const queryForm = reactive({
  status: '',
  businessKey: ''
})

const pagination = reactive({
  page: 1,
  size: 10,
  total: 0
})

const submitForm = reactive({
  definitionId: '',
  businessKey: '',
  formDataJson: '{}',
  onBehalfOf: '',
  isUrgent: false
})

const statusMap = {
  running: { label: '进行中', type: 'primary' },
  completed: { label: '已完成', type: 'success' },
  rejected: { label: '已驳回', type: 'danger' },
  withdrawn: { label: '已撤回', type: 'info' }
}

const actionMap = {
  submit: '提交',
  approve: '通过',
  reject: '驳回',
  return: '退回',
  withdraw: '撤回',
  countersign: '加签',
  delegate: '委托'
}

function getStatusLabel(status) {
  return statusMap[status]?.label || status
}

function getStatusType(status) {
  return statusMap[status]?.type || 'info'
}

function getActionLabel(action) {
  return actionMap[action] || action
}

function getProcessName(definitionId) {
  const def = activeDefinitions.value.find(d => d.id === definitionId)
  return def?.name || definitionId?.slice(0, 8) + '...'
}

function formatTime(time) {
  if (!time) return '-'
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

async function loadData() {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.size
    }
    if (queryForm.status) params.status = queryForm.status
    if (queryForm.businessKey) params.businessKey = queryForm.businessKey
    
    const result = await listInstances(params)
    instanceList.value = result.list || []
    pagination.total = result.total || 0
  } catch (error) {
    console.error('Failed to load instances:', error)
  } finally {
    loading.value = false
  }
}

async function loadDefinitions() {
  try {
    const result = await listDefinitions({ status: 'active', page: 1, size: 100 })
    activeDefinitions.value = result.list || []
  } catch (error) {
    console.error('Failed to load definitions:', error)
  }
}

function handleSearch() {
  pagination.page = 1
  loadData()
}

function handleReset() {
  queryForm.status = ''
  queryForm.businessKey = ''
  handleSearch()
}

function handleSizeChange(size) {
  pagination.size = size
  loadData()
}

function handlePageChange(page) {
  pagination.page = page
  loadData()
}

function handleCreate() {
  submitForm.definitionId = ''
  submitForm.businessKey = `BUS-${Date.now()}`
  submitForm.formDataJson = JSON.stringify({ amount: 1000, reason: '测试申请' }, null, 2)
  submitForm.onBehalfOf = ''
  submitForm.isUrgent = false
  submitVisible.value = true
}

async function confirmSubmit() {
  if (!submitForm.definitionId) {
    ElMessage.warning('请选择流程定义')
    return
  }
  if (!submitForm.businessKey) {
    ElMessage.warning('请输入业务单号')
    return
  }
  
  submitting.value = true
  try {
    let formData = {}
    try {
      formData = JSON.parse(submitForm.formDataJson || '{}')
    } catch {
      ElMessage.warning('表单数据格式不正确')
      return
    }
    
    await submitProcess({
      definitionId: submitForm.definitionId,
      businessKey: submitForm.businessKey,
      formData,
      onBehalfOf: submitForm.onBehalfOf || undefined,
      isUrgent: submitForm.isUrgent
    })
    
    ElMessage.success('提交成功')
    submitVisible.value = false
    loadData()
  } catch (error) {
    console.error('Submit failed:', error)
  } finally {
    submitting.value = false
  }
}

async function handleView(row) {
  try {
    const detail = await getInstance(row.id)
    currentInstance.value = detail
    detailVisible.value = true
  } catch (error) {
    console.error('Failed to get instance detail:', error)
  }
}

async function handleWithdraw(row) {
  try {
    await ElMessageBox.confirm(
      `确定要撤回流程 "${row.businessKey}" 吗？`,
      '确认撤回',
      { type: 'warning' }
    )
    await withdrawInstance(row.id, '用户主动撤回')
    ElMessage.success('撤回成功')
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Withdraw failed:', error)
    }
  }
}

async function handleMarkUrgent(row) {
  try {
    await markUrgent(row.id)
    ElMessage.success('标记加急成功')
    loadData()
  } catch (error) {
    console.error('Mark urgent failed:', error)
  }
}

onMounted(() => {
  loadData()
  loadDefinitions()
})
</script>

<style scoped>
.process-instance-page {
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

.search-card {
  margin-bottom: 0;
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}


</style>
