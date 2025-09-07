<template>
  <div class="not-found-container">
    <div class="not-found-content">
      <div class="error-illustration">
        <svg width="200" height="200" viewBox="0 0 200 200" fill="none" xmlns="http://www.w3.org/2000/svg">
          <circle cx="100" cy="100" r="80" fill="#f0f0f0" stroke="#d9d9d9" stroke-width="2"/>
          <text x="100" y="110" text-anchor="middle" font-size="48" font-weight="bold" fill="#8c8c8c">404</text>
        </svg>
      </div>
      
      <h1 class="error-title">页面不存在</h1>
      <p class="error-description">
        抱歉，您访问的页面不存在或已被删除。
      </p>
      
      <div class="error-actions">
        <a-button type="primary" size="large" @click="goHome">
          返回首页
        </a-button>
        <a-button size="large" @click="goBack">
          返回上一页
        </a-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

const goHome = () => {
  router.push('/')
}

const goBack = () => {
  router.back()
}
</script>

<style scoped>
.not-found-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f5f5f5;
}

.not-found-content {
  text-align: center;
  padding: 48px;
}

.error-illustration {
  margin-bottom: 32px;
}

.error-title {
  font-size: 32px;
  font-weight: 600;
  color: #262626;
  margin-bottom: 16px;
}

.error-description {
  font-size: 16px;
  color: #8c8c8c;
  margin-bottom: 32px;
  line-height: 1.6;
}

.error-actions {
  display: flex;
  gap: 16px;
  justify-content: center;
}

@media (max-width: 480px) {
  .not-found-content {
    padding: 24px;
  }
  
  .error-title {
    font-size: 24px;
  }
  
  .error-actions {
    flex-direction: column;
    align-items: center;
  }
}
</style>