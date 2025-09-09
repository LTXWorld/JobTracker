<template>
  <div class="status-timeline">
    <!-- 标题和操作栏 -->
    <div class="timeline-header">
      <div class="header-title">
        <h3>
          <HistoryOutlined />
          状态流转历史
        </h3>
        <span class="timeline-meta" v-if="timelineData.length > 0">
          共 {{ timelineData.length }} 次变更，
          总用时 {{ formatDuration(totalDuration) }}
        </span>
        <!-- 流转链路概览 -->
        <div class="flow-chain" v-if="timelineData.length > 0">
          <span v-if="statusHistory?.metadata.initial_status" class="flow-item">
            <a-tag :color="'#1890ff'" class="flow-tag">{{ statusHistory!.metadata.initial_status }}</a-tag>
            <span class="flow-arrow">→</span>
          </span>
          <span v-for="(item, idx) in timelineData" :key="item.id + '_chain'" class="flow-item">
            <a-tag :color="item.color" class="flow-tag">{{ item.status }}</a-tag>
            <span v-if="idx < timelineData.length - 1" class="flow-arrow">→</span>
          </span>
        </div>
      </div>
      <div class="header-actions">
        <a-button
          :loading="loading"
          size="small"
          @click="refreshTimeline"
        >
          <template #icon><ReloadOutlined /></template>
          刷新
        </a-button>
        <a-button
          type="primary"
          size="small"
          @click="showUpdateModal = true"
        >
          <template #icon><EditOutlined /></template>
          更新状态
        </a-button>
      </div>
    </div>

    <!-- 时间轴内容 -->
    <div class="timeline-content" v-loading="loading">
      <a-empty
        v-if="!loading && timelineData.length === 0"
        description="暂无状态历史记录"
      />
      
      <a-timeline v-else class="status-timeline-wrapper">
        <a-timeline-item
          v-for="(item, index) in timelineData"
          :key="item.id"
          :color="getTimelineColor(item)"
        >
          <template #dot>
            <div 
              class="timeline-dot" 
              :class="{
                'current-dot': item.is_current,
                'failed-dot': item.is_failed,
                'passed-dot': item.is_passed
              }"
            >
              <component 
                :is="getIconComponent(item.icon)" 
                :style="{ fontSize: '14px', color: '#fff' }"
              />
            </div>
          </template>
          
          <div class="timeline-item-content">
            <!-- 状态标题 -->
            <div class="status-header">
              <h4 class="status-title">
                <a-tag 
                  :color="item.color"
                  class="status-tag"
                >
                  {{ item.status }}
                </a-tag>
                <span 
                  v-if="item.is_current" 
                  class="current-badge"
                >
                  当前状态
                </span>
              </h4>
              <span class="status-time">
                {{ formatTimestamp(item.timestamp, 'MM-DD HH:mm') }}
              </span>
            </div>

            <!-- 状态详情 -->
            <div class="status-details">
              <div class="detail-row" v-if="item.duration && index > 0">
                <ClockCircleOutlined />
                <span>停留时长：{{ formatDuration(item.duration) }}</span>
              </div>
              
              <div class="detail-row" v-if="item.interview_scheduled">
                <CalendarOutlined />
                <span>
                  面试安排：{{ formatTimestamp(item.interview_scheduled, 'YYYY-MM-DD HH:mm') }}
                </span>
              </div>
              
              <div class="detail-row" v-if="item.note">
                <FileTextOutlined />
                <span class="status-note">{{ item.note }}</span>
              </div>
            </div>

            <!-- 时间进度条（对于当前状态显示） -->
            <div v-if="item.is_current && !item.is_failed" class="time-progress">
              <div class="progress-info">
                <span>当前阶段进度</span>
                <span class="progress-time">
                  {{ formatDuration(getCurrentStatusDuration(item.timestamp)) }}
                </span>
              </div>
              <a-progress
                :percent="getProgressPercent(item.status)"
                :status="item.is_failed ? 'exception' : 'active'"
                :stroke-color="item.color"
                size="small"
              />
            </div>
          </div>
        </a-timeline-item>
      </a-timeline>
    </div>

    <!-- 时间统计摘要 -->
    <div class="timeline-summary" v-if="statusHistory?.metadata">
      <a-row :gutter="16">
        <a-col :span="6">
          <a-statistic
            title="总流程时间"
            :value="totalDuration"
            :formatter="(value) => formatDuration(Number(value))"
          />
        </a-col>
        <a-col :span="6">
          <a-statistic
            title="状态变更次数"
            :value="statusHistory.metadata.status_count"
            suffix="次"
          />
        </a-col>
        <a-col :span="6">
          <a-statistic
            title="当前阶段"
            :value="statusHistory.metadata.current_stage"
            :value-style="{ color: getCurrentStageColor() }"
          />
        </a-col>
        <a-col :span="6">
          <a-statistic
            title="最后更新"
            :value="statusHistory.metadata.last_updated"
            :formatter="(value) => (value ? formatTimestamp(String(value), 'MM-DD HH:mm') : '-')"
          />
        </a-col>
      </a-row>
    </div>

    <!-- 状态更新弹窗 -->
    <a-modal
      v-model:visible="showUpdateModal"
      title="更新状态"
      width="500px"
      :footer="null"
    >
      <StatusQuickUpdate
        :application-id="applicationId"
        :current-status="currentStatus || '已投递'"
        mode="inline"
        @updated="handleStatusUpdated"
        @cancelled="showUpdateModal = false"
      />
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { 
  HistoryOutlined, 
  ReloadOutlined, 
  EditOutlined,
  ClockCircleOutlined,
  CalendarOutlined,
  FileTextOutlined,
  SendOutlined,
  EyeOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  UserOutlined,
  TeamOutlined,
  CrownOutlined,
  ContactsOutlined,
  GiftOutlined,
  StopOutlined,
  TrophyOutlined,
  FlagOutlined,
  QuestionCircleOutlined
} from '@ant-design/icons-vue'
import { useStatusTrackingStore } from '../stores/statusTracking'
import { type ApplicationStatus, type StatusTimelineItem, type StatusHistory } from '../types'
import StatusQuickUpdate from './StatusQuickUpdate.vue'
import dayjs from 'dayjs'

// Props
interface Props {
  applicationId: number
  currentStatus?: ApplicationStatus
  compact?: boolean // 紧凑模式
  maxHeight?: string // 最大高度
}

const props = withDefaults(defineProps<Props>(), {
  compact: false,
  maxHeight: '600px'
})

// Emits
const emit = defineEmits<{
  statusUpdated: [status: ApplicationStatus]
}>()

// Store
const statusTrackingStore = useStatusTrackingStore()

// 响应式数据
const loading = ref(false)
const showUpdateModal = ref(false)
const statusHistory = ref<StatusHistory | null>(null)

// 计算属性
const timelineData = computed((): StatusTimelineItem[] => {
  if (!statusHistory.value) return []
  return statusTrackingStore.convertToTimelineData(statusHistory.value)
})

const totalDuration = computed((): number => {
  return statusHistory.value?.metadata.total_duration || 0
})

// 图标组件映射
const iconComponents = {
  SendOutlined,
  EyeOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  UserOutlined,
  TeamOutlined,
  CrownOutlined,
  ContactsOutlined,
  GiftOutlined,
  StopOutlined,
  TrophyOutlined,
  FlagOutlined,
  QuestionCircleOutlined
}

// 方法
const getIconComponent = (iconName: string) => {
  return (iconComponents as any)[iconName] || QuestionCircleOutlined
}

const getTimelineColor = (item: StatusTimelineItem): string => {
  if (item.is_failed) return '#ff4d4f'
  if (item.is_passed) return '#52c41a'
  if (item.is_current) return '#1890ff'
  return '#d9d9d9'
}

const getCurrentStatusDuration = (timestamp: string): number => {
  return statusTrackingStore.calculateDuration(timestamp)
}

const getProgressPercent = (status: ApplicationStatus): number => {
  // 根据状态估算进度百分比
  const progressMap: Record<string, number> = {
    '已投递': 10,
    '简历筛选中': 20,
    '笔试中': 30,
    '笔试通过': 40,
    '一面中': 50,
    '一面通过': 60,
    '二面中': 70,
    '二面通过': 80,
    '三面中': 85,
    '三面通过': 90,
    'HR面中': 95,
    'HR面通过': 98,
    '待发offer': 99,
    '已收到offer': 100,
    '已接受offer': 100
  }
  return progressMap[status] || 0
}

const getCurrentStageColor = (): string => {
  if (!statusHistory.value) return '#666'
  const stage = statusHistory.value.metadata.current_stage
  
  if (stage.includes('interview') || stage.includes('面试')) return '#1890ff'
  if (stage.includes('offer')) return '#52c41a'
  if (stage.includes('screening') || stage.includes('筛选')) return '#faad14'
  return '#666'
}

const refreshTimeline = async () => {
  await fetchStatusHistory(true)
}

const fetchStatusHistory = async (forceRefresh = false) => {
  loading.value = true
  try {
    const history = await statusTrackingStore.fetchStatusHistory(props.applicationId, forceRefresh)
    statusHistory.value = history
  } finally {
    loading.value = false
  }
}

const handleStatusUpdated = (newStatus: ApplicationStatus) => {
  emit('statusUpdated', newStatus)
  refreshTimeline()
}

// 工具方法
const { formatDuration, formatTimestamp } = statusTrackingStore

// 生命周期
onMounted(() => {
  fetchStatusHistory()
})

// 监听器
watch(() => props.applicationId, (newId) => {
  if (newId) {
    fetchStatusHistory(true)
  }
}, { immediate: false })
</script>

<style scoped>
.status-timeline {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.timeline-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.header-title h3 {
  margin: 0 0 4px 0;
  color: #262626;
  display: flex;
  align-items: center;
  gap: 8px;
}

.timeline-meta {
  color: #8c8c8c;
  font-size: 12px;
}

.flow-chain {
  margin-top: 6px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.flow-item {
  display: flex;
  align-items: center;
}

.flow-tag {
  border-radius: 10px;
}

.flow-arrow {
  margin: 0 4px;
  color: #bfbfbf;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.timeline-content {
  max-height: v-bind(maxHeight);
  overflow-y: auto;
}

.status-timeline-wrapper :deep(.ant-timeline-item-tail) {
  border-left: 2px solid #f0f0f0;
}

.timeline-dot {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #d9d9d9;
  border: 2px solid #fff;
  box-shadow: 0 0 0 2px #f0f0f0;
}

.timeline-dot.current-dot {
  background: #1890ff;
  box-shadow: 0 0 0 2px #e6f7ff;
  animation: pulse 2s infinite;
}

.timeline-dot.failed-dot {
  background: #ff4d4f;
  box-shadow: 0 0 0 2px #fff2f0;
}

.timeline-dot.passed-dot {
  background: #52c41a;
  box-shadow: 0 0 0 2px #f6ffed;
}

@keyframes pulse {
  0% {
    box-shadow: 0 0 0 2px #e6f7ff, 0 0 0 4px rgba(24, 144, 255, 0.2);
  }
  50% {
    box-shadow: 0 0 0 2px #e6f7ff, 0 0 0 8px rgba(24, 144, 255, 0.1);
  }
  100% {
    box-shadow: 0 0 0 2px #e6f7ff, 0 0 0 4px rgba(24, 144, 255, 0.2);
  }
}

.timeline-item-content {
  padding-left: 16px;
}

.status-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.status-title {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-tag {
  margin: 0;
  font-weight: 500;
}

.current-badge {
  background: linear-gradient(135deg, #1890ff, #40a9ff);
  color: white;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 500;
}

.status-time {
  color: #8c8c8c;
  font-size: 12px;
}

.status-details {
  margin: 12px 0;
}

.detail-row {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #666;
  font-size: 13px;
  margin: 4px 0;
}

.detail-row .anticon {
  color: #8c8c8c;
}

.status-note {
  color: #262626;
  font-style: italic;
  background: #f8f8f8;
  padding: 4px 8px;
  border-radius: 4px;
  border-left: 3px solid #d9d9d9;
}

.time-progress {
  margin-top: 12px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 6px;
  border: 1px solid #e9ecef;
}

.progress-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.progress-info span:first-child {
  font-size: 12px;
  color: #666;
}

.progress-time {
  font-size: 12px;
  font-weight: 500;
  color: #1890ff;
}

.timeline-summary {
  margin-top: 24px;
  padding: 20px;
  background: #fafafa;
  border-radius: 8px;
}

.timeline-summary :deep(.ant-statistic-title) {
  color: #8c8c8c;
  font-size: 12px;
}

.timeline-summary :deep(.ant-statistic-content) {
  color: #262626;
  font-size: 16px;
  font-weight: 500;
}

/* 紧凑模式样式 */
.status-timeline.compact {
  padding: 12px;
}

.status-timeline.compact .timeline-header {
  margin-bottom: 12px;
  padding-bottom: 8px;
}

.status-timeline.compact .header-title h3 {
  font-size: 16px;
}

.status-timeline.compact .timeline-dot {
  width: 20px;
  height: 20px;
}

.status-timeline.compact .timeline-item-content {
  padding-left: 12px;
}

/* 响应式适配 */
@media (max-width: 768px) {
  .timeline-header {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }

  .header-actions {
    align-self: stretch;
    justify-content: flex-end;
  }

  .timeline-summary .ant-row {
    flex-direction: column;
    gap: 16px;
  }

  .timeline-summary .ant-col {
    width: 100% !important;
    text-align: center;
  }
}
</style>
