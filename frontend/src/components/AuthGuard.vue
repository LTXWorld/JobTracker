<template>
  <div class="auth-guard">
    <!-- 加载状态 -->
    <div v-if="loading" class="auth-loading">
      <a-spin size="large">
        <template #tip>验证登录状态...</template>
      </a-spin>
    </div>
    
    <!-- 已认证，显示内容 -->
    <slot v-else-if="authStore.isLoggedIn" />
    
    <!-- 未认证，显示提示 -->
    <div v-else class="auth-required">
      <a-result
        status="warning"
        title="需要登录"
        sub-title="请先登录后访问此页面"
      >
        <template #extra>
          <a-space>
            <a-button type="primary" @click="goToLogin">
              立即登录
            </a-button>
            <a-button @click="goToRegister">
              注册账号
            </a-button>
          </a-space>
        </template>
      </a-result>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(true)

onMounted(async () => {
  // 初始化认证状态
  authStore.initAuth()
  
  // 如果有token，验证其有效性
  if (authStore.accessToken) {
    try {
      await authStore.validateToken()
    } catch (error) {
      console.error('Token validation failed:', error)
    }
  }
  
  loading.value = false
})

const goToLogin = () => {
  router.push('/login')
}

const goToRegister = () => {
  router.push('/register')
}
</script>

<style scoped>
.auth-loading {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
}

.auth-required {
  padding: 48px 24px;
}
</style>