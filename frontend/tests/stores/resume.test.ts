import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useResumeStore } from '../../src/stores/resume'
import { ResumeAPI } from '../../src/api/resume'

// Mock ResumeAPI
vi.mock('../../src/api/resume', () => ({
  ResumeAPI: {
    getMy: vi.fn(),
    listSections: vi.fn(),
    listAttachments: vi.fn(),
    upsertSection: vi.fn(),
    uploadAttachment: vi.fn()
  }
}))

// Mock ant-design-vue message，避免测试时产生实际提示
vi.mock('ant-design-vue', () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn()
  }
}))

describe('ResumeStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('fetchMyResume 应并发加载分区与附件并填充状态', async () => {
    const store = useResumeStore()

    // Arrange mocks
    vi.mocked(ResumeAPI.getMy).mockResolvedValueOnce({
      resume: { id: 1, title: '默认简历', completeness: 20 }
    })
    vi.mocked(ResumeAPI.listSections).mockResolvedValueOnce([
      { type: 'base', content: { name: '张三', phone: '13800000000', email: 'a@b.com' } },
      { type: 'exp', content: { items: [] } }
    ])
    vi.mocked(ResumeAPI.listAttachments).mockResolvedValueOnce([
      { id: 10, file_name: 'cv.pdf', url: 'http://localhost:8010/static/resumes/1/1/cv.pdf' }
    ])

    // Act
    await store.fetchMyResume()

    // Assert
    expect(ResumeAPI.getMy).toHaveBeenCalledTimes(1)
    expect(ResumeAPI.listSections).toHaveBeenCalledWith(1)
    expect(ResumeAPI.listAttachments).toHaveBeenCalledWith(1)
    expect(store.resume?.id).toBe(1)
    expect(store.sections['base']).toBeTruthy()
    expect(store.attachments.length).toBe(1)
  })

  it('upsertSection 后应更新本地分区并刷新完善度', async () => {
    const store = useResumeStore()
    // 初始化简历元信息
    store.resume = { id: 1, title: '默认简历', completeness: 20 } as any

    // Arrange mocks
    vi.mocked(ResumeAPI.upsertSection).mockResolvedValueOnce({})
    vi.mocked(ResumeAPI.getMy).mockResolvedValueOnce({
      resume: { id: 1, title: '默认简历', completeness: 35 }
    })

    // Act
    await store.upsertSection('exp', { items: [{ company: 'ABC', position: 'Intern' }] })

    // Assert
    expect(ResumeAPI.upsertSection).toHaveBeenCalledWith(1, 'exp', { items: [{ company: 'ABC', position: 'Intern' }] })
    expect(store.sections['exp']).toBeTruthy()
    expect(store.resume?.completeness).toBe(35)
  })

  it('uploadAttachment 后应把最新附件置顶并刷新完善度', async () => {
    const store = useResumeStore()
    store.resume = { id: 1, title: '默认简历', completeness: 20 } as any

    const fakeFile = new File([new Uint8Array([1, 2, 3])], 'cv.pdf', { type: 'application/pdf' })

    vi.mocked(ResumeAPI.uploadAttachment).mockResolvedValueOnce({
      attachment: { id: 101, file_name: 'cv.pdf' },
      url: 'http://localhost:8010/static/resumes/1/1/cv.pdf'
    })
    vi.mocked(ResumeAPI.getMy).mockResolvedValueOnce({
      resume: { id: 1, title: '默认简历', completeness: 25 }
    })

    await store.uploadAttachment(fakeFile)

    expect(ResumeAPI.uploadAttachment).toHaveBeenCalled()
    expect(store.attachments.length).toBe(1)
    expect(store.attachments[0].url).toContain('/static/resumes/1/1/cv.pdf')
    expect(store.resume?.completeness).toBe(25)
  })
})

