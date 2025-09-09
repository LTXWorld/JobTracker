<template>
  <a-modal
    v-model:visible="visible"
    :title="modalTitle"
    width="900px"
    :mask-closable="false"
    :footer="null"
    class="status-detail-modal"
    @cancel="handleCancel"
  >
    <div class="modal-content">
      <!-- 头部信息 -->
      <div class="application-header">
        <div class="header-info">
          <h3>{{ applicationData?.company_name }} - {{ applicationData?.position_title }}</h3>
          <div class="meta-info">
            <a-tag :color="StatusHelper.getStatusColor(currentStatus)">
              {{ currentStatus }}
            </a-tag>
            <span class="apply-date">
              投递日期：{{ formatTimestamp(applicationData?.application_date || '', 'YYYY-MM-DD') }}
            </span>
          </div>
        </div>
        <div class="header-actions">
          <StatusQuickUpdate
            v-if="applicationId"
            :application-id="applicationId"
            :current-status="currentStatus"
            mode="button"
            button-text="更新状态"
            button-size="small"
            @updated="handleStatusUpdated"
          />
        </div>
      </div>

      <!-- 标签页内容 -->
      <a-tabs v-model:active-key="activeTab" class="detail-tabs">
        <!-- 状态时间轴 -->
        <a-tab-pane key="timeline" tab="状态流转" class="tab-timeline">
          <template #tab>
            <HistoryOutlined />
            状态流转
          </template>
          
          <StatusTimeline
            :application-id="applicationId"
            :current-status="currentStatus"
            max-height="500px"
            @status-updated="handleStatusUpdated"
          />
        </a-tab-pane>

        <!-- 状态统计 -->
        <a-tab-pane key="stats" tab="流程统计" class="tab-stats">
          <template #tab>
            <BarChartOutlined />
            流程统计
          </template>

          <div class="stats-content">
            <a-spin :spinning="statsLoading">
              <!-- 统计卡片 -->
              <a-row :gutter="[16, 16]" class="stats-cards">
              <a-col :span="6">
                <a-card size="small">
                  <a-statistic
                    title="总流程时间"
                    :value="statusHistory?.metadata?.total_duration || 0"
                    :formatter="(value) => formatDuration(value as number)"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card size="small">
                  <a-statistic
                    title="状态变更"
                    :value="statusHistory?.metadata?.status_count || 0"
                    suffix="次"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card size="small">
                  <a-statistic
                    title="当前阶段"
                    :value="currentStage"
                    :value-style="{ fontSize: '16px', color: getCurrentStageColor() }"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card size="small">
                  <a-statistic
                    title="预计完成"
                    :value="estimatedCompletion"
                    :value-style="{ fontSize: '14px' }"
                  />
                </a-card>
              </a-col>
            </a-row>

            <!-- 状态持续时间图表 -->
            <a-card title="各阶段耗时分析" class="duration-chart" size="small">
              <div ref="durationChartRef" style="height: 300px;"></div>
            </a-card>
            </a-spin>
          </div>
        </a-tab-pane>

        <!-- 详细信息 -->
        <a-tab-pane key="details" tab="详细信息" class="tab-details">
          <template #tab>
            <InfoCircleOutlined />
            详细信息
          </template>

          <div class="details-content">
            <a-descriptions :column="2" bordered size="small">
              <a-descriptions-item label="公司名称">
                {{ applicationData?.company_name }}
              </a-descriptions-item>
              <a-descriptions-item label="职位标题">
                {{ applicationData?.position_title }}
              </a-descriptions-item>
              <a-descriptions-item label="工作地点">
                {{ applicationData?.work_location || '未填写' }}
              </a-descriptions-item>
              <a-descriptions-item label="薪资范围">
                {{ applicationData?.salary_range || '面谈' }}
              </a-descriptions-item>
              <a-descriptions-item label="面试时间">
                {{ formatTimestamp(applicationData?.interview_time || '', 'YYYY-MM-DD HH:mm') || '未安排' }}
              </a-descriptions-item>
              <a-descriptions-item label="面试类型">
                {{ applicationData?.interview_type || '未知' }}
              </a-descriptions-item>
              <a-descriptions-item label="HR姓名">
                {{ applicationData?.hr_name || '未填写' }}
              </a-descriptions-item>
              <a-descriptions-item label="HR联系方式">
                <div v-if="applicationData?.hr_phone || applicationData?.hr_email">
                  <div v-if="applicationData?.hr_phone">
                    电话：{{ applicationData.hr_phone }}
                  </div>
                  <div v-if="applicationData?.hr_email">
                    邮箱：{{ applicationData.hr_email }}
                  </div>
                </div>
                <span v-else>未填写</span>
              </a-descriptions-item>
            </a-descriptions>

            <!-- 职位描述 -->
            <a-card title="职位描述" size="small" class="job-description">
              <div v-if="applicationData?.job_description" class="description-content">
                {{ applicationData.job_description }}
              </div>
              <a-empty v-else description="暂无职位描述" :image="'simple'" />
            </a-card>

            <!-- 个人备注 -->
            <a-card title="个人备注" size="small" class="notes">
              <div v-if="applicationData?.notes" class="notes-content">
                {{ applicationData.notes }}
              </div>
              <a-empty v-else description="暂无个人备注" :image="'simple'" />
            </a-card>
          </div>
        </a-tab-pane>

        <!-- 智能建议 -->
        <a-tab-pane key="insights" tab="智能建议" class="tab-insights">
          <template #tab>
            <BulbOutlined />
            智能建议
          </template>

          <div class="insights-content">
            <a-spin :spinning="insightsLoading">
            <!-- 流程洞察 -->
            <a-card title="流程洞察" size="small" class="insights-card">
              <div v-if="insights.length > 0">
                <div
                  v-for="(insight, index) in insights"
                  :key="index"
                  class="insight-item"
                >
                  <a-alert
                    :type="getInsightType(insight.type)"
                    :message="insight.title"
                    :description="insight.description"
                    :show-icon="true"
                    class="insight-alert"
                  />
                  <div v-if="insight.action_suggestion" class="action-suggestion">
                    <strong>建议行动：</strong>{{ insight.action_suggestion }}
                  </div>
                </div>
              </div>
              <a-empty v-else description="暂无流程洞察" :image="'simple'" />
            </a-card>

            <!-- 成功率预测 -->
            <a-card title="成功率分析" size="small" class="success-prediction">
              <div class="prediction-content">
                <div class="success-rate">
                  <a-progress
                    type="circle"
                    :percent="successProbability"
                    :format="(percent) => `${percent}%`"
                    :stroke-color="getSuccessColor(successProbability)"
                  />
                  <div class="rate-text">
                    <h4>预估成功率</h4>
                    <p>基于历史数据和当前进展分析</p>
                  </div>
                </div>
                
                <div class="factors">
                  <h5>影响因素</h5>
                  <ul>
                    <li>当前状态：{{ currentStatus }}</li>
                    <li>流程时长：{{ formatDuration(totalDuration) }}</li>
                    <li>行业平均：相对{{ averageComparison }}</li>
                  </ul>
                </div>
              </div>
            </a-card>
            </a-spin>
          </div>
        </a-tab-pane>
      </a-tabs>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import {
  HistoryOutlined,
  BarChartOutlined,
  InfoCircleOutlined,
  BulbOutlined
} from '@ant-design/icons-vue'
import { useStatusTrackingStore } from '../stores/statusTracking'
import { useJobApplicationStore } from '../stores/jobApplication'
import { StatusHelper, type ApplicationStatus, type StatusHistory, type JobApplication } from '../types'
import StatusTimeline from './StatusTimeline.vue'
import StatusQuickUpdate from './StatusQuickUpdate.vue'
import * as echarts from 'echarts'
import dayjs from 'dayjs'

// Props
interface Props {
  visible: boolean
  applicationId: number
  currentStatus: ApplicationStatus
}

const props = defineProps<Props>()

// Emits
const emit = defineEmits<{
  'update:visible': [value: boolean]
  statusUpdated: [status: ApplicationStatus]
}>()

// Store
const statusTrackingStore = useStatusTrackingStore()
const jobApplicationStore = useJobApplicationStore()

// 响应式数据
const activeTab = ref('timeline')
const statsLoading = ref(false)
const insightsLoading = ref(false)
const statusHistory = ref<StatusHistory | null>(null)
const applicationData = ref<JobApplication | null>(null)
const insights = ref<any[]>([])
const durationChartRef = ref<HTMLElement>()
const durationChart = ref<echarts.ECharts>()

// 创建本地的visible响应式引用
const visible = computed({
  get() {
    return props.visible
  },
  set(value: boolean) {
    emit('update:visible', value)
  }
})

// 计算属性
const modalTitle = computed(() => {
  if (applicationData.value) {
    return `${applicationData.value.company_name} - 状态详情`
  }
  return '状态详情'
})

const currentStage = computed(() => {
  return statusHistory.value?.metadata?.current_stage || '未知阶段'
})

const totalDuration = computed(() => {
  return statusHistory.value?.metadata?.total_duration || 0
})

const successProbability = computed(() => {
  // 基于状态计算成功概率
  const statusScores: Record<string, number> = {
    '已投递': 20,
    '简历筛选中': 30,
    '笔试中': 45,
    '笔试通过': 55,
    '一面中': 60,
    '一面通过': 75,
    '二面中': 80,
    '二面通过': 90,
    '三面中': 92,
    '三面通过': 95,
    'HR面中': 98,
    'HR面通过': 99,
    '待发offer': 95,
    '已收到offer': 100,
    '已接受offer': 100
  }
  return statusScores[props.currentStatus] || 20
})

const estimatedCompletion = computed(() => {
  // 基于当前状态估算完成时间
  const averageDays: Record<string, number> = {
    '已投递': 30,
    '简历筛选中': 25,
    '笔试中': 20,
    '一面中': 15,
    '二面中': 10,
    'HR面中': 5,
    '待发offer': 3
  }
  
  const days = averageDays[props.currentStatus] || 0
  if (days > 0) {
    return dayjs().add(days, 'day').format('MM-DD')
  }
  return '已完成'
})

const averageComparison = computed(() => {
  // 与平均水平比较
  if (totalDuration.value > 0) {
    const avgDuration = 20 * 24 * 60 // 20天的分钟数
    if (totalDuration.value > avgDuration * 1.2) return '偏慢'
    if (totalDuration.value < avgDuration * 0.8) return '较快'
    return '正常'
  }
  return '正常'
})

// 方法
const getCurrentStageColor = (): string => {
  if (currentStage.value.includes('interview') || currentStage.value.includes('面试')) return '#1890ff'
  if (currentStage.value.includes('offer')) return '#52c41a'
  if (currentStage.value.includes('screening') || currentStage.value.includes('筛选')) return '#faad14'
  return '#666'
}

const getInsightType = (type: string): 'success' | 'info' | 'warning' | 'error' => {
  const typeMap: Record<string, 'success' | 'info' | 'warning' | 'error'> = {
    success: 'success',
    info: 'info',
    warning: 'warning',
    error: 'error'
  }
  return typeMap[type] || 'info'
}

const getSuccessColor = (probability: number): string => {
  if (probability >= 80) return '#52c41a'
  if (probability >= 60) return '#faad14'
  return '#ff4d4f'
}

const initDurationChart = () => {
  if (!durationChartRef.value || !statusHistory.value) return

  nextTick(() => {
    if (durationChart.value) {
      durationChart.value.dispose()
    }

    durationChart.value = echarts.init(durationChartRef.value!)

    const history = statusHistory.value!.history
    const chartData = history.map((entry, index) => ({
      name: entry.status,
      value: entry.duration || 0,
      color: StatusHelper.getStatusColor(entry.status as ApplicationStatus)
    }))

    const option = {
      title: {
        text: '各状态停留时长',
        left: 'center',
        textStyle: { fontSize: 14 }
      },
      tooltip: {
        trigger: 'item',
        formatter: (params: any) => {
          return `${params.name}<br/>停留时长: ${statusTrackingStore.formatDuration(params.value)}`
        }
      },
      series: [
        {
          type: 'pie',
          radius: ['40%', '70%'],
          center: ['50%', '60%'],
          data: chartData,
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

    durationChart.value.setOption(option)
  })
}

const fetchData = async () => {
  await Promise.all([
    fetchApplicationData(),
    fetchStatusHistory(),
    fetchInsights()
  ])
}

const fetchApplicationData = async () => {
  try {
    applicationData.value = await jobApplicationStore.fetchApplicationById(props.applicationId)
  } catch (error) {
    console.error('获取申请数据失败:', error)
  }
}

const fetchStatusHistory = async () => {
  statsLoading.value = true
  try {
    statusHistory.value = await statusTrackingStore.fetchStatusHistory(props.applicationId, true)
    if (activeTab.value === 'stats') {
      nextTick(initDurationChart)
    }
  } finally {
    statsLoading.value = false
  }
}

const fetchInsights = async () => {
  insightsLoading.value = true
  try {
    const processInsights = await statusTrackingStore.fetchProcessInsights()
    insights.value = processInsights.insights || []
  } catch (error) {
    console.error('获取智能建议失败:', error)
    insights.value = []
  } finally {
    insightsLoading.value = false
  }
}

const handleCancel = () => {
  emit('update:visible', false)
}

const handleStatusUpdated = (newStatus: ApplicationStatus) => {
  emit('statusUpdated', newStatus)
  // 重新获取数据
  fetchData()
}

// 工具方法
const { formatDuration, formatTimestamp } = statusTrackingStore

// 监听器
watch(() => props.visible, (visible) => {
  if (visible) {
    fetchData()
  }
})

watch(activeTab, (tab) => {
  if (tab === 'stats' && statusHistory.value) {
    nextTick(initDurationChart)
  }
})

// 生命周期
onMounted(() => {
  if (props.visible) {
    fetchData()
  }
})
</script>

<style scoped>
.status-detail-modal :deep(.ant-modal-body) {
  padding: 16px;
}

.modal-content {
  min-height: 600px;
}

.application-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 16px 0;
  border-bottom: 1px solid #f0f0f0;
  margin-bottom: 20px;
}

.header-info h3 {
  margin: 0 0 8px 0;
  color: #262626;
  font-size: 18px;
}

.meta-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.apply-date {
  color: #8c8c8c;
  font-size: 13px;
}

.detail-tabs {
  min-height: 500px;
}

.detail-tabs :deep(.ant-tabs-content-holder) {
  padding-top: 16px;
}

.stats-cards {
  margin-bottom: 24px;
}

.stats-cards .ant-card {
  text-align: center;
}

.duration-chart {
  margin-top: 16px;
}

.details-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.job-description,
.notes {
  margin-top: 16px;
}

.description-content,
.notes-content {
  line-height: 1.6;
  white-space: pre-wrap;
  color: #262626;
}

.insights-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.insight-item {
  margin-bottom: 16px;
}

.insight-alert {
  margin-bottom: 8px;
}

.action-suggestion {
  color: #1890ff;
  font-size: 13px;
  padding: 8px 12px;
  background: #f0f9ff;
  border-radius: 4px;
  border-left: 3px solid #1890ff;
}

.success-prediction {
  margin-top: 16px;
}

.prediction-content {
  display: flex;
  gap: 24px;
  align-items: center;
}

.success-rate {
  display: flex;
  align-items: center;
  gap: 16px;
}

.rate-text h4 {
  margin: 0 0 4px 0;
  color: #262626;
}

.rate-text p {
  margin: 0;
  color: #8c8c8c;
  font-size: 12px;
}

.factors {
  flex: 1;
}

.factors h5 {
  margin: 0 0 12px 0;
  color: #262626;
}

.factors ul {
  margin: 0;
  padding-left: 16px;
}

.factors li {
  color: #595959;
  margin-bottom: 4px;
}

/* 响应式适配 */
@media (max-width: 768px) {
  .status-detail-modal {
    width: 95vw !important;
    margin: 0 auto;
  }

  .application-header {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }

  .prediction-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .success-rate {
    flex-direction: column;
    text-align: center;
  }
}
</style>