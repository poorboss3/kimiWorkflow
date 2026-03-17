<template>
  <div class="proxy-config-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <h2>代理配置管理</h2>
      <el-button type="primary" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        添加代理配置
      </el-button>
    </div>
    
    <!-- 说明卡片 -->
    <el-alert
      title="代理配置说明"
      type="info"
      :closable="false"
      show-icon
      class="info-alert"
    >
      <p>代理配置允许一个用户（代理人）代替另一个用户（被代理人）提交流程。</p>
      <p>与委托不同，代理是以被代理人的名义发起流程，而非审批过程中的委托。</p>
    </el-alert>
    
    <!-- 数据表格 -->
    <el-card>
      <el-table 
        :data="proxyList" 
        v-loading="loading"
        stripe
        border
      >
        <el-table-column prop="principalName" label="被代理人" min-width="150">
          <template #default="{ row }">
            <div class="user-cell">
              <span class="user-avatar">{{ getUserAvatar(row.principalId) }}</span>
              <span>{{ getUserName(row.principalId) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="agentName" label="代理人" min-width="150">
          <template #default="{ row }">
            <div class="user-cell">
              <span class="user-avatar">{{ getUserAvatar(row.agentId) }}</span>
              <span>{{ getUserName(row.agentId) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="有效时间" min-width="240">
          <template #default="{ row }">
            <div class="time-range">
              <div>
                <el-tag size="small" type="info">开始</el-tag>
                {{ formatTime(row.validFrom) }}
              </div>
              <div v-if="row.validTo">
                <el-tag size="small" type="info">结束</el-tag>
                {{ formatTime(row.validTo) }}
              </div>
              <div v-else>
                <el-tag size="small" type="success">永久有效</el-tag>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.isActive ? 'success' : 'info'">
              {{ row.isActive ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button 
              type="warning" 
              link 
              @click="handleToggleStatus(row)"
            >
              {{ row.isActive ? '禁用' : '启用' }}
            </el-button>
            <el-button type="danger" link @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    
    <!-- 添加/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑代理配置' : '添加代理配置'"
      width="500px"
    >
      <el-form :model="form" label-width="100px" ref="formRef">
        <el-form-item label="被代理人" prop="principalId" required>
          <el-select 
            v-model="form.principalId" 
            placeholder="选择被代理人"
            style="width: 100%"
            :disabled="isEdit"
          >
            <el-option
              v-for="user in TEST_USERS"
              :key="user.id"
              :label="user.name"
              :value="user.id"
            >
              <span>{{ user.avatar }} {{ user.name }}</span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="代理人" prop="agentId" required>
          <el-select 
            v-model="form.agentId" 
            placeholder="选择代理人"
            style="width: 100%"
            :disabled="isEdit"
          >
            <el-option
              v-for="user in availableAgents"
              :key="user.id"
              :label="user.name"
              :value="user.id"
            >
              <span>{{ user.avatar }} {{ user.name }}</span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="开始时间" prop="validFrom" required>
          <el-date-picker
            v-model="form.validFrom"
            type="datetime"
            placeholder="选择开始时间"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="结束时间">
          <el-date-picker
            v-model="form.validTo"
            type="datetime"
            placeholder="留空表示永久有效"
            style="width: 100%"
            clearable
          />
        </el-form-item>
        <el-form-item label="允许流程">
          <el-select
            v-model="form.allowedProcessTypes"
            multiple
            filterable
            allow-create
            default-first-option
            placeholder="留空表示允许所有流程"
            style="width: 100%"
          >
            <el-option label="请假流程" value="leave" />
            <el-option label="报销流程" value="expense" />
            <el-option label="采购流程" value="purchase" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          {{ isEdit ? '保存' : '添加' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore, TEST_USERS } from '@/stores/user'
import dayjs from 'dayjs'

// 注意：实际项目中应该调用后端 API
// 这里使用内存存储模拟
const userStore = useUserStore()

const loading = ref(false)
const proxyList = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref(null)

const form = reactive({
  id: '',
  principalId: '',
  agentId: '',
  validFrom: new Date(),
  validTo: null,
  allowedProcessTypes: [],
  isActive: true
})

// 可用代理人（排除被代理人自己）
const availableAgents = computed(() => {
  return TEST_USERS.filter(u => u.id !== form.principalId)
})

function getUserName(userId) {
  const user = TEST_USERS.find(u => u.id === userId)
  return user?.name || userId
}

function getUserAvatar(userId) {
  const user = TEST_USERS.find(u => u.id === userId)
  return user?.avatar || '👤'
}

function formatTime(time) {
  if (!time) return '-'
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

function loadData() {
  // 模拟从后端加载数据
  // 实际项目中应该调用 API
  const stored = localStorage.getItem('proxyConfigs')
  if (stored) {
    proxyList.value = JSON.parse(stored)
  }
}

function saveData() {
  localStorage.setItem('proxyConfigs', JSON.stringify(proxyList.value))
}

function handleCreate() {
  isEdit.value = false
  form.id = ''
  form.principalId = ''
  form.agentId = ''
  form.validFrom = new Date()
  form.validTo = null
  form.allowedProcessTypes = []
  form.isActive = true
  dialogVisible.value = true
}

function handleEdit(row) {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    principalId: row.principalId,
    agentId: row.agentId,
    validFrom: new Date(row.validFrom),
    validTo: row.validTo ? new Date(row.validTo) : null,
    allowedProcessTypes: row.allowedProcessTypes || [],
    isActive: row.isActive
  })
  dialogVisible.value = true
}

function handleSubmit() {
  if (!form.principalId) {
    ElMessage.warning('请选择被代理人')
    return
  }
  if (!form.agentId) {
    ElMessage.warning('请选择代理人')
    return
  }
  if (!form.validFrom) {
    ElMessage.warning('请选择开始时间')
    return
  }
  
  submitting.value = true
  
  setTimeout(() => {
    const data = {
      id: form.id || Date.now().toString(),
      principalId: form.principalId,
      agentId: form.agentId,
      principalName: getUserName(form.principalId),
      agentName: getUserName(form.agentId),
      validFrom: form.validFrom.toISOString(),
      validTo: form.validTo?.toISOString() || null,
      allowedProcessTypes: form.allowedProcessTypes,
      isActive: form.isActive
    }
    
    if (isEdit.value) {
      const index = proxyList.value.findIndex(p => p.id === form.id)
      if (index > -1) {
        proxyList.value[index] = data
      }
      ElMessage.success('更新成功')
    } else {
      proxyList.value.push(data)
      ElMessage.success('添加成功')
    }
    
    saveData()
    dialogVisible.value = false
    submitting.value = false
  }, 500)
}

async function handleToggleStatus(row) {
  const action = row.isActive ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(
      `确定要${action}该代理配置吗？`,
      `确认${action}`,
      { type: 'warning' }
    )
    row.isActive = !row.isActive
    saveData()
    ElMessage.success(`${action}成功`)
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Toggle failed:', error)
    }
  }
}

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(
      '确定要删除该代理配置吗？',
      '确认删除',
      { type: 'danger' }
    )
    const index = proxyList.value.findIndex(p => p.id === row.id)
    if (index > -1) {
      proxyList.value.splice(index, 1)
      saveData()
      ElMessage.success('删除成功')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete failed:', error)
    }
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.proxy-config-page {
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

.info-alert {
  margin-bottom: 0;
}

.info-alert p {
  margin: 4px 0;
  font-size: 13px;
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-avatar {
  font-size: 20px;
}

.time-range {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 13px;
}
</style>
