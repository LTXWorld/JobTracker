<template>
  <div class="simple-kanban">
    <div class="kanban-header">
      <h2>求职看板 (简化版)</h2>
      <div class="kanban-actions">
        <a-button type="primary">
          <template #icon><PlusOutlined /></template>
          添加投递
        </a-button>
        <a-button>
          <template #icon><ReloadOutlined /></template>
          刷新
        </a-button>
      </div>
    </div>

    <!-- 测试数据显示 -->
    <div class="test-info">
      <p>应用数据加载状态: {{ loading ? '加载中...' : '已加载' }}</p>
      <p>总记录数: {{ applications.length }}</p>
      <p>看板分组数量: {{ kanbanGroups.length }}</p>
    </div>

    <!-- 分组展示 -->
    <div class="groups-container" v-if="!loading">
      <div class="group-section" v-for="group in kanbanGroups" :key="group.key">
        <div class="group-title">
          <h3>{{ group.title }} ({{ getGroupTotalCount(group) }})</h3>
          <a-button 
            v-if="group.collapsible"
            type="text" 
            size="small"
            @click="toggleGroupCollapse(group.key)"
          >
            {{ groupCollapsedStates[group.key] ? '展开' : '收起' }}
          </a-button>
        </div>

        <div v-if="!groupCollapsedStates[group.key]" class="columns-row">
          <div 
            class="simple-column" 
            v-for="column in getGroupColumns(group)" 
            :key="column.status"
          >
            <div class="column-header">
              <span>{{ column.title }}</span>
              <a-badge :count="column.items.length" />
            </div>
            <div class="column-cards">
              <div 
                class="simple-card" 
                v-for="item in column.items" 
                :key="item.id"
              >
                <div class="card-title">{{ item.company_name }}</div>
                <div class="card-position">{{ item.position_title }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-else class="loading-container">
      <a-spin size="large" tip="加载中..." />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons-vue'
import { useJobApplicationStore } from '../stores/jobApplication'
import { ApplicationStatus, type JobApplication } from '../types'

// 简化的分组定义
interface SimpleGroup {
  key: string
  title: string
  collapsible: boolean
  defaultCollapsed: boolean
  columns: {
    status: ApplicationStatus
    title: string
    color: string
  }[]
}

const kanbanGroups: SimpleGroup[] = [
  {
    key: 'in-progress',
    title: '进行中状态',
    collapsible: false,
    defaultCollapsed: false,
    columns: [
      { status: ApplicationStatus.APPLIED, title: '已投递', color: '#1890ff' },
      { status: ApplicationStatus.RESUME_SCREENING, title: '简历筛选中', color: '#13c2c2' },
      { status: ApplicationStatus.WRITTEN_TEST, title: '笔试中', color: '#fa8c16' },
      { status: ApplicationStatus.FIRST_INTERVIEW, title: '一面中', color: '#722ed1' },
      { status: ApplicationStatus.SECOND_INTERVIEW, title: '二面中', color: '#eb2f96' },
      { status: ApplicationStatus.THIRD_INTERVIEW, title: '三面中', color: '#13c2c2' }
    ]
  },
  {
    key: 'success',
    title: '成功状态',
    collapsible: false,
    defaultCollapsed: false,
    columns: [
      { status: ApplicationStatus.OFFER_WAITING, title: '待发Offer', color: '#faad14' },
      { status: ApplicationStatus.OFFER_RECEIVED, title: '已收Offer', color: '#52c41a' },
      { status: ApplicationStatus.PROCESS_FINISHED, title: '流程结束', color: '#52c41a' }
    ]
  },
  {
    key: 'failed',
    title: '失败状态',
    collapsible: true,
    defaultCollapsed: true,
    columns: [
      { status: ApplicationStatus.RESUME_SCREENING_FAIL, title: '简历被拒', color: '#ff7875' },
      { status: ApplicationStatus.WRITTEN_TEST_FAIL, title: '笔试未过', color: '#ff7875' },
      { status: ApplicationStatus.FIRST_FAIL, title: '一面未过', color: '#ff7875' }
    ]
  }
]

const jobStore = useJobApplicationStore()
const { applications, loading } = storeToRefs(jobStore)

// 分组折叠状态
const groupCollapsedStates = ref<Record<string, boolean>>({})

// 初始化折叠状态
kanbanGroups.forEach(group => {
  groupCollapsedStates.value[group.key] = group.defaultCollapsed
})

// 切换分组折叠状态
const toggleGroupCollapse = (groupKey: string) => {
  groupCollapsedStates.value[groupKey] = !groupCollapsedStates.value[groupKey]
}

// 获取分组的列数据
const getGroupColumns = (group: SimpleGroup) => {
  return group.columns.map(col => ({
    ...col,
    items: applications.value.filter(app => app.status === col.status)
  }))
}

// 获取分组总数量
const getGroupTotalCount = (group: SimpleGroup): number => {
  return group.columns.reduce((total, col) => {
    return total + applications.value.filter(app => app.status === col.status).length
  }, 0)
}

// 获取数据
const fetchData = () => jobStore.fetchApplications()

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.simple-kanban {
  padding: 24px;
  background: #f0f2f5;
  min-height: 100vh;
}

.kanban-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.kanban-header h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.kanban-actions {
  display: flex;
  gap: 12px;
}

.test-info {
  background: #fff;
  padding: 16px;
  border-radius: 8px;
  margin-bottom: 24px;
  border-left: 4px solid #1890ff;
}

.test-info p {
  margin: 4px 0;
  font-size: 14px;
}

.groups-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.group-section {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.group-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.group-title h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.columns-row {
  display: flex;
  gap: 16px;
  overflow-x: auto;
}

.simple-column {
  flex: 0 0 200px;
  background: #fafafa;
  border-radius: 6px;
  padding: 12px;
}

.column-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-weight: 600;
  font-size: 14px;
}

.column-cards {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.simple-card {
  background: #fff;
  border-radius: 4px;
  padding: 8px;
  border-left: 3px solid #1890ff;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.card-title {
  font-weight: 600;
  font-size: 13px;
  margin-bottom: 4px;
}

.card-position {
  font-size: 12px;
  color: #666;
}

.loading-container {
  height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>