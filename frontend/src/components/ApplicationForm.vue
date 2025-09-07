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
      <a-form-item label="公司名称" name="company_name" required>
        <a-input v-model:value="formData.company_name" placeholder="请输入公司名称" />
      </a-form-item>

      <a-form-item label="职位标题" name="position_title" required>
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
          <a-select-option
            v-for="status in statusOptions"
            :key="status.value"
            :value="status.value"
          >
            {{ status.label }}
          </a-select-option>
        </a-select>
      </a-form-item>

      <a-form-item label="薪资范围" name="salary_range">
        <a-input v-model:value="formData.salary_range" placeholder="如：15-25K" />
      </a-form-item>

      <a-form-item label="工作地点" name="work_location">
        <a-input v-model:value="formData.work_location" placeholder="如：北京/上海" />
      </a-form-item>

      <a-form-item label="联系信息" name="contact_info">
        <a-textarea
          v-model:value="formData.contact_info"
          placeholder="HR联系方式、邮箱等"
          :rows="2"
        />
      </a-form-item>

      <a-form-item label="职位描述" name="job_description">
        <a-textarea
          v-model:value="formData.job_description"
          placeholder="职位要求和描述"
          :rows="3"
        />
      </a-form-item>

      <a-form-item label="备注信息" name="notes">
        <a-textarea
          v-model:value="formData.notes"
          placeholder="投递渠道、面试反馈等"
          :rows="3"
        />
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue'
import dayjs, { type Dayjs } from 'dayjs'
import { useJobApplicationStore } from '../stores/jobApplication'
import { ApplicationStatus, type JobApplication } from '../types'

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
const formData = reactive({
  company_name: '',
  position_title: '',
  application_date: null as Dayjs | null,
  status: ApplicationStatus.APPLIED as ApplicationStatus,
  salary_range: '',
  work_location: '',
  contact_info: '',
  job_description: '',
  notes: ''
})

// 状态选项
const statusOptions = computed(() => [
  { label: '已投递', value: ApplicationStatus.APPLIED },
  { label: '笔试中', value: ApplicationStatus.WRITTEN_TEST },
  { label: '笔试通过', value: ApplicationStatus.WRITTEN_TEST_PASS },
  { label: '一面中', value: ApplicationStatus.FIRST_INTERVIEW },
  { label: '一面通过', value: ApplicationStatus.FIRST_PASS },
  { label: '二面中', value: ApplicationStatus.SECOND_INTERVIEW },
  { label: '二面通过', value: ApplicationStatus.SECOND_PASS },
  { label: '三面中', value: ApplicationStatus.THIRD_INTERVIEW },
  { label: '三面通过', value: ApplicationStatus.THIRD_PASS },
  { label: 'HR面中', value: ApplicationStatus.HR_INTERVIEW },
  { label: 'HR面通过', value: ApplicationStatus.HR_PASS },
  { label: '待发offer', value: ApplicationStatus.OFFER_WAITING },
  { label: '已拒绝', value: ApplicationStatus.REJECTED },
  { label: '已收到offer', value: ApplicationStatus.OFFER_RECEIVED },
  { label: '已接受offer', value: ApplicationStatus.OFFER_ACCEPTED },
  { label: '流程结束', value: ApplicationStatus.PROCESS_FINISHED }
])

// 表单验证规则
const rules = {
  company_name: [
    { required: true, message: '请输入公司名称', trigger: 'blur' },
    { min: 1, max: 100, message: '公司名称长度应在1-100个字符之间', trigger: 'blur' }
  ],
  position_title: [
    { required: true, message: '请输入职位标题', trigger: 'blur' },
    { min: 1, max: 100, message: '职位标题长度应在1-100个字符之间', trigger: 'blur' }
  ],
  salary_range: [
    { max: 50, message: '薪资范围长度不能超过50个字符', trigger: 'blur' }
  ],
  work_location: [
    { max: 100, message: '工作地点长度不能超过100个字符', trigger: 'blur' }
  ]
}

// 计算属性
const isEdit = computed(() => !!props.initialData)

// 监听initialData变化，填充表单
watch(() => props.initialData, (app) => {
  if (app) {
    formData.company_name = app.company_name
    formData.position_title = app.position_title
    formData.application_date = app.application_date ? dayjs(app.application_date) : null
    formData.status = app.status
    formData.salary_range = app.salary_range || ''
    formData.work_location = app.work_location || ''
    formData.contact_info = app.contact_info || ''
    formData.job_description = app.job_description || ''
    formData.notes = app.notes || ''
  } else {
    resetForm()
  }
}, { immediate: true })

// 重置表单
const resetForm = () => {
  formData.company_name = ''
  formData.position_title = ''
  formData.application_date = dayjs()
  formData.status = ApplicationStatus.APPLIED as ApplicationStatus
  formData.salary_range = ''
  formData.work_location = ''
  formData.contact_info = ''
  formData.job_description = ''
  formData.notes = ''
  formRef.value?.clearValidate()
}

// 提交表单
const handleSubmit = async () => {
  try {
    await formRef.value.validateFields()
    loading.value = true

    const submitData = {
      company_name: formData.company_name,
      position_title: formData.position_title,
      application_date: formData.application_date?.format('YYYY-MM-DD') || dayjs().format('YYYY-MM-DD'),
      status: formData.status,
      salary_range: formData.salary_range || undefined,
      work_location: formData.work_location || undefined,
      contact_info: formData.contact_info || undefined,
      job_description: formData.job_description || undefined,
      notes: formData.notes || undefined
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

// 取消操作
const handleCancel = () => {
  emit('update:visible', false)
  setTimeout(() => resetForm(), 300) // 延迟重置，避免闪烁
}
</script>

<style scoped>
.ant-form-item {
  margin-bottom: 16px;
}
</style>