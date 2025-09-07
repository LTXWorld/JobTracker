<template>
  <a-modal
    :open="visible"
    title="简化表单测试"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <p>这是一个简化的ApplicationForm组件</p>
    <p>initialData: {{ initialData?.company_name || '无数据' }}</p>
    <a-input v-model:value="testValue" placeholder="测试输入" />
  </a-modal>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { JobApplication } from '../types'

interface Props {
  visible: boolean
  initialData?: JobApplication | null
}

interface Emits {
  (e: 'update:visible', value: boolean): void
  (e: 'success'): void
}

defineProps<Props>()
const emit = defineEmits<Emits>()

const testValue = ref('')

const handleOk = () => {
  console.log('表单确定', testValue.value)
  emit('success')
}

const handleCancel = () => {
  console.log('表单取消')
  emit('update:visible', false)
}

console.log('SimpleApplicationForm组件加载成功')
</script>