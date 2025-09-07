<template>
  <a-layout class="app-layout">
    <!-- 顶部导航栏 -->
    <a-layout-header class="app-header">
      <div class="header-content">
        <div class="logo">
          <div class="logo-icon">
            <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <rect x="3" y="4" width="18" height="18" rx="2" stroke="#1890ff" stroke-width="2" fill="none"/>
              <path d="M16 10l-4 4-2-2" stroke="#1890ff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </div>
          <h1>JobView</h1>
        </div>
        
        <div class="header-actions">
          <!-- 用户信息区域 -->
          <div class="user-section" v-if="authStore.isLoggedIn">
            <span class="welcome-text">欢迎，{{ authStore.userName }}</span>
            <a-dropdown>
              <template #overlay>
                <a-menu>
                  <a-menu-item key="profile" @click="goToProfile">
                    <template #icon><UserOutlined /></template>
                    个人资料
                  </a-menu-item>
                  <a-menu-divider />
                  <a-menu-item key="export">
                    <template #icon><ExportOutlined /></template>
                    导出数据
                  </a-menu-item>
                  <a-menu-item key="settings">
                    <template #icon><SettingOutlined /></template>
                    系统设置
                  </a-menu-item>
                  <a-menu-divider />
                  <a-menu-item key="logout" @click="handleLogout">
                    <template #icon><LogoutOutlined /></template>
                    退出登录
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button type="text" class="user-avatar-btn">
                <a-avatar :size="32" class="user-avatar">
                  <template #icon>
                    <UserOutlined />
                  </template>
                </a-avatar>
                <DownOutlined class="dropdown-icon" />
              </a-button>
            </a-dropdown>
          </div>
          
          <!-- 未登录状态 -->
          <div class="auth-actions" v-else>
            <a-button type="text" @click="goToLogin" class="auth-btn">
              登录
            </a-button>
            <a-button type="primary" ghost size="small" @click="goToRegister">
              注册
            </a-button>
          </div>
        </div>
      </div>
    </a-layout-header>

    <!-- 标签导航栏 -->
    <div class="tab-navigation" v-if="authStore.isLoggedIn">
      <a-tabs 
        v-model:activeKey="activeTab" 
        @change="handleTabChange"
        type="card"
        class="main-tabs"
      >
        <a-tab-pane key="kanban" tab="看板视图">
          <template #tab>
            <span>
              <AppstoreOutlined />
              看板视图
            </span>
          </template>
        </a-tab-pane>
        <a-tab-pane key="timeline" tab="投递记录">
          <template #tab>
            <span>
              <ClockCircleOutlined />
              投递记录
            </span>
          </template>
        </a-tab-pane>
        <a-tab-pane key="reminders" tab="提醒中心">
          <template #tab>
            <span>
              <BellOutlined />
              提醒中心
            </span>
          </template>
        </a-tab-pane>
        <a-tab-pane key="statistics" tab="数据统计">
          <template #tab>
            <span>
              <BarChartOutlined />
              数据统计
            </span>
          </template>
        </a-tab-pane>
      </a-tabs>
      <div class="tab-actions">
        <a-button type="primary" ghost :icon="h(ReloadOutlined)" @click="handleRefresh" :loading="refreshing">
          刷新数据
        </a-button>
      </div>
    </div>

    <!-- 内容区域 -->
    <a-layout-content class="app-content" :class="{ 'no-tabs': !authStore.isLoggedIn }">
      <div class="content-wrapper">
        <router-view />
      </div>
    </a-layout-content>

    <!-- 底部 -->
    <a-layout-footer class="app-footer">
      <div class="footer-content">
        <p>
          JobView 求职进程管理系统 © {{ currentYear }} 
          <a-divider type="vertical" />
          <a href="https://github.com" target="_blank">
            <GithubOutlined /> GitHub
          </a>
        </p>
      </div>
    </a-layout-footer>
  </a-layout>
</template>

<script setup lang="ts">
import { ref, computed, h, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import {
  AppstoreOutlined,
  ClockCircleOutlined,
  BarChartOutlined,
  BellOutlined,
  ReloadOutlined,
  UserOutlined,
  DownOutlined,
  ExportOutlined,
  SettingOutlined,
  LogoutOutlined,
  GithubOutlined
} from '@ant-design/icons-vue'
import { useJobApplicationStore } from '../stores/jobApplication'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const route = useRoute()
const jobStore = useJobApplicationStore()
const authStore = useAuthStore()

const refreshing = ref(false)

// 当前激活的标签
const activeTab = computed({
  get: () => {
    const routeName = route.name as string
    if (!routeName || routeName === 'home') {
      return 'kanban'
    }
    // 如果当前在认证相关页面，不显示标签
    if (['login', 'register', 'profile'].includes(routeName)) {
      return ''
    }
    return routeName
  },
  set: (value: string) => {
    if (value) {
      router.push({ name: value })
    }
  }
})

// 当前年份
const currentYear = computed(() => new Date().getFullYear())

// 组件挂载时初始化认证状态
onMounted(() => {
  authStore.initAuth()
})

// 标签切换处理
const handleTabChange = (key: string) => {
  router.push({ name: key })
}

// 刷新数据
const handleRefresh = async () => {
  refreshing.value = true
  try {
    await jobStore.fetchApplications()
    message.success('数据刷新成功')
  } catch (error) {
    message.error('数据刷新失败')
  } finally {
    refreshing.value = false
  }
}

// 跳转到个人资料页
const goToProfile = () => {
  router.push('/profile')
}

// 跳转到登录页
const goToLogin = () => {
  router.push('/login')
}

// 跳转到注册页
const goToRegister = () => {
  router.push('/register')
}

// 处理登出
const handleLogout = () => {
  Modal.confirm({
    title: '确认退出？',
    content: '您确定要退出登录吗？',
    okText: '确认',
    cancelText: '取消',
    onOk: async () => {
      try {
        await authStore.logout()
        router.push('/login')
      } catch (error) {
        console.error('登出失败:', error)
      }
    }
  })
}
</script>

<style scoped>
.app-layout {
  min-height: 100vh;
}

.app-header {
  padding: 0;
  background: #001529;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  height: 48px;
  line-height: 48px;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  padding: 0 24px;
  max-width: 1400px;
  margin: 0 auto;
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-icon {
  display: flex;
  align-items: center;
}

.logo h1 {
  color: #fff;
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  white-space: nowrap;
}

.header-actions {
  display: flex;
  align-items: center;
}

.user-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.welcome-text {
  color: rgba(255, 255, 255, 0.85);
  font-size: 14px;
}

.user-avatar-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  color: rgba(255, 255, 255, 0.85);
  padding: 4px 8px;
  height: auto;
}

.user-avatar-btn:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
}

.user-avatar {
  background: #1890ff;
}

.dropdown-icon {
  font-size: 12px;
}

.auth-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.auth-btn {
  color: rgba(255, 255, 255, 0.85);
}

.auth-btn:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
}

/* 标签导航栏样式 */
.tab-navigation {
  background: #fff;
  padding: 0 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.main-tabs {
  flex: 1;
}

.main-tabs :deep(.ant-tabs-nav) {
  margin: 0;
  border: none;
}

.main-tabs :deep(.ant-tabs-tab) {
  padding: 12px 16px;
  font-size: 14px;
  border: none;
  background: transparent;
  border-radius: 6px 6px 0 0;
  transition: all 0.3s;
}

.main-tabs :deep(.ant-tabs-tab:hover) {
  background: #e6f7ff;
  color: #1890ff;
}

.main-tabs :deep(.ant-tabs-tab-active) {
  background: #1890ff;
  color: #fff;
}

.main-tabs :deep(.ant-tabs-tab-active .anticon) {
  color: #fff;
}

.main-tabs :deep(.ant-tabs-tab span) {
  display: flex;
  align-items: center;
  gap: 8px;
}

.main-tabs :deep(.ant-tabs-ink-bar) {
  display: none;
}

.main-tabs :deep(.ant-tabs-nav::before) {
  border: none;
}

.tab-actions {
  margin-left: 16px;
}

.app-content {
  background: #f0f2f5;
  padding: 0;
  min-height: calc(100vh - 48px - 56px - 70px);
}

.app-content.no-tabs {
  min-height: calc(100vh - 48px - 70px);
}

.content-wrapper {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0;
}

.app-footer {
  background: #f0f2f5;
  padding: 24px 0;
  border-top: 1px solid #d9d9d9;
  text-align: center;
}

.footer-content {
  max-width: 1200px;
  margin: 0 auto;
  color: #666;
}

.footer-content p {
  margin: 0;
}

.footer-content a {
  color: #666;
  text-decoration: none;
  transition: color 0.3s;
}

.footer-content a:hover {
  color: #1890ff;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .header-content {
    padding: 0 16px;
  }
  
  .logo h1 {
    font-size: 18px;
  }
  
  .welcome-text {
    display: none;
  }
  
  .tab-navigation {
    padding: 0 16px;
  }
  
  .content-wrapper {
    padding: 16px;
  }
}

@media (max-width: 480px) {
  .header-content {
    padding: 0 12px;
  }
  
  .logo h1 {
    font-size: 16px;
  }
  
  .tab-navigation {
    padding: 0 12px;
  }
  
  .content-wrapper {
    padding: 12px;
  }
  
  .main-tabs :deep(.ant-tabs-tab) {
    padding: 8px 12px;
    font-size: 12px;
  }
  
  .main-tabs :deep(.ant-tabs-tab span) {
    gap: 4px;
  }
}
</style>

<style scoped>
.app-layout {
  min-height: 100vh;
}

.app-header {
  padding: 0;
  background: #001529;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  height: 48px;
  line-height: 48px;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  padding: 0 24px;
  max-width: 1400px;
  margin: 0 auto;
}

.logo h1 {
  color: #fff;
  margin: 0;
  font-size: 18px;
  font-weight: 500;
  white-space: nowrap;
}

.header-actions {
  display: flex;
  align-items: center;
}

.header-actions :deep(.ant-btn) {
  color: rgba(255, 255, 255, 0.85);
}

.header-actions :deep(.ant-btn:hover) {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
}

/* 标签导航栏样式 */
.tab-navigation {
  background: #fff;
  padding: 0 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.main-tabs {
  flex: 1;
}

.main-tabs :deep(.ant-tabs-nav) {
  margin: 0;
  border: none;
}

.main-tabs :deep(.ant-tabs-tab) {
  padding: 12px 16px;
  font-size: 14px;
  border: none;
  background: transparent;
}

.main-tabs :deep(.ant-tabs-tab-active) {
  background: #1890ff;
  color: #fff;
}

.main-tabs :deep(.ant-tabs-tab-active .anticon) {
  color: #fff;
}

.main-tabs :deep(.ant-tabs-tab span) {
  display: flex;
  align-items: center;
  gap: 8px;
}

.main-tabs :deep(.ant-tabs-ink-bar) {
  display: none;
}

.main-tabs :deep(.ant-tabs-nav::before) {
  border: none;
}

.tab-actions {
  margin-left: 16px;
}

.app-content {
  background: #f0f2f5;
  padding: 0;
  min-height: calc(100vh - 48px - 56px - 70px);
}

.content-wrapper {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0;
}

.app-footer {
  background: #f0f2f5;
  padding: 24px 0;
  border-top: 1px solid #d9d9d9;
  text-align: center;
}

.footer-content {
  max-width: 1200px;
  margin: 0 auto;
  color: #666;
}

.footer-content p {
  margin: 0;
}

.footer-content a {
  color: #666;
  text-decoration: none;
  transition: color 0.3s;
}

.footer-content a:hover {
  color: #1890ff;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .header-content {
    padding: 0 16px;
  }
  
  .logo h1 {
    font-size: 16px;
  }
  
  .content-wrapper {
    padding: 16px;
  }
  
  .nav-menu {
    display: none; /* 移动端隐藏导航菜单，可后续添加抽屉菜单 */
  }
}

@media (max-width: 480px) {
  .header-content {
    padding: 0 12px;
  }
  
  .content-wrapper {
    padding: 12px;
  }
}
</style>