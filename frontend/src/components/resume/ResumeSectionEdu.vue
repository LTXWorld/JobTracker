<template>
  <a-card title="教育经历">
    <template #extra>
      <a-space>
        <a-button type="primary" size="small" @click="saveEdu">保存</a-button>
      </a-space>
    </template>
    <draggable v-model="items" handle=".drag-handle" item-key="id">
      <template #item="{ element, index }">
        <a-card class="edu-item" size="small">
          <template #title>
            <span class="drag-handle">⋮⋮</span>
            {{ element.school || '未命名学校' }} · {{ element.major || '专业' }}
          </template>
          <template #extra>
            <a-space>
              <a-button size="small" @click="remove(index)" danger>删除</a-button>
            </a-space>
          </template>
          <a-form layout="vertical">
            <a-row :gutter="12">
              <a-col :span="12">
                <a-form-item label="学校">
                  <a-input v-model:value="element.school" placeholder="学校名称" />
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item label="专业">
                  <a-input v-model:value="element.major" placeholder="专业" />
                </a-form-item>
              </a-col>
            </a-row>
            <a-row :gutter="12">
              <a-col :span="8">
                <a-form-item label="学历">
                  <a-input v-model:value="element.degree" placeholder="本科/硕士" />
                </a-form-item>
              </a-col>
              <a-col :span="8">
                <a-form-item label="开始时间">
                  <a-input v-model:value="element.from" placeholder="2021-09" />
                </a-form-item>
              </a-col>
              <a-col :span="8">
                <a-form-item label="结束时间">
                  <a-input v-model:value="element.to" placeholder="2025-06" />
                </a-form-item>
              </a-col>
            </a-row>
            <a-row :gutter="12">
              <a-col :span="12">
                <a-form-item label="GPA/排名(可选)">
                  <a-input v-model:value="element.gpa" placeholder="如 3.8/4.0 或 前10%" />
                </a-form-item>
              </a-col>
            </a-row>
          </a-form>
        </a-card>
      </template>
    </draggable>
    <a-button style="margin-top:8px" @click="add">新增教育经历</a-button>
  </a-card>
 </template>

<script setup lang="ts">
import draggable from 'vuedraggable'
import { reactive, watch } from 'vue'

interface EduItem { id: string; school?: string; major?: string; degree?: string; from?: string; to?: string; gpa?: string }

const props = defineProps<{ modelValue?: any }>()
const emit = defineEmits<{'update:modelValue':[any],'save':[any]}>()

const items = reactive<Array<EduItem>>([])

watch(() => props.modelValue, (v) => {
  items.splice(0, items.length, ...((v?.items || []).map((it:any, idx:number)=>({ id: `${Date.now()}_${idx}`, ...it }))))
}, { immediate: true })

const add = () => { items.push({ id: `${Date.now()}` }) }
const remove = (idx:number) => { items.splice(idx,1) }

const saveEdu = () => {
  emit('save', { items: items.map(({ id, ...rest }) => rest) })
}
</script>

<style scoped>
.edu-item{ margin-bottom:12px }
.drag-handle{ cursor:grab; margin-right:8px; color:#999 }
</style>
