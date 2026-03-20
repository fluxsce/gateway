<template>
  <GModal
    :visible="show"
    preset="dialog"
    :title="title"
    width="720px"
    :mask="true"
    :block-scroll="true"
    :draggable="true"
    :show-fullscreen-toggle="true"
    :show-cancel="true"
    :show-confirm="true"
    cancel-text="取消"
    confirm-text="确定"
    @update:visible="handleUpdateVisible"
    @confirm="handleConfirm"
    @cancel="handleCancel"
    @close="handleCancel"
  >
    <GTextShow
      :content="content"
      format="auto"
      :show-copy-button="true"
      :auto-format="true"
      :show-line-numbers="showLineNumbers"
      max-height="52vh"
    />
  </GModal>
</template>

<script setup lang="ts">
import { GModal } from '@/components/gmodal'
import { GTextShow } from '@/components'

defineOptions({ name: 'CustomRenderGModalTextDialog' })

const props = withDefaults(
  defineProps<{
    show: boolean
    title?: string
    content?: string
    showLineNumbers?: boolean
  }>(),
  {
    title: 'GModal + GTextShow',
    content: '',
    showLineNumbers: false,
  }
)

const emit = defineEmits<{
  (e: 'update:show', v: boolean): void
  (e: 'success', data?: unknown): void
}>()

function handleUpdateVisible(v: boolean) {
  emit('update:show', v)
}

function handleCancel() {
  emit('update:show', false)
}

function handleConfirm() {
  emit('success', { confirmed: true, title: props.title })
  emit('update:show', false)
}

const showLineNumbers = props.showLineNumbers
</script>

