<template>
  <div class="user-guide-overlay" v-if="visible">
    <div class="guide-container">
      <!-- 指导页面头部 -->
      <div class="guide-header">
        <div class="logo-section">
          <div class="logo">
            <span class="logo-icon">📋</span>
            <h1>JobView</h1>
          </div>
          <p class="subtitle">智能求职投递管理系统</p>
        </div>

        <div class="progress-section">
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: `${(currentSlide + 1) / totalSlides * 100}%` }"></div>
          </div>
          <span class="progress-text">{{ currentSlide + 1 }} / {{ totalSlides }}</span>
        </div>
      </div>

      <!-- 指导页面内容 -->
      <div class="guide-content">
        <!-- 欢迎页面 -->
        <div v-if="currentSlide === 0" class="slide welcome-slide">
          <div class="slide-content">
            <div class="welcome-icon">🎉</div>
            <h2>欢迎使用 JobView!</h2>
            <p class="description">
              JobView 覆盖投递记录、流程推进、提醒复盘与数据洞察，<br>
              帮助你用一套工具掌控整个求职旅程。
            </p>
            <div class="features-preview">
              <div class="feature-item">
                <span class="feature-icon">📋</span>
                <span>看板流转</span>
              </div>
              <div class="feature-item">
                <span class="feature-icon">📅</span>
                <span>时间线追踪</span>
              </div>
              <div class="feature-item">
                <span class="feature-icon">🔔</span>
                <span>智能提醒</span>
              </div>
              <div class="feature-item">
                <span class="feature-icon">📊</span>
                <span>数据洞察</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 看板功能介绍 -->
        <div v-if="currentSlide === 1" class="slide kanban-slide">
          <div class="slide-content">
            <div class="feature-icon-large">📋</div>
            <h2>看板管理 - 拖拽式状态更新</h2>
            <p class="description">
              将所有岗位集中在看板上，拖拽即可推进状态，随时掌握当前进展
            </p>

            <div class="demo-kanban">
              <div class="kanban-column">
                <div class="column-header">已投递</div>
                <div class="demo-card">
                  <div class="company-name">阿里巴巴</div>
                  <div class="position">前端工程师</div>
                </div>
              </div>
              <div class="drag-arrow">➡️</div>
              <div class="kanban-column">
                <div class="column-header">简历筛选中</div>
                <div class="demo-card highlighted">
                  <div class="company-name">阿里巴巴</div>
                  <div class="position">前端工程师</div>
                </div>
              </div>
              <div class="drag-arrow">➡️</div>
              <div class="kanban-column">
                <div class="column-header">一面中</div>
                <div class="demo-card">
                  <div class="company-name">阿里巴巴</div>
                  <div class="position">前端工程师</div>
                </div>
              </div>
            </div>

            <div class="key-features">
              <div class="feature">✨ 拖拽流转配合进度条展示</div>
              <div class="feature">🔍 智能搜索定位并高亮结果</div>
              <div class="feature">📦 批量导入、导出与状态概览</div>
            </div>
          </div>
        </div>

        <!-- 时间线功能介绍 -->
        <div v-if="currentSlide === 2" class="slide timeline-slide">
          <div class="slide-content">
            <div class="feature-icon-large">📅</div>
            <h2>时间线视图 - 全流程追踪</h2>
            <p class="description">
              将投递记录按时间顺序展开，结合筛选条件快速回顾每一步进展
            </p>

            <div class="timeline-demo">
              <div class="timeline-column">
                <div class="timeline-item-card">
                  <div class="timeline-item-header">
                    <span class="timeline-company">阿里巴巴 · 前端工程师</span>
                    <span class="timeline-status in-progress">一面中</span>
                  </div>
                  <div class="timeline-meta">
                    <span>投递：4 月 8 日</span>
                    <span>最新更新：4 月 16 日</span>
                  </div>
                  <p class="timeline-note">与面试官交流前端工程化经验，等待反馈</p>
                </div>
                <div class="timeline-item-card">
                  <div class="timeline-item-header">
                    <span class="timeline-company">字节跳动 · 客户端</span>
                    <span class="timeline-status waiting">笔试中</span>
                  </div>
                  <div class="timeline-meta">
                    <span>投递：4 月 3 日</span>
                    <span>笔试：4 月 12 日</span>
                  </div>
                  <p class="timeline-note">系统提醒我笔试前 30 分钟准备</p>
                </div>
              </div>
            </div>

            <div class="key-features">
              <div class="feature">🎯 多条件筛选与关键字高亮</div>
              <div class="feature">📄 批量导入与 Excel 导出</div>
              <div class="feature">🧭 快速跳转编辑投递信息</div>
            </div>

            <div class="note">
              <strong>技巧：</strong> 筛选条件会联动顶部统计卡片，帮你即时聚焦当前关注的岗位集合。
            </div>
          </div>
        </div>

        <!-- 统计分析功能 -->
        <div v-if="currentSlide === 3" class="slide stats-slide">
          <div class="slide-content">
            <div class="feature-icon-large">📊</div>
            <h2>数据统计 - 多维度投递分析</h2>
            <p class="description">
              自动汇总状态分布、趋势与转化率，快速洞察求职策略的有效性
            </p>

            <div class="stats-demo">
              <div class="chart-container">
                <div class="chart-item">
                  <div class="chart-title">状态分布</div>
                  <div class="pie-chart">
                    <div class="pie-slice slice-1" style="--percentage: 35%"></div>
                    <div class="pie-slice slice-2" style="--percentage: 25%"></div>
                    <div class="pie-slice slice-3" style="--percentage: 40%"></div>
                  </div>
                </div>

                <div class="chart-item">
                  <div class="chart-title">投递趋势</div>
                  <div class="line-chart">
                    <div class="chart-line"></div>
                  </div>
                </div>
              </div>

              <div class="stats-cards">
                <div class="stat-card">
                  <div class="stat-number">124</div>
                  <div class="stat-label">总投递量</div>
                </div>
                <div class="stat-card">
                  <div class="stat-number">23</div>
                  <div class="stat-label">面试邀请</div>
                </div>
                <div class="stat-card">
                  <div class="stat-number">18.5%</div>
                  <div class="stat-label">面试成功率</div>
                </div>
              </div>
            </div>

            <div class="key-features">
              <div class="feature">📈 趋势曲线与状态分布一目了然</div>
              <div class="feature">🎯 各阶段通过率智能计算</div>
              <div class="feature">📤 一键导出统计报告与历史版本</div>
            </div>
          </div>
        </div>

        <!-- 提醒中心功能 -->
        <div v-if="currentSlide === 4" class="slide reminders-slide">
          <div class="slide-content">
            <div class="feature-icon-large">🔔</div>
            <h2>提醒中心 - 智能跟进提醒</h2>
            <p class="description">
              将提醒统一管理，结合邮箱授权与自动策略，确保不遗漏每一次面试
            </p>

            <div class="reminders-demo">
              <div class="reminder-card urgent">
                <div class="reminder-icon">⚠️</div>
                <div class="reminder-content">
                  <div class="reminder-title">面试提醒</div>
                  <div class="reminder-text">明天 14:00 腾讯终面，请提前准备</div>
                </div>
                <div class="reminder-time">1 小时后提醒</div>
              </div>

              <div class="reminder-card normal">
                <div class="reminder-icon">📞</div>
                <div class="reminder-content">
                  <div class="reminder-title">跟进提醒</div>
                  <div class="reminder-text">阿里内推 7 天未反馈，建议主动联系</div>
                </div>
                <div class="reminder-time">2 天前创建</div>
              </div>

              <div class="reminder-card info">
                <div class="reminder-icon">📋</div>
                <div class="reminder-content">
                  <div class="reminder-title">状态更新</div>
                  <div class="reminder-text">状态流转到二面，系统自动同步提醒</div>
                </div>
                <div class="reminder-time">系统生成</div>
              </div>
            </div>

            <div class="key-features">
              <div class="feature">⏰ 默认提前提醒时间可配置</div>
              <div class="feature">📬 支持 IMAP 邮箱授权同步</div>
              <div class="feature">🤖 面试状态触发自动提醒</div>
            </div>
          </div>
        </div>

        <!-- 其他亮眼功能 -->
        <div v-if="currentSlide === 5" class="slide features-slide">
          <div class="slide-content">
            <div class="feature-icon-large">🧰</div>
            <h2>更多实用模块</h2>
            <p class="description">
              除核心流程外，JobView 还提供多个辅助模块，帮助你形成完整的求职档案
            </p>

            <div class="features-grid">
              <div class="feature-card">
                <div class="card-icon">🗂️</div>
                <h3>投递详情页</h3>
                <p>随时查看岗位字段、更新时间并支持快速编辑</p>
              </div>

              <div class="feature-card">
                <div class="card-icon">🧠</div>
                <h3>面试体验记录</h3>
                <p>状态推进时记录面试反馈，沉淀复盘素材</p>
              </div>

              <div class="feature-card">
                <div class="card-icon">📝</div>
                <h3>简历管理</h3>
                <p>分块维护简历信息，支持上传 PDF 附件</p>
              </div>

              <div class="feature-card">
                <div class="card-icon">📦</div>
                <h3>批量工具</h3>
                <p>看板与时间线共享导入、导出能力，便于备份</p>
              </div>

              <div class="feature-card">
                <div class="card-icon">📬</div>
                <h3>邮箱授权</h3>
                <p>配置 IMAP 授权码后，可同步邮件提醒</p>
              </div>

              <div class="feature-card">
                <div class="card-icon">🔄</div>
                <h3>统一数据</h3>
                <p>看板、时间线、提醒中心共享同一份投递数据</p>
              </div>
            </div>
          </div>
        </div>

        <!-- 开始使用页面 -->
        <div v-if="currentSlide === 6" class="slide start-slide">
          <div class="slide-content">
            <div class="feature-icon-large">🎯</div>
            <h2>准备好开始新的求职节奏了吗？</h2>
            <p class="description">
              选择一个入口开始操作，系统会在后台同步所有信息
            </p>

            <div class="start-actions">
              <div class="action-card" @click="goToKanban">
                <div class="action-icon">📋</div>
                <h3>开始管理投递</h3>
                <p>使用看板管理您的求职进度</p>
              </div>

              <div class="action-card" @click="goToTimeline">
                <div class="action-icon">📅</div>
                <h3>回顾历史</h3>
                <p>结合筛选条件复盘过往投递</p>
              </div>

              <div class="action-card" @click="goToStats">
                <div class="action-icon">📊</div>
                <h3>查看数据分析</h3>
                <p>用实时统计检验求职策略</p>
              </div>
            </div>

            <div class="start-tip">
              <p>💡 小贴士：在首页或设置入口可以重新唤起本指南，随时复习关键功能。</p>
            </div>
          </div>
        </div>
      </div>

      <!-- 导航控制 -->
      <div class="guide-footer">
        <div class="nav-buttons">
          <button
            class="nav-btn prev-btn"
            @click="prevSlide"
            :disabled="currentSlide === 0"
          >
            ← 上一步
          </button>

          <div class="slide-indicators">
            <span
              v-for="index in totalSlides"
              :key="index"
              class="indicator"
              :class="{ active: currentSlide === index - 1 }"
              @click="goToSlide(index - 1)"
            ></span>
          </div>

          <button
            v-if="currentSlide < totalSlides - 1"
            class="nav-btn next-btn"
            @click="nextSlide"
          >
            下一步 →
          </button>

          <button
            v-else
            class="nav-btn finish-btn"
            @click="finishGuide"
          >
            开始使用 🚀
          </button>
        </div>

        <div class="footer-actions">
          <button class="skip-btn" @click="finishGuide">
            跳过指导
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'

// Props
interface Props {
  visible: boolean
}

const props = withDefaults(defineProps<Props>(), {
  visible: false
})

// Emits
const emit = defineEmits<{
  close: []
  finished: []
}>()

// Router
const router = useRouter()

// 指导状态
const currentSlide = ref(0)
const totalSlides = computed(() => 7)

// 导航方法
const nextSlide = () => {
  if (currentSlide.value < totalSlides.value - 1) {
    currentSlide.value++
  }
}

const prevSlide = () => {
  if (currentSlide.value > 0) {
    currentSlide.value--
  }
}

const goToSlide = (index: number) => {
  if (index >= 0 && index < totalSlides.value) {
    currentSlide.value = index
  }
}

// 完成指导
const finishGuide = () => {
  emit('finished')
  emit('close')
}

// 快速跳转方法
const goToKanban = () => {
  finishGuide()
  router.push('/kanban')
}

const goToTimeline = () => {
  finishGuide()
  router.push('/timeline')
}

const goToStats = () => {
  finishGuide()
  router.push('/statistics')
}

// 键盘事件监听
const handleKeydown = (event: KeyboardEvent) => {
  if (!props.visible) return

  switch (event.key) {
    case 'ArrowLeft':
      prevSlide()
      break
    case 'ArrowRight':
      nextSlide()
      break
    case 'Escape':
      finishGuide()
      break
  }
}

// 生命周期
import { onMounted, onUnmounted } from 'vue'

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.user-guide-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
  padding: 20px;
}

.guide-container {
  background: white;
  border-radius: 24px;
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.15);
  width: 100%;
  max-width: 900px;
  height: 90vh;
  max-height: 700px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 头部样式 */
.guide-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 24px 32px;
  display: flex;
  justify-content: flex-start;
  align-items: center;
  position: relative;
}

.logo-section {
  display: flex;
  flex-direction: column;
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.logo-icon {
  font-size: 28px;
}

.logo h1 {
  margin: 0;
  font-size: 28px;
  font-weight: 700;
}

.subtitle {
  margin: 0;
  opacity: 0.9;
  font-size: 16px;
}

.progress-section {
  position: absolute;
  right: 32px;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8px;
}

.progress-bar {
  width: 200px;
  height: 6px;
  background: rgba(255, 255, 255, 0.3);
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: white;
  border-radius: 3px;
  transition: width 0.3s ease;
}

.progress-text {
  font-size: 14px;
  opacity: 0.9;
}

/* 内容样式 */
.guide-content {
  flex: 1;
  padding: 40px 48px 40px 40px;
  overflow-y: auto;
  display: flex;
  align-items: center;
  justify-content: center;
}

.slide {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 20px 0;
}

.slide-content {
  width: 100%;
  max-width: 600px;
  text-align: center;
  padding: 20px 0;
}

/* 通用样式 */
.feature-icon-large {
  font-size: 48px;
  margin-bottom: 24px;
}

.slide h2 {
  font-size: 28px;
  margin-bottom: 16px;
  color: #2c3e50;
}

.description {
  font-size: 16px;
  color: #7f8c8d;
  margin-bottom: 20px;
  line-height: 1.6;
}

/* 欢迎页面样式 */
.welcome-icon {
  font-size: 64px;
  margin-bottom: 24px;
}

.features-preview {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-top: 32px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 12px;
  font-weight: 500;
}

.feature-icon {
  font-size: 20px;
}

/* 看板演示样式 */
.demo-kanban {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin: 32px 0;
  padding: 24px;
  background: #f8f9fa;
  border-radius: 16px;
}

.kanban-column {
  background: white;
  border-radius: 8px;
  padding: 12px;
  min-width: 120px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.column-header {
  font-weight: 600;
  margin-bottom: 8px;
  color: #2c3e50;
  font-size: 12px;
}

.demo-card {
  background: #e3f2fd;
  border: 1px solid #bbdefb;
  border-radius: 6px;
  padding: 8px;
  font-size: 12px;
}

.demo-card.highlighted {
  background: #fff3e0;
  border-color: #ffcc02;
  animation: pulse 2s infinite;
}

.company-name {
  font-weight: 600;
  color: #1565c0;
}

.position {
  color: #757575;
  margin-top: 4px;
}

.drag-arrow {
  font-size: 20px;
  color: #4caf50;
  animation: bounce 2s infinite;
}

/* 插件演示样式 */
.timeline-demo {
  margin: 32px 0;
  display: flex;
  justify-content: center;
}

.timeline-column {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
  max-width: 420px;
}

.timeline-item-card {
  background: #f8f9fa;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  text-align: left;
}

.timeline-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  gap: 12px;
}

.timeline-company {
  font-weight: 600;
  color: #2c3e50;
}

.timeline-status {
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 999px;
  font-weight: 500;
}

.timeline-status.in-progress {
  background: rgba(102, 126, 234, 0.15);
  color: #4c51bf;
}

.timeline-status.waiting {
  background: rgba(250, 173, 20, 0.18);
  color: #d48806;
}

.timeline-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  font-size: 13px;
  color: #718096;
  margin-bottom: 8px;
}

.timeline-note {
  margin: 0;
  font-size: 14px;
  line-height: 1.6;
  color: #4a5568;
}

/* 统计演示样式 */
.stats-demo {
  margin: 32px 0;
}

.chart-container {
  display: flex;
  gap: 24px;
  justify-content: center;
  margin-bottom: 24px;
}

.chart-item {
  text-align: center;
}

.chart-title {
  font-weight: 600;
  margin-bottom: 12px;
  color: #2c3e50;
}

.pie-chart {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: conic-gradient(
    #667eea 0% 35%,
    #48bb78 35% 60%,
    #ed8936 60% 100%
  );
  margin: 0 auto;
}

.line-chart {
  width: 120px;
  height: 60px;
  position: relative;
  background: #f8f9fa;
  border-radius: 6px;
  margin: 0 auto;
  overflow: hidden;
}

.chart-line {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 70%;
  background: linear-gradient(45deg, #667eea, #48bb78);
  clip-path: polygon(0 100%, 20% 80%, 40% 60%, 60% 40%, 80% 20%, 100% 30%, 100% 100%);
}

.stats-cards {
  display: flex;
  gap: 16px;
  justify-content: center;
}

.stat-card {
  background: #f8f9fa;
  padding: 16px;
  border-radius: 12px;
  text-align: center;
  min-width: 80px;
}

.stat-number {
  font-size: 20px;
  font-weight: 700;
  color: #667eea;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: #7f8c8d;
}

/* 提醒演示样式 */
.reminders-demo {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin: 32px 0;
  max-width: 400px;
  margin-left: auto;
  margin-right: auto;
}

.reminder-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  border-radius: 12px;
  text-align: left;
}

.reminder-card.urgent {
  background: #fff5f5;
  border: 1px solid #fed7d7;
}

.reminder-card.normal {
  background: #f0fff4;
  border: 1px solid #c6f6d5;
}

.reminder-card.info {
  background: #ebf8ff;
  border: 1px solid #bee3f8;
}

.reminder-icon {
  font-size: 20px;
}

.reminder-content {
  flex: 1;
}

.reminder-title {
  font-weight: 600;
  margin-bottom: 4px;
  color: #2c3e50;
}

.reminder-text {
  font-size: 14px;
  color: #7f8c8d;
}

.reminder-time {
  font-size: 12px;
  color: #a0aec0;
}

/* 功能网格样式 */
.features-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin: 20px 0;
}

.feature-card {
  background: #f8f9fa;
  padding: 16px;
  border-radius: 12px;
  text-align: center;
}

.card-icon {
  font-size: 32px;
  margin-bottom: 12px;
}

.feature-card h3 {
  margin: 0 0 8px 0;
  color: #2c3e50;
  font-size: 16px;
}

.feature-card p {
  margin: 0;
  color: #7f8c8d;
  font-size: 14px;
  line-height: 1.4;
}

/* 开始页面样式 */
.start-actions {
  display: flex;
  gap: 16px;
  justify-content: center;
  margin: 32px 0;
  flex-wrap: wrap;
}

.action-card {
  background: #f8f9fa;
  border: 2px solid transparent;
  border-radius: 16px;
  padding: 24px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s ease;
  min-width: 160px;
}

.action-card:hover {
  border-color: #667eea;
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(102, 126, 234, 0.2);
}

.action-icon {
  font-size: 32px;
  margin-bottom: 12px;
}

.action-card h3 {
  margin: 0 0 8px 0;
  color: #2c3e50;
  font-size: 16px;
}

.action-card p {
  margin: 0;
  color: #7f8c8d;
  font-size: 14px;
}

.start-tip {
  margin-top: 32px;
  padding: 16px;
  background: #f0f7ff;
  border-radius: 12px;
  font-size: 14px;
  color: #4a5568;
}

/* 通用功能样式 */
.key-features {
  display: flex;
  gap: 24px;
  justify-content: center;
  margin-top: 24px;
  flex-wrap: wrap;
}

.feature {
  font-size: 14px;
  color: #4a5568;
  font-weight: 500;
}

.note {
  margin-top: 24px;
  padding: 16px;
  background: #fff7ed;
  border: 1px solid #fed7aa;
  border-radius: 12px;
  font-size: 14px;
  color: #9a3412;
}

/* 底部导航样式 */
.guide-footer {
  background: #f8f9fa;
  padding: 24px 32px;
  border-top: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.nav-buttons {
  display: flex;
  align-items: center;
  gap: 24px;
}

.nav-btn {
  background: #667eea;
  color: white;
  border: none;
  padding: 12px 24px;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.3s ease;
}

.nav-btn:hover:not(:disabled) {
  background: #5a67d8;
  transform: translateY(-1px);
}

.nav-btn:disabled {
  background: #e2e8f0;
  color: #a0aec0;
  cursor: not-allowed;
}

.finish-btn {
  background: #48bb78;
}

.finish-btn:hover {
  background: #38a169;
}

.slide-indicators {
  display: flex;
  gap: 8px;
}

.indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #cbd5e0;
  cursor: pointer;
  transition: all 0.3s ease;
}

.indicator.active {
  background: #667eea;
  width: 24px;
  border-radius: 4px;
}

.skip-btn {
  background: none;
  border: none;
  color: #a0aec0;
  cursor: pointer;
  font-size: 14px;
  transition: color 0.3s ease;
}

.skip-btn:hover {
  color: #718096;
}

/* 动画 */
@keyframes pulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.05); }
}

@keyframes bounce {
  0%, 100% { transform: translateX(0); }
  50% { transform: translateX(3px); }
}

/* 响应式设计 */
@media (max-width: 768px) {
  .guide-container {
    height: 95vh;
    margin: 0;
    border-radius: 0;
  }

  .guide-header {
    padding: 16px 20px;
  }

  .progress-section {
    right: 20px;
  }

  .progress-bar {
    width: 150px;
  }

  .guide-content {
    padding: 24px 20px;
  }

  .logo h1 {
    font-size: 24px;
  }

  .features-preview,
  .features-grid {
    grid-template-columns: 1fr;
  }

  .chart-container {
    flex-direction: column;
    align-items: center;
    gap: 16px;
  }

  .stats-cards {
    flex-wrap: wrap;
  }

  .start-actions {
    flex-direction: column;
    align-items: center;
  }

  .demo-kanban {
    flex-direction: column;
    gap: 8px;
  }
  
  .timeline-demo {
    margin: 24px 0;
  }

  .timeline-column {
    max-width: 100%;
  }

  .drag-arrow {
    transform: rotate(90deg);
  }

  .key-features {
    flex-direction: column;
    text-align: center;
    gap: 12px;
  }
}

@media (max-width: 480px) {
  .guide-footer {
    flex-direction: column;
    gap: 16px;
  }

  .nav-buttons {
    gap: 12px;
  }

  .nav-btn {
    padding: 10px 16px;
    font-size: 14px;
  }
}

/* 自定义滚动条样式 - 让滚动条更靠右边 */
.guide-content::-webkit-scrollbar {
  width: 8px;
}

.guide-content::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.05);
  border-radius: 4px;
  margin: 4px;
}

.guide-content::-webkit-scrollbar-thumb {
  background: rgba(102, 126, 234, 0.3);
  border-radius: 4px;
  transition: all 0.3s ease;
}

.guide-content::-webkit-scrollbar-thumb:hover {
  background: rgba(102, 126, 234, 0.5);
}

.guide-content::-webkit-scrollbar-corner {
  background: transparent;
}
</style>
