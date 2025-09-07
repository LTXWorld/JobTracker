<template>
  <div class="reminder-container">
    <!-- 提醒列表卡片 -->
    <a-card title="待办提醒" :loading="loading">
      <template #extra>
        <a-badge :count="activeReminders.length" :showZero="true">
          <BellOutlined style="font-size: 20px" />
        </a-badge>
      </template>

      <!-- 快速筛选 -->
      <div class="filter-bar">
        <a-radio-group v-model:value="filterType" button-style="solid">
          <a-radio-button value="all">全部</a-radio-button>
          <a-radio-button value="interview">面试提醒</a-radio-button>
          <a-radio-button value="follow_up">跟进提醒</a-radio-button>
          <a-radio-button value="today">今日待办</a-radio-button>
          <a-radio-button value="upcoming">即将到来</a-radio-button>
        </a-radio-group>
      </div>

      <a-divider />

      <!-- 提醒列表 -->
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

      <!-- 空状态 -->
      <a-empty 
        v-else
        description="暂无待办提醒"
        :image="Empty.PRESENTED_IMAGE_SIMPLE"
      />
    </a-card>

    <!-- 设置提醒弹窗 -->
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
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { 
  BellOutlined, CalendarOutlined, ClockCircleOutlined,
  MessageOutlined
} from '@ant-design/icons-vue'
import { message, Empty } from 'ant-design-vue'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'
import { useJobApplicationStore } from '../stores/jobApplication'
import type { JobApplication, Reminder } from '../types'
import { useRouter } from 'vue-router'

dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

const props = defineProps<{
  application?: JobApplication
}>()

const emit = defineEmits<{
  'update': (data: any) => void
}>()

const router = useRouter()
const jobStore = useJobApplicationStore()

// 响应式数据
const loading = ref(false)
const filterType = ref('all')
const showReminderModal = ref(false)
const reminders = ref<any[]>([])
const reminderForm = ref({
  type: 'interview' as 'interview' | 'follow_up',
  interview_time: null as any,
  reminder_time: null as any,
  notification_methods: ['browser', 'sound'],
  message: '',
  repeat: true
})

// 定时器
let checkInterval: NodeJS.Timeout | null = null

// 活动提醒（未发送的）
const activeReminders = computed(() => 
  reminders.value.filter(r => !r.is_sent)
)

// 筛选后的提醒
const filteredReminders = computed(() => {
  let result = activeReminders.value

  switch (filterType.value) {
    case 'interview':
      result = result.filter(r => r.type === 'interview')
      break
    case 'follow_up':
      result = result.filter(r => r.type === 'follow_up')
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

  // 按时间排序
  return result.sort((a, b) => 
    dayjs(a.reminder_time).valueOf() - dayjs(b.reminder_time).valueOf()
  )
})

// 格式化日期时间
const formatDateTime = (datetime: string) => {
  return dayjs(datetime).format('YYYY-MM-DD HH:mm')
}

// 获取剩余时间
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
  } else {
    return `${diffDays} 天后`
  }
}

// 获取紧急程度颜色
const getUrgencyColor = (reminderTime: string) => {
  const diffHours = dayjs(reminderTime).diff(dayjs(), 'hour')
  
  if (diffHours < 0) return 'default'
  if (diffHours <= 1) return 'error'
  if (diffHours <= 6) return 'warning'
  if (diffHours <= 24) return 'processing'
  return 'success'
}

// 获取提醒类型文本
const getReminderTypeText = (type: string) => {
  return type === 'interview' ? '面试提醒' : '跟进提醒'
}

// 获取提醒颜色
const getReminderColor = (reminder: any) => {
  return reminder.type === 'interview' ? 'blue' : 'green'
}

// 查看申请详情
const viewApplication = (applicationId: number) => {
  router.push(`/application/${applicationId}`)
}

// 忽略提醒
const dismissReminder = async (reminder: any) => {
  try {
    // 标记为已发送
    reminder.is_sent = true
    message.success('已忽略该提醒')
    
    // 从列表中移除
    reminders.value = reminders.value.filter(r => r.id !== reminder.id)
  } catch (error) {
    message.error('操作失败')
  }
}

// 保存提醒
const saveReminder = async () => {
  if (!reminderForm.value.reminder_time) {
    message.error('请选择提醒时间')
    return
  }

  if (reminderForm.value.type === 'interview' && !reminderForm.value.interview_time) {
    message.error('请选择面试时间')
    return
  }

  try {
    // 更新应用数据
    if (props.application) {
      await jobStore.updateApplication(props.application.id, {
        interview_time: reminderForm.value.interview_time?.format('YYYY-MM-DD HH:mm:ss'),
        reminder_time: reminderForm.value.reminder_time?.format('YYYY-MM-DD HH:mm:ss'),
        reminder_enabled: true
      })
    }

    message.success('提醒设置成功')
    showReminderModal.value = false
    
    // 请求浏览器通知权限
    if (reminderForm.value.notification_methods.includes('browser')) {
      requestNotificationPermission()
    }
    
    // 刷新提醒列表
    loadReminders()
  } catch (error) {
    message.error('设置失败')
  }
}

// 取消设置提醒
const cancelReminder = () => {
  showReminderModal.value = false
  resetReminderForm()
}

// 重置表单
const resetReminderForm = () => {
  reminderForm.value = {
    type: 'interview',
    interview_time: null,
    reminder_time: null,
    notification_methods: ['browser', 'sound'],
    message: '',
    repeat: true
  }
}

// 请求通知权限
const requestNotificationPermission = async () => {
  if ('Notification' in window && Notification.permission !== 'granted') {
    await Notification.requestPermission()
  }
}

// 发送浏览器通知
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

// 检查并触发提醒
const checkReminders = () => {
  const now = dayjs()
  
  activeReminders.value.forEach(reminder => {
    const reminderTime = dayjs(reminder.reminder_time)
    
    // 如果提醒时间已到
    if (reminderTime.isBefore(now) || reminderTime.isSame(now, 'minute')) {
      // 发送通知
      sendBrowserNotification(reminder)
      
      // 标记为已发送
      reminder.is_sent = true
      
      // 播放声音
      playReminderSound()
    }
  })
}

// 播放提醒声音
const playReminderSound = () => {
  const audio = new Audio('data:audio/wav;base64,UklGRnoGAABXQVZFZm10IBAAAAABAAEAQB8AAEAfAAABAAgAZGF0YQoGAACBhYqFbF1fdJivrJBhNjVgodDbq2EcBj+a2/LDciUFLIHO8tiJNwgZaLvt559NEAxQp+PwtmMcBjiR1/LMeSwFJHfH8N2QQAoUXrTp66hVFApGn+DyvmwhBTGBzvLZiTYIG2m98OScTgwOUant7blmFgU7k9n1unEiBC13yO/eizEIHWq+8+OWT')
  audio.play().catch(() => {})
}

// 加载提醒列表
const loadReminders = async () => {
  loading.value = true
  
  try {
    // 先获取应用列表
    await jobStore.fetchApplications()
    const generatedReminders: any[] = []
    
    // 使用 store 中的 applications
    jobStore.applications.forEach((app: JobApplication) => {
      if (app.reminder_enabled && app.reminder_time) {
        generatedReminders.push({
          id: app.id,
          application_id: app.id,
          type: app.interview_time ? 'interview' : 'follow_up',
          reminder_time: app.reminder_time,
          interview_time: app.interview_time || '',
          is_sent: false,
          company_name: app.company_name,
          position_title: app.position_title,
          message: app.notes || ''
        })
      }
    })
    
    reminders.value = generatedReminders
  } catch (error) {
    message.error('加载提醒失败')
  } finally {
    loading.value = false
  }
}

// 打开设置提醒弹窗
const openReminderModal = (application?: JobApplication) => {
  if (application) {
    // 预填充数据
    if (application.interview_time) {
      reminderForm.value.interview_time = dayjs(application.interview_time)
    }
    if (application.reminder_time) {
      reminderForm.value.reminder_time = dayjs(application.reminder_time)
    }
  }
  showReminderModal.value = true
}

// 暴露方法给父组件
defineExpose({
  openReminderModal
})

onMounted(() => {
  // 请求通知权限
  requestNotificationPermission()
  
  // 加载提醒
  loadReminders()
  
  // 启动定时检查
  checkInterval = setInterval(checkReminders, 60000) // 每分钟检查一次
})

onUnmounted(() => {
  // 清理定时器
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
</style>