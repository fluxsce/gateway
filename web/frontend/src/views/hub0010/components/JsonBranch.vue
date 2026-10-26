<template>
  <div class="json-branch">
    <div
      class="json-branch__row"
      :class="{ 'json-branch__row--openable': openable }"
      @click="onRowClick"
    >
      <button
        v-if="openable"
        type="button"
        class="json-branch__twist"
        :aria-expanded="open"
        :aria-label="node.label || node.preview"
        @click.stop="emit('toggle', node.id)"
      >
        <RsIcon :name="open ? 'chevron-down' : 'chevron-right'" :size="14" />
      </button>
      <span v-else class="json-branch__twist" />
      <span v-if="node.label" class="json-branch__key" :class="{ 'json-branch__key--index': indexKey }">{{ node.label }}</span>
      <span v-if="node.label" class="json-branch__colon">:</span>
      <span class="json-branch__value" :class="`json-branch__value--${node.kind}`" :title="node.title || undefined">{{ node.preview }}</span>
      <button
        v-if="node.id !== '$'"
        type="button"
        class="json-branch__path"
        :title="node.id"
        :aria-label="copyPathLabel"
        @click.stop="emit('copyPath', node.id)"
      >
        <RsIcon name="link" :size="12" />
      </button>
    </div>
    <div v-if="open" class="json-branch__children">
      <JsonBranch
        v-for="child in node.children"
        :key="child.id"
        :node="child"
        :expanded="expanded"
        @toggle="emit('toggle', $event)"
        @copy-path="emit('copyPath', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { RsIcon } from '@/ui'
import { computed } from 'vue'
import type { JsonTreeNode } from '../jsonTree'

defineOptions({ name: 'JsonBranch' })

const props = defineProps<{
  node: JsonTreeNode
  expanded: Set<string>
}>()

const emit = defineEmits<{
  toggle: [id: string]
  copyPath: [path: string]
}>()

const { t } = useModuleI18n('hub0010')
const copyPathLabel = computed(() => t('work.copyPath'))

const openable = computed(() => props.node.children.length > 0)
const open = computed(() => openable.value && props.expanded.has(props.node.id))
const indexKey = computed(() => /^\d+$/.test(props.node.label))

const onRowClick = () => {
  if (openable.value) emit('toggle', props.node.id)
}
</script>

<style lang="scss" scoped>
.json-branch__row {
  display: flex;
  align-items: baseline;
  gap: 4px;
  min-height: 24px;
  padding-right: 12px;
  border-radius: 6px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 13px;
  line-height: 24px;
}

.json-branch__row--openable {
  cursor: pointer;
}

.json-branch__row--openable:hover {
  background: color-mix(in srgb, var(--rs-primary) 8%, transparent);
}

.json-branch__twist {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 24px;
  flex: none;
  padding: 0;
  border: 0;
  background: transparent;
  color: var(--rs-muted);
}

.json-branch__key {
  flex: none;
  color: var(--rs-primary);
}

.json-branch__key--index {
  color: var(--rs-muted);
  font-variant-numeric: tabular-nums;
}

.json-branch__colon {
  flex: none;
  color: var(--rs-muted);
}

.json-branch__value {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.json-branch__path {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  margin-left: auto;
  flex: none;
  padding: 0;
  border: 0;
  border-radius: 4px;
  background: transparent;
  color: var(--rs-muted);
  opacity: 0;
  cursor: pointer;
}

.json-branch__row:hover .json-branch__path,
.json-branch__path:focus-visible {
  opacity: 1;
}

.json-branch__path:hover {
  color: var(--rs-primary);
  background: color-mix(in srgb, var(--rs-primary) 10%, transparent);
}

.json-branch__value--string {
  color: var(--rs-success);
}

.json-branch__value--number {
  color: var(--rs-info);
}

.json-branch__value--boolean,
.json-branch__value--null {
  color: var(--rs-warning);
}

.json-branch__value--object,
.json-branch__value--array {
  color: var(--rs-muted);
}

.json-branch__children {
  margin-left: 7px;
  padding-left: 12px;
  border-left: 1px solid color-mix(in srgb, var(--rs-border) 80%, var(--rs-muted));
}
</style>
