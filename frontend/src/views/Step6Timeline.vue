<template>
  <div class="timeline-page">
    <h1>Step 6: 测试新ApplicationForm</h1>
    <a-card title="测试新ApplicationForm" :loading="loading">
      <p>applications count: {{ applications.length }}</p>
      <a-button type="primary" @click="showModal = true">显示新表单</a-button>
      
      <!-- 使用新ApplicationForm -->
      <NewApplicationForm
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
import NewApplicationForm from '../components/NewApplicationForm.vue'

const jobStore = useJobApplicationStore()
const { applications, loading } = storeToRefs(jobStore)

const showModal = ref(false)

const onSuccess = () => {
  console.log('新表单提交成功')
  showModal.value = false
}

console.log('Step6Timeline加载成功')
</script>

<style scoped>
.timeline-page {
  padding: 20px;
}
</style>