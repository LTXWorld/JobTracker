<template>
  <div class="detail-page">
    <a-card title="投递详情" :loading="loading">
      <p>详情页面开发中...</p>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { useJobApplicationStore } from '../stores/jobApplication'

const { loading } = useJobApplicationStore()
</script>

<style scoped>
.detail-page {
  padding: 16px;
}
</style>