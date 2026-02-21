<template>
  <!-- 遮罩 -->
  <div 
    v-if="visible" 
    class="fixed inset-0 bg-[#0f0721]/60 backdrop-blur-md opacity-100 transition-opacity duration-300 z-[60]"
    @click="visible = false"
  ></div>
  
  <!-- 抽屉 -->
  <aside 
    v-if="visible"
    class="fixed top-6 right-6 bottom-6 w-full max-w-sm bg-white/95 backdrop-blur-xl border border-white/40 rounded-[40px] z-[70] shadow-2xl flex flex-col overflow-hidden animate-scale-in"
  >
    <div class="p-8 border-b border-primary/5 flex items-center justify-between">
      <div>
        <h3 class="font-extrabold text-2xl text-[#1e0a3d]">任务中心</h3>
        <p class="text-[10px] font-bold text-primary/50 mt-1 uppercase tracking-widest">Active Tasks</p>
      </div>
      <button 
        class="p-2.5 rounded-2xl bg-red-50 text-red-500 hover:bg-red-100 transition-colors"
        @click="visible = false"
      >
        <X class="w-6 h-6" />
      </button>
    </div>
    
    <div class="flex-1 overflow-y-auto p-8 space-y-6">
      <!-- 过滤标签 -->
      <div class="flex gap-2 flex-wrap">
        <button 
          v-for="tab in tabs" 
          :key="tab.key"
          @click="filter = tab.key"
          class="px-4 py-2 rounded-xl text-xs font-bold transition-all"
          :class="filter === tab.key ? 'bg-primary text-white shadow-lg shadow-primary/20' : 'bg-primary/5 text-primary hover:bg-primary/10'"
        >
          {{ tab.label }}
        </button>
      </div>

      <!-- 任务列表 -->
      <div class="space-y-4">
        <div 
          v-for="task in filteredTasks" 
          :key="task.id"
          class="p-6 bg-primary/5 border border-primary/10 rounded-[32px] space-y-4"
        >
          <div class="flex justify-between items-start">
            <div class="flex gap-4">
              <div class="w-12 h-12 rounded-2xl bg-primary text-white flex items-center justify-center shrink-0 shadow-lg shadow-primary/20">
                <Upload v-if="task.type === 'upload'" class="w-6 h-6" />
                <Download v-else class="w-6 h-6" />
              </div>
              <div class="truncate">
                <div class="text-sm font-extrabold text-[#1e0a3d] truncate">{{ task.name }}</div>
                <div class="flex items-center gap-2 mt-1">
                  <span class="text-[10px] font-bold text-primary px-2 py-0.5 bg-primary/10 rounded-full">{{ task.progress }}%</span>
                  <span class="text-[10px] font-bold text-primary/40">{{ task.sizeLabel }}</span>
                </div>
              </div>
            </div>
            <button 
              v-if="task.status === 'running'"
              @click="cancelTask(task.id)"
              class="p-2 text-primary/40 hover:text-red-500 transition-colors"
            >
              <X class="w-4 h-4" />
            </button>
          </div>
          <div class="h-2.5 bg-white rounded-full overflow-hidden shadow-inner">
            <div 
              class="h-full bg-primary shadow-[0_0_12px_rgba(124,58,237,0.4)] transition-all duration-500"
              :style="{ width: task.progress + '%' }"
            ></div>
          </div>
        </div>

        <!-- 空状态 -->
        <div v-if="filteredTasks.length === 0" class="text-center py-12 text-primary/40">
          <Inbox class="w-12 h-12 mx-auto mb-4 opacity-30" />
          <p class="text-sm font-bold">暂无任务</p>
        </div>
      </div>
    </div>
  </aside>
</template>

<script setup>
import { ref, computed } from 'vue'
import { X, Upload, Download, Inbox } from 'lucide-vue-next'
import { useTaskCenter } from '../store/taskCenter'

const visible = defineModel('visible', { default: false })
const state = useTaskCenter()

const filter = ref('all')
const tabs = [
  { key: 'all', label: '全部' },
  { key: 'upload', label: '上传' },
  { key: 'download', label: '下载' },
  { key: 'done', label: '已完成' },
]

const filteredTasks = computed(() => {
  if (filter.value === 'all') return state.tasks
  if (filter.value === 'done') return state.tasks.filter(t => t.status === 'done')
  return state.tasks.filter(t => t.type === filter.value && t.status === 'running')
})

function cancelTask(id) {
  const index = state.tasks.findIndex(t => t.id === id)
  if (index > -1) {
    state.tasks.splice(index, 1)
  }
}
</script>
