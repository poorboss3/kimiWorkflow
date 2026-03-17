import request from '@/utils/request'

// ==================== 流程定义 API ====================

// 创建流程定义
export function createDefinition(data) {
  return request.post('/processes/definitions', data)
}

// 获取流程定义列表
export function listDefinitions(params) {
  return request.get('/processes/definitions', { params })
}

// 获取流程定义详情
export function getDefinition(id) {
  return request.get(`/processes/definitions/${id}`)
}

// 更新流程定义
export function updateDefinition(id, data) {
  return request.put(`/processes/definitions/${id}`, data)
}

// 激活流程定义
export function activateDefinition(id, version) {
  return request.post(`/processes/definitions/${id}/activate`, { version })
}

// 归档流程定义
export function archiveDefinition(id) {
  return request.post(`/processes/definitions/${id}/archive`)
}

// ==================== 流程实例 API ====================

// 提交流程
export function submitProcess(data) {
  return request.post('/processes/instances', data)
}

// 获取流程实例列表
export function listInstances(params) {
  return request.get('/processes/instances', { params })
}

// 获取流程实例详情
export function getInstance(id) {
  return request.get(`/processes/instances/${id}`)
}

// 撤回流程
export function withdrawInstance(id, reason) {
  return request.post(`/processes/instances/${id}/withdraw`, { reason })
}

// 获取流程审批历史
export function getInstanceHistory(id) {
  return request.get(`/processes/instances/${id}/history`)
}

// 修改审批步骤
export function modifySteps(id, data) {
  return request.put(`/processes/instances/${id}/steps`, data)
}

// 标记加急
export function markUrgent(id) {
  return request.post(`/processes/instances/${id}/urgent`)
}
