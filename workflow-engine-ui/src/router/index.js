import { createRouter, createWebHistory } from 'vue-router'
import Home from '@/views/Home.vue'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: Home,
    meta: { title: '首页' }
  },
  {
    path: '/definitions',
    name: 'ProcessDefinition',
    component: () => import('@/views/ProcessDefinition.vue'),
    meta: { title: '流程定义' }
  },
  {
    path: '/instances',
    name: 'ProcessInstance',
    component: () => import('@/views/ProcessInstance.vue'),
    meta: { title: '流程实例' }
  },
  {
    path: '/tasks',
    name: 'TaskList',
    component: () => import('@/views/TaskList.vue'),
    meta: { title: '任务中心' }
  },
  {
    path: '/proxy',
    name: 'ProxyConfig',
    component: () => import('@/views/ProxyConfig.vue'),
    meta: { title: '代理配置' }
  },
  {
    path: '/delegation',
    name: 'DelegationConfig',
    component: () => import('@/views/DelegationConfig.vue'),
    meta: { title: '委托配置' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  document.title = to.meta.title 
    ? `${to.meta.title} - Workflow Engine` 
    : 'Workflow Engine'
  next()
})

export default router
