import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('@ant-design/icons-vue', () => {
  const stub = { template: '<span />' }
  return {
    HistoryOutlined: stub,
    ReloadOutlined: stub,
    ClockCircleOutlined: stub,
    CalendarOutlined: stub,
    FileTextOutlined: stub,
    SendOutlined: stub,
    EyeOutlined: stub,
    CheckCircleOutlined: stub,
    CloseCircleOutlined: stub,
    UserOutlined: stub,
    TeamOutlined: stub,
    CrownOutlined: stub,
    ContactsOutlined: stub,
    StopOutlined: stub,
    QuestionCircleOutlined: stub,
    InfoCircleOutlined: stub,
    SmileOutlined: stub,
    MehOutlined: stub,
    FrownOutlined: stub
  }
})

import StatusTimeline from '../../src/components/StatusTimeline.vue'
import { useStatusTrackingStore } from '../../src/stores/statusTracking'
import { ApplicationStatus as ApplicationStatusEnum } from '../../src/types'
import type { StatusHistory } from '../../src/types'

const blockStub = { template: '<div><slot /></div>' }
const inlineStub = { template: '<span><slot /></span>' }

describe('StatusTimeline 面试体验展示', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('展示评分与跳过两类面试体验信息', async () => {
    const store = useStatusTrackingStore()
    const history: StatusHistory = {
      history: [
        {
          status: ApplicationStatusEnum.FIRST_INTERVIEW,
          timestamp: '2025-02-01T09:00:00Z',
          metadata: {
            interview_experience: {
              rating: 'good',
              skip: false,
              note: '面试官沟通顺畅',
              recorded_at: '2025-02-01T10:00:00Z'
            }
          }
        },
        {
          status: ApplicationStatusEnum.SECOND_INTERVIEW,
          timestamp: '2025-02-05T09:30:00Z',
          metadata: {
            interview_experience: {
              skip: true,
              skip_reason: '候选人时间冲突',
              note: '等待重新安排'
            }
          }
        }
      ],
      metadata: {
        total_duration: 0,
        status_count: 2,
        last_updated: '2025-02-05T09:30:00Z',
        current_stage: '二面中'
      }
    }

    vi.spyOn(store, 'fetchStatusHistory').mockResolvedValue(history)

    const wrapper = mount(StatusTimeline, {
      props: {
        applicationId: 101
      },
      global: {
        stubs: {
          'a-button': { template: '<button @click="$emit(\'click\')"><slot /></button>' },
          'a-spin': blockStub,
          'a-empty': blockStub,
          'a-timeline': blockStub,
          'a-timeline-item': blockStub,
          'a-tag': inlineStub,
          'a-progress': inlineStub,
          'a-statistic': blockStub,
          'a-row': blockStub,
          'a-col': blockStub
        }
      }
    })

    await flushPromises()
    await wrapper.vm.$nextTick()

    expect(store.fetchStatusHistory).toHaveBeenCalledWith(101, false)
    const textContent = wrapper.text()
    expect(textContent).toContain('体验良好')
    expect(textContent).toContain('面试官沟通顺畅')
    expect(textContent).toContain('记录于 2025-02-01')
    expect(textContent).toContain('已跳过反馈')
    expect(textContent).toContain('原因：候选人时间冲突')
    expect(textContent).toContain('备注：等待重新安排')
  })
})
