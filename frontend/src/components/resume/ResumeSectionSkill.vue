<template>
  <a-card title="技能与标签">
    <template #extra>
      <a-space>
        <a-button type="primary" size="small" @click="saveAll">保存</a-button>
      </a-space>
    </template>
    <draggable v-model="items" handle=".drag-handle" item-key="id">
      <template #item="{ element, index }">
        <a-card class="skill-item" size="small">
          <template #title>
            <span class="drag-handle">⋮⋮</span>
            {{ element.name || '技能' }} · {{ element.level || '熟练度' }}
          </template>
          <template #extra>
            <a-button size="small" danger @click="remove(index)">删除</a-button>
          </template>
          <a-form layout="vertical">
            <a-row :gutter="12">
              <a-col :span="10"><a-form-item label="技能名称"><a-input v-model:value="element.name" placeholder="如：Golang" /></a-form-item></a-col>
              <a-col :span="6">
                <a-form-item label="熟练度">
                  <a-select v-model:value="element.level" placeholder="选择">
                    <a-select-option value="掌握">掌握</a-select-option>
                    <a-select-option value="熟练">熟练</a-select-option>
                    <a-select-option value="精通">精通</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
              <a-col :span="8"><a-form-item label="年限(可选)"><a-input v-model:value="element.years" placeholder="如：2 年" /></a-form-item></a-col>
            </a-row>
            <a-form-item label="标签(逗号分隔)">
              <a-input v-model:value="element.tagsText" placeholder="如：微服务, RPC, 并发" />
            </a-form-item>
          </a-form>
        </a-card>
      </template>
    </draggable>
    <a-button style="margin-top:8px" @click="add">新增技能</a-button>
  </a-card>
</template>

<script setup lang="ts">
import draggable from 'vuedraggable'
import { reactive, watch } from 'vue'

interface SkillItem { id:string; name?:string; level?:string; years?:string; tags?:string[]; tagsText?:string }

const props = defineProps<{ modelValue?: any }>()
const emit = defineEmits<{'update:modelValue':[any],'save':[any]}>()

const items = reactive<Array<SkillItem>>([])

watch(() => props.modelValue, (v) => {
  const arr = (v?.items || [])
  items.splice(0, items.length, ...arr.map((it:any, idx:number)=>({ id:`${Date.now()}_${idx}`, ...it, tagsText:(it.tags||[]).join(', ') })))
}, { immediate: true })

const add = () => items.push({ id:`${Date.now()}` })
const remove = (idx:number) => items.splice(idx,1)

const saveAll = () => {
  const payload = { items: items.map(({id,tagsText,...rest})=>({ ...rest, tags: (tagsText||'').split(/[,，]/).map(s=>s.trim()).filter(Boolean) })) }
  emit('save', payload)
}
</script>

<style scoped>
.skill-item{ margin-bottom:12px }
.drag-handle{ cursor:grab; margin-right:8px; color:#999 }
</style>

