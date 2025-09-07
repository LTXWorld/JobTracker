import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { JobApplicationAPI } from '../api/jobApplication'
import type { 
  JobApplication, 
  CreateJobApplicationRequest, 
  UpdateJobApplicationRequest, 
  FilterOptions, 
  ApplicationStatus,
  JobApplicationStatistics 
} from '../types'
import { message } from 'ant-design-vue'

export const useJobApplicationStore = defineStore('jobApplication', () => {
  // 状态
  const applications = ref<JobApplication[]>([])
  const loading = ref(false)
  const currentApplication = ref<JobApplication | null>(null)
  const statistics = ref<JobApplicationStatistics | null>(null)
  const statisticsLoading = ref(false)

  // 计算属性
  const totalCount = computed(() => applications.value.length)
  
  const statusCounts = computed(() => {
    const counts: Record<ApplicationStatus, number> = {} as any
    applications.value.forEach(app => {
      counts[app.status] = (counts[app.status] || 0) + 1
    })
    return counts
  })

  // 获取所有投递记录
  const fetchApplications = async () => {
    loading.value = true
    try {
      applications.value = await JobApplicationAPI.getAll()
    } catch (error) {
      message.error('获取数据失败: ' + (error as Error).message)
    } finally {
      loading.value = false
    }
  }

  // 获取单个投递记录
  const fetchApplicationById = async (id: number) => {
    loading.value = true
    try {
      currentApplication.value = await JobApplicationAPI.getById(id)
    } catch (error) {
      message.error('获取详情失败: ' + (error as Error).message)
    } finally {
      loading.value = false
    }
  }

  // 创建投递记录
  const createApplication = async (data: CreateJobApplicationRequest) => {
    loading.value = true
    try {
      const newApp = await JobApplicationAPI.create(data)
      applications.value.unshift(newApp) // 添加到列表开头
      message.success('创建成功')
      return newApp
    } catch (error) {
      message.error('创建失败: ' + (error as Error).message)
      throw error
    } finally {
      loading.value = false
    }
  }

  // 更新投递记录
  const updateApplication = async (id: number, data: UpdateJobApplicationRequest) => {
    loading.value = true
    try {
      const updatedApp = await JobApplicationAPI.update(id, data)
      const index = applications.value.findIndex(app => app.id === id)
      if (index !== -1) {
        applications.value[index] = updatedApp
      }
      message.success('更新成功')
      return updatedApp
    } catch (error) {
      message.error('更新失败: ' + (error as Error).message)
      throw error
    } finally {
      loading.value = false
    }
  }

  // 删除投递记录
  const deleteApplication = async (id: number) => {
    loading.value = true
    try {
      await JobApplicationAPI.delete(id)
      applications.value = applications.value.filter(app => app.id !== id)
      message.success('删除成功')
    } catch (error) {
      message.error('删除失败: ' + (error as Error).message)
      throw error
    } finally {
      loading.value = false
    }
  }

  // 筛选应用
  const getFilteredApplications = (filters: FilterOptions) => {
    return computed(() => {
      let filtered = applications.value

      if (filters.status) {
        filtered = filtered.filter(app => app.status === filters.status)
      }

      if (filters.company) {
        filtered = filtered.filter(app => 
          app.company_name.toLowerCase().includes(filters.company!.toLowerCase())
        )
      }

      if (filters.dateRange) {
        const [startDate, endDate] = filters.dateRange
        filtered = filtered.filter(app => {
          const appDate = new Date(app.application_date)
          return appDate >= new Date(startDate) && appDate <= new Date(endDate)
        })
      }

    return filtered
      })
  }

  // 获取统计数据
  const fetchStatistics = async () => {
    statisticsLoading.value = true
    try {
      statistics.value = await JobApplicationAPI.getStatistics()
    } catch (error) {
      message.error('获取统计数据失败: ' + (error as Error).message)
    } finally {
      statisticsLoading.value = false
    }
  }

  return {
    // 状态
    applications,
    loading,
    currentApplication,
    statistics,
    statisticsLoading,
    
    // 计算属性
    totalCount,
    statusCounts,
    
    // 方法
    fetchApplications,
    fetchApplicationById,
    createApplication,
    updateApplication,
    deleteApplication,
    getFilteredApplications,
    fetchStatistics
  }
})