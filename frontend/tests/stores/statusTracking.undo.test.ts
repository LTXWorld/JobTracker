import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import dayjs from 'dayjs'
import { useStatusTrackingStore } from '../../src/stores/statusTracking'
import { useJobApplicationStore } from '../../src/stores/jobApplication'
import { StatusTrackingAPI } from '../../src/api/statusTracking'
import { message, notification } from 'ant-design-vue'
import type { ApplicationStatus, StatusHistory } from '../../src/types'

vi.mock('ant-design-vue', () => {
  const messageMock = {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn()
  }
  const notificationMock = {
    open: vi.fn(),
    close: vi.fn()
  }
  return {
    message: messageMock,
    notification: notificationMock,
    Button: {
      name: 'MockButton',
      render: () => null
    }
  }
})

describe('StatusTrackingStore 撤销流程', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers()
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('状态更新后触发撤销通知并完成撤销流程', async () => {
    const store = useStatusTrackingStore()
    const jobStore = useJobApplicationStore()

    const status: ApplicationStatus = '简历筛选中'
    const historyResponse: StatusHistory = {
      history: [],
      metadata: {
        total_duration: 0,
        status_count: 0,
        last_updated: '',
        current_stage: ''
      }
    }

    const updateResult = {
      job: {
        id: 1,
        company_name: '测试公司',
        position_title: '前端工程师',
        application_date: '2025-01-01',
        status,
        status_version: 4,
        created_at: '2025-01-01T00:00:00Z',
        updated_at: '2025-01-01T00:00:00Z'
      },
      history_id: 99,
      status_version: 4,
      undo_available_until: dayjs().add(10, 'second').toISOString(),
      metadata: {}
    }

    const undoResult = {
      job: {
        id: 1,
        company_name: '测试公司',
        position_title: '前端工程师',
        application_date: '2025-01-01',
        status: '已投递' as ApplicationStatus,
        status_version: 3,
        created_at: '2025-01-01T00:00:00Z',
        updated_at: '2025-01-01T00:00:00Z'
      },
      reverted_to: '已投递' as ApplicationStatus,
      undo_history_id: 100,
      source_history_id: 99,
      status_version: 3,
      undo_completed_at: new Date().toISOString()
    }

    const onUndoSpy = vi.fn()

    vi.spyOn(StatusTrackingAPI, 'updateStatus').mockResolvedValue(updateResult as any)
    vi.spyOn(StatusTrackingAPI, 'getStatusHistory').mockResolvedValue(historyResponse)
    vi.spyOn(StatusTrackingAPI, 'undoStatus').mockResolvedValue(undoResult as any)
    vi.spyOn(jobStore, 'fetchApplications').mockResolvedValue()

    const result = await store.updateApplicationStatus(1, { status }, { onUndo: onUndoSpy })

    expect(result?.history_id).toBe(99)
    expect(StatusTrackingAPI.updateStatus).toHaveBeenCalledTimes(1)
    expect(notification.open).toHaveBeenCalled()

    await store.undoStatusUpdate(1)

    expect(StatusTrackingAPI.undoStatus).toHaveBeenCalledWith(1, {
      history_id: 99,
      version: 4
    })
    expect(jobStore.fetchApplications).toHaveBeenCalled()
    expect(message.success).toHaveBeenCalledWith(expect.stringContaining('已撤销'))
    expect(notification.close).toHaveBeenCalled()
    expect(onUndoSpy).toHaveBeenCalled()

    vi.runOnlyPendingTimers()
  })
})
