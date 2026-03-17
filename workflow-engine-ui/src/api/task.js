import request from '@/utils/request'

// ==================== 任务 API ====================

// 获取待办任务
export function getPendingTasks(params) {
  return request.get('/tasks/pending', { params })
}

// 获取已办任务
export function getCompletedTasks(params) {
  return request.get('/tasks/completed', { params })
}

// 获取任务统计
export function getTaskStatistics() {
  return request.get('/tasks/statistics')
}

// 获取任务详情
export function getTaskDetail(id) {
  return request.get(`/tasks/${id}`)
}

// 处理任务
export function processTask(id, data) {
  return request.post(`/tasks/${id}/action`, data)
}

// ==================== 快捷操作 ====================

// 通过任务
export function approveTask(id, comment) {
  return processTask(id, { action: 'approve', comment })
}

// 驳回任务
export function rejectTask(id, comment) {
  return processTask(id, { action: 'reject', comment })
}

// 退回任务
export function returnTask(id, comment, returnToStep) {
  return processTask(id, { action: 'return', comment, returnToStep })
}

// 加签
export function countersignTask(id, comment, countersignData) {
  return processTask(id, { 
    action: 'countersign', 
    comment, 
    countersignData 
  })
}

// 委托任务
export function delegateTask(id, comment, delegateeId) {
  return processTask(id, { 
    action: 'delegate', 
    comment,
    delegateeId
  })
}
