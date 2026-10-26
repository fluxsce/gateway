<template>
  <section class="json-editor" :class="{ 'json-editor--fault': diagnostics.length > 0 }">
    <header class="json-editor__bar">
      <span class="json-editor__label">{{ label }}</span>
      <span v-if="errorAt" class="json-editor__error">{{ errorAt }}</span>
      <span class="json-editor__spacer" />
      <RsButton
        variant="text"
        size="sm"
        icon="unfold-vertical"
        icon-only
        :disabled="!model"
        :tooltip="t('work.expandAll')"
        :aria-label="t('work.expandAll')"
        @click="unfoldAll"
      />
      <RsButton
        variant="text"
        size="sm"
        icon="fold-vertical"
        icon-only
        :disabled="!model"
        :tooltip="t('work.collapseAll')"
        :aria-label="t('work.collapseAll')"
        @click="foldAll"
      />
      <RsButton
        v-if="model"
        variant="text"
        size="sm"
        icon="eraser"
        icon-only
        :tooltip="t('work.clear')"
        :aria-label="t('work.clear')"
        @click="model = ''"
      />
      <RsButton
        variant="text"
        size="sm"
        icon="copy"
        icon-only
        :disabled="!model"
        :tooltip="t('work.copy')"
        :aria-label="t('work.copy')"
        @click="copy"
      />
    </header>
    <div class="json-editor__body">
      <RsCodeEditor
        ref="editorRef"
        v-model="model"
        language="json"
        height="100%"
        embedded
        :show-toolbar="false"
        :wrap="false"
        :fold-gutter="true"
        :diagnostics="diagnostics"
        :aria-label="label"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { RsButton } from '@/ui'
import { RsCodeEditor, type RsCodeEditorDiagnostic } from '@/ui/code-editor'
import { copyToClipboardAsync } from '@/utils/clipboard'
import { ref } from 'vue'

withDefaults(
  defineProps<{
    label: string
    errorAt?: string
    diagnostics?: RsCodeEditorDiagnostic[]
  }>(),
  { diagnostics: () => [] },
)

const model = defineModel<string>({ default: '' })
const { t } = useModuleI18n('hub0010')
const editorRef = ref<{
  goToPosition: (line: number, column?: number) => void
  foldAll: () => void
  unfoldAll: () => void
} | null>(null)

const goToPosition = (line: number, column?: number) => {
  editorRef.value?.goToPosition(line, column)
}

const foldAll = () => {
  editorRef.value?.foldAll()
}

const unfoldAll = () => {
  editorRef.value?.unfoldAll()
}

defineExpose({ goToPosition, foldAll, unfoldAll })

const copy = () => {
  if (!model.value) return
  void copyToClipboardAsync(model.value, {
    successMessage: t('work.copied'),
    errorMessage: t('work.copyFail'),
  })
}
</script>

<style lang="scss" scoped>
.json-editor {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  min-width: 0;
  overflow: hidden;
  border: 1px solid var(--rs-border);
  border-radius: 12px;
  background: var(--rs-surface);
  box-shadow: var(--rs-shadow-sm);
}

.json-editor--fault {
  border-color: color-mix(in srgb, var(--rs-danger) 55%, var(--rs-border));
}

.json-editor__bar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: none;
  height: 40px;
  padding: 0 8px 0 14px;
  border-bottom: 1px solid var(--rs-border);
  background: color-mix(in srgb, var(--rs-muted) 6%, var(--rs-surface));
}

.json-editor__label {
  font-size: 12px;
  font-weight: 650;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--rs-text);
}

.json-editor__error {
  flex: none;
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
  color: var(--rs-danger);
}

.json-editor__spacer {
  flex: 1;
}

.json-editor__body {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
  min-width: 0;
}

.json-editor__body :deep(.rs-code-editor__diagnostics) {
  display: none;
}

.json-editor__body :deep(.cm-foldGutter .cm-gutterElement) {
  cursor: pointer;
}
</style>
