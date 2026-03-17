<template>
  <div class="user-switcher">
    <el-dropdown @command="handleSwitchUser" trigger="click">
      <div class="current-user">
        <el-avatar :size="32" class="user-avatar">
          {{ currentUser?.avatar }}
        </el-avatar>
        <div class="user-info">
          <span class="user-name">{{ currentUser?.name }}</span>
          <span class="user-role">{{ getRoleLabel(currentUser?.role) }}</span>
        </div>
        <el-icon class="dropdown-icon"><ArrowDown /></el-icon>
      </div>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item 
            v-for="user in TEST_USERS" 
            :key="user.id"
            :command="user"
            :disabled="user.id === currentUser?.id"
          >
            <div class="dropdown-user-item">
              <span class="user-avatar-small">{{ user.avatar }}</span>
              <span class="user-name-text">{{ user.name }}</span>
              <el-tag size="small" type="info">{{ getRoleLabel(user.role) }}</el-tag>
              <el-icon v-if="user.id === currentUser?.id" class="check-icon"><Check /></el-icon>
            </div>
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useUserStore, TEST_USERS } from '@/stores/user'
import { ElMessage } from 'element-plus'

const userStore = useUserStore()
const currentUser = computed(() => userStore.currentUser)

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

function handleSwitchUser(user) {
  userStore.switchUser(user)
  ElMessage.success(`已切换到: ${user.name}`)
  // 刷新页面以更新数据
  window.location.reload()
}
</script>

<style scoped>
.user-switcher {
  display: inline-block;
}

.current-user {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  cursor: pointer;
  border-radius: 8px;
  transition: background-color 0.2s;
}

.current-user:hover {
  background-color: var(--el-fill-color-light);
}

.user-info {
  display: flex;
  flex-direction: column;
  line-height: 1.2;
}

.user-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.user-role {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.dropdown-icon {
  color: var(--el-text-color-secondary);
}

.dropdown-user-item {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 160px;
}

.user-avatar-small {
  font-size: 18px;
}

.user-name-text {
  flex: 1;
}

.check-icon {
  color: var(--el-color-success);
}
</style>
