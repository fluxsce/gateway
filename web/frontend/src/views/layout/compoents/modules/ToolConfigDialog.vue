<template>
  <RsDialog
    :open="show"
    :title="t('config.title')"
    layout="confirm"
    width="md"
    @update:open="(val: boolean) => emit('update:show', val)"
  >
    <template #body>
      <div v-if="tool" class="config-dialog">
        <div class="config-field">
          <label class="config-label">{{ t('config.name') }}</label>
          <RsInput :model-value="tool.displayName || tool.name" readonly />
        </div>

        <div v-for="field in tool.configSchema || []" :key="field.key" class="config-field">
          <label class="config-label">{{ field.label }}</label>
          <RsSelect
            v-if="field.type === 'select'"
            v-model="form[field.key]"
            :options="field.options || []"
            block
          />
        </div>
      </div>
    </template>

    <template #footer>
      <div class="config-footer">
        <RsButton variant="ghost" @click="emit('update:show', false)">{{ t('config.cancel') }}</RsButton>
        <RsButton variant="primary" @click="handleSave">{{ t('config.save') }}</RsButton>
      </div>
    </template>
  </RsDialog>
</template>

<script setup lang="ts">
import { RsButton, RsDialog, RsInput, RsSelect } from '@/ui'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { reactive, watch } from 'vue'
import type { Tool } from '../../types/toolMarketplace'

interface Props {
  show: boolean
  tool?: Tool | null
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  save: [config: Record<string, string | number>]
}>()

const { t } = useModuleI18n('hub0010')
const form = reactive<Record<string, string | number>>({})

watch(
  () => props.tool,
  (tool) => {
    for (const key of Object.keys(form)) delete form[key]
    const source = tool?.userConfig || tool?.defaultConfig || {}
    for (const field of tool?.configSchema || []) {
      const value = source[field.key] ?? field.defaultValue ?? ''
      form[field.key] = typeof value === 'number' || typeof value === 'string' ? value : String(value)
    }
  },
  { immediate: true },
)

const handleSave = () => {
  emit('save', { ...form })
}
</script>

<style lang="scss" scoped>
.config-dialog {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.config-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.config-label {
  font-size: var(--rs-font-size-sm, 13px);
  color: var(--text-color-secondary);
}

.config-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
