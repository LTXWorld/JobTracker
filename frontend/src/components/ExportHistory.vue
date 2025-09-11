<template>
  <a-modal
    :open="visible"
    @update:open="(val) => emit('update:visible', val)"
    title="导出历史"
    width="900px"
    :footer="null"
    @cancel="handleCancel"
  >
    <div class="export-history">
      <!-- 工具栏 -->
      <div class="history-toolbar">
        <a-row :gutter="16">
          <a-col :flex="1">
            <a-space>
              <a-button @click="refreshHistory" :loading="loading">
                <template #icon><ReloadOutlined /></template>
                刷新
              </a-button>
              <a-button @click="cleanupExpired" :loading="cleanupLoading">
                <template #icon><DeleteOutlined /></template>
                清理过期文件
              </a-button>
            </a-space>
          </a-col>
          <a-col>
            <a-input-search
              v-model:value="searchKeyword"
              placeholder="搜索文件名或任务ID"
              @search="handleSearch"
              style="width: 250px;"
            />
          </a-col>
        </a-row>
      </div>

      <!-- 统计信息 -->
      <div class="history-stats" v-if="historyData">
        <a-row :gutter="16">
          <a-col :span="6">
            <a-statistic title="总导出次数" :value="historyData.pagination.total_count" />
          </a-col>
          <a-col :span="6">
            <a-statistic 
              title="成功导出" 
              :value="getStatusCount('completed')" 
              :value-style="{ color: '#52c41a' }"
            />
          </a-col>
          <a-col :span="6">
            <a-statistic 
              title="失败导出" 
              :value="getStatusCount('failed')" 
              :value-style="{ color: '#ff4d4f' }"
            />
          </a-col>
          <a-col :span="6">
            <a-statistic 
              title="处理中" 
              :value="getStatusCount('processing')" 
              :value-style="{ color: '#1890ff' }"
            />
          </a-col>
        </a-row>
      </div>

      <a-divider />

      <!-- 导出历史列表 -->
      <div class="history-list">
        <a-table
          :dataSource="filteredHistoryList"
          :columns="columns"
          :loading="loading"
          :pagination="{
            current: currentPage,
            pageSize: pageSize,
            total: historyData?.pagination.total_count || 0,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
            onChange: handlePageChange,
            onShowSizeChange: handlePageSizeChange
          }"
          :scroll="{ x: 800 }"
          row-key="task_id"
        >
          <!-- 文件名列 -->
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'filename'">
              <div class="filename-cell">
                <FileExcelOutlined style="color: #52c41a; margin-right: 8px;" />
                <a-tooltip :title="record.filename">
                  <span class="filename-text">{{ record.filename }}</span>
                </a-tooltip>
              </div>
            </template>

            <!-- 状态列 -->
            <template v-else-if="column.key === 'status'">
              <a-tag :color="getTaskStatusColor(record.status)">
                <component :is="getTaskStatusIcon(record.status)" style="margin-right: 4px;" />
                {{ getTaskStatusLabel(record.status) }}
              </a-tag>
            </template>

            <!-- 文件大小列 -->
            <template v-else-if="column.key === 'file_size'">
              {{ formatFileSize(record.file_size) }}
            </template>

            <!-- 创建时间列 -->
            <template v-else-if="column.key === 'created_at'">
              <div class="time-cell">
                <div>{{ formatTime(record.created_at) }}</div>
                <small style="color: #999;">{{ formatRelativeTime(record.created_at) }}</small>
              </div>
            </template>

            <!-- 过期时间列 -->
            <template v-else-if="column.key === 'expires_at'">
              <div v-if="record.expires_at" class="expire-cell">
                <div :class="{ 'expire-warning': isExpiringSoon(record.expires_at) }">
                  {{ formatRemainingTime(record.expires_at) }}
                </div>
              </div>
              <span v-else>--</span>
            </template>

            <!-- 操作列 -->
            <template v-else-if="column.key === 'actions'">
              <a-space>
                <a-button 
                  v-if="record.status === 'completed' && record.download_url" 
                  type="primary" 
                  size="small"
                  @click="downloadFile(record)"
                  :loading="downloadingTasks.has(record.task_id)"
                >
                  <template #icon><DownloadOutlined /></template>
                  下载
                </a-button>

                <a-button 
                  v-if="record.status === 'processing'" 
                  type="default" 
                  size="small"
                  @click="viewTaskProgress(record)"
                >
                  <template #icon><EyeOutlined /></template>
                  查看进度
                </a-button>

                <a-popconfirm
                  title="确定要删除这个导出记录吗？"
                  @confirm="deleteHistoryItem(record.task_id)"
                >
                  <a-button 
                    type="text" 
                    size="small" 
                    danger
                    :loading="deletingTasks.has(record.task_id)"
                  >
                    <template #icon><DeleteOutlined /></template>
                    删除
                  </a-button>
                </a-popconfirm>
              </a-space>
            </template>
          </template>
        </a-table>
      </div>

      <!-- 任务进度查看弹窗 -->
      <a-modal
        v-model:visible="showProgressModal"
        title="导出进度详情"
        :footer="null"
        width="600px"
      >
        <div v-if="selectedTask" class="progress-detail">
          <a-descriptions :column="2" bordered>
            <a-descriptions-item label="任务ID">
              {{ selectedTask.task_id }}
            </a-descriptions-item>
            <a-descriptions-item label="状态">
              <a-tag :color="getTaskStatusColor(selectedTask.status)">
                {{ getTaskStatusLabel(selectedTask.status) }}
              </a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="文件名">
              {{ selectedTask.filename }}
            </a-descriptions-item>
            <a-descriptions-item label="记录数">
              {{ selectedTask.record_count }}
            </a-descriptions-item>
            <a-descriptions-item label="创建时间" :span="2">
              {{ formatTime(selectedTask.created_at) }}
            </a-descriptions-item>
          </a-descriptions>

          <div v-if="taskProgress" class="progress-info">
            <a-divider>实时进度</a-divider>
            <a-progress 
              :percent="taskProgress.progress" 
              :status="getProgressStatus(taskProgress.status)"
            >
              <template #format="{ percent }">
                {{ percent }}% ({{ taskProgress.processed_records }}/{{ taskProgress.total_records }})
              </template>
            </a-progress>

            <div class="progress-meta">
              <a-descriptions size="small" :column="2">
                <a-descriptions-item label="预计剩余时间">
                  {{ formatEstimatedTime(taskProgress.estimated_time) }}
                </a-descriptions-item>
                <a-descriptions-item label="处理速度">
                  {{ calculateProcessingSpeed() }} 条/分钟
                </a-descriptions-item>
              </a-descriptions>
            </div>

            <div v-if="taskProgress.status === 'failed'" class="error-info">
              <a-alert
                type="error"
                :message="taskProgress.error_message"
                show-icon
              />
            </div>
          </div>
        </div>
      </a-modal>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { message } from 'ant-design-vue'
import {
  ReloadOutlined,
  DeleteOutlined,
  FileExcelOutlined,
  DownloadOutlined,
  EyeOutlined,
  CheckCircleOutlined,
  LoadingOutlined,
  ExclamationCircleOutlined,
  ClockCircleOutlined,
  StopOutlined
} from '@ant-design/icons-vue'
import type { 
  ExportHistoryItem, 
  ExportHistoryResponse,
  ExportTask,
  TaskStatus
} from '../types'
import { 
  ExportAPI, 
  TaskStatusConfig,
  formatFileSize,
  formatRemainingTime
} from '../api/export'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'

// dayjs插件
dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

interface Props {
  visible: boolean
}

interface Emits {
  (e: 'update:visible', value: boolean): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 响应式数据
const loading = ref(false)
const cleanupLoading = ref(false)
const historyData = ref<ExportHistoryResponse['data'] | null>(null)
const searchKeyword = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const downloadingTasks = ref(new Set<string>())
const deletingTasks = ref(new Set<string>())

// 进度查看相关
const showProgressModal = ref(false)
const selectedTask = ref<ExportHistoryItem | null>(null)
const taskProgress = ref<ExportTask | null>(null)
const progressPollingTimer = ref<NodeJS.Timeout | null>(null)

// 表格列配置
const columns = [
  {
    title: '文件名',
    key: 'filename',
    width: 200,
    ellipsis: true
  },
  {
    title: '状态',
    key: 'status',
    width: 120
  },
  {
    title: '记录数',
    dataIndex: 'record_count',
    key: 'record_count',
    width: 100,
    sorter: (a: ExportHistoryItem, b: ExportHistoryItem) => a.record_count - b.record_count
  },
  {
    title: '文件大小',
    key: 'file_size',
    width: 100,
    sorter: (a: ExportHistoryItem, b: ExportHistoryItem) => {
      const sizeA = parseFloat(a.file_size || '0')
      const sizeB = parseFloat(b.file_size || '0')
      return sizeA - sizeB
    }
  },
  {
    title: '创建时间',
    key: 'created_at',
    width: 180,
    sorter: (a: ExportHistoryItem, b: ExportHistoryItem) => 
      new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
  },
  {
    title: '剩余时间',
    key: 'expires_at',
    width: 120
  },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    fixed: 'right' as const
  }
]

// 计算属性
const filteredHistoryList = computed(() => {
  if (!historyData.value?.exports) return []
  
  const keyword = searchKeyword.value.toLowerCase()
  if (!keyword) return historyData.value.exports
  
  return historyData.value.exports.filter(item =>
    item.filename.toLowerCase().includes(keyword) ||
    item.task_id.toLowerCase().includes(keyword)
  )
})

// 方法
const refreshHistory = async () => {
  try {
    loading.value = true
    const data = await ExportAPI.getExportHistory(currentPage.value, pageSize.value)
    historyData.value = data
  } catch (error: any) {
    console.error('获取导出历史失败:', error)
    message.error(error.message || '获取导出历史失败')
  } finally {
    loading.value = false
  }
}

const cleanupExpired = async () => {
  try {
    cleanupLoading.value = true
    const result = await ExportAPI.cleanupExpiredFiles()
    message.success(`已清理 ${result.cleaned} 个过期文件`)
    await refreshHistory()
  } catch (error: any) {
    console.error('清理失败:', error)
    message.error(error.message || '清理过期文件失败')
  } finally {
    cleanupLoading.value = false
  }
}

const handleSearch = () => {
  // 搜索逻辑在计算属性中处理，这里可以添加额外处理
}

const handlePageChange = (page: number) => {
  currentPage.value = page
  refreshHistory()
}

const handlePageSizeChange = (current: number, size: number) => {
  pageSize.value = size
  currentPage.value = 1
  refreshHistory()
}

const downloadFile = async (item: ExportHistoryItem) => {
  try {
    downloadingTasks.value.add(item.task_id)
    await ExportAPI.downloadFile(item.task_id)
    message.success('文件下载完成')
  } catch (error: any) {
    console.error('文件下载失败:', error)
    message.error(error.message || '文件下载失败')
  } finally {
    downloadingTasks.value.delete(item.task_id)
  }
}

const deleteHistoryItem = async (taskId: string) => {
  try {
    deletingTasks.value.add(taskId)
    // 这里需要后端提供删除导出记录的API
    // await ExportAPI.deleteExportRecord(taskId)
    message.success('删除成功')
    await refreshHistory()
  } catch (error: any) {
    console.error('删除失败:', error)
    message.error(error.message || '删除失败')
  } finally {
    deletingTasks.value.delete(taskId)
  }
}

const viewTaskProgress = async (item: ExportHistoryItem) => {
  selectedTask.value = item
  showProgressModal.value = true
  
  if (item.status === 'processing') {
    await fetchTaskProgress(item.task_id)
    startProgressPolling(item.task_id)
  }
}

const fetchTaskProgress = async (taskId: string) => {
  try {
    const progress = await ExportAPI.getTaskStatus(taskId)
    taskProgress.value = progress
  } catch (error) {
    console.error('获取任务进度失败:', error)
  }
}

const startProgressPolling = (taskId: string) => {
  if (progressPollingTimer.value) {
    clearInterval(progressPollingTimer.value)
  }
  
  progressPollingTimer.value = setInterval(async () => {
    await fetchTaskProgress(taskId)
    
    // 如果任务完成或失败，停止轮询
    if (taskProgress.value && 
        (taskProgress.value.status === 'completed' || taskProgress.value.status === 'failed')) {
      stopProgressPolling()
      // 刷新历史列表
      await refreshHistory()
    }
  }, 3000)
}

const stopProgressPolling = () => {
  if (progressPollingTimer.value) {
    clearInterval(progressPollingTimer.value)
    progressPollingTimer.value = null
  }
}

const getStatusCount = (status: TaskStatus): number => {
  if (!historyData.value?.exports) return 0
  return historyData.value.exports.filter(item => item.status === status).length
}

const getTaskStatusColor = (status: TaskStatus) => {
  return TaskStatusConfig[status]?.color || 'default'
}

const getTaskStatusLabel = (status: TaskStatus) => {
  return TaskStatusConfig[status]?.label || status
}

const getTaskStatusIcon = (status: TaskStatus) => {
  const iconMap = {
    pending: ClockCircleOutlined,
    processing: LoadingOutlined,
    completed: CheckCircleOutlined,
    failed: ExclamationCircleOutlined,
    cancelled: StopOutlined,
    expired: ClockCircleOutlined
  }
  return iconMap[status] || ClockCircleOutlined
}

const formatTime = (time: string) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

const formatRelativeTime = (time: string) => {
  return dayjs(time).fromNow()
}

const isExpiringSoon = (expiresAt: string) => {
  const now = dayjs()
  const expiry = dayjs(expiresAt)
  const hoursUntilExpiry = expiry.diff(now, 'hour')
  return hoursUntilExpiry <= 2 && hoursUntilExpiry > 0
}

const getProgressStatus = (status: TaskStatus): 'success' | 'exception' | 'active' | 'normal' => {
  switch (status) {
    case 'completed':
      return 'success'
    case 'failed':
      return 'exception'
    case 'processing':
      return 'active'
    default:
      return 'normal'
  }
}

const formatEstimatedTime = (estimatedTime?: number) => {
  if (!estimatedTime) return '--'
  
  if (estimatedTime < 1) return '即将完成'
  if (estimatedTime < 60) return `约 ${Math.ceil(estimatedTime)} 分钟`
  
  const hours = Math.floor(estimatedTime / 60)
  const minutes = Math.ceil(estimatedTime % 60)
  return `约 ${hours} 小时 ${minutes} 分钟`
}

const calculateProcessingSpeed = (): string => {
  if (!taskProgress.value || !selectedTask.value) return '--'
  
  const startTime = dayjs(selectedTask.value.created_at)
  const now = dayjs()
  const elapsedMinutes = now.diff(startTime, 'minute')
  
  if (elapsedMinutes === 0) return '--'
  
  const speed = Math.round(taskProgress.value.processed_records / elapsedMinutes)
  return speed.toString()
}

const handleCancel = () => {
  emit('update:visible', false)
}

// 监听弹窗显示状态
watch(() => props.visible, (newValue) => {
  if (newValue) {
    refreshHistory()
  }
})

// 监听进度弹窗关闭
watch(() => showProgressModal.value, (newValue) => {
  if (!newValue) {
    stopProgressPolling()
    selectedTask.value = null
    taskProgress.value = null
  }
})

// 组件卸载时清理
onUnmounted(() => {
  stopProgressPolling()
})
</script>

<style scoped>
.export-history {
  padding: 0;
}

.history-toolbar {
  margin-bottom: 16px;
}

.history-stats {
  margin-bottom: 16px;
  padding: 16px;
  background: #fafafa;
  border-radius: 6px;
}

.filename-cell {
  display: flex;
  align-items: center;
}

.filename-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.time-cell {
  line-height: 1.4;
}

.expire-cell .expire-warning {
  color: #fa8c16;
  font-weight: 500;
}

.progress-detail {
  padding: 8px 0;
}

.progress-info {
  margin-top: 16px;
}

.progress-meta {
  margin-top: 16px;
}

.error-info {
  margin-top: 16px;
}

/* 表格样式优化 */
.ant-table-tbody > tr > td {
  padding: 12px 8px;
}

.ant-table-thead > tr > th {
  background: #fafafa;
  font-weight: 600;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .history-toolbar {
    margin-bottom: 12px;
  }
  
  .history-stats {
    margin-bottom: 12px;
    padding: 12px;
  }
}
</style>
