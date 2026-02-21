<template>
  <section class="view active space-y-6">
    <!-- 头部区域 -->
    <div class="flex flex-col gap-4">
      <div class="flex items-end justify-between">
        <div>
          <h2 class="text-2xl md:text-3xl font-extrabold tracking-tight">{{ currentFolderName }}</h2>
          <!-- 面包屑导航 -->
          <nav class="flex items-center gap-2 text-xs md:text-sm font-bold mt-1">
            <template v-for="(crumb, index) in breadcrumbs" :key="crumb.folder_id || index">
              <span v-if="index > 0" class="text-primary/30">/</span>
              <button 
                @click="goToFolder(crumb.folder_id)"
                class="text-primary hover:underline"
              >
                {{ crumb.name }}
              </button>
            </template>
            <template v-if="currentFolderId">
              <span v-if="breadcrumbs.length > 0" class="text-primary/30">/</span>
              <span class="text-primary/50">{{ currentFolderName }}</span>
            </template>
          </nav>
        </div>
        <!-- 视图切换 -->
        <div class="flex items-center gap-1 p-1 bg-white/40 rounded-2xl border border-white/60 shadow-sm">
          <button 
            class="p-2 rounded-xl transition-all duration-300"
            :class="layout === 'list' ? 'bg-white shadow-sm text-primary' : 'text-primary/40 hover:text-primary'"
            @click="layout = 'list'"
          >
            <List class="w-4 h-4" />
          </button>
          <button 
            class="p-2 rounded-xl transition-all duration-300"
            :class="layout === 'grid' ? 'bg-white shadow-sm text-primary' : 'text-primary/40 hover:text-primary'"
            @click="layout = 'grid'"
          >
            <LayoutGrid class="w-4 h-4" />
          </button>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="flex items-center gap-2 overflow-x-auto no-scrollbar pb-1">
        <router-link :to="uploadLink" class="btn-primary-glow px-5 py-2.5 rounded-2xl flex items-center gap-2 font-bold text-xs whitespace-nowrap">
          <Plus class="w-4 h-4" />
          <span>新建/上传</span>
        </router-link>
        <button 
          @click="showNewFolderModal = true"
          class="bg-white/60 hover:bg-white border border-white/60 px-5 py-2.5 rounded-2xl flex items-center gap-2 font-bold text-xs text-primary transition-all shadow-sm whitespace-nowrap"
        >
          <FolderPlus class="w-4 h-4" />
          <span>新建文件夹</span>
        </button>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="flex items-center justify-center py-20">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
    </div>

    <!-- 空状态 -->
    <div v-else-if="folders.length === 0 && files.length === 0" class="flex flex-col items-center justify-center py-20 text-primary/50">
      <FolderOpen class="w-16 h-16 mb-4 opacity-30" />
      <p class="font-bold">当前文件夹为空</p>
      <p class="text-sm mt-2">点击上方按钮创建文件夹或上传文件</p>
    </div>

    <!-- 内容区域 -->
    <div v-else class="space-y-6">
      <!-- 文件夹列表 -->
      <div v-if="folders.length > 0" class="space-y-2">
        <h3 class="text-xs font-bold text-primary/40 uppercase tracking-widest px-2">文件夹</h3>
        
        <!-- 列表视图 -->
        <div v-if="layout === 'list'" class="space-y-2">
          <div 
            v-for="folder in folders" 
            :key="folder.folder_id"
            class="folder-row flex items-center gap-4 px-4 py-4 bg-white/60 hover:bg-white border border-white/20 rounded-2xl cursor-pointer shadow-sm"
            @click="goToFolder(folder.folder_id)"
          >
            <div class="w-11 h-11 rounded-xl bg-amber-100 text-amber-600 flex items-center justify-center shrink-0 shadow-inner">
              <Folder class="w-6 h-6 fill-amber-600/20" />
            </div>
            <div class="flex-1 min-w-0">
              <div class="font-extrabold text-sm text-[#1e0a3d] truncate">{{ folder.name }}</div>
              <div class="text-[10px] text-primary/40 font-bold uppercase tracking-tight">{{ folder.item_count }} 个项目</div>
            </div>
            <div class="hidden md:block text-xs font-bold text-primary/30">--</div>
            <div class="hidden md:block text-xs font-bold text-primary/30">{{ formatDate(folder.created_at) }}</div>
            <div class="relative">
              <button 
                class="p-2 rounded-xl hover:bg-primary/10 text-primary transition-colors"
                @click.stop="showFolderMenu($event, folder)"
              >
                <MoreVertical class="w-5 h-5" />
              </button>
            </div>
          </div>
        </div>

        <!-- 网格视图 -->
        <div v-else class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 xl:grid-cols-6 gap-4">
          <div 
            v-for="folder in folders" 
            :key="folder.folder_id"
            class="group relative flex flex-col items-center p-6 bg-white/60 hover:bg-white border border-white/20 rounded-[32px] cursor-pointer shadow-sm hover:shadow-xl hover:-translate-y-1 transition-all duration-300"
            @click="goToFolder(folder.folder_id)"
          >
            <div class="w-20 h-20 mb-4 rounded-3xl bg-amber-100 text-amber-600 flex items-center justify-center shadow-inner group-hover:scale-110 transition-transform">
              <Folder class="w-10 h-10 fill-amber-600/20" />
            </div>
            <div class="font-bold text-sm text-center truncate w-full px-2">{{ folder.name }}</div>
            <div class="text-[10px] font-bold text-primary/30 mt-1 uppercase">{{ folder.item_count }} items</div>
            <button 
              class="absolute top-2 right-2 p-2 opacity-0 group-hover:opacity-100 transition-opacity text-primary"
              @click.stop="showFolderMenu($event, folder)"
            >
              <MoreVertical class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      <!-- 文件列表 -->
      <div v-if="files.length > 0" class="space-y-2">
        <h3 class="text-xs font-bold text-primary/40 uppercase tracking-widest px-2">文件</h3>
        
        <!-- 列表视图 -->
        <div v-if="layout === 'list'" class="space-y-2">
          <div 
            v-for="file in files" 
            :key="file.file_id"
            class="file-row flex items-center gap-4 px-4 py-4 bg-white/60 hover:bg-white border border-white/20 rounded-2xl cursor-pointer shadow-sm"
            @click="goToFile(file.file_id)"
          >
            <div class="w-11 h-11 rounded-xl flex items-center justify-center shrink-0 shadow-inner"
              :class="getFileIconClass(file.mime_type)">
              <FileText class="w-6 h-6" />
            </div>
            <div class="flex-1 min-w-0">
              <div class="font-extrabold text-sm text-[#1e0a3d] truncate">{{ file.original_name }}</div>
              <div class="text-[10px] text-primary/40 font-mono font-bold tracking-tight">filehub://{{ file.file_id }}</div>
            </div>
            <div class="hidden md:block text-xs font-bold text-primary/40 w-24 text-right">{{ formatSize(file.size) }}</div>
            <div class="hidden md:block text-xs font-bold text-primary/30 w-32 text-right">{{ formatDate(file.created_at) }}</div>
            <div class="flex items-center gap-1">
              <button 
                class="p-2 rounded-xl hover:bg-primary/10 text-primary transition-colors"
                @click.stop="handleDownload(file)"
              >
                <Download class="w-5 h-5" />
              </button>
              <button 
                class="p-2 rounded-xl hover:bg-primary/10 text-primary transition-colors"
                @click.stop="showFileMenu($event, file)"
              >
                <MoreVertical class="w-5 h-5" />
              </button>
            </div>
          </div>
        </div>

        <!-- 网格视图 -->
        <div v-else class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 xl:grid-cols-6 gap-4">
          <div 
            v-for="file in files" 
            :key="file.file_id"
            class="group relative flex flex-col items-center p-6 bg-white/60 hover:bg-white border border-white/20 rounded-[32px] cursor-pointer shadow-sm hover:shadow-xl hover:-translate-y-1 transition-all duration-300"
            @click="goToFile(file.file_id)"
          >
            <div class="w-20 h-20 mb-4 rounded-3xl flex items-center justify-center shadow-inner group-hover:scale-110 transition-transform"
              :class="getFileIconClass(file.mime_type)">
              <FileText class="w-10 h-10" />
            </div>
            <div class="font-bold text-sm text-center truncate w-full px-2">{{ file.original_name }}</div>
            <div class="text-[10px] font-bold text-primary/30 mt-1 uppercase">{{ formatSize(file.size) }}</div>
            <button 
              class="absolute top-2 right-2 p-2 opacity-0 group-hover:opacity-100 transition-opacity text-primary"
              @click.stop="showFileMenu($event, file)"
            >
              <MoreVertical class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </section>

  <!-- 新建文件夹模态框 -->
  <div v-if="showNewFolderModal" class="fixed inset-0 z-[150] bg-[#0f0721]/60 backdrop-blur-md flex items-center justify-center p-6">
    <div class="w-full max-w-sm glass-solid p-8 rounded-[40px] border border-white/60 shadow-2xl flex flex-col animate-scale-in">
      <div class="w-14 h-14 bg-amber-100 text-amber-600 rounded-2xl flex items-center justify-center mb-6 shadow-inner">
        <FolderPlus class="w-8 h-8" />
      </div>
      <h3 class="text-2xl font-extrabold text-[#1e0a3d] mb-2">新建文件夹</h3>
      <p class="text-sm text-primary/50 font-bold mb-6">输入名称以创建新的存储目录</p>
      <input 
        v-model="newFolderName"
        type="text" 
        placeholder="输入文件夹名称..."
        class="w-full bg-white/60 border border-white/40 focus:bg-white focus:ring-4 focus:ring-primary/5 rounded-2xl px-6 py-4 outline-none transition-all font-bold mb-6"
        @keyup.enter="createFolder"
      >
      <div class="flex gap-3">
        <button 
          @click="showNewFolderModal = false"
          class="flex-1 py-4 bg-white border border-primary/10 text-primary font-bold rounded-2xl hover:bg-primary/5 transition-all"
        >
          取消
        </button>
        <button 
          @click="createFolder"
          :disabled="!newFolderName.trim() || creatingFolder"
          class="flex-1 py-4 bg-primary text-white font-bold rounded-2xl shadow-lg shadow-primary/20 hover:scale-105 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ creatingFolder ? '创建中...' : '创建' }}
        </button>
      </div>
    </div>
  </div>

  <!-- 下拉菜单 -->
  <div 
    v-if="showMenu" 
    class="fixed z-[120] bg-white/95 backdrop-blur-xl border border-primary/10 rounded-2xl shadow-2xl w-48 overflow-hidden py-1 animate-fade-in"
    :style="{ top: menuPosition.top + 'px', left: menuPosition.left + 'px' }"
  >
    <template v-if="menuTarget?.folder_id">
      <button 
        @click="renameFolder"
        class="w-full flex items-center gap-3 px-4 py-2.5 text-xs font-bold text-primary hover:bg-primary/5 transition-colors text-left"
      >
        <Edit3 class="w-4 h-4" />重命名
      </button>
      <button 
        @click="deleteFolder"
        class="w-full flex items-center gap-3 px-4 py-2.5 text-xs font-bold text-red-500 hover:bg-red-50 transition-colors text-left"
      >
        <Trash2 class="w-4 h-4" />删除
      </button>
    </template>
    <template v-else>
      <button 
        @click="copyFilehub"
        class="w-full flex items-center gap-3 px-4 py-2.5 text-xs font-bold text-primary hover:bg-primary/5 transition-colors text-left"
      >
        <Copy class="w-4 h-4" />复制 filehub://
      </button>
      <button 
        @click="copyShareLink"
        class="w-full flex items-center gap-3 px-4 py-2.5 text-xs font-bold text-primary hover:bg-primary/5 transition-colors text-left"
      >
        <Link class="w-4 h-4" />分享链接
      </button>
      <button 
        @click="showMoveFileDialogFn"
        class="w-full flex items-center gap-3 px-4 py-2.5 text-xs font-bold text-primary hover:bg-primary/5 transition-colors text-left"
      >
        <ArrowRight class="w-4 h-4" />移动到
      </button>
      <div class="h-[1px] bg-primary/5 my-1"></div>
      <button 
        @click="removeFile"
        class="w-full flex items-center gap-3 px-4 py-2.5 text-xs font-bold text-red-500 hover:bg-red-50 transition-colors text-left"
      >
        <Trash2 class="w-4 h-4" />删除
      </button>
    </template>
  </div>

  <!-- 移动文件对话框 -->
  <div 
    v-if="showMoveFileDialog" 
    class="fixed inset-0 z-[150] bg-[#0f0721]/60 backdrop-blur-md flex items-center justify-center p-6"
    @click.self="showMoveFileDialog = false"
  >
    <div class="w-full max-w-sm glass-solid p-8 rounded-[40px] border border-white/60 shadow-2xl flex flex-col animate-scale-in">
      <div class="w-14 h-14 bg-amber-100 text-amber-600 rounded-2xl flex items-center justify-center mb-6 shadow-inner">
        <Folder class="w-8 h-8" />
      </div>
      <h3 class="text-2xl font-extrabold text-[#1e0a3d] mb-2">移动文件</h3>
      <p class="text-sm text-primary/50 font-bold mb-6">{{ menuTarget?.original_name }}</p>
      
      <div class="max-h-60 overflow-y-auto space-y-2 mb-6">
        <button 
          @click="moveFileTo(null)"
          class="w-full flex items-center gap-3 px-4 py-3 rounded-xl hover:bg-primary/5 text-left transition-colors"
          :class="moveTargetFolderId === null ? 'bg-primary/10 text-primary' : 'text-primary'"
        >
          <Folder class="w-5 h-5" />根目录
        </button>
        <button 
          v-for="folder in allFolders" 
          :key="folder.folder_id"
          @click="moveFileTo(folder.folder_id)"
          class="w-full flex items-center gap-3 px-4 py-3 rounded-xl hover:bg-primary/5 text-left transition-colors"
          :class="moveTargetFolderId === folder.folder_id ? 'bg-primary/10 text-primary' : 'text-primary'"
        >
          <Folder class="w-5 h-5" :class="moveTargetFolderId === folder.folder_id ? 'fill-primary/20' : ''" />
          {{ folder.name }}
        </button>
      </div>
      
      <div class="flex gap-3">
        <button 
          @click="showMoveFileDialog = false"
          class="flex-1 py-4 bg-white border border-primary/10 text-primary font-bold rounded-2xl hover:bg-primary/5 transition-all"
        >
          取消
        </button>
        <button 
          @click="confirmMoveFile"
          :disabled="movingFile"
          class="flex-1 py-4 bg-primary text-white font-bold rounded-2xl shadow-lg shadow-primary/20 hover:scale-105 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ movingFile ? '移动中...' : '移动' }}
        </button>
      </div>
    </div>
  </div>

  <!-- Toast 通知 -->
  <div 
    id="toast" 
    class="fixed bottom-12 left-1/2 -translate-x-1/2 px-8 py-4 bg-[#1e0a3d] text-white rounded-2xl border border-white/10 shadow-2xl translate-y-20 opacity-0 transition-all duration-300 z-[200] font-extrabold text-sm pointer-events-none"
    :class="{ 'show': toast.show }"
  >
    {{ toast.message }}
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { 
  Folder, FolderPlus, Plus, List, LayoutGrid, ChevronRight, 
  MoreVertical, FileText, Download, FolderOpen,
  Edit3, Trash2, Copy, Link, ArrowRight
} from 'lucide-vue-next'
import { listFiles, deleteFile, downloadFile, shareFile } from '../api/files'
import { 
  listFolders, getFolderContents, createFolder as apiCreateFolder,
  renameFolder as apiRenameFolder, deleteFolder as apiDeleteFolder,
  moveFile
} from '../api/folders'
import { eventBus } from '../store/eventBus'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const folders = ref([])
const files = ref([])
const breadcrumbs = ref([])
const currentFolderInfo = ref(null)
const layout = ref('list')
const currentFolderId = computed(() => route.params.id || null)
const currentFolderName = computed(() => {
  if (!currentFolderId.value) return '所有文件'
  if (currentFolderInfo.value?.name) return currentFolderInfo.value.name
  return '文件夹'
})

// 上传链接，携带当前文件夹ID
const uploadLink = computed(() => {
  if (currentFolderId.value) {
    return `/upload?folder=${currentFolderId.value}`
  }
  return '/upload'
})

// 新建文件夹
const showNewFolderModal = ref(false)
const newFolderName = ref('')
const creatingFolder = ref(false)

// 下拉菜单
const showMenu = ref(false)
const menuPosition = ref({ top: 0, left: 0 })
const menuTarget = ref(null)

// 移动文件对话框
const showMoveFileDialog = ref(false)
const moveTargetFolderId = ref(null)
const allFolders = ref([])
const movingFile = ref(false)

// Toast
const toast = ref({ show: false, message: '' })

function showToast(message) {
  toast.value = { show: true, message }
  setTimeout(() => toast.value.show = false, 3000)
}

// 加载数据
async function loadData() {
  loading.value = true
  try {
    if (currentFolderId.value) {
      // 加载文件夹内容
      const res = await getFolderContents(currentFolderId.value)
      const data = res.data?.data
      currentFolderInfo.value = { name: data?.name }
      folders.value = data?.folders || []
      files.value = data?.files || []
      breadcrumbs.value = data?.breadcrumbs || []
    } else {
      // 加载根目录
      currentFolderInfo.value = null
      const [foldersRes, filesRes] = await Promise.all([
        listFolders(),
        listFiles({ limit: 200, offset: 0, order: 'desc' })
      ])
      folders.value = foldersRes.data?.data?.folders || []
      files.value = filesRes.data?.data?.files || []
      breadcrumbs.value = []
    }
  } catch (error) {
    console.error('加载失败:', error)
    showToast('加载失败')
  } finally {
    loading.value = false
  }
}

// 导航
function goToFolder(folderId) {
  if (folderId) {
    router.push(`/folder/${folderId}`)
  } else {
    router.push('/')
  }
}

function goToFile(fileId) {
  router.push(`/file/${fileId}`)
}

// 创建文件夹
async function createFolder() {
  if (!newFolderName.value.trim()) return
  creatingFolder.value = true
  try {
    await apiCreateFolder({
      name: newFolderName.value.trim(),
      parent_id: currentFolderId.value
    })
    showToast('文件夹创建成功')
    showNewFolderModal.value = false
    newFolderName.value = ''
    await loadData()
  } catch (error) {
    const msg = error.response?.data?.message || '创建失败'
    showToast(msg)
  } finally {
    creatingFolder.value = false
  }
}

// 显示菜单
function showFolderMenu(event, folder) {
  menuTarget.value = folder
  const rect = event.target.getBoundingClientRect()
  menuPosition.value = {
    top: rect.bottom + 10,
    left: rect.right - 192
  }
  showMenu.value = true
}

function showFileMenu(event, file) {
  menuTarget.value = file
  const rect = event.target.getBoundingClientRect()
  menuPosition.value = {
    top: rect.bottom + 10,
    left: rect.right - 192
  }
  showMenu.value = true
}

// 菜单操作
async function renameFolder() {
  const newName = prompt('请输入新名称:', menuTarget.value.name)
  if (!newName || newName === menuTarget.value.name) return
  try {
    await apiRenameFolder(menuTarget.value.folder_id, newName)
    showToast('重命名成功')
    await loadData()
  } catch (error) {
    showToast(error.response?.data?.message || '重命名失败')
  }
  showMenu.value = false
}

async function deleteFolder() {
  if (!confirm(`确定要删除文件夹 "${menuTarget.value.name}" 吗?\n注意：只能删除空文件夹。`)) return
  try {
    await apiDeleteFolder(menuTarget.value.folder_id)
    showToast('删除成功')
    await loadData()
  } catch (error) {
    showToast(error.response?.data?.message || '删除失败')
  }
  showMenu.value = false
}

async function copyFilehub() {
  const text = `filehub://${menuTarget.value.file_id}`
  try {
    await navigator.clipboard.writeText(text)
    showToast('已复制 filehub://')
  } catch {
    showToast('复制失败')
  }
  showMenu.value = false
}

async function copyShareLink() {
  try {
    const res = await shareFile(menuTarget.value.file_id)
    const url = res.data?.data?.url
    if (url) {
      await navigator.clipboard.writeText(url)
      showToast('已复制分享链接')
    }
  } catch {
    showToast('获取分享链接失败')
  }
  showMenu.value = false
}

async function removeFile() {
  if (!confirm(`确定要删除文件 "${menuTarget.value.original_name}" 吗?`)) return
  try {
    await deleteFile(menuTarget.value.file_id)
    showToast('删除成功')
    await loadData()
  } catch {
    showToast('删除失败')
  }
  showMenu.value = false
}

// 移动文件
async function loadAllFolders() {
  try {
    const res = await listFolders()
    allFolders.value = res.data?.data?.folders || []
  } catch (error) {
    console.error('加载文件夹列表失败:', error)
    allFolders.value = []
  }
}

function showMoveFileDialogFn() {
  if (!menuTarget.value?.file_id) return
  moveTargetFolderId.value = null
  loadAllFolders()
  showMoveFileDialog.value = true
  showMenu.value = false
}

function moveFileTo(folderId) {
  moveTargetFolderId.value = folderId
}

async function confirmMoveFile() {
  if (!menuTarget.value?.file_id) return
  movingFile.value = true
  try {
    await moveFile(menuTarget.value.file_id, moveTargetFolderId.value)
    showToast('文件移动成功')
    showMoveFileDialog.value = false
    await loadData()
  } catch (error) {
    showToast(error.response?.data?.message || '移动失败')
  } finally {
    movingFile.value = false
  }
}

async function handleDownload(file) {
  try {
    const res = await downloadFile(file.file_id)
    const blob = new Blob([res.data])
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = file.original_name
    link.click()
    window.URL.revokeObjectURL(url)
    showToast('开始下载')
  } catch {
    showToast('下载失败')
  }
}

// 工具函数
function formatSize(size) {
  if (!size) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let value = size
  let idx = 0
  while (value > 1024 && idx < units.length - 1) {
    value /= 1024
    idx++
  }
  return `${value.toFixed(1)} ${units[idx]}`
}

function formatDate(value) {
  return value?.replace('T', ' ').slice(0, 16) || '--'
}

function getFileIconClass(mimeType) {
  if (mimeType?.startsWith('image/')) return 'bg-blue-100 text-blue-600'
  if (mimeType?.startsWith('video/')) return 'bg-purple-100 text-purple-600'
  if (mimeType?.includes('pdf')) return 'bg-red-100 text-red-600'
  return 'bg-gray-100 text-gray-600'
}

// 监听路由变化
watch(() => route.params.id, () => {
  loadData()
})

// 监听刷新事件
watch(() => eventBus.refreshFiles, (newVal) => {
  if (newVal && !currentFolderId.value) {
    loadData()
  }
})

// 点击外部关闭菜单
window.addEventListener('click', (e) => {
  if (!e.target.closest('.dropdown-menu')) {
    showMenu.value = false
  }
})

onMounted(() => {
  loadData()
})
</script>
