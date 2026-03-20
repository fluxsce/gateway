<template>
  <NModal
    :show="show"
    preset="card"
    :title="title"
    style="width: 400px"
    @update:show="$emit('update:show', $event)"
  >
    <GTextShow
      :content="content"
      format="auto"
      :show-copy-button="true"
      :auto-format="true"
      max-height="220px"
    />
    <template #footer>
      <NSpace justify="end">
        <NButton @click="handleCancel">取消</NButton>
        <NButton type="primary" @click="handleConfirm">确定</NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<script setup lang="ts">
import { NModal, NButton, NSpace } from 'naive-ui'
import { GTextShow } from '@/components'

defineOptions({ name: 'CustomRenderDemoDialog' })

const props = withDefaults(
  defineProps<{
    show: boolean
    title?: string
    content?: string
  }>(),
  {
    title: '自定义渲染弹窗',
    content: '由 $gRender.show() 打开，子组件通过 emit 关闭。',
  }
)

const emit = defineEmits<{
  (e: 'update:show', v: boolean): void
  (e: 'success', data?: unknown): void
}>()

function handleCancel() {
  emit('update:show', false)
}

function handleConfirm() {
  emit('success', { confirmed: true, title: props.title })
  emit('update:show', false)
}
</script>
