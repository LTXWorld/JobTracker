import request from './request'
import type { MailboxResponse, MailboxBindRequest } from '../types'

export const MailboxAPI = {
  async get(): Promise<MailboxResponse | null> {
    const response = await request.get('/v1/mailbox')
    return response.data.data || null
  },

  async bind(payload: MailboxBindRequest): Promise<MailboxResponse> {
    const response = await request.post('/v1/mailbox', payload)
    if (!response.data.data) {
      throw new Error('邮箱绑定失败')
    }
    return response.data.data
  },

  async remove(): Promise<void> {
    await request.delete('/v1/mailbox')
  }
}
