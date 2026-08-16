<template>
  <div class="rewrite-rules-editor">
    <div v-for="(rule, index) in rules" :key="index" class="rewrite-rules-editor__row">
      <select v-model="rule.mode" class="rewrite-rules-editor__mode" @change="emitRules">
        <option value="prefix">前缀</option>
        <option value="exact">精确</option>
        <option value="regex">正则</option>
      </select>
      <input
        v-model="rule.from"
        class="rewrite-rules-editor__input"
        placeholder="匹配 /old"
        @change="emitRules"
      />
      <input
        v-model="rule.to"
        class="rewrite-rules-editor__input"
        placeholder="替换 /new"
        @change="emitRules"
      />
      <button type="button" class="rewrite-rules-editor__btn" @click="removeRule(index)">删除</button>
    </div>
    <button type="button" class="rewrite-rules-editor__btn rewrite-rules-editor__btn--add" @click="addRule">
      添加规则
    </button>
  </div>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue'

defineOptions({
  name: 'RewriteRulesEditor',
})

interface RuleRow {
  mode: string
  from: string
  to: string
}

const props = defineProps<{
  modelValue?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const rules = ref<RuleRow[]>([])

function parseRules(raw?: string): RuleRow[] {
  const text = (raw || '').trim()
  if (!text) {
    return []
  }
  try {
    const parsed = JSON.parse(text) as Array<{ mode?: string; from?: string; to?: string }>
    if (Array.isArray(parsed)) {
      return parsed
        .filter((item) => item && typeof item.from === 'string' && item.from.trim())
        .map((item) => ({
          mode: item.mode || 'prefix',
          from: item.from || '',
          to: item.to || '',
        }))
    }
  } catch {
    // 兼容逐行文本
  }
  return text
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line && !line.startsWith('#'))
    .map((line) => {
      const fields = line.replace('=>', ' ').split(/\s+/).filter(Boolean)
      if (fields[0] === 'prefix' || fields[0] === 'exact' || fields[0] === 'regex') {
        return { mode: fields[0], from: fields[1] || '', to: fields.slice(2).join(' ') }
      }
      return { mode: 'prefix', from: fields[0] || '', to: fields.slice(1).join(' ') }
    })
}

function emitRules() {
  const cleaned = rules.value.filter((item) => item.from.trim())
  emit('update:modelValue', cleaned.length ? JSON.stringify(cleaned) : '')
}

function addRule() {
  rules.value.push({ mode: 'prefix', from: '', to: '' })
}

function removeRule(index: number) {
  rules.value.splice(index, 1)
  emitRules()
}

watch(
  () => props.modelValue,
  (value) => {
    const next = parseRules(value)
    const current = JSON.stringify(rules.value.filter((item) => item.from.trim()))
    const incoming = JSON.stringify(next)
    if (current !== incoming) {
      rules.value = next
    }
  },
  { immediate: true },
)
</script>

<style scoped>
.rewrite-rules-editor {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}
.rewrite-rules-editor__row {
  display: grid;
  grid-template-columns: 88px 1fr 1fr auto;
  gap: 8px;
  align-items: center;
}
.rewrite-rules-editor__mode,
.rewrite-rules-editor__input,
.rewrite-rules-editor__btn {
  height: 32px;
  border: 1px solid var(--rs-border-color, #d9d9d9);
  border-radius: 4px;
  padding: 0 8px;
  background: transparent;
  color: inherit;
}
.rewrite-rules-editor__btn--add {
  width: fit-content;
}
</style>
