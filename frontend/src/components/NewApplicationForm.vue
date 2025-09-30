<template>
  <a-modal
    :open="visible"
    :title="isEdit ? '编辑投递记录' : '添加投递记录'"
    :confirm-loading="loading"
    @ok="handleSubmit"
    @cancel="handleCancel"
    width="600px"
  >
    <a-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      :label-col="{ span: 6 }"
      :wrapper-col="{ span: 18 }"
    >
      <a-form-item label="公司名称" name="company_name">
        <a-input v-model:value="formData.company_name" placeholder="请输入公司名称" />
      </a-form-item>

      <a-form-item label="职位标题" name="position_title">
        <a-input v-model:value="formData.position_title" placeholder="请输入职位标题" />
      </a-form-item>

      <a-form-item label="企业属性" name="company_attribute">
        <a-select v-model:value="formData.company_attribute" placeholder="请选择企业属性">
          <a-select-option value="央国企">央国企</a-select-option>
          <a-select-option value="私企">私企</a-select-option>
        </a-select>
      </a-form-item>

      <a-form-item label="投递日期" name="application_date">
        <a-date-picker
          v-model:value="formData.application_date"
          placeholder="选择投递日期"
          format="YYYY-MM-DD"
          style="width: 100%"
        />
      </a-form-item>

      <a-form-item label="当前状态" name="status">
        <a-select v-model:value="formData.status" placeholder="选择当前状态">
          <a-select-opt-group label="基础状态">
            <a-select-option :value="ApplicationStatus.APPLIED">{{ ApplicationStatus.APPLIED }}</a-select-option>
            <a-select-option :value="ApplicationStatus.RESUME_SCREENING">{{ ApplicationStatus.RESUME_SCREENING }}</a-select-option>
            <a-select-option :value="ApplicationStatus.RESUME_SCREENING_FAIL">
              <span :style="{ color: StatusHelper.getStatusColor(ApplicationStatus.RESUME_SCREENING_FAIL) }">
                {{ ApplicationStatus.RESUME_SCREENING_FAIL }}
              </span>
            </a-select-option>
          </a-select-opt-group>
          
          <a-select-opt-group label="笔试阶段">
            <a-select-option :value="ApplicationStatus.WRITTEN_TEST">{{ ApplicationStatus.WRITTEN_TEST }}</a-select-option>
            <a-select-option :value="ApplicationStatus.WRITTEN_TEST_PASS">
              <span :style="{ color: StatusHelper.getStatusColor(ApplicationStatus.WRITTEN_TEST_PASS) }">
                {{ ApplicationStatus.WRITTEN_TEST_PASS }}
              </span>
            </a-select-option>
            <a-select-option :value="ApplicationStatus.WRITTEN_TEST_FAIL">
              <span :style="{ color: StatusHelper.getStatusColor(ApplicationStatus.WRITTEN_TEST_FAIL) }">
                {{ ApplicationStatus.WRITTEN_TEST_FAIL }}
              </span>
            </a-select-option>
          </a-select-opt-group>

          <a-select-opt-group label="一面阶段">
            <a-select-option :value="ApplicationStatus.FIRST_INTERVIEW">{{ ApplicationStatus.FIRST_INTERVIEW }}</a-select-option>
            <a-select-option :value="ApplicationStatus.FIRST_PASS">
              <span :style="{ color: StatusHelper.getStatusColor(ApplicationStatus.FIRST_PASS) }">
                {{ ApplicationStatus.FIRST_PASS }}
              </span>
            </a-select-option>
            <a-select-option :value="ApplicationStatus.FIRST_FAIL">
              <span :style="{ color: StatusHelper.getStatusColor(ApplicationStatus.FIRST_FAIL) }">
                {{ ApplicationStatus.FIRST_FAIL }}
              </span>
            </a-select-option>
          </a-select-opt-group>

          <a-select-opt-group label="二面阶段">
            <a-select-option :value="ApplicationStatus.SECOND_INTERVIEW">{{ ApplicationStatus.SECOND_INTERVIEW }}</a-select-option>
            <a-select-option :value="ApplicationStatus.SECOND_PASS">
              <span :style="{ color: StatusHelper.getStatusColor(ApplicationStatus.SECOND_PASS) }">
                {{ ApplicationStatus.SECOND_PASS }}
              </span>
            </a-select-option>
            <a-select-option :value="ApplicationStatus.SECOND_FAIL">
              <span :style="{ color: StatusHelper.getStatusColor(ApplicationStatus.SECOND_FAIL) }">
                {{ ApplicationStatus.SECOND_FAIL }}
              </span>
            </a-select-option>
          </a-select-opt-group>

          <a-select-opt-group label="三面阶段">
            <a-select-option :value="ApplicationStatus.THIRD_INTERVIEW">{{ ApplicationStatus.THIRD_INTERVIEW }}</a-select-option>
            <a-select-option :value="ApplicationStatus.THIRD_PASS">
              <span :style="{ color: StatusHelper.getStatusColor(ApplicationStatus.THIRD_PASS) }">
                {{ ApplicationStatus.THIRD_PASS }}
              </span>
            </a-select-option>
            <a-select-option :value="ApplicationStatus.THIRD_FAIL">
              <span :style="{ color: StatusHelper.getStatusColor(ApplicationStatus.THIRD_FAIL) }">
                {{ ApplicationStatus.THIRD_FAIL }}
              </span>
            </a-select-option>
          </a-select-opt-group>

          <a-select-opt-group label="HR面阶段">
            <a-select-option :value="ApplicationStatus.HR_INTERVIEW">{{ ApplicationStatus.HR_INTERVIEW }}</a-select-option>
            <a-select-option :value="ApplicationStatus.HR_PASS">
              <span :style="{ color: StatusHelper.getStatusColor(ApplicationStatus.HR_PASS) }">
                {{ ApplicationStatus.HR_PASS }}
              </span>
            </a-select-option>
            <a-select-option :value="ApplicationStatus.HR_FAIL">
              <span :style="{ color: StatusHelper.getStatusColor(ApplicationStatus.HR_FAIL) }">
                {{ ApplicationStatus.HR_FAIL }}
              </span>
            </a-select-option>
          </a-select-opt-group>

          <a-select-opt-group label="最终状态">
            <a-select-option :value="ApplicationStatus.OFFER_WAITING">
              <span :style="{ color: StatusHelper.getStatusColor(ApplicationStatus.OFFER_WAITING) }">
                {{ ApplicationStatus.OFFER_WAITING }}
              </span>
            </a-select-option>
            <a-select-option :value="ApplicationStatus.OFFER_RECEIVED">
              <span :style="{ color: StatusHelper.getStatusColor(ApplicationStatus.OFFER_RECEIVED) }">
                {{ ApplicationStatus.OFFER_RECEIVED }}
              </span>
            </a-select-option>
            <a-select-option :value="ApplicationStatus.OFFER_ACCEPTED">
              <span :style="{ color: StatusHelper.getStatusColor(ApplicationStatus.OFFER_ACCEPTED) }">
                {{ ApplicationStatus.OFFER_ACCEPTED }}
              </span>
            </a-select-option>
            <a-select-option :value="ApplicationStatus.REJECTED">
              <span :style="{ color: StatusHelper.getStatusColor(ApplicationStatus.REJECTED) }">
                {{ ApplicationStatus.REJECTED }}
              </span>
            </a-select-option>
            <a-select-option :value="ApplicationStatus.PROCESS_FINISHED">
              <span :style="{ color: StatusHelper.getStatusColor(ApplicationStatus.PROCESS_FINISHED) }">
                {{ ApplicationStatus.PROCESS_FINISHED }}
              </span>
            </a-select-option>
          </a-select-opt-group>
        </a-select>
      </a-form-item>

      <a-form-item label="薪资范围" name="salary_range">
        <a-input v-model:value="formData.salary_range" placeholder="如：15-25K" />
      </a-form-item>

      <a-form-item label="工作地点" name="work_location">
        <a-input v-model:value="formData.work_location" placeholder="如：北京/上海" />
      </a-form-item>

      <a-form-item label="备注信息" name="notes">
        <a-textarea
          v-model:value="formData.notes"
          placeholder="投递渠道、面试反馈等"
          :rows="3"
        />
      </a-form-item>

      <a-divider orientation="left">考试/面试提醒设置</a-divider>

      <a-form-item label="提醒类别" name="reminder_type">
        <a-select
          v-model:value="formData.reminder_category"
          placeholder="选择提醒类别"
          style="width: 100%"
        >
          <a-select-option value="interview">面试提醒</a-select-option>
          <a-select-option value="written">笔试提醒</a-select-option>
        </a-select>
      </a-form-item>

      <a-form-item v-if="formData.reminder_category === 'interview'" label="面试时间" name="interview_time">
        <a-date-picker
          v-model:value="formData.interview_time"
          show-time
          placeholder="选择面试时间"
          format="YYYY-MM-DD HH:mm"
          style="width: 100%"
        />
      </a-form-item>

      <a-form-item v-if="formData.reminder_category === 'written'" label="笔试时间" name="written_time">
        <a-date-picker
          v-model:value="formData.written_time"
          show-time
          placeholder="选择笔试时间"
          format="YYYY-MM-DD HH:mm"
          style="width: 100%"
        />
      </a-form-item>

      <a-form-item label="启用提醒" name="reminder_enabled">
        <a-switch v-model:checked="formData.reminder_enabled" />
        <span style="margin-left: 10px">开启提醒</span>
      </a-form-item>

      <a-form-item v-if="formData.reminder_enabled" label="提醒时间" name="reminder_time">
        <a-date-picker
          v-model:value="formData.reminder_time"
          show-time
          placeholder="选择提醒时间"
          format="YYYY-MM-DD HH:mm"
          style="width: 100%"
        />
        <div style="margin-top: 5px">
          <a-space>
            <a-button size="small" @click="setReminderTime(15)">提前15分钟</a-button>
            <a-button size="small" @click="setReminderTime(30)">提前30分钟</a-button>
            <a-button size="small" @click="setReminderTime(60)">提前1小时</a-button>
          </a-space>
        </div>
      </a-form-item>

      <a-form-item label="跟进日期" name="follow_up_date">
        <a-date-picker
          v-model:value="formData.follow_up_date"
          placeholder="选择跟进日期"
          format="YYYY-MM-DD"
          style="width: 100%"
        />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue'
import dayjs, { type Dayjs } from 'dayjs'
import { useJobApplicationStore } from '../stores/jobApplication'
import type { JobApplication } from '../types'
import { ApplicationStatus, StatusHelper } from '../types'

interface Props {
  visible: boolean
  initialData?: JobApplication | null
}

interface Emits {
  (e: 'update:visible', value: boolean): void
  (e: 'success'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const jobStore = useJobApplicationStore()
const formRef = ref()
const loading = ref(false)

// 表单数据
const formData = reactive<{
  company_name: string
  position_title: string
  application_date: Dayjs | null
  status: ApplicationStatus
  company_attribute: '' | '央国企' | '私企'
  salary_range: string
  work_location: string
  notes: string
  interview_time: Dayjs | null
  written_time: Dayjs | null
  reminder_time: Dayjs | null
  reminder_enabled: boolean
  reminder_category: 'interview' | 'written' | 'follow_up'
  follow_up_date: Dayjs | null
}>({
  company_name: '',
  position_title: '',
  application_date: null,
  status: '已投递' as ApplicationStatus,
  company_attribute: '',
  salary_range: '',
  work_location: '',
  notes: '',
  interview_time: null,
  written_time: null,
  reminder_time: null,
  reminder_enabled: false,
  reminder_category: 'interview',
  follow_up_date: null
})

// 表单验证规则（编辑时企业属性可选）
const rules = computed(() => ({
  company_name: [
    { required: true, message: '请输入公司名称', trigger: 'blur' }
  ],
  position_title: [
    { required: true, message: '请输入职位标题', trigger: 'blur' }
  ],
  company_attribute: [
    { required: !isEdit.value, message: '请选择企业属性', trigger: 'change' }
  ]
}))

// 计算属性
const isEdit = computed(() => !!props.initialData)

// 将 Dayjs 值转换为 ISO 字符串，便于后端解析
const toISOStringOrUndefined = (value: Dayjs | null) => {
  return value ? value.toDate().toISOString() : undefined
}

// 设置提醒时间（提前N分钟）
const setReminderTime = (minutes: number) => {
  const baseTime = formData.reminder_category === 'written'
    ? formData.written_time
    : formData.interview_time
  if (baseTime) {
    formData.reminder_time = baseTime.subtract(minutes, 'minute')
  }
}

// 重置表单
const resetForm = () => {
  formData.company_name = ''
  formData.position_title = ''
  formData.application_date = dayjs()
  formData.status = '已投递' as ApplicationStatus
  formData.company_attribute = ''
  formData.salary_range = ''
  formData.work_location = ''
  formData.notes = ''
  formData.interview_time = null
  formData.written_time = null
  formData.reminder_time = null
  formData.reminder_enabled = false
  formData.reminder_category = 'interview'
  formData.follow_up_date = null
  formRef.value?.clearValidate()
}

// 监听initialData变化，填充表单
watch(() => props.initialData, (app) => {
  if (app) {
    formData.company_name = app.company_name
    formData.position_title = app.position_title
    formData.application_date = app.application_date ? dayjs(app.application_date) : null
    formData.status = app.status
    formData.company_attribute = (app.company_attribute as any) || ''
    formData.salary_range = app.salary_range || ''
    formData.work_location = app.work_location || ''
    formData.notes = app.notes || ''
    formData.interview_time = app.interview_time && app.interview_type !== '笔试'
      ? dayjs(app.interview_time)
      : null
    formData.written_time = app.interview_time && app.interview_type === '笔试'
      ? dayjs(app.interview_time)
      : null
    formData.reminder_time = app.reminder_time ? dayjs(app.reminder_time) : null
    formData.reminder_enabled = app.reminder_enabled || false
    if (app.reminder_enabled) {
      if (app.interview_time && app.interview_type === '笔试') {
        formData.reminder_category = 'written'
      } else if (app.interview_time) {
        formData.reminder_category = 'interview'
      } else {
        formData.reminder_category = 'follow_up'
      }
    } else {
      formData.reminder_category = 'interview'
    }
    formData.follow_up_date = app.follow_up_date ? dayjs(app.follow_up_date) : null
  } else {
    resetForm()
  }
}, { immediate: true })

// 提交表单
const handleSubmit = async () => {
  try {
    await formRef.value.validateFields()
    loading.value = true

    const submitData = {
      company_name: formData.company_name,
      position_title: formData.position_title,
      application_date: formData.application_date?.format('YYYY-MM-DD') || dayjs().format('YYYY-MM-DD'),
      status: formData.status as ApplicationStatus,
      company_attribute: formData.company_attribute as '央国企' | '私企',
      salary_range: formData.salary_range || undefined,
      work_location: formData.work_location || undefined,
      notes: formData.notes || undefined,
      interview_time: formData.reminder_category === 'interview'
        ? toISOStringOrUndefined(formData.interview_time)
        : toISOStringOrUndefined(formData.written_time),
      reminder_time: toISOStringOrUndefined(formData.reminder_time),
      reminder_enabled: formData.reminder_enabled,
      follow_up_date: formData.follow_up_date?.format('YYYY-MM-DD') || undefined,
      hr_name: formData.hr_name || undefined,
      hr_phone: formData.hr_phone || undefined,
      hr_email: formData.hr_email || undefined,
      interview_location: formData.interview_location || undefined,
      interview_type: formData.reminder_category === 'written'
        ? '笔试'
        : formData.interview_type || undefined
    }

    if (isEdit.value && props.initialData) {
      await jobStore.updateApplication(props.initialData.id, submitData)
    } else {
      await jobStore.createApplication(submitData)
    }

    emit('success')
  } catch (error) {
    console.error('表单提交失败:', error)
  } finally {
    loading.value = false
  }
}

// 取消
const handleCancel = () => {
  emit('update:visible', false)
  resetForm()
}

console.log('NewApplicationForm组件加载成功')
</script>
