<template>
  <div class="homepage">
    <!-- 主页头部 -->
    <div class="homepage-header">
      <div class="brand-section">
        <div class="logo">
          <span class="logo-icon">📋</span>
          <h1>JobView</h1>
        </div>
        <p class="tagline">智能求职投递管理系统</p>
        <p class="description">让数据驱动您的求职成功</p>
      </div>

      <div class="quick-actions">
        <a-button
          type="primary"
          size="large"
          @click="startGuide"
          :icon="h(PlayCircleOutlined)"
        >
          观看使用指导
        </a-button>

        <a-button
          size="large"
          @click="goToKanban"
          :icon="h(AppstoreOutlined)"
        >
          直接开始使用
        </a-button>
      </div>
    </div>

    <!-- 功能特色展示 -->
    <div class="features-showcase">
      <h2>核心功能特色</h2>

      <div class="features-grid">
        <div class="feature-card" @click="goToKanban">
          <div class="feature-icon">📋</div>
          <h3>看板管理</h3>
          <p>拖拽式状态更新，直观管理求职进度</p>
          <div class="feature-badge">推荐</div>
        </div>

        <div class="feature-card" @click="goToTimeline">
          <div class="feature-icon">📅</div>
          <h3>时间线记录</h3>
          <p>详细记录每一次投递，追踪时间轴</p>
        </div>

        <div class="feature-card" @click="goToStats">
          <div class="feature-icon">📊</div>
          <h3>数据统计</h3>
          <p>多维度分析投递数据，洞察成功率</p>
        </div>

        <div class="feature-card" @click="goToReminders">
          <div class="feature-icon">🔔</div>
          <h3>智能提醒</h3>
          <p>永不错过面试和跟进时机</p>
        </div>

        <div class="feature-card" @click="goToResume">
          <div class="feature-icon">📝</div>
          <h3>简历管理</h3>
          <p>集中管理简历，匹配不同职位</p>
        </div>
      </div>
    </div>

    <!-- 使用统计 -->
    <div class="stats-section">
      <h2>系统概览</h2>
      <div class="stats-cards">
        <div class="stat-card">
          <div class="stat-icon">🚀</div>
          <div class="stat-content">
            <div class="stat-number">{{ totalApplications }}</div>
            <div class="stat-label">总投递记录</div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon">⚡</div>
          <div class="stat-content">
            <div class="stat-number">89%</div>
            <div class="stat-label">性能提升</div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon">👥</div>
          <div class="stat-content">
            <div class="stat-number">200+</div>
            <div class="stat-label">并发支持</div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon">📱</div>
          <div class="stat-content">
            <div class="stat-number">100%</div>
            <div class="stat-label">响应式设计</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 快速入门 -->
    <div class="getting-started">
      <h2>快速入门</h2>
      <div class="steps">
        <div class="step">
          <div class="step-number">1</div>
          <div class="step-content">
            <h3>观看指导</h3>
            <p>观看详细的使用指导了解所有功能</p>
          </div>
        </div>

        <div class="step">
          <div class="step-number">2</div>
          <div class="step-content">
            <h3>添加投递记录</h3>
            <p>开始记录您的求职投递信息</p>
          </div>
        </div>

        <div class="step">
          <div class="step-number">3</div>
          <div class="step-content">
            <h3>管理进度</h3>
            <p>使用看板拖拽更新状态</p>
          </div>
        </div>

        <div class="step">
          <div class="step-number">4</div>
          <div class="step-content">
            <h3>分析数据</h3>
            <p>查看统计数据优化求职策略</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 用户指导模态框 -->
    <UserGuide
      :visible="showGuide"
      @close="closeGuide"
      @finished="onGuideFinished"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, h, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  PlayCircleOutlined,
  AppstoreOutlined
} from '@ant-design/icons-vue'
import UserGuide from '../components/UserGuide.vue'
import { UserGuideManager } from '../utils/userGuide'
import { useJobApplicationStore } from '../stores/jobApplication'

const router = useRouter()
const jobStore = useJobApplicationStore()

// State
const showGuide = ref(false)

// Computed
const totalApplications = computed(() => {
  return jobStore.applications?.length || 0
})

// Methods
const startGuide = () => {
  showGuide.value = true
}

const closeGuide = () => {
  showGuide.value = false
}

const onGuideFinished = () => {
  UserGuideManager.markGuideCompleted()
  showGuide.value = false
  message.success('欢迎使用 JobView！开始您的求职之旅吧！')
}

// 导航方法
const goToKanban = () => {
  router.push('/kanban')
}

const goToTimeline = () => {
  router.push('/timeline')
}

const goToStats = () => {
  router.push('/statistics')
}

const goToReminders = () => {
  router.push('/reminders')
}

const goToResume = () => {
  router.push('/resume')
}

// 生命周期
onMounted(async () => {
  // 尝试获取应用数据
  try {
    await jobStore.fetchApplications()
  } catch (error) {
    console.log('获取应用数据失败，可能是后端未连接')
  }

  // 检查是否应该自动显示用户指导
  if (UserGuideManager.shouldShowGuide()) {
    setTimeout(() => {
      showGuide.value = true
    }, 1000)
  }
})
</script>

<style scoped>
.homepage {
  min-height: calc(100vh - 120px);
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

/* 主页头部 */
.homepage-header {
  text-align: center;
  padding: 60px 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  position: relative;
  overflow: hidden;
}

.homepage-header::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="20" cy="20" r="2" fill="rgba(255,255,255,0.1)"/><circle cx="80" cy="20" r="2" fill="rgba(255,255,255,0.1)"/><circle cx="20" cy="80" r="2" fill="rgba(255,255,255,0.1)"/><circle cx="80" cy="80" r="2" fill="rgba(255,255,255,0.1)"/><circle cx="50" cy="50" r="2" fill="rgba(255,255,255,0.1)"/></svg>');
  animation: float 20s linear infinite;
  pointer-events: none;
}

@keyframes float {
  0% { transform: translate(-50%, -50%) rotate(0deg); }
  100% { transform: translate(-50%, -50%) rotate(360deg); }
}

.brand-section {
  position: relative;
  z-index: 1;
  margin-bottom: 40px;
}

.logo {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-bottom: 16px;
}

.logo-icon {
  font-size: 48px;
}

.logo h1 {
  margin: 0;
  font-size: 48px;
  font-weight: 700;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.tagline {
  font-size: 20px;
  margin: 0 0 8px 0;
  opacity: 0.9;
}

.description {
  font-size: 16px;
  margin: 0;
  opacity: 0.8;
}

.quick-actions {
  display: flex;
  gap: 16px;
  justify-content: center;
  flex-wrap: wrap;
  position: relative;
  z-index: 1;
}

/* 功能特色展示 */
.features-showcase {
  padding: 60px 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.features-showcase h2 {
  text-align: center;
  font-size: 32px;
  margin-bottom: 48px;
  color: #2d3748;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 24px;
}

.feature-card {
  background: white;
  border-radius: 16px;
  padding: 32px;
  text-align: center;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
  cursor: pointer;
  position: relative;
  overflow: hidden;
}

.feature-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(102, 126, 234, 0.1), transparent);
  transition: left 0.5s ease;
}

.feature-card:hover::before {
  left: 100%;
}

.feature-card:hover {
  transform: translateY(-8px);
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12);
}

.feature-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.feature-card h3 {
  font-size: 20px;
  margin-bottom: 12px;
  color: #2d3748;
}

.feature-card p {
  color: #718096;
  line-height: 1.6;
  margin: 0;
}

.feature-badge {
  position: absolute;
  top: 16px;
  right: 16px;
  background: #48bb78;
  color: white;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}

.feature-badge.new {
  background: #ed8936;
}

/* 使用统计 */
.stats-section {
  padding: 60px 24px;
  background: white;
}

.stats-section h2 {
  text-align: center;
  font-size: 32px;
  margin-bottom: 48px;
  color: #2d3748;
}

.stats-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 24px;
  max-width: 1000px;
  margin: 0 auto;
}

.stat-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 32px;
  border-radius: 16px;
  text-align: center;
  position: relative;
  overflow: hidden;
}

.stat-card::before {
  content: '';
  position: absolute;
  top: -50%;
  right: -50%;
  width: 100%;
  height: 100%;
  background: radial-gradient(circle, rgba(255,255,255,0.1) 0%, transparent 70%);
  animation: pulse 3s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { transform: scale(1); opacity: 0.5; }
  50% { transform: scale(1.1); opacity: 0.8; }
}

.stat-icon {
  font-size: 32px;
  margin-bottom: 16px;
}

.stat-number {
  font-size: 32px;
  font-weight: 700;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  opacity: 0.9;
}

/* 快速入门 */
.getting-started {
  padding: 60px 24px;
  max-width: 1000px;
  margin: 0 auto;
}

.getting-started h2 {
  text-align: center;
  font-size: 32px;
  margin-bottom: 48px;
  color: #2d3748;
}

.steps {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 32px;
}

.step {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  padding: 24px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.step-number {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 18px;
}

.step-content h3 {
  margin: 0 0 8px 0;
  color: #2d3748;
  font-size: 18px;
}

.step-content p {
  margin: 0;
  color: #718096;
  line-height: 1.6;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .homepage-header {
    padding: 40px 16px;
  }

  .logo h1 {
    font-size: 36px;
  }

  .logo-icon {
    font-size: 36px;
  }

  .tagline {
    font-size: 18px;
  }

  .features-showcase,
  .stats-section,
  .getting-started {
    padding: 40px 16px;
  }

  .features-showcase h2,
  .stats-section h2,
  .getting-started h2 {
    font-size: 24px;
    margin-bottom: 32px;
  }

  .features-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .feature-card {
    padding: 24px;
  }

  .quick-actions {
    flex-direction: column;
    align-items: center;
  }

  .steps {
    grid-template-columns: 1fr;
    gap: 24px;
  }

  .step {
    flex-direction: column;
    text-align: center;
  }
}

@media (max-width: 480px) {
  .feature-card {
    padding: 20px;
  }

  .stat-card {
    padding: 24px;
  }

  .step {
    padding: 20px;
  }
}
</style>