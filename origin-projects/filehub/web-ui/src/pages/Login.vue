<template>
  <section class="fixed inset-0 z-[100] bg-[#FAF5FF] flex items-center justify-center p-6">
    <!-- 背景光晕 -->
    <div class="fixed top-[-10%] left-[-10%] w-[40%] h-[40%] bg-primary/10 rounded-full blur-[150px] pointer-events-none"></div>
    <div class="fixed bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-accent/5 rounded-full blur-[150px] pointer-events-none"></div>

    <div class="w-full max-w-md glass-solid p-10 md:p-12 rounded-[48px] border border-white/60 shadow-2xl flex flex-col items-center animate-scale-in relative z-10">
      <!-- Logo -->
      <div class="w-20 h-20 bg-gradient-to-br from-primary to-primary-light rounded-3xl flex items-center justify-center text-white shadow-2xl mb-8">
        <FolderKey class="w-10 h-10" />
      </div>
      
      <!-- 标题 -->
      <h2 class="text-4xl font-extrabold text-center tracking-tight text-[#1e0a3d]">FileHub Pro</h2>
      <p class="text-primary/60 text-center font-bold mt-2 mb-10 uppercase tracking-widest text-xs">Secure Control Center</p>
      
      <!-- 表单 -->
      <form @submit.prevent="submit" class="w-full space-y-5">
        <div class="space-y-2">
          <label class="text-[11px] font-bold text-primary/40 ml-3 uppercase tracking-widest">用户名</label>
          <input 
            v-model="username" 
            type="text" 
            placeholder="admin"
            autocomplete="username"
            class="w-full bg-white/60 border border-white/40 focus:bg-white focus:ring-4 focus:ring-primary/5 rounded-2xl px-6 py-4 outline-none transition-all font-bold"
          >
        </div>
        
        <div class="space-y-2">
          <label class="text-[11px] font-bold text-primary/40 ml-3 uppercase tracking-widest">密码</label>
          <input 
            v-model="password" 
            type="password" 
            placeholder="••••••••"
            autocomplete="current-password"
            class="w-full bg-white/60 border border-white/40 focus:bg-white focus:ring-4 focus:ring-primary/5 rounded-2xl px-6 py-4 outline-none transition-all font-bold"
          >
        </div>
        
        <button 
          type="submit"
          :disabled="loading"
          class="w-full py-5 bg-primary text-white font-extrabold rounded-2xl shadow-xl shadow-primary/30 hover:scale-[1.02] active:scale-95 transition-all mt-6 text-sm uppercase tracking-widest disabled:opacity-70 disabled:cursor-not-allowed"
        >
          {{ loading ? '登录中...' : '进入系统' }}
        </button>
      </form>
      
      <p class="text-[10px] text-primary/40 font-bold mt-8 uppercase tracking-widest">Secure private storage</p>
    </div>

    <!-- Toast -->
    <div 
      class="fixed bottom-12 left-1/2 -translate-x-1/2 px-8 py-4 bg-[#1e0a3d] text-white rounded-2xl border border-white/10 shadow-2xl transition-all duration-300 z-[200] font-extrabold text-sm"
      :class="toast.show ? 'translate-y-0 opacity-100' : 'translate-y-20 opacity-0 pointer-events-none'"
    >
      {{ toast.message }}
    </div>
  </section>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { FolderKey } from 'lucide-vue-next'
import { login } from '../api/auth'
import { setTokens } from '../store/auth'

const username = ref('admin')
const password = ref('')
const loading = ref(false)
const route = useRoute()
const router = useRouter()
const toast = ref({ show: false, message: '' })

function showToast(message) {
  toast.value = { show: true, message }
  setTimeout(() => toast.value.show = false, 3000)
}

const submit = async () => {
  if (!username.value || !password.value) {
    showToast('请输入用户名和密码')
    return
  }
  
  loading.value = true
  try {
    const response = await login(username.value, password.value)
    const data = response.data?.data
    if (!data?.access_token) throw new Error('login failed')
    setTokens(data.access_token, data.refresh_token)
    router.push(route.query.redirect || '/')
  } catch {
    showToast('登录失败，请检查用户名和密码')
  } finally {
    loading.value = false
  }
}
</script>
