<template>
  <div class="status-tracking-view">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <div class="header-title">
          <h1>
            <DashboardOutlined />
            状态跟踪分析
          </h1>
          <p>全面掌握求职进展，智能分析流程数据</p>
        </div>
        <div class="header-actions">
          <a-space>
            <a-button @click="refreshData" :loading="loading">
              <template #icon><ReloadOutlined /></template>
              刷新数据
            </a-button>
            <a-button type="primary" @click="showSettingsModal = true">
              <template #icon><SettingOutlined /></template>
              偏好设置
            </a-button>
          </a-space>
        </div>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-overview">
      <a-row :gutter="[16, 16]">
        <a-col 
          v-for="card in statusStatsCards" 
          :key="card.title"
          :xs="24" :sm="12" :md="6"
        >
          <a-card 
            :loading="analyticsLoading"
            class="stats-card"
            :class="`stats-card-${card.color.replace('#', '')}`"
          >
            <div class="card-content">
              <div class="card-icon" :style="{ color: card.color }">
                <component :is="getIconComponent(card.icon)" />
              </div>
              <div class="card-info">
                <div class="card-value">{{ card.value }}</div>
                <div class="card-title">{{ card.title }}</div>
                <div v-if="card.trend" class="card-trend">
                  <span 
                    class="trend-indicator" 
                    :class="`trend-${card.trend.direction}`"
                  >
                    <component :is="getTrendIcon(card.trend.direction)" />
                    {{ card.trend.value }}
                  </span>
                  <span class="trend-period">{{ card.trend.period }}</span>
                </div>
              </div>
            </div>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- 主要内容区域 -->
    <div class="main-content">
      <a-row :gutter="24">
        <!-- 左侧：状态分布 -->
        <a-col :xs="24" :lg="12">
          <!-- 状态分布图 -->
          <a-card 
            title="状态分布" 
            class="chart-card distribution-chart"
            :loading="analyticsLoading"
          >
            <template #extra>
              <a-select
                v-model:value="chartTimeRange"
                size="small"
                style="width: 100px"
                @change="handleTimeRangeChange"
              >
                <a-select-option value="week">本周</a-select-option>
                <a-select-option value="month">本月</a-select-option>
                <a-select-option value="quarter">本季</a-select-option>
              </a-select>
            </template>
            
            <div ref="distributionChartRef" style="height: 350px;"></div>
          </a-card>
        </a-col>

        <!-- 右侧：智能洞察 -->
        <a-col :xs="24" :lg="12">
          <!-- 智能洞察 -->
          <a-card 
            title="智能洞察" 
            class="insights-card"
            :loading="insightsLoading"
          >
            <template #extra>
              <a-button type="link" size="small" @click="refreshInsights">
                <ReloadOutlined />
              </a-button>
            </template>
            
            <div class="insights-content">
              <div 
                v-for="(insight, index) in processInsights.slice(0, 6)"
                :key="index"
                class="insight-item"
              >
                <a-alert
                  :type="getAlertType(insight.type)"
                  :message="insight.title"
                  :description="insight.description"
                  :show-icon="true"
                  class="insight-alert"
                />
              </div>
              <a-empty 
                v-if="processInsights.length === 0"
                description="暂无洞察数据"
                :image="'simple'"
              />
            </div>
          </a-card>
        </a-col>
      </a-row>

      <a-row :gutter="24" style="margin-top: 24px;">
        <!-- 趋势分析图 - 全宽 -->
        <a-col :span="24">
          <a-card 
            title="申请趋势" 
            class="chart-card trend-chart"
            :loading="trendsLoading"
          >
            <template #extra>
              <a-space>
                <a-checkbox v-model:checked="showSuccessRate">成功率</a-checkbox>
                <a-checkbox v-model:checked="showApplicationCount">申请量</a-checkbox>
              </a-space>
            </template>
            
            <div ref="trendsChartRef" style="height: 300px;"></div>
          </a-card>
        </a-col>
      </a-row>

      <a-row :gutter="24" style="margin-top: 24px;">
        <!-- 最新活动 -->
        <a-col :xs="24" :lg="12">
          <a-card 
            title="最新活动" 
            class="activity-card"
            :loading="dashboardLoading"
          >
            <div class="activity-list">
              <div
                v-for="activity in recentActivities.slice(0, 8)"
                :key="`${activity.application_id}_${activity.timestamp}`"
                class="activity-item"
              >
                <div class="activity-icon">
                  <component :is="getActivityIcon(activity.new_status)" />
                </div>
                <div class="activity-content">
                  <div class="activity-title">
                    {{ activity.company_name }} - {{ activity.position_title }}
                  </div>
                  <div class="activity-desc">
                    状态从「{{ activity.old_status }}」更新为
                    <a-tag :color="StatusHelper.getStatusColor(activity.new_status)">
                      {{ activity.new_status }}
                    </a-tag>
                  </div>
                  <div class="activity-time">
                    {{ formatTimestamp(activity.timestamp, 'MM-DD HH:mm') }}
                  </div>
                </div>
              </div>
              <a-empty 
                v-if="recentActivities.length === 0"
                description="暂无活动记录"
                :image="'simple'"
              />
            </div>
          </a-card>
        </a-col>

        <!-- 待办提醒 -->
        <a-col :xs="24" :lg="12">
          <a-card 
            title="待办提醒" 
            class="todo-card"
            :loading="dashboardLoading"
          >
            <div class="todo-list">
              <div
                v-for="todo in upcomingActions.slice(0, 8)"
                :key="`${todo.application_id}_${todo.type}_${todo.scheduled_time}`"
                class="todo-item"
                :class="`todo-${todo.priority}`"
              >
                <div class="todo-icon">
                  <component :is="getTodoIcon(todo.type)" />
                </div>
                <div class="todo-content">
                  <div class="todo-title">
                    {{ todo.company_name }} - {{ getTodoTitle(todo.type) }}
                  </div>
                  <div class="todo-time">
                    {{ formatTimestamp(todo.scheduled_time, 'MM-DD HH:mm') }}
                  </div>
                </div>
                <div class="todo-actions">
                  <a-button 
                    type="link" 
                    size="small"
                    @click="markTodoCompleted(todo)"
                  >
                    完成
                  </a-button>
                </div>
              </div>
              <a-empty 
                v-if="upcomingActions.length === 0"
                description="暂无待办事项"
                :image="'simple'"
              />
            </div>
          </a-card>
        </a-col>
      </a-row>

      <!-- 申请成功率分析 -->
      <a-row :gutter="24" style="margin-top: 24px;">
        <a-col :span="24">
          <a-card 
            title="申请成功率" 
            class="chart-card success-rate-chart"
            :loading="analyticsLoading"
          >
            <template #extra>
              <a-space>
                <a-tag color="green">成功率</a-tag>
                <a-tag color="orange">申请量</a-tag>
              </a-space>
            </template>
            
            <div class="success-rate-content">
              <!-- 成功率柱状图区域预留 -->
              <div class="success-rate-placeholder">
                <a-empty description="申请成功率图表开发中" />
              </div>
            </div>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- 用户偏好设置弹窗 -->
    <UserPreferencesModal
      v-model:visible="showSettingsModal"
      @updated="handlePreferencesUpdated"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import {
  DashboardOutlined,
  ReloadOutlined,
  SettingOutlined,
  FileTextOutlined,
  ClockCircleOutlined,
  TrophyOutlined,
  FieldTimeOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
  MinusOutlined,
  SendOutlined,
  EyeOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  CalendarOutlined,
  PhoneOutlined,
  BellOutlined
} from '@ant-design/icons-vue'
import { useStatusTrackingStore } from '../stores/statusTracking'
import { StatusHelper, type ApplicationStatus } from '../types'
import UserPreferencesModal from '../components/UserPreferencesModal.vue'
import * as echarts from 'echarts'

// Store
const statusTrackingStore = useStatusTrackingStore()

// 响应式数据
const loading = ref(false)
const showSettingsModal = ref(false)
const chartTimeRange = ref('month')
const showSuccessRate = ref(true)
const showApplicationCount = ref(true)
const distributionChartRef = ref<HTMLElement>()
const trendsChartRef = ref<HTMLElement>()
const distributionChart = ref<echarts.ECharts>()
const trendsChart = ref<echarts.ECharts>()
const trendsLoading = ref(false)
const insightsLoading = ref(false)

// 计算属性
const { 
  statusStatsCards, 
  statusDistributionData, 
  processInsights,
  analyticsLoading,
  dashboardLoading
} = statusTrackingStore

const recentActivities = computed(() => {
  return statusTrackingStore.dashboardData?.recent_activity || []
})

const upcomingActions = computed(() => {
  return statusTrackingStore.dashboardData?.upcoming_actions || []
})

// 图标组件映射
const iconComponents = {
  FileTextOutlined,
  ClockCircleOutlined, 
  TrophyOutlined,
  FieldTimeOutlined,
  SendOutlined,
  EyeOutlined,
  CheckCircleOutlined,
  CalendarOutlined,
  PhoneOutlined,
  BellOutlined
}

// 方法
const getIconComponent = (iconName: string) => {
  return (iconComponents as any)[iconName] || FileTextOutlined
}

const getTrendIcon = (direction: string) => {
  const iconMap = {
    up: ArrowUpOutlined,
    down: ArrowDownOutlined,
    stable: MinusOutlined
  }
  return (iconMap as any)[direction] || MinusOutlined
}

const getAlertType = (type: string): 'success' | 'info' | 'warning' | 'error' => {
  const typeMap: Record<string, 'success' | 'info' | 'warning' | 'error'> = {
    success: 'success',
    info: 'info', 
    warning: 'warning',
    error: 'error'
  }
  return typeMap[type] || 'info'
}

const getActivityIcon = (status: ApplicationStatus) => {
  if (StatusHelper.isFailedStatus(status)) return CloseCircleOutlined
  if (StatusHelper.isPassedStatus(status)) return CheckCircleOutlined
  return ClockCircleOutlined
}

const getTodoIcon = (type: string) => {
  const iconMap = {
    interview: CalendarOutlined,
    follow_up: PhoneOutlined,
    reminder: BellOutlined
  }
  return (iconMap as any)[type] || BellOutlined
}

const getTodoTitle = (type: string): string => {
  const titleMap: Record<string, string> = {
    interview: '面试提醒',
    follow_up: '跟进提醒', 
    reminder: '其他提醒'
  }
  return titleMap[type] || '提醒'
}

const initDistributionChart = async () => {
  // 等待DOM稳定，避免在loading骨架期间初始化
  await nextTick()
  const el = distributionChartRef.value as HTMLElement | undefined
  if (!el || !el.isConnected) return

  if (distributionChart.value) {
    distributionChart.value.dispose()
  }

  distributionChart.value = echarts.init(el)

    const option = {
      tooltip: {
        trigger: 'item',
        formatter: '{a} <br/>{b}: {c} ({d}%)'
      },
      legend: {
        orient: 'vertical',
        left: 10,
        data: statusDistributionData.map(item => item.name)
      },
      series: [
        {
          name: '状态分布',
          type: 'pie',
          radius: '55%',
          center: ['60%', '50%'],
          data: statusDistributionData,
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowOffsetX: 0,
              shadowColor: 'rgba(0, 0, 0, 0.5)'
            }
          }
        }
      ]
    }

    distributionChart.value.setOption(option)
}

const initTrendsChart = async () => {
  // 开启loading以显示骨架
  trendsLoading.value = true
  try {
    const trendsData = await statusTrackingStore.fetchStatusTrends({
      period: chartTimeRange.value as any
    })

    // 关闭loading后再等待DOM回流，确保图表容器已渲染
    trendsLoading.value = false
    await nextTick()

    const el = trendsChartRef.value as HTMLElement | undefined
    if (!el || !el.isConnected) {
      return
    }

    if (trendsChart.value) {
      trendsChart.value.dispose()
    }

    trendsChart.value = echarts.init(el)

      const dates = trendsData.trends.map(item => item.date)
      const applicationCounts = trendsData.trends.map(item => item.total_applications)
      const successRates = trendsData.trends.map(item => item.success_rate * 100)

      const series = []

      if (showApplicationCount.value) {
        series.push({
          name: '申请数量',
          type: 'bar',
          yAxisIndex: 0,
          data: applicationCounts
        })
      }

      if (showSuccessRate.value) {
        series.push({
          name: '成功率(%)',
          type: 'line',
          yAxisIndex: 1,
          data: successRates
        })
      }

      const option = {
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'cross'
          }
        },
        legend: {
          data: series.map(s => s.name)
        },
        xAxis: {
          type: 'category',
          data: dates
        },
        yAxis: [
          {
            type: 'value',
            name: '申请数量',
            position: 'left'
          },
          {
            type: 'value',
            name: '成功率(%)',
            position: 'right',
            max: 100
          }
        ],
        series: series
      }

      trendsChart.value.setOption(option)
  } finally {
    // 若前面已置为false，这里保持幂等
    trendsLoading.value = false
  }
}

const refreshData = async () => {
  loading.value = true
  try {
    await Promise.all([
      statusTrackingStore.fetchAnalytics(true),
      statusTrackingStore.fetchDashboardData(true)
    ])
    await nextTick()
    initDistributionChart()
    initTrendsChart()
  } finally {
    loading.value = false
  }
}

const refreshInsights = async () => {
  insightsLoading.value = true
  try {
    await statusTrackingStore.fetchProcessInsights()
  } finally {
    insightsLoading.value = false
  }
}

const handleTimeRangeChange = () => {
  initTrendsChart()
}

const handlePreferencesUpdated = () => {
  refreshData()
}

const markTodoCompleted = (todo: any) => {
  // TODO: 实现待办事项完成逻辑
  console.log('标记完成:', todo)
}

// 工具方法
const { formatTimestamp } = statusTrackingStore

// 监听器
watch([showSuccessRate, showApplicationCount], () => {
  initTrendsChart()
})

// 生命周期
onMounted(async () => {
  await refreshData()
})
</script>

<style scoped>
.status-tracking-view {
  padding: 24px;
  background: #f5f5f5;
  min-height: 100vh;
}

.page-header {
  background: #fff;
  border-radius: 8px;
  padding: 24px;
  margin-bottom: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.header-title h1 {
  margin: 0 0 8px 0;
  color: #262626;
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-title p {
  margin: 0;
  color: #8c8c8c;
}

.stats-overview {
  margin-bottom: 24px;
}

.stats-card {
  border-radius: 8px;
  transition: all 0.3s ease;
}

.stats-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.card-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.card-icon {
  font-size: 28px;
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.1);
}

.card-info {
  flex: 1;
}

.card-value {
  font-size: 24px;
  font-weight: 600;
  color: #262626;
  margin-bottom: 4px;
}

.card-title {
  color: #8c8c8c;
  font-size: 13px;
  margin-bottom: 4px;
}

.card-trend {
  display: flex;
  align-items: center;
  gap: 8px;
}

.trend-indicator {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  font-weight: 500;
}

.trend-up {
  color: #52c41a;
}

.trend-down {
  color: #ff4d4f;
}

.trend-stable {
  color: #faad14;
}

.trend-period {
  color: #8c8c8c;
  font-size: 11px;
}

.main-content {
  /* 移除flex布局，改为普通块级元素 */
}

.chart-card {
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.distribution-chart {
  height: 100%;
}

.trend-chart {
  margin-bottom: 0;
}

.success-rate-chart {
  margin-bottom: 0;
}

.insights-card,
.activity-card,
.todo-card {
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  height: 100%;
}

.insights-content {
  max-height: 380px;
  overflow-y: auto;
}

.success-rate-content {
  height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.success-rate-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.insight-item {
  margin-bottom: 12px;
}

.insight-alert {
  margin-bottom: 0;
}

.activity-list,
.todo-list {
  max-height: 380px;
  overflow-y: auto;
}

.activity-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.activity-item:last-child {
  border-bottom: none;
}

.activity-icon {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0f9ff;
  color: #1890ff;
  flex-shrink: 0;
}

.activity-content {
  flex: 1;
}

.activity-title {
  font-weight: 500;
  color: #262626;
  margin-bottom: 4px;
}

.activity-desc {
  color: #595959;
  font-size: 13px;
  margin-bottom: 4px;
}

.activity-time {
  color: #8c8c8c;
  font-size: 11px;
}

.todo-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 6px;
  margin-bottom: 8px;
  border: 1px solid #f0f0f0;
}

.todo-high {
  border-color: #ff4d4f;
  background: #fff2f0;
}

.todo-medium {
  border-color: #faad14;
  background: #fffbe6;
}

.todo-low {
  border-color: #52c41a;
  background: #f6ffed;
}

.todo-icon {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #1890ff;
  color: white;
  flex-shrink: 0;
}

.todo-content {
  flex: 1;
}

.todo-title {
  font-weight: 500;
  color: #262626;
  margin-bottom: 2px;
}

.todo-time {
  color: #8c8c8c;
  font-size: 11px;
}

.todo-actions {
  flex-shrink: 0;
}

/* 响应式适配 */
@media (max-width: 1200px) {
  .insights-content,
  .activity-list,
  .todo-list {
    max-height: 300px;
  }
}

@media (max-width: 768px) {
  .status-tracking-view {
    padding: 16px;
  }

  .header-content {
    flex-direction: column;
    gap: 16px;
    align-items: flex-start;
  }

  .card-content {
    flex-direction: column;
    text-align: center;
    gap: 8px;
  }

  .card-icon {
    align-self: center;
  }

  .activity-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
}
</style>
