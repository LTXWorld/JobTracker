import { defineStore } from 'pinia'
import { ref } from 'vue'
import { MailEventAPI } from '../api/mailEvent'
import type { MailEventPendingItem, MailEventStatusUpdateRequest } from '../types'
import { message } from 'ant-design-vue'

export const useMailEventStore = defineStore('mailEvent', () => {
  const pendingEvents = ref<MailEventPendingItem[]>([])
  const loading = ref(false)

  const fetchPending = async () => {
    loading.value = true
    try {
      const events = await MailEventAPI.getPending()
      pendingEvents.value = events
    } catch (error) {
      message.error('获取待确认事件失败')
    } finally {
      loading.value = false
    }
  }

  const updateEventStatus = async (id: number, payload: MailEventStatusUpdateRequest) => {
    try {
      const updated = await MailEventAPI.updateStatus(id, payload)
      if (updated.status === 'pending' || updated.status === 'needs_review') {
        pendingEvents.value = pendingEvents.value.map(event =>
          event.id === id ? updated : event
        )
      } else {
        // 处理完成或忽略的事件不再保留
        pendingEvents.value = pendingEvents.value.filter(event => event.id !== id)
      }
      message.success('事件状态已更新')
      return updated
    } catch (error) {
      message.error('事件状态更新失败')
      throw error
    }
  }

  const removeEvent = (id: number) => {
    pendingEvents.value = pendingEvents.value.filter(event => event.id !== id)
  }

  return {
    pendingEvents,
    loading,
    fetchPending,
    updateEventStatus,
    removeEvent
  }
})
