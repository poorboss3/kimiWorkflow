import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useProcessStore = defineStore('process', () => {
  // State
  const currentDefinition = ref(null)
  const currentInstance = ref(null)
  const currentTask = ref(null)
  
  // Actions
  function setCurrentDefinition(definition) {
    currentDefinition.value = definition
  }
  
  function setCurrentInstance(instance) {
    currentInstance.value = instance
  }
  
  function setCurrentTask(task) {
    currentTask.value = task
  }
  
  function clearCurrent() {
    currentDefinition.value = null
    currentInstance.value = null
    currentTask.value = null
  }
  
  return {
    currentDefinition,
    currentInstance,
    currentTask,
    setCurrentDefinition,
    setCurrentInstance,
    setCurrentTask,
    clearCurrent
  }
})
