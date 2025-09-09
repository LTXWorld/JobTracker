<template>
  <a-modal
    :visible="visible"
    title="偏好设置" 
    width="600px"
    :mask-closable="false"
    @ok="handleSave"
    @cancel="handleCancel"
    @update:visible="emit('update:visible', $event)"
  >
    <div class="preferences-content">
      <a-form
        :model="formData"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <!-- 通知设置 -->
        <a-divider orientation="left">通知设置</a-divider>
        
        <a-form-item label="邮件通知">
          <a-switch
            v-model:checked="formData.notification_settings.email_enabled"
            checked-children="开"
            un-checked-children="关"
          />
        </a-form-item>

        <a-form-item label="推送通知">
          <a-switch
            v-model:checked="formData.notification_settings.push_enabled"
            checked-children="开"
            un-checked-children="关"
          />
        </a-form-item>

        <a-form-item label="提醒频率">
          <a-radio-group v-model:value="formData.notification_settings.reminder_frequency">
            <a-radio value="daily">每日</a-radio>
            <a-radio value="weekly">每周</a-radio>
            <a-radio value="custom">自定义</a-radio>
          </a-radio-group>
        </a-form-item>

        <!-- 显示偏好 -->
        <a-divider orientation="left">显示偏好</a-divider>

        <a-form-item label="显示持续时间">
          <a-switch
            v-model:checked="formData.display_preferences.show_durations"
            checked-children="显示"
            un-checked-children="隐藏"
          />
        </a-form-item>

        <a-form-item label="显示成功概率">
          <a-switch
            v-model:checked="formData.display_preferences.show_probabilities"
            checked-children="显示"
            un-checked-children="隐藏"
          />
        </a-form-item>

        <a-form-item label="紧凑时间轴">
          <a-switch
            v-model:checked="formData.display_preferences.timeline_compact"
            checked-children="紧凑"
            un-checked-children="标准"
          />
        </a-form-item>

        <a-form-item label="看板显示计数">
          <a-switch
            v-model:checked="formData.display_preferences.kanban_show_counts"
            checked-children="显示"
            un-checked-children="隐藏"
          />
        </a-form-item>

        <!-- 自动提醒规则 -->
        <a-divider orientation="left">自动提醒规则</a-divider>

        <div class="reminder-rules">
          <div
            v-for="(rule, index) in formData.auto_reminder_rules"
            :key="index"
            class="rule-item"
          >
            <a-row :gutter="12" align="middle">
              <a-col :span="8">
                <a-select
                  v-model:value="rule.status"
                  placeholder="选择状态"
                  size="small"
                >
                  <a-select-option
                    v-for="status in availableStatuses"
                    :key="status"
                    :value="status"
                  >
                    {{ status }}
                  </a-select-option>
                </a-select>
              </a-col>
              <a-col :span="6">
                <a-input-number
                  v-model:value="rule.delay_days"
                  :min="1"
                  :max="30"
                  placeholder="天数"
                  size="small"
                />
              </a-col>
              <a-col :span="6">
                <a-switch
                  v-model:checked="rule.enabled"
                  size="small"
                  checked-children="启用"
                  un-checked-children="禁用"
                />
              </a-col>
              <a-col :span="4">
                <a-button
                  type="link"
                  size="small"
                  danger
                  @click="removeReminderRule(index)"
                >
                  删除
                </a-button>
              </a-col>
            </a-row>
          </div>
          
          <a-button
            type="dashed"
            block
            size="small"
            @click="addReminderRule"
          >
            <template #icon><PlusOutlined /></template>
            添加提醒规则
          </a-button>
        </div>
      </a-form>
    </div>

    <template #footer>
      <a-space>
        <a-button @click="handleCancel">取消</a-button>
        <a-button type="default" @click="handleReset">重置为默认</a-button>
        <a-button type="primary" :loading="loading" @click="handleSave">
          保存设置
        </a-button>
      </a-space>
    </template>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { useStatusTrackingStore } from '../stores/statusTracking'
import { ApplicationStatus, type UserStatusPreferences } from '../types'
import { message } from 'ant-design-vue'

// Props
interface Props {
  visible: boolean
}

const props = defineProps<Props>()

// Emits
const emit = defineEmits<{
  'update:visible': [value: boolean]
  updated: []
}>()

// Store
const statusTrackingStore = useStatusTrackingStore()

// 响应式数据
const loading = ref(false)
const formData = reactive({
  notification_settings: {
    email_enabled: true,
    push_enabled: true,
    reminder_frequency: 'daily' as 'daily' | 'weekly' | 'custom'
  },
  display_preferences: {
    show_durations: true,
    show_probabilities: true,
    timeline_compact: false,
    kanban_show_counts: true
  },
  auto_reminder_rules: [] as Array<{
    status: string;
    delay_days: number;
    enabled: boolean;
  }>
})

// 可用状态列表
const availableStatuses = Object.values(ApplicationStatus)

// 默认设置
const defaultPreferences = {
  notification_settings: {
    email_enabled: true,
    push_enabled: true,
    reminder_frequency: 'daily' as const
  },
  display_preferences: {
    show_durations: true,
    show_probabilities: true,
    timeline_compact: false,
    kanban_show_counts: true
  },
  auto_reminder_rules: [
    { status: '已投递', delay_days: 7, enabled: true },
    { status: '简历筛选中', delay_days: 5, enabled: true },
    { status: '一面中', delay_days: 3, enabled: true }
  ]
}

// 方法
const loadPreferences = async () => {
  try {
    const preferences = await statusTrackingStore.fetchUserPreferences()
    if (preferences?.preference_config) {
      Object.assign(formData, preferences.preference_config)
    }
  } catch (error) {
    console.error('加载用户偏好失败:', error)
    // 使用默认设置
    Object.assign(formData, defaultPreferences)
  }
}

const handleSave = async () => {
  loading.value = true
  try {
    await statusTrackingStore.updateUserPreferences(formData)
    message.success('偏好设置保存成功')
    emit('updated')
    emit('update:visible', false)
  } catch (error) {
    message.error('保存失败: ' + (error as Error).message)
  } finally {
    loading.value = false
  }
}

const handleCancel = () => {
  emit('update:visible', false)
  // 重新加载设置
  loadPreferences()
}

const handleReset = () => {
  Object.assign(formData, defaultPreferences)
  message.info('已重置为默认设置')
}

const addReminderRule = () => {
  formData.auto_reminder_rules.push({
    status: '已投递',
    delay_days: 7,
    enabled: true
  })
}

const removeReminderRule = (index: number) => {
  formData.auto_reminder_rules.splice(index, 1)
}

// 监听器
watch(() => props.visible, (visible) => {
  if (visible) {
    loadPreferences()
  }
})

// 生命周期
onMounted(() => {
  if (props.visible) {
    loadPreferences()
  }
})
</script>

<style scoped>
.preferences-content {
  max-height: 600px;
  overflow-y: auto;
}

.reminder-rules {
  border: 1px solid #f0f0f0;
  border-radius: 6px;
  padding: 16px;
  background: #fafafa;
}

.rule-item {
  margin-bottom: 12px;
  padding: 8px;
  background: white;
  border-radius: 4px;
  border: 1px solid #f0f0f0;
}

.rule-item:last-child {
  margin-bottom: 16px;
}

:deep(.ant-divider-horizontal.ant-divider-with-text-left) {
  margin: 24px 0 16px 0;
}

:deep(.ant-divider-horizontal.ant-divider-with-text-left::before) {
  width: 5%;
}

:deep(.ant-divider-horizontal.ant-divider-with-text-left::after) {
  width: 95%;
}
</style>