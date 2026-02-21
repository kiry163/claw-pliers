<template>
  <section class="h-full flex flex-col p-4">
    <!-- 目标文件夹选择/显示 -->
    <div class="mb-4 bg-white/60 border border-white/40 rounded-2xl p-4 shadow-sm">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <Folder class="w-5 h-5 text-amber-500" />
            <div>
              <span class="text-xs font-bold text-primary/40 uppercase tracking-widest">上传到</span>
              <div class="flex items-center gap-1 text-sm font-bold text-[#1e0a3d]" v-if="targetFolderPath.length > 0">
                <span v-for="(item, index) in targetFolderPath" :key="index" class="flex items-center">
                  <span v-if="index > 0" class="text-primary/30 mx-1">/</span>
                  <span :class="index === targetFolderPath.length - 1 ? 'text-primary' : ''">{{ item.name }}</span>
                </span>
              </div>
              <div v-else class="text-sm font-bold text-primary">根目录</div>
            </div>
        </div>
        
        <!-- 更改目标按钮 -->
        <button 
          @click="showFolderSelector = true"
          class="px-4 py-2 text-xs font-bold text-primary bg-primary/5 hover:bg-primary/10 rounded-xl transition-all border border-primary/10"
        >
          更改
        </button>
      </div>
    </div>

    <!-- 上传区域 -->
    <div 
      class="flex-1 bg-white/40 border-2 border-dashed border-primary/20 rounded-[48px] flex flex-col items-center justify-center gap-8 group cursor-pointer transition-all"
      @dragover.prevent
      @drop.prevent="onDrop"
      @click="selectFiles"
    >
      <div class="w-32 h-32 rounded-[40px] bg-primary/10 text-primary flex items-center justify-center group-hover:scale-110 transition-transform shadow-lg shadow-primary/5">
        <UploadCloud class="w-16 h-16" />
      </div>
      <div class="text-center">
        <h3 class="text-3xl font-extrabold text-[#1e0a3d]">点击或拖拽上传</h3>
        <p class="text-primary/50 font-bold mt-3 text-lg">任务状态将在右上角任务中心显示</p>
      </div>
      <button class="px-12 py-5 bg-primary text-white font-bold rounded-2xl shadow-xl shadow-primary/30 hover:scale-105 active:scale-95 transition-all text-lg">
        选择文件
      </button>
      <input 
        ref="fileInput" 
        type="file" 
        class="hidden" 
        multiple 
        @change="onSelect" 
      />
    </div>
  </section>

  <!-- 文件夹选择器模态框 -->
  <div v-if="showFolderSelector" class="fixed inset-0 z-[150] bg-[#0f0721]/60 backdrop-blur-md flex items-center justify-center p-6">
    <div class="w-full max-w-md glass-solid p-6 rounded-[32px] border border-white/60 shadow-2xl flex flex-col animate-scale-in max-h-[80vh]">
      <div class="flex items-center justify-between mb-6">
        <div>
          <h3 class="text-xl font-extrabold text-[#1e0a3d]">选择目标文件夹</h3>
          <p class="text-xs text-primary/50 font-bold mt-1">选择文件上传的目标位置</p>
        </div>
        <button 
          @click="showFolderSelector = false"
          class="p-2 rounded-xl hover:bg-primary/10 text-primary transition-colors"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- 面包屑导航 -->
      <div class="flex items-center gap-1 text-xs font-bold text-primary/50 mb-4 flex-wrap">
        <button 
          @click="selectorFolderId = null; loadSelectorFolders()"
          class="hover:text-primary transition-colors"
          :class="{ 'text-primary': !selectorFolderId }"
        >
          根目录
        </button>
        <template v-for="(crumb, index) in selectorBreadcrumbs" :key="index">
          <ChevronRight class="w-3 h-3" />
          <button 
            @click="selectorFolderId = crumb.folder_id; loadSelectorFolders()"
            class="hover:text-primary transition-colors"
            :class="{ 'text-primary': crumb.folder_id === selectorFolderId }"
          >
            {{ crumb.name }}
          </button>
        </template>
      </div>

      <!-- 文件夹列表 -->
      <div class="flex-1 overflow-y-auto space-y-2 min-h-[200px] max-h-[400px]">
        <!-- 根目录选项 -->
        <button 
          v-if="selectorFolderId"
          @click="selectTargetFolder(null)"
          class="w-full flex items-center gap-3 p-4 rounded-2xl hover:bg-primary/5 transition-all text-left group"
        >
          <div class="w-10 h-10 rounded-xl bg-primary/10 text-primary flex items-center justify-center group-hover:bg-primary group-hover:text-white transition-all">
            <ArrowLeft class="w-5 h-5" />
          </div>
          <div>
            <div class="font-bold text-sm text-[#1e0a3d]">根目录</div>
            <div class="text-xs text-primary/40">返回根目录上传</div>
          </div>
        </button>

        <!-- 当前文件夹中的子文件夹 -->
        <button 
          v-for="folder in selectorFolders" 
          :key="folder.folder_id"
          @click="enterSelectorFolder(folder)"
          class="w-full flex items-center gap-3 p-4 rounded-2xl hover:bg-primary/5 transition-all text-left group"
        >
          <div class="w-10 h-10 rounded-xl bg-amber-100 text-amber-600 flex items-center justify-center">
            <Folder class="w-5 h-5" />
          </div>
          <div class="flex-1">
            <div class="font-bold text-sm text-[#1e0a3d]">{{ folder.name }}</div>
            <div class="text-xs text-primary/40">{{ folder.item_count }} 个项目</div>
          </div>
          <ChevronRight class="w-4 h-4 text-primary/30" />
        </button>

        <!-- 空状态 -->
        <div v-if="selectorFolders.length === 0 && !selectorFolderId" class="text-center py-8 text-primary/40">
          <FolderOpen class="w-12 h-12 mx-auto mb-3 opacity-30" />
          <p class="text-sm font-bold">根目录没有子文件夹</p>
          <p class="text-xs mt-1">文件将上传到根目录</p>
        </div>
      </div>

      <!-- 底部操作 -->
      <div class="flex gap-3 mt-6 pt-4 border-t border-primary/10">
        <button 
          @click="showFolderSelector = false"
          class="flex-1 py-3 bg-white border border-primary/10 text-primary font-bold rounded-2xl hover:bg-primary/5 transition-all"
        >
          取消
        </button>
        <button 
          @click="confirmFolderSelection"
          class="flex-1 py-3 bg-primary text-white font-bold rounded-2xl hover:scale-105 transition-all shadow-lg shadow-primary/20"
        >
          确认选择
        </button>
      </div>
    </div>
  </div>

  <!-- 复制对话框 -->
  <div v-if="copyDialogOpen" class="fixed inset-0 z-[150] bg-[#0f0721]/60 backdrop-blur-md flex items-center justify-center p-6">
    <div class="w-full max-w-sm glass-solid p-8 rounded-[40px] border border-white/60 shadow-2xl flex flex-col animate-scale-in">
      <h3 class="text-xl font-extrabold text-[#1e0a3d] mb-2">手动复制</h3>
      <p class="text-sm text-primary/50 font-bold mb-4">自动复制失败，请手动复制：</p>
      <input 
        :value="copyDialogText"
        readonly
        class="w-full bg-white/60 border border-white/40 rounded-2xl px-4 py-3 font-mono text-sm font-bold mb-6 outline-none"
      >
      <button 
        @click="copyDialogOpen = false"
        class="w-full py-3 bg-primary text-white font-bold rounded-2xl hover:scale-105 transition-all"
      >
        关闭
      </button>
    </div>
  </div>

  <!-- Toast -->
  <div 
    class="fixed bottom-12 left-1/2 -translate-x-1/2 px-8 py-4 bg-[#1e0a3d] text-white rounded-2xl border border-white/10 shadow-2xl transition-all duration-300 z-[200] font-extrabold text-sm"
    :class="toast.show ? 'translate-y-0 opacity-100' : 'translate-y-20 opacity-0 pointer-events-none'"
  >
    {{ toast.message }}
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { UploadCloud, Folder, X, ChevronRight, ArrowLeft, FolderOpen } from 'lucide-vue-next'
import { uploadFile } from '../api/files'
import { listFolders, getFolderContents } from '../api/folders'
import { addTask, updateTask, completeTask, failTask } from '../store/taskCenter'
import { eventBus } from '../store/eventBus'

const route = useRoute()
const fileInput = ref(null)
const copyDialogOpen = ref(false)
const copyDialogText = ref('')
const toast = ref({ show: false, message: '' })

// 目标文件夹相关
const targetFolderId = ref(null)
const targetFolderPath = ref([])
const showFolderSelector = ref(false)
const selectorFolderId = ref(null)
const selectorFolders = ref([])
const selectorBreadcrumbs = ref([])

// 从 URL 查询参数获取 folder_id
onMounted(async () => {
  const folderIdFromUrl = route.query.folder
  if (folderIdFromUrl) {
    targetFolderId.value = folderIdFromUrl
    await loadTargetFolderPath(folderIdFromUrl)
  }
})

// 加载目标文件夹路径
async function loadTargetFolderPath(folderId) {
  try {
    const res = await getFolderContents(folderId)
    const data = res.data?.data
    if (data?.breadcrumbs) {
      // 添加当前文件夹到路径末尾
      targetFolderPath.value = [...data.breadcrumbs, { folder_id: folderId, name: data.name }]
    }
  } catch (error) {
    console.error('加载文件夹路径失败:', error)
  }
}

// 加载选择器中的文件夹
async function loadSelectorFolders() {
  try {
    const res = await listFolders(selectorFolderId.value)
    selectorFolders.value = res.data?.data?.folders || []
    
    // 更新面包屑
    if (selectorFolderId.value) {
      const contentsRes = await getFolderContents(selectorFolderId.value)
      const data = contentsRes.data?.data
      selectorBreadcrumbs.value = data?.breadcrumbs?.slice(1) || [] // 去掉根目录
    } else {
      selectorBreadcrumbs.value = []
    }
  } catch (error) {
    console.error('加载文件夹列表失败:', error)
  }
}

// 进入选择器中的子文件夹
function enterSelectorFolder(folder) {
  selectorFolderId.value = folder.folder_id
  loadSelectorFolders()
}

// 选择目标文件夹
function selectTargetFolder(folderId) {
  targetFolderId.value = folderId
  if (folderId) {
    loadTargetFolderPath(folderId)
  } else {
    targetFolderPath.value = []
  }
  showFolderSelector.value = false
}

// 确认文件夹选择
function confirmFolderSelection() {
  selectTargetFolder(selectorFolderId.value)
}

// 打开选择器时加载数据
watch(() => showFolderSelector.value, (newVal) => {
  if (newVal) {
    selectorFolderId.value = targetFolderId.value
    loadSelectorFolders()
  }
})

function showToast(message) {
  toast.value = { show: true, message }
  setTimeout(() => toast.value.show = false, 3000)
}

const selectFiles = () => fileInput.value?.click()

const onSelect = (event) => {
  const files = Array.from(event.target.files || [])
  handleFiles(files)
  event.target.value = ''
}

const onDrop = (event) => {
  const files = Array.from(event.dataTransfer.files || [])
  handleFiles(files)
}

const handleFiles = (files) => {
  files.forEach((file) => startUpload(file))
}

const formatSize = (size) => {
  if (!size) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let value = size
  let idx = 0
  while (value > 1024 && idx < units.length - 1) {
    value /= 1024
    idx += 1
  }
  return `${value.toFixed(1)} ${units[idx]}`
}

const startUpload = async (file) => {
  const id = `${Date.now()}-${file.name}`
  addTask({
    id,
    name: file.name,
    type: 'upload',
    progress: 0,
    status: 'running',
    sizeLabel: formatSize(file.size),
  })
  try {
    const response = await uploadFile(file, targetFolderId.value, (event) => {
      const percent = event.total ? Math.round((event.loaded / event.total) * 100) : 0
      updateTask(id, { progress: percent })
    })
    completeTask(id)
    showToast('上传完成')
    // 触发文件列表刷新
    eventBus.triggerRefresh()
  } catch (error) {
    failTask(id)
    showToast('上传失败')
    console.error('上传错误:', error)
  }
}
</script>
