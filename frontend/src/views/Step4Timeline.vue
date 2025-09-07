<template>
  <div class="timeline-page">
    <h1>Step 4: 实际使用ApplicationForm组件</h1>
    <a-card title="测试ApplicationForm使用" :loading="loading">
      <p>applications count: {{ applications.length }}</p>
      <a-button type="primary" @click="showModal = true">显示表单</a-button>
      
      <!-- 这里实际使用ApplicationForm -->
      <ApplicationForm
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
import ApplicationForm from '../components/ApplicationForm.vue'

const jobStore = useJobApplicationStore()
const { applications, loading } = storeToRefs(jobStore)

const showModal = ref(false)

const onSuccess = () => {
  console.log('表单提交成功')
  showModal.value = false
}

console.log('Step4Timeline加载成功')
</script>

<style scoped>
.timeline-page {
  padding: 20px;
}
</style>