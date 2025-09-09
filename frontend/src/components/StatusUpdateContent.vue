<template>
  <div class="status-update-content" :class="{ 'compact': compact }">
    <!-- 当前状态显示 -->
    <div v-if="showCurrent" class="current-status">
      <label>当前状态：</label>
      <a-tag :color="StatusHelper.getStatusColor(currentStatus)">
        {{ currentStatus }}
      </a-tag>
      <span class="status-category">
        ({{ StatusHelper.getStatusCategory(currentStatus) }})
      </span>
    </div>

    <!-- 状态选择 -->
    <div class="form-item">
      <label>更新状态：</label>
      <a-select
        :value="selectedStatus"
        @update:value="$emit('update:selectedStatus', $event)"
        placeholder="选择新状态"
        :style="{ width: compact ? '200px' : '100%' }"
        :loading="loading"
        show-search
        :filter-option="(input: string, option: any) => 
          option.label.toLowerCase().includes(input.toLowerCase())"
      >
        <a-select-option
          v-for="option in statusOptions"
          :key="option.value"
          :value="option.value"
        >
          <div class="status-option">
            <a-tag :color="option.color" size="small">
              {{ option.label }}
            </a-tag>
            <span class="option-category">{{ option.category }}</span>
          </div>
        </a-select-option>
      </a-select>
    </div>

    <!-- 面试时间（条件显示） -->
    <div v-if="needsInterviewTime" class="form-item">
      <label>面试时间：</label>
      <a-date-picker
        :value="interviewTime"
        @update:value="$emit('update:interviewTime', $event)"
        show-time
        placeholder="选择面试时间"
        :style="{ width: compact ? '200px' : '100%' }"
        format="YYYY-MM-DD HH:mm"
      />
    </div>

    <!-- 备注信息 -->
    <div class="form-item">
      <label>备注信息：</label>
      <a-textarea
        :value="note"
        @update:value="$emit('update:note', $event)"
        placeholder="可选：添加状态变更的备注信息..."
        :rows="compact ? 2 : 3"
        :maxlength="200"
        show-count
      />
    </div>

    <!-- 操作按钮（内联模式） -->
    <div v-if="!compact && showActions" class="form-actions">
      <a-space>
        <a-button 
          type="primary" 
          :loading="loading"
          :disabled="!selectedStatus || selectedStatus === currentStatus"
          @click="$emit('update')"
        >
          确定更新
        </a-button>
        <a-button @click="$emit('cancel')">
          取消
        </a-button>
      </a-space>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { StatusHelper, type ApplicationStatus } from '../types'

// Props
interface Props {
  selectedStatus: ApplicationStatus | ''
  note: string
  interviewTime: string
  currentStatus: ApplicationStatus
  availableStatuses: ApplicationStatus[]
  loading: boolean
  compact?: boolean
  showCurrent?: boolean
  showActions?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  compact: false,
  showCurrent: true,
  showActions: true
})

// Emits
defineEmits<{
  'update:selectedStatus': [value: ApplicationStatus | '']
  'update:note': [value: string]
  'update:interviewTime': [value: string]
  update: []
  cancel: []
}>()

// 状态选项
const statusOptions = computed(() => {
  return props.availableStatuses.map(status => ({
    value: status,
    label: status,
    color: StatusHelper.getStatusColor(status),
    category: StatusHelper.getStatusCategory(status)
  }))
})

// 是否需要面试时间
const needsInterviewTime = computed(() => {
  return props.selectedStatus && isInterviewStatus(props.selectedStatus as ApplicationStatus)
})

const isInterviewStatus = (status: ApplicationStatus): boolean => {
  return ['一面中', '二面中', '三面中', 'HR面中', '笔试中'].includes(status)
}
</script>

<style scoped>
.status-update-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.status-update-content.compact {
  gap: 12px;
}

.current-status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #f0f9ff;
  border-radius: 6px;
  border: 1px solid #bae7ff;
}

.current-status label {
  color: #1890ff;
  font-weight: 500;
  margin: 0;
}

.status-category {
  color: #8c8c8c;
  font-size: 12px;
}

.form-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-item label {
  color: #262626;
  font-weight: 500;
  margin: 0;
}

.status-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.option-category {
  color: #8c8c8c;
  font-size: 12px;
}

.form-actions {
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
}

/* 紧凑模式下的布局调整 */
.status-update-content.compact .form-item {
  flex-direction: row;
  align-items: center;
  gap: 12px;
}

.status-update-content.compact .form-item label {
  min-width: 80px;
  flex-shrink: 0;
}

.status-update-content.compact .current-status {
  flex-direction: row;
  align-items: center;
}

/* 响应式适配 */
@media (max-width: 768px) {
  .status-update-content.compact .form-item {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .status-update-content.compact .form-item label {
    min-width: auto;
  }
}
</style>