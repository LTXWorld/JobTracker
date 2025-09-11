<template>
  <div class="kanban-board">
    <!-- 看板头部 -->
    <div class="kanban-header">
      <h2>求职看板</h2>
      <div class="header-right">
        <!-- 搜索定位 -->
        <div class="kanban-search">
          <a-auto-complete
            v-model:value="searchText"
            :options="searchOptions"
            :filter-option="false"
            style="width: 320px"
            @select="onSearchSelect"
          >
            <a-input
              :placeholder="'搜索公司/职位... (回车定位)'"
              allow-clear
              @pressEnter="onSearchEnter"
            >
              <template #prefix>
                <SearchOutlined />
              </template>
            </a-input>
          </a-auto-complete>
        </div>
        <!-- 状态切换标签 -->
        <div class="status-tabs">
          <a-button 
            :type="activeTab === 'in-progress' ? 'primary' : 'default'"
            @click="setActiveTab('in-progress')"
            class="tab-button"
          >
            <ClockCircleOutlined />
            <div class="tab-content">
              <div class="tab-title">进行中</div>
              <div class="tab-count">{{ inProgressCount }}</div>
            </div>
          </a-button>
          <a-button 
            :type="activeTab === 'failed' ? 'primary' : 'default'"
            @click="setActiveTab('failed')"
            class="tab-button"
          >
            <CloseCircleOutlined />
            <div class="tab-content">
              <div class="tab-title">失败状态</div>
              <div class="tab-count">{{ failedCount }}</div>
            </div>
          </a-button>
        </div>
        
        <!-- 操作按钮 -->
        <div class="kanban-actions">
          <a-button type="primary" @click="showCreateModal = true">
            <template #icon><PlusOutlined /></template>
            添加投递
          </a-button>
          <a-button @click="showImportModal = true">
            <template #icon><UploadOutlined /></template>
            批量导入
          </a-button>
          <a-button @click="showExportModal = true">
            <template #icon><DownloadOutlined /></template>
            导出Excel
          </a-button>
          <a-button @click="fetchData">
            <template #icon><ReloadOutlined /></template>
            刷新
          </a-button>
        </div>
      </div>
    </div>

    <!-- 看板主体 -->
    <div class="kanban-main" v-if="!loading">
      <div class="kanban-columns">
        <div class="kanban-column" v-for="column in currentStatusColumns" :key="column.status">
          <div class="column-header">
            <h3>{{ column.title }}</h3>
            <a-badge :count="column.items.length" :color="column.color" />
          </div>
          
          <!-- 拖拽容器 -->
          <draggable
            v-model="column.items"
            group="applications"
            item-key="id"
            class="column-content"
            :animation="200"
            ghost-class="ghost-card"
            @change="handleDragChange($event, column.status)"
          >
            <template #item="{ element }">
              <div 
                class="job-card"
                :class="{ 'bounce': highlightedId === element.id }"
                :tabindex="0"
                :data-id="element.id"
                :ref="(el) => registerCardRef(element.id, el as HTMLElement | null)"
                @animationend="onBounceEnd(element.id)"
                @click="openStatusDetail(element)"
              >
                <div class="card-header">
                  <h4 v-html="highlightText(element.company_name)"></h4>
                  <a-dropdown @click.stop>
                    <template #overlay>
                      <a-menu @click="handleCardAction($event, element)">
                        <a-menu-item key="status-detail">
                          <HistoryOutlined /> 状态详情
                        </a-menu-item>
                        <a-menu-item key="quick-update">
                          <EditOutlined /> 快速更新
                        </a-menu-item>
                        <a-menu-divider />
                        <a-menu-item key="edit">
                          <SettingOutlined /> 编辑
                        </a-menu-item>
                        <a-menu-item key="delete">
                          <DeleteOutlined /> 删除
                        </a-menu-item>
                      </a-menu>
                    </template>
                    <a-button type="text" size="small">
                      <MoreOutlined />
                    </a-button>
                  </a-dropdown>
                </div>
                
                <div class="card-body">
                  <p class="position" v-html="highlightText(element.position_title)"></p>
                  
                  <!-- 状态持续时间指示器 -->
                  <div class="status-duration" v-if="getStatusDuration(element)">
                    <ClockCircleOutlined />
                    <span>{{ getStatusDuration(element) }}</span>
                  </div>
                  
                  <div class="card-date">
                    <CalendarOutlined /> {{ formatDate(element.application_date) }}
                  </div>
                  
                  <!-- 进度指示器 -->
                  <div class="progress-indicator">
                    <a-progress 
                      :percent="getProgressPercent(element.status)" 
                      size="small" 
                      :show-info="false"
                      :stroke-color="getProgressColor(element.status)"
                    />
                  </div>
                </div>
              </div>
            </template>
          </draggable>
          
          <!-- 空状态 -->
          <div v-if="column.items.length === 0" class="empty-column">
            <a-empty :description="`暂无${column.title}的投递`" />
          </div>
        </div>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-else class="loading-container">
      <a-spin size="large" tip="加载中..." />
    </div>

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

    <!-- 状态详情弹窗 -->
    <StatusDetailModal
      v-model:visible="showStatusDetailModal"
      :application-id="selectedApplicationId"
      :current-status="selectedApplicationStatus"
      @status-updated="handleStatusUpdated"
    />

    <!-- 状态快速更新弹窗 -->
    <StatusQuickUpdate
      v-if="showQuickUpdateModal && selectedApplication"
      :application-id="selectedApplication.id"
      :current-status="selectedApplication.status"
      mode="button"
      @updated="handleStatusUpdated"
      @cancelled="showQuickUpdateModal = false"
    />

    <!-- Excel导出弹窗 -->
    <ExportDialog
      v-model:visible="showExportModal"
      :applications="applications"
      @success="handleExportSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, h } from 'vue'
import { storeToRefs } from 'pinia'
import draggable from 'vuedraggable'
import { 
  PlusOutlined, ReloadOutlined, CalendarOutlined, DollarOutlined, 
  EnvironmentOutlined, MoreOutlined, EditOutlined, DeleteOutlined,
  UploadOutlined, BellOutlined, BellFilled, DownOutlined, UpOutlined,
  ClockCircleOutlined, TrophyOutlined, CloseCircleOutlined,
  SearchOutlined,
  HistoryOutlined, SettingOutlined, DownloadOutlined
} from '@ant-design/icons-vue'
import { useJobApplicationStore } from '../stores/jobApplication'
import { useStatusTrackingStore } from '../stores/statusTracking'
import { ApplicationStatus, StatusHelper, type JobApplication, type ApplicationStatus as AppStatus } from '../types'
import NewApplicationForm from '../components/NewApplicationForm.vue'
import BatchImport from '../components/BatchImport.vue'
import ExportDialog from '../components/ExportDialog.vue'
import StatusDetailModal from '../components/StatusDetailModal.vue'
import StatusQuickUpdate from '../components/StatusQuickUpdate.vue'
import dayjs from 'dayjs'
import { message, Modal } from 'ant-design-vue'

interface KanbanColumn {
  status: ApplicationStatus
  title: string
  color: string
  items: JobApplication[]
}

const jobStore = useJobApplicationStore()
const statusTrackingStore = useStatusTrackingStore()
const { applications, loading } = storeToRefs(jobStore)

const showCreateModal = ref(false)
const showImportModal = ref(false)
const showExportModal = ref(false)
const showStatusDetailModal = ref(false)
const showQuickUpdateModal = ref(false)
const editingApplication = ref<JobApplication | null>(null)
const selectedApplication = ref<JobApplication | null>(null)
const selectedApplicationId = ref<number>(0)
const selectedApplicationStatus = ref<AppStatus>('已投递')

// 当前活跃的标签页
const activeTab = ref<'in-progress' | 'failed'>('in-progress')

// 搜索相关
const searchText = ref('')
const highlightedId = ref<number | null>(null)
const cardRefs = new Map<number, HTMLElement>()

const failedStatusSet = new Set([
  ApplicationStatus.RESUME_SCREENING_FAIL,
  ApplicationStatus.WRITTEN_TEST_FAIL,
  ApplicationStatus.FIRST_FAIL,
  ApplicationStatus.SECOND_FAIL,
  ApplicationStatus.THIRD_FAIL,
  ApplicationStatus.HR_FAIL,
  ApplicationStatus.REJECTED
])

const escapeRegExp = (s: string) => s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

const renderHighlightedLabel = (label: string, keyword: string) => {
  if (!keyword) return label
  const rx = new RegExp(escapeRegExp(keyword), 'ig')
  const nodes: any[] = []
  let lastIndex = 0
  let match: RegExpExecArray | null
  // 为了性能，限制最多渲染10个高亮片段
  let segments = 0
  while ((match = rx.exec(label)) && segments < 10) {
    const start = match.index
    const end = start + match[0].length
    if (start > lastIndex) nodes.push(label.slice(lastIndex, start))
    nodes.push(h('mark', match[0]))
    lastIndex = end
    segments++
  }
  if (lastIndex < label.length) nodes.push(label.slice(lastIndex))
  return h('span', nodes)
}

const highlightText = (text: string) => {
  const kw = searchText.value.trim()
  if (!kw) return text
  const safe = (s: string) => s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
  const rx = new RegExp(escapeRegExp(kw), 'ig')
  return safe(text).replace(rx, (m) => `<mark>${m}</mark>`)
}

const searchOptions = computed(() => {
  const list = Array.isArray(applications.value) ? applications.value : []
  const keyword = searchText.value.trim().toLowerCase()
  // 生成选项并按关键字过滤（前端过滤，避免额外请求）
  const items = list.map(app => {
    const rawLabel = `${app.company_name} - ${app.position_title}`
    return {
      value: String(app.id),
      label: renderHighlightedLabel(rawLabel, searchText.value.trim()) as any,
      _plain: rawLabel.toLowerCase(),
    }
  })
  if (!keyword) return items.slice(0, 50)
  return items.filter(o => o._plain.includes(keyword)).slice(0, 50)
})

const onSearchEnter = async () => {
  const opt = searchOptions.value.find(o => o.label.toLowerCase().includes(searchText.value.trim().toLowerCase()))
  if (opt) {
    await locateCardById(Number(opt.value))
  } else if (searchText.value.trim()) {
    message.warning('未找到匹配的投递记录')
  }
}

const onSearchSelect = async (value: string) => {
  await locateCardById(Number(value))
}

const registerCardRef = (id: number, el: HTMLElement | null) => {
  if (el) {
    cardRefs.set(id, el)
  } else {
    cardRefs.delete(id)
  }
}

const onBounceEnd = (id: number) => {
  if (highlightedId.value === id) {
    highlightedId.value = null
  }
}

const locateCardById = async (id: number) => {
  const list = Array.isArray(applications.value) ? applications.value : []
  const app = list.find(a => a.id === id)
  if (!app) {
    message.warning('记录不存在或未加载')
    return
  }

  // 切换到对应分组标签（进行中/失败），以确保卡片渲染出来
  const targetTab: 'in-progress' | 'failed' = failedStatusSet.has(app.status) ? 'failed' : 'in-progress'
  if (activeTab.value !== targetTab) {
    activeTab.value = targetTab
    await nextTick()
  }

  // 等待DOM渲染并获取卡片引用
  await nextTick()
  let el = cardRefs.get(id)
  if (!el) {
    // 再次等待一次（拖拽列表渲染可能稍有延迟）
    await new Promise(r => setTimeout(r, 50))
    el = cardRefs.get(id)
  }
  if (!el) {
    message.warning('未能定位到卡片')
    return
  }

  // 滚动并聚焦
  el.scrollIntoView({ behavior: 'smooth', block: 'center', inline: 'center' })
  // 短暂延时，待滚动生效后再聚焦与动画
  setTimeout(() => {
    try { el?.focus({ preventScroll: true }) } catch {}
    highlightedId.value = id
  }, 220)
}

// 定义进行中状态列
const inProgressColumns = [
  { status: ApplicationStatus.APPLIED, title: '已投递', color: '#1890ff' },
  { status: ApplicationStatus.RESUME_SCREENING, title: '简历筛选中', color: '#13c2c2' },
  { status: ApplicationStatus.WRITTEN_TEST, title: '笔试中', color: '#fa8c16' },
  { status: ApplicationStatus.FIRST_INTERVIEW, title: '一面中', color: '#722ed1' },
  { status: ApplicationStatus.SECOND_INTERVIEW, title: '二面中', color: '#eb2f96' },
  { status: ApplicationStatus.THIRD_INTERVIEW, title: '三面中', color: '#13c2c2' },
  { status: ApplicationStatus.HR_INTERVIEW, title: 'HR面中', color: '#fa541c' }
]

// 定义失败状态列（与统计页一致：包含已拒绝、HR面未通过）
const failedColumns = [
  { status: ApplicationStatus.RESUME_SCREENING_FAIL, title: '简历挂', color: '#ff7875' },
  { status: ApplicationStatus.WRITTEN_TEST_FAIL, title: '笔试挂', color: '#ff7875' },
  { status: ApplicationStatus.FIRST_FAIL, title: '一面挂', color: '#ff7875' },
  { status: ApplicationStatus.SECOND_FAIL, title: '二面挂', color: '#ff7875' },
  { status: ApplicationStatus.THIRD_FAIL, title: '三面挂', color: '#ff7875' },
  { status: ApplicationStatus.HR_FAIL, title: 'HR面挂', color: '#ff7875' },
  { status: ApplicationStatus.REJECTED, title: '已拒绝', color: '#ff4d4f' }
]

// 切换标签页
const setActiveTab = (tab: 'in-progress' | 'failed') => {
  activeTab.value = tab
}

// 获取当前活跃状态的列
const currentStatusColumns = computed(() => {
  const columns = activeTab.value === 'in-progress' ? inProgressColumns : failedColumns
  return columns.map(col => ({
    ...col,
    items: Array.isArray(applications.value) ? applications.value.filter(app => app.status === col.status) : []
  }))
})

// 计算进行中状态数量
const inProgressCount = computed(() => {
  if (!Array.isArray(applications.value)) return 0
  return inProgressColumns.reduce((total, col) => {
    return total + applications.value.filter(app => app.status === col.status).length
  }, 0)
})

// 计算失败状态数量（包含 HR面未通过、已拒绝）  
const failedCount = computed(() => {
  if (!Array.isArray(applications.value)) return 0
  return failedColumns.reduce((total, col) => {
    return total + applications.value.filter(app => app.status === col.status).length
  }, 0)
})

// 保留原有的kanbanColumns计算属性以兼容其他组件
const kanbanColumns = computed((): KanbanColumn[] => {
  const allColumns = [...inProgressColumns, ...failedColumns]
  return allColumns.map(col => ({
    ...col,
    items: Array.isArray(applications.value) ? applications.value.filter(app => app.status === col.status) : []
  }))
})

// 格式化日期
const formatDate = (date: string) => dayjs(date).format('MM-DD')

// 格式化日期时间
const formatDateTime = (datetime: string) => dayjs(datetime).format('MM-DD HH:mm')

// 获取数据
const fetchData = () => jobStore.fetchApplications()

// 处理拖拽变化
const handleDragChange = async (evt: any, newStatus: ApplicationStatus) => {
  if (!evt.added) return
  const app = evt.added.element as JobApplication

  // 预检合法流转（若配置了限制，则前端先挡住无效拖拽）
  try {
    const rules = await statusTrackingStore.getAvailableTransitions(app.status)
    const allowed = rules?.some(r => (r.to || []).includes(newStatus)) ?? true
    if (!allowed) {
      message.warning(`不允许从「${app.status}」到「${newStatus}」`)
      await fetchData()
      return
    }
  } catch {
    // 若获取失败，不阻断操作，交由后端校验
  }

  try {
    // 使用状态跟踪接口以保证合法流转并记录历史
    await statusTrackingStore.updateApplicationStatus(app.id, { status: newStatus })
    message.success(`已更新状态为: ${newStatus}`)
    // 刷新应用列表，确保列分组一致
    await fetchData()
  } catch (error: any) {
    const msg = (error?.message as string) || '状态更新失败'
    message.error(msg)
    // 恢复原数据
    await fetchData()
  }
}

// 处理卡片操作
const handleCardAction = ({ key }: { key: string }, app: JobApplication) => {
  if (key === 'status-detail') {
    openStatusDetail(app)
  } else if (key === 'quick-update') {
    selectedApplication.value = app
    showQuickUpdateModal.value = true
  } else if (key === 'edit') {
    editingApplication.value = app
    showCreateModal.value = true
  } else if (key === 'delete') {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除 ${app.company_name} - ${app.position_title} 的投递记录吗？`,
      onOk: async () => {
        await jobStore.deleteApplication(app.id)
      }
    })
  }
}

// 打开状态详情
const openStatusDetail = (app: JobApplication) => {
  selectedApplicationId.value = app.id
  selectedApplicationStatus.value = app.status
  showStatusDetailModal.value = true
}

// 获取状态持续时间
const getStatusDuration = (app: JobApplication): string => {
  const updatedTime = dayjs(app.updated_at)
  const now = dayjs()
  const duration = now.diff(updatedTime, 'day')
  
  if (duration === 0) {
    const hours = now.diff(updatedTime, 'hour')
    return hours > 0 ? `${hours}小时` : '刚刚'
  } else if (duration < 30) {
    return `${duration}天`
  } else {
    return '超过1月'
  }
}

// 获取进度百分比
const getProgressPercent = (status: AppStatus): number => {
  const progressMap: Record<string, number> = {
    '已投递': 10,
    '简历筛选中': 20,
    '简历筛选未通过': 0,
    '笔试中': 30,
    '笔试通过': 40,
    '笔试未通过': 0,
    '一面中': 50,
    '一面通过': 60,
    '一面未通过': 0,
    '二面中': 70,
    '二面通过': 80,
    '二面未通过': 0,
    '三面中': 85,
    '三面通过': 90,
    '三面未通过': 0,
    'HR面中': 95,
    'HR面通过': 98,
    'HR面未通过': 0,
    '待发offer': 99,
    '已拒绝': 0,
    '已收到offer': 100,
    '已接受offer': 100,
    '流程结束': 100
  }
  return progressMap[status] || 0
}

// 获取进度条颜色
const getProgressColor = (status: AppStatus): string => {
  if (StatusHelper.isFailedStatus(status)) return '#ff4d4f'
  if (StatusHelper.isPassedStatus(status)) return '#52c41a'
  return '#1890ff'
}

// 处理状态更新成功
const handleStatusUpdated = (newStatus: AppStatus) => {
  showStatusDetailModal.value = false
  showQuickUpdateModal.value = false
  selectedApplication.value = null
  fetchData() // 刷新数据以反映状态变更
}

// 表单成功回调
const handleFormSuccess = () => {
  showCreateModal.value = false
  editingApplication.value = null
  fetchData()
}

const handleImportSuccess = () => {
  showImportModal.value = false
  fetchData()
  message.success('批量导入成功')
}

const handleExportSuccess = () => {
  showExportModal.value = false
  // 导出成功后不需要刷新数据
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.kanban-board {
  height: calc(100vh - 48px - 56px - 70px);
  padding: 24px;
  background: #f0f2f5;
}

.kanban-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.kanban-search {
  display: flex;
  align-items: center;
}

.kanban-header h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 24px;
}

.kanban-actions {
  display: flex;
  gap: 12px;
}

/* 看板主容器 */
.kanban-main {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 48px - 56px - 100px);
  overflow: hidden;
}

/* 状态切换标签 - 头部水平布局 */
.status-tabs {
  display: flex;
  gap: 12px;
}

.tab-button {
  height: 44px !important;
  padding: 8px 16px !important;
  border-radius: 8px !important;
  font-weight: 500;
  display: flex !important;
  align-items: center !important;
  gap: 8px !important;
  transition: all 0.3s ease;
  min-width: 100px;
}

.tab-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.tab-content {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  line-height: 1.2;
}

.tab-title {
  font-size: 12px;
  font-weight: 600;
}

.tab-count {
  font-size: 14px;
  font-weight: 700;
  margin-top: 1px;
}

/* 看板列容器 */
.kanban-columns {
  display: flex;
  gap: 18px;
  flex: 1;
  overflow-x: auto;
  overflow-y: hidden;
  padding: 0 4px;
}

.kanban-column {
  flex: 0 0 240px;
  background: #fff;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.06);
  border: 1px solid #f0f0f0;
  height: calc(100vh - 180px);
}

.column-header {
  padding: 12px 16px; /* 从18px 20px减少到12px 16px */
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fafafa;
  border-radius: 8px 8px 0 0;
}

.column-header h3 {
  margin: 0;
  font-size: 14px; /* 从16px减少到14px */
  font-weight: 600;
  color: #262626;
}

.column-content {
  flex: 1;
  padding: 12px; /* 从16px减少到12px */
  overflow-y: auto;
  min-height: 200px;
}

.job-card {
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 6px;
  padding: 8px;
  margin-bottom: 8px;
  cursor: move;
  transition: all 0.3s ease;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
  min-height: 70px;
  display: flex;
  flex-direction: column;
  position: relative;
}

.job-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
  border-color: #1890ff;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 10px; /* 从14px减少到10px */
}

.card-header h4 {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: #262626;
  flex: 1;
  line-height: 1.3;
  padding-right: 6px;
}

.card-body {
  flex: 1;
  margin-bottom: 8px; /* 从12px减少到8px */
}

.position {
  margin: 0 0 6px 0;
  font-size: 12px;
  color: #1890ff;
  font-weight: 500;
  line-height: 1.3;
}

.card-info {
  display: flex;
  flex-direction: column;
  gap: 4px; /* 从8px减少到4px */
  margin-bottom: 8px; /* 从12px减少到8px */
  font-size: 12px; /* 从14px减少到12px */
  color: #666;
}

.card-info span {
  display: flex;
  align-items: center;
  gap: 4px; /* 从8px减少到4px */
  padding: 1px 0; /* 从2px减少到1px */
}

.card-date {
  font-size: 10px;
  color: #999;
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 4px;
}

.interview-info {
  font-size: 11px; /* 从13px减少到11px */
  color: #ff4d4f;
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 6px;
  padding: 4px 8px; /* 从6px 10px减少到4px 8px */
  background: linear-gradient(135deg, #fff2f0 0%, #ffebe8 100%);
  border-radius: 4px;
  border: 1px solid #ffccc7;
  font-weight: 500;
}

.card-footer {
  padding-top: 10px;
  border-top: 1px solid #f0f0f0;
  margin-top: 8px;
}

.notes {
  margin: 0;
  font-size: 12px;
  color: #8c8c8c;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

/* 不同分组的特殊样式 */
.group-in-progress .kanban-column {
  border-top: 3px solid #1890ff;
}

.group-success .kanban-column {
  border-top: 3px solid #52c41a;
}

.group-failed .kanban-column {
  border-top: 3px solid #ff4d4f;
}

/* 折叠状态下的汇总信息 */
.group-collapsed-summary {
  padding: 12px 20px;
  background: #fafafa;
  border-top: 1px solid #f0f0f0;
}

.summary-stats {
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
}

.summary-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #666;
}

.ghost-card {
  opacity: 0.5;
  background: #f0f2f5;
}

.empty-column {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.loading-container {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .kanban-main {
    height: calc(100vh - 48px - 56px - 120px);
  }
  
  .kanban-columns {
    gap: 16px;
  }
  
  .kanban-column {
    flex: 0 0 220px;
    height: calc(100vh - 200px);
  }
  
  .status-tabs {
    width: 140px;
  }
  
  .tab-button {
    height: 70px !important;
  }
}

@media (max-width: 992px) {
  .kanban-main {
    flex-direction: column;
    height: calc(100vh - 48px - 56px - 140px);
  }
  
  .status-tabs {
    width: 100%;
    flex-direction: row;
    justify-content: center;
    order: -1;
    margin-bottom: 16px;
  }
  
  .tab-button {
    height: 50px !important;
    flex: 1;
    max-width: 200px;
  }
  
  .tab-content {
    align-items: center;
  }
  
  .kanban-columns {
    gap: 12px;
  }
  
  .kanban-column {
    flex: 0 0 200px;
    height: calc(100vh - 260px);
  }
}

@media (max-width: 768px) {
  .kanban-board {
    padding: 16px;
  }
  
  .kanban-main {
    height: calc(100vh - 48px - 56px - 160px);
  }
  
  .kanban-columns {
    gap: 12px;
    padding: 0;
  }
  
  .kanban-column {
    flex: 0 0 180px;
    height: calc(100vh - 280px);
  }
  
  .job-card {
    padding: 6px;
    margin-bottom: 6px;
    min-height: 60px;
  }
  
  .card-header h4 {
    font-size: 12px;
  }
  
  .position {
    font-size: 11px;
    margin: 0 0 4px 0;
  }
  
  .card-date {
    font-size: 9px;
    margin-top: 3px;
  }
}

@media (max-width: 480px) {
  .kanban-columns {
    justify-content: flex-start;
    padding-right: 16px;
  }
  
  .kanban-column {
    flex: 0 0 160px;
  }
  
  .status-tabs {
    gap: 8px;
  }
  
  .tab-button {
    height: 40px !important;
    padding: 8px 12px !important;
  }
  
  .tab-title {
    font-size: 12px;
  }
  
  .tab-count {
    font-size: 14px;
  }
}

/* 滚动条样式 */
.kanban-container::-webkit-scrollbar,
.column-content::-webkit-scrollbar {
  height: 8px;
  width: 8px;
}

.kanban-container::-webkit-scrollbar-track,
.column-content::-webkit-scrollbar-track {
  background: #f0f0f0;
  border-radius: 4px;
}

.kanban-container::-webkit-scrollbar-thumb,
.column-content::-webkit-scrollbar-thumb {
  background: #d9d9d9;
  border-radius: 4px;
}

.kanban-container::-webkit-scrollbar-thumb:hover,
.column-content::-webkit-scrollbar-thumb:hover {
  background: #bfbfbf;
}

/* 状态跟踪相关样式 */
.status-duration {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: #8c8c8c;
  margin-bottom: 6px;
}

.status-duration .anticon {
  font-size: 10px;
}

.progress-indicator {
  margin-top: 8px;
}

.card-date {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: #8c8c8c;
  margin-bottom: 8px;
}

.card-date .anticon {
  font-size: 10px;
}

/* 增强卡片交互效果 */
.job-card {
  position: relative;
}

/* 聚焦/定位视觉强化 */
.job-card:focus {
  outline: none;
  box-shadow: 0 0 0 3px rgba(24, 144, 255, 0.35);
}

/* 弹跳动画 */
@keyframes bounceOnce {
  0%, 20%, 53%, 80%, 100% { transform: translate3d(0,0,0); }
  40%, 43% { transform: translate3d(0, -10px, 0); }
  70% { transform: translate3d(0, -6px, 0); }
  90% { transform: translate3d(0, -2px, 0); }
}
.bounce {
  animation: bounceOnce 0.9s ease;
}

.job-card::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 3px;
  height: 100%;
  background: transparent;
  border-radius: 0 8px 8px 0;
  transition: all 0.2s ease;
}

.job-card:hover::after {
  background: #1890ff;
}

/* 状态持续时间警告色 */
.status-duration.warning {
  color: #faad14;
}

.status-duration.danger {
  color: #ff4d4f;
}

/* 进度条样式调整 */
.progress-indicator :deep(.ant-progress-bg) {
  border-radius: 2px;
}

.progress-indicator :deep(.ant-progress-outer) {
  padding-right: 0;
}

/* 搜索匹配高亮样式 */
.kanban-board :deep(mark) {
  background: #ffe58f;
  color: inherit;
  padding: 0 2px;
  border-radius: 2px;
}
</style>
