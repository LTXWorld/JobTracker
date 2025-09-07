<template>
  <div class="timeline-page">
    <h1>Step 5: 使用简化ApplicationForm</h1>
    <a-card title="测试简化ApplicationForm" :loading="loading">
      <p>applications count: {{ applications.length }}</p>
      <a-button type="primary" @click="showModal = true">显示简化表单</a-button>
      
      <!-- 使用简化版ApplicationForm -->
      <SimpleApplicationForm
        v-model:visible="showModal"
        :initial-data="null"
        @success="onSuccess"
      />
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useJobApplicationStore } from '../stores/jobApplication'
import SimpleApplicationForm from '../components/SimpleApplicationForm.vue'

const jobStore = useJobApplicationStore()
const { applications, loading } = storeToRefs(jobStore)

const showModal = ref(false)

const onSuccess = () => {
  console.log('简化表单提交成功')
  showModal.value = false
}

console.log('Step5Timeline加载成功')
</script>

<style scoped>
.timeline-page {
  padding: 20px;
}
</style>