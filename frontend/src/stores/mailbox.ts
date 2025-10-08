import { defineStore } from 'pinia'
import { ref } from 'vue'
import { MailboxAPI } from '../api/mailbox'
import type { MailboxResponse, MailboxBindRequest } from '../types'
import { message } from 'ant-design-vue'

export const useMailboxStore = defineStore('mailbox', () => {
  const mailbox = ref<MailboxResponse | null>(null)
  const loading = ref(false)
  const saving = ref(false)

  const fetchMailbox = async () => {
    loading.value = true
    try {
      mailbox.value = await MailboxAPI.get()
    } catch (error) {
      message.error('获取邮箱配置失败')
    } finally {
      loading.value = false
    }
  }

  const bindMailbox = async (payload: MailboxBindRequest) => {
    saving.value = true
    try {
      const result = await MailboxAPI.bind(payload)
      mailbox.value = result
      message.success('邮箱授权已更新')
      return result
    } catch (error) {
      message.error('邮箱授权失败')
      throw error
    } finally {
      saving.value = false
    }
  }

  const removeMailbox = async () => {
    saving.value = true
    try {
      await MailboxAPI.remove()
      mailbox.value = null
      message.success('邮箱授权已解除')
    } catch (error) {
      message.error('解除邮箱授权失败')
      throw error
    } finally {
      saving.value = false
    }
  }

  return {
    mailbox,
    loading,
    saving,
    fetchMailbox,
    bindMailbox,
    removeMailbox
  }
})
