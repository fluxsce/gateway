<template>
  <ToolPage
    icon="braces"
    :title="t('tools.jsonFormatter.name')"
    :description="t('tools.jsonFormatter.description')"
    :loading="loading"
    :denied="denied"
    :failed="failed"
  >
    <template #actions>
      <span class="json-formatter__indent">{{ t('config.indent') }} {{ indent }}</span>
      <RsButton variant="primary" icon="wrap-text" @click="format">{{ t('work.format') }}</RsButton>
      <RsButton variant="ghost" icon="minimize-2" @click="minify">{{ t('work.minify') }}</RsButton>
    </template>

    <div class="json-formatter">
      <div v-if="fault" class="json-formatter__fault" role="alert">
        <p class="json-formatter__where">{{ faultWhere }}</p>
        <p class="json-formatter__why">{{ faultWhy }}</p>
        <pre class="json-formatter__snippet">{{ snippet }}</pre>
      </div>
      <RsSplitPane class="json-formatter__split" fill with-handle :panes="panes">
        <template #input>
          <JsonEditor
            ref="inputEditor"
            v-model="input"
            :label="t('work.input')"
            :error-at="faultMark"
            :diagnostics="diagnostics"
          />
        </template>
        <template #output>
          <JsonResult :node="tree" :text="output" :stale="stale" />
        </template>
      </RsSplitPane>
    </div>
  </ToolPage>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { RsButton, RsSplitPane, type RsSplitPaneItem } from '@/ui'
import type { RsCodeEditorDiagnostic } from '@/ui/code-editor'
import { computed, nextTick, onUnmounted, ref, watch } from 'vue'
import { formatJsonText, jsonFaultSnippet, locateJsonFault, type JsonFault, type JsonFaultCode } from '../jsonFormat'
import { buildJsonTree, type JsonTreeNode } from '../jsonTree'
import { useGrantedTool } from '../useGrantedTool'
import JsonEditor from './JsonEditor.vue'
import JsonResult from './JsonResult.vue'
import ToolPage from './ToolPage.vue'

defineOptions({ name: 'JsonFormatter' })

const panes: RsSplitPaneItem[] = [
  { key: 'input', min: 28 },
  { key: 'output', min: 28 },
]

const { t } = useModuleI18n('hub0010')
const { loading, denied, failed, item } = useGrantedTool('json-formatter')
const input = ref('')
const output = ref('')
const tree = ref<JsonTreeNode | null>(null)
const stale = ref(false)
const fault = ref<JsonFault | null>(null)
const snippet = ref('')
const inputEditor = ref<{ goToPosition: (line: number, column?: number) => void } | null>(null)
let skipSync = 0
let syncTimer = 0

const reasonKey: Record<JsonFaultCode, string> = {
  empty: 'work.jsonEmpty',
  unexpected: 'work.jsonUnexpected',
  trailing: 'work.jsonTrailing',
  'trailing-comma': 'work.jsonTrailingComma',
  'unterminated-string': 'work.jsonUnterminated',
  'bad-escape': 'work.jsonBadEscape',
  'bad-unicode': 'work.jsonBadUnicode',
  'control-char': 'work.jsonControl',
  'bad-number': 'work.jsonBadNumber',
  'expected-colon': 'work.jsonExpectedColon',
  'expected-comma-or-end': 'work.jsonExpectedComma',
  'expected-value': 'work.jsonExpectedValue',
  eof: 'work.jsonEof',
}

const indent = computed(() => {
  const raw = item.value?.userConfig?.indent ?? item.value?.defaultConfig?.indent ?? 2
  return Number(raw) === 4 ? 4 : 2
})

const tokenLabel = (token: string) => {
  const char = token[0] ?? ''
  if (!char) return t('work.jsonTokenEof')
  if (char === '\n') return t('work.jsonTokenNl')
  if (char === '\t') return t('work.jsonTokenTab')
  if (char === '\r') return 'CR'
  const code = char.charCodeAt(0)
  if (code < 32) return `U+${code.toString(16).toUpperCase().padStart(4, '0')}`
  return char
}

const faultWhere = computed(() => {
  if (!fault.value) return ''
  return t('work.jsonAt', { line: fault.value.line, column: fault.value.column })
})

const faultMark = computed(() => (fault.value ? `${fault.value.line}:${fault.value.column}` : ''))

const faultWhy = computed(() => {
  if (!fault.value) return ''
  return t(reasonKey[fault.value.code], { token: tokenLabel(fault.value.token) })
})

const diagnostics = computed<RsCodeEditorDiagnostic[]>(() => {
  if (!fault.value) return []
  return [
    {
      severity: 'error',
      line: fault.value.line,
      column: fault.value.column,
      message: faultWhy.value,
    },
  ]
})

const applyTree = (source: string) => {
  if (!source.trim()) {
    tree.value = null
    output.value = ''
    stale.value = false
    return
  }
  if (locateJsonFault(source)) {
    stale.value = tree.value !== null
    return
  }
  tree.value = buildJsonTree(JSON.parse(source) as unknown)
  output.value = source
  stale.value = false
  fault.value = null
  snippet.value = ''
}

watch(input, () => {
  if (skipSync > 0) {
    skipSync -= 1
    return
  }
  fault.value = null
  snippet.value = ''
  window.clearTimeout(syncTimer)
  syncTimer = window.setTimeout(() => applyTree(input.value), 160)
})

onUnmounted(() => {
  window.clearTimeout(syncTimer)
})

const run = (space: number | undefined) => {
  window.clearTimeout(syncTimer)
  const result = formatJsonText(input.value, space)
  if (result.ok) {
    fault.value = null
    snippet.value = ''
    output.value = result.text
    tree.value = buildJsonTree(JSON.parse(result.text) as unknown)
    stale.value = false
    if (input.value !== result.text) {
      skipSync += 1
      input.value = result.text
    }
    return
  }
  fault.value = result.fault
  snippet.value = jsonFaultSnippet(input.value, result.fault)
  const located = result.fault
  void nextTick(() => inputEditor.value?.goToPosition(located.line, located.column))
}

const format = () => run(indent.value)
const minify = () => run(undefined)
</script>

<style lang="scss" scoped>
.json-formatter {
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: 100%;
  min-height: 0;
}

.json-formatter__fault {
  flex: none;
  padding: 10px 14px;
  border-radius: 10px;
  border: 1px solid color-mix(in srgb, var(--rs-danger) 35%, var(--rs-border));
  background: color-mix(in srgb, var(--rs-danger) 8%, var(--rs-surface));
}

.json-formatter__where,
.json-formatter__why {
  margin: 0;
  font-size: 13px;
  line-height: 1.45;
}

.json-formatter__where {
  font-weight: 650;
  font-variant-numeric: tabular-nums;
  color: var(--rs-danger);
}

.json-formatter__why {
  margin-top: 2px;
  color: var(--rs-text);
}

.json-formatter__snippet {
  margin: 8px 0 0;
  padding: 8px 10px;
  overflow: auto;
  border-radius: 8px;
  background: var(--rs-surface);
  color: var(--rs-text);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre;
}

.json-formatter__split {
  width: 100%;
  flex: 1;
  min-height: 0;
}

.json-formatter__indent {
  margin-right: 4px;
  padding: 4px 10px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--rs-primary) 12%, var(--rs-surface));
  color: var(--rs-primary);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
}
</style>
