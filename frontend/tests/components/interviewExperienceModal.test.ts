import { describe, it, expect, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import InterviewExperienceModal from '../../src/components/InterviewExperienceModal.vue'

const modalStub = {
  template: '<div><slot /><slot name="footer" /></div>',
  props: ['open']
}

const buttonStub = { template: '<button @click="$emit(\'click\')"><slot /></button>' }
const textareaStub = {
  template: '<textarea :value="value" @input="$emit(\'update:value\', $event.target.value)" />',
  props: ['value']
}
const inputStub = {
  template: '<input :value="value" @input="$emit(\'update:value\', $event.target.value)" />',
  props: ['value']
}
const alertStub = { template: '<div><slot /></div>' }
const formStub = { template: '<form><slot /></form>' }
const formItemStub = { template: '<div><slot /></div>' }
const radioGroupStub = {
  template: '<div><slot /></div>',
  props: ['value'],
  emits: ['update:value']
}
const radioButtonStub = {
  template: '<button @click="$emit(\'update:value\', value)"><slot /></button>',
  props: ['value']
}
const radioStub = {
  inheritAttrs: false,
  props: ['value'],
  template: '<label :class="$attrs.class"><slot /></label>'
}
const cardStub = { template: '<div class="card"><slot /></div>' }
const dividerStub = { template: '<hr />' }

describe('InterviewExperienceModal', () => {
  const mountModal = (props?: Record<string, unknown>) => {
    return mount(InterviewExperienceModal, {
      props: {
        open: true,
        fromStatus: '一面中',
        toStatus: '二面中',
        ...props
      },
      global: {
        stubs: {
          'a-modal': modalStub,
          'a-alert': alertStub,
          'a-form': formStub,
          'a-form-item': formItemStub,
          'a-radio-group': radioGroupStub,
          'a-radio-button': radioButtonStub,
          'a-radio': radioStub,
          'a-textarea': textareaStub,
          'a-input': inputStub,
          'a-button': buttonStub,
          'a-card': cardStub,
          'a-divider': dividerStub
        }
      }
    })
  }

  beforeEach(() => {
    // wrapper 会在每个用例中单独挂载，确保状态干净
  })

  it('提交评分反馈时输出结构化数据', async () => {
    const wrapper = mountModal()
    const vm = wrapper.vm as any
    vm.rating = 'good'
    vm.note = '  面试官体验良好  '

    vm.handleSubmit()

    const submitted = wrapper.emitted('submit')?.[0]?.[0]
    expect(submitted).toEqual({
      skip: false,
      rating: 'good',
      note: '面试官体验良好'
    })
  })

  it('跳过反馈时携带原因与备注', async () => {
    const wrapper = mountModal({
      defaultValue: {
        skip: true,
        skip_reason: '时间紧张',
        note: '后续补记'
      }
    })
    const vm = wrapper.vm as any
    vm.mode = 'skip'
    vm.skipReason = '  暂无时间填写 '
    vm.note = '  待补充 '

    vm.handleSkip()

    const submitted = wrapper.emitted('submit')?.[0]?.[0]
    expect(submitted).toEqual({
      skip: true,
      skip_reason: '暂无时间填写',
      note: '待补充'
    })
  })

  it('点击取消时只发送取消事件', async () => {
    const wrapper = mountModal()
    const vm = wrapper.vm as any

    vm.handleCancel()

    expect(wrapper.emitted('cancel')).toBeTruthy()
    expect(wrapper.emitted('submit')).toBeUndefined()
  })
})
