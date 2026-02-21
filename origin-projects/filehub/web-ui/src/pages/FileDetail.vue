<template>
  <section class="space-y-6">
    <!-- 头部 -->
    <div class="flex items-center gap-4">
      <button 
        @click="goBack"
        class="p-2.5 rounded-xl hover:bg-white text-primary border border-white/50 shadow-sm transition-all"
      >
        <ArrowLeft class="w-6 h-6" />
      </button>
      <h2 class="text-3xl font-extrabold tracking-tight">文件详情</h2>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="flex items-center justify-center py-20">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
    </div>

    <!-- 内容区域 -->
    <div v-else class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- 左侧预览 -->
      <div class="lg:col-span-2 space-y-4">
        <div class="bg-white/60 border border-white/40 rounded-3xl p-6 shadow-sm">
          <!-- 文件信息头部 -->
          <div class="flex items-start justify-between mb-6">
            <div class="flex items-center gap-4">
              <div class="w-14 h-14 rounded-2xl flex items-center justify-center shrink-0 shadow-inner"
                :class="getFileIconClass(file?.mime_type)">
                <FileText class="w-8 h-8" />
              </div>
              <div>
                <h3 class="font-extrabold text-lg text-[#1e0a3d]">{{ file?.original_name }}</h3>
                <p class="text-sm text-primary/50 font-bold mt-1">{{ metaLine }}</p>
              </div>
            </div>
            <span class="px-3 py-1 bg-primary/10 text-primary text-xs font-bold rounded-full">
              {{ previewLabel }}
            </span>
          </div>

          <!-- 预览区域 -->
          <div class="min-h-[400px] bg-primary/5 rounded-2xl flex items-center justify-center overflow-hidden">
            <!-- 文件过大提示 -->
            <div v-if="isFileTooLarge" class="text-center py-20">
              <FileX class="w-16 h-16 text-primary/30 mx-auto mb-4" />
              <p class="text-primary/50 font-bold">文件过大，无法在线预览</p>
              <p class="text-sm text-primary/30 mt-2">请下载后查看</p>
            </div>

            <!-- 图片预览 -->
            <div v-else-if="previewType === 'image'" class="w-full h-full">
              <img v-if="previewUrl" :src="previewUrl" class="w-full h-full object-contain" :alt="file?.original_name" />
              <div v-else class="text-primary/40 font-bold">图片加载中...</div>
            </div>

            <!-- 视频预览 -->
            <div v-else-if="previewType === 'video'" class="w-full">
              <video v-if="previewUrl" controls class="w-full rounded-xl" :src="previewUrl"></video>
              <div v-else class="text-primary/40 font-bold">视频加载中...</div>
            </div>

            <!-- 文本预览 -->
            <div v-else-if="previewType === 'text'" class="w-full p-6 overflow-auto max-h-[600px]">
              <pre class="whitespace-pre-wrap text-sm font-mono text-primary/80">{{ textContent }}</pre>
            </div>

            <!-- Markdown预览 -->
            <div v-else-if="previewType === 'markdown'" class="w-full p-6">
              <article class="prose prose-sm max-w-none" v-html="markdownHtml"></article>
            </div>

            <!-- 不支持预览 -->
            <div v-else class="text-center py-20">
              <FileX class="w-16 h-16 text-primary/30 mx-auto mb-4" />
              <p class="text-primary/50 font-bold">不支持预览此文件类型</p>
              <p class="text-sm text-primary/30 mt-2">请下载后查看</p>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧操作 -->
      <div class="space-y-4">
        <div class="bg-white/60 border border-white/40 rounded-3xl p-6 shadow-sm">
          <h4 class="font-extrabold text-[#1e0a3d] mb-4">操作</h4>
          
          <div class="space-y-3">
            <button 
              @click="download"
              class="w-full py-3 bg-primary text-white font-bold rounded-2xl shadow-lg shadow-primary/20 hover:scale-[1.02] transition-all flex items-center justify-center gap-2"
            >
              <Download class="w-5 h-5" />
              下载文件
            </button>
            
            <button 
              @click="copyFilehub"
              class="w-full py-3 bg-white border border-primary/20 text-primary font-bold rounded-2xl hover:bg-primary/5 transition-all flex items-center justify-center gap-2"
            >
              <Copy class="w-5 h-5" />
              复制 filehub://
            </button>
            
            <button 
              @click="copyShare"
              class="w-full py-3 bg-white border border-primary/20 text-primary font-bold rounded-2xl hover:bg-primary/5 transition-all flex items-center justify-center gap-2"
            >
              <Share2 class="w-5 h-5" />
              复制分享链接
            </button>
            
            <div class="h-[1px] bg-primary/10 my-4"></div>
            
            <button 
              @click="remove"
              class="w-full py-3 bg-red-50 border border-red-200 text-red-500 font-bold rounded-2xl hover:bg-red-100 transition-all flex items-center justify-center gap-2"
            >
              <Trash2 class="w-5 h-5" />
              删除文件
            </button>
          </div>

          <!-- 文件信息 -->
          <div class="mt-6 pt-6 border-t border-primary/10 space-y-2">
            <div class="flex justify-between text-sm">
              <span class="text-primary/50 font-bold">文件类型</span>
              <span class="font-bold text-[#1e0a3d]">{{ file?.mime_type || '-' }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-primary/50 font-bold">文件大小</span>
              <span class="font-bold text-[#1e0a3d]">{{ formatSize(file?.size) }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-primary/50 font-bold">上传时间</span>
              <span class="font-bold text-[#1e0a3d]">{{ formatDate(file?.created_at) }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-primary/50 font-bold">文件ID</span>
              <span class="font-mono font-bold text-primary">{{ file?.file_id }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>

  <!-- 删除确认对话框 -->
  <div v-if="showDeleteDialog" class="fixed inset-0 z-[150] bg-[#0f0721]/60 backdrop-blur-md flex items-center justify-center p-6">
    <div class="w-full max-w-sm glass-solid p-8 rounded-[40px] border border-white/60 shadow-2xl flex flex-col animate-scale-in">
      <div class="w-14 h-14 bg-red-100 text-red-500 rounded-2xl flex items-center justify-center mb-6">
        <Trash2 class="w-8 h-8" />
      </div>
      <h3 class="text-2xl font-extrabold text-[#1e0a3d] mb-2">确认删除</h3>
      <p class="text-sm text-primary/50 font-bold mb-6">确定要删除文件 "{{ file?.original_name }}" 吗？此操作不可恢复。</p>
      <div class="flex gap-3">
        <button 
          @click="showDeleteDialog = false"
          class="flex-1 py-4 bg-white border border-primary/10 text-primary font-bold rounded-2xl hover:bg-primary/5 transition-all"
        >
          取消
        </button>
        <button 
          @click="confirmDelete"
          class="flex-1 py-4 bg-red-500 text-white font-bold rounded-2xl hover:scale-105 transition-all shadow-lg shadow-red-500/20"
        >
          删除
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
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import MarkdownIt from 'markdown-it'
import { 
  ArrowLeft, FileText, FileX, Download, Copy, Share2, Trash2 
} from 'lucide-vue-next'
import { getFile, downloadFile, deleteFile, getPreviewUrl, shareFile } from '../api/files'

const route = useRoute()
const router = useRouter()
const file = ref(null)
const loading = ref(false)
const markdownHtml = ref('')
const markdown = new MarkdownIt({ html: false, linkify: true })
const previewUrl = ref('')
const textContent = ref('')
const copyDialogOpen = ref(false)
const copyDialogText = ref('')
const showDeleteDialog = ref(false)
const toast = ref({ show: false, message: '' })

function showToast(message) {
  toast.value = { show: true, message }
  setTimeout(() => toast.value.show = false, 3000)
}

const metaLine = computed(() => {
  if (!file.value) return ''
  return `${formatSize(file.value.size)} · ${formatDate(file.value.created_at)}`
})

const previewType = computed(() => {
  if (!file.value) return 'none'
  const name = file.value.original_name?.toLowerCase() || ''
  const mime = file.value.mime_type || ''
  if (mime.startsWith('image/')) return 'image'
  if (mime.startsWith('video/')) return 'video'
  if (mime.includes('markdown') || name.endsWith('.md') || name.endsWith('.markdown')) return 'markdown'
  if (name.endsWith('.txt') || mime === 'text/plain') return 'text'
  if (mime.startsWith('text/')) return 'text'
  return 'none'
})

const previewLabel = computed(() => {
  if (previewType.value === 'image') return '图片'
  if (previewType.value === 'video') return '视频'
  if (previewType.value === 'markdown') return 'Markdown'
  if (previewType.value === 'text') return '文本'
  return '不支持预览'
})

// 大文件限制（5MB）
const MAX_PREVIEW_SIZE = 5 * 1024 * 1024
const isFileTooLarge = computed(() => {
  if (!file.value?.size) return false
  return file.value.size > MAX_PREVIEW_SIZE
})

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

const formatDate = (value) => value?.replace('T', ' ').slice(0, 16) || '--'

const getFileIconClass = (mimeType) => {
  if (mimeType?.startsWith('image/')) return 'bg-blue-100 text-blue-600'
  if (mimeType?.startsWith('video/')) return 'bg-purple-100 text-purple-600'
  if (mimeType?.includes('pdf')) return 'bg-red-100 text-red-600'
  return 'bg-gray-100 text-gray-600'
}

const fetchFile = async () => {
  loading.value = true
  try {
    const response = await getFile(route.params.id)
    file.value = response.data?.data
    if (!file.value) throw new Error('not found')
    await loadPreview()
  } catch (error) {
    showToast('文件不存在或加载失败')
    console.error('加载文件失败:', error)
  } finally {
    loading.value = false
  }
}

const loadPreview = async () => {
  if (!file.value?.file_id) return
  if (isFileTooLarge.value) return
  
  if (previewType.value === 'text') {
    try {
      const response = await downloadFile(file.value.file_id)
      let text = await response.data.text()
      // 限制显示内容长度
      const maxLength = 500 * 1024 // 只显示前 500KB
      if (text.length > maxLength) {
        text = text.substring(0, maxLength) + '\n\n... (文件过大，仅显示部分内容)'
      }
      textContent.value = text
    } catch (error) {
      console.error('加载文本失败:', error)
    }
    return
  }
  
  if (previewType.value === 'markdown') {
    try {
      const response = await downloadFile(file.value.file_id)
      const text = await response.data.text()
      markdownHtml.value = markdown.render(text)
    } catch (error) {
      console.error('加载Markdown失败:', error)
    }
    return
  }
  
  if (previewType.value === 'image' || previewType.value === 'video') {
    try {
      const response = await getPreviewUrl(file.value.file_id)
      previewUrl.value = response.data?.data?.url || ''
    } catch (error) {
      console.error('获取预览URL失败:', error)
    }
  }
}

const copyFilehub = async () => {
  const text = `filehub://${file.value?.file_id}`
  try {
    await navigator.clipboard.writeText(text)
    showToast('已复制 filehub://')
  } catch {
    copyDialogText.value = text
    copyDialogOpen.value = true
  }
}

const copyShare = async () => {
  try {
    const response = await shareFile(file.value?.file_id)
    const url = response.data?.data?.url
    if (!url) throw new Error('share failed')
    try {
      await navigator.clipboard.writeText(url)
      showToast('已复制分享链接')
    } catch {
      copyDialogText.value = url
      copyDialogOpen.value = true
    }
  } catch (error) {
    showToast('分享链接生成失败')
    console.error('获取分享链接失败:', error)
  }
}

const download = async () => {
  if (!file.value?.file_id) return
  try {
    const response = await downloadFile(file.value.file_id)
    const blob = new Blob([response.data])
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = file.value.original_name
    link.click()
    window.URL.revokeObjectURL(url)
    showToast('开始下载')
  } catch (error) {
    showToast('下载失败')
    console.error('下载失败:', error)
  }
}

const remove = () => {
  showDeleteDialog.value = true
}

const confirmDelete = async () => {
  try {
    await deleteFile(file.value.file_id)
    showToast('已删除')
    showDeleteDialog.value = false
    router.push('/')
  } catch (error) {
    showToast('删除失败')
    console.error('删除失败:', error)
  }
}

const goBack = () => router.push('/')

onMounted(fetchFile)
</script>
