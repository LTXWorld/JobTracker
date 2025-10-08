import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useMailboxStore } from '../../src/stores/mailbox'
import { MailboxAPI } from '../../src/api/mailbox'
import type { MailboxResponse, MailboxBindRequest } from '../../src/types'

vi.mock('../../src/api/mailbox', () => ({
  MailboxAPI: {
    get: vi.fn(),
    bind: vi.fn(),
    remove: vi.fn()
  }
}))

vi.mock('ant-design-vue', () => ({
  message: {
    success: vi.fn(),
    error: vi.fn()
  }
}))

describe('MailboxStore', () => {
  const mockMailbox: MailboxResponse = {
    email_address: 'user@qq.com',
    provider: 'qq',
    protocol: 'imap',
    host: 'imap.qq.com',
    port: 993,
    use_ssl: true,
    status: 'active',
    requires_attention: false
  }

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('fetchMailbox 成功更新状态', async () => {
    vi.mocked(MailboxAPI.get).mockResolvedValueOnce(mockMailbox)
    const store = useMailboxStore()

    await store.fetchMailbox()

    expect(MailboxAPI.get).toHaveBeenCalled()
    expect(store.mailbox).toEqual(mockMailbox)
  })

  it('fetchMailbox 失败时保持状态不变', async () => {
    vi.mocked(MailboxAPI.get).mockRejectedValueOnce(new Error('network'))
    const store = useMailboxStore()

    await expect(store.fetchMailbox()).resolves.toBeUndefined()
    expect(store.mailbox).toBeNull()
  })

  it('bindMailbox 保存后更新 mailbox', async () => {
    vi.mocked(MailboxAPI.bind).mockResolvedValueOnce(mockMailbox)
    const store = useMailboxStore()
    const payload: MailboxBindRequest = {
      email_address: 'user@qq.com',
      provider: 'qq',
      protocol: 'imap',
      host: 'imap.qq.com',
      port: 993,
      use_ssl: true,
      authorization_code: 'abc'
    }

    const result = await store.bindMailbox(payload)

    expect(MailboxAPI.bind).toHaveBeenCalledWith(payload)
    expect(result).toEqual(mockMailbox)
    expect(store.mailbox).toEqual(mockMailbox)
  })

  it('removeMailbox 清理状态', async () => {
    const store = useMailboxStore()
    store.mailbox = mockMailbox
    vi.mocked(MailboxAPI.remove).mockResolvedValueOnce()

    await store.removeMailbox()

    expect(MailboxAPI.remove).toHaveBeenCalled()
    expect(store.mailbox).toBeNull()
  })
})
