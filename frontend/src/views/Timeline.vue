<template>
  <div class="timeline-page">
    <!-- 筛选栏 -->
    <FilterBar @change="handleFilterChange" />
    
    <a-card title="投递记录 Timeline" :loading="loading">
      <template #extra>
        <a-space>
          <a-button type="primary" @click="showCreateModal = true">
            <template #icon><PlusOutlined /></template>
            添加投递
          </a-button>
          <a-button @click="showImportModal = true">
            <template #icon><UploadOutlined /></template>
            批量导入
          </a-button>
          <a-button type="default" @click="showExportModal = true" :disabled="filteredApplications.length === 0">
            <template #icon><DownloadOutlined /></template>
            导出Excel
          </a-button>
          <a-button @click="fetchData">
            <template #icon><ReloadOutlined /></template>
            刷新
          </a-button>
        </a-space>
      </template>

      <!-- 快速统计（基于筛选结果） -->
      <div class="stats-row" v-if="filteredApplications.length > 0">
        <div class="stats-grid">
          <a-statistic title="总投递数" :value="totalApplicationsCount" />
          <a-statistic title="笔试中" :value="writtenCount" />
          <a-statistic title="面试中" :value="interviewingCount" />
          <a-statistic title="已挂" :value="failedSumCount" />
          <a-statistic title="已接offer" :value="acceptedOfferCount" />
        </div>
      </div>

      <a-divider />

      <!-- Timeline展示 -->
      <div v-if="filteredApplications.length === 0 && !loading" class="empty-state">
        <a-empty :description="hasFilters ? '没有符合条件的投递记录' : '暂无投递记录'">
          <a-button v-if="!hasFilters" type="primary" @click="showCreateModal = true">添加第一条记录</a-button>
          <a-button v-else @click="clearFilters">清除筛选条件</a-button>
        </a-empty>
      </div>

      <a-timeline v-else>
        <a-timeline-item 
          v-for="app in paginatedApplications" 
          :key="app.id"
          :color="getStatusColor(app.status)"
        >
          <template #dot>
            <component :is="getStatusIcon(app.status)" />
          </template>
          
          <div class="timeline-item">
            <div class="timeline-header">
              <h3>
                <span v-html="highlightKeyword(app.company_name)"></span> - 
                <span v-html="highlightKeyword(app.position_title)"></span>
              </h3>
              <a-tag :color="getStatusColor(app.status)">{{ app.status }}</a-tag>
            </div>
            
            <div class="timeline-content">
              <p><CalendarOutlined /> 投递日期: {{ formatDate(app.application_date) }}</p>
              <p v-if="app.salary_range"><DollarOutlined /> 薪资范围: {{ app.salary_range }}</p>
              <p v-if="app.work_location"><EnvironmentOutlined /> 工作地点: {{ app.work_location }}</p>
              <p v-if="app.notes"><FileTextOutlined /> 备注: {{ app.notes }}</p>
            </div>

            <div class="timeline-actions">
              <a-space>
                <a-button size="small" @click="editApplication(app)">编辑</a-button>
                <a-popconfirm title="确定要删除这条记录吗？" @confirm="deleteApp(app.id)">
                  <a-button size="small" danger>删除</a-button>
                </a-popconfirm>
              </a-space>
            </div>
          </div>
        </a-timeline-item>
      </a-timeline>

      <!-- 分页 -->
      <div v-if="filteredApplications.length > pageSize" class="pagination">
        <a-pagination
          v-model:current="currentPage"
          :total="filteredApplications.length"
          :page-size="pageSize"
          show-size-changer
          :page-size-options="['10', '20', '50', '100']"
          @change="handlePageChange"
          @showSizeChange="handlePageSizeChange"
        />
      </div>
    </a-card>

    <!-- 创建/编辑弹窗 -->
    <NewApplicationForm 
      v-model:visible="showCreateModal"
      :initial-data="editingApplication"
      @success="handleFormSuccess"
    />
    
    <!-- 批量导入弹窗 -->
    <BatchImport
      v-model:visible="showImportModal"
      @success="handleImportSuccess"
    />

    <!-- Excel导出弹窗 -->
    <ExportDialog
      v-model:visible="showExportModal"
      :applications="filteredApplications"
      :current-filters="currentFilters"
      @success="handleExportSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { 
  PlusOutlined, ReloadOutlined, CalendarOutlined, DollarOutlined, 
  EnvironmentOutlined, FileTextOutlined,
  CheckCircleOutlined, ClockCircleOutlined, ExclamationCircleOutlined,
  UploadOutlined, DownloadOutlined
} from '@ant-design/icons-vue'
import { useJobApplicationStore } from '../stores/jobApplication'
import { ApplicationStatus, StatusHelper, type JobApplication } from '../types'
import NewApplicationForm from '../components/NewApplicationForm.vue'
import BatchImport from '../components/BatchImport.vue'
import ExportDialog from '../components/ExportDialog.vue'
import FilterBar from '../components/FilterBar.vue'
import dayjs from 'dayjs'

const jobStore = useJobApplicationStore()
const { applications, loading } = storeToRefs(jobStore)

// 响应式数据
const showCreateModal = ref(false)
const showImportModal = ref(false)
const showExportModal = ref(false)
const editingApplication = ref<JobApplication | null>(null)
const currentPage = ref(1)
const pageSize = ref(20)
const currentFilters = ref<any>({})

// 筛选后的数据
const filteredApplications = computed(() => {
  let result = [...applications.value]
  
  // 关键词搜索
  if (currentFilters.value.keyword) {
    const keyword = currentFilters.value.keyword.toLowerCase()
    result = result.filter(app => 
      app.company_name.toLowerCase().includes(keyword) ||
      app.position_title.toLowerCase().includes(keyword)
    )
  }
  
  // 状态筛选
  if (currentFilters.value.status) {
    result = result.filter(app => app.status === currentFilters.value.status)
  }
  
  // 日期范围筛选
  if (currentFilters.value.dateRange) {
    const [start, end] = currentFilters.value.dateRange
    result = result.filter(app => {
      const appDate = dayjs(app.application_date)
      return appDate.isAfter(start) && appDate.isBefore(end.add(1, 'day'))
    })
  }
  
  // 薪资范围筛选
  if (currentFilters.value.salaryRange) {
    const [min, max] = currentFilters.value.salaryRange.split('-').map((v: string) => 
      v === '+' ? Infinity : parseInt(v)
    )
    result = result.filter(app => {
      if (!app.salary_range) return false
      const match = app.salary_range.match(/(\d+)/)
      if (match) {
        const salary = parseInt(match[1])
        if (max === Infinity) return salary >= min
        return salary >= min && salary < max
      }
      return false
    })
  }
  
  // 地点筛选
  if (currentFilters.value.location) {
    result = result.filter(app => 
      app.work_location?.includes(currentFilters.value.location)
    )
  }
  
  // 排序
  const sortBy = currentFilters.value.sortBy || 'date_desc'
  result.sort((a, b) => {
    switch (sortBy) {
      case 'date_asc':
        return new Date(a.application_date).getTime() - new Date(b.application_date).getTime()
      case 'date_desc':
        return new Date(b.application_date).getTime() - new Date(a.application_date).getTime()
      case 'salary_asc':
      case 'salary_desc':
        const getSalary = (app: JobApplication) => {
          const match = app.salary_range?.match(/(\d+)/)
          return match ? parseInt(match[1]) : 0
        }
        const diff = getSalary(a) - getSalary(b)
        return sortBy === 'salary_asc' ? diff : -diff
      default:
        return 0
    }
  })
  
  return result
})

// 分页后的数据
const paginatedApplications = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredApplications.value.slice(start, end)
})

// 是否有筛选条件
const hasFilters = computed(() => {
  return !!(
    currentFilters.value.keyword ||
    currentFilters.value.status ||
    currentFilters.value.dateRange ||
    currentFilters.value.salaryRange ||
    currentFilters.value.location
  )
})

// 统计数据
const writtenCount = computed(() => 
  filteredApplications.value.filter(app => app.status === ApplicationStatus.WRITTEN_TEST).length
)

const interviewingCount = computed(() => 
  filteredApplications.value.filter(app => {
    const interviewStatuses: ApplicationStatus[] = [
      ApplicationStatus.FIRST_INTERVIEW, 
      ApplicationStatus.SECOND_INTERVIEW, 
      ApplicationStatus.THIRD_INTERVIEW, 
      ApplicationStatus.HR_INTERVIEW
    ]
    return interviewStatuses.includes(app.status)
  }).length
)

const failedSumCount = computed(() => 
  filteredApplications.value.filter(app => {
    const failedStatuses: ApplicationStatus[] = [
      ApplicationStatus.RESUME_SCREENING_FAIL,
      ApplicationStatus.WRITTEN_TEST_FAIL,
      ApplicationStatus.FIRST_FAIL,
      ApplicationStatus.SECOND_FAIL,
      ApplicationStatus.THIRD_FAIL
    ]
    return failedStatuses.includes(app.status)
  }).length
)

const acceptedOfferCount = computed(() => 
  filteredApplications.value.filter(app => app.status === ApplicationStatus.OFFER_ACCEPTED).length
)

const totalApplicationsCount = computed(() => applications.value.length)

// 方法
const fetchData = () => jobStore.fetchApplications()

const formatDate = (date: string) => dayjs(date).format('YYYY-MM-DD')

const getStatusColor = (status: ApplicationStatus) => {
  return StatusHelper.getStatusColor(status)
}

const getStatusIcon = (status: ApplicationStatus) => {
  const successStatuses: ApplicationStatus[] = [
    ApplicationStatus.OFFER_RECEIVED, 
    ApplicationStatus.OFFER_ACCEPTED, 
    ApplicationStatus.PROCESS_FINISHED
  ]
  if (successStatuses.includes(status)) {
    return CheckCircleOutlined
  }
  if (status === ApplicationStatus.REJECTED) {
    return ExclamationCircleOutlined
  }
  return ClockCircleOutlined
}

// 高亮关键词
const highlightKeyword = (text: string) => {
  if (!currentFilters.value.keyword) return text
  const keyword = currentFilters.value.keyword
  const regex = new RegExp(`(${keyword})`, 'gi')
  return text.replace(regex, '<mark>$1</mark>')
}

// 处理筛选变化
const handleFilterChange = (filters: any) => {
  currentFilters.value = filters
  currentPage.value = 1 // 重置到第一页
}

// 清除筛选条件
const clearFilters = () => {
  currentFilters.value = {}
  currentPage.value = 1
}

// 分页处理
const handlePageChange = (page: number) => {
  currentPage.value = page
}

const handlePageSizeChange = (current: number, size: number) => {
  pageSize.value = size
  currentPage.value = 1
}

const editApplication = (app: JobApplication) => {
  editingApplication.value = app
  showCreateModal.value = true
}

const deleteApp = async (id: number) => {
  await jobStore.deleteApplication(id)
}

const handleFormSuccess = () => {
  showCreateModal.value = false
  editingApplication.value = null
  fetchData()
}

const handleImportSuccess = () => {
  showImportModal.value = false
  fetchData()
}

const handleExportSuccess = () => {
  showExportModal.value = false
  // 导出成功后可以显示成功提示，但不需要刷新数据
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.timeline-page {
  padding: 24px;
  background: #f0f2f5;
  min-height: calc(100vh - 48px - 56px - 70px);
}

.stats-row {
  margin-bottom: 16px;
}

.timeline-item {
  padding: 12px 0;
}

/* 统计行：五等分一行展示 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(120px, 1fr));
  gap: 16px;
  align-items: center;
}

.stats-grid :deep(.ant-statistic-title) {
  color: #8c8c8c;
}

.stats-grid :deep(.ant-statistic-content) {
  font-weight: 600;
}

.timeline-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.timeline-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 500;
}

.timeline-content {
  margin-bottom: 12px;
  color: #666;
}

.timeline-content p {
  margin: 4px 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.timeline-actions {
  margin-top: 8px;
}

.empty-state {
  text-align: center;
  padding: 40px 0;
}

.pagination {
  margin-top: 24px;
  text-align: center;
}

/* 搜索关键词高亮 */
.timeline-header :deep(mark) {
  background-color: #ffe58f;
  padding: 0 2px;
  border-radius: 2px;
}
</style>
