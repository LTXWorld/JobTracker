<template>
  <div class="detail-page">
    <a-card title="投递详情" :loading="loading">
      <template #extra>
        <a-button type="link" @click="openEdit">编辑</a-button>
      </template>

      <a-descriptions bordered :column="1">
        <a-descriptions-item label="公司名称">{{ app?.company_name }}</a-descriptions-item>
        <a-descriptions-item label="职位标题">{{ app?.position_title }}</a-descriptions-item>
        <a-descriptions-item label="投递日期">{{ app?.application_date }}</a-descriptions-item>
        <a-descriptions-item label="当前状态">{{ app?.status }}</a-descriptions-item>
        <a-descriptions-item label="企业属性">
          <template v-if="app?.company_attribute">{{ app?.company_attribute }}</template>
          <template v-else>
            <a-tag color="orange">未填写</a-tag>
            <a-button type="link" @click="openEdit">去完善</a-button>
          </template>
        </a-descriptions-item>
        <a-descriptions-item label="工作地点">{{ app?.work_location || '-' }}</a-descriptions-item>
        <a-descriptions-item label="薪资范围">{{ app?.salary_range || '-' }}</a-descriptions-item>
        <a-descriptions-item label="备注">{{ app?.notes || '-' }}</a-descriptions-item>
        <a-descriptions-item label="创建时间">{{ app?.created_at }}</a-descriptions-item>
        <a-descriptions-item label="更新时间">{{ app?.updated_at }}</a-descriptions-item>
      </a-descriptions>
    </a-card>

    <ApplicationForm v-model:visible="showEditor" :initialData="app" @success="handleUpdated" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useJobApplicationStore } from '../stores/jobApplication'
import ApplicationForm from '../components/ApplicationForm.vue'

const route = useRoute()
const jobStore = useJobApplicationStore()
const loading = ref(false)
const app = ref<any>(null)
const showEditor = ref(false)

const fetch = async () => {
  loading.value = true
  try {
    const id = Number(route.params.id)
    app.value = await jobStore.fetchApplicationById(id)
  } finally {
    loading.value = false
  }
}

onMounted(fetch)

const openEdit = () => {
  showEditor.value = true
}
const handleUpdated = async () => {
  showEditor.value = false
  await fetch()
}
</script>

<style scoped>
.detail-page {
  padding: 16px;
}
</style>
