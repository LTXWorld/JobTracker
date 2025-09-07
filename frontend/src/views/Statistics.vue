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
              title="已通过"
              :value="statisticsData?.passed || offerCount"
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
              title="总通过率"
              :value="statisticsData?.pass_rate || `${successRate}%`"
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
            <v-chart class="chart" :option="statusPieOption" />
          </a-card>
        </a-col>

        <!-- 投递趋势折线图 -->
        <a-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
          <a-card title="投递趋势（最近30天）" :bordered="false" class="chart-card">
            <v-chart class="chart" :option="trendLineOption" />
          </a-card>
        </a-col>

        <!-- 各阶段通过率 -->
        <a-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
          <a-card title="各阶段通过率" :bordered="false" class="chart-card">
            <v-chart class="chart" :option="stageBarOption" />
          </a-card>
        </a-col>

        <!-- 公司和薪资分布 -->
        <a-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
          <a-card title="薪资分布" :bordered="false" class="chart-card">
            <v-chart class="chart" :option="salaryBarOption" />
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- 详细数据表格 -->
    <a-card title="投递详情统计" :bordered="false" class="detail-card">
      <a-table 
        :columns="tableColumns" 
        :data-source="tableData"
        :pagination="false"
        size="middle"
      />
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import {
  CanvasRenderer
} from 'echarts/renderers'
import {
  PieChart,
  LineChart,
  BarChart
} from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
} from 'echarts/components'
import {
  SendOutlined,
  ClockCircleOutlined,
  TrophyOutlined,
  RiseOutlined,
  CloseCircleOutlined,
  CalendarOutlined
} from '@ant-design/icons-vue'
import { useJobApplicationStore } from '../stores/jobApplication'
import { ApplicationStatus, StatusHelper } from '../types'
import dayjs from 'dayjs'

// 注册ECharts组件
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

const jobStore = useJobApplicationStore()
const { applications, loading, statistics: statisticsData, statisticsLoading } = storeToRefs(jobStore)

// 统计数据计算
const totalApplications = computed(() => applications.value.length)

// 使用StatusHelper进行状态分类
const inProgressCount = computed(() => {
  return applications.value.filter(app => StatusHelper.isInProgressStatus(app.status)).length
})

const offerCount = computed(() => {
  return applications.value.filter(app => StatusHelper.isPassedStatus(app.status)).length
})

const failedCount = computed(() => {
  return applications.value.filter(app => StatusHelper.isFailedStatus(app.status)).length
})

const successRate = computed(() => {
  const total = applications.value.length
  if (total === 0) return 0
  return (offerCount.value / total) * 100
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

// 各阶段通过率柱状图配置
const stageBarOption = computed(() => {
  const stages = [
    { name: '笔试', total: ApplicationStatus.WRITTEN_TEST, pass: ApplicationStatus.WRITTEN_TEST_PASS },
    { name: '一面', total: ApplicationStatus.FIRST_INTERVIEW, pass: ApplicationStatus.FIRST_PASS },
    { name: '二面', total: ApplicationStatus.SECOND_INTERVIEW, pass: ApplicationStatus.SECOND_PASS },
    { name: '三面', total: ApplicationStatus.THIRD_INTERVIEW, pass: ApplicationStatus.THIRD_PASS },
    { name: 'HR面', total: ApplicationStatus.HR_INTERVIEW, pass: ApplicationStatus.HR_PASS }
  ]

  const names = stages.map(s => s.name)
  const rates = stages.map(stage => {
    const totalCount = applications.value.filter(app => 
      app.status === stage.total || app.status === stage.pass
    ).length
    const passCount = applications.value.filter(app => app.status === stage.pass).length
    return totalCount > 0 ? ((passCount / totalCount) * 100).toFixed(1) : 0
  })

  return {
    tooltip: {
      trigger: 'axis',
      formatter: '{b}: {c}%'
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: names
    },
    yAxis: {
      type: 'value',
      max: 100,
      axisLabel: {
        formatter: '{value}%'
      }
    },
    series: [
      {
        type: 'bar',
        data: rates,
        itemStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [{
              offset: 0, color: '#52c41a'
            }, {
              offset: 1, color: '#a0d911'
            }]
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

onMounted(async () => {
  await jobStore.fetchApplications()
  await jobStore.fetchStatistics() // 获取服务器端统计数据
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