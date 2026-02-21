import { reactive } from 'vue'

// 简单的事件总线
export const eventBus = reactive({
  refreshFiles: false,
  triggerRefresh() {
    this.refreshFiles = true
    setTimeout(() => {
      this.refreshFiles = false
    }, 100)
  }
})
