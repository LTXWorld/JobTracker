<template>
  <div class="statistics-page">
    <!-- 统计概览卡片 -->
    <div class="stats-overview">
      <a-row :gutter="16">
        <a-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
          <a-card :bordered="false" class="stat-card">
            <a-statistic
              title="总投递数"
              :value="statisticsData?.total_applications || totalApplications"
              :value-style="{ color: '#1890ff' }"
            >
              <template #prefix>
                <SendOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
          <a-card :bordered="false" class="stat-card">
            <a-statistic
              title="进行中"
              :value="statisticsData?.in_progress || inProgressCount"
              :value-style="{ color: '#fa8c16' }"
            >
              <template #prefix>
                <ClockCircleOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
          <a-card :bordered="false" class="stat-card">
            <a-statistic
              title="已OC"
              :value="ocCount"
              :value-style="{ color: '#52c41a' }"
            >
              <template #prefix>
                <TrophyOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
          <a-card :bordered="false" class="stat-card">
            <a-statistic
              title="已失败"
              :value="statisticsData?.failed || failedCount"
              :value-style="{ color: '#ff4d4f' }"
            >
              <template #prefix>
                <CloseCircleOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
      </a-row>
      
      <!-- 第二行：通过率和详细分析 -->
      <a-row :gutter="16" style="margin-top: 16px;">
        <a-col :xs="12" :sm="12" :md="8" :lg="8" :xl="8">
          <a-card :bordered="false" class="stat-card">
            <a-statistic
              title="OC率"
              :value="`${ocRate}%`"
              :value-style="{ color: '#722ed1' }"
            >
              <template #prefix>
                <RiseOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :xs="12" :sm="12" :md="8" :lg="8" :xl="8">
          <a-card :bordered="false" class="stat-card">
            <a-statistic
              title="本月投递"
              :value="monthlyCount"
              :value-style="{ color: '#13c2c2' }"
            >
              <template #prefix>
                <CalendarOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :xs="12" :sm="12" :md="8" :lg="8" :xl="8">
          <a-card :bordered="false" class="stat-card">
            <a-statistic
              title="本周投递"
              :value="weeklyCount"
              :value-style="{ color: '#fa541c' }"
            >
              <template #prefix>
                <ClockCircleOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- 图表区域 -->
    <div class="charts-container">
      <a-row :gutter="16">
        <!-- 状态分布饼图 -->
        <a-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
          <a-card title="投递状态分布" :bordered="false" class="chart-card">
            <component :is="VChart" class="chart" :option="statusPieOption" />
          </a-card>
        </a-col>

        <!-- 投递趋势折线图 -->
        <a-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
          <a-card title="投递趋势（最近30天）" :bordered="false" class="chart-card">
            <component :is="VChart" class="chart" :option="trendLineOption" />
          </a-card>
        </a-col>

        <!-- 各阶段通过率 -->
        <a-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
          <a-card title="各阶段通过率" :bordered="false" class="chart-card">
            <component :is="VChart" class="chart" :option="stageBarOption" />
          </a-card>
        </a-col>

        <!-- 公司和薪资分布 -->
        <a-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
          <a-card title="薪资分布" :bordered="false" class="chart-card">
            <component :is="VChart" class="chart" :option="salaryBarOption" />
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- 详细数据表格 -->
    <a-card :bordered="false" class="detail-card">
      <template #title>
        <span>投递详情统计</span>
      </template>
      <template #extra>
        <a-space>
          <a-button type="default" @click="showExportHistoryModal = true">
            <template #icon><HistoryOutlined /></template>
            导出历史
          </a-button>
          <a-button type="primary" @click="showExportModal = true">
            <template #icon><DownloadOutlined /></template>
            导出统计报告
          </a-button>
        </a-space>
      </template>
      
      <a-table 
        :columns="tableColumns" 
        :data-source="tableData"
        :pagination="false"
        size="middle"
      />
    </a-card>

    <!-- 导出历史弹窗 -->
    <ExportHistory
      v-model:visible="showExportHistoryModal"
    />

    <!-- 导出统计报告弹窗 -->
    <ExportDialog
      v-model:visible="showExportModal"
      :applications="applications"
      @success="handleExportSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, defineAsyncComponent } from 'vue'
import { storeToRefs } from 'pinia'
// 延迟加载图表相关依赖，避免首次进入时动态导入失败导致整页加载失败
const VChart = defineAsyncComponent(async () => {
  try {
    const [{ default: Comp }, core, renderers, charts, comps] = await Promise.all([
      import('vue-echarts'),
      import('echarts/core'),
      import('echarts/renderers'),
      import('echarts/charts'),
      import('echarts/components')
    ])
    // 运行时按需注册
    const { use } = core as any
    const { CanvasRenderer } = renderers as any
    const { PieChart, LineChart, BarChart } = charts as any
    const { TitleComponent, TooltipComponent, LegendComponent, GridComponent } = comps as any
    use([
      CanvasRenderer,
      PieChart,
      LineChart,
      BarChart,
      TitleComponent,
      TooltipComponent,
      LegendComponent,
      GridComponent
    ])
    return (Comp as any)
  } catch (e) {
    console.error('VChart load failed:', e)
    // 返回空渲染组件，保证页面其余部分可用
    return {
      name: 'ChartFallback',
      render() { return null }
    } as any
  }
})
import {
  SendOutlined,
  ClockCircleOutlined,
  TrophyOutlined,
  RiseOutlined,
  CloseCircleOutlined,
  CalendarOutlined,
  HistoryOutlined,
  DownloadOutlined
} from '@ant-design/icons-vue'
import { useJobApplicationStore } from '../stores/jobApplication'
import { useStatusTrackingStore } from '../stores/statusTracking'
import { ApplicationStatus, StatusHelper } from '../types'
import ExportHistory from '../components/ExportHistory.vue'
import ExportDialog from '../components/ExportDialog.vue'
import dayjs from 'dayjs'

// 由 defineAsyncComponent 内部在运行时注册 ECharts 依赖

const jobStore = useJobApplicationStore()
const statusTrackingStore = useStatusTrackingStore()
const { analytics: analyticsData } = storeToRefs(statusTrackingStore)
const { applications, loading, statistics: statisticsData, statisticsLoading } = storeToRefs(jobStore)

// 弹窗状态
const showExportHistoryModal = ref(false)
const showExportModal = ref(false)

// 统计数据计算
const totalApplications = computed(() => applications.value.length)

// 使用StatusHelper进行状态分类
const inProgressCount = computed(() => {
  return applications.value.filter(app => StatusHelper.isInProgressStatus(app.status)).length
})

// 已通过（旧口径）仍保留用于 successRate 计算
const offerCount = computed(() => {
  return applications.value.filter(app => StatusHelper.isPassedStatus(app.status)).length
})

// 已OC：仅统计“已收到offer”
const ocCount = computed(() => {
  return applications.value.filter(app => app.status === ApplicationStatus.OFFER_RECEIVED).length
})

const failedCount = computed(() => {
  return applications.value.filter(app => StatusHelper.isFailedStatus(app.status)).length
})

const successRate = computed(() => {
  const total = applications.value.length
  if (total === 0) return 0
  return (offerCount.value / total) * 100
})

// OC率：仅以“已收到offer”占比计算
const ocRate = computed(() => {
  const total = applications.value.length
  if (total === 0) return 0
  return Number(((ocCount.value / total) * 100).toFixed(1))
})

// 本月投递数
const monthlyCount = computed(() => {
  const currentMonth = dayjs().startOf('month')
  return applications.value.filter(app => 
    dayjs(app.application_date).isAfter(currentMonth)
  ).length
})

// 本周投递数  
const weeklyCount = computed(() => {
  const currentWeek = dayjs().startOf('week')
  return applications.value.filter(app => 
    dayjs(app.application_date).isAfter(currentWeek)
  ).length
})

// 状态分布饼图配置
const statusPieOption = computed(() => {
  const statusMap = new Map<string, number>()
  applications.value.forEach(app => {
    statusMap.set(app.status, (statusMap.get(app.status) || 0) + 1)
  })
  
  const data = Array.from(statusMap.entries()).map(([name, value]) => ({
    name,
    value
  }))

  return {
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 'left'
    },
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: {
          show: false,
          position: 'center'
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 20,
            fontWeight: 'bold'
          }
        },
        labelLine: {
          show: false
        },
        data
      }
    ]
  }
})

// 投递趋势折线图配置
const trendLineOption = computed(() => {
  const last30Days = []
  const counts = []
  
  for (let i = 29; i >= 0; i--) {
    const date = dayjs().subtract(i, 'day')
    last30Days.push(date.format('MM-DD'))
    
    const count = applications.value.filter(app => 
      dayjs(app.application_date).format('YYYY-MM-DD') === date.format('YYYY-MM-DD')
    ).length
    counts.push(count)
  }

  return {
    tooltip: {
      trigger: 'axis'
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: last30Days
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: '投递数',
        type: 'line',
        smooth: true,
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [{
              offset: 0, color: 'rgba(24, 144, 255, 0.3)'
            }, {
              offset: 1, color: 'rgba(24, 144, 255, 0.1)'
            }]
          }
        },
        data: counts,
        itemStyle: {
          color: '#1890ff'
        }
      }
    ]
  }
})

// 各阶段通过率柱状图配置（优先使用后端StageAnalysis；若无则用前端推断，考虑“直通”场景）
const stageBarOption = computed(() => {
  // 优先后端口径
  const sa: any = analyticsData.value?.StageAnalysis || (analyticsData.value as any)?.stage_analysis
  if (sa && Object.keys(sa).length > 0) {
    const order = ['written', 'first', 'second', 'third', 'hr']
    const names = order
      .filter(k => sa[k])
      .map(k => ({ key: k, name: k === 'written' ? '笔试' : k === 'first' ? '一面' : k === 'second' ? '二面' : k === 'third' ? '三面' : 'HR面' }))
    const rates = names.map(n => Number((sa[n.key].success_rate || sa[n.key].SuccessRate || 0).toFixed?.(1) ?? sa[n.key].success_rate ?? 0))
    return {
      tooltip: { trigger: 'axis', formatter: '{b}: {c}%' },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: { type: 'category', data: names.map(n => n.name) },
      yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
      series: [{ type: 'bar', data: rates, itemStyle: { color: '#52c41a' } }]
    }
  }
  const S = ApplicationStatus
  const stages = [
    {
      name: '笔试',
      entry: S.WRITTEN_TEST,
      pass: [S.WRITTEN_TEST_PASS],
      next: [S.FIRST_INTERVIEW, S.FIRST_PASS, S.SECOND_INTERVIEW, S.SECOND_PASS, S.THIRD_INTERVIEW, S.THIRD_PASS, S.HR_INTERVIEW, S.HR_PASS, S.OFFER_WAITING, S.OFFER_RECEIVED, S.OFFER_ACCEPTED]
    },
    {
      name: '一面',
      entry: S.FIRST_INTERVIEW,
      pass: [S.FIRST_PASS],
      next: [S.SECOND_INTERVIEW, S.SECOND_PASS, S.THIRD_INTERVIEW, S.THIRD_PASS, S.HR_INTERVIEW, S.HR_PASS, S.OFFER_WAITING, S.OFFER_RECEIVED, S.OFFER_ACCEPTED]
    },
    {
      name: '二面',
      entry: S.SECOND_INTERVIEW,
      pass: [S.SECOND_PASS],
      next: [S.THIRD_INTERVIEW, S.THIRD_PASS, S.HR_INTERVIEW, S.HR_PASS, S.OFFER_WAITING, S.OFFER_RECEIVED, S.OFFER_ACCEPTED]
    },
    {
      name: '三面',
      entry: S.THIRD_INTERVIEW,
      pass: [S.THIRD_PASS],
      next: [S.HR_INTERVIEW, S.HR_PASS, S.OFFER_WAITING, S.OFFER_RECEIVED, S.OFFER_ACCEPTED]
    },
    {
      name: 'HR面',
      entry: S.HR_INTERVIEW,
      pass: [S.HR_PASS],
      next: [S.OFFER_WAITING, S.OFFER_RECEIVED, S.OFFER_ACCEPTED]
    }
  ]

  const inSet = (st: ApplicationStatus, list: ApplicationStatus[]) => list.includes(st)
  const names = stages.map(s => s.name)
  const rates = stages.map(stage => {
    const totalCount = applications.value.filter(app => inSet(app.status as ApplicationStatus, [stage.entry, ...stage.pass, ...stage.next])).length
    const passCount = applications.value.filter(app => inSet(app.status as ApplicationStatus, [...stage.pass, ...stage.next])).length
    return totalCount > 0 ? Number(((passCount / totalCount) * 100).toFixed(1)) : 0
  })

  return {
    tooltip: { trigger: 'axis', formatter: '{b}: {c}%' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: names },
    yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
    series: [
      {
        type: 'bar',
        data: rates,
        itemStyle: {
          color: {
            type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [{ offset: 0, color: '#52c41a' }, { offset: 1, color: '#a0d911' }]
          },
          borderRadius: [5, 5, 0, 0]
        }
      }
    ]
  }
})

// 薪资分布柱状图配置
const salaryBarOption = computed(() => {
  const salaryRanges = {
    '10K以下': 0,
    '10-15K': 0,
    '15-20K': 0,
    '20-25K': 0,
    '25-30K': 0,
    '30K以上': 0
  }

  applications.value.forEach(app => {
    if (app.salary_range) {
      const match = app.salary_range.match(/(\d+)/)
      if (match) {
        const salary = parseInt(match[1])
        if (salary < 10) salaryRanges['10K以下']++
        else if (salary < 15) salaryRanges['10-15K']++
        else if (salary < 20) salaryRanges['15-20K']++
        else if (salary < 25) salaryRanges['20-25K']++
        else if (salary < 30) salaryRanges['25-30K']++
        else salaryRanges['30K以上']++
      }
    }
  })

  return {
    tooltip: {
      trigger: 'axis'
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: Object.keys(salaryRanges)
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        type: 'bar',
        data: Object.values(salaryRanges),
        itemStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [{
              offset: 0, color: '#722ed1'
            }, {
              offset: 1, color: '#b37feb'
            }]
          },
          borderRadius: [5, 5, 0, 0]
        }
      }
    ]
  }
})

// 表格配置
const tableColumns = [
  {
    title: '公司',
    dataIndex: 'company',
    key: 'company'
  },
  {
    title: '投递数',
    dataIndex: 'count',
    key: 'count',
    sorter: (a: any, b: any) => a.count - b.count
  },
  {
    title: '面试中',
    dataIndex: 'interviewing',
    key: 'interviewing'
  },
  {
    title: '已收Offer',
    dataIndex: 'offer',
    key: 'offer'
  },
  {
    title: '已挂',
    dataIndex: 'rejected',
    key: 'rejected'
  }
]

const tableData = computed(() => {
  const companyMap = new Map<string, any>()
  
  applications.value.forEach(app => {
    if (!companyMap.has(app.company_name)) {
      companyMap.set(app.company_name, {
        company: app.company_name,
        count: 0,
        interviewing: 0,
        offer: 0,
        rejected: 0
      })
    }
    
    const data = companyMap.get(app.company_name)!
    data.count++
    
    const interviewStatuses = [
      ApplicationStatus.WRITTEN_TEST,
      ApplicationStatus.FIRST_INTERVIEW,
      ApplicationStatus.SECOND_INTERVIEW,
      ApplicationStatus.THIRD_INTERVIEW,
      ApplicationStatus.HR_INTERVIEW
    ]
    
    if (interviewStatuses.includes(app.status)) {
      data.interviewing++
    } else if (app.status === ApplicationStatus.OFFER_RECEIVED || app.status === ApplicationStatus.OFFER_ACCEPTED) {
      data.offer++
    } else if (app.status === ApplicationStatus.REJECTED) {
      data.rejected++
    }
  })
  
  return Array.from(companyMap.values()).sort((a, b) => b.count - a.count)
})

// 导出成功处理函数
const handleExportSuccess = () => {
  showExportModal.value = false
  // 导出成功后可以刷新导出历史
}

onMounted(async () => {
  await jobStore.fetchApplications()
  await jobStore.fetchStatistics() // 获取服务器端统计数据（通用）
  try {
    await statusTrackingStore.fetchAnalytics(true) // 获取带StageAnalysis的分析数据
  } catch (e) {
    console.warn('获取状态分析失败，使用前端推断通过率', e)
  }
})
</script>

<style scoped>
.statistics-page {
  padding: 24px;
  background: #f0f2f5;
  min-height: calc(100vh - 48px - 56px - 70px);
}

.stats-overview {
  margin-bottom: 24px;
}

.stat-card {
  height: 100%;
}

.stat-card :deep(.ant-card-body) {
  padding: 20px;
}

.charts-container {
  margin-bottom: 24px;
}

.chart-card {
  margin-bottom: 16px;
  height: 400px;
}

.chart {
  height: 320px;
  width: 100%;
}

.detail-card {
  margin-top: 24px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .statistics-page {
    padding: 16px;
  }
  
  .chart-card {
    height: 350px;
  }
  
  .chart {
    height: 280px;
  }
}
</style>
