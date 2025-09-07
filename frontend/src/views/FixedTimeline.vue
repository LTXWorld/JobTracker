<template>
  <div class="timeline-page">
    <a-card title="投递记录 Timeline" :loading="loading">
      <template #extra>
        <a-space>
          <a-button type="primary" @click="showCreateModal = true">
            <template #icon><PlusOutlined /></template>
            添加投递
          </a-button>
          <a-button @click="fetchData">
            <template #icon><ReloadOutlined /></template>
            刷新
          </a-button>
        </a-space>
      </template>

      <!-- 快速统计 -->
      <div class="stats-row" v-if="applications.length > 0">
        <a-row :gutter="16">
          <a-col :xs="12" :sm="6">
            <a-statistic title="总投递数" :value="totalCount" />
          </a-col>
          <a-col :xs="12" :sm="6">
            <a-statistic title="面试中" :value="interviewingCount" />
          </a-col>
          <a-col :xs="12" :sm="6">
            <a-statistic title="已拒绝" :value="rejectedCount" />
          </a-col>
          <a-col :xs="12" :sm="6">
            <a-statistic title="已收offer" :value="offerCount" />
          </a-col>
        </a-row>
      </div>

      <a-divider />

      <!-- Timeline展示 -->
      <div v-if="applications.length === 0 && !loading" class="empty-state">
        <a-empty description="暂无投递记录">
          <a-button type="primary" @click="showCreateModal = true">添加第一条记录</a-button>
        </a-empty>
      </div>

      <a-timeline v-else>
        <a-timeline-item 
          v-for="app in applications" 
          :key="app.id"
          :color="getStatusColor(app.status)"
        >
          <template #dot>
            <component :is="getStatusIcon(app.status)" />
          </template>
          
          <div class="timeline-content">
            <div class="timeline-header">
              <h3>{{ app.company_name }} - {{ app.position_title }}</h3>
              <a-tag :color="getStatusColor(app.status)">{{ app.status }}</a-tag>
            </div>
            
            <div class="timeline-details">
              <div class="detail-item">
                <CalendarOutlined />
                <span>投递时间: {{ formatDate(app.application_date) }}</span>
              </div>
              
              <div class="detail-item" v-if="app.salary_range">
                <DollarOutlined />
                <span>薪资范围: {{ app.salary_range }}</span>
              </div>
              
              <div class="detail-item" v-if="app.work_location">
                <EnvironmentOutlined />
                <span>工作地点: {{ app.work_location }}</span>
              </div>
              
              <div class="detail-item" v-if="app.notes">
                <FileTextOutlined />
                <span>备注: {{ app.notes }}</span>
              </div>
            </div>
            
            <div class="timeline-actions">
              <a-dropdown>
                <template #overlay>
                  <a-menu @click="(e: any) => handleStatusChange(app.id, e.key)">
                    <a-menu-item key="已投递">已投递</a-menu-item>
                    <a-menu-item key="笔试中">笔试中</a-menu-item>
                    <a-menu-item key="一面中">一面中</a-menu-item>
                    <a-menu-item key="二面中">二面中</a-menu-item>
                    <a-menu-item key="三面中">三面中</a-menu-item>
                    <a-menu-item key="已挂">已挂</a-menu-item>
                  </a-menu>
                </template>
                <a-button>
                  状态变更 <DownOutlined />
                </a-button>
              </a-dropdown>
              
              <a-button @click="handleEdit(app)">编辑</a-button>
              <a-popconfirm 
                title="确定要删除这条记录吗？" 
                @confirm="handleDelete(app.id)"
                ok-text="确定" 
                cancel-text="取消"
              >
                <a-button danger>删除</a-button>
              </a-popconfirm>
            </div>
          </div>
        </a-timeline-item>
      </a-timeline>
    </a-card>

    <!-- 创建/编辑弹窗 -->
    <ApplicationForm
      v-model:visible="showCreateModal"
      :initial-data="editingApplication"
      @success="handleFormSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { storeToRefs } from 'pinia' // 重要：导入storeToRefs
import { 
  PlusOutlined, ReloadOutlined, CalendarOutlined, DollarOutlined, 
  EnvironmentOutlined, FileTextOutlined, DownOutlined,
  CheckCircleOutlined, ClockCircleOutlined, ExclamationCircleOutlined
} from '@ant-design/icons-vue'
import { useJobApplicationStore } from '../stores/jobApplication'
import { ApplicationStatus, StatusHelper, type JobApplication } from '../types'
import ApplicationForm from '../components/ApplicationForm.vue'
import dayjs from 'dayjs'

const jobStore = useJobApplicationStore()

// 响应式数据
const showCreateModal = ref(false)
const editingApplication = ref<JobApplication | null>(null)

// 正确的方式：使用storeToRefs保持响应性
const { applications, loading, totalCount } = storeToRefs(jobStore)

const interviewingCount = computed(() => 
  applications.value.filter(app => {
    const interviewStatuses: ApplicationStatus[] = [
      ApplicationStatus.FIRST_INTERVIEW, 
      ApplicationStatus.SECOND_INTERVIEW, 
      ApplicationStatus.THIRD_INTERVIEW, 
      ApplicationStatus.HR_INTERVIEW
    ]
    return interviewStatuses.includes(app.status)
  }).length
)
const rejectedCount = computed(() => applications.value.filter(app => app.status === ApplicationStatus.REJECTED).length)
const offerCount = computed(() => applications.value.filter(app => {
  const offerStatuses: ApplicationStatus[] = [
    ApplicationStatus.OFFER_RECEIVED, 
    ApplicationStatus.OFFER_ACCEPTED
  ]
  return offerStatuses.includes(app.status)
}).length)

// 方法
const fetchData = () => jobStore.fetchApplications()

const formatDate = (date: string) => dayjs(date).format('YYYY-MM-DD')

const getStatusColor = (status: ApplicationStatus) => {
  return StatusHelper.getStatusColor(status)
}

const getStatusIcon = (status: ApplicationStatus) => {
  if (status === ApplicationStatus.REJECTED) {
    return ExclamationCircleOutlined
  } else if ([ApplicationStatus.OFFER_RECEIVED, ApplicationStatus.OFFER_ACCEPTED].includes(status)) {
    return CheckCircleOutlined
  } else {
    return ClockCircleOutlined
  }
}

const handleStatusChange = async (id: number, newStatus: ApplicationStatus) => {
  try {
    await jobStore.updateApplication(id, { status: newStatus })
  } catch (error) {
    console.error('状态更新失败:', error)
  }
}

const handleEdit = (application: JobApplication) => {
  editingApplication.value = application
  showCreateModal.value = true
}

const handleDelete = async (id: number) => {
  try {
    await jobStore.deleteApplication(id)
  } catch (error) {
    console.error('删除失败:', error)
  }
}

const handleFormSuccess = () => {
  showCreateModal.value = false
  editingApplication.value = null
  fetchData()
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.timeline-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px;
}

.stats-row {
  margin-bottom: 24px;
}

.empty-state {
  text-align: center;
  margin: 48px 0;
}

.timeline-content {
  min-width: 0;
  flex: 1;
}

.timeline-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.timeline-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  flex: 1;
}

.timeline-details {
  margin-bottom: 16px;
}

.detail-item {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
  color: #666;
}

.detail-item :deep(.anticon) {
  margin-right: 8px;
  color: #999;
}

.timeline-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

@media (max-width: 768px) {
  .timeline-page {
    padding: 16px;
  }
  
  .timeline-header {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .timeline-header h3 {
    margin-bottom: 8px;
  }
  
  .timeline-actions {
    margin-top: 12px;
  }
}
</style>