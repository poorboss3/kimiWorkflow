<template>
  <div class="process-definition-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <h2>流程定义管理</h2>
      <el-button type="primary" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        创建流程定义
      </el-button>
    </div>
    
    <!-- 搜索栏 -->
    <el-card class="search-card">
      <el-form :model="queryForm" inline>
        <el-form-item label="状态">
          <el-select v-model="queryForm.status" placeholder="全部状态" clearable>
            <el-option label="草稿" value="draft" />
            <el-option label="激活" value="active" />
            <el-option label="归档" value="archived" />
          </el-select>
        </el-form-item>
        <el-form-item label="名称">
          <el-input 
            v-model="queryForm.name" 
            placeholder="搜索流程名称"
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
        :data="definitionList" 
        v-loading="loading"
        stripe
        border
      >
        <el-table-column prop="name" label="流程名称" min-width="150" show-overflow-tooltip />
        <el-table-column prop="version" label="版本" width="80" align="center" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="审批步骤" width="120" align="center">
          <template #default="{ row }">
            <el-tag type="info" effect="plain">
              {{ getStepCount(row.nodeTemplates) }} 步
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleView(row)">
              查看
            </el-button>
            <el-button 
              v-if="row.status === 'draft'" 
              type="primary" 
              link 
              @click="handleEdit(row)"
            >
              编辑
            </el-button>
            <el-button 
              v-if="row.status === 'draft'" 
              type="success" 
              link 
              @click="handleActivate(row)"
            >
              激活
            </el-button>
            <el-button 
              v-if="row.status === 'active'" 
              type="warning" 
              link 
              @click="handleArchive(row)"
            >
              归档
            </el-button>
            <el-button type="primary" link @click="handleSubmitInstance(row)">
              提交
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
    
    <!-- 创建/编辑对话框 -->
    <process-definition-form
      ref="formRef"
      :definition="currentDefinition"
      @success="handleFormSuccess"
    />
    
    <!-- 查看详情对话框 -->
    <el-dialog v-model="detailVisible" title="流程定义详情" width="600px">
      <el-descriptions :column="1" border v-if="currentDefinition">
        <el-descriptions-item label="流程名称">
          {{ currentDefinition.name }}
        </el-descriptions-item>
        <el-descriptions-item label="版本">
          {{ currentDefinition.version }}
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(currentDefinition.status)">
            {{ getStatusLabel(currentDefinition.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="审批步骤">
          <step-timeline :steps="parseSteps(currentDefinition.nodeTemplates)" />
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">
          {{ formatTime(currentDefinition.createdAt) }}
        </el-descriptions-item>
        <el-descriptions-item label="更新时间">
          {{ formatTime(currentDefinition.updatedAt) }}
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
    
    <!-- 提交流程对话框 -->
    <el-dialog v-model="submitVisible" title="提交流程" width="500px">
      <el-form :model="submitForm" label-width="100px">
        <el-form-item label="业务单号" required>
          <el-input v-model="submitForm.businessKey" placeholder="请输入业务单号" />
        </el-form-item>
        <el-form-item label="表单数据">
          <el-input
            v-model="submitForm.formDataJson"
            type="textarea"
            :rows="4"
            placeholder='{"key": "value"}'
          />
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
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  listDefinitions, 
  activateDefinition, 
  archiveDefinition,
  submitProcess 
} from '@/api/process'
import ProcessDefinitionForm from '@/components/ProcessDefinitionForm.vue'
import StepTimeline from '@/components/StepTimeline.vue'
import dayjs from 'dayjs'

const loading = ref(false)
const definitionList = ref([])
const currentDefinition = ref(null)
const formRef = ref(null)
const detailVisible = ref(false)
const submitVisible = ref(false)
const submitting = ref(false)

const queryForm = reactive({
  status: '',
  name: ''
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
  isUrgent: false
})

const statusMap = {
  draft: { label: '草稿', type: 'info' },
  active: { label: '激活', type: 'success' },
  archived: { label: '归档', type: 'danger' }
}

function getStatusLabel(status) {
  return statusMap[status]?.label || status
}

function getStatusType(status) {
  return statusMap[status]?.type || 'info'
}

function getStepCount(nodeTemplates) {
  if (!nodeTemplates) return 0
  try {
    const parsed = typeof nodeTemplates === 'string' ? JSON.parse(nodeTemplates) : nodeTemplates
    return parsed.length
  } catch {
    return 0
  }
}

function parseSteps(nodeTemplates) {
  if (!nodeTemplates) return []
  try {
    const parsed = typeof nodeTemplates === 'string' ? JSON.parse(nodeTemplates) : nodeTemplates
    return parsed.map((node, index) => ({
      stepIndex: index,
      type: node.type || 'approval',
      status: 'pending',
      assignees: node.assignees || [],
      jointSignPolicy: node.jointSignPolicy
    }))
  } catch {
    return []
  }
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
    if (queryForm.name) params.name = queryForm.name
    
    const result = await listDefinitions(params)
    definitionList.value = result.list || []
    pagination.total = result.total || 0
  } catch (error) {
    console.error('Failed to load definitions:', error)
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  loadData()
}

function handleReset() {
  queryForm.status = ''
  queryForm.name = ''
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
  currentDefinition.value = null
  formRef.value?.open()
}

function handleEdit(row) {
  currentDefinition.value = row
  formRef.value?.open()
}

function handleView(row) {
  currentDefinition.value = row
  detailVisible.value = true
}

async function handleActivate(row) {
  try {
    await ElMessageBox.confirm(
      `确定要激活流程 "${row.name}" 吗？激活后将可以创建流程实例。`,
      '确认激活',
      { type: 'warning' }
    )
    await activateDefinition(row.id, row.version)
    ElMessage.success('激活成功')
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Activate failed:', error)
    }
  }
}

async function handleArchive(row) {
  try {
    await ElMessageBox.confirm(
      `确定要归档流程 "${row.name}" 吗？归档后将不能再创建流程实例。`,
      '确认归档',
      { type: 'warning' }
    )
    await archiveDefinition(row.id)
    ElMessage.success('归档成功')
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Archive failed:', error)
    }
  }
}

function handleSubmitInstance(row) {
  if (row.status !== 'active') {
    ElMessage.warning('只有激活状态的流程才能提交')
    return
  }
  submitForm.definitionId = row.id
  submitForm.businessKey = `BUS-${Date.now()}`
  submitForm.formDataJson = JSON.stringify({ amount: 1000, reason: '测试' }, null, 2)
  submitForm.isUrgent = false
  submitVisible.value = true
}

async function confirmSubmit() {
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
      isUrgent: submitForm.isUrgent
    })
    
    ElMessage.success('提交成功')
    submitVisible.value = false
  } catch (error) {
    console.error('Submit failed:', error)
  } finally {
    submitting.value = false
  }
}

function handleFormSuccess() {
  loadData()
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.process-definition-page {
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
