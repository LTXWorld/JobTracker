import request from './request'

export type SectionType = 'base'|'intent'|'edu'|'exp'|'project'|'skill'|'cert'|'honor'|'summary'|'links'

export const ResumeAPI = {
  async getMy() {
    const res = await request.get('/api/v1/resumes/me')
    return res.data.data
  },
  async getById(id: number) {
    const res = await request.get(`/api/v1/resumes/${id}`)
    return res.data.data
  },
  async create() {
    const res = await request.post('/api/v1/resumes')
    return res.data.data
  },
  async updateMeta(id: number, payload: { title?: string; privacy?: string }) {
    const res = await request.put(`/api/v1/resumes/${id}`, payload)
    return res.data.data
  },
  async listSections(id: number) {
    const res = await request.get(`/api/v1/resumes/${id}/sections`)
    return res.data.data
  },
  async upsertSection(id: number, type: SectionType, content: any) {
    const res = await request.put(`/api/v1/resumes/${id}/sections/${type}`, content)
    return res.data.data
  },
  async uploadAttachment(id: number, file: File) {
    const form = new FormData()
    form.append('file', file)
    const res = await request.post(`/api/v1/resumes/${id}/attachments`, form, { headers: { 'Content-Type': 'multipart/form-data' } })
    return res.data.data
  },
  async listAttachments(id: number) {
    const res = await request.get(`/api/v1/resumes/${id}/attachments`)
    return res.data.data
  }
}
