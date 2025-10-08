<template>
  <div class="reminder-page">
    <a-row :gutter="24">
      <!-- 左侧：提醒列表 -->
      <a-col :xs="24" :lg="16">
        <ReminderManager ref="reminderManager" />
      </a-col>

      <!-- 右侧：邮箱授权 + 快捷操作 -->
      <a-col :xs="24" :lg="8">
        <!-- 邮箱授权 -->
        <a-card title="邮箱授权" style="margin-bottom: 24px" :loading="mailboxLoading">
          <a-alert
            v-if="mailboxInfo"
            :message="mailboxStatusText"
            :description="mailboxStatusDescription"
            :type="mailboxAlertType"
            show-icon
            style="margin-bottom: 16px"
          />

          <a-form layout="vertical">
            <a-form-item label="邮箱地址" required>
              <a-input
                v-model:value="mailboxForm.email_address"
                placeholder="例如 123456@qq.com"
              />
            </a-form-item>

            <a-form-item label="协议">
              <a-select v-model:value="mailboxForm.protocol" disabled>
                <a-select-option value="imap">IMAP</a-select-option>
              </a-select>
            </a-form-item>

            <a-form-item label="收件服务器" required>
              <a-input v-model:value="mailboxForm.host" />
            </a-form-item>

            <a-form-item label="端口" required>
              <a-input-number v-model:value="mailboxForm.port" :min="1" :max="65535" style="width: 100%" />
            </a-form-item>

            <a-form-item label="SSL 加密">
              <a-switch v-model:checked="mailboxForm.use_ssl" />
            </a-form-item>

            <a-form-item label="授权码" required>
              <a-input-password
                v-model:value="mailboxForm.authorization_code"
                placeholder="请输入邮箱授权码"
              />
            </a-form-item>

            <div class="mailbox-actions">
              <a-space>
                <a-button size="small" @click="useQQPreset">填入 QQ 邮箱预设</a-button>
                <a-button type="primary" :loading="mailboxSaving" @click="submitMailbox">保存绑定</a-button>
                <a-button
                  v-if="mailboxInfo"
                  danger
                  :loading="mailboxSaving"
                  @click="removeMailbox"
                >
                  解除绑定
                </a-button>
              </a-space>
            </div>
          </a-form>
        </a-card>

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
            <a-col :span="8">
              <a-statistic 
                title="面试提醒" 
                :value="interviewCount"
              >
                <template #prefix>
                  <TeamOutlined />
                </template>
              </a-statistic>
            </a-col>
            <a-col :span="8">
              <a-statistic 
                title="笔试提醒" 
                :value="writtenCount"
              >
                <template #prefix>
                  <CalendarOutlined />
                </template>
              </a-statistic>
            </a-col>
            <a-col :span="8">
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
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { 
  PlusOutlined, CalendarOutlined, ClockCircleOutlined,
  TeamOutlined, PhoneOutlined
} from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import ReminderManager from '../components/ReminderManager.vue'
import { useJobApplicationStore } from '../stores/jobApplication'
import { useMailboxStore } from '../stores/mailbox'
import { storeToRefs } from 'pinia'
import dayjs from 'dayjs'

const jobStore = useJobApplicationStore()
const { applications } = storeToRefs(jobStore)

const reminderManager = ref()
const selectedApplicationId = ref<number | null>(null)

const mailboxStore = useMailboxStore()
const { mailbox: mailboxInfo, loading: mailboxLoading, saving: mailboxSaving } = storeToRefs(mailboxStore)

const mailboxForm = reactive({
  email_address: '',
  provider: 'qq',
  protocol: 'imap',
  host: 'imap.qq.com',
  port: 993,
  use_ssl: true,
  authorization_code: ''
})

const mailboxStatusText = computed(() => {
  if (!mailboxInfo.value) {
    return '尚未绑定邮箱，填写 QQ 邮箱授权后可自动同步邀请邮件'
  }
  switch (mailboxInfo.value.status) {
    case 'active':
      return `邮箱 ${mailboxInfo.value.email_address} 同步正常`
    case 'error':
      return `邮箱 ${mailboxInfo.value.email_address} 同步异常`
    case 'pending_review':
      return `邮箱 ${mailboxInfo.value.email_address} 待人工检查`
    default:
      return `邮箱 ${mailboxInfo.value.email_address} 当前状态：${mailboxInfo.value.status}`
  }
})

const mailboxStatusDescription = computed(() => {
  if (!mailboxInfo.value) return ''
  const last = mailboxInfo.value.last_synced_at
  const timeText = last ? `最近同步：${dayjs(last).format('YYYY-MM-DD HH:mm')}` : '尚未同步'
  if (mailboxInfo.value.error_message) {
    return `${mailboxInfo.value.error_message}，${timeText}`
  }
  return timeText
})

const mailboxAlertType = computed(() => {
  if (!mailboxInfo.value) return 'info'
  if (mailboxInfo.value.requires_attention) return 'warning'
  return 'success'
})

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
    app.reminder_enabled && app.interview_type !== '笔试' && app.interview_time
  ).length
)

const followUpCount = computed(() => 
  applications.value.filter(app => 
    app.follow_up_date && app.reminder_enabled && !app.interview_time
  ).length
)

const writtenCount = computed(() =>
  applications.value.filter(app =>
    app.reminder_enabled && app.interview_type === '笔试'
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

const fillFormFromMailbox = () => {
  if (!mailboxInfo.value) {
    useQQPreset()
    return
  }
  mailboxForm.email_address = mailboxInfo.value.email_address
  mailboxForm.host = mailboxInfo.value.host
  mailboxForm.port = mailboxInfo.value.port
  mailboxForm.protocol = mailboxInfo.value.protocol
  mailboxForm.use_ssl = mailboxInfo.value.use_ssl
  mailboxForm.authorization_code = ''
}

const useQQPreset = () => {
  mailboxForm.provider = 'qq'
  mailboxForm.protocol = 'imap'
  mailboxForm.host = 'imap.qq.com'
  mailboxForm.port = 993
  mailboxForm.use_ssl = true
}

const submitMailbox = async () => {
  if (!mailboxForm.email_address) {
    message.warning('请输入邮箱地址')
    return
  }
  if (!mailboxForm.host || !mailboxForm.port) {
    message.warning('请完整填写服务器信息')
    return
  }
  if (!mailboxForm.authorization_code) {
    message.warning('请输入授权码')
    return
  }

  try {
    await mailboxStore.bindMailbox({
      email_address: mailboxForm.email_address,
      provider: mailboxForm.provider,
      protocol: mailboxForm.protocol,
      host: mailboxForm.host,
      port: mailboxForm.port,
      use_ssl: mailboxForm.use_ssl,
      authorization_code: mailboxForm.authorization_code
    })
    mailboxForm.authorization_code = ''
  } catch (error) {
    /* message 已提示 */
  }
}

const removeMailbox = async () => {
  try {
    await mailboxStore.removeMailbox()
    mailboxForm.authorization_code = ''
    useQQPreset()
  } catch (error) {
    /* message 已提示 */
  }
}

watch(mailboxInfo, () => {
  fillFormFromMailbox()
}, { immediate: true })

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
  mailboxStore.fetchMailbox()
  loadSettings()
})
</script>

<style scoped>
.reminder-page {
  padding: 24px;
  background: var(--bg-page);
  min-height: calc(100vh - 48px - 56px - 70px);
}

@media (max-width: 992px) {
  .reminder-page {
    padding: 16px;
  }
}

.reminder-page :deep(.ant-empty) {
  background: var(--bg-card);
  padding: 24px;
  border-radius: 8px;
}

.mailbox-actions {
  display: flex;
  justify-content: flex-end;
}
</style>
