<template>
  <div class="timeline-page">
    <h1>Step 2: 添加Ant Design组件测试</h1>
    <a-card title="测试卡片" :loading="loading">
      <p>applications count: {{ applications.length }}</p>
      <a-button type="primary">测试按钮</a-button>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useJobApplicationStore } from '../stores/jobApplication'

const jobStore = useJobApplicationStore()
const { applications, loading } = storeToRefs(jobStore)

console.log('Step2Timeline加载成功')
</script>

<style scoped>
.timeline-page {
  padding: 20px;
}
</style>