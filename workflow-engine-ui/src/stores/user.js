import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

// 预定义测试用户（ID需与后端扩展点默认审批人一致）
export const TEST_USERS = [
  { id: 'user_001', name: '张三', role: 'employee', avatar: '👤' },
  { id: 'manager_001', name: '李四', role: 'manager', avatar: '👨‍💼' },
  { id: 'director_001', name: '王五', role: 'director', avatar: '👔' },
  { id: 'hr_001', name: '赵六', role: 'hr', avatar: '👩‍💼' },
  { id: 'admin_001', name: '管理员', role: 'admin', avatar: '🔧' }
]

// 从 localStorage 读取保存的用户
const getStoredUser = () => {
  try {
    const stored = localStorage.getItem('currentUser')
    if (stored) {
      const parsed = JSON.parse(stored)
      // 验证用户是否存在于 TEST_USERS
      if (TEST_USERS.find(u => u.id === parsed.id)) {
        return parsed
      }
    }
  } catch (e) {
    console.error('Failed to parse stored user:', e)
  }
  return null
}

export const useUserStore = defineStore('user', () => {
  // State - 优先从 localStorage 读取
  const currentUser = ref(getStoredUser() || TEST_USERS[0])
  
  // Getters
  const isManager = computed(() => 
    ['manager', 'director', 'admin'].includes(currentUser.value?.role)
  )
  
  const isAdmin = computed(() => 
    currentUser.value?.role === 'admin'
  )
  
  // Actions
  function switchUser(user) {
    currentUser.value = user
    // 持久化到 localStorage
    try {
      localStorage.setItem('currentUser', JSON.stringify(user))
    } catch (e) {
      console.error('Failed to save user:', e)
    }
  }
  
  function getUserById(userId) {
    return TEST_USERS.find(u => u.id === userId)
  }
  
  return {
    currentUser,
    isManager,
    isAdmin,
    switchUser,
    getUserById,
    TEST_USERS
  }
})
