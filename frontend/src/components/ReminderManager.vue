<template>
  <div class="reminder-container">
    <a-card :loading="cardLoading" title="提醒中心">
      <template #extra>
        <a-badge :count="activeBadgeCount" :showZero="true">
          <BellOutlined style="font-size: 20px" />
        </a-badge>
      </template>

      <a-tabs v-model:activeKey="activeTab">
        <a-tab-pane key="reminders" tab="待办提醒">
          <div class="filter-bar">
            <a-radio-group v-model:value="filterType" button-style="solid">
              <a-radio-button value="all">全部</a-radio-button>
              <a-radio-button value="interview">面试提醒</a-radio-button>
              <a-radio-button value="written">笔试提醒</a-radio-button>
              <a-radio-button value="follow_up">跟进提醒</a-radio-button>
              <a-radio-button value="today">今日待办</a-radio-button>
              <a-radio-button value="upcoming">即将到来</a-radio-button>
            </a-radio-group>
          </div>

          <a-divider />

          <a-list
            v-if="filteredReminders.length > 0"
            :data-source="filteredReminders"
            item-layout="horizontal"
          >
            <template #renderItem="{ item }">
              <a-list-item>
                <template #actions>
                  <a-button
                    size="small"
                    type="primary"
                    @click="viewApplication(item.application_id)"
                  >
                    查看详情
                  </a-button>
                  <a-button
                    size="small"
                    danger
                    @click="dismissReminder(item)"
                  >
                    忽略
                  </a-button>
                </template>

                <a-list-item-meta>
                  <template #avatar>
                    <a-avatar
                      :style="{
                        backgroundColor: item.type === 'interview' ? '#1890ff' : '#52c41a'
                      }"
                    >
                      <template v-if="item.type === 'interview'">
                        <CalendarOutlined />
                      </template>
                      <template v-else>
                        <ClockCircleOutlined />
                      </template>
                    </a-avatar>
                  </template>

                  <template #title>
                    <div class="reminder-title">
                      <span>{{ item.company_name }} - {{ item.position_title }}</span>
                      <a-tag :color="getReminderColor(item)">
                        {{ getReminderTypeText(item.type) }}
                      </a-tag>
                    </div>
                  </template>

                  <template #description>
                    <div class="reminder-info">
                      <p>
                        <ClockCircleOutlined />
                        提醒时间：{{ formatDateTime(item.reminder_time) }}
                      </p>
                      <p v-if="item.interview_time">
                        <CalendarOutlined />
                        面试时间：{{ formatDateTime(item.interview_time) }}
                      </p>
                      <p v-if="item.message">
                        <MessageOutlined />
                        备注：{{ item.message }}
                      </p>
                      <a-tag :color="getUrgencyColor(item.reminder_time)">
                        {{ getTimeRemaining(item.reminder_time) }}
                      </a-tag>
                    </div>
                  </template>
                </a-list-item-meta>
              </a-list-item>
            </template>
          </a-list>

          <a-empty
            v-else
            description="暂无待办提醒"
            :image="Empty.PRESENTED_IMAGE_SIMPLE"
          />
        </a-tab-pane>

        <a-tab-pane key="mailEvents" tab="待确认邮件">
          <div class="mail-events-header">
            <a-space>
              <a-button type="link" @click="refreshMailEvents" :loading="mailEventsLoading">
                刷新
              </a-button>
            </a-space>
          </div>
          <a-spin :spinning="mailEventsLoading">
            <template v-if="mailEvents.length > 0">
              <a-list :data-source="mailEvents" item-layout="vertical">
                <template #renderItem="{ item }">
                  <a-list-item>
                    <template #actions>
                      <a-button
                        size="small"
                        type="primary"
                        :disabled="!item.application_id"
                        @click="openReminderFromEvent(item)"
                      >
                        设置提醒
                      </a-button>
                      <a-button size="small" @click="markEventProcessed(item)">
                        标记完成
                      </a-button>
                      <a-button size="small" danger @click="dismissEvent(item)">
                        忽略
                      </a-button>
                    </template>
                    <a-list-item-meta>
                      <template #avatar>
                        <a-avatar style="background-color: #faad14">
                          <MailOutlined />
                        </a-avatar>
                      </template>
                      <template #title>
                        <div class="mail-event-title">
                          <span>{{ item.subject }}</span>
                          <a-tag :color="getClassificationColor(item.classification)">
                            {{ getClassificationText(item.classification) }}
                          </a-tag>
                          <a-tag v-if="item.application" color="blue">
                            已匹配：{{ item.application.company_name }}
                          </a-tag>
                        </div>
                      </template>
                      <template #description>
                        <div class="mail-event-info">
                          <p>
                            <FieldTimeOutlined />
                            接收时间：{{ formatDateTime(item.received_at) }}
                          </p>
                          <p>
                            <MessageOutlined />
                            发件人：{{ item.sender }}
                          </p>
                          <p v-if="item.snippet" class="mail-event-snippet">
                            {{ item.snippet }}
                          </p>
                          <div class="mail-event-confidence">
                            <span>识别置信度</span>
                            <a-progress :percent="getConfidencePercent(item.confidence)" :show-info="true" />
                          </div>
                          <p v-if="item.payload.exam_link">
                            <LinkOutlined />
                            笔试链接：
                            <a :href="item.payload.exam_link" target="_blank" rel="noopener">
                              前往笔试
                            </a>
                          </p>
                          <p v-if="item.payload.meeting_link">
                            <LinkOutlined />
                            会议链接：
                            <a :href="item.payload.meeting_link" target="_blank" rel="noopener">
                              立即加入
                            </a>
                          </p>
                          <p v-if="item.payload.meeting_id">
                            <CalendarOutlined />
                            会议号：{{ item.payload.meeting_id }}
                          </p>
                          <div v-if="item.payload.raw_links && item.payload.raw_links.length" class="mail-event-links">
                            <span>相关链接：</span>
                            <a-tag v-for="link in item.payload.raw_links" :key="link" color="purple">
                              <a :href="link" target="_blank" rel="noopener">链接</a>
                            </a-tag>
                          </div>
                        </div>
                      </template>
                    </a-list-item-meta>
                  </a-list-item>
                </template>
              </a-list>
            </template>
            <a-empty
              v-else
              description="暂无待确认事件"
              :image="Empty.PRESENTED_IMAGE_SIMPLE"
            />
          </a-spin>
        </a-tab-pane>
      </a-tabs>
    </a-card>

    <a-modal
      v-model:visible="showReminderModal"
      title="设置提醒"
      @ok="saveReminder"
      @cancel="cancelReminder"
      width="600px"
    >
      <a-form
        :model="reminderForm"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="提醒类型" required>
          <a-radio-group v-model:value="reminderForm.type">
            <a-radio value="interview">面试提醒</a-radio>
            <a-radio value="written">笔试提醒</a-radio>
            <a-radio value="follow_up">跟进提醒</a-radio>
          </a-radio-group>
        </a-form-item>

        <a-form-item
          v-if="reminderForm.type === 'interview'"
          label="面试时间"
          required
        >
          <a-date-picker
            v-model:value="reminderForm.interview_time"
            show-time
            placeholder="选择面试时间"
            style="width: 100%"
            :format="'YYYY-MM-DD HH:mm'"
          />
        </a-form-item>

        <a-form-item
          v-if="reminderForm.type === 'written'"
          label="笔试时间"
          required
        >
          <a-date-picker
            v-model:value="reminderForm.written_time"
            show-time
            placeholder="选择笔试时间"
            style="width: 100%"
            :format="'YYYY-MM-DD HH:mm'"
          />
        </a-form-item>

        <a-form-item label="提醒时间" required>
          <a-date-picker
            v-model:value="reminderForm.reminder_time"
            show-time
            placeholder="选择提醒时间"
            style="width: 100%"
            :format="'YYYY-MM-DD HH:mm'"
          />
        </a-form-item>

        <a-form-item label="提醒方式">
          <a-checkbox-group v-model:value="reminderForm.notification_methods">
            <a-checkbox value="browser">浏览器通知</a-checkbox>
            <a-checkbox value="email">邮件提醒</a-checkbox>
            <a-checkbox value="sound">声音提醒</a-checkbox>
          </a-checkbox-group>
        </a-form-item>

        <a-form-item label="备注信息">
          <a-textarea
            v-model:value="reminderForm.message"
            placeholder="输入提醒备注"
            :rows="3"
          />
        </a-form-item>

        <a-form-item label="重复提醒">
          <a-switch v-model:checked="reminderForm.repeat" />
          <span style="margin-left: 10px">面试前30分钟再次提醒</span>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import {
  BellOutlined,
  CalendarOutlined,
  ClockCircleOutlined,
  MessageOutlined,
  MailOutlined,
  LinkOutlined,
  FieldTimeOutlined
} from '@ant-design/icons-vue'
import { message, Empty } from 'ant-design-vue'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'
import { useJobApplicationStore } from '../stores/jobApplication'
import { useMailEventStore } from '../stores/mailEvent'
import { storeToRefs } from 'pinia'
import type { JobApplication, Reminder, MailEventPendingItem } from '../types'
import { useRouter } from 'vue-router'

dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

const props = defineProps<{ application?: JobApplication }>()
const emit = defineEmits<{ (e: 'update', data: JobApplication): void }>()

const router = useRouter()
const jobStore = useJobApplicationStore()
const mailEventStore = useMailEventStore()

const { applications } = storeToRefs(jobStore)
const { pendingEvents, loading: mailEventsLoadingRef } = storeToRefs(mailEventStore)

const loading = ref(false)
const activeTab = ref<'reminders' | 'mailEvents'>('reminders')
const filterType = ref('all')
const showReminderModal = ref(false)
const reminders = ref<any[]>([])
const reminderManagerLoadedMailEvents = ref(false)
const selectedApplication = ref<JobApplication | null>(props.application || null)

const mailEvents = computed(() => pendingEvents.value)
const mailEventsLoading = computed(() => mailEventsLoadingRef.value)

const cardLoading = computed(() =>
  activeTab.value === 'reminders' ? loading.value : mailEventsLoading.value
)

const activeBadgeCount = computed(() =>
  activeTab.value === 'reminders' ? activeReminders.value.length : mailEvents.value.length
)

const reminderForm = ref({
  type: 'interview' as 'interview' | 'written' | 'follow_up',
  interview_time: null as any,
  written_time: null as any,
  reminder_time: null as any,
  notification_methods: ['browser', 'sound'],
  message: '',
  repeat: true
})

watch(() => reminderForm.value.type, (type) => {
  if (type === 'interview') {
    reminderForm.value.written_time = null
  }
  if (type === 'written') {
    reminderForm.value.interview_time = null
  }
})

watch(activeTab, async (tab) => {
  if (tab === 'mailEvents' && !reminderManagerLoadedMailEvents.value) {
    await mailEventStore.fetchPending()
    reminderManagerLoadedMailEvents.value = true
  }
})

watch(() => props.application, (app) => {
  selectedApplication.value = app || null
})

let checkInterval: NodeJS.Timeout | null = null

const activeReminders = computed(() =>
  reminders.value.filter(r => !r.is_sent)
)

const filteredReminders = computed(() => {
  let result = activeReminders.value
  switch (filterType.value) {
    case 'interview':
      result = result.filter(r => r.type === 'interview')
      break
    case 'follow_up':
      result = result.filter(r => r.type === 'follow_up')
      break
    case 'written':
      result = result.filter(r => r.type === 'written')
      break
    case 'today':
      result = result.filter(r =>
        dayjs(r.reminder_time).isSame(dayjs(), 'day')
      )
      break
    case 'upcoming':
      result = result.filter(r => {
        const diff = dayjs(r.reminder_time).diff(dayjs(), 'hour')
        return diff >= 0 && diff <= 24
      })
      break
  }
  return result.sort((a, b) =>
    dayjs(a.reminder_time).valueOf() - dayjs(b.reminder_time).valueOf()
  )
})

const formatDateTime = (datetime: string) => dayjs(datetime).format('YYYY-MM-DD HH:mm')

const getTimeRemaining = (reminderTime: string) => {
  const now = dayjs()
  const target = dayjs(reminderTime)
  if (target.isBefore(now)) {
    return '已过期'
  }
  const diffMinutes = target.diff(now, 'minute')
  const diffHours = target.diff(now, 'hour')
  const diffDays = target.diff(now, 'day')
  if (diffMinutes < 60) {
    return `${diffMinutes} 分钟后`
  } else if (diffHours < 24) {
    return `${diffHours} 小时后`
  }
  return `${diffDays} 天后`
}

const getUrgencyColor = (reminderTime: string) => {
  const diffHours = dayjs(reminderTime).diff(dayjs(), 'hour')
  if (diffHours < 0) return 'default'
  if (diffHours <= 1) return 'error'
  if (diffHours <= 6) return 'warning'
  if (diffHours <= 24) return 'processing'
  return 'success'
}

const getReminderTypeText = (type: string) => {
  if (type === 'interview') return '面试提醒'
  if (type === 'written') return '笔试提醒'
  return '跟进提醒'
}

const getReminderColor = (reminder: any) => {
  if (reminder.type === 'interview') return 'blue'
  if (reminder.type === 'written') return 'purple'
  return 'green'
}

const getClassificationText = (classification: string) => {
  switch (classification) {
    case 'exam':
      return '笔试'
    case 'interview':
      return '面试'
    case 'information':
      return '信息通知'
    default:
      return '待分类'
  }
}

const getClassificationColor = (classification: string) => {
  switch (classification) {
    case 'exam':
      return 'purple'
    case 'interview':
      return 'blue'
    case 'information':
      return 'green'
    default:
      return 'default'
  }
}

const getConfidencePercent = (value: number) => {
  if (value <= 1) {
    return Math.round(value * 100)
  }
  return Math.min(100, Math.round(value))
}

const viewApplication = (applicationId: number) => {
  router.push(`/application/${applicationId}`)
}

const dismissReminder = async (reminder: any) => {
  try {
    await jobStore.updateApplication(reminder.application_id, {
      reminder_enabled: false,
      reminder_time: undefined,
      reminder_category: undefined,
      follow_up_date: undefined
    })
    reminders.value = reminders.value.filter(r => r.id !== reminder.id)
    message.success('提醒已归档')
  } catch (error) {
    message.error('操作失败')
  }
}

const openReminderFromEvent = async (event: MailEventPendingItem) => {
  if (!event.application_id) {
    message.warning('事件尚未关联投递记录')
    return
  }
  let target = applications.value.find(app => app.id === event.application_id)
  if (!target) {
    try {
      target = await jobStore.fetchApplicationById(event.application_id)
    } catch (error) {
      message.error('加载关联投递失败')
      return
    }
  }
  if (target) {
    selectedApplication.value = target
    openReminderModal(target)
  }
}

const markEventProcessed = async (event: MailEventPendingItem) => {
  try {
    await mailEventStore.updateEventStatus(event.id, { status: 'processed' })
  } catch (error) {
    // store 内已提示
  }
}

const dismissEvent = async (event: MailEventPendingItem) => {
  try {
    await mailEventStore.updateEventStatus(event.id, { status: 'dismissed' })
  } catch (error) {
    // store 内已提示
  }
}

const refreshMailEvents = async () => {
  await mailEventStore.fetchPending()
}

const saveReminder = async () => {
  if (!reminderForm.value.reminder_time) {
    message.error('请选择提醒时间')
    return
  }
  if (reminderForm.value.type === 'interview' && !reminderForm.value.interview_time) {
    message.error('请选择面试时间')
    return
  }
  if (reminderForm.value.type === 'written' && !reminderForm.value.written_time) {
    message.error('请选择笔试时间')
    return
  }
  try {
    if (selectedApplication.value) {
      const baseTime = reminderForm.value.type === 'written'
        ? reminderForm.value.written_time
        : reminderForm.value.interview_time
      const interviewTypePayload = reminderForm.value.type === 'written'
        ? '笔试'
        : reminderForm.value.type === 'interview'
          ? selectedApplication.value.interview_type || undefined
          : undefined
      const updated = await jobStore.updateApplication(selectedApplication.value.id, {
        interview_time: baseTime ? baseTime.toDate().toISOString() : undefined,
        reminder_time: reminderForm.value.reminder_time
          ? reminderForm.value.reminder_time.toDate().toISOString()
          : undefined,
        reminder_enabled: true,
        interview_type: interviewTypePayload
      })
      emit('update', updated)
    }
    reminderForm.value.notification_methods.includes('browser') && requestNotificationPermission()
    loadReminders()
    message.success('提醒已设置')
    showReminderModal.value = false
    resetReminderForm()
  } catch (error) {
    message.error('设置失败')
  }
}

const cancelReminder = () => {
  showReminderModal.value = false
  resetReminderForm()
}

const resetReminderForm = () => {
  reminderForm.value = {
    type: 'interview',
    interview_time: null,
    written_time: null,
    reminder_time: null,
    notification_methods: ['browser', 'sound'],
    message: '',
    repeat: true
  }
}

const requestNotificationPermission = async () => {
  if ('Notification' in window && Notification.permission !== 'granted') {
    await Notification.requestPermission()
  }
}

const sendBrowserNotification = (reminder: Reminder) => {
  if ('Notification' in window && Notification.permission === 'granted') {
    const notification = new Notification('求职提醒', {
      body: `${reminder.company_name} - ${reminder.position_title}\n${reminder.message || '您有一个待办事项'}`,
      icon: '/favicon.ico',
      requireInteraction: true
    })
    notification.onclick = () => {
      viewApplication(reminder.application_id)
      notification.close()
    }
  }
}

const checkReminders = () => {
  const now = dayjs()
  activeReminders.value.forEach(reminder => {
    const reminderTime = dayjs(reminder.reminder_time)
    if (reminderTime.isBefore(now) || reminderTime.isSame(now, 'minute')) {
      sendBrowserNotification(reminder)
      reminder.is_sent = true
      playReminderSound()
    }
  })
}

const playReminderSound = () => {
  const audio = new Audio('data:audio/wav;base64,UklGRnoGAABXQVZFZm10IBAAAAABAAEAQB8AAEAfAAABAAgAZGF0YQoGAACBhYqFbF1fdJivrJBhNjVgodDbq2EcBj+a2/LDciUFLIHO8tiJNwgZaLvt559NEAxQp+PwtmMcBjiR1/LMeSwFJHfH8N2QQAoUXrTp66hVFApGn+DyvmwhBTGBzvLZiTYIG2m98OScTgwOUant7blmFgU7k9n1unEiBC13yO/eizEIHWq+8+OWT')
  audio.play().catch(() => {})
}

const loadReminders = async () => {
  loading.value = true
  try {
    await jobStore.fetchApplications()
    const generated: any[] = []
    applications.value.forEach((app: JobApplication) => {
      if (app.reminder_enabled && app.reminder_time) {
        const reminderType = app.interview_type === '笔试'
          ? 'written'
          : (app.interview_time ? 'interview' : 'follow_up')
        generated.push({
          id: app.id,
          application_id: app.id,
          type: reminderType,
          reminder_time: app.reminder_time,
          interview_time: app.interview_time || '',
          is_sent: false,
          company_name: app.company_name,
          position_title: app.position_title,
          message: app.notes || ''
        })
      }
    })
    reminders.value = generated
  } catch (error) {
    message.error('加载提醒失败')
  } finally {
    loading.value = false
  }
}

const openReminderModal = (application?: JobApplication) => {
  if (application) {
    selectedApplication.value = application
    if (application.interview_time && application.interview_type !== '笔试') {
      reminderForm.value.interview_time = dayjs(application.interview_time)
      reminderForm.value.written_time = null
      reminderForm.value.type = 'interview'
    } else if (application.interview_time && application.interview_type === '笔试') {
      reminderForm.value.written_time = dayjs(application.interview_time)
      reminderForm.value.interview_time = null
      reminderForm.value.type = 'written'
    }
    if (application.reminder_time) {
      reminderForm.value.reminder_time = dayjs(application.reminder_time)
    }
    if (!application.interview_time && application.follow_up_date) {
      reminderForm.value.type = 'follow_up'
    }
  }
  showReminderModal.value = true
}

defineExpose({
  openReminderModal
})

onMounted(() => {
  requestNotificationPermission()
  loadReminders()
  mailEventStore.fetchPending().finally(() => {
    reminderManagerLoadedMailEvents.value = true
  })
  checkInterval = setInterval(checkReminders, 60000)
})

onUnmounted(() => {
  if (checkInterval) {
    clearInterval(checkInterval)
  }
})
</script>

<style scoped>
.reminder-container {
  height: 100%;
}

.filter-bar {
  margin-bottom: 16px;
}

.reminder-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.reminder-info p {
  margin: 4px 0;
  color: #666;
  display: flex;
  align-items: center;
  gap: 6px;
}

.reminder-info p:last-child {
  margin-top: 8px;
}

.mail-events-header {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 12px;
}

.mail-event-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.mail-event-info {
  display: flex;
  flex-direction: column;
  gap: 6px;
  color: #666;
}

.mail-event-info p {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 6px;
}

.mail-event-snippet {
  background: var(--bg-card, #fafafa);
  padding: 8px;
  border-radius: 6px;
  line-height: 1.5;
}

.mail-event-confidence {
  margin-top: 8px;
}

.mail-event-links {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}
</style>
