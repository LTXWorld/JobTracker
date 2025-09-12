<template>
  <div class="resume-page">
    <div class="header">
      <h2>{{ resume?.title || '我的简历' }}</h2>
      <div class="meta">
        <span>完善度：<b>{{ resume?.completeness || 0 }}%</b></span>
        <span v-if="lastSavedAt">已保存于 {{ lastSavedAt }}</span>
      </div>
    </div>

    <a-row :gutter="16">
      <a-col :xs="24" :md="6">
        <a-menu v-model:selectedKeys="selected" mode="inline">
          <a-menu-item key="base">基本信息</a-menu-item>
          <a-menu-item key="intent">求职意向</a-menu-item>
          <a-menu-item key="edu">教育经历</a-menu-item>
          <a-menu-item key="exp">工作/实习</a-menu-item>
          <a-menu-item key="project">项目经历</a-menu-item>
          <a-menu-item key="skill">技能与证书</a-menu-item>
          <a-menu-item key="cert">证书</a-menu-item>
          <a-menu-item key="summary">自我评价</a-menu-item>
          <a-menu-item key="attachment">附件简历</a-menu-item>
        </a-menu>
      </a-col>
      <a-col :xs="24" :md="18">
        <a-card v-if="selected[0]==='base'" title="基本信息">
          <a-form layout="vertical">
            <a-form-item label="姓名">
              <a-input v-model:value="base.name" placeholder="姓名" />
            </a-form-item>
            <a-form-item label="手机号">
              <a-input v-model:value="base.phone" placeholder="手机号" />
            </a-form-item>
            <a-form-item label="邮箱">
              <a-input v-model:value="base.email" placeholder="邮箱" />
            </a-form-item>
            <a-form-item label="所在城市">
              <a-input v-model:value="base.city" placeholder="城市" />
            </a-form-item>
            <a-button type="primary" :loading="saving" @click="save('base', base)">保存</a-button>
          </a-form>
        </a-card>

        <a-card v-else-if="selected[0]==='intent'" title="求职意向">
          <a-form layout="vertical">
            <a-form-item label="期望职位">
              <a-input v-model:value="intent.position" placeholder="如：Golang 开发" />
            </a-form-item>
            <a-form-item label="期望城市">
              <a-input v-model:value="intent.city" placeholder="如：上海" />
            </a-form-item>
            <a-form-item label="期望薪资">
              <a-input v-model:value="intent.salary" placeholder="如：20-30K" />
            </a-form-item>
            <a-button type="primary" :loading="saving" @click="save('intent', intent)">保存</a-button>
          </a-form>
        </a-card>

        <ResumeSectionEdu v-else-if="selected[0]==='edu'" :model-value="sections['edu']" @save="(p:any)=>save('edu', p)" />

        <ResumeSectionExp v-else-if="selected[0]==='exp'" :model-value="sections['exp']" @save="(p:any)=>save('exp', p)" />

        <ResumeSectionProject v-else-if="selected[0]==='project'" :model-value="sections['project']" @save="(p:any)=>save('project', p)" />

        <ResumeSectionSkill v-else-if="selected[0]==='skill'" :model-value="sections['skill']" @save="(p:any)=>save('skill', p)" />

        <ResumeSectionCert v-else-if="selected[0]==='cert'" :model-value="sections['cert']" @save="(p:any)=>save('cert', p)" />

        <ResumeSectionSummary v-else-if="selected[0]==='summary'" :model-value="sections['summary']" @save="(p:any)=>save('summary', p)" />

        <a-card v-else-if="selected[0]==='attachment'" title="附件简历">
          <a-upload :show-upload-list="false" :before-upload="beforeUpload" :custom-request="onUpload">
            <a-button>上传 PDF</a-button>
          </a-upload>
          <div v-if="uploadUrl" class="upload-tip">已上传：<a :href="uploadUrl" target="_blank">点击查看</a></div>
        </a-card>

        <a-result v-else title="即将到来">
          <template #subTitle>本区块将在后续迭代中补充详细表单。现在可先填写基本信息与求职意向。</template>
        </a-result>
      </a-col>
    </a-row>
  </div>
 </template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useResumeStore } from '../stores/resume'
import { message } from 'ant-design-vue'
import ResumeSectionEdu from '../components/resume/ResumeSectionEdu.vue'
import ResumeSectionExp from '../components/resume/ResumeSectionExp.vue'
import ResumeSectionProject from '../components/resume/ResumeSectionProject.vue'
import ResumeSectionSkill from '../components/resume/ResumeSectionSkill.vue'
import ResumeSectionCert from '../components/resume/ResumeSectionCert.vue'
import ResumeSectionSummary from '../components/resume/ResumeSectionSummary.vue'

const store = useResumeStore()
// 关键修复：使用 storeToRefs 确保刷新后仍保持响应性
const { resume, sections, saving, lastSavedAt, attachments } = storeToRefs(store)
const selected = ref<string[]>(['base'])

const base = reactive<any>({ name: '', phone: '', email: '', city: '' })
const intent = reactive<any>({ position: '', city: '', salary: '' })
const uploadUrl = ref<string>('')

onMounted(async () => {
  await store.fetchMyResume()
  // 初始化表单（注意 sections 是 ref，需要取 value）
  Object.assign(base, (sections.value && sections.value['base']) || {})
  Object.assign(intent, (sections.value && sections.value['intent']) || {})
  // 初始化附件显示：取最新一条附件的 URL
  if (attachments.value && attachments.value.length > 0) {
    uploadUrl.value = attachments.value[0].url || ''
  }
})

watch(() => sections.value && sections.value['base'], (v) => { if (v) Object.assign(base, v) })
watch(() => sections.value && sections.value['intent'], (v) => { if (v) Object.assign(intent, v) })

const save = async (type: any, payload: any) => {
  await store.upsertSection(type, payload)
  message.success('已保存')
}

const beforeUpload = (file: File) => {
  if (file.type !== 'application/pdf') { message.error('仅支持 PDF'); return false }
  if (file.size/1024/1024 > 5) { message.error('文件不能超过 5MB'); return false }
  return true
}

const onUpload = async (options: any) => {
  try {
    const res = await store.uploadAttachment(options.file)
    uploadUrl.value = res.url
    options.onSuccess({}, options.file)
  } catch (e:any) {
    options.onError(e)
  }
}

// 监听附件列表变化，保持链接为最新上传
watch(() => attachments.value && attachments.value[0]?.url, (u) => {
  if (u) uploadUrl.value = u
})
</script>

<style scoped>
.resume-page { padding: 24px; }
.header { display:flex; align-items:center; justify-content:space-between; margin-bottom:16px; }
.meta { color:#888; display:flex; gap:16px; }
.upload-tip { margin-top: 12px; }
</style>
