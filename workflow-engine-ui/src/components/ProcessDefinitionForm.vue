<template>
  <el-dialog
    v-model="visible"
    :title="isEdit ? '编辑流程定义' : '创建流程定义'"
    width="700px"
    destroy-on-close
  >
    <el-form 
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
      class="definition-form"
    >
      <el-form-item label="流程名称" prop="name">
        <el-input 
          v-model="form.name" 
          placeholder="请输入流程名称"
          maxlength="100"
          show-word-limit
        />
      </el-form-item>
      
      <!-- 审批步骤 -->
      <el-form-item label="审批步骤">
        <div class="steps-container">
          <div 
            v-for="(step, index) in form.nodeTemplates" 
            :key="index"
            class="step-item"
          >
            <el-card shadow="never" class="step-card">
              <template #header>
                <div class="step-header">
                  <span>步骤 {{ index + 1 }}</span>
                  <el-button
                    type="danger"
                    link
                    size="small"
                    @click="removeStep(index)"
                    :disabled="form.nodeTemplates.length <= 1"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
              </template>
              
              <el-form-item label="步骤类型" :required="true">
                <el-radio-group v-model="step.type">
                  <el-radio-button label="approval">审批</el-radio-button>
                  <el-radio-button label="joint_sign">会签</el-radio-button>
                  <el-radio-button label="notify">通知</el-radio-button>
                </el-radio-group>
              </el-form-item>
              
              <el-form-item 
                v-if="step.type === 'joint_sign'" 
                label="会签策略"
                :required="true"
              >
                <el-radio-group v-model="step.jointSignPolicy">
                  <el-radio-button label="ALL_PASS">全部通过</el-radio-button>
                  <el-radio-button label="ANY_ONE">任一人通过</el-radio-button>
                  <el-radio-button label="MAJORITY">多数通过</el-radio-button>
                </el-radio-group>
              </el-form-item>
              
              <el-form-item label="审批人" :required="true">
                <el-select
                  v-model="step.assignees"
                  multiple
                  filterable
                  placeholder="选择审批人"
                  style="width: 100%"
                >
                  <el-option-group label="用户">
                    <el-option
                      v-for="user in TEST_USERS"
                      :key="user.id"
                      :label="user.name"
                      :value="JSON.stringify({ type: 'user', value: user.id, name: user.name })"
                    >
                      <span>{{ user.avatar }} {{ user.name }}</span>
                    </el-option>
                  </el-option-group>
                  <el-option-group label="角色">
                    <el-option
                      label="部门主管"
                      value='{"type":"direct_supervisor","value":"direct_supervisor","name":"部门主管"}'
                    />
                    <el-option
                      label="HR"
                      value='{"type":"role","value":"hr","name":"HR"}'
                    />
                  </el-option-group>
                </el-select>
              </el-form-item>
            </el-card>
          </div>
          
          <el-button type="primary" plain @click="addStep" class="add-step-btn">
            <el-icon><Plus /></el-icon>
            添加步骤
          </el-button>
        </div>
      </el-form-item>
      
      <!-- 扩展点配置 -->
      <el-form-item label="扩展点配置">
        <el-collapse>
          <el-collapse-item title="HTTP 扩展点" name="extension">
            <el-form-item label="审批人解析">
              <el-input 
                v-model="form.extensionPoints.approverResolverUrl" 
                placeholder="https://example.com/resolve-approvers"
              />
            </el-form-item>
            <el-form-item label="权限验证">
              <el-input 
                v-model="form.extensionPoints.permissionValidatorUrl" 
                placeholder="https://example.com/validate-permissions"
              />
            </el-form-item>
            <el-form-item label="超时时间(秒)">
              <el-input-number 
                v-model="form.extensionPoints.timeoutSeconds" 
                :min="1" 
                :max="30"
                :default-value="3"
              />
            </el-form-item>
          </el-collapse-item>
        </el-collapse>
      </el-form-item>
    </el-form>
    
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="handleSubmit" :loading="loading">
        {{ isEdit ? '保存' : '创建' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, defineExpose } from 'vue'
import { ElMessage } from 'element-plus'
import { createDefinition, updateDefinition } from '@/api/process'
import { TEST_USERS } from '@/stores/user'

const props = defineProps({
  definition: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['success'])

const visible = ref(false)
const loading = ref(false)
const formRef = ref(null)

const isEdit = computed(() => !!props.definition?.id)

const form = ref({
  name: '',
  nodeTemplates: [],
  extensionPoints: {
    approverResolverUrl: '',
    permissionValidatorUrl: '',
    timeoutSeconds: 3
  }
})

const rules = {
  name: [
    { required: true, message: '请输入流程名称', trigger: 'blur' },
    { max: 100, message: '名称长度不能超过100个字符', trigger: 'blur' }
  ]
}

function initForm() {
  if (props.definition) {
    form.value = {
      name: props.definition.name || '',
      nodeTemplates: parseNodeTemplates(props.definition.nodeTemplates),
      extensionPoints: parseExtensionPoints(props.definition.extensionPoints)
    }
  } else {
    form.value = {
      name: '',
      nodeTemplates: [createDefaultStep()],
      extensionPoints: {
        approverResolverUrl: '',
        permissionValidatorUrl: '',
        timeoutSeconds: 3
      }
    }
  }
}

function parseNodeTemplates(templates) {
  if (!templates) return [createDefaultStep()]
  try {
    const parsed = typeof templates === 'string' ? JSON.parse(templates) : templates
    return parsed.map(t => ({
      ...t,
      assignees: (t.assignees || []).map(a => JSON.stringify(a))
    }))
  } catch {
    return [createDefaultStep()]
  }
}

function parseExtensionPoints(points) {
  if (!points) {
    return {
      approverResolverUrl: '',
      permissionValidatorUrl: '',
      timeoutSeconds: 3
    }
  }
  try {
    return typeof points === 'string' ? JSON.parse(points) : points
  } catch {
    return {
      approverResolverUrl: '',
      permissionValidatorUrl: '',
      timeoutSeconds: 3
    }
  }
}

function createDefaultStep() {
  return {
    type: 'approval',
    jointSignPolicy: 'ALL_PASS',
    assignees: []
  }
}

function addStep() {
  form.value.nodeTemplates.push(createDefaultStep())
}

function removeStep(index) {
  form.value.nodeTemplates.splice(index, 1)
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  
  // 验证步骤
  for (let i = 0; i < form.value.nodeTemplates.length; i++) {
    const step = form.value.nodeTemplates[i]
    if (!step.assignees || step.assignees.length === 0) {
      ElMessage.warning(`步骤 ${i + 1} 需要至少一个审批人`)
      return
    }
  }
  
  loading.value = true
  
  try {
    const data = {
      name: form.value.name,
      nodeTemplates: form.value.nodeTemplates.map((step, index) => ({
        stepIndex: index,
        type: step.type,
        jointSignPolicy: step.type === 'joint_sign' ? step.jointSignPolicy : undefined,
        assignees: step.assignees.map(a => {
          try {
            return JSON.parse(a)
          } catch {
            return a
          }
        })
      })),
      extensionPoints: form.value.extensionPoints
    }
    
    if (isEdit.value) {
      await updateDefinition(props.definition.id, data)
      ElMessage.success('更新成功')
    } else {
      await createDefinition(data)
      ElMessage.success('创建成功')
    }
    
    visible.value = false
    emit('success')
  } catch (error) {
    console.error('Submit failed:', error)
  } finally {
    loading.value = false
  }
}

function open() {
  initForm()
  visible.value = true
}

defineExpose({ open })
</script>

<style scoped>
.definition-form {
  max-height: 60vh;
  overflow-y: auto;
}

.steps-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.step-card {
  background: var(--el-fill-color-light);
}

.step-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.add-step-btn {
  align-self: flex-start;
}
</style>
