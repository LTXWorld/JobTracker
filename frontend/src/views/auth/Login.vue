<template>
  <div class="login-container">
    <!-- 背景装饰 -->
    <div class="login-background">
      <div class="bg-shape shape-1"></div>
      <div class="bg-shape shape-2"></div>
      <div class="bg-shape shape-3"></div>
    </div>

    <div class="login-content">
      <!-- Logo和标题区域 -->
      <div class="login-header">
        <div class="logo">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <rect x="3" y="4" width="18" height="18" rx="2" stroke="#1890ff" stroke-width="2" fill="none"/>
            <path d="M16 10l-4 4-2-2" stroke="#1890ff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <h1 class="title">JobView</h1>
        <p class="subtitle">求职进程管理系统</p>
      </div>

      <!-- 登录表单 -->
      <div class="login-form-wrapper">
        <a-form
          :model="formData"
          :rules="formRules"
          @finish="handleLogin"
          @finish-failed="handleLoginFailed"
          layout="vertical"
          size="large"
        >
          <a-form-item name="username" label="用户名或邮箱">
            <a-input
              v-model:value="formData.username"
              placeholder="请输入用户名或邮箱"
              allow-clear
              autocomplete="username"
            >
              <template #prefix>
                <UserOutlined style="color: rgba(0,0,0,.25)" />
              </template>
            </a-input>
          </a-form-item>

          <a-form-item name="password" label="密码">
            <a-input-password
              v-model:value="formData.password"
              placeholder="请输入密码"
              autocomplete="current-password"
            >
              <template #prefix>
                <LockOutlined style="color: rgba(0,0,0,.25)" />
              </template>
            </a-input-password>
          </a-form-item>

          <a-form-item>
            <div class="login-options">
              <a-checkbox v-model:checked="rememberMe">记住登录状态</a-checkbox>
              <a href="#" class="forgot-password" @click.prevent="handleForgotPassword">
                忘记密码？
              </a>
            </div>
          </a-form-item>

          <a-form-item>
            <a-button
              type="primary"
              html-type="submit"
              :loading="authStore.loading"
              :disabled="isLocked"
              :block="true"
              size="large"
              class="login-button"
            >
              {{ isLocked ? `锁定中 (${lockoutTime}s)` : '登录' }}
            </a-button>
          </a-form-item>

          <a-form-item class="register-link">
            <div class="text-center">
              <span>还没有账号？</span>
              <router-link to="/register" class="register-link-btn">立即注册</router-link>
            </div>
          </a-form-item>
        </a-form>
      </div>

      <!-- 快速登录提示 -->
      <div class="quick-login-tip" v-if="isDevelopment">
        <a-alert
          message="开发环境快速登录"
          description="用户名: testuser，密码: TestPass123!"
          type="info"
          show-icon
          :closable="true"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import { UserOutlined, LockOutlined } from '@ant-design/icons-vue'
import { useAuthStore } from '../../stores/auth'
import type { LoginCredentials } from '../../types/auth'
import type { Rule } from 'ant-design-vue/es/form'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

// 检查是否为开发环境
const isDevelopment = ref(import.meta.env.DEV)

// 表单数据
const formData = reactive<LoginCredentials>({
  username: '',
  password: ''
})

// 记住登录状态
const rememberMe = ref(false)

// 登录尝试次数
const loginAttempts = ref(0)
const maxAttempts = 5
const lockoutTime = ref(0)
const isLocked = computed(() => lockoutTime.value > 0)

// 倒计时定时器
let lockoutTimer: NodeJS.Timeout | null = null

// 表单验证规则
const formRules: Record<string, Rule[]> = {
  username: [
    { required: true, message: '请输入用户名或邮箱', trigger: 'blur' },
    { min: 3, message: '用户名至少3个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6个字符', trigger: 'blur' }
  ]
}

// 处理登录
const handleLogin = async (values: LoginCredentials) => {
  if (isLocked.value) {
    message.error(`登录已被锁定，请等待 ${lockoutTime.value} 秒后再试`)
    return
  }

  try {
    // 添加记住我的参数
    const loginData = {
      ...values,
      remember_me: rememberMe.value
    }
    
    const success = await authStore.login(loginData)
    if (success) {
      // 登录成功，重置尝试次数
      loginAttempts.value = 0
      clearLockout()
      
      // 重定向到原来要访问的页面或首页
      const redirect = route.query.redirect as string || '/'
      await router.push(redirect)
    } else {
      // 登录失败，增加尝试次数
      handleFailedAttempt()
    }
  } catch (error) {
    console.error('登录过程出错:', error)
    handleFailedAttempt()
  }
}

// 处理登录失败尝试
const handleFailedAttempt = () => {
  loginAttempts.value++
  
  if (loginAttempts.value >= maxAttempts) {
    // 达到最大尝试次数，锁定账号
    const lockoutSeconds = 300 // 5分钟
    lockoutTime.value = lockoutSeconds
    message.error(`登录失败次数过多，账号已被锁定 ${lockoutSeconds} 秒`)
    
    // 启动倒计时
    startLockoutTimer()
  } else {
    const remainingAttempts = maxAttempts - loginAttempts.value
    message.warning(`登录失败，还可尝试 ${remainingAttempts} 次`)
  }
}

// 启动锁定倒计时
const startLockoutTimer = () => {
  if (lockoutTimer) {
    clearInterval(lockoutTimer)
  }
  
  lockoutTimer = setInterval(() => {
    lockoutTime.value--
    if (lockoutTime.value <= 0) {
      clearLockout()
    }
  }, 1000)
}

// 清除锁定状态
const clearLockout = () => {
  if (lockoutTimer) {
    clearInterval(lockoutTimer)
    lockoutTimer = null
  }
  lockoutTime.value = 0
}

// 处理登录失败
const handleLoginFailed = (errorInfo: any) => {
  console.error('表单验证失败:', errorInfo)
  message.error('请检查输入信息')
}

// 忘记密码处理
const handleForgotPassword = () => {
  message.info('忘记密码功能开发中，请联系管理员重置密码')
}

// 组件挂载时的处理
onMounted(() => {
  // 如果已经登录，直接跳转到首页
  if (authStore.isLoggedIn) {
    router.push('/')
  }
  
  // 检查是否有登录错误信息
  if (route.query.error === 'unauthorized') {
    message.warning('登录已过期，请重新登录')
  }
  
  // 恢复记住我的状态
  const rememberedUser = localStorage.getItem('remembered_username')
  if (rememberedUser) {
    formData.username = rememberedUser
    rememberMe.value = true
  }
})
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  position: relative;
  overflow: hidden;
}

.login-background {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
  z-index: 0;
}

.bg-shape {
  position: absolute;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  animation: float 6s ease-in-out infinite;
}

.shape-1 {
  width: 200px;
  height: 200px;
  top: 20%;
  left: 10%;
  animation-delay: 0s;
}

.shape-2 {
  width: 150px;
  height: 150px;
  top: 60%;
  right: 15%;
  animation-delay: 2s;
}

.shape-3 {
  width: 100px;
  height: 100px;
  bottom: 20%;
  left: 70%;
  animation-delay: 4s;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0px);
  }
  50% {
    transform: translateY(-20px);
  }
}

.login-content {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 400px;
  padding: 0 24px;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.logo {
  margin-bottom: 16px;
}

.title {
  color: white;
  font-size: 32px;
  font-weight: bold;
  margin: 0 0 8px 0;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.subtitle {
  color: rgba(255, 255, 255, 0.8);
  font-size: 16px;
  margin: 0;
}

.login-form-wrapper {
  background: white;
  padding: 32px;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.login-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.forgot-password {
  color: #1890ff;
  text-decoration: none;
  font-size: 14px;
}

.forgot-password:hover {
  color: #40a9ff;
}

.login-button {
  height: 48px;
  border-radius: 8px;
  font-weight: 500;
  font-size: 16px;
  background: linear-gradient(135deg, #1890ff 0%, #096dd9 100%);
  border: none;
}

.login-button:hover {
  background: linear-gradient(135deg, #40a9ff 0%, #1890ff 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(24, 144, 255, 0.3);
}

.register-link {
  text-align: center;
  margin-bottom: 0;
}

.register-link-btn {
  color: #1890ff;
  text-decoration: none;
  font-weight: 500;
  margin-left: 8px;
}

.register-link-btn:hover {
  color: #40a9ff;
}

.text-center {
  text-align: center;
}

.quick-login-tip {
  margin-top: 24px;
}

/* 响应式设计 */
@media (max-width: 480px) {
  .login-content {
    max-width: 100%;
    padding: 0 16px;
  }

  .login-form-wrapper {
    padding: 24px 20px;
  }

  .title {
    font-size: 28px;
  }

  .subtitle {
    font-size: 14px;
  }

  .login-options {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
}

/* 深色模式适配 */
@media (prefers-color-scheme: dark) {
  .login-form-wrapper {
    background: rgba(255, 255, 255, 0.95);
  }
}
</style>