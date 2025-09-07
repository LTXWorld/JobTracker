import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useJobApplicationStore } from '../../src/stores/jobApplication'
import { JobApplicationAPI } from '../../src/api/jobApplication'
import type { 
  JobApplication, 
  CreateJobApplicationRequest, 
  UpdateJobApplicationRequest,
  ApplicationStatus,
  JobApplicationStatistics
} from '../../src/types'

// Mock JobApplicationAPI
vi.mock('../../src/api/jobApplication', () => ({
  JobApplicationAPI: {
    getAll: vi.fn(),
    getById: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    getStatistics: vi.fn()
  }
}))

// Mock ant-design-vue message
vi.mock('ant-design-vue', () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn()
  }
}))

describe('JobApplicationStore', () => {
  let mockApplications: JobApplication[]

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()

    // 创建模拟数据
    mockApplications = [
      {
        id: 1,
        company: 'Google',
        position: 'Software Engineer',
        status: 'applied' as ApplicationStatus,
        applied_date: '2024-01-01',
        notes: 'Initial application',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        user_id: 1
      },
      {
        id: 2,
        company: 'Microsoft',
        position: 'Frontend Developer',
        status: 'interview' as ApplicationStatus,
        applied_date: '2024-01-02',
        notes: 'Phone interview scheduled',
        created_at: '2024-01-02T00:00:00Z',
        updated_at: '2024-01-02T00:00:00Z',
        user_id: 1
      },
      {
        id: 3,
        company: 'Meta',
        position: 'Full Stack Developer',
        status: 'rejected' as ApplicationStatus,
        applied_date: '2024-01-03',
        notes: 'Not a good fit',
        created_at: '2024-01-03T00:00:00Z',
        updated_at: '2024-01-03T00:00:00Z',
        user_id: 1
      }
    ]
  })

  describe('初始状态', () => {
    it('应该有正确的初始状态', () => {
      const store = useJobApplicationStore()
      
      expect(store.applications).toEqual([])
      expect(store.loading).toBe(false)
      expect(store.currentApplication).toBe(null)
      expect(store.statistics).toBe(null)
      expect(store.statisticsLoading).toBe(false)
      expect(store.totalCount).toBe(0)
    })
  })

  describe('获取投递记录', () => {
    it('fetchApplications 应该成功获取所有记录', async () => {
      const store = useJobApplicationStore()
      
      vi.mocked(JobApplicationAPI.getAll).mockResolvedValueOnce(mockApplications)

      await store.fetchApplications()

      expect(JobApplicationAPI.getAll).toHaveBeenCalledOnce()
      expect(store.applications).toEqual(mockApplications)
      expect(store.loading).toBe(false)
      expect(store.totalCount).toBe(3)
    })

    it('fetchApplications 处理API错误时应该显示错误信息', async () => {
      const store = useJobApplicationStore()
      
      const errorMessage = '网络请求失败'
      vi.mocked(JobApplicationAPI.getAll).mockRejectedValueOnce(new Error(errorMessage))

      await store.fetchApplications()

      expect(store.applications).toEqual([])
      expect(store.loading).toBe(false)
    })

    it('fetchApplicationById 应该成功获取单个记录', async () => {
      const store = useJobApplicationStore()
      const targetApp = mockApplications[0]
      
      vi.mocked(JobApplicationAPI.getById).mockResolvedValueOnce(targetApp)

      await store.fetchApplicationById(1)

      expect(JobApplicationAPI.getById).toHaveBeenCalledWith(1)
      expect(store.currentApplication).toEqual(targetApp)
      expect(store.loading).toBe(false)
    })

    it('fetchApplicationById 处理不存在的记录', async () => {
      const store = useJobApplicationStore()
      
      vi.mocked(JobApplicationAPI.getById).mockRejectedValueOnce(new Error('记录不存在'))

      await store.fetchApplicationById(999)

      expect(store.currentApplication).toBe(null)
      expect(store.loading).toBe(false)
    })
  })

  describe('创建投递记录', () => {
    it('createApplication 应该成功创建新记录', async () => {
      const store = useJobApplicationStore()
      
      const newApplicationData: CreateJobApplicationRequest = {
        company: 'Amazon',
        position: 'Backend Engineer',
        status: 'applied' as ApplicationStatus,
        applied_date: '2024-01-04',
        notes: 'Excited about this opportunity'
      }

      const createdApplication: JobApplication = {
        id: 4,
        ...newApplicationData,
        created_at: '2024-01-04T00:00:00Z',
        updated_at: '2024-01-04T00:00:00Z',
        user_id: 1
      }

      vi.mocked(JobApplicationAPI.create).mockResolvedValueOnce(createdApplication)

      const result = await store.createApplication(newApplicationData)

      expect(result).toBe(true)
      expect(JobApplicationAPI.create).toHaveBeenCalledWith(newApplicationData)
      expect(store.applications).toContain(createdApplication)
      expect(store.totalCount).toBe(1) // 初始store是空的，添加一个后为1
    })

    it('createApplication 处理创建失败', async () => {
      const store = useJobApplicationStore()
      
      const newApplicationData: CreateJobApplicationRequest = {
        company: 'Amazon',
        position: 'Backend Engineer',
        status: 'applied' as ApplicationStatus,
        applied_date: '2024-01-04',
        notes: 'Test application'
      }

      vi.mocked(JobApplicationAPI.create).mockRejectedValueOnce(new Error('创建失败'))

      const result = await store.createApplication(newApplicationData)

      expect(result).toBe(false)
      expect(store.applications).toEqual([])
    })
  })

  describe('更新投递记录', () => {
    it('updateApplication 应该成功更新记录', async () => {
      const store = useJobApplicationStore()
      store.applications = [...mockApplications] // 设置初始数据
      
      const updateData: UpdateJobApplicationRequest = {
        status: 'interview' as ApplicationStatus,
        notes: 'Updated notes'
      }

      const updatedApplication: JobApplication = {
        ...mockApplications[0],
        ...updateData,
        updated_at: '2024-01-05T00:00:00Z'
      }

      vi.mocked(JobApplicationAPI.update).mockResolvedValueOnce(updatedApplication)

      const result = await store.updateApplication(1, updateData)

      expect(result).toBe(true)
      expect(JobApplicationAPI.update).toHaveBeenCalledWith(1, updateData)
      
      // 验证store中的记录已更新
      const updatedInStore = store.applications.find(app => app.id === 1)
      expect(updatedInStore?.status).toBe('interview')
      expect(updatedInStore?.notes).toBe('Updated notes')
    })

    it('updateApplication 处理不存在的记录', async () => {
      const store = useJobApplicationStore()
      store.applications = [...mockApplications]
      
      vi.mocked(JobApplicationAPI.update).mockRejectedValueOnce(new Error('记录不存在'))

      const result = await store.updateApplication(999, { notes: 'test' })

      expect(result).toBe(false)
    })
  })

  describe('删除投递记录', () => {
    it('deleteApplication 应该成功删除记录', async () => {
      const store = useJobApplicationStore()
      store.applications = [...mockApplications]
      
      vi.mocked(JobApplicationAPI.delete).mockResolvedValueOnce()

      const result = await store.deleteApplication(1)

      expect(result).toBe(true)
      expect(JobApplicationAPI.delete).toHaveBeenCalledWith(1)
      
      // 验证记录已从store中移除
      expect(store.applications.find(app => app.id === 1)).toBeUndefined()
      expect(store.totalCount).toBe(2) // 原来3个，删除1个后剩2个
    })

    it('deleteApplication 处理删除失败', async () => {
      const store = useJobApplicationStore()
      store.applications = [...mockApplications]
      const originalLength = mockApplications.length
      
      vi.mocked(JobApplicationAPI.delete).mockRejectedValueOnce(new Error('删除失败'))

      const result = await store.deleteApplication(1)

      expect(result).toBe(false)
      expect(store.applications.length).toBe(originalLength) // 没有删除
    })
  })

  describe('统计功能', () => {
    it('fetchStatistics 应该成功获取统计数据', async () => {
      const store = useJobApplicationStore()
      
      const mockStatistics: JobApplicationStatistics = {
        total: 10,
        applied: 5,
        interview: 3,
        offer: 1,
        rejected: 1,
        recent_applications: 3
      }

      vi.mocked(JobApplicationAPI.getStatistics).mockResolvedValueOnce(mockStatistics)

      await store.fetchStatistics()

      expect(JobApplicationAPI.getStatistics).toHaveBeenCalledOnce()
      expect(store.statistics).toEqual(mockStatistics)
      expect(store.statisticsLoading).toBe(false)
    })

    it('statusCounts 计算属性应该正确统计状态数量', () => {
      const store = useJobApplicationStore()
      store.applications = [...mockApplications]

      const counts = store.statusCounts

      expect(counts.applied).toBe(1)
      expect(counts.interview).toBe(1)
      expect(counts.rejected).toBe(1)
      expect(counts.offer).toBeUndefined() // 没有这个状态的记录
    })

    it('totalCount 计算属性应该返回正确的总数', () => {
      const store = useJobApplicationStore()
      
      expect(store.totalCount).toBe(0)
      
      store.applications = [...mockApplications]
      expect(store.totalCount).toBe(3)
    })
  })

  describe('筛选和搜索', () => {
    it('应该正确筛选特定状态的记录', () => {
      const store = useJobApplicationStore()
      store.applications = [...mockApplications]

      const interviewApps = store.applications.filter(app => app.status === 'interview')
      expect(interviewApps.length).toBe(1)
      expect(interviewApps[0].company).toBe('Microsoft')
    })

    it('应该正确筛选特定公司的记录', () => {
      const store = useJobApplicationStore()
      store.applications = [...mockApplications]

      const googleApps = store.applications.filter(app => app.company === 'Google')
      expect(googleApps.length).toBe(1)
      expect(googleApps[0].position).toBe('Software Engineer')
    })
  })

  describe('loading状态管理', () => {
    it('API调用期间应该设置loading状态', async () => {
      const store = useJobApplicationStore()
      
      let loadingDuringCall = false
      
      vi.mocked(JobApplicationAPI.getAll).mockImplementationOnce(() => {
        loadingDuringCall = store.loading
        return Promise.resolve(mockApplications)
      })

      await store.fetchApplications()

      expect(loadingDuringCall).toBe(true)
      expect(store.loading).toBe(false) // 调用完成后应该重置
    })

    it('statisticsLoading应该在统计API调用期间设置', async () => {
      const store = useJobApplicationStore()
      
      let statisticsLoadingDuringCall = false
      
      vi.mocked(JobApplicationAPI.getStatistics).mockImplementationOnce(() => {
        statisticsLoadingDuringCall = store.statisticsLoading
        return Promise.resolve({
          total: 5,
          applied: 2,
          interview: 2,
          offer: 1,
          rejected: 0,
          recent_applications: 1
        })
      })

      await store.fetchStatistics()

      expect(statisticsLoadingDuringCall).toBe(true)
      expect(store.statisticsLoading).toBe(false)
    })
  })

  describe('错误处理', () => {
    it('应该正确处理网络错误', async () => {
      const store = useJobApplicationStore()
      
      const networkError = new Error('网络连接超时')
      vi.mocked(JobApplicationAPI.getAll).mockRejectedValueOnce(networkError)

      await store.fetchApplications()

      expect(store.applications).toEqual([])
      expect(store.loading).toBe(false)
    })

    it('应该正确处理服务器错误', async () => {
      const store = useJobApplicationStore()
      
      const serverError = new Error('服务器内部错误')
      vi.mocked(JobApplicationAPI.create).mockRejectedValueOnce(serverError)

      const result = await store.createApplication({
        company: 'Test Company',
        position: 'Test Position',
        status: 'applied' as ApplicationStatus,
        applied_date: '2024-01-01',
        notes: 'Test'
      })

      expect(result).toBe(false)
    })
  })

  describe('数据一致性', () => {
    it('创建记录后应该保持数据一致性', async () => {
      const store = useJobApplicationStore()
      const initialCount = store.totalCount

      const newApp: JobApplication = {
        id: 10,
        company: 'Netflix',
        position: 'DevOps Engineer',
        status: 'applied' as ApplicationStatus,
        applied_date: '2024-01-10',
        notes: 'Streaming technology',
        created_at: '2024-01-10T00:00:00Z',
        updated_at: '2024-01-10T00:00:00Z',
        user_id: 1
      }

      vi.mocked(JobApplicationAPI.create).mockResolvedValueOnce(newApp)

      await store.createApplication({
        company: newApp.company,
        position: newApp.position,
        status: newApp.status,
        applied_date: newApp.applied_date,
        notes: newApp.notes
      })

      expect(store.totalCount).toBe(initialCount + 1)
      expect(store.applications).toContain(newApp)
    })

    it('删除记录后应该保持数据一致性', async () => {
      const store = useJobApplicationStore()
      store.applications = [...mockApplications]
      const initialCount = store.totalCount

      vi.mocked(JobApplicationAPI.delete).mockResolvedValueOnce()

      await store.deleteApplication(1)

      expect(store.totalCount).toBe(initialCount - 1)
      expect(store.applications.find(app => app.id === 1)).toBeUndefined()
    })
  })
})