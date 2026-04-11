<template>
  <div
    class="g-key-value-editor"
    :class="{ 'g-key-value-editor--table': variant === 'table' }"
  >
    <!-- Query / Header：勾选 + 键 + = + 值 -->
    <template v-if="variant === 'table' && tableVariant === 'query'">
      <div class="g-key-value-editor__table">
        <div class="g-key-value-editor__thead g-key-value-editor__thead--query">
          <span class="g-key-value-editor__th g-key-value-editor__th--check" aria-hidden="true" />
          <span class="g-key-value-editor__th">{{ keyColumnLabel }}</span>
          <span class="g-key-value-editor__th g-key-value-editor__th--eq" aria-hidden="true" />
          <span class="g-key-value-editor__th">{{ valueColumnLabel }}</span>
          <span class="g-key-value-editor__th g-key-value-editor__th--action" aria-hidden="true" />
        </div>
        <div
          v-for="row in rows"
          :key="row.id"
          class="g-key-value-editor__tr g-key-value-editor__tr--query"
          :class="{ 'g-key-value-editor__tr--auto-body': row.autoFromBody }"
        >
          <div class="g-key-value-editor__td g-key-value-editor__td--check">
            <n-checkbox v-model:checked="row.enabled" />
          </div>
          <div class="g-key-value-editor__td">
            <n-input
              v-model:value="row.key"
              size="small"
              :placeholder="keyPlaceholder"
              class="g-key-value-editor__cell-input"
            />
          </div>
          <div class="g-key-value-editor__td g-key-value-editor__td--eq" aria-hidden="true">
            <span class="g-key-value-editor__equals">=</span>
          </div>
          <div class="g-key-value-editor__td g-key-value-editor__td--grow">
            <div class="g-key-value-editor__value-cell">
              <n-input
                v-model:value="row.value"
                size="small"
                :placeholder="valuePlaceholder"
                class="g-key-value-editor__cell-input"
              />
              <span
                v-if="row.autoFromBody && showAutoBodyHint"
                class="g-key-value-editor__auto-hint"
              >来自 Body 设置</span>
            </div>
          </div>
          <div class="g-key-value-editor__td g-key-value-editor__td--action">
            <n-button
              quaternary
              circle
              size="small"
              class="g-key-value-editor__remove"
              :aria-label="'删除该行'"
              @click="removeRow(row.id)"
            >
              <template #icon>
                <g-icon icon="RemoveOutline" size="small" />
              </template>
            </n-button>
          </div>
        </div>
      </div>
    </template>

    <!-- x-www-form-urlencoded / form-data：参数名、类型、参数值 -->
    <template v-else-if="variant === 'table' && tableVariant === 'form'">
      <div class="g-key-value-editor__table">
        <div class="g-key-value-editor__thead g-key-value-editor__thead--form">
          <span class="g-key-value-editor__th g-key-value-editor__th--check" aria-hidden="true" />
          <span class="g-key-value-editor__th">{{ keyColumnLabel }}</span>
          <span class="g-key-value-editor__th">{{ typeColumnLabel }}</span>
          <span class="g-key-value-editor__th">{{ valueColumnLabel }}</span>
          <span class="g-key-value-editor__th g-key-value-editor__th--action" aria-hidden="true" />
        </div>
        <div
          v-for="row in rows"
          :key="row.id"
          class="g-key-value-editor__tr g-key-value-editor__tr--form"
        >
          <div class="g-key-value-editor__td g-key-value-editor__td--check">
            <n-checkbox v-model:checked="row.enabled" />
          </div>
          <div class="g-key-value-editor__td">
            <n-input
              v-model:value="row.key"
              size="small"
              :placeholder="keyPlaceholder"
              class="g-key-value-editor__cell-input"
            />
          </div>
          <div class="g-key-value-editor__td g-key-value-editor__td--type">
            <n-select
              size="small"
              :value="row.fieldKind ?? 'text'"
              :options="fieldKindOptions"
              :disabled="formTableKind === 'urlencoded'"
              class="g-key-value-editor__type-select"
              @update:value="(v) => setFieldKind(row, v)"
            />
          </div>
          <div class="g-key-value-editor__td g-key-value-editor__td--grow g-key-value-editor__td--value">
            <template v-if="(row.fieldKind ?? 'text') === 'text' || formTableKind === 'urlencoded'">
              <n-input
                v-model:value="row.value"
                size="small"
                :placeholder="valuePlaceholder"
                class="g-key-value-editor__cell-input"
              />
            </template>
            <template v-else>
              <div class="g-key-value-editor__file-cell">
                <label class="g-key-value-editor__file-label">
                  <input
                    type="file"
                    class="g-key-value-editor__file-native"
                    @change="(e: Event) => onFilePick(row, e)"
                  />
                  <span class="g-key-value-editor__file-btn">选择文件</span>
                </label>
                <span
                  v-if="row.file"
                  class="g-key-value-editor__file-name"
                  :title="row.file.name"
                >{{ row.file.name }}</span>
              </div>
            </template>
          </div>
          <div class="g-key-value-editor__td g-key-value-editor__td--action">
            <n-button
              quaternary
              circle
              size="small"
              class="g-key-value-editor__remove"
              :aria-label="'删除该行'"
              @click="removeRow(row.id)"
            >
              <template #icon>
                <g-icon icon="RemoveOutline" size="small" />
              </template>
            </n-button>
          </div>
        </div>
      </div>
    </template>

    <!-- 紧凑行内 -->
    <template v-else>
      <div
        v-for="row in rows"
        :key="row.id"
        class="g-key-value-editor__row"
      >
        <n-checkbox v-model:checked="row.enabled" />
        <n-input
          v-model:value="row.key"
          size="small"
          :placeholder="keyPlaceholder"
          class="g-key-value-editor__key"
        />
        <n-input
          v-model:value="row.value"
          size="small"
          :placeholder="valuePlaceholder"
          class="g-key-value-editor__val"
        />
        <n-button
          size="tiny"
          quaternary
          @click="removeRow(row.id)"
        >
          删除
        </n-button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { GIcon } from '@/components/gicon'
import { NButton, NCheckbox, NInput, NSelect } from 'naive-ui'
import { computed, nextTick, onMounted, watch } from 'vue'
import type { RestFormFieldKind, RestKeyValueRow } from './types'
import { createKeyValueRow } from './types'

/**
 * 键值编辑器：支持 Query/Header 表格、表单三列表格与行内堆叠。
 * 表格模式在底部自动保留一行空行，无需单独「添加」按钮。
 */
defineOptions({
  name: 'KeyValueEditor'
})

const rows = defineModel<RestKeyValueRow[]>('rows', { required: true })

const props = withDefaults(
  defineProps<{
    variant?: 'inline' | 'table'
    /** table 时：query 为键=值；form 为参数名/类型/参数值 */
    tableVariant?: 'query' | 'form'
    /** form 表格时：urlencoded 仅文本；form-data 可选文件 */
    formTableKind?: 'urlencoded' | 'multipart'
    keyColumnLabel?: string
    valueColumnLabel?: string
    typeColumnLabel?: string
    keyPlaceholder?: string
    valuePlaceholder?: string
    /** 为 true 时在由 Body 推导的 Content-Type 行右侧显示「来自 Body 设置」 */
    showAutoBodyHint?: boolean
  }>(),
  {
    variant: 'inline',
    tableVariant: 'query',
    formTableKind: 'urlencoded',
    keyColumnLabel: '参数名',
    valueColumnLabel: '参数值',
    typeColumnLabel: '类型',
    keyPlaceholder: '',
    valuePlaceholder: '',
    showAutoBodyHint: false
  }
)

const fieldKindOptions = computed(() => {
  if (props.formTableKind === 'urlencoded') {
    return [{ label: 'Text', value: 'text' as RestFormFieldKind }]
  }
  return [
    { label: 'Text', value: 'text' as RestFormFieldKind },
    { label: 'File', value: 'file' as RestFormFieldKind }
  ]
})

/**
 * 判断一行是否视为「有内容」，用于决定是否须在末尾再追加空行。
 */
function rowHasContent(r: RestKeyValueRow): boolean {
  if (r.key.trim()) {
    return true
  }
  if ((r.fieldKind ?? 'text') === 'file' && r.file) {
    return true
  }
  if (r.value.trim()) {
    return true
  }
  return false
}

/**
 * 比较两行数组是否同一组 id（用于避免无意义的 v-model 回写触发循环）。
 */
function sameRowIds(a: RestKeyValueRow[], b: RestKeyValueRow[]): boolean {
  if (a.length !== b.length) {
    return false
  }
  for (let i = 0; i < a.length; i++) {
    if (a[i].id !== b[i].id) {
      return false
    }
  }
  return true
}

/**
 * 尾部仅保留一个空行；若最后一行有内容则自动在底部追加一空行。
 */
function ensureTrailingEmptyRow(): void {
  if (props.variant !== 'table') {
    return
  }
  let list = rows.value.length > 0 ? [...rows.value] : [createKeyValueRow()]

  while (list.length >= 2) {
    const prev = list[list.length - 2]
    const last = list[list.length - 1]
    if (!rowHasContent(prev) && !rowHasContent(last)) {
      list = list.slice(0, -1)
    } else {
      break
    }
  }

  const last = list[list.length - 1]
  if (rowHasContent(last)) {
    list = [...list, createKeyValueRow()]
  }

  if (!sameRowIds(rows.value, list)) {
    rows.value = list
  }
}

/**
 * 切换字段为文本或文件；切回文本时清除已选文件。
 */
function setFieldKind(row: RestKeyValueRow, kind: RestFormFieldKind): void {
  row.fieldKind = kind
  if (kind === 'text') {
    row.file = null
  }
}

/**
 * 本地选择文件后写入行数据。
 */
function onFilePick(row: RestKeyValueRow, e: Event): void {
  const input = e.target as HTMLInputElement
  const f = input.files?.[0]
  row.file = f ?? null
  row.value = f?.name ?? ''
  input.value = ''
}

/**
 * 删除指定 id 的行；至少保留一行并由 ensureTrailingEmptyRow 补齐尾部空行。
 */
function removeRow(id: string): void {
  const next = rows.value.filter((r) => r.id !== id)
  rows.value = next.length > 0 ? next : [createKeyValueRow()]
  ensureTrailingEmptyRow()
}

watch(
  rows,
  () => {
    nextTick(() => {
      ensureTrailingEmptyRow()
    })
  },
  { deep: true }
)

onMounted(() => {
  nextTick(() => {
    ensureTrailingEmptyRow()
  })
})
</script>

<style scoped lang="scss">
.g-key-value-editor {
  &__table {
    border: 1px solid var(--n-border-color);
    border-radius: var(--n-border-radius);
    overflow: hidden;
    background: var(--n-color);
  }

  &__thead {
    display: grid;
    align-items: center;
    gap: 0;
    padding: 8px 10px;
    font-size: 12px;
    color: var(--n-text-color-3);
    background: var(--n-color-modal);
    border-bottom: 1px solid var(--n-border-color);

    &--query {
      grid-template-columns: 36px 1fr 28px 1.4fr 40px;
    }

    &--form {
      grid-template-columns: 36px minmax(72px, 1fr) 100px minmax(100px, 1.4fr) 40px;
    }
  }

  &__th {
    &--check {
      justify-self: center;
    }
    &--eq {
      text-align: center;
    }
    &--action {
      width: 40px;
    }
  }

  &__tr {
    display: grid;
    min-height: 40px;
    border-bottom: 1px solid var(--n-border-color);
    background: var(--n-color);

    &--query {
      grid-template-columns: 36px 1fr 28px 1.4fr 40px;
      align-items: center;
    }

    &--auto-body {
      background: var(--n-color-embedded);
    }

    &--form {
      grid-template-columns: 36px minmax(72px, 1fr) 100px minmax(100px, 1.4fr) 40px;
      align-items: center;
    }
  }

  &__td {
    display: flex;
    align-items: center;
    min-width: 0;
    padding: 2px 6px;

    &--check {
      justify-content: center;
      padding-left: 8px;
    }
    &--eq {
      justify-content: center;
    }
    &--type {
      align-items: center;
    }
    &--grow {
      flex: 1;
    }
    &--value {
      align-items: center;
    }
    &--action {
      justify-content: center;
      padding-right: 4px;
    }
  }

  &__type-select {
    width: 100%;
    min-width: 0;
  }

  &__equals {
    color: var(--n-text-color-3);
    font-size: 14px;
    user-select: none;
  }

  &__value-cell {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    min-width: 0;
  }

  &__auto-hint {
    font-size: 12px;
    color: var(--n-text-color-3);
    flex-shrink: 0;
    white-space: nowrap;
  }

  &__cell-input {
    width: 100%;

    :deep(.n-input__input-el) {
      padding-left: 4px;
      padding-right: 4px;
    }
  }

  &__file-cell {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    width: 100%;
    min-height: 28px;
  }

  &__file-native {
    display: none;
  }

  &__file-label {
    display: inline-flex;
    align-items: center;
    cursor: pointer;
    flex-shrink: 0;
  }

  &__file-btn {
    display: inline-flex;
    align-items: center;
    padding: 4px 10px;
    font-size: 12px;
    line-height: 1.25;
    color: var(--n-primary-color);
    border: 1px dashed var(--n-border-color);
    border-radius: var(--n-border-radius);
    background: var(--n-color-modal);
  }

  &__file-name {
    font-size: 12px;
    line-height: 1.25;
    color: var(--n-text-color);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 100%;
  }

  &__remove {
    color: var(--n-text-color-3);

    &:hover {
      color: var(--n-error-color);
    }
  }

  &__row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
  }

  &__key {
    flex: 1 1 120px;
    min-width: 0;
    max-width: 240px;
  }

  &__val {
    flex: 2 1 180px;
    min-width: 0;
  }
}
</style>
