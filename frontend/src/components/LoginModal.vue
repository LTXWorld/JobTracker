<template>
  <a-modal
    v-model:open="visible"
    title="快捷登录"
    :footer="null"
    width="400px"
    @cancel="handleCancel"
  >
    <a-form
      :model="formData"
      :rules="formRules"
      @finish="handleLogin"
      layout="vertical"
    >
      <a-form-item name="username" label="用户名">
        <a-input
          v-model:value="formData.username"
          placeholder="请输入用户名"
          allow-clear
          autocomplete="username"
        >
          <template #prefix>
            <UserOutlined />
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
            <LockOutlined />
          </template>
        </a-input-password>
      </a-form-item>

      <a-form-item>
        <div class="modal-actions">
          <a-button @click="handleCancel">
            取消
          </a-button>
          <a-button 
            type="primary" 
            html-type="submit"
            :loading="authStore.loading"
          >
            登录
          </a-button>
        </div>
      </a-form-item>
    </a-form>

    <div class="modal-footer">
      <p>还没有账号？<a @click="goToRegister">立即注册</a></p>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { UserOutlined, LockOutlined } from '@ant-design/icons-vue'
import { useAuthStore } from '../stores/auth'
import type { LoginCredentials } from '../types/auth'
import type { Rule } from 'ant-design-vue/es/form'

interface Props {
  open?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  open: false
})

const emit = defineEmits<{
  'update:open': [value: boolean]
  'login-success': []
}>()

const router = useRouter()
const authStore = useAuthStore()

// 模态框显示状态
const visible = ref(props.open)

// 表单数据
const formData = reactive<LoginCredentials>({
  username: '',
  password: ''
})

// 表单验证规则
const formRules: Record<string, Rule[]> = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

// 监听父组件传入的open状态
watch(() => props.open, (newVal) => {
  visible.value = newVal
})

// 监听内部visible状态，同步到父组件
watch(visible, (newVal) => {
  emit('update:open', newVal)
  if (!newVal) {
    // 关闭时重置表单
    formData.username = ''
    formData.password = ''
  }
})

// 处理登录
const handleLogin = async (values: LoginCredentials) => {
  try {
    const success = await authStore.login(values)
    if (success) {
      visible.value = false
      emit('login-success')
    }
  } catch (error) {
    console.error('登录失败:', error)
  }
}

// 处理取消
const handleCancel = () => {
  visible.value = false
}

// 跳转到注册页
const goToRegister = () => {
  visible.value = false
  router.push('/register')
}
</script>

<style scoped>
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.modal-footer {
  text-align: center;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
  margin-top: 24px;
}

.modal-footer p {
  margin: 0;
  color: #666;
}

.modal-footer a {
  color: #1890ff;
  text-decoration: none;
  cursor: pointer;
}

.modal-footer a:hover {
  color: #40a9ff;
}
</style>