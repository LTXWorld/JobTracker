<template>
  <a-card title="证书与资质">
    <template #extra>
      <a-space>
        <a-button type="primary" size="small" @click="saveAll">保存</a-button>
      </a-space>
    </template>
    <draggable v-model="items" handle=".drag-handle" item-key="id">
      <template #item="{ element, index }">
        <a-card class="cert-item" size="small">
          <template #title>
            <span class="drag-handle">⋮⋮</span>
            {{ element.name || '证书' }} · {{ element.issuer || '机构' }}
          </template>
          <template #extra>
            <a-button size="small" danger @click="remove(index)">删除</a-button>
          </template>
          <a-form layout="vertical">
            <a-row :gutter="12">
              <a-col :span="12"><a-form-item label="证书名称"><a-input v-model:value="element.name" /></a-form-item></a-col>
              <a-col :span="12"><a-form-item label="颁发机构"><a-input v-model:value="element.issuer" /></a-form-item></a-col>
            </a-row>
            <a-row :gutter="12">
              <a-col :span="8"><a-form-item label="日期"><a-input v-model:value="element.date" placeholder="2024-06"/></a-form-item></a-col>
              <a-col :span="8"><a-form-item label="证书编号"><a-input v-model:value="element.code" /></a-form-item></a-col>
              <a-col :span="8"><a-form-item label="链接(可选)"><a-input v-model:value="element.link" placeholder="https://..." /></a-form-item></a-col>
            </a-row>
          </a-form>
        </a-card>
      </template>
    </draggable>
    <a-button style="margin-top:8px" @click="add">新增证书</a-button>
  </a-card>
</template>

<script setup lang="ts">
import draggable from 'vuedraggable'
import { reactive, watch } from 'vue'

interface CertItem { id:string; name?:string; issuer?:string; date?:string; code?:string; link?:string }

const props = defineProps<{ modelValue?: any }>()
const emit = defineEmits<{'update:modelValue':[any],'save':[any]}>()

const items = reactive<Array<CertItem>>([])

watch(() => props.modelValue, (v) => {
  const arr = (v?.items || [])
  items.splice(0, items.length, ...arr.map((it:any, idx:number)=>({ id:`${Date.now()}_${idx}`, ...it })))
}, { immediate: true })

const add = () => items.push({ id:`${Date.now()}` })
const remove = (idx:number) => items.splice(idx,1)

const saveAll = () => {
  const payload = { items: items.map(({id, ...rest})=>rest) }
  emit('save', payload)
}
</script>

<style scoped>
.cert-item{ margin-bottom:12px }
.drag-handle{ cursor:grab; margin-right:8px; color:#999 }
</style>

