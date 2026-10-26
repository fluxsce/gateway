<template>
  <section class="tool-editor" :class="{ 'tool-editor--readonly': readonly, 'tool-editor--fault': !!errorAt }">
    <header class="tool-editor__bar">
      <span class="tool-editor__label">{{ label }}</span>
      <span v-if="hint" class="tool-editor__hint">{{ hint }}</span>
      <span v-if="errorAt" class="tool-editor__error">{{ errorAt }}</span>
      <span class="tool-editor__spacer" />
      <RsButton
        v-if="!readonly && model"
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
    <div class="tool-editor__body">
      <div
        v-if="lines"
        ref="gutterRef"
        class="tool-editor__gutter"
        :style="{ width: gutterWidth }"
        aria-hidden="true"
        @wheel.prevent="onGutterWheel"
      >
        <span
          v-for="n in lineCount"
          :key="n"
          class="tool-editor__ln"
          :class="{ 'is-error': errorLine === n }"
        >{{ n }}</span>
      </div>
      <textarea
        ref="areaRef"
        v-model="model"
        class="tool-editor__input"
        :readonly="readonly"
        :placeholder="placeholder"
        :wrap="lines ? 'off' : 'soft'"
        :style="tabSize ? { tabSize: String(tabSize) } : undefined"
        spellcheck="false"
        @scroll="syncGutter"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { RsButton } from '@/ui'
import { copyToClipboardAsync } from '@/utils/clipboard'
import { computed, ref } from 'vue'

const props = defineProps<{
  label: string
  hint?: string
  placeholder?: string
  readonly?: boolean
  errorAt?: string
  errorLine?: number
  lines?: boolean
  tabSize?: number
}>()

const model = defineModel<string>({ default: '' })
const { t } = useModuleI18n('hub0010')
const areaRef = ref<HTMLTextAreaElement | null>(null)
const gutterRef = ref<HTMLElement | null>(null)

const lineCount = computed(() => (model.value ? model.value.split('\n').length : 1))
const gutterWidth = computed(() => `calc(${Math.max(2, String(lineCount.value).length)}ch + 18px)`)

const syncGutter = () => {
  const gutter = gutterRef.value
  const area = areaRef.value
  if (!props.lines || !gutter || !area) return
  gutter.scrollTop = area.scrollTop
}

const onGutterWheel = (event: WheelEvent) => {
  const area = areaRef.value
  if (!area) return
  area.scrollTop += event.deltaY
  syncGutter()
}

/** 选中源文本中的一段，并把所在行滚进可视区。index 与 length 按 UTF-16 码元计。 */
const reveal = (index: number, length: number) => {
  const el = areaRef.value
  if (!el) return
  const start = Math.max(0, Math.min(index, model.value.length))
  const end = Math.max(start, Math.min(start + Math.max(length, 0), model.value.length))
  el.focus()
  el.setSelectionRange(start, end)
  const line = model.value.slice(0, start).split('\n').length
  const style = getComputedStyle(el)
  const fontSize = Number.parseFloat(style.fontSize) || 13
  const lineHeight = Number.parseFloat(style.lineHeight) || fontSize * 1.6
  el.scrollTop = Math.max(0, (line - 3) * lineHeight)
  syncGutter()
}

defineExpose({ reveal })

const copy = () => {
  if (!model.value) return
  void copyToClipboardAsync(model.value, {
    successMessage: t('work.copied'),
    errorMessage: t('work.copyFail'),
  })
}
</script>

<style lang="scss" scoped>
.tool-editor {
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
  --tool-line: 20.8px;
  --tool-pad: 14px;
}

.tool-editor__bar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: none;
  height: 40px;
  padding: 0 8px 0 14px;
  border-bottom: 1px solid var(--rs-border);
  background: color-mix(in srgb, var(--rs-muted) 6%, var(--rs-surface));
}

.tool-editor__label {
  font-size: 12px;
  font-weight: 650;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--rs-text);
}

.tool-editor__hint {
  font-size: 12px;
  color: var(--rs-muted);
}

.tool-editor--fault {
  border-color: color-mix(in srgb, var(--rs-danger) 55%, var(--rs-border));
}

.tool-editor__error {
  flex: none;
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
  color: var(--rs-danger);
}

.tool-editor__spacer {
  flex: 1;
}

.tool-editor__body {
  display: flex;
  flex: 1;
  min-height: 0;
  min-width: 0;
}

.tool-editor__gutter {
  flex: none;
  overflow: hidden;
  box-sizing: border-box;
  padding: var(--tool-pad) 8px var(--tool-pad) 0;
  border-right: 1px solid var(--rs-border);
  background: color-mix(in srgb, var(--rs-muted) 5%, var(--rs-surface));
  color: var(--rs-muted);
  text-align: right;
  user-select: none;
}

.tool-editor__ln {
  display: block;
  height: var(--tool-line);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 13px;
  font-variant-numeric: tabular-nums;
  line-height: var(--tool-line);
}

.tool-editor__ln.is-error {
  color: var(--rs-danger);
  font-weight: 650;
}

.tool-editor__input {
  flex: 1;
  min-width: 0;
  min-height: 0;
  width: 100%;
  height: 100%;
  margin: 0;
  padding: var(--tool-pad) 16px;
  border: 0;
  resize: none;
  outline: none;
  box-sizing: border-box;
  background: transparent;
  color: var(--rs-text);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 13px;
  line-height: var(--tool-line);
}

.tool-editor--readonly .tool-editor__input {
  background: color-mix(in srgb, var(--rs-muted) 4%, var(--rs-surface));
}

.tool-editor__input::placeholder {
  color: var(--rs-muted);
}
</style>
