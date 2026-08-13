<template>
  <RsDialog
    :open="show"
    title="工具配置"
    layout="confirm"
    width="md"
    @update:open="(val: boolean) => emit('update:show', val)"
  >
    <template #body>
      <div v-if="tool" class="config-dialog">
        <div class="config-field">
          <label class="config-label">工具名称</label>
          <RsInput :model-value="tool.displayName || tool.name" readonly />
        </div>

        <div class="config-field">
          <label class="config-label">版本</label>
          <RsInput :model-value="tool.version" readonly />
        </div>

        <div class="config-field">
          <label class="config-label">状态</label>
          <RsTag :variant="tool.status === ToolStatus.INSTALLED ? 'success' : 'default'" size="sm">
            {{ tool.status }}
          </RsTag>
        </div>
      </div>
    </template>

    <template #footer>
      <div class="config-footer">
        <RsButton variant="ghost" @click="emit('update:show', false)">取消</RsButton>
        <RsButton variant="primary" @click="handleSave">保存</RsButton>
      </div>
    </template>
  </RsDialog>
</template>

<script setup lang="ts">
import { RsButton, RsDialog, RsInput, RsTag } from '@/ui'
import { ToolStatus, type Tool } from '../../types/toolMarketplace'

interface Props {
  show: boolean
  tool?: Tool | null
}

defineProps<Props>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  save: [config: Record<string, any>]
}>()

const handleSave = () => {
  emit('save', {})
  emit('update:show', false)
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
