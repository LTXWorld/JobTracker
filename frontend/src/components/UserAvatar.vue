<template>
  <a-avatar 
    :size="size" 
    :src="avatarSrc"
    :class="avatarClass"
    @click="handleClick"
  >
    <template #icon v-if="!avatarSrc">
      <UserOutlined />
    </template>
    {{ displayName }}
  </a-avatar>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { UserOutlined } from '@ant-design/icons-vue'
import { useAuthStore } from '../stores/auth'

interface Props {
  size?: number | 'large' | 'small' | 'default'
  clickable?: boolean
  showName?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  size: 'default',
  clickable: false,
  showName: false
})

const emit = defineEmits<{
  click: []
}>()

const authStore = useAuthStore()

// 头像图片源（后续可以实现头像上传功能）
const avatarSrc = computed(() => {
  // 可以基于用户邮箱生成 Gravatar 或使用上传的头像
  return null
})

// 显示名称（如果没有头像图片时显示用户名首字母）
const displayName = computed(() => {
  if (!authStore.user?.username) return ''
  return avatarSrc.value ? '' : authStore.user.username.charAt(0).toUpperCase()
})

// 头像样式类
const avatarClass = computed(() => ({
  'user-avatar-clickable': props.clickable
}))

// 处理点击事件
const handleClick = () => {
  if (props.clickable) {
    emit('click')
  }
}
</script>

<style scoped>
.user-avatar-clickable {
  cursor: pointer;
  transition: all 0.3s;
}

.user-avatar-clickable:hover {
  transform: scale(1.05);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}
</style>