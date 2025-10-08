import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('@ant-design/icons-vue', () => {
  const stub = { template: '<span />' }
  return {
    BellOutlined: stub,
    CalendarOutlined: stub,
    ClockCircleOutlined: stub,
    MessageOutlined: stub,
    MailOutlined: stub,
    LinkOutlined: stub,
    FieldTimeOutlined: stub,
    PlusOutlined: stub,
    TeamOutlined: stub,
    PhoneOutlined: stub
  }
})

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() })
}))

import ReminderManager from '../../src/components/ReminderManager.vue'
import { useMailEventStore } from '../../src/stores/mailEvent'
import { useJobApplicationStore } from '../../src/stores/jobApplication'
import type { MailEventPendingItem, JobApplication } from '../../src/types'

const simpleStub = { template: '<div><slot /></div>' }
const buttonStub = { template: '<button @click="$emit(\'click\')"><slot /></button>' }
const listStub = {
  props: ['dataSource'],
  template: '<div><div v-for="item in dataSource" class="list-item"><slot name="renderItem" :item="item" /></div></div>'
}
const listItemStub = { template: '<div class="list-item"><slot /></div>' }
const listItemMetaStub = { template: '<div class="meta"><slot name="title" /><slot name="description" /><slot name="avatar" /></div>' }

describe('ReminderManager mail-events integration', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('在邮件事件标签页展示待确认事件并触发状态更新', async () => {
    const mailEventStore = useMailEventStore()
    const jobStore = useJobApplicationStore()

    vi.spyOn(mailEventStore, 'fetchPending').mockResolvedValue()
    vi.spyOn(jobStore, 'fetchApplications').mockResolvedValue()

    const application: JobApplication = {
      id: 101,
      company_name: '示例公司',
      position_title: '后端实习生',
      application_date: '2024-01-01',
      status: '已投递',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z'
    } as JobApplication
    jobStore.applications = [application]

    const mailEvent: MailEventPendingItem = {
      id: 1,
      user_id: 1,
      mailbox_id: 10,
      application_id: 101,
      subject: '笔试通知',
      sender: 'hr@example.com',
      received_at: '2024-01-02T09:00:00Z',
      classification: 'exam',
      confidence: 0.85,
      payload: { exam_link: 'https://exam.example.com' },
      status: 'needs_review',
      application: {
        id: 101,
        company_name: '示例公司',
        position_title: '后端实习生',
        status: '已投递',
        reminder_enabled: false
      }
    }
    mailEventStore.pendingEvents = [mailEvent]

    const wrapper = mount(ReminderManager, {
      global: {
        stubs: {
          'a-card': simpleStub,
          'a-badge': simpleStub,
          'a-tabs': { template: '<div><slot /></div>' },
          'a-tab-pane': { template: '<div><slot /></div>' },
          'a-radio-group': simpleStub,
          'a-radio-button': simpleStub,
          'a-divider': simpleStub,
          'a-list': listStub,
          'a-list-item': listItemStub,
          'a-list-item-meta': listItemMetaStub,
          'a-button': buttonStub,
          'a-avatar': simpleStub,
          'a-tag': simpleStub,
          'a-empty': simpleStub,
          'a-spin': simpleStub,
          'a-space': simpleStub,
          'a-progress': simpleStub,
          'a-modal': { template: '<div><slot /></div>' },
          'a-form': simpleStub,
          'a-form-item': simpleStub,
          'a-radio': simpleStub,
          'a-date-picker': { template: '<input />' },
          'a-checkbox-group': simpleStub,
          'a-checkbox': simpleStub,
          'a-textarea': { template: '<textarea />' },
          'a-switch': { template: '<input type="checkbox" />' }
        }
      }
    })

    ;(wrapper.vm as any).activeTab = 'mailEvents'
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('笔试通知')
    expect(wrapper.text()).toContain('笔试链接')

    const componentVm = wrapper.vm as any
    expect(typeof componentVm.markEventProcessed).toBe('function')

    vi.spyOn(mailEventStore, 'updateEventStatus').mockResolvedValueOnce({ ...mailEvent, status: 'processed' })
    await componentVm.markEventProcessed(mailEvent)
    expect(mailEventStore.updateEventStatus).toHaveBeenCalledWith(1, { status: 'processed' })
  })
})
