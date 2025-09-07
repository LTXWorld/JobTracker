<template>
  <div class="timeline-page">
    <h1>Step 1: 基础显示测试</h1>
    <p>如果你看到这个，说明基础渲染OK</p>
    <p>loading: {{ loading }}</p>
    <p>applications count: {{ applications.length }}</p>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useJobApplicationStore } from '../stores/jobApplication'

const jobStore = useJobApplicationStore()
const { applications, loading } = storeToRefs(jobStore)

console.log('Step1Timeline加载成功')
</script>

<style scoped>
.timeline-page {
  padding: 20px;
}
</style>