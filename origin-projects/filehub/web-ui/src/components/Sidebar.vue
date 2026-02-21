<template>
  <!-- 桌面端侧边栏 -->
  <aside class="sidebar-desktop fixed inset-y-0 left-0 z-[60] w-72 lg:flex flex-col glass-solid border-r border-white/40 m-0 lg:m-4 lg:rounded-3xl shadow-2xl lg:shadow-xl overflow-hidden shrink-0 hidden">
    <div class="p-8 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 bg-gradient-to-br from-primary to-primary-light rounded-xl flex items-center justify-center text-white shadow-lg shadow-primary/30">
          <FolderKey class="w-6 h-6" />
        </div>
        <div>
          <h1 class="font-bold text-lg leading-tight text-[#1e0a3d]">FileHub</h1>
          <p class="text-[10px] text-primary/70 uppercase tracking-widest font-bold">Pro Control</p>
        </div>
      </div>
    </div>

    <nav class="flex-1 px-4 py-2 space-y-1">
      <div class="px-4 py-2 text-[10px] font-bold text-primary/40 uppercase tracking-[0.2em]">文件中心</div>
      <router-link 
        to="/" 
        class="nav-item w-full flex items-center gap-3 px-4 py-3 rounded-2xl transition-all duration-300"
        :class="{ 'active': $route.path === '/' || $route.path.startsWith('/folder') }"
      >
        <Files class="w-5 h-5" />
        <span class="font-bold text-sm">所有文件</span>
      </router-link>
      <router-link 
        to="/upload" 
        class="nav-item w-full flex items-center gap-3 px-4 py-3 rounded-2xl transition-all duration-300"
        :class="{ 'active': $route.path === '/upload' }"
      >
        <UploadCloud class="w-5 h-5" />
        <span class="font-bold text-sm">上传中心</span>
      </router-link>
      
      <div class="pt-4 px-4 py-2 text-[10px] font-bold text-primary/40 uppercase tracking-[0.2em]">活动追踪</div>
      <button class="nav-item w-full flex items-center gap-3 px-4 py-3 rounded-2xl transition-all duration-300 opacity-60 cursor-not-allowed">
        <DownloadCloud class="w-5 h-5" />
        <span class="font-bold text-sm">下载记录</span>
      </button>
    </nav>

    <div class="p-6">
      <div class="bg-primary/5 p-4 rounded-2xl border border-primary/10 mb-4">
        <div class="flex justify-between text-[11px] font-bold mb-2">
          <span class="text-primary/70">存储空间</span>
          <span class="text-primary">{{ usagePercent }}%</span>
        </div>
        <div class="h-1.5 bg-white/50 rounded-full overflow-hidden">
          <div class="h-full bg-primary rounded-full shadow-[0_0_8px_rgba(124,58,237,0.4)] transition-all duration-500" :style="{ width: usagePercent + '%' }"></div>
        </div>
        <p class="text-[10px] text-primary/60 font-medium mt-2">{{ formatSize(usedBytes) }} / {{ formatSize(capacityBytes) }} 已使用</p>
      </div>
      <button 
        @click="handleLogout"
        class="w-full flex items-center justify-center gap-2 px-4 py-3 bg-white hover:bg-red-50 rounded-2xl text-sm font-bold border border-white transition-all duration-300 text-red-500 shadow-sm"
      >
        <LogOut class="w-4 h-4" />
        <span>退出系统</span>
      </button>
    </div>
  </aside>

  <!-- 移动端侧边栏遮罩 -->
  <div 
    v-if="mobileMenuOpen" 
    class="fixed inset-0 bg-[#0f0721]/60 backdrop-blur-sm z-50 lg:hidden"
    @click="mobileMenuOpen = false"
  ></div>

  <!-- 移动端侧边栏 -->
  <aside 
    v-if="mobileMenuOpen"
    class="sidebar-mobile fixed inset-y-0 left-0 z-[60] w-72 flex flex-col glass-solid border-r border-white/40 shadow-2xl overflow-hidden shrink-0 lg:hidden"
  >
    <div class="p-6 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 bg-gradient-to-br from-primary to-primary-light rounded-xl flex items-center justify-center text-white shadow-lg shadow-primary/30">
          <FolderKey class="w-6 h-6" />
        </div>
        <div>
          <h1 class="font-bold text-lg leading-tight text-[#1e0a3d]">FileHub</h1>
        </div>
      </div>
      <button @click="mobileMenuOpen = false" class="p-2 text-primary/40">
        <X class="w-6 h-6" />
      </button>
    </div>

    <nav class="flex-1 px-4 py-2 space-y-1">
      <router-link 
        to="/" 
        class="nav-item w-full flex items-center gap-3 px-4 py-3 rounded-2xl transition-all duration-300"
        :class="{ 'active': $route.path === '/' || $route.path.startsWith('/folder') }"
        @click="mobileMenuOpen = false"
      >
        <Files class="w-5 h-5" />
        <span class="font-bold text-sm">所有文件</span>
      </router-link>
      <router-link 
        to="/upload" 
        class="nav-item w-full flex items-center gap-3 px-4 py-3 rounded-2xl transition-all duration-300"
        :class="{ 'active': $route.path === '/upload' }"
        @click="mobileMenuOpen = false"
      >
        <UploadCloud class="w-5 h-5" />
        <span class="font-bold text-sm">上传中心</span>
      </router-link>
    </nav>

    <div class="p-6">
      <button 
        @click="handleLogout"
        class="w-full flex items-center justify-center gap-2 px-4 py-3 bg-white hover:bg-red-50 rounded-2xl text-sm font-bold border border-white transition-all duration-300 text-red-500 shadow-sm"
      >
        <LogOut class="w-4 h-4" />
        <span>退出系统</span>
      </button>
    </div>
  </aside>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { FolderKey, Files, UploadCloud, DownloadCloud, LogOut, X } from 'lucide-vue-next'
import { clearTokens } from '../store/auth'

const router = useRouter()
const mobileMenuOpen = defineModel('mobileMenuOpen', { default: false })

const props = defineProps({
  usedBytes: { type: Number, default: 0 }
})

const capacityBytes = 50 * 1024 * 1024 * 1024 // 50GB

const usagePercent = computed(() => {
  return Math.min(100, Math.round((props.usedBytes / capacityBytes) * 100))
})

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

function handleLogout() {
  clearTokens()
  router.push('/login')
}
</script>
