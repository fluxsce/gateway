<template>
  <section class="json-result" :class="{ 'json-result--stale': stale }">
    <header class="json-result__bar">
      <span class="json-result__label">{{ t('work.structure') }}</span>
      <input
        v-model="query"
        class="json-result__search"
        type="search"
        :placeholder="t('work.searchStructure')"
        :disabled="!node"
        :aria-label="t('work.searchStructure')"
      />
      <span v-if="stale" class="json-result__stale" :title="t('work.structureStale')" />
      <RsButton
        variant="text"
        size="sm"
        icon="unfold-vertical"
        icon-only
        :disabled="!hasBranches"
        :tooltip="t('work.expandAll')"
        :aria-label="t('work.expandAll')"
        @click="expandAll"
      />
      <RsButton
        variant="text"
        size="sm"
        icon="fold-vertical"
        icon-only
        :disabled="!hasBranches"
        :tooltip="t('work.collapseAll')"
        :aria-label="t('work.collapseAll')"
        @click="expanded = new Set()"
      />
      <RsButton
        variant="text"
        size="sm"
        icon="copy"
        icon-only
        :disabled="!text"
        :tooltip="t('work.copy')"
        :aria-label="t('work.copy')"
        @click="copy"
      />
    </header>

    <div class="json-result__body">
      <p v-if="summary" class="json-result__meta">{{ summary }}</p>
      <p v-if="!node" class="json-result__empty">{{ t('work.resultEmpty') }}</p>
      <p v-else-if="noMatch" class="json-result__empty">{{ t('work.searchEmpty') }}</p>
      <p v-else-if="visible && visible.children.length === 0 && (visible.kind === 'object' || visible.kind === 'array')" class="json-result__literal">
        {{ visible.kind === 'array' ? '[]' : '{}' }}
      </p>
      <div v-else-if="visible" class="json-result__tree">
        <template v-if="visible.kind === 'object' || visible.kind === 'array'">
          <JsonBranch
            v-for="child in visible.children"
            :key="child.id"
            :node="child"
            :expanded="expanded"
            @toggle="toggle"
            @copy-path="copyPath"
          />
        </template>
        <JsonBranch v-else :node="visible" :expanded="expanded" @toggle="toggle" @copy-path="copyPath" />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { RsButton } from '@/ui'
import { copyToClipboardAsync } from '@/utils/clipboard'
import { computed, ref, watch } from 'vue'
import { allBranchIds, filterJsonTree, initialOpenIds, type JsonTreeNode } from '../jsonTree'
import JsonBranch from './JsonBranch.vue'

const props = defineProps<{
  node: JsonTreeNode | null
  text: string
  stale?: boolean
}>()

const { t } = useModuleI18n('hub0010')
const query = ref('')
const expanded = ref(new Set<string>())

const visible = computed(() => (props.node ? filterJsonTree(props.node, query.value) : null))
const noMatch = computed(() => Boolean(props.node && query.value.trim() && !visible.value))
const hasBranches = computed(() => Boolean(visible.value && allBranchIds(visible.value).length > 0))

const summary = computed(() => {
  const node = props.node
  if (!node) return ''
  if (node.kind === 'object') return `${t('work.jsonObject')} · ${t('work.jsonKeys', { count: node.children.length })}`
  if (node.kind === 'array') return `${t('work.jsonArray')} · ${t('work.jsonItems', { count: node.children.length })}`
  return node.preview
})

watch(
  () => props.node,
  (node, previous) => {
    if (!node) {
      expanded.value = new Set()
      return
    }
    if (query.value.trim()) {
      const matched = filterJsonTree(node, query.value)
      expanded.value = new Set(matched ? allBranchIds(matched) : [])
      return
    }
    if (!previous) {
      expanded.value = new Set(initialOpenIds(node))
      return
    }
    const alive = new Set(allBranchIds(node))
    const kept = [...expanded.value].filter((id) => alive.has(id))
    expanded.value = new Set(kept.length > 0 ? kept : initialOpenIds(node))
  },
  { immediate: true },
)

watch(query, () => {
  if (!props.node) return
  if (!query.value.trim()) {
    expanded.value = new Set(initialOpenIds(props.node))
    return
  }
  const matched = filterJsonTree(props.node, query.value)
  expanded.value = new Set(matched ? allBranchIds(matched) : [])
})

const toggle = (id: string) => {
  const next = new Set(expanded.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  expanded.value = next
}

const expandAll = () => {
  if (!visible.value) return
  expanded.value = new Set(allBranchIds(visible.value))
}

const copy = () => {
  if (!props.text) return
  void copyToClipboardAsync(props.text, {
    successMessage: t('work.copied'),
    errorMessage: t('work.copyFail'),
  })
}

const copyPath = (path: string) => {
  void copyToClipboardAsync(path, {
    successMessage: t('work.pathCopied'),
    errorMessage: t('work.copyFail'),
  })
}
</script>

<style lang="scss" scoped>
.json-result {
  container-type: inline-size;
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

.json-result--stale {
  border-color: color-mix(in srgb, var(--rs-warning) 55%, var(--rs-border));
}

.json-result__bar {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: none;
  height: 40px;
  padding: 0 8px 0 14px;
  border-bottom: 1px solid var(--rs-border);
  background: color-mix(in srgb, var(--rs-muted) 6%, var(--rs-surface));
}

.json-result__label {
  flex: none;
  font-size: 12px;
  font-weight: 650;
  letter-spacing: 0.02em;
  text-transform: uppercase;
  color: var(--rs-text);
}

@container (max-width: 340px) {
  .json-result__label {
    display: none;
  }
}

.json-result__search {
  flex: 1 1 72px;
  min-width: 0;
  height: 26px;
  margin: 0;
  padding: 0 8px;
  border: 1px solid var(--rs-border);
  border-radius: 7px;
  background: var(--rs-surface);
  color: var(--rs-text);
  font-size: 12px;
  outline: none;
}

.json-result__search:focus {
  border-color: color-mix(in srgb, var(--rs-primary) 55%, var(--rs-border));
}

.json-result__search:disabled {
  opacity: 0.55;
}

.json-result__stale {
  flex: none;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--rs-warning);
}

.json-result__body {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 8px 8px 12px 10px;
}

.json-result__meta {
  margin: 0 4px 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--rs-muted);
  font-size: 12px;
}

.json-result__empty,
.json-result__literal {
  margin: 28px 16px;
  color: var(--rs-muted);
  font-size: 13px;
  line-height: 1.5;
}

.json-result__literal {
  margin: 4px 14px 10px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}
</style>
