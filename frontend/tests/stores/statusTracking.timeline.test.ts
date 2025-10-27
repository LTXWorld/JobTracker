import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useStatusTrackingStore } from '../../src/stores/statusTracking'
import { ApplicationStatus as ApplicationStatusEnum } from '../../src/types'
import type { StatusHistory } from '../../src/types'

describe('StatusTrackingStore 时间轴面试体验转换', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('保留评分类面试体验字段并进行必要归一化', () => {
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
              note: ' 面试官沟通顺畅 ',
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
              skip_reason: '候选人时间冲突 ',
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

    const timeline = store.convertToTimelineData(history)

    expect(timeline).toHaveLength(2)
    const [first, second] = timeline

    expect(first.interviewExperience).toMatchObject({
      skip: false,
      rating: 'good',
      note: '面试官沟通顺畅',
      recordedAt: '2025-02-01T10:00:00Z'
    })
    expect(first.interviewExperience?.raw).toMatchObject({
      rating: 'good',
      skip: false
    })

    expect(second.interviewExperience).toMatchObject({
      skip: true,
      rating: undefined,
      skipReason: '候选人时间冲突',
      note: '等待重新安排'
    })
    expect(second.interviewExperience?.raw).toMatchObject({
      skip: true,
      skip_reason: '候选人时间冲突 '
    })
  })
})
