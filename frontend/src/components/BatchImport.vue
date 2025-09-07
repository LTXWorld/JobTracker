<template>
  <a-modal
    :visible="visible"
    title="批量导入投递记录"
    width="800px"
    :confirmLoading="loading"
    @ok="handleImport"
    @cancel="handleCancel"
  >
    <div class="import-container">
      <!-- 步骤条 -->
      <a-steps :current="currentStep" class="steps-container">
        <a-step title="上传文件" />
        <a-step title="映射字段" />
        <a-step title="预览确认" />
      </a-steps>

      <!-- 步骤1: 上传文件 -->
      <div v-if="currentStep === 0" class="step-content">
        <a-upload-dragger
          v-model:fileList="fileList"
          :maxCount="1"
          :beforeUpload="beforeUpload"
          accept=".xlsx,.xls,.csv"
        >
          <p class="ant-upload-drag-icon">
            <InboxOutlined />
          </p>
          <p class="ant-upload-text">点击或拖拽文件到此区域上传</p>
          <p class="ant-upload-hint">
            支持 .xlsx, .xls, .csv 格式文件，最大 10MB
          </p>
        </a-upload-dragger>

        <a-divider />
        
        <div class="template-section">
          <h4>下载模板</h4>
          <a-space>
            <a-button @click="downloadTemplate('excel')">
              <template #icon><FileExcelOutlined /></template>
              Excel 模板
            </a-button>
            <a-button @click="downloadTemplate('csv')">
              <template #icon><FileTextOutlined /></template>
              CSV 模板
            </a-button>
          </a-space>
          <div class="template-hint">
            <p>模板包含以下字段：</p>
            <ul>
              <li>公司名称（必填）</li>
              <li>职位名称（必填）</li>
              <li>投递日期（必填，格式：YYYY-MM-DD）</li>
              <li>投递状态（可选，默认：已投递）</li>
              <li>薪资范围（可选）</li>
              <li>工作地点（可选）</li>
              <li>备注（可选）</li>
            </ul>
          </div>
        </div>
      </div>

      <!-- 步骤2: 字段映射 -->
      <div v-if="currentStep === 1" class="step-content">
        <a-alert
          message="请将文件中的列映射到对应的字段"
          type="info"
          showIcon
          class="mapping-alert"
        />
        
        <a-table
          :columns="mappingColumns"
          :dataSource="mappingData"
          :pagination="false"
          size="small"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'target'">
              <a-select
                v-model:value="fieldMapping[record.source]"
                style="width: 200px"
                placeholder="选择映射字段"
              >
                <a-select-option
                  v-for="field in targetFields"
                  :key="field.value"
                  :value="field.value"
                  :disabled="isFieldMapped(field.value, record.source)"
                >
                  {{ field.label }}
                  {{ field.required ? '*' : '' }}
                </a-select-option>
              </a-select>
            </template>
            <template v-else-if="column.key === 'preview'">
              {{ record.preview }}
            </template>
          </template>
        </a-table>
      </div>

      <!-- 步骤3: 预览确认 -->
      <div v-if="currentStep === 2" class="step-content">
        <a-alert
          :message="`即将导入 ${previewData.length} 条记录`"
          :type="invalidRows.length > 0 ? 'warning' : 'success'"
          showIcon
          class="preview-alert"
        />
        
        <div v-if="invalidRows.length > 0" class="invalid-section">
          <h4>以下记录存在问题，将被跳过：</h4>
          <a-table
            :columns="previewColumns"
            :dataSource="invalidRows"
            :pagination="false"
            size="small"
            :scroll="{ y: 200 }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'error'">
                <a-tag color="error">{{ record.error }}</a-tag>
              </template>
            </template>
          </a-table>
        </div>

        <div class="valid-section">
          <h4>有效记录预览（前10条）：</h4>
          <a-table
            :columns="previewColumns"
            :dataSource="validRows.slice(0, 10)"
            :pagination="false"
            size="small"
            :scroll="{ y: 300 }"
          />
        </div>
      </div>
    </div>

    <template #footer>
      <a-button @click="handleCancel">取消</a-button>
      <a-button v-if="currentStep > 0" @click="prevStep">上一步</a-button>
      <a-button 
        v-if="currentStep < 2" 
        type="primary" 
        @click="nextStep"
        :disabled="!canProceed"
      >
        下一步
      </a-button>
      <a-button
        v-if="currentStep === 2"
        type="primary"
        @click="handleImport"
        :loading="loading"
        :disabled="validRows.length === 0"
      >
        确认导入
      </a-button>
    </template>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { message } from 'ant-design-vue'
import { 
  InboxOutlined, FileExcelOutlined, FileTextOutlined 
} from '@ant-design/icons-vue'
import * as XLSX from 'xlsx'
import Papa from 'papaparse'
import dayjs from 'dayjs'
import { useJobApplicationStore } from '../stores/jobApplication'
import { ApplicationStatus, type JobApplication } from '../types'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  'update:visible': (value: boolean) => void
  'success': () => void
}>()

const jobStore = useJobApplicationStore()

// 响应式数据
const currentStep = ref(0)
const loading = ref(false)
const fileList = ref<any[]>([])
const parsedData = ref<any[]>([])
const fieldMapping = ref<Record<string, string>>({})
const previewData = ref<any[]>([])
const validRows = ref<any[]>([])
const invalidRows = ref<any[]>([])

// 目标字段配置
const targetFields = [
  { label: '公司名称', value: 'company_name', required: true },
  { label: '职位名称', value: 'position_title', required: true },
  { label: '投递日期', value: 'application_date', required: true },
  { label: '投递状态', value: 'status', required: false },
  { label: '薪资范围', value: 'salary_range', required: false },
  { label: '工作地点', value: 'work_location', required: false },
  { label: '备注', value: 'notes', required: false }
]

// 状态映射
const statusMap: Record<string, ApplicationStatus> = {
  '已投递': ApplicationStatus.APPLIED,
  '笔试中': ApplicationStatus.WRITTEN_TEST,
  '笔试通过': ApplicationStatus.WRITTEN_TEST_PASS,
  '一面中': ApplicationStatus.FIRST_INTERVIEW,
  '一面通过': ApplicationStatus.FIRST_PASS,
  '二面中': ApplicationStatus.SECOND_INTERVIEW,
  '二面通过': ApplicationStatus.SECOND_PASS,
  '三面中': ApplicationStatus.THIRD_INTERVIEW,
  '三面通过': ApplicationStatus.THIRD_PASS,
  'HR面': ApplicationStatus.HR_INTERVIEW,
  'HR通过': ApplicationStatus.HR_PASS,
  '等待offer': ApplicationStatus.OFFER_WAITING,
  '已挂': ApplicationStatus.REJECTED,
  '已拒绝': ApplicationStatus.REJECTED,
  '收到offer': ApplicationStatus.OFFER_RECEIVED,
  '接受offer': ApplicationStatus.OFFER_ACCEPTED,
  '流程结束': ApplicationStatus.PROCESS_FINISHED
}

// 映射表格列配置
const mappingColumns = [
  { title: '文件列名', dataIndex: 'source', key: 'source' },
  { title: '示例数据', dataIndex: 'preview', key: 'preview' },
  { title: '映射到字段', key: 'target' }
]

// 映射数据
const mappingData = computed(() => {
  if (parsedData.value.length === 0) return []
  const headers = Object.keys(parsedData.value[0])
  return headers.map(header => ({
    key: header,
    source: header,
    preview: parsedData.value[0][header]
  }))
})

// 预览表格列配置
const previewColumns = computed(() => {
  const cols = [
    { title: '公司名称', dataIndex: 'company_name', key: 'company_name' },
    { title: '职位名称', dataIndex: 'position_title', key: 'position_title' },
    { title: '投递日期', dataIndex: 'application_date', key: 'application_date' },
    { title: '状态', dataIndex: 'status', key: 'status' },
    { title: '薪资', dataIndex: 'salary_range', key: 'salary_range' },
    { title: '地点', dataIndex: 'work_location', key: 'work_location' }
  ]
  if (invalidRows.value.length > 0) {
    cols.push({ title: '错误原因', key: 'error' })
  }
  return cols
})

// 是否可以进入下一步
const canProceed = computed(() => {
  if (currentStep.value === 0) {
    return fileList.value.length > 0 && parsedData.value.length > 0
  }
  if (currentStep.value === 1) {
    // 检查必填字段是否都已映射
    const requiredFields = targetFields.filter(f => f.required)
    return requiredFields.every(field => 
      Object.values(fieldMapping.value).includes(field.value)
    )
  }
  return true
})

// 检查字段是否已被映射
const isFieldMapped = (fieldValue: string, currentSource: string) => {
  return Object.entries(fieldMapping.value).some(
    ([source, target]) => target === fieldValue && source !== currentSource
  )
}

// 文件上传前处理
const beforeUpload = (file: File) => {
  const isValidType = ['application/vnd.ms-excel', 
    'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    'text/csv'].includes(file.type) || 
    /\.(xlsx|xls|csv)$/i.test(file.name)
  
  if (!isValidType) {
    message.error('只支持 Excel 或 CSV 文件！')
    return false
  }
  
  const isLt10M = file.size / 1024 / 1024 < 10
  if (!isLt10M) {
    message.error('文件大小不能超过 10MB！')
    return false
  }
  
  parseFile(file)
  return false
}

// 解析文件
const parseFile = async (file: File) => {
  const extension = file.name.split('.').pop()?.toLowerCase()
  
  try {
    if (extension === 'csv') {
      // 解析 CSV
      Papa.parse(file, {
        header: true,
        complete: (results) => {
          parsedData.value = results.data.filter((row: any) => 
            Object.values(row).some(v => v)
          )
          autoMapFields()
        },
        error: () => {
          message.error('CSV 文件解析失败')
        }
      })
    } else {
      // 解析 Excel
      const reader = new FileReader()
      reader.onload = (e) => {
        const data = e.target?.result
        const workbook = XLSX.read(data, { type: 'binary' })
        const sheetName = workbook.SheetNames[0]
        const worksheet = workbook.Sheets[sheetName]
        const jsonData = XLSX.utils.sheet_to_json(worksheet)
        parsedData.value = jsonData
        autoMapFields()
      }
      reader.readAsBinaryString(file)
    }
  } catch (error) {
    message.error('文件解析失败')
  }
}

// 自动映射字段
const autoMapFields = () => {
  if (parsedData.value.length === 0) return
  
  const headers = Object.keys(parsedData.value[0])
  const mapping: Record<string, string> = {}
  
  // 尝试自动映射
  headers.forEach(header => {
    const lowerHeader = header.toLowerCase()
    if (lowerHeader.includes('公司') || lowerHeader.includes('company')) {
      mapping[header] = 'company_name'
    } else if (lowerHeader.includes('职位') || lowerHeader.includes('岗位') || 
               lowerHeader.includes('position') || lowerHeader.includes('title')) {
      mapping[header] = 'position_title'
    } else if (lowerHeader.includes('日期') || lowerHeader.includes('时间') || 
               lowerHeader.includes('date')) {
      mapping[header] = 'application_date'
    } else if (lowerHeader.includes('状态') || lowerHeader.includes('status')) {
      mapping[header] = 'status'
    } else if (lowerHeader.includes('薪资') || lowerHeader.includes('salary')) {
      mapping[header] = 'salary_range'
    } else if (lowerHeader.includes('地点') || lowerHeader.includes('地址') || 
               lowerHeader.includes('location')) {
      mapping[header] = 'work_location'
    } else if (lowerHeader.includes('备注') || lowerHeader.includes('note')) {
      mapping[header] = 'notes'
    }
  })
  
  fieldMapping.value = mapping
}

// 生成预览数据
const generatePreview = () => {
  validRows.value = []
  invalidRows.value = []
  
  parsedData.value.forEach((row, index) => {
    const mappedRow: any = { key: index }
    let error = ''
    
    // 映射字段
    Object.entries(fieldMapping.value).forEach(([source, target]) => {
      if (row[source]) {
        if (target === 'application_date') {
          // 处理日期
          const date = dayjs(row[source])
          if (date.isValid()) {
            mappedRow[target] = date.format('YYYY-MM-DD')
          } else {
            error = '日期格式错误'
          }
        } else if (target === 'status') {
          // 处理状态
          const status = statusMap[row[source]] || ApplicationStatus.APPLIED
          mappedRow[target] = status
        } else {
          mappedRow[target] = row[source]
        }
      }
    })
    
    // 验证必填字段
    if (!mappedRow.company_name) {
      error = '缺少公司名称'
    } else if (!mappedRow.position_title) {
      error = '缺少职位名称'
    } else if (!mappedRow.application_date) {
      error = '缺少投递日期'
    }
    
    // 设置默认值
    if (!mappedRow.status) {
      mappedRow.status = ApplicationStatus.APPLIED
    }
    
    if (error) {
      invalidRows.value.push({ ...mappedRow, error })
    } else {
      validRows.value.push(mappedRow)
    }
  })
  
  previewData.value = [...validRows.value, ...invalidRows.value]
}

// 下载模板
const downloadTemplate = (type: 'excel' | 'csv') => {
  const templateData = [
    {
      '公司名称': '示例科技有限公司',
      '职位名称': '前端开发工程师',
      '投递日期': dayjs().format('YYYY-MM-DD'),
      '投递状态': '已投递',
      '薪资范围': '15k-20k',
      '工作地点': '北京',
      '备注': '通过官网投递'
    }
  ]
  
  if (type === 'excel') {
    const ws = XLSX.utils.json_to_sheet(templateData)
    const wb = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(wb, ws, 'Sheet1')
    XLSX.writeFile(wb, '投递记录导入模板.xlsx')
  } else {
    const csv = Papa.unparse(templateData)
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.download = '投递记录导入模板.csv'
    link.click()
  }
}

// 上一步
const prevStep = () => {
  currentStep.value--
}

// 下一步
const nextStep = () => {
  if (currentStep.value === 1) {
    generatePreview()
  }
  currentStep.value++
}

// 确认导入
const handleImport = async () => {
  loading.value = true
  
  try {
    // 批量创建投递记录
    for (const row of validRows.value) {
      await jobStore.createApplication({
        company_name: row.company_name,
        position_title: row.position_title,
        application_date: row.application_date,
        status: row.status,
        salary_range: row.salary_range || '',
        work_location: row.work_location || '',
        notes: row.notes || ''
      })
    }
    
    message.success(`成功导入 ${validRows.value.length} 条记录`)
    emit('success')
    handleCancel()
  } catch (error) {
    message.error('导入失败，请重试')
  } finally {
    loading.value = false
  }
}

// 取消
const handleCancel = () => {
  emit('update:visible', false)
  // 重置状态
  currentStep.value = 0
  fileList.value = []
  parsedData.value = []
  fieldMapping.value = {}
  previewData.value = []
  validRows.value = []
  invalidRows.value = []
}

// 监听 visible 变化
watch(() => props.visible, (newVal) => {
  if (!newVal) {
    handleCancel()
  }
})
</script>

<style scoped>
.import-container {
  min-height: 400px;
}

.steps-container {
  margin-bottom: 24px;
}

.step-content {
  min-height: 350px;
}

.template-section {
  margin-top: 16px;
}

.template-hint {
  margin-top: 16px;
  padding: 12px;
  background: #f5f5f5;
  border-radius: 4px;
}

.template-hint ul {
  margin: 8px 0 0 20px;
}

.mapping-alert,
.preview-alert {
  margin-bottom: 16px;
}

.invalid-section,
.valid-section {
  margin-top: 16px;
}

.invalid-section h4,
.valid-section h4 {
  margin-bottom: 8px;
}
</style>