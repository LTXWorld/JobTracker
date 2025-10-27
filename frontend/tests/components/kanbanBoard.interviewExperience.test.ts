import { describe, it, expect, beforeEach, vi } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import type { ApplicationStatus } from '../../src/types'

var messageMock: any
var modalConfirmMock: any

vi.mock('ant-design-vue', () => {
  messageMock = {
    warning: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
    success: vi.fn()
  }
  modalConfirmMock = vi.fn()
  return {
    message: messageMock,
    Modal: {
      confirm: modalConfirmMock
    }
  }
})

vi.mock('vuedraggable', () => ({
  default: {
    name: 'DraggableStub',
    template: '<div><slot /></div>'
  }
}))

vi.mock('pinia', () => ({
  storeToRefs: (store: any) => ({
    applications: store.applications,
    loading: store.loading
  })
}))

const requestInterviewExperienceMock = vi.fn()
const shouldCaptureInterviewExperienceMock = vi.fn()

vi.mock('../../src/composables/useInterviewExperienceCapture', () => ({
  useInterviewExperienceCapture: () => ({
    requestInterviewExperience: requestInterviewExperienceMock,
    shouldCaptureInterviewExperience: shouldCaptureInterviewExperienceMock
  })
}))

const getAvailableTransitionsMock = vi.fn()
const updateApplicationStatusMock = vi.fn()
const assembleUpdateRequestMock = vi.fn((base: any, _ctx: any, submission: any) => ({
  ...base,
  metadata: {
    ...(base.metadata || {}),
    interview_experience: submission
  }
}))

vi.mock('../../src/stores/statusTracking', () => ({
  useStatusTrackingStore: () => ({
    getAvailableTransitions: getAvailableTransitionsMock,
    updateApplicationStatus: updateApplicationStatusMock,
    assembleUpdateRequestWithInterviewExperience: assembleUpdateRequestMock
  })
}))

const fetchApplicationsMock = vi.fn()
const applicationsRef = { value: [] as any[] }
const loadingRef = { value: false }

vi.mock('../../src/stores/jobApplication', () => ({
  useJobApplicationStore: () => ({
    fetchApplications: fetchApplicationsMock,
    deleteApplication: vi.fn(),
    applications: applicationsRef,
    loading: loadingRef
  })
}))

import KanbanBoard from '../../src/views/KanbanBoard.vue'

const mountKanban = () => {
  return shallowMount(KanbanBoard, {
    global: {
      config: {
        compilerOptions: {
          isCustomElement: (tag: string) => tag.startsWith('a-')
        }
      },
      stubs: {
        draggable: { template: '<div><slot /></div>' },
        'a-button': { template: '<button><slot /></button>' },
        'a-card': { template: '<div><slot /></div>' },
        'a-space': { template: '<div><slot /></div>' },
        'a-modal': { template: '<div><slot /><slot name="footer" /></div>' }
      }
    }
  })
}

const buildEvent = (status: ApplicationStatus) => ({
  added: {
    element: {
      id: 28,
      company_name: '测试公司',
      position_title: '高级前端',
      status
    }
  }
})

describe('KanbanBoard 面试体验采集流程', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    applicationsRef.value = []
    fetchApplicationsMock.mockResolvedValue(undefined)
    getAvailableTransitionsMock.mockResolvedValue([])
    shouldCaptureInterviewExperienceMock.mockReturnValue(true)
  })

  it('拖拽面试状态时写入评分 metadata', async () => {
    requestInterviewExperienceMock.mockResolvedValue({
      cancelled: false,
      submission: {
        skip: false,
        rating: 'good',
        note: '沟通顺畅'
      }
    })

    updateApplicationStatusMock.mockResolvedValue(undefined)

    const wrapper = mountKanban()
    await flushPromises()
    fetchApplicationsMock.mockClear()

    const evt = buildEvent('一面中')
    await (wrapper.vm as any).handleDragChange(evt, '二面中')

    expect(requestInterviewExperienceMock).toHaveBeenCalledWith({
      fromStatus: '一面中',
      toStatus: '二面中'
    })
    const enrichedPayload = assembleUpdateRequestMock.mock.results[0].value
    expect(enrichedPayload.metadata.interview_experience).toEqual({
      skip: false,
      rating: 'good',
      note: '沟通顺畅'
    })
    expect(updateApplicationStatusMock).toHaveBeenCalledWith(
      28,
      enrichedPayload,
      expect.any(Object)
    )
    expect(fetchApplicationsMock).toHaveBeenCalledTimes(1)
  })

  it('用户取消反馈后不会调用状态更新接口', async () => {
    requestInterviewExperienceMock.mockResolvedValue({
      cancelled: true
    })

    const wrapper = mountKanban()
    await flushPromises()
    fetchApplicationsMock.mockClear()

    const evt = buildEvent('一面中')
    await (wrapper.vm as any).handleDragChange(evt, '二面中')

    expect(updateApplicationStatusMock).not.toHaveBeenCalled()
    expect(assembleUpdateRequestMock).not.toHaveBeenCalled()
    expect(fetchApplicationsMock).toHaveBeenCalledTimes(1)
    expect(messageMock.info).toHaveBeenCalledWith(expect.stringContaining('已取消'))
  })
})
