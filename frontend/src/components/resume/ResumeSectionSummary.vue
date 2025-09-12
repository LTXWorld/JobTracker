<template>
  <a-card title="自我评价">
    <template #extra>
      <a-space>
        <span style="color:#999">{{ wordCount }}/500</span>
        <a-button type="primary" size="small" @click="save">保存</a-button>
      </a-space>
    </template>
    <a-form layout="vertical">
      <a-form-item label="自我评价/个人简介">
        <a-textarea v-model:value="text" :rows="6" :maxlength="500" show-count placeholder="建议概括优势与亮点，突出量化成果" />
      </a-form-item>
    </a-form>
  </a-card>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
const props = defineProps<{ modelValue?: any }>()
const emit = defineEmits<{'update:modelValue':[any],'save':[any]}>()

const text = ref('')
watch(() => props.modelValue, (v) => { text.value = v?.text || '' }, { immediate: true })

const wordCount = computed(() => text.value.length)
const save = () => emit('save', { text: text.value })
</script>

