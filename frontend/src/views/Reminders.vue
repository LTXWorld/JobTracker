<template>
  <div class="reminder-page">
    <a-row :gutter="24">
      <!-- 左侧：提醒列表 -->
      <a-col :xs="24" :lg="16">
        <ReminderManager ref="reminderManager" />
      </a-col>
      
      <!-- 右侧：快速操作和统计 -->
      <a-col :xs="24" :lg="8">
        <!-- 快速添加提醒 -->
        <a-card title="快速添加提醒" style="margin-bottom: 24px">
          <a-form layout="vertical">
            <a-form-item label="选择投递记录">
              <a-select
                v-model:value="selectedApplicationId"
                placeholder="选择要设置提醒的投递"
                style="width: 100%"
                :options="applicationOptions"
                show-search
                :filter-option="filterOption"
              />
            </a-form-item>
            
            <a-button 
              type="primary" 
              block
              :disabled="!selectedApplicationId"
              @click="setReminder"
            >
              <PlusOutlined /> 设置提醒
            </a-button>
          </a-form>
        </a-card>

        <!-- 提醒统计 -->
        <a-card title="提醒统计">
          <a-row :gutter="16">
            <a-col :span="12">
              <a-statistic 
                title="今日待办" 
                :value="todayCount"
                :value-style="{ color: '#3f8600' }"
              >
                <template #prefix>
                  <CalendarOutlined />
                </template>
              </a-statistic>
            </a-col>
            <a-col :span="12">
              <a-statistic 
                title="本周待办" 
                :value="weekCount"
                :value-style="{ color: '#1890ff' }"
              >
                <template #prefix>
                  <ClockCircleOutlined />
                </template>
              </a-statistic>
            </a-col>
          </a-row>
          
          <a-divider />
          
          <a-row :gutter="16">
            <a-col :span="12">
              <a-statistic 
                title="面试提醒" 
                :value="interviewCount"
              >
                <template #prefix>
                  <TeamOutlined />
                </template>
              </a-statistic>
            </a-col>
            <a-col :span="12">
              <a-statistic 
                title="跟进提醒" 
                :value="followUpCount"
              >
                <template #prefix>
                  <PhoneOutlined />
                </template>
              </a-statistic>
            </a-col>
          </a-row>
        </a-card>

        <!-- 提醒设置 -->
        <a-card title="提醒设置" style="margin-top: 24px">
          <a-form layout="vertical">
            <a-form-item label="默认提前提醒时间">
              <a-select v-model:value="settings.defaultReminderTime" style="width: 100%">
                <a-select-option :value="15">15分钟</a-select-option>
                <a-select-option :value="30">30分钟</a-select-option>
                <a-select-option :value="60">1小时</a-select-option>
                <a-select-option :value="120">2小时</a-select-option>
                <a-select-option :value="1440">1天</a-select-option>
              </a-select>
            </a-form-item>
            
            <a-form-item label="提醒方式">
              <a-checkbox-group v-model:value="settings.notificationMethods">
                <a-checkbox value="browser">浏览器通知</a-checkbox>
                <a-checkbox value="sound">声音提醒</a-checkbox>
                <a-checkbox value="email">邮件提醒</a-checkbox>
              </a-checkbox-group>
            </a-form-item>
            
            <a-form-item label="自动提醒">
              <a-switch v-model:checked="settings.autoReminder" />
              <span style="margin-left: 10px">面试状态自动创建提醒</span>
            </a-form-item>
            
            <a-button type="primary" @click="saveSettings">保存设置</a-button>
          </a-form>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { 
  PlusOutlined, CalendarOutlined, ClockCircleOutlined,
  TeamOutlined, PhoneOutlined
} from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import ReminderManager from '../components/ReminderManager.vue'
import { useJobApplicationStore } from '../stores/jobApplication'
import { storeToRefs } from 'pinia'
import dayjs from 'dayjs'

const jobStore = useJobApplicationStore()
const { applications } = storeToRefs(jobStore)

const reminderManager = ref()
const selectedApplicationId = ref<number | null>(null)

// 设置
const settings = ref({
  defaultReminderTime: 30,
  notificationMethods: ['browser', 'sound'],
  autoReminder: true
})

// 应用选项
const applicationOptions = computed(() => 
  applications.value.map(app => ({
    value: app.id,
    label: `${app.company_name} - ${app.position_title}`
  }))
)

// 统计数据
const todayCount = computed(() => {
  const today = dayjs()
  return applications.value.filter(app => 
    app.reminder_time && 
    dayjs(app.reminder_time).isSame(today, 'day')
  ).length
})

const weekCount = computed(() => {
  const startOfWeek = dayjs().startOf('week')
  const endOfWeek = dayjs().endOf('week')
  return applications.value.filter(app => 
    app.reminder_time && 
    dayjs(app.reminder_time).isAfter(startOfWeek) &&
    dayjs(app.reminder_time).isBefore(endOfWeek)
  ).length
})

const interviewCount = computed(() => 
  applications.value.filter(app => 
    app.interview_time && app.reminder_enabled
  ).length
)

const followUpCount = computed(() => 
  applications.value.filter(app => 
    app.follow_up_date && app.reminder_enabled && !app.interview_time
  ).length
)

// 筛选选项
const filterOption = (input: string, option: any) => {
  return option.label.toLowerCase().includes(input.toLowerCase())
}

// 设置提醒
const setReminder = () => {
  if (!selectedApplicationId.value) {
    message.warning('请先选择投递记录')
    return
  }
  
  const application = applications.value.find(
    app => app.id === selectedApplicationId.value
  )
  
  if (application) {
    reminderManager.value?.openReminderModal(application)
  }
}

// 保存设置
const saveSettings = () => {
  // 保存到本地存储
  localStorage.setItem('reminderSettings', JSON.stringify(settings.value))
  message.success('设置已保存')
}

// 加载设置
const loadSettings = () => {
  const saved = localStorage.getItem('reminderSettings')
  if (saved) {
    try {
      settings.value = JSON.parse(saved)
    } catch (e) {
      console.error('Failed to load settings')
    }
  }
}

onMounted(() => {
  jobStore.fetchApplications()
  loadSettings()
})
</script>

<style scoped>
.reminder-page {
  padding: 24px;
  background: #f0f2f5;
  min-height: calc(100vh - 48px - 56px - 70px);
}

@media (max-width: 992px) {
  .reminder-page {
    padding: 16px;
  }
}
</style>