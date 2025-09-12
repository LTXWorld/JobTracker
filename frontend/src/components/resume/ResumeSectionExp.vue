<template>
  <a-card title="工作/实习经历">
    <template #extra>
      <a-space>
        <a-button type="primary" size="small" @click="saveAll">保存</a-button>
      </a-space>
    </template>
    <draggable v-model="items" handle=".drag-handle" item-key="id">
      <template #item="{ element, index }">
        <a-card class="exp-item" size="small">
          <template #title>
            <span class="drag-handle">⋮⋮</span>
            {{ element.company || '未命名公司' }} · {{ element.position || '职位' }}
          </template>
          <template #extra>
            <a-space>
              <a-button size="small" @click="remove(index)" danger>删除</a-button>
            </a-space>
          </template>
          <a-form layout="vertical">
            <a-row :gutter="12">
              <a-col :span="12"><a-form-item label="公司"><a-input v-model:value="element.company" /></a-form-item></a-col>
              <a-col :span="12"><a-form-item label="部门"><a-input v-model:value="element.department" /></a-form-item></a-col>
            </a-row>
            <a-row :gutter="12">
              <a-col :span="12"><a-form-item label="职位"><a-input v-model:value="element.position" /></a-form-item></a-col>
              <a-col :span="6"><a-form-item label="开始"><a-input v-model:value="element.from" placeholder="2023-06" /></a-form-item></a-col>
              <a-col :span="6"><a-form-item label="结束"><a-input v-model:value="element.to" placeholder="2023-09" /></a-form-item></a-col>
            </a-row>
            <a-form-item label="职责/业绩要点（每行一条）">
              <a-textarea v-model:value="element.highlightsText" :rows="3" placeholder="示例：将服务 QPS 提升 30%" />
            </a-form-item>
          </a-form>
        </a-card>
      </template>
    </draggable>
    <a-button style="margin-top:8px" @click="add">新增经历</a-button>
  </a-card>
 </template>

<script setup lang="ts">
import draggable from 'vuedraggable'
import { reactive, watch } from 'vue'

interface ExpItem { id:string; company?:string; department?:string; position?:string; from?:string; to?:string; highlights?:string[]; highlightsText?:string }

const props = defineProps<{ modelValue?: any }>()
const emit = defineEmits<{'update:modelValue':[any],'save':[any]}>()

const items = reactive<Array<ExpItem>>([])

watch(() => props.modelValue, (v) => {
  const arr = (v?.items || [])
  items.splice(0, items.length, ...arr.map((it:any, idx:number)=>({ id: `${Date.now()}_${idx}`, ...it, highlightsText: (it.highlights||[]).join('\n') })))
}, { immediate: true })

const add = () => { items.push({ id:`${Date.now()}`, highlightsText:'' }) }
const remove = (idx:number) => { items.splice(idx,1) }

const saveAll = () => {
  const payload = { items: items.map(({id,highlightsText,...rest})=>({ ...rest, highlights: (highlightsText||'').split(/\n+/).filter(Boolean) })) }
  emit('save', payload)
}
</script>

<style scoped>
.exp-item{ margin-bottom:12px }
.drag-handle{ cursor:grab; margin-right:8px; color:#999 }
</style>
