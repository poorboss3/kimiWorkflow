<template>
  <el-container class="app-container">
    <!-- 侧边栏 -->
    <el-aside width="220px" class="sidebar">
      <div class="logo">
        <el-icon :size="28" color="#409EFF"><Connection /></el-icon>
        <span class="logo-text">Workflow Engine</span>
      </div>
      
      <el-menu
        :default-active="$route.path"
        router
        class="nav-menu"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
      >
        <el-menu-item index="/">
          <el-icon><HomeFilled /></el-icon>
          <span>首页</span>
        </el-menu-item>
        
        <el-menu-item index="/definitions">
          <el-icon><Document /></el-icon>
          <span>流程定义</span>
        </el-menu-item>
        
        <el-menu-item index="/instances">
          <el-icon><List /></el-icon>
          <span>流程实例</span>
        </el-menu-item>
        
        <el-menu-item index="/tasks">
          <el-icon><Bell /></el-icon>
          <span>任务中心</span>
        </el-menu-item>
        
        <el-sub-menu index="/config">
          <template #title>
            <el-icon><Setting /></el-icon>
            <span>配置管理</span>
          </template>
          <el-menu-item index="/proxy">
            <el-icon><UserFilled /></el-icon>
            <span>代理配置</span>
          </el-menu-item>
          <el-menu-item index="/delegation">
            <el-icon><User /></el-icon>
            <span>委托配置</span>
          </el-menu-item>
        </el-sub-menu>
      </el-menu>
    </el-aside>
    
    <el-container>
      <!-- 顶部栏 -->
      <el-header class="header">
        <div class="header-left">
          <breadcrumb />
        </div>
        <div class="header-right">
          <el-tag type="info" effect="plain" class="api-status">
            <el-icon v-if="apiConnected" color="#67C23A"><CircleCheck /></el-icon>
            <el-icon v-else color="#F56C6C"><CircleClose /></el-icon>
            API: {{ apiConnected ? '已连接' : '未连接' }}
          </el-tag>
          <user-switcher />
        </div>
      </el-header>
      
      <!-- 主内容区 -->
      <el-main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import UserSwitcher from '@/components/UserSwitcher.vue'

const apiConnected = ref(false)

onMounted(async () => {
  // 检查 API 连接状态（健康检查端点是 /health，不是 /api/v1/health）
  try {
    const response = await fetch('/health', { 
      method: 'GET',
      headers: { 'X-User-ID': 'system' }
    })
    apiConnected.value = response.ok
  } catch {
    apiConnected.value = false
  }
})
</script>

<style scoped>
.app-container {
  height: 100vh;
}

/* 重置 el-aside 默认边距 */
:deep(.el-aside) {
  margin: 0 !important;
  padding: 0 !important;
}

.sidebar {
  background-color: #304156;
  display: flex;
  flex-direction: column;
  margin: 0;
  padding: 0;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  border-bottom: 1px solid #1f2d3d;
}

.logo-text {
  font-size: 18px;
  font-weight: 600;
  color: #fff;
}

.nav-menu {
  flex: 1;
  border-right: none !important;
  margin: 0 !important;
  padding: 0 !important;
}

/* 移除菜单项的默认左边距 */
:deep(.el-menu) {
  border-right: none !important;
}

:deep(.el-menu-item),
:deep(.el-sub-menu__title) {
  margin: 0 !important;
}

.header {
  background-color: #fff;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.api-status {
  display: flex;
  align-items: center;
  gap: 4px;
}

.main-content {
  background-color: #f0f2f5;
  padding: 20px;
  overflow-y: auto;
}

/* 页面切换动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
