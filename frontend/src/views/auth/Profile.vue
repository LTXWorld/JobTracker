<template>
  <div class="profile-container">
    <a-row :gutter="24">
      <!-- 左侧用户信息卡片 -->
      <a-col :xs="24" :md="8">
        <a-card class="profile-card" :bordered="false">
          <div class="user-avatar-section">
            <div class="avatar-upload-container">
              <a-upload
                name="avatar"
                list-type="picture-card"
                class="avatar-uploader"
                :show-upload-list="false"
                :before-upload="beforeAvatarUpload"
                :custom-request="handleAvatarUpload"
              >
                <a-avatar 
                  :size="100" 
                  :src="userAvatar"
                  class="user-avatar"
                >
                  <template #icon>
                    <UserOutlined />
                  </template>
                </a-avatar>
                <div class="avatar-upload-overlay">
                  <CameraOutlined />
                  <span>更换头像</span>
                </div>
              </a-upload>
            </div>
            <div class="user-basic-info">
              <h3 class="username">{{ authStore.user?.full_name || authStore.userName }}</h3>
              <p class="user-email">{{ authStore.userEmail }}</p>
              <p class="join-date">
                <CalendarOutlined /> 
                加入时间: {{ formatDate(authStore.user?.created_at) }}
              </p>
              <p class="last-login" v-if="authStore.user?.last_login_at">
                <ClockCircleOutlined />
                最后登录: {{ formatDate(authStore.user?.last_login_at) }}
              </p>
            </div>
          </div>
        </a-card>

        <!-- 账户统计 -->
        <a-card title="账户统计" class="stats-card" :bordered="false">
          <div class="stat-item">
            <span class="stat-label">求职记录数</span>
            <span class="stat-value">{{ userStats.totalApplications || 0 }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">进行中申请</span>
            <span class="stat-value">{{ userStats.activeApplications || 0 }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">收到Offer</span>
            <span class="stat-value">{{ userStats.receivedOffers || 0 }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">成功率</span>
            <span class="stat-value">{{ userStats.successRate || '0%' }}</span>
          </div>
        </a-card>
      </a-col>

      <!-- 右侧表单区域 -->
      <a-col :xs="24" :md="16">
        <a-card title="个人资料" :bordered="false" class="form-card">
          <a-tabs v-model:activeKey="activeTab" type="card">
            <!-- 基本信息标签页 -->
            <a-tab-pane key="profile" tab="基本信息">
              <a-form
                :model="profileForm"
                :rules="profileRules"
                @finish="handleUpdateProfile"
                layout="vertical"
                size="large"
              >
                <a-row :gutter="16">
                  <a-col :span="12">
                    <a-form-item name="username" label="用户名">
                      <a-input
                        v-model:value="profileForm.username"
                        placeholder="请输入用户名"
                        allow-clear
                        :disabled="profileUpdateLoading"
                      >
                        <template #prefix>
                          <UserOutlined style="color: rgba(0,0,0,.25)" />
                        </template>
                      </a-input>
                    </a-form-item>
                  </a-col>
                  <a-col :span="12">
                    <a-form-item name="email" label="邮箱地址">
                      <a-input
                        v-model:value="profileForm.email"
                        type="email"
                        placeholder="请输入邮箱地址"
                        allow-clear
                        :disabled="profileUpdateLoading"
                      >
                        <template #prefix>
                          <MailOutlined style="color: rgba(0,0,0,.25)" />
                        </template>
                      </a-input>
                    </a-form-item>
                  </a-col>
                </a-row>

                <a-row :gutter="16">
                  <a-col :span="12">
                    <a-form-item name="full_name" label="真实姓名">
                      <a-input
                        v-model:value="profileForm.full_name"
                        placeholder="请输入真实姓名"
                        allow-clear
                        :disabled="profileUpdateLoading"
                      >
                        <template #prefix>
                          <IdcardOutlined style="color: rgba(0,0,0,.25)" />
                        </template>
                      </a-input>
                    </a-form-item>
                  </a-col>
                  <a-col :span="12">
                    <a-form-item name="phone" label="电话号码">
                      <a-input
                        v-model:value="profileForm.phone"
                        placeholder="请输入电话号码"
                        allow-clear
                        :disabled="profileUpdateLoading"
                      >
                        <template #prefix>
                          <PhoneOutlined style="color: rgba(0,0,0,.25)" />
                        </template>
                      </a-input>
                    </a-form-item>
                  </a-col>
                </a-row>

                <a-row :gutter="16">
                  <a-col :span="12">
                    <a-form-item name="location" label="所在地区">
                      <a-input
                        v-model:value="profileForm.location"
                        placeholder="请输入所在地区"
                        allow-clear
                        :disabled="profileUpdateLoading"
                      >
                        <template #prefix>
                          <EnvironmentOutlined style="color: rgba(0,0,0,.25)" />
                        </template>
                      </a-input>
                    </a-form-item>
                  </a-col>
                  <a-col :span="12">
                    <a-form-item name="website" label="个人网站">
                      <a-input
                        v-model:value="profileForm.website"
                        placeholder="请输入个人网站URL"
                        allow-clear
                        :disabled="profileUpdateLoading"
                      >
                        <template #prefix>
                          <GlobalOutlined style="color: rgba(0,0,0,.25)" />
                        </template>
                      </a-input>
                    </a-form-item>
                  </a-col>
                </a-row>

                <a-form-item name="bio" label="个人简介">
                  <a-textarea
                    v-model:value="profileForm.bio"
                    placeholder="请输入个人简介（最多500字）"
                    :rows="4"
                    :maxlength="500"
                    show-count
                    allow-clear
                    :disabled="profileUpdateLoading"
                  />
                </a-form-item>

                <a-form-item>
                  <a-space>
                    <a-button 
                      type="primary" 
                      html-type="submit"
                      :loading="profileUpdateLoading"
                    >
                      保存更改
                    </a-button>
                    <a-button @click="resetProfileForm">
                      重置
                    </a-button>
                  </a-space>
                </a-form-item>
              </a-form>
            </a-tab-pane>

            <!-- 修改密码标签页 -->
            <a-tab-pane key="password" tab="修改密码">
              <a-form
                :model="passwordForm"
                :rules="passwordRules"
                @finish="handleChangePassword"
                layout="vertical"
                size="large"
              >
                <a-form-item name="currentPassword" label="当前密码">
                  <a-input-password
                    v-model:value="passwordForm.currentPassword"
                    placeholder="请输入当前密码"
                    autocomplete="current-password"
                    :disabled="passwordUpdateLoading"
                  >
                    <template #prefix>
                      <LockOutlined style="color: rgba(0,0,0,.25)" />
                    </template>
                  </a-input-password>
                </a-form-item>

                <a-form-item name="newPassword" label="新密码">
                  <a-input-password
                    v-model:value="passwordForm.newPassword"
                    placeholder="请输入新密码"
                    autocomplete="new-password"
                    :disabled="passwordUpdateLoading"
                    @input="checkNewPasswordStrength"
                  >
                    <template #prefix>
                      <LockOutlined style="color: rgba(0,0,0,.25)" />
                    </template>
                  </a-input-password>
                  <!-- 密码强度指示器 -->
                  <div class="password-strength" v-if="passwordForm.newPassword">
                    <div class="strength-bar">
                      <div 
                        class="strength-fill"
                        :class="newPasswordStrengthClass"
                        :style="{ width: newPasswordStrengthWidth }"
                      ></div>
                    </div>
                    <span class="strength-text" :class="newPasswordStrengthClass">
                      密码强度: {{ newPasswordStrengthText }}
                    </span>
                  </div>
                </a-form-item>

                <a-form-item name="confirmPassword" label="确认新密码">
                  <a-input-password
                    v-model:value="passwordForm.confirmPassword"
                    placeholder="请再次输入新密码"
                    autocomplete="new-password"
                    :disabled="passwordUpdateLoading"
                  >
                    <template #prefix>
                      <LockOutlined style="color: rgba(0,0,0,.25)" />
                    </template>
                  </a-input-password>
                </a-form-item>

                <a-form-item>
                  <a-space>
                    <a-button 
                      type="primary" 
                      html-type="submit"
                      :loading="passwordUpdateLoading"
                      danger
                    >
                      修改密码
                    </a-button>
                    <a-button @click="resetPasswordForm">
                      重置
                    </a-button>
                  </a-space>
                </a-form-item>
              </a-form>
            </a-tab-pane>

            <!-- 账户设置标签页 -->
            <a-tab-pane key="settings" tab="账户设置">
              <div class="settings-section">
                <h4>危险操作</h4>
                <a-alert
                  message="注意"
                  description="以下操作不可逆转，请谨慎操作！"
                  type="warning"
                  show-icon
                  :closable="false"
                  style="margin-bottom: 16px;"
                />
                
                <a-space direction="vertical" size="large" class="danger-actions">
                  <div class="danger-action">
                    <div class="action-info">
                      <h5>清空所有求职记录</h5>
                      <p>这将删除您的所有求职记录，此操作不可恢复</p>
                    </div>
                    <a-button 
                      danger 
                      @click="showClearDataConfirm"
                      :disabled="loading"
                    >
                      清空数据
                    </a-button>
                  </div>

                  <a-divider />

                  <div class="danger-action">
                    <div class="action-info">
                      <h5>注销账户</h5>
                      <p>这将永久删除您的账户和所有相关数据</p>
                    </div>
                    <a-button 
                      danger 
                      @click="showDeleteAccountConfirm"
                      :disabled="loading"
                    >
                      注销账户
                    </a-button>
                  </div>
                </a-space>
              </div>
            </a-tab-pane>
          </a-tabs>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import { 
  UserOutlined, 
  MailOutlined, 
  LockOutlined,
  CalendarOutlined
} from '@ant-design/icons-vue'
import { useAuthStore } from '../../stores/auth'
import { useJobApplicationStore } from '../../stores/jobApplication'
import type { UpdateProfileData } from '../../types/auth'
import type { Rule } from 'ant-design-vue/es/form'
import dayjs from 'dayjs'

const router = useRouter()
const authStore = useAuthStore()
const applicationStore = useJobApplicationStore()

// 当前激活的标签页
const activeTab = ref('profile')

// 加载状态
const profileUpdateLoading = ref(false)
const passwordUpdateLoading = ref(false)
const loading = ref(false)

// 用户统计数据
const userStats = ref({
  totalApplications: 0,
  activeApplications: 0,
  receivedOffers: 0,
  successRate: '0%'
})

// 用户头像（使用默认的 Gravatar 或初始字符）
const userAvatar = computed(() => {
  // 可以后续实现头像上传功能
  return null
})

// 个人资料表单
const profileForm = reactive<UpdateProfileData>({
  username: '',
  email: ''
})

// 密码修改表单
const passwordForm = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 新密码强度
const newPasswordStrength = ref(0)

// 表单验证规则
const profileRules: Record<string, Rule[]> = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度为3-20个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_]+$/, message: '用户名只能包含字母、数字和下划线', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }
  ]
}

const passwordRules: Record<string, Rule[]> = {
  currentPassword: [
    { required: true, message: '请输入当前密码', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 8, message: '新密码至少8个字符', trigger: 'blur' },
    {
      validator: (rule: any, value: string) => {
        if (value && newPasswordStrength.value < 3) {
          return Promise.reject('密码强度太弱，请包含大小写字母、数字和特殊字符')
        }
        return Promise.resolve()
      },
      trigger: 'blur'
    }
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (rule: any, value: string) => {
        if (value && value !== passwordForm.newPassword) {
          return Promise.reject('两次输入的密码不匹配')
        }
        return Promise.resolve()
      },
      trigger: 'blur'
    }
  ]
}

// 新密码强度计算
const newPasswordStrengthText = computed(() => {
  switch (newPasswordStrength.value) {
    case 1: return '弱'
    case 2: return '中'
    case 3: return '强'
    case 4: return '很强'
    default: return ''
  }
})

const newPasswordStrengthClass = computed(() => {
  switch (newPasswordStrength.value) {
    case 1: return 'weak'
    case 2: return 'medium'
    case 3: return 'strong'
    case 4: return 'very-strong'
    default: return ''
  }
})

const newPasswordStrengthWidth = computed(() => {
  return `${(newPasswordStrength.value / 4) * 100}%`
})

// 检查新密码强度
const checkNewPasswordStrength = () => {
  const password = passwordForm.newPassword
  let strength = 0
  
  if (password.length >= 8) strength++
  if (/[a-z]/.test(password)) strength++
  if (/[A-Z]/.test(password)) strength++
  if (/[0-9]/.test(password)) strength++
  if (/[^a-zA-Z0-9]/.test(password)) strength++
  
  newPasswordStrength.value = Math.min(strength, 4)
}

// 格式化日期
const formatDate = (dateString?: string) => {
  if (!dateString) return ''
  return dayjs(dateString).format('YYYY年MM月DD日')
}

// 初始化表单数据
const initFormData = () => {
  if (authStore.user) {
    profileForm.username = authStore.user.username
    profileForm.email = authStore.user.email
  }
}

// 重置个人资料表单
const resetProfileForm = () => {
  initFormData()
}

// 重置密码表单
const resetPasswordForm = () => {
  passwordForm.currentPassword = ''
  passwordForm.newPassword = ''
  passwordForm.confirmPassword = ''
  newPasswordStrength.value = 0
}

// 更新个人资料
const handleUpdateProfile = async (values: UpdateProfileData) => {
  profileUpdateLoading.value = true
  try {
    const success = await authStore.updateProfile(values)
    if (success) {
      message.success('个人资料更新成功')
    }
  } catch (error) {
    console.error('更新个人资料失败:', error)
  } finally {
    profileUpdateLoading.value = false
  }
}

// 修改密码
const handleChangePassword = async () => {
  passwordUpdateLoading.value = true
  try {
    const success = await authStore.changePassword(
      passwordForm.currentPassword,
      passwordForm.newPassword
    )
    if (success) {
      resetPasswordForm()
      // 密码修改成功后会自动登出，跳转到登录页
      await router.push('/login')
    }
  } catch (error) {
    console.error('修改密码失败:', error)
  } finally {
    passwordUpdateLoading.value = false
  }
}

// 显示清空数据确认
const showClearDataConfirm = () => {
  Modal.confirm({
    title: '确认清空所有数据？',
    content: '此操作将删除您的所有求职记录，且不可恢复。您确定要继续吗？',
    okText: '确认清空',
    okType: 'danger',
    cancelText: '取消',
    onOk: handleClearAllData
  })
}

// 显示注销账户确认
const showDeleteAccountConfirm = () => {
  Modal.confirm({
    title: '确认注销账户？',
    content: '此操作将永久删除您的账户和所有相关数据，且不可恢复。您确定要继续吗？',
    okText: '确认注销',
    okType: 'danger',
    cancelText: '取消',
    onOk: handleDeleteAccount
  })
}

// 处理清空所有数据
const handleClearAllData = async () => {
  loading.value = true
  try {
    // 这里需要实现清空数据的API
    message.info('清空数据功能开发中')
  } catch (error) {
    console.error('清空数据失败:', error)
    message.error('清空数据失败')
  } finally {
    loading.value = false
  }
}

// 处理注销账户
const handleDeleteAccount = async () => {
  loading.value = true
  try {
    // 这里需要实现注销账户的API
    message.info('注销账户功能开发中')
  } catch (error) {
    console.error('注销账户失败:', error)
    message.error('注销账户失败')
  } finally {
    loading.value = false
  }
}

// 获取用户统计数据
const fetchUserStats = async () => {
  try {
    await applicationStore.fetchApplications()
    const applications = applicationStore.applications
    
    userStats.value = {
      totalApplications: applications.length,
      activeApplications: applications.filter(app => 
        !['已拒绝', '已接受offer', '简历筛选未通过', '笔试未通过', '一面未通过', '二面未通过', '三面未通过'].includes(app.status)
      ).length,
      receivedOffers: applications.filter(app => 
        ['已收到offer', '已接受offer'].includes(app.status)
      ).length,
      successRate: applications.length > 0 
        ? `${Math.round((applications.filter(app => ['已收到offer', '已接受offer'].includes(app.status)).length / applications.length) * 100)}%`
        : '0%'
    }
  } catch (error) {
    console.error('获取用户统计失败:', error)
  }
}

// 组件挂载
onMounted(() => {
  if (!authStore.isLoggedIn) {
    router.push('/login')
    return
  }
  
  initFormData()
  fetchUserStats()
})
</script>

<style scoped>
.profile-container {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.profile-card {
  margin-bottom: 24px;
}

.user-avatar-section {
  text-align: center;
}

.user-avatar {
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.user-basic-info {
  text-align: center;
}

.username {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 4px;
  color: #262626;
}

.user-email {
  color: #8c8c8c;
  margin-bottom: 8px;
}

.join-date {
  color: #8c8c8c;
  font-size: 14px;
  margin-bottom: 0;
}

.stats-card {
  margin-bottom: 24px;
}

.stat-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.stat-item:last-child {
  border-bottom: none;
}

.stat-label {
  color: #595959;
  font-size: 14px;
}

.stat-value {
  font-weight: 600;
  color: #1890ff;
  font-size: 16px;
}

.form-card {
  min-height: 500px;
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

.settings-section h4 {
  color: #262626;
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
}

.danger-actions {
  width: 100%;
}

.danger-action {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border: 1px solid #ffccc7;
  border-radius: 6px;
  background-color: #fff2f0;
}

.action-info h5 {
  margin: 0 0 4px 0;
  font-size: 14px;
  font-weight: 600;
  color: #262626;
}

.action-info p {
  margin: 0;
  font-size: 12px;
  color: #8c8c8c;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .profile-container {
    padding: 16px;
  }
  
  .danger-action {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
}
</style>