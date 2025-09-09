<template>
  <div class="status-quick-update">
    <!-- 卡片模式 -->
    <a-card 
      v-if="mode === 'card'" 
      :title="cardTitle"
      size="small"
      class="update-card"
      :class="{ 'compact': compact }"
    >
      <template #extra>
        <a-button
          type="link"
          size="small"
          @click="showDetailModal = true"
        >
          <HistoryOutlined />
          查看历史
        </a-button>
      </template>
      
      <StatusUpdateContent
        v-model:selectedStatus="selectedStatus"
        v-model:note="note"
        v-model:interviewTime="interviewTime"
        :current-status="currentStatus"
        :available-statuses="availableStatuses"
        :loading="loading"
        :compact="compact"
        @update="handleQuickUpdate"
        @cancel="handleCancel"
      />
    </a-card>

    <!-- 内联模式 -->
    <div v-else-if="mode === 'inline'" class="inline-update">
      <StatusUpdateContent
        v-model:selectedStatus="selectedStatus"
        v-model:note="note"
        v-model:interviewTime="interviewTime"
        :current-status="currentStatus"
        :available-statuses="availableStatuses"
        :loading="loading"
        :compact="compact"
        :show-current="showCurrent"
        @update="handleQuickUpdate"
        @cancel="handleCancel"
      />
    </div>

    <!-- 按钮触发模式 -->
    <div v-else class="button-trigger">
      <a-button
        type="primary"
        :size="buttonSize"
        :loading="loading"
        @click="showUpdateModal = true"
      >
        <template #icon><EditOutlined /></template>
        {{ buttonText }}
      </a-button>
      
      <!-- 弹窗 -->
      <a-modal
        v-model:visible="showUpdateModal"
        title="快速更新状态"
        width="600px"
        :mask-closable="false"
        @cancel="handleModalCancel"
      >
        <StatusUpdateContent
          v-model:selectedStatus="selectedStatus"
          v-model:note="note"
          v-model:interviewTime="interviewTime"
          :current-status="currentStatus"
          :available-statuses="availableStatuses"
          :loading="loading"
          :show-current="true"
          @update="handleQuickUpdate"
          @cancel="handleModalCancel"
        />
        
        <template #footer>
          <a-space>
            <a-button @click="handleModalCancel">取消</a-button>
            <a-button 
              type="primary" 
              :loading="loading"
              :disabled="!selectedStatus || selectedStatus === currentStatus"
              @click="handleQuickUpdate"
            >
              确定更新
            </a-button>
          </a-space>
        </template>
      </a-modal>
    </div>

    <!-- 状态详情弹窗 -->
    <a-modal
      v-model:visible="showDetailModal"
      title="状态流转历史"
      width="800px"
      footer=""
      class="status-detail-modal"
    >
      <StatusTimeline 
        :application-id="applicationId"
        :current-status="currentStatus"
        compact
        max-height="400px"
        @status-updated="handleStatusUpdated"
      />
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { 
  EditOutlined,
  HistoryOutlined
} from '@ant-design/icons-vue'
import { useStatusTrackingStore } from '../stores/statusTracking'
import { StatusHelper, type ApplicationStatus } from '../types'
import StatusTimeline from './StatusTimeline.vue'
import StatusUpdateContent from './StatusUpdateContent.vue'
import { message } from 'ant-design-vue'

// Props
interface Props {
  applicationId: number
  currentStatus: ApplicationStatus
  mode?: 'card' | 'inline' | 'button' // 显示模式
  compact?: boolean // 紧凑模式
  showCurrent?: boolean // 是否显示当前状态
  buttonText?: string // 按钮文字
  buttonSize?: 'large' | 'middle' | 'small'
  autoFetch?: boolean // 是否自动获取可用状态
}

const props = withDefaults(defineProps<Props>(), {
  mode: 'button',
  compact: false,
  showCurrent: true,
  buttonText: '更新状态',
  buttonSize: 'middle',
  autoFetch: true
})

// Emits
const emit = defineEmits<{
  updated: [status: ApplicationStatus, note?: string]
  cancelled: []
}>()

// Store
const statusTrackingStore = useStatusTrackingStore()

// 响应式数据
const loading = ref(false)
const showUpdateModal = ref(false)
const showDetailModal = ref(false)
const selectedStatus = ref<ApplicationStatus | ''>('')
const note = ref('')
const interviewTime = ref('')
const availableStatuses = ref<ApplicationStatus[]>([])

// 计算属性
const cardTitle = computed(() => {
  const statusTag = StatusHelper.getStatusCategory(props.currentStatus)
  return `当前状态：${props.currentStatus} (${statusTag})`
})

// 方法
const fetchAvailableStatuses = async () => {
  if (!props.autoFetch) return
  
  try {
    const transitions = await statusTrackingStore.getAvailableTransitions(props.currentStatus)
    availableStatuses.value = transitions.reduce((acc: ApplicationStatus[], rule) => {
      acc.push(...rule.to)
      return acc
    }, [])
  } catch (error) {
    console.error('获取可用状态失败:', error)
    // 降级到默认状态列表
    availableStatuses.value = getDefaultNextStatuses(props.currentStatus)
  }
}

const getDefaultNextStatuses = (currentStatus: ApplicationStatus): ApplicationStatus[] => {
  // 基于业务逻辑的默认状态转换
  const statusFlow: Record<ApplicationStatus, ApplicationStatus[]> = {
    '已投递': ['简历筛选中', '简历筛选未通过'],
    '简历筛选中': ['笔试中', '一面中', '简历筛选未通过'],
    '笔试中': ['笔试通过', '笔试未通过'],
    '笔试通过': ['一面中'],
    '一面中': ['一面通过', '一面未通过'],
    '一面通过': ['二面中', 'HR面中', '待发offer'],
    '二面中': ['二面通过', '二面未通过'],
    '二面通过': ['三面中', 'HR面中', '待发offer'],
    '三面中': ['三面通过', '三面未通过'],
    '三面通过': ['HR面中', '待发offer'],
    'HR面中': ['HR面通过', 'HR面未通过'],
    'HR面通过': ['待发offer'],
    '待发offer': ['已收到offer', '已拒绝'],
    '已收到offer': ['已接受offer', '已拒绝'],
    '已接受offer': ['流程结束'],
    '已拒绝': ['流程结束'],
    '简历筛选未通过': ['流程结束'],
    '笔试未通过': ['流程结束'],
    '一面未通过': ['流程结束'],
    '二面未通过': ['流程结束'],
    '三面未通过': ['流程结束'],
    'HR面未通过': ['流程结束'],
    '流程结束': []
  }
  
  return statusFlow[currentStatus] || []
}

const handleQuickUpdate = async () => {
  if (!selectedStatus.value || selectedStatus.value === props.currentStatus) {
    message.warning('请选择不同的状态')
    return
  }

  loading.value = true
  try {
    const updateData: any = {
      status: selectedStatus.value,
      note: note.value || undefined
    }
    
    // 如果是面试相关状态且设置了面试时间
    if (interviewTime.value && isInterviewStatus(selectedStatus.value)) {
      updateData.interview_scheduled = interviewTime.value
    }

    await statusTrackingStore.updateApplicationStatus(props.applicationId, updateData)
    
    emit('updated', selectedStatus.value, note.value || undefined)
    resetForm()
    
    if (props.mode === 'button') {
      showUpdateModal.value = false
    }
    
  } finally {
    loading.value = false
  }
}

const handleCancel = () => {
  resetForm()
  emit('cancelled')
}

const handleModalCancel = () => {
  handleCancel()
  showUpdateModal.value = false
}

const handleStatusUpdated = (newStatus: ApplicationStatus) => {
  emit('updated', newStatus)
  showDetailModal.value = false
}

const resetForm = () => {
  selectedStatus.value = ''
  note.value = ''
  interviewTime.value = ''
}

const isInterviewStatus = (status: ApplicationStatus): boolean => {
  return ['一面中', '二面中', '三面中', 'HR面中', '笔试中'].includes(status)
}

// 生命周期
onMounted(() => {
  fetchAvailableStatuses()
})

// 监听器
watch(() => props.currentStatus, () => {
  resetForm()
  fetchAvailableStatuses()
}, { immediate: false })
</script>

<!-- StatusUpdateContent 子组件 -->

<style scoped>
.status-quick-update {
  width: 100%;
}

.update-card {
  border-radius: 8px;
}

.update-card.compact {
  margin-bottom: 16px;
}

.inline-update {
  padding: 16px;
  background: #fafafa;
  border-radius: 8px;
  border: 1px solid #f0f0f0;
}

/* 状态详情弹窗样式 */
:deep(.status-detail-modal .ant-modal-body) {
  padding: 16px;
}

/* 响应式适配 */
@media (max-width: 768px) {
  .button-trigger .ant-modal {
    width: 90vw !important;
    margin: 0 auto;
  }
}
</style>