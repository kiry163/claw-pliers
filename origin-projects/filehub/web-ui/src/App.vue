<template>
  <div class="flex h-screen overflow-hidden bg-bg">
    <!-- 背景光晕 -->
    <div class="fixed top-[-5%] left-[-5%] w-[35%] h-[35%] bg-primary/10 rounded-full blur-[120px] -z-10 pointer-events-none"></div>
    <div class="fixed bottom-[-5%] right-[-5%] w-[35%] h-[35%] bg-accent/5 rounded-full blur-[120px] -z-10 pointer-events-none"></div>

    <!-- 侧边栏 -->
    <Sidebar v-model:mobileMenuOpen="mobileMenuOpen" :used-bytes="usedBytes" />

    <!-- 主内容区域 -->
    <main v-if="!isAuthRoute" class="flex-1 flex flex-col min-w-0 overflow-hidden relative lg:ml-80">
      <!-- 顶部栏 -->
      <header class="h-20 flex items-center justify-between px-6 lg:px-8 shrink-0 bg-white/40 backdrop-blur-lg border-b border-white/30 z-40 lg:rounded-tl-none">
        <div class="flex items-center gap-4">
          <button 
            class="lg:hidden p-2 text-primary" 
            @click="mobileMenuOpen = true"
          >
            <Menu class="w-6 h-6" />
          </button>
          <div class="hidden lg:block relative group w-64 xl:w-96">
            <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-primary/50 group-focus-within:text-primary transition-colors" />
            <input 
              type="text" 
              placeholder="搜索文件或路径..."
              v-model="searchKeyword"
              @keyup.enter="handleSearch"
              class="w-full bg-white/60 border border-white/60 focus:bg-white focus:border-primary/40 focus:ring-4 focus:ring-primary/5 rounded-2xl py-2.5 pl-11 pr-4 outline-none transition-all duration-300 text-sm"
            >
          </div>
        </div>

        <div class="flex items-center gap-2 md:gap-4">
          <button 
            class="flex items-center gap-2 px-3 py-2 md:px-4 md:py-2.5 bg-primary/5 hover:bg-primary/10 text-primary border border-primary/10 rounded-2xl font-bold text-xs md:text-sm transition-all shadow-sm group"
            @click="showTaskDrawer = true"
          >
            <RefreshCw class="w-4 h-4 group-hover:animate-spin-slow" />
            <span class="hidden md:inline">{{ taskCount }} 个任务</span>
            <span class="md:hidden">{{ taskCount }}</span>
          </button>
          <div class="w-[1px] h-8 bg-primary/10"></div>
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-primary/10 border-2 border-white flex items-center justify-center font-bold text-primary shadow-sm">
              {{ userInitials }}
            </div>
          </div>
        </div>
      </header>

      <!-- 视图容器 -->
      <div class="flex-1 overflow-y-auto px-4 md:px-8 pb-32 lg:pb-8 pt-6 relative w-full">
        <router-view />
      </div>
    </main>

    <!-- 登录页 -->
    <router-view v-else />

    <!-- 任务抽屉 -->
    <TaskDrawer v-model:visible="showTaskDrawer" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Menu, Search, RefreshCw } from 'lucide-vue-next'
import Sidebar from './components/Sidebar.vue'
import TaskDrawer from './components/TaskDrawer.vue'

const route = useRoute()
const router = useRouter()

const mobileMenuOpen = ref(false)
const showTaskDrawer = ref(false)
const searchKeyword = ref('')
const usedBytes = ref(0)
const taskCount = ref(0)

const isAuthRoute = computed(() => route.path === '/login')
const userInitials = computed(() => 'AD') // 从用户信息获取

function handleSearch() {
  if (searchKeyword.value.trim()) {
    // 触发搜索，可以在当前页面搜索或跳转到搜索结果页
    console.log('搜索:', searchKeyword.value)
  }
}
</script>
