<script setup lang="ts">
import AppLayout from './components/AppLayout.vue'
import NetworkStatus from './components/NetworkStatus.vue'
import { onMounted } from 'vue'
import { JobApplicationAPI } from './api/jobApplication'
import { AuthAPI } from './api/auth'
import { useAuthStore } from './stores/auth'
import { message } from 'ant-design-vue'

const authStore = useAuthStore()

// 应用启动时的初始化
onMounted(async () => {
  // 初始化认证状态
  authStore.initAuth()
  
  // 检查后端连接
  try {
    const [jobApiHealth, authApiHealth] = await Promise.all([
      JobApplicationAPI.healthCheck(),
      AuthAPI.healthCheck()
    ])
    
    if (!jobApiHealth) {
      message.warning('求职记录服务连接异常，部分功能可能不可用')
    }
    
    if (!authApiHealth) {
      message.error('认证服务连接异常，请检查后端服务状态')
    }
    
    if (jobApiHealth && authApiHealth) {
      console.log('✅ 所有后端服务连接正常')
    }
  } catch (error) {
    console.error('健康检查失败:', error)
    message.warning('无法连接到后端服务器，请确保后端服务正在运行')
  }
})
</script>

<template>
  <div id="app">
    <!-- 网络状态指示器 -->
    <NetworkStatus />
    <!-- 主应用布局 -->
    <AppLayout />
  </div>
</template>

<style>
/* 全局样式重置 */
* {
  box-sizing: border-box;
}

html, body {
  margin: 0;
  padding: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

#app {
  height: 100vh;
  width: 100%;
}

/* 自定义滚动条样式 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* 自定义Ant Design样式覆盖 */
.ant-timeline-item-content {
  margin-left: 20px;
}

.ant-card {
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.ant-btn {
  border-radius: 6px;
}

.ant-input, .ant-select-selector {
  border-radius: 6px;
}

/* 响应式工具类 */
@media (max-width: 768px) {
  .hide-on-mobile {
    display: none !important;
  }
}

@media (min-width: 769px) {
  .show-on-mobile {
    display: none !important;
  }
}
</style>
