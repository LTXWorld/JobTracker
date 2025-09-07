<template>
  <div class="timeline-page">
    <h1>Step 3: 仅导入ApplicationForm（不使用）</h1>
    <a-card title="测试ApplicationForm导入" :loading="loading">
      <p>applications count: {{ applications.length }}</p>
      <a-button type="primary">测试按钮</a-button>
      <p>ApplicationForm已导入但未使用: {{ ApplicationForm ? '成功' : '失败' }}</p>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useJobApplicationStore } from '../stores/jobApplication'
import ApplicationForm from '../components/ApplicationForm.vue' // 仅导入，不使用

const jobStore = useJobApplicationStore()
const { applications, loading } = storeToRefs(jobStore)

console.log('Step3Timeline加载成功', ApplicationForm)
</script>

<style scoped>
.timeline-page {
  padding: 20px;
}
</style>