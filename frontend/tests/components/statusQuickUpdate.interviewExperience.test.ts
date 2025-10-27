import { describe, it, expect, beforeEach, vi } from 'vitest'
import { shallowMount, flushPromises } from '@vue/test-utils'
import StatusQuickUpdate from '../../src/components/StatusQuickUpdate.vue'
import type { ApplicationStatus, UpdateStatusRequest } from '../../src/types'

vi.mock('ant-design-vue', () => {
  const messageMock = {
    warning: vi.fn(),
    info: vi.fn(),
    success: vi.fn(),
    error: vi.fn()
  }

  return {
    message: messageMock,
    Modal: {
      confirm: vi.fn()
    }
  }
})

const requestInterviewExperienceMock = vi.fn()
const shouldCaptureInterviewExperienceMock = vi.fn()

vi.mock('../../src/composables/useInterviewExperienceCapture', () => ({
  useInterviewExperienceCapture: () => ({
    requestInterviewExperience: requestInterviewExperienceMock,
    shouldCaptureInterviewExperience: shouldCaptureInterviewExperienceMock
  })
}))

const updateApplicationStatusMock = vi.fn()
const assembleUpdateRequestMock = vi.fn((base: UpdateStatusRequest, _ctx: any, submission: any) => ({
  ...base,
  metadata: {
    ...(base.metadata || {}),
    interview_experience: submission
  }
}))
const getAvailableTransitionsMock = vi.fn().mockResolvedValue([])

vi.mock('../../src/stores/statusTracking', () => ({
  useStatusTrackingStore: () => ({
    updateApplicationStatus: updateApplicationStatusMock,
    assembleUpdateRequestWithInterviewExperience: assembleUpdateRequestMock,
    getAvailableTransitions: getAvailableTransitionsMock
  })
}))

const simpleStub = { template: '<div><slot /></div>' }
const buttonStub = { template: '<button @click="$emit(\'click\')"><slot /></button>' }
const modalStub = { template: '<div><slot /><slot name="footer" /></div>' }

const baseGlobal = {
  config: {
    compilerOptions: {
      isCustomElement: (tag: string) => tag.startsWith('a-')
    }
  },
  stubs: {
    StatusUpdateContent: { template: '<div />' },
    StatusTimeline: { template: '<div />' },
    'a-card': simpleStub,
    'a-space': simpleStub,
    'a-modal': modalStub,
    'a-button': buttonStub
  }
}

const mountComponent = () => {
  return shallowMount(StatusQuickUpdate, {
    props: {
      applicationId: 11,
      currentStatus: '一面中' as ApplicationStatus,
      mode: 'inline',
      compact: true
    },
    global: baseGlobal
  })
}

describe('StatusQuickUpdate 面试体验写入', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    shouldCaptureInterviewExperienceMock.mockReturnValue(true)
    getAvailableTransitionsMock.mockResolvedValue([])
  })

  it('评分路径会注入 metadata', async () => {
    requestInterviewExperienceMock.mockResolvedValue({
      cancelled: false,
      submission: {
        skip: false,
        rating: 'good',
        note: '体验优秀'
      }
    })

    const wrapper = mountComponent()
    await flushPromises()

    const vm = wrapper.vm as any
    vm.selectedStatus = '二面中'
    vm.note = '  整体沟通顺畅 '

    updateApplicationStatusMock.mockResolvedValue(undefined)

    await vm.handleQuickUpdate()

    expect(requestInterviewExperienceMock).toHaveBeenCalledWith({
      fromStatus: '一面中',
      toStatus: '二面中'
    })

    expect(assembleUpdateRequestMock).toHaveBeenCalled()
    const assembledPayload = assembleUpdateRequestMock.mock.results[0].value
    expect(assembledPayload.metadata.interview_experience).toEqual({
      skip: false,
      rating: 'good',
      note: '体验优秀'
    })

    expect(updateApplicationStatusMock).toHaveBeenCalledWith(
      11,
      assembledPayload,
      expect.any(Object)
    )
  })

  it('跳过路径会写入 skip metadata', async () => {
    requestInterviewExperienceMock.mockResolvedValue({
      cancelled: false,
      submission: {
        skip: true,
        skip_reason: '暂无时间记录',
        note: '待补充'
      }
    })

    const wrapper = mountComponent()
    await flushPromises()

    const vm = wrapper.vm as any
    vm.selectedStatus = '一面通过'
    vm.note = ''

    updateApplicationStatusMock.mockResolvedValue(undefined)

    await vm.handleQuickUpdate()

    const assembledPayload = assembleUpdateRequestMock.mock.results[0].value
    expect(assembledPayload.metadata.interview_experience).toEqual({
      skip: true,
      skip_reason: '暂无时间记录',
      note: '待补充'
    })
    expect(updateApplicationStatusMock).toHaveBeenCalled()
  })

  it('取消反馈时不触发状态更新', async () => {
    requestInterviewExperienceMock.mockResolvedValue({
      cancelled: true
    })

    const wrapper = mountComponent()
    await flushPromises()

    const vm = wrapper.vm as any
    vm.selectedStatus = '二面中'

    await vm.handleQuickUpdate()

    expect(updateApplicationStatusMock).not.toHaveBeenCalled()
    expect(assembleUpdateRequestMock).not.toHaveBeenCalled()
    expect(vm.loading).toBe(false)
  })
})
