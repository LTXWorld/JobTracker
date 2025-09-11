<template>
  <a-modal
    :visible="visible"
    title="导出Excel"
    width="800px"
    :confirm-loading="exportLoading"
    @ok="handleExport"
    @cancel="handleCancel"
    :ok-button-props="{ disabled: selectedFields.length === 0 || exportLoading }"
  >
    <template #footer>
      <a-space>
        <a-button @click="handleCancel">取消</a-button>
        <a-button
          type="primary"
          :loading="exportLoading"
          :disabled="selectedFields.length === 0"
          @click="handleExport"
        >
          {{ taskStatus === 'processing' ? '导出中...' : '开始导出' }}
        </a-button>
      </a-space>
    </template>

    <div class="export-dialog">
      <!-- 基本信息 -->
      <div class="export-section">
        <h4><FileExcelOutlined /> 导出信息</h4>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="导出格式">
              <a-select v-model:value="exportConfig.format" style="width: 100%">
                <a-select-option value="xlsx">Excel (.xlsx)</a-select-option>
                <!-- <a-select-option value="csv">CSV (.csv)</a-select-option> -->
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="文件名称">
              <a-input 
                v-model:value="exportConfig.options.filename" 
                placeholder="我的求职记录"
                :maxlength="50"
              />
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-row :gutter="16">
          <a-col :span="24">
            <a-statistic 
              title="将导出记录数" 
              :value="applications.length"
              suffix="条"
              :value-style="{ color: '#1890ff' }"
            />
          </a-col>
        </a-row>
      </div>

      <a-divider />

      <!-- 字段选择 -->
      <div class="export-section">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
          <h4><DatabaseOutlined /> 导出字段选择</h4>
          <a-space>
            <a-button size="small" @click="selectAllFields">全选</a-button>
            <a-button size="small" @click="selectDefaultFields">默认字段</a-button>
            <a-button size="small" @click="clearAllFields">清空</a-button>
          </a-space>
        </div>

        <div class="field-groups">
          <a-collapse v-model:activeKey="activeFieldGroups" ghost>
            <a-collapse-panel 
              v-for="group in exportFieldGroups" 
              :key="group.group" 
              :header="group.label"
            >
              <template #extra>
                <a-tag :color="getGroupSelectedCount(group) > 0 ? 'blue' : 'default'">
                  {{ getGroupSelectedCount(group) }}/{{ group.fields.length }}
                </a-tag>
              </template>

              <div class="field-checkboxes">
                <a-row :gutter="[16, 8]">
                  <a-col :span="12" v-for="field in group.fields" :key="field.field">
                    <a-checkbox 
                      v-model:checked="fieldSelections[field.field]"
                      @change="updateFieldSelection(field.field, $event)"
                      :disabled="field.required"
                    >
                      {{ field.label }}
                      <a-tag v-if="field.required" color="red" size="small">必选</a-tag>
                    </a-checkbox>
                  </a-col>
                </a-row>
              </div>
            </a-collapse-panel>
          </a-collapse>
        </div>

        <div class="selected-summary">
          <a-alert 
            :message="`已选择 ${selectedFields.length} 个字段`"
            :description="selectedFields.length === 0 ? '请至少选择一个字段' : `${selectedFieldLabels.join(', ')}`"
            :type="selectedFields.length === 0 ? 'warning' : 'info'"
            show-icon
          />
        </div>
      </div>

      <a-divider />

      <!-- 筛选条件 -->
      <div class="export-section">
        <h4><FilterOutlined /> 筛选条件</h4>
        
        <a-form layout="vertical">
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="状态筛选">
                <a-select 
                  v-model:value="exportConfig.filters.status"
                  mode="multiple"
                  placeholder="选择要导出的状态"
                  style="width: 100%"
                  :max-tag-count="2"
                >
                  <a-select-option 
                    v-for="status in availableStatuses" 
                    :key="status" 
                    :value="status"
                  >
                    <a-tag :color="getStatusColor(status)" size="small">{{ status }}</a-tag>
                  </a-select-option>
                </a-select>
              </a-form-item>
            </a-col>
            
            <a-col :span="12">
              <a-form-item label="日期范围">
                <a-range-picker 
                  v-model:value="dateRangeValue"
                  style="width: 100%"
                  :placeholder="['开始日期', '结束日期']"
                  format="YYYY-MM-DD"
                  @change="handleDateRangeChange"
                />
              </a-form-item>
            </a-col>
          </a-row>

          <a-row :gutter="16">
            <a-col :span="24">
              <a-form-item label="公司筛选">
                <a-select 
                  v-model:value="exportConfig.filters.company_names"
                  mode="multiple"
                  placeholder="选择要导出的公司"
                  style="width: 100%"
                  :max-tag-count="3"
                  show-search
                  :filter-option="filterCompanyOption"
                >
                  <a-select-option 
                    v-for="company in availableCompanies" 
                    :key="company" 
                    :value="company"
                  >
                    {{ company }}
                  </a-select-option>
                </a-select>
              </a-form-item>
            </a-col>
          </a-row>
        </a-form>

        <div class="filter-summary">
          <a-descriptions size="small" :column="2" bordered>
            <a-descriptions-item label="状态筛选">
              {{ exportConfig.filters.status?.length ? `${exportConfig.filters.status.length} 个状态` : '全部状态' }}
            </a-descriptions-item>
            <a-descriptions-item label="日期范围">
              {{ formatDateRange() }}
            </a-descriptions-item>
            <a-descriptions-item label="公司筛选" :span="2">
              {{ exportConfig.filters.company_names?.length ? `${exportConfig.filters.company_names.length} 个公司` : '全部公司' }}
            </a-descriptions-item>
          </a-descriptions>
        </div>
      </div>

      <a-divider />

      <!-- 高级选项 -->
      <div class="export-section">
        <h4><SettingOutlined /> 高级选项</h4>
        
        <a-space direction="vertical" style="width: 100%;">
          <a-checkbox v-model:checked="exportConfig.options.include_statistics">
            包含统计信息表
            <a-tooltip title="在Excel中生成一个额外的统计表，包含状态分布、成功率等信息">
              <QuestionCircleOutlined style="margin-left: 4px; color: #999;" />
            </a-tooltip>
          </a-checkbox>

          <a-checkbox v-model:checked="exportConfig.options.include_status_history" disabled>
            包含状态变更历史
            <a-tooltip title="包含每个投递记录的详细状态变更历史（暂不支持）">
              <QuestionCircleOutlined style="margin-left: 4px; color: #999;" />
            </a-tooltip>
          </a-checkbox>
        </a-space>
      </div>

      <!-- 导出进度 -->
      <div v-if="currentTask" class="export-progress">
        <a-divider />
        <div class="progress-section">
          <h4><LoadingOutlined v-if="taskStatus === 'processing'" /> 导出进度</h4>
          
          <a-progress 
            :percent="currentTask.progress" 
            :status="getProgressStatus()"
            :show-info="true"
          >
            <template #format="{ percent }">
              {{ percent }}% ({{ currentTask.processed_records }}/{{ currentTask.total_records }})
            </template>
          </a-progress>

          <div class="task-info">
            <a-descriptions size="small" :column="3">
              <a-descriptions-item label="任务ID">
                {{ currentTask.task_id }}
              </a-descriptions-item>
              <a-descriptions-item label="状态">
                <a-tag :color="getTaskStatusColor(currentTask.status)">
                  {{ getTaskStatusLabel(currentTask.status) }}
                </a-tag>
              </a-descriptions-item>
              <a-descriptions-item label="预计剩余时间">
                {{ formatEstimatedTime() }}
              </a-descriptions-item>
            </a-descriptions>
          </div>

          <div v-if="currentTask.status === 'completed'" class="download-section">
            <a-result
              status="success"
              title="导出完成！"
              :sub-title="`文件大小: ${currentTask.file_size}`"
            >
              <template #extra>
                <a-space>
                  <a-button type="primary" @click="downloadFile">
                    <template #icon><DownloadOutlined /></template>
                    下载文件
                  </a-button>
                  <a-button @click="resetExport">重新导出</a-button>
                </a-space>
              </template>
            </a-result>
          </div>

          <div v-if="currentTask.status === 'failed'" class="error-section">
            <a-result
              status="error"
              title="导出失败"
              :sub-title="currentTask.error_message"
            >
              <template #extra>
                <a-space>
                  <a-button type="primary" @click="retryExport">重试</a-button>
                  <a-button @click="resetExport">重新配置</a-button>
                </a-space>
              </template>
            </a-result>
          </div>
        </div>
      </div>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue'
import { message } from 'ant-design-vue'
import {
  FileExcelOutlined,
  DatabaseOutlined,
  FilterOutlined,
  SettingOutlined,
  QuestionCircleOutlined,
  LoadingOutlined,
  DownloadOutlined
} from '@ant-design/icons-vue'
import type { 
  JobApplication, 
  ApplicationStatus,
  ExportRequest,
  ExportTask,
  ExportableField,
  TaskStatus
} from '../types'
import { 
  StatusHelper,
  ExportableFields,
  FieldDisplayNames
} from '../types'
import { ExportAPI, ExportFieldGroups, DefaultExportFields, TaskStatusConfig } from '../api/export'
import dayjs, { type Dayjs } from 'dayjs'

interface Props {
  visible: boolean
  applications: JobApplication[]
  currentFilters?: Record<string, any>
}

interface Emits {
  (e: 'update:visible', value: boolean): void
  (e: 'success'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 响应式数据
const exportLoading = ref(false)
const currentTask = ref<ExportTask | null>(null)
const pollingTimer = ref<NodeJS.Timeout | null>(null)
const activeFieldGroups = ref<string[]>(['basic']) // 默认展开基础信息组
const dateRangeValue = ref<[Dayjs, Dayjs] | null>(null)

// 导出配置
const exportConfig = ref<ExportRequest>({
  format: 'xlsx',
  fields: [...DefaultExportFields],
  filters: {},
  options: {
    include_statistics: true,
    include_status_history: false,
    filename: '我的求职记录'
  }
})

// 字段选择状态
const fieldSelections = ref<Record<ExportableField, boolean>>({} as Record<ExportableField, boolean>)

// 导出字段分组
const exportFieldGroups = ref(ExportFieldGroups)

// 计算属性
const taskStatus = computed<TaskStatus>(() => currentTask.value?.status || 'pending')

const selectedFields = computed(() => {
  return Object.keys(fieldSelections.value).filter(
    field => fieldSelections.value[field as ExportableField]
  ) as ExportableField[]
})

const selectedFieldLabels = computed(() => {
  return selectedFields.value.map(field => FieldDisplayNames[field])
})

const availableStatuses = computed(() => {
  const statuses = new Set<ApplicationStatus>()
  props.applications.forEach(app => statuses.add(app.status))
  return Array.from(statuses).sort()
})

const availableCompanies = computed(() => {
  const companies = new Set<string>()
  props.applications.forEach(app => companies.add(app.company_name))
  return Array.from(companies).sort()
})

// 方法
const getStatusColor = (status: ApplicationStatus) => {
  return StatusHelper.getStatusColor(status)
}

const getGroupSelectedCount = (group: typeof exportFieldGroups.value[0]) => {
  return group.fields.filter(field => fieldSelections.value[field.field]).length
}

const updateFieldSelection = (field: ExportableField, event: any) => {
  fieldSelections.value[field] = event.target.checked
  exportConfig.value.fields = selectedFields.value
}

const selectAllFields = () => {
  exportFieldGroups.value.forEach(group => {
    group.fields.forEach(field => {
      fieldSelections.value[field.field] = true
    })
  })
  exportConfig.value.fields = selectedFields.value
}

const selectDefaultFields = () => {
  // 先清空所有选择
  Object.keys(fieldSelections.value).forEach(field => {
    fieldSelections.value[field as ExportableField] = false
  })
  
  // 选择默认字段
  DefaultExportFields.forEach(field => {
    fieldSelections.value[field] = true
  })
  
  exportConfig.value.fields = selectedFields.value
}

const clearAllFields = () => {
  // 保留必选字段
  exportFieldGroups.value.forEach(group => {
    group.fields.forEach(field => {
      if (!field.required) {
        fieldSelections.value[field.field] = false
      }
    })
  })
  exportConfig.value.fields = selectedFields.value
}

const handleDateRangeChange = (dates: [Dayjs, Dayjs] | null) => {
  if (dates) {
    exportConfig.value.filters.date_range = {
      start: dates[0].format('YYYY-MM-DD'),
      end: dates[1].format('YYYY-MM-DD')
    }
  } else {
    delete exportConfig.value.filters.date_range
  }
}

const formatDateRange = () => {
  if (exportConfig.value.filters.date_range) {
    return `${exportConfig.value.filters.date_range.start} 至 ${exportConfig.value.filters.date_range.end}`
  }
  return '全部时间'
}

const filterCompanyOption = (input: string, option: any) => {
  return option.value.toLowerCase().includes(input.toLowerCase())
}

const handleExport = async () => {
  try {
    exportLoading.value = true
    
    // 验证配置
    if (selectedFields.value.length === 0) {
      message.warning('请至少选择一个导出字段')
      return
    }

    // 启动导出任务
    const task = await ExportAPI.startExport(exportConfig.value)
    currentTask.value = task
    
    message.success('导出任务已启动')
    
    // 如果是同步处理（小数据量），直接下载
    if (task.status === 'completed') {
      await downloadFile()
      emit('success')
      return
    }
    
    // 异步处理，开始轮询状态
    startStatusPolling()
    
  } catch (error: any) {
    console.error('导出失败:', error)
    message.error(error.message || '导出失败，请重试')
  } finally {
    exportLoading.value = false
  }
}

const startStatusPolling = () => {
  if (pollingTimer.value) {
    clearInterval(pollingTimer.value)
  }
  
  pollingTimer.value = setInterval(async () => {
    if (!currentTask.value) return
    
    try {
      const updatedTask = await ExportAPI.getTaskStatus(currentTask.value.task_id)
      currentTask.value = updatedTask
      
      // 如果任务完成或失败，停止轮询
      if (updatedTask.status === 'completed' || updatedTask.status === 'failed') {
        stopStatusPolling()
        
        if (updatedTask.status === 'completed') {
          message.success('导出完成！')
        } else if (updatedTask.status === 'failed') {
          message.error('导出失败：' + updatedTask.error_message)
        }
      }
    } catch (error) {
      console.error('获取任务状态失败:', error)
    }
  }, 2000) // 每2秒轮询一次
}

const stopStatusPolling = () => {
  if (pollingTimer.value) {
    clearInterval(pollingTimer.value)
    pollingTimer.value = null
  }
}

const downloadFile = async () => {
  if (!currentTask.value) return
  
  try {
    await ExportAPI.downloadFile(currentTask.value.task_id)
    message.success('文件下载完成')
    emit('success')
  } catch (error: any) {
    console.error('文件下载失败:', error)
    message.error(error.message || '文件下载失败')
  }
}

const retryExport = () => {
  currentTask.value = null
  handleExport()
}

const resetExport = () => {
  currentTask.value = null
  stopStatusPolling()
}

const handleCancel = () => {
  // 如果有进行中的任务，询问是否取消
  if (currentTask.value && currentTask.value.status === 'processing') {
    // 可以添加确认对话框
  }
  
  stopStatusPolling()
  emit('update:visible', false)
}

const getProgressStatus = (): 'success' | 'exception' | 'active' | 'normal' => {
  if (!currentTask.value) return 'normal'
  
  switch (currentTask.value.status) {
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

const getTaskStatusColor = (status: TaskStatus) => {
  return TaskStatusConfig[status]?.color || 'default'
}

const getTaskStatusLabel = (status: TaskStatus) => {
  return TaskStatusConfig[status]?.label || status
}

const formatEstimatedTime = () => {
  if (!currentTask.value?.estimated_time) return '--'
  
  const minutes = currentTask.value.estimated_time
  if (minutes < 1) return '即将完成'
  if (minutes < 60) return `约 ${Math.ceil(minutes)} 分钟`
  
  const hours = Math.floor(minutes / 60)
  const remainingMinutes = Math.ceil(minutes % 60)
  return `约 ${hours} 小时 ${remainingMinutes} 分钟`
}

// 初始化字段选择状态
const initFieldSelections = () => {
  const selections: Record<ExportableField, boolean> = {} as Record<ExportableField, boolean>
  
  exportFieldGroups.value.forEach(group => {
    group.fields.forEach(field => {
      selections[field.field] = DefaultExportFields.includes(field.field) || field.required || false
    })
  })
  
  fieldSelections.value = selections
  exportConfig.value.fields = selectedFields.value
}

// 监听对话框显示状态
watch(() => props.visible, (newValue) => {
  if (newValue) {
    // 重置状态
    currentTask.value = null
    exportLoading.value = false
    initFieldSelections()
    
    // 应用当前筛选条件到导出配置
    if (props.currentFilters) {
      if (props.currentFilters.status) {
        exportConfig.value.filters.status = [props.currentFilters.status]
      }
      if (props.currentFilters.dateRange) {
        const [start, end] = props.currentFilters.dateRange
        exportConfig.value.filters.date_range = {
          start: start.format('YYYY-MM-DD'),
          end: end.format('YYYY-MM-DD')
        }
        dateRangeValue.value = [start, end]
      }
    }
  } else {
    // 清理轮询
    stopStatusPolling()
  }
})

// 组件卸载时清理
onUnmounted(() => {
  stopStatusPolling()
})

// 初始化
initFieldSelections()
</script>

<style scoped>
.export-dialog {
  max-height: 600px;
  overflow-y: auto;
}

.export-section {
  margin-bottom: 24px;
}

.export-section h4 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  color: #1890ff;
  font-weight: 600;
}

.field-groups {
  margin-bottom: 16px;
}

.field-checkboxes {
  padding: 8px 0;
}

.selected-summary {
  margin-top: 16px;
}

.filter-summary {
  margin-top: 16px;
}

.export-progress {
  background: #fafafa;
  padding: 16px;
  border-radius: 6px;
  margin-top: 16px;
}

.progress-section h4 {
  color: #52c41a;
}

.task-info {
  margin: 16px 0;
}

.download-section,
.error-section {
  margin-top: 24px;
}

/* 滚动条样式 */
.export-dialog::-webkit-scrollbar {
  width: 6px;
}

.export-dialog::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.export-dialog::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.export-dialog::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .field-checkboxes .ant-col {
    span: 24 !important;
  }
}
</style>