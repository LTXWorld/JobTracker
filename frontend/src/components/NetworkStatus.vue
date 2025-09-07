<template>
  <div class="network-status" v-if="showStatus">
    <a-alert
      :type="alertType"
      :message="statusMessage"
      :description="statusDescription"
      :show-icon="true"
      :closable="isOnline"
      banner
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

const isOnline = ref(navigator.onLine)
const showStatus = ref(!navigator.onLine)
const lastOnlineTime = ref<Date | null>(null)

let reconnectAttempts = 0
const maxReconnectAttempts = 5

// 网络状态变化处理
const handleOnline = () => {
  isOnline.value = true
  reconnectAttempts = 0
  
  if (lastOnlineTime.value) {
    const offlineDuration = Date.now() - lastOnlineTime.value.getTime()
    if (offlineDuration > 5000) { // 离线超过5秒才显示恢复消息
      showStatus.value = true
      // 3秒后自动隐藏
      setTimeout(() => {
        showStatus.value = false
      }, 3000)
    }
  } else {
    showStatus.value = false
  }
  
  lastOnlineTime.value = null
}

const handleOffline = () => {
  isOnline.value = false
  showStatus.value = true
  lastOnlineTime.value = new Date()
}

// 计算状态信息
const alertType = computed(() => {
  return isOnline.value ? 'success' : 'error'
})

const statusMessage = computed(() => {
  if (isOnline.value) {
    return '网络连接已恢复'
  } else {
    return '网络连接已断开'
  }
})

const statusDescription = computed(() => {
  if (isOnline.value) {
    return '您现在可以正常使用所有功能'
  } else {
    return '请检查网络连接，某些功能可能不可用'
  }
})

// 生命周期管理
onMounted(() => {
  window.addEventListener('online', handleOnline)
  window.addEventListener('offline', handleOffline)
  
  // 如果初始状态是离线，显示状态
  if (!navigator.onLine) {
    handleOffline()
  }
})

onUnmounted(() => {
  window.removeEventListener('online', handleOnline)
  window.removeEventListener('offline', handleOffline)
})
</script>

<style scoped>
.network-status {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
}

.network-status :deep(.ant-alert) {
  border-radius: 0;
  margin-bottom: 0;
}
</style>