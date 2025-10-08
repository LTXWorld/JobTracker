import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useJobApplicationStore } from '../../src/stores/jobApplication'
import { JobApplicationAPI } from '../../src/api/jobApplication'
import type { JobApplication, CreateJobApplicationRequest, UpdateJobApplicationRequest, ApplicationStatus } from '../../src/types'

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

vi.mock('ant-design-vue', () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn()
  }
}))

describe('JobApplication Store', () => {
  const mockApplications: JobApplication[] = [
    {
      id: 1,
      company_name: 'Company A',
      position_title: 'Backend Engineer',
      application_date: '2024-01-01',
      status: '已投递' as ApplicationStatus,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
      reminder_enabled: false
    } as JobApplication,
    {
      id: 2,
      company_name: 'Company B',
      position_title: 'Frontend Engineer',
      application_date: '2024-01-02',
      status: '笔试中' as ApplicationStatus,
      created_at: '2024-01-02T00:00:00Z',
      updated_at: '2024-01-02T00:00:00Z',
      reminder_enabled: true,
      reminder_time: '2024-01-03T08:00:00Z',
      interview_type: '笔试'
    } as JobApplication
  ]

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('fetchApplications 会加载并格式化数据', async () => {
    vi.mocked(JobApplicationAPI.getAll).mockResolvedValueOnce(mockApplications)
    const store = useJobApplicationStore()

    await store.fetchApplications()

    expect(JobApplicationAPI.getAll).toHaveBeenCalled()
    expect(store.applications.length).toBe(2)
    expect(store.totalCount).toBe(2)
    // 笔试提醒会被归类为 written
    const writtenItem = store.applications.find(app => app.id === 2)
    expect((writtenItem as any).reminder_category).toBe('written')
  })

  it('createApplication 新增记录并插入列表', async () => {
    const store = useJobApplicationStore()
    const newAppRequest: CreateJobApplicationRequest = {
      company_name: 'Company C',
      position_title: 'DevOps',
      application_date: '2024-01-05',
      status: '已投递' as ApplicationStatus,
      company_attribute: '私企'
    }
    const createdApp: JobApplication = {
      id: 3,
      company_name: 'Company C',
      position_title: 'DevOps',
      application_date: '2024-01-05',
      status: '已投递' as ApplicationStatus,
      created_at: '2024-01-05T00:00:00Z',
      updated_at: '2024-01-05T00:00:00Z'
    } as JobApplication

    vi.mocked(JobApplicationAPI.create).mockResolvedValueOnce(createdApp)

    const result = await store.createApplication(newAppRequest)

    expect(JobApplicationAPI.create).toHaveBeenCalledWith(newAppRequest)
    expect(result).toEqual(createdApp)
    expect(store.applications[0]).toEqual(createdApp)
  })

  it('updateApplication 成功更新本地缓存', async () => {
    const store = useJobApplicationStore()
    store.applications = [...mockApplications]
    const updateRequest: UpdateJobApplicationRequest = {
      status: '笔试中' as ApplicationStatus,
      reminder_enabled: true,
      reminder_time: '2024-01-04T08:00:00Z'
    }
    const updatedApp: JobApplication = {
      ...mockApplications[0],
      status: '笔试中' as ApplicationStatus,
      reminder_enabled: true,
      reminder_time: '2024-01-04T08:00:00Z',
      updated_at: '2024-01-04T00:00:00Z'
    } as JobApplication

    vi.mocked(JobApplicationAPI.update).mockResolvedValueOnce(updatedApp)

    const result = await store.updateApplication(1, updateRequest)

    expect(JobApplicationAPI.update).toHaveBeenCalledWith(1, updateRequest)
    expect(result).toEqual(updatedApp)
    expect(store.applications[0]).toEqual(updatedApp)
  })

  it('deleteApplication 会移除记录', async () => {
    const store = useJobApplicationStore()
    store.applications = [...mockApplications]

    vi.mocked(JobApplicationAPI.delete).mockResolvedValueOnce(void 0)

    await store.deleteApplication(1)

    expect(JobApplicationAPI.delete).toHaveBeenCalledWith(1)
    expect(store.applications.find(app => app.id === 1)).toBeUndefined()
  })

  it('fetchApplicationById 会更新 currentApplication', async () => {
    const store = useJobApplicationStore()
    const target = mockApplications[0]
    vi.mocked(JobApplicationAPI.getById).mockResolvedValueOnce(target)

    const result = await store.fetchApplicationById(1)

    expect(JobApplicationAPI.getById).toHaveBeenCalledWith(1)
    expect(result).toEqual(target)
    expect(store.currentApplication).toEqual(target)
  })
})
