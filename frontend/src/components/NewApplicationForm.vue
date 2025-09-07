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

      <a-divider orientation="left">面试提醒设置</a-divider>

      <a-form-item label="面试时间" name="interview_time">
        <a-date-picker
          v-model:value="formData.interview_time"
          show-time
          placeholder="选择面试时间"
          format="YYYY-MM-DD HH:mm"
          style="width: 100%"
        />
      </a-form-item>

      <a-form-item label="启用提醒" name="reminder_enabled">
        <a-switch v-model:checked="formData.reminder_enabled" />
        <span style="margin-left: 10px">在面试前提醒</span>
      </a-form-item>

      <a-form-item 
        v-if="formData.reminder_enabled" 
        label="提醒时间" 
        name="reminder_time"
      >
        <a-date-picker
          v-model:value="formData.reminder_time"
          show-time
          placeholder="选择提醒时间"
          format="YYYY-MM-DD HH:mm"
          style="width: 100%"
        />
        <div style="margin-top: 5px">
          <a-space>
            <a-button size="small" @click="setReminderTime(15)">面试前15分钟</a-button>
            <a-button size="small" @click="setReminderTime(30)">面试前30分钟</a-button>
            <a-button size="small" @click="setReminderTime(60)">面试前1小时</a-button>
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
  salary_range: string
  work_location: string
  notes: string
  interview_time: Dayjs | null
  reminder_time: Dayjs | null
  reminder_enabled: boolean
  follow_up_date: Dayjs | null
}>({
  company_name: '',
  position_title: '',
  application_date: null,
  status: '已投递' as ApplicationStatus,
  salary_range: '',
  work_location: '',
  notes: '',
  interview_time: null,
  reminder_time: null,
  reminder_enabled: false,
  follow_up_date: null
})

// 表单验证规则
const rules = {
  company_name: [
    { required: true, message: '请输入公司名称', trigger: 'blur' }
  ],
  position_title: [
    { required: true, message: '请输入职位标题', trigger: 'blur' }
  ]
}

// 计算属性
const isEdit = computed(() => !!props.initialData)

// 设置提醒时间（提前N分钟）
const setReminderTime = (minutes: number) => {
  if (formData.interview_time) {
    formData.reminder_time = formData.interview_time.subtract(minutes, 'minute')
  }
}

// 重置表单
const resetForm = () => {
  formData.company_name = ''
  formData.position_title = ''
  formData.application_date = dayjs()
  formData.status = '已投递' as ApplicationStatus
  formData.salary_range = ''
  formData.work_location = ''
  formData.notes = ''
  formData.interview_time = null
  formData.reminder_time = null
  formData.reminder_enabled = false
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
    formData.salary_range = app.salary_range || ''
    formData.work_location = app.work_location || ''
    formData.notes = app.notes || ''
    formData.interview_time = app.interview_time ? dayjs(app.interview_time) : null
    formData.reminder_time = app.reminder_time ? dayjs(app.reminder_time) : null
    formData.reminder_enabled = app.reminder_enabled || false
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
      salary_range: formData.salary_range || undefined,
      work_location: formData.work_location || undefined,
      notes: formData.notes || undefined,
      interview_time: formData.interview_time?.format('YYYY-MM-DD HH:mm:ss') || undefined,
      reminder_time: formData.reminder_time?.format('YYYY-MM-DD HH:mm:ss') || undefined,
      reminder_enabled: formData.reminder_enabled,
      follow_up_date: formData.follow_up_date?.format('YYYY-MM-DD') || undefined
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