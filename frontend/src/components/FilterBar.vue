<template>
  <div class="filter-bar">
    <a-card :bordered="false" class="filter-card">
      <a-row :gutter="16" align="middle">
        <!-- 搜索框 -->
        <a-col :xs="24" :sm="12" :md="8" :lg="6">
          <a-input
            v-model:value="filters.keyword"
            placeholder="搜索公司名称或职位"
            allow-clear
            @change="handleSearch"
          >
            <template #prefix>
              <SearchOutlined />
            </template>
          </a-input>
        </a-col>

        <!-- 状态筛选 -->
        <a-col :xs="24" :sm="12" :md="8" :lg="6">
          <a-select
            v-model:value="filters.status"
            placeholder="选择状态"
            allow-clear
            style="width: 100%"
            @change="handleFilterChange"
          >
            <a-select-option value="">全部状态</a-select-option>
            <a-select-option v-for="status in statusOptions" :key="status" :value="status">
              {{ status }}
            </a-select-option>
          </a-select>
        </a-col>

        <!-- 日期范围选择 -->
        <a-col :xs="24" :sm="12" :md="8" :lg="6">
          <a-range-picker
            v-model:value="filters.dateRange"
            format="YYYY-MM-DD"
            style="width: 100%"
            :placeholder="['开始日期', '结束日期']"
            @change="handleDateChange"
          />
        </a-col>

        <!-- 薪资范围筛选 -->
        <a-col :xs="24" :sm="12" :md="8" :lg="4">
          <a-select
            v-model:value="filters.salaryRange"
            placeholder="薪资范围"
            allow-clear
            style="width: 100%"
            @change="handleFilterChange"
          >
            <a-select-option value="">全部薪资</a-select-option>
            <a-select-option value="0-10">10K以下</a-select-option>
            <a-select-option value="10-15">10-15K</a-select-option>
            <a-select-option value="15-20">15-20K</a-select-option>
            <a-select-option value="20-25">20-25K</a-select-option>
            <a-select-option value="25-30">25-30K</a-select-option>
            <a-select-option value="30+">30K以上</a-select-option>
          </a-select>
        </a-col>

        <!-- 操作按钮 -->
        <a-col :xs="24" :sm="24" :md="24" :lg="2">
          <a-space>
            <a-button @click="handleReset">重置</a-button>
          </a-space>
        </a-col>
      </a-row>

      <!-- 高级筛选选项 -->
      <div v-if="showAdvanced" class="advanced-filters">
        <a-divider />
        <a-row :gutter="16">
          <a-col :span="8">
            <a-input
              v-model:value="filters.location"
              placeholder="工作地点"
              allow-clear
              @change="handleFilterChange"
            >
              <template #prefix>
                <EnvironmentOutlined />
              </template>
            </a-input>
          </a-col>
          <a-col :span="8">
            <a-select
              v-model:value="filters.sortBy"
              placeholder="排序方式"
              style="width: 100%"
              @change="handleFilterChange"
            >
              <a-select-option value="date_desc">最新投递</a-select-option>
              <a-select-option value="date_asc">最早投递</a-select-option>
              <a-select-option value="salary_desc">薪资从高到低</a-select-option>
              <a-select-option value="salary_asc">薪资从低到高</a-select-option>
            </a-select>
          </a-col>
        </a-row>
      </div>

      <!-- 高级筛选切换 -->
      <div class="advanced-toggle">
        <a @click="showAdvanced = !showAdvanced">
          {{ showAdvanced ? '收起' : '高级筛选' }}
          <component :is="showAdvanced ? 'UpOutlined' : 'DownOutlined'" />
        </a>
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { 
  SearchOutlined, 
  EnvironmentOutlined,
  UpOutlined,
  DownOutlined
} from '@ant-design/icons-vue'
import { ApplicationStatus } from '../types'
import type { Dayjs } from 'dayjs'

interface FilterOptions {
  keyword: string
  status: string
  dateRange: [Dayjs, Dayjs] | null
  salaryRange: string
  location: string
  sortBy: string
}

interface Emits {
  (e: 'change', filters: FilterOptions): void
}

const emit = defineEmits<Emits>()

const showAdvanced = ref(false)

const filters = reactive<FilterOptions>({
  keyword: '',
  status: '',
  dateRange: null,
  salaryRange: '',
  location: '',
  sortBy: 'date_desc'
})

// 状态选项
const statusOptions = Object.values(ApplicationStatus)

// 处理搜索
const handleSearch = () => {
  emit('change', filters)
}

// 处理筛选变化
const handleFilterChange = () => {
  emit('change', filters)
}

// 处理日期变化
const handleDateChange = () => {
  emit('change', filters)
}

// 重置筛选
const handleReset = () => {
  filters.keyword = ''
  filters.status = ''
  filters.dateRange = null
  filters.salaryRange = ''
  filters.location = ''
  filters.sortBy = 'date_desc'
  showAdvanced.value = false
  emit('change', filters)
}

// 监听所有筛选条件变化
watch(filters, (newFilters) => {
  emit('change', newFilters)
}, { deep: true })
</script>

<style scoped>
.filter-bar {
  margin-bottom: 16px;
}

.filter-card {
  background: #fff;
}

.filter-card :deep(.ant-card-body) {
  padding: 16px;
}

.advanced-filters {
  margin-top: 16px;
}

.advanced-toggle {
  text-align: center;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

.advanced-toggle a {
  color: #1890ff;
  font-size: 14px;
}

.advanced-toggle a:hover {
  color: #40a9ff;
}

/* 响应式布局 */
@media (max-width: 768px) {
  .filter-card :deep(.ant-col) {
    margin-bottom: 8px;
  }
}
</style>