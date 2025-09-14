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
        <div class="flow-chain" v-if="timelineData.length > 0" ref="flowChainRef" :class="{ 'has-oc': showOcIndicator }">
          <span v-if="statusHistory?.metadata.initial_status" class="flow-item">
            <a-tag 
              :color="'#1890ff'" 
              class="flow-tag"
              :ref="(el:any) => assignStatusRef(String(statusHistory!.metadata.initial_status), el)"
            >
              {{ statusHistory!.metadata.initial_status }}
            </a-tag>
            <span class="flow-arrow">→</span>
          </span>
          <span v-for="(item, idx) in timelineData" :key="item.id + '_chain'" class="flow-item">
            <a-tag 
              :color="item.color" 
              class="flow-tag"
              :ref="(el:any) => assignStatusRef(item.status, el)"
            >
              {{ item.status }}
            </a-tag>
            <span v-if="idx < timelineData.length - 1" class="flow-arrow">→</span>
          </span>
          <div v-if="showOcIndicator" class="oc-indicator" :style="ocStyle">
            <div class="oc-track"></div>
            <div 
              class="oc-carrier" 
              :key="ocAnimKey"
              :class="{ traveling: !ocSettled && !ocPaused, paused: ocPaused }"
              :style="carrierStyle"
              @animationend="handleTravelEnd"
            >
              <div class="oc-ball" :class="{ paused: ocPaused }">OC</div>
            </div>
          </div>
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

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { 
  HistoryOutlined, 
  ReloadOutlined, 
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
import { StatusHelper, type ApplicationStatus, type StatusTimelineItem, type StatusHistory } from '../types'
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
const statusHistory = ref<StatusHistory | null>(null)
const flowChainRef = ref<HTMLElement | null>(null)
// 关键锚点与通用映射
const appliedTagEl = ref<HTMLElement | null>(null)
const screeningTagEl = ref<HTMLElement | null>(null)
const tagElMap = new Map<string, HTMLElement>()
const firstChainTagEl = ref<HTMLElement | null>(null) // 兜底：链路首个标签
const secondChainTagEl = ref<HTMLElement | null>(null) // 兜底：链路第二个标签
const trackLeft = ref(0)
const trackWidth = ref(0)

// 兼容 antd-vue 组件实例与真实 DOM
const unwrapEl = (node: any): HTMLElement | null => {
  if (!node) return null
  if (node instanceof HTMLElement) return node
  if (node.$el && node.$el instanceof HTMLElement) return node.$el
  if (node.$?.vnode?.el && node.$?.vnode?.el instanceof HTMLElement) return node.$?.vnode?.el as HTMLElement
  return null
}

const assignStatusRef = (status: string, el: any) => {
  const dom = unwrapEl(el)
  if (!dom) return
  if (status === '已投递') appliedTagEl.value = dom
  if (status === '简历筛选中') screeningTagEl.value = dom
  // 兜底：记录链路中前两个标签
  if (!firstChainTagEl.value) {
    firstChainTagEl.value = dom
  } else if (!secondChainTagEl.value && dom !== firstChainTagEl.value) {
    secondChainTagEl.value = dom
  }
  // 通用映射（覆盖为最新一次渲染）
  tagElMap.set(status, dom)
}

// 计算属性
const timelineData = computed((): StatusTimelineItem[] => {
  if (!statusHistory.value) return []
  return statusTrackingStore.convertToTimelineData(statusHistory.value)
})

const totalDuration = computed((): number => {
  return statusHistory.value?.metadata.total_duration || 0
})

const offerReached = computed(() => {
  const terminalNames = new Set(['已收到offer', '已接受offer', '待发offer', '流程结束', '已拒绝'])
  return timelineData.value.some(i => terminalNames.has(i.status) || StatusHelper.isFailedStatus(i.status))
})

const showOcIndicator = computed(() => {
  // 只要能定位到起点（已投递或链路第一个标签），就显示
  return !!flowChainRef.value && (!!appliedTagEl.value || !!firstChainTagEl.value)
})

const ocPaused = computed(() => offerReached.value)
const ocSettled = ref(false)
const ocHasAnimated = ref(false)
const ocAnimKey = ref(0)

const ocStyle = computed(() => {
  const w = Math.max(trackWidth.value || 0, 140)
  return {
    left: trackLeft.value + 'px',
    width: w + 'px',
    '--oc-size': '18px'
  } as any
})

const carrierStyle = computed(() => {
  const size = 18
  const distance = Math.max((trackWidth.value || 0) - size, 0)
  return ocSettled.value
    ? { transform: `translateX(${distance}px)` }
    : { ['--oc-distance' as any]: `${distance}px` }
})

const handleTravelEnd = () => {
  ocSettled.value = true
  ocHasAnimated.value = true
}

let resizeObs: ResizeObserver | null = null
const updateOcTrack = () => {
  const container = flowChainRef.value
  // 起点：优先“已投递”，否则元数据 initial_status，再否则链路第一项
  let startEl = appliedTagEl.value 
    || (statusHistory.value?.metadata?.initial_status ? tagElMap.get(String(statusHistory.value?.metadata?.initial_status)) || null : null)
    || firstChainTagEl.value
  if (!container) return
  // 兜底：如果没采集到 ref，尝试从 DOM 查询 ant-tag
  if (!startEl) {
    const tags = container.querySelectorAll('.ant-tag')
    if (tags && tags.length > 0) startEl = tags[0] as HTMLElement
    if (!startEl) return
  }
  const cRect = container.getBoundingClientRect()
  const aRect = startEl.getBoundingClientRect()
  let left = aRect.left - cRect.left
  let width = Math.max(container.clientWidth * 0.3, 140)
  // 终点：优先当前状态，其次“简历筛选中”，最后链路第二项
  let endEl: HTMLElement | null | undefined = undefined
  if (props.currentStatus) {
    endEl = tagElMap.get(props.currentStatus) || undefined
  }
  if (!endEl) endEl = screeningTagEl.value
  if (!endEl) endEl = secondChainTagEl.value
  if (!endEl) {
    const tags = container.querySelectorAll('.ant-tag')
    if (tags && tags.length > 1) endEl = tags[1] as HTMLElement
  }
  if (endEl) {
    const sRect = endEl.getBoundingClientRect()
    width = Math.max(60, sRect.right - aRect.left)
  }
  trackLeft.value = Math.max(0, Math.round(left))
  trackWidth.value = Math.round(width)
}

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
  nextTick(() => {
    updateOcTrack()
    if (flowChainRef.value && 'ResizeObserver' in window) {
      resizeObs = new ResizeObserver(() => updateOcTrack())
      resizeObs.observe(flowChainRef.value)
    }
    window.addEventListener('resize', updateOcTrack)
  })
})

// 监听器
watch(() => props.applicationId, (newId) => {
  if (newId) {
    ocHasAnimated.value = false
    ocSettled.value = false
    ocAnimKey.value++
    fetchStatusHistory(true)
  }
}, { immediate: false })

watch([timelineData, () => statusHistory.value?.metadata?.initial_status], () => {
  // 重置锚点，等待新渲染后重新采集
  appliedTagEl.value = null
  screeningTagEl.value = null
  firstChainTagEl.value = null
  secondChainTagEl.value = null
  tagElMap.clear()
  ocSettled.value = false
  // 仅在本次会话首次动画时重启动画
  if (!ocHasAnimated.value) {
    ocAnimKey.value++
  }
  nextTick(() => updateOcTrack())
})

watch([appliedTagEl, screeningTagEl, firstChainTagEl, secondChainTagEl, () => props.currentStatus], () => {
  if (!ocHasAnimated.value) ocAnimKey.value++
  nextTick(() => updateOcTrack())
})

onBeforeUnmount(() => {
  if (resizeObs) resizeObs.disconnect()
  window.removeEventListener('resize', updateOcTrack)
})

// 当当前状态变化时，从起点再次滚到新的当前位置
watch(() => props.currentStatus, (cur, prev) => {
  if (cur !== prev) {
    ocSettled.value = false
    ocAnimKey.value++
    nextTick(() => updateOcTrack())
  }
})
</script>

<style scoped>
.status-timeline {
  background: var(--bg-card);
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
  position: relative;
}

.flow-chain.has-oc {
  padding-bottom: 24px; /* 为OC指示器预留空间，仅在需要时添加 */
}

.flow-item {
  display: flex;
  align-items: center;
}

.flow-tag {
  border-radius: 10px;
}

/* OC 动态指示器 */
.oc-indicator {
  position: absolute;
  top: calc(100% + 8px);
  height: 20px;
  pointer-events: none;
  z-index: 1;
}

.oc-track {
  position: absolute;
  left: 0; right: 0;
  top: 50%;
  height: 2px;
  background: var(--border-color);
  transform: translateY(-50%);
  opacity: 0.6;
  border-radius: 2px;
}

.oc-carrier {
  position: absolute;
  top: -7px;
  left: 0;
}

.oc-carrier.traveling {
  animation: oc-travel var(--oc-speed, 2.4s) linear forwards;
}

.oc-carrier.paused { animation-play-state: paused; }

.oc-ball {
  position: absolute;
  top: 0;
  left: 0;
  width: var(--oc-size, 18px);
  height: var(--oc-size, 18px);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: 700;
  color: #fff;
  background: linear-gradient(135deg, #69c0ff, #1890ff);
  border: 1px solid rgba(255,255,255,0.25);
  box-shadow: 0 2px 6px rgba(24, 144, 255, 0.4);
  animation: oc-spin 0.9s linear infinite;
  will-change: transform;
  z-index: 1;
}

.oc-ball.paused { animation-play-state: paused; }

@keyframes oc-travel {
  from { transform: translateX(0); }
  to { transform: translateX(var(--oc-distance)); }
}

@keyframes oc-spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@media (prefers-reduced-motion: reduce) {
  .oc-ball { animation: none; }
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

/* 放大自定义圆点并为内容留出空间，避免遮挡“当前状态/状态标签” */
.status-timeline-wrapper :deep(.ant-timeline-item-head),
.status-timeline-wrapper :deep(.ant-timeline-item-head-custom) {
  width: 32px;
  height: 32px;
  left: 8px;
  inset-inline-start: 8px;
}

.status-timeline-wrapper :deep(.ant-timeline-item-tail) {
  left: 23px; /* 32/2 + 8 - 1 对齐圆点中心 */
  inset-inline-start: 23px;
}

.status-timeline-wrapper :deep(.ant-timeline-item-content) {
  margin-left: 56px; /* 为内容增加左侧余量，避免覆盖圆点 */
  margin-inline-start: 56px;
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
  z-index: 2;
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
  flex-wrap: wrap; /* 文本过长时换行，避免被遮挡 */
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
  background: var(--bg-muted);
  padding: 4px 8px;
  border-radius: 4px;
  border-left: 3px solid var(--border-color);
}

.time-progress {
  margin-top: 12px;
  padding: 12px;
  background: var(--bg-muted);
  border-radius: 6px;
  border: 1px solid var(--border-color);
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
  background: var(--bg-muted);
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
