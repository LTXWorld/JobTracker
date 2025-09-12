import { defineStore } from 'pinia'
import { ref } from 'vue'
import { ResumeAPI, type SectionType } from '../api/resume'
import { message } from 'ant-design-vue'

export const useResumeStore = defineStore('resume', () => {
  const resume = ref<any>(null)
  const sections = ref<Record<string, any>>({})
  const attachments = ref<Array<any>>([])
  const loading = ref(false)
  const saving = ref(false)
  const lastSavedAt = ref<string>('')

  const fetchMyResume = async () => {
    loading.value = true
    try {
      const data = await ResumeAPI.getMy()
      resume.value = data.resume
      // 并发加载分区与附件，提升首屏速度
      await Promise.all([fetchSections(), fetchAttachments()])
    } finally {
      loading.value = false
    }
  }

  const fetchSections = async () => {
    if (!resume.value) return
    const list = await ResumeAPI.listSections(resume.value.id)
    const map: Record<string, any> = {}
    list.forEach((s: any) => { map[s.type] = s.content })
    sections.value = map
  }

  const fetchAttachments = async () => {
    if (!resume.value) return
    const list = await ResumeAPI.listAttachments(resume.value.id)
    attachments.value = Array.isArray(list) ? list : []
  }

  const upsertSection = async (type: SectionType, content: any) => {
    if (!resume.value) return
    saving.value = true
    try {
      await ResumeAPI.upsertSection(resume.value.id, type, content)
      // 本地状态先行更新
      sections.value[type] = content
      await refreshResumeMeta()
      lastSavedAt.value = new Date().toLocaleTimeString()
    } catch (e:any) {
      message.error('保存失败: ' + (e?.message || '未知错误'))
      throw e
    } finally { saving.value = false }
  }

  const uploadAttachment = async (file: File) => {
    if (!resume.value) return
    const data = await ResumeAPI.uploadAttachment(resume.value.id, file)
    // 更新附件列表（置顶最新）
    if (data?.attachment) {
      attachments.value = [{ ...data.attachment, url: data.url }, ...attachments.value]
    }
    await refreshResumeMeta()
    return data
  }

  // 刷新简历元信息（完善度等），统一调用入口
  const refreshResumeMeta = async () => {
    try {
      const data = await ResumeAPI.getMy()
      resume.value = data.resume
    } catch (_) { /* 静默失败 */ }
  }

  return { resume, sections, attachments, loading, saving, lastSavedAt, fetchMyResume, fetchSections, fetchAttachments, upsertSection, uploadAttachment, refreshResumeMeta }
})
