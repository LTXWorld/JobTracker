<template>
  <div class="register-container">
    <!-- 背景装饰 -->
    <div class="register-background">
      <div class="bg-shape shape-1"></div>
      <div class="bg-shape shape-2"></div>
      <div class="bg-shape shape-3"></div>
      <div class="bg-shape shape-4"></div>
    </div>

    <div class="register-content">
      <!-- Logo和标题区域 -->
      <div class="register-header">
        <div class="logo">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <rect x="3" y="4" width="18" height="18" rx="2" stroke="#1890ff" stroke-width="2" fill="none"/>
            <path d="M16 10l-4 4-2-2" stroke="#1890ff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <h1 class="title">加入 JobView</h1>
        <p class="subtitle">开始管理您的求职进程</p>
      </div>

      <!-- 注册表单 -->
      <div class="register-form-wrapper">
        <a-form
          :model="formData"
          :rules="formRules"
          @finish="handleRegister"
          @finish-failed="handleRegisterFailed"
          layout="vertical"
          size="large"
        >
          <a-form-item name="username" label="用户名">
            <a-input
              v-model:value="formData.username"
              placeholder="请输入用户名 (3-20字符)"
              allow-clear
              autocomplete="username"
              @input="checkUsernameAvailability"
            >
              <template #prefix>
                <UserOutlined style="color: rgba(0,0,0,.25)" />
              </template>
              <template #suffix v-if="usernameChecking">
                <LoadingOutlined style="color: #1890ff" />
              </template>
              <template #suffix v-else-if="usernameStatus === 'success'">
                <CheckCircleOutlined style="color: #52c41a" />
              </template>
              <template #suffix v-else-if="usernameStatus === 'error'">
                <CloseCircleOutlined style="color: #ff4d4f" />
              </template>
            </a-input>
          </a-form-item>

          <a-form-item name="email" label="邮箱">
            <a-input
              v-model:value="formData.email"
              type="email"
              placeholder="请输入邮箱地址"
              allow-clear
              autocomplete="email"
              @input="checkEmailAvailability"
            >
              <template #prefix>
                <MailOutlined style="color: rgba(0,0,0,.25)" />
              </template>
              <template #suffix v-if="emailChecking">
                <LoadingOutlined style="color: #1890ff" />
              </template>
              <template #suffix v-else-if="emailStatus === 'success'">
                <CheckCircleOutlined style="color: #52c41a" />
              </template>
              <template #suffix v-else-if="emailStatus === 'error'">
                <CloseCircleOutlined style="color: #ff4d4f" />
              </template>
            </a-input>
          </a-form-item>

          <a-form-item name="password" label="密码">
            <a-input-password
              v-model:value="formData.password"
              placeholder="请输入密码"
              autocomplete="new-password"
              @input="checkPasswordStrength"
            >
              <template #prefix>
                <LockOutlined style="color: rgba(0,0,0,.25)" />
              </template>
            </a-input-password>
            <!-- 密码强度指示器 -->
            <div class="password-strength" v-if="formData.password">
              <div class="strength-bar">
                <div 
                  class="strength-fill"
                  :class="passwordStrengthClass"
                  :style="{ width: passwordStrengthWidth }"
                ></div>
              </div>
              <span class="strength-text" :class="passwordStrengthClass">
                {{ passwordStrengthText }}
              </span>
            </div>
          </a-form-item>

          <a-form-item name="confirmPassword" label="确认密码">
            <a-input-password
              v-model:value="formData.confirmPassword"
              placeholder="请再次输入密码"
              autocomplete="new-password"
            >
              <template #prefix>
                <LockOutlined style="color: rgba(0,0,0,.25)" />
              </template>
            </a-input-password>
          </a-form-item>

          <!-- 服务条款 -->
          <a-form-item name="agreement">
            <a-checkbox v-model:checked="agreement">
              我已阅读并同意
              <a href="#" @click.prevent="showTerms">《服务条款》</a>
              和
              <a href="#" @click.prevent="showPrivacy">《隐私政策》</a>
            </a-checkbox>
          </a-form-item>

          <a-form-item>
            <a-button
              type="primary"
              html-type="submit"
              :loading="authStore.loading"
              :block="true"
              size="large"
              class="register-button"
            >
              立即注册
            </a-button>
          </a-form-item>

          <a-form-item class="login-link">
            <div class="text-center">
              <span>已有账号？</span>
              <router-link to="/login" class="login-link-btn">立即登录</router-link>
            </div>
          </a-form-item>
        </a-form>
      </div>
    </div>

    <!-- 服务条款模态框 -->
    <a-modal
      v-model:open="termsVisible"
      title="服务条款"
      :footer="null"
      width="600px"
    >
      <div class="terms-content">
        <p>欢迎使用 JobView 求职进程管理系统。在使用我们的服务之前，请仔细阅读以下服务条款：</p>
        <h4>1. 服务说明</h4>
        <p>JobView 是一个帮助用户管理求职进程的工具，提供求职记录管理、面试跟踪等功能。</p>
        <h4>2. 用户责任</h4>
        <p>用户需要保证注册信息的真实性，并妥善保管账户信息。</p>
        <h4>3. 隐私保护</h4>
        <p>我们承诺保护用户隐私，不会泄露用户的个人信息和求职数据。</p>
        <h4>4. 免责声明</h4>
        <p>本系统仅为工具性质，不对求职结果承担任何责任。</p>
      </div>
      <template #footer>
        <a-button type="primary" @click="termsVisible = false">我已了解</a-button>
      </template>
    </a-modal>

    <!-- 隐私政策模态框 -->
    <a-modal
      v-model:open="privacyVisible"
      title="隐私政策"
      :footer="null"
      width="600px"
    >
      <div class="privacy-content">
        <p>我们重视您的隐私保护，本政策说明我们如何收集、使用和保护您的个人信息：</p>
        <h4>1. 信息收集</h4>
        <p>我们只收集您主动提供的信息，包括用户名、邮箱和求职相关数据。</p>
        <h4>2. 信息使用</h4>
        <p>收集的信息仅用于提供服务功能，不会用于其他商业目的。</p>
        <h4>3. 信息保护</h4>
        <p>我们采用安全技术保护您的数据，包括数据加密和访问控制。</p>
        <h4>4. 信息共享</h4>
        <p>我们不会与第三方共享您的个人信息，除非得到您的明确同意。</p>
      </div>
      <template #footer>
        <a-button type="primary" @click="privacyVisible = false">我已了解</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { 
  UserOutlined, 
  LockOutlined, 
  MailOutlined,
  LoadingOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined
} from '@ant-design/icons-vue'
import { useAuthStore } from '../../stores/auth'
import { AuthAPI } from '../../api/auth'
import type { RegisterData } from '../../types/auth'
import type { Rule } from 'ant-design-vue/es/form'

const router = useRouter()
const authStore = useAuthStore()

// 表单数据
const formData = reactive<RegisterData>({
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})

// 服务条款同意状态
const agreement = ref(false)

// 用户名和邮箱检查状态
const usernameChecking = ref(false)
const emailChecking = ref(false)
const usernameStatus = ref<'success' | 'error' | ''>('')
const emailStatus = ref<'success' | 'error' | ''>('')

// 密码强度
const passwordStrength = ref(0)

// 模态框状态
const termsVisible = ref(false)
const privacyVisible = ref(false)

// 防抖定时器
let usernameDebounceTimer: ReturnType<typeof setTimeout> | null = null
let emailDebounceTimer: ReturnType<typeof setTimeout> | null = null

// 清理定时器
onUnmounted(() => {
  if (usernameDebounceTimer) {
    clearTimeout(usernameDebounceTimer)
  }
  if (emailDebounceTimer) {
    clearTimeout(emailDebounceTimer)
  }
})

// 表单验证规则
const formRules: Record<string, Rule[]> = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度为3-20个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_]+$/, message: '用户名只能包含字母、数字和下划线', trigger: 'blur' },
    { 
      validator: async (rule: any, value: string) => {
        if (value && usernameStatus.value === 'error') {
          throw new Error('该用户名已被使用')
        }
      },
      trigger: 'blur'
    }
  ],
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' },
    {
      validator: async (rule: any, value: string) => {
        if (value && emailStatus.value === 'error') {
          throw new Error('该邮箱已被注册')
        }
      },
      trigger: 'blur'
    }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 8, message: '密码至少8个字符', trigger: 'blur' },
    {
      validator: (rule: any, value: string) => {
        if (value && passwordStrength.value < 3) {
          return Promise.reject('密码强度太弱，请包含大小写字母、数字和特殊字符')
        }
        return Promise.resolve()
      },
      trigger: 'blur'
    }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (rule: any, value: string) => {
        if (value && value !== formData.password) {
          return Promise.reject('两次输入的密码不匹配')
        }
        return Promise.resolve()
      },
      trigger: 'blur'
    }
  ],
  agreement: [
    {
      validator: (rule: any, value: boolean) => {
        if (!agreement.value) {
          return Promise.reject('请阅读并同意服务条款和隐私政策')
        }
        return Promise.resolve()
      },
      trigger: 'change'
    }
  ]
}

// 密码强度计算
const passwordStrengthText = computed(() => {
  switch (passwordStrength.value) {
    case 1: return '弱'
    case 2: return '中'
    case 3: return '强'
    case 4: return '很强'
    default: return ''
  }
})

const passwordStrengthClass = computed(() => {
  switch (passwordStrength.value) {
    case 1: return 'weak'
    case 2: return 'medium'
    case 3: return 'strong'
    case 4: return 'very-strong'
    default: return ''
  }
})

const passwordStrengthWidth = computed(() => {
  return `${(passwordStrength.value / 4) * 100}%`
})

// 检查密码强度
const checkPasswordStrength = () => {
  const password = formData.password
  let strength = 0
  
  if (password.length >= 8) strength++
  if (/[a-z]/.test(password)) strength++
  if (/[A-Z]/.test(password)) strength++
  if (/[0-9]/.test(password)) strength++
  if (/[^a-zA-Z0-9]/.test(password)) strength++
  
  passwordStrength.value = Math.min(strength, 4)
}

// 检查用户名可用性（带防抖）
const checkUsernameAvailability = async () => {
  // 清除之前的定时器
  if (usernameDebounceTimer) {
    clearTimeout(usernameDebounceTimer)
  }
  
  if (!formData.username || formData.username.length < 3) {
    usernameStatus.value = ''
    return
  }
  
  // 设置防抖延迟
  usernameDebounceTimer = setTimeout(async () => {
    usernameChecking.value = true
    usernameStatus.value = ''
    
    try {
      const response = await AuthAPI.checkUsernameAvailability(formData.username)
      usernameStatus.value = response.available ? 'success' : 'error'
      if (!response.available && response.message) {
        // 显示友好的错误提示
        message.warning(response.message)
      } else if (response.available) {
        // 可选：显示成功提示
        message.success('用户名可用', 2)
      }
    } catch (error) {
      usernameStatus.value = 'error'
      message.error('检查用户名可用性失败，请重试')
      console.error('检查用户名可用性失败:', error)
    } finally {
      usernameChecking.value = false
    }
  }, 500) // 500ms防抖延迟
}

// 检查邮箱可用性（带防抖）
const checkEmailAvailability = async () => {
  // 清除之前的定时器
  if (emailDebounceTimer) {
    clearTimeout(emailDebounceTimer)
  }
  
  if (!formData.email || !/\S+@\S+\.\S+/.test(formData.email)) {
    emailStatus.value = ''
    return
  }
  
  // 设置防抖延迟
  emailDebounceTimer = setTimeout(async () => {
    emailChecking.value = true
    emailStatus.value = ''
    
    try {
      const response = await AuthAPI.checkEmailAvailability(formData.email)
      emailStatus.value = response.available ? 'success' : 'error'
      if (!response.available && response.message) {
        // 显示友好的错误提示
        message.warning(response.message)
      } else if (response.available) {
        // 可选：显示成功提示
        message.success('邮箱可用', 2)
      }
    } catch (error) {
      emailStatus.value = 'error'
      message.error('检查邮箱可用性失败，请重试')
      console.error('检查邮箱可用性失败:', error)
    } finally {
      emailChecking.value = false
    }
  }, 500) // 500ms防抖延迟
}

// 处理注册
const handleRegister = async (values: RegisterData) => {
  // 检查服务条款同意状态
  if (!agreement.value) {
    message.error('请阅读并同意服务条款和隐私政策')
    return
  }
  
  try {
    const success = await authStore.register(values)
    if (success) {
      message.success('注册成功，欢迎使用 JobView!')
      await router.push('/')
    }
  } catch (error) {
    console.error('注册过程出错:', error)
  }
}

// 处理注册失败
const handleRegisterFailed = (errorInfo: any) => {
  console.error('表单验证失败:', errorInfo)
  message.error('请检查输入信息')
}

// 显示服务条款
const showTerms = () => {
  termsVisible.value = true
}

// 显示隐私政策
const showPrivacy = () => {
  privacyVisible.value = true
}
</script>

<style scoped>
.register-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  position: relative;
  overflow: hidden;
  padding: 20px 0;
}

.register-background {
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
  animation: float 8s ease-in-out infinite;
}

.shape-1 {
  width: 180px;
  height: 180px;
  top: 15%;
  left: 5%;
  animation-delay: 0s;
}

.shape-2 {
  width: 120px;
  height: 120px;
  top: 65%;
  right: 10%;
  animation-delay: 2s;
}

.shape-3 {
  width: 90px;
  height: 90px;
  bottom: 25%;
  left: 75%;
  animation-delay: 4s;
}

.shape-4 {
  width: 60px;
  height: 60px;
  top: 40%;
  left: 20%;
  animation-delay: 6s;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0px) rotate(0deg);
  }
  50% {
    transform: translateY(-30px) rotate(180deg);
  }
}

.register-content {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 450px;
  padding: 0 24px;
}

.register-header {
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

.register-form-wrapper {
  background: white;
  padding: 32px;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.password-strength {
  margin-top: 8px;
}

.strength-bar {
  height: 4px;
  background-color: #f0f0f0;
  border-radius: 2px;
  overflow: hidden;
  margin-bottom: 4px;
}

.strength-fill {
  height: 100%;
  transition: width 0.3s ease;
  border-radius: 2px;
}

.strength-fill.weak {
  background-color: #ff4d4f;
}

.strength-fill.medium {
  background-color: #fa8c16;
}

.strength-fill.strong {
  background-color: #52c41a;
}

.strength-fill.very-strong {
  background-color: #1890ff;
}

.strength-text {
  font-size: 12px;
  font-weight: 500;
}

.strength-text.weak {
  color: #ff4d4f;
}

.strength-text.medium {
  color: #fa8c16;
}

.strength-text.strong {
  color: #52c41a;
}

.strength-text.very-strong {
  color: #1890ff;
}

.register-button {
  height: 48px;
  border-radius: 8px;
  font-weight: 500;
  font-size: 16px;
  background: linear-gradient(135deg, #1890ff 0%, #096dd9 100%);
  border: none;
}

.register-button:hover {
  background: linear-gradient(135deg, #40a9ff 0%, #1890ff 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(24, 144, 255, 0.3);
}

.login-link {
  text-align: center;
  margin-bottom: 0;
}

.login-link-btn {
  color: #1890ff;
  text-decoration: none;
  font-weight: 500;
  margin-left: 8px;
}

.login-link-btn:hover {
  color: #40a9ff;
}

.text-center {
  text-align: center;
}

.terms-content, .privacy-content {
  max-height: 400px;
  overflow-y: auto;
}

.terms-content h4, .privacy-content h4 {
  color: #1890ff;
  margin-top: 16px;
  margin-bottom: 8px;
}

.terms-content p, .privacy-content p {
  margin-bottom: 12px;
  line-height: 1.6;
}

/* 响应式设计 */
@media (max-width: 480px) {
  .register-content {
    max-width: 100%;
    padding: 0 16px;
  }

  .register-form-wrapper {
    padding: 24px 20px;
  }

  .title {
    font-size: 28px;
  }

  .subtitle {
    font-size: 14px;
  }
}

/* 深色模式适配 */
@media (prefers-color-scheme: dark) {
  .register-form-wrapper {
    background: rgba(255, 255, 255, 0.95);
  }
}
</style>