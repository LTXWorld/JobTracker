import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useMailEventStore } from '../../src/stores/mailEvent'
import { MailEventAPI } from '../../src/api/mailEvent'
import type { MailEventPendingItem } from '../../src/types'

vi.mock('../../src/api/mailEvent', () => ({
  MailEventAPI: {
    getPending: vi.fn(),
    updateStatus: vi.fn()
  }
}))

vi.mock('ant-design-vue', () => ({
  message: {
    success: vi.fn(),
    error: vi.fn()
  }
}))

describe('MailEventStore', () => {
  const sampleEvents: MailEventPendingItem[] = [
    {
      id: 1,
      user_id: 1,
      mailbox_id: 1,
      application_id: 10,
      subject: '笔试通知',
      sender: 'hr@example.com',
      received_at: '2024-01-01T10:00:00Z',
      classification: 'exam',
      confidence: 0.8,
      payload: { exam_link: 'https://exam.example.com' },
      status: 'needs_review',
      application: null
    },
    {
      id: 2,
      user_id: 1,
      mailbox_id: 1,
      application_id: null,
      subject: '面试邀请',
      sender: 'teamlead@example.com',
      received_at: '2024-01-02T09:00:00Z',
      classification: 'interview',
      confidence: 0.7,
      payload: {},
      status: 'pending',
      application: null
    }
  ] as any

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('fetchPending 应该载入事件列表', async () => {
    vi.mocked(MailEventAPI.getPending).mockResolvedValueOnce(sampleEvents)
    const store = useMailEventStore()

    await store.fetchPending()

    expect(MailEventAPI.getPending).toHaveBeenCalled()
    expect(store.pendingEvents).toEqual(sampleEvents)
  })

  it('fetchPending 失败时不应抛出异常', async () => {
    vi.mocked(MailEventAPI.getPending).mockRejectedValueOnce(new Error('network'))
    const store = useMailEventStore()

    await expect(store.fetchPending()).resolves.not.toThrow()
    expect(store.pendingEvents).toEqual([])
  })

  it('updateEventStatus 处理完成后应移除事件', async () => {
    const store = useMailEventStore()
    store.pendingEvents = [...sampleEvents]

    vi.mocked(MailEventAPI.updateStatus).mockResolvedValueOnce({
      ...sampleEvents[0],
      status: 'processed'
    } as MailEventPendingItem)

    await store.updateEventStatus(1, { status: 'processed' })

    expect(MailEventAPI.updateStatus).toHaveBeenCalledWith(1, { status: 'processed' })
    expect(store.pendingEvents.find(event => event.id === 1)).toBeUndefined()
  })

  it('updateEventStatus 仍为待确认状态时应更新本地数据', async () => {
    const store = useMailEventStore()
    store.pendingEvents = [...sampleEvents]

    vi.mocked(MailEventAPI.updateStatus).mockResolvedValueOnce({
      ...sampleEvents[1],
      subject: '面试邀请提醒'
    } as MailEventPendingItem)

    await store.updateEventStatus(2, { status: 'pending' })

    expect(store.pendingEvents.find(event => event.id === 2)?.subject).toBe('面试邀请提醒')
  })
})
