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
              :value="`${conversionRate}%`"
              :value-style="{ color: '#722ed1' }"
            >
              <template #title>
                <span>转化率</span>
                <a-tooltip placement="top">
                  <template #title>
                    <div>
                      <div><strong>转化率计算方法：</strong></div>
                      <div>进入面试次数 ÷ 总投递数 × 100%</div>
                      <div style="margin-top: 8px;">
                        <span style="color: #52c41a;">{{ interviewCount }}</span> ÷
                        <span style="color: #1890ff;">{{ totalApplications }}</span> × 100% =
                        <span style="color: #722ed1;">{{ conversionRate }}%</span>
                      </div>
                      <div style="margin-top: 8px; font-size: 12px; opacity: 0.9;">
                        进入面试包括：一面、二面、三面、HR面及其通过/挂的状态，还包括后续的offer相关状态
                      </div>
                    </div>
                  </template>
                  <InfoCircleOutlined style="margin-left: 4px; font-size: 14px; color: #999; cursor: help;" />
                </a-tooltip>
              </template>
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
          <a-card :bordered="false" class="chart-card">
            <template #title>
              <span>各阶段通过率</span>
              <a-tooltip placement="top">
                <template #title>
                  <div>
                    <div><strong>统计说明:</strong></div>
                    <div>只统计实际经历过该阶段的岗位</div>
                    <div style="margin-top: 8px;">
                      例如:从一面直接到HR面的岗位<br/>
                      <span style="color: #52c41a;">✓</span> 会计入一面通过率<br/>
                      <span style="color: #ff4d4f;">✗</span> 不会计入二面/三面通过率
                    </div>
                  </div>
                </template>
                <InfoCircleOutlined style="margin-left: 4px; font-size: 14px; color: #999; cursor: help;" />
              </a-tooltip>
            </template>
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
  DownloadOutlined,
  InfoCircleOutlined
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

// 已OC：统计进入过HR面通过阶段的岗位（含已接受/拒绝offer）
const ocCount = computed(() => {
  const stats = statisticsData.value
  if (stats && typeof stats.hr_passed === 'number') {
    return stats.hr_passed
  }
  const hrPassedStatuses: ApplicationStatus[] = [
    ApplicationStatus.HR_PASS,
    ApplicationStatus.OFFER_ACCEPTED,
    ApplicationStatus.REJECTED
  ]
  return applications.value.filter(app => hrPassedStatuses.includes(app.status)).length
})

const failedCount = computed(() => {
  return applications.value.filter(app => StatusHelper.isFailedStatus(app.status)).length
})

const successRate = computed(() => {
  const total = applications.value.length
  if (total === 0) return 0
  return (offerCount.value / total) * 100
})

// OC率：仅以"已接受offer"占比计算
const ocRate = computed(() => {
  const total = applications.value.length
  if (total === 0) return 0
  return Number(((ocCount.value / total) * 100).toFixed(1))
})

// 转化率：进入面试的次数 / 总投递数
const conversionRate = computed(() => {
  const total = applications.value.length
  if (total === 0) return 0

  // 统计真正进入面试的申请（一面、二面、三面、HR面及其通过/未通过状态）
  const interviewStatuses = [
    ApplicationStatus.FIRST_INTERVIEW, ApplicationStatus.FIRST_PASS, ApplicationStatus.FIRST_FAIL,
    ApplicationStatus.SECOND_INTERVIEW, ApplicationStatus.SECOND_PASS, ApplicationStatus.SECOND_FAIL,
    ApplicationStatus.THIRD_INTERVIEW, ApplicationStatus.THIRD_PASS, ApplicationStatus.THIRD_FAIL,
    ApplicationStatus.HR_INTERVIEW, ApplicationStatus.HR_PASS, ApplicationStatus.HR_FAIL,
    ApplicationStatus.OFFER_ACCEPTED, ApplicationStatus.REJECTED
  ]

  const interviewCount = applications.value.filter(app =>
    interviewStatuses.includes(app.status)
  ).length

  return Number(((interviewCount / total) * 100).toFixed(1))
})

// 进入面试次数（用于tooltip显示）
const interviewCount = computed(() => {
  // 统计真正进入面试的申请（一面、二面、三面、HR面及其通过/未通过状态）
  const interviewStatuses = [
    ApplicationStatus.FIRST_INTERVIEW, ApplicationStatus.FIRST_PASS, ApplicationStatus.FIRST_FAIL,
    ApplicationStatus.SECOND_INTERVIEW, ApplicationStatus.SECOND_PASS, ApplicationStatus.SECOND_FAIL,
    ApplicationStatus.THIRD_INTERVIEW, ApplicationStatus.THIRD_PASS, ApplicationStatus.THIRD_FAIL,
    ApplicationStatus.HR_INTERVIEW, ApplicationStatus.HR_PASS, ApplicationStatus.HR_FAIL,
    ApplicationStatus.OFFER_ACCEPTED, ApplicationStatus.REJECTED
  ]

  return applications.value.filter(app =>
    interviewStatuses.includes(app.status)
  ).length
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

// 各阶段通过率柱状图配置（优先使用后端StageAnalysis；若无则用前端推断，考虑"直通"场景）
const stageBarOption = computed(() => {
  // 优先后端口径
  const sa: any = (analyticsData.value as any)?.stage_analysis
  if (sa && Object.keys(sa).length > 0) {
    const order = ['written', 'first', 'second', 'third', 'hr']
    const names = order
      .filter(k => sa[k])
      .map(k => ({ key: k, name: k === 'written' ? '笔试' : k === 'first' ? '一面' : k === 'second' ? '二面' : k === 'third' ? '三面' : 'HR面' }))
    const rates = names.map(n => Number((sa[n.key].success_rate || sa[n.key].SuccessRate || 0).toFixed?.(1) ?? sa[n.key].success_rate ?? 0))

    return {
      tooltip: {
        trigger: 'axis',
        formatter: (params: any) => {
          const idx = params[0].dataIndex
          const stageKey = order.filter(k => sa[k])[idx]
          const stageData = sa[stageKey]
          const pass = stageData.success_count || stageData.SuccessCount || 0
          const total = stageData.total_count || stageData.TotalCount || 0
          const rate = params[0].value
          return `${params[0].axisValue}<br/>通过: ${pass}/${total}<br/>通过率: ${rate}%`
        }
      },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: { type: 'category', data: names.map(n => n.name) },
      yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
      series: [{
        type: 'bar',
        data: rates,
        itemStyle: { color: '#52c41a' },
        label: {
          show: true,
          position: 'top',
          formatter: (params: any) => {
            const idx = params.dataIndex
            const stageKey = order.filter(k => sa[k])[idx]
            const stageData = sa[stageKey]
            const pass = stageData.success_count || stageData.SuccessCount || 0
            const total = stageData.total_count || stageData.TotalCount || 0
            return `${pass}/${total}`
          },
          color: '#666',
          fontSize: 12
        }
      }]
    }
  }
  const S = ApplicationStatus
  const stages = [
    {
      name: '笔试',
      entry: S.WRITTEN_TEST,
      pass: [S.WRITTEN_TEST_PASS],
      next: [S.FIRST_INTERVIEW, S.FIRST_PASS, S.SECOND_INTERVIEW, S.SECOND_PASS, S.THIRD_INTERVIEW, S.THIRD_PASS, S.HR_INTERVIEW, S.HR_PASS, S.OFFER_ACCEPTED, S.REJECTED]
    },
    {
      name: '一面',
      entry: S.FIRST_INTERVIEW,
      pass: [S.FIRST_PASS],
      next: [S.SECOND_INTERVIEW, S.SECOND_PASS, S.THIRD_INTERVIEW, S.THIRD_PASS, S.HR_INTERVIEW, S.HR_PASS, S.OFFER_ACCEPTED, S.REJECTED]
    },
    {
      name: '二面',
      entry: S.SECOND_INTERVIEW,
      pass: [S.SECOND_PASS],
      next: [S.THIRD_INTERVIEW, S.THIRD_PASS, S.HR_INTERVIEW, S.HR_PASS, S.OFFER_ACCEPTED, S.REJECTED]
    },
    {
      name: '三面',
      entry: S.THIRD_INTERVIEW,
      pass: [S.THIRD_PASS],
      next: [S.HR_INTERVIEW, S.HR_PASS, S.OFFER_ACCEPTED, S.REJECTED]
    },
    {
      name: 'HR面',
      entry: S.HR_INTERVIEW,
      pass: [S.HR_PASS],
      next: [S.OFFER_ACCEPTED, S.REJECTED]
    }
  ]

  const inSet = (st: ApplicationStatus, list: ApplicationStatus[]) => list.includes(st)
  const names = stages.map(s => s.name)

  // 计算每个阶段的详细数据
  const stageData = stages.map(stage => {
    const totalCount = applications.value.filter(app => inSet(app.status as ApplicationStatus, [stage.entry, ...stage.pass, ...stage.next])).length
    const passCount = applications.value.filter(app => inSet(app.status as ApplicationStatus, [...stage.pass, ...stage.next])).length
    const rate = totalCount > 0 ? Number(((passCount / totalCount) * 100).toFixed(1)) : 0
    return { totalCount, passCount, rate }
  })

  const rates = stageData.map(d => d.rate)

  return {
    tooltip: {
      trigger: 'axis',
      formatter: (params: any) => {
        const idx = params[0].dataIndex
        const data = stageData[idx]
        return `${params[0].axisValue}<br/>通过: ${data.passCount}/${data.totalCount}<br/>通过率: ${data.rate}%`
      }
    },
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
        },
        label: {
          show: true,
          position: 'top',
          formatter: (params: any) => {
            const idx = params.dataIndex
            const data = stageData[idx]
            return `${data.passCount}/${data.totalCount}`
          },
          color: '#666',
          fontSize: 12
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
    
    const interviewStatuses: ApplicationStatus[] = [
      ApplicationStatus.WRITTEN_TEST,
      ApplicationStatus.FIRST_INTERVIEW,
      ApplicationStatus.SECOND_INTERVIEW,
      ApplicationStatus.THIRD_INTERVIEW,
      ApplicationStatus.HR_INTERVIEW
    ]
    
    if (interviewStatuses.includes(app.status)) {
      data.interviewing++
    } else if ([ApplicationStatus.HR_PASS, ApplicationStatus.OFFER_ACCEPTED].includes(app.status)) {
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
