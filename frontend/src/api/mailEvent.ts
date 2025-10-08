import request from './request'
import type { MailEventPendingItem, MailEventStatusUpdateRequest } from '../types'

export const MailEventAPI = {
  async getPending(): Promise<MailEventPendingItem[]> {
    const response = await request.get('/v1/mail-events/pending')
    return Array.isArray(response.data.data) ? response.data.data : []
  },
  async updateStatus(id: number, data: MailEventStatusUpdateRequest): Promise<MailEventPendingItem> {
    const response = await request.patch(`/v1/mail-events/${id}/status`, data)
    if (!response.data.data) {
      throw new Error('事件状态更新失败')
    }
    return response.data.data
  }
}
