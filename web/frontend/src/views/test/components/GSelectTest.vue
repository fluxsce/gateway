<template>
  <div class="gselect-test-page">
    <div class="page-header">
      <h1>GSelect 选择器测试</h1>
      <p class="page-description">
        基于 Naive NSelect：单选、多选、filterable、disabled、size、分组选项
      </p>
    </div>

    <div class="test-sections">
      <section class="test-section">
        <h2>基础单选</h2>
        <div class="demo-row">
          <GSelect
            v-model:value="singleValue"
            :options="singleOptions"
            placeholder="请选择"
            style="max-width: 240px"
          />
        </div>
      </section>

      <section class="test-section">
        <h2>多选</h2>
        <div class="demo-row">
          <GSelect
            v-model:value="multipleValue"
            :options="singleOptions"
            multiple
            placeholder="多选"
            style="max-width: 320px"
          />
        </div>
      </section>

      <section class="test-section">
        <h2>可过滤</h2>
        <div class="demo-row">
          <GSelect
            v-model:value="filterValue"
            :options="filterOptions"
            filterable
            placeholder="输入搜索"
            style="max-width: 240px"
          />
        </div>
      </section>

      <section class="test-section">
        <h2>禁用</h2>
        <div class="demo-row">
          <GSelect
            v-model:value="singleValue"
            :options="singleOptions"
            disabled
            placeholder="禁用状态"
            style="max-width: 240px"
          />
        </div>
      </section>

      <section class="test-section">
        <h2>Size</h2>
        <div class="demo-row wrap">
          <GSelect
            v-model:value="singleValue"
            :options="singleOptions"
            size="small"
            placeholder="small"
            style="max-width: 160px"
          />
          <GSelect
            v-model:value="singleValue"
            :options="singleOptions"
            size="medium"
            placeholder="medium"
            style="max-width: 160px"
          />
          <GSelect
            v-model:value="singleValue"
            :options="singleOptions"
            size="large"
            placeholder="large"
            style="max-width: 160px"
          />
        </div>
      </section>

      <section class="test-section">
        <h2>分组选项</h2>
        <div class="demo-row">
          <GSelect
            v-model:value="groupValue"
            :options="groupOptions"
            placeholder="选择分组项"
            style="max-width: 240px"
          />
        </div>
      </section>

      <section class="test-section">
        <h2>不可清空</h2>
        <div class="demo-row">
          <GSelect
            v-model:value="singleValue"
            :options="singleOptions"
            :clearable="false"
            placeholder="不可清空"
            style="max-width: 240px"
          />
        </div>
      </section>

      <section class="test-section result-section">
        <h2>当前值</h2>
        <div class="result-output">
          <pre>{{ resultOutput }}</pre>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { GSelect } from '@/components'
import type { GSelectOption } from '@/components/gselect/types'
import { computed, ref } from 'vue'

defineOptions({ name: 'GSelectTest' })

const singleValue = ref<string | number | null>(null)
const multipleValue = ref<Array<string | number>>([])
const filterValue = ref<string | number | null>(null)
const groupValue = ref<string | number | null>(null)

const singleOptions: GSelectOption[] = [
  { label: '选项 A', value: 'a' },
  { label: '选项 B', value: 'b' },
  { label: '选项 C', value: 'c' },
  { label: '选项 D', value: 'd' },
]

const filterOptions: GSelectOption[] = [
  { label: '苹果', value: 'apple' },
  { label: '香蕉', value: 'banana' },
  { label: '橙子', value: 'orange' },
  { label: '葡萄', value: 'grape' },
  { label: '西瓜', value: 'watermelon' },
]

const groupOptions: GSelectOption[] = [
  {
    type: 'group',
    label: '水果',
    key: 'fruit',
    children: [
      { label: '苹果', value: 'apple' },
      { label: '香蕉', value: 'banana' },
    ],
  },
  {
    type: 'group',
    label: '蔬菜',
    key: 'veg',
    children: [
      { label: '胡萝卜', value: 'carrot' },
      { label: '土豆', value: 'potato' },
    ],
  },
]

const resultOutput = computed(() =>
  JSON.stringify(
    {
      singleValue: singleValue.value,
      multipleValue: multipleValue.value,
      filterValue: filterValue.value,
      groupValue: groupValue.value,
    },
    null,
    2
  )
)
</script>

<style scoped lang="scss">
.gselect-test-page {
  padding: var(--g-padding-lg);
  max-width: 900px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: var(--g-space-xl);

  h1 {
    font-size: 24px;
    font-weight: 600;
    color: var(--g-text-primary);
    margin: 0 0 var(--g-space-sm);
  }

  .page-description {
    font-size: var(--g-font-size-sm);
    color: var(--g-text-secondary);
    margin: 0;
  }
}

.test-sections {
  display: flex;
  flex-direction: column;
  gap: var(--g-space-xl);
}

.test-section {
  padding: var(--g-padding-md);
  background: var(--g-bg-secondary);
  border: 1px solid var(--g-border-primary);
  border-radius: var(--g-radius-lg);

  h2 {
    font-size: var(--g-font-size-lg);
    font-weight: 600;
    color: var(--g-text-primary);
    margin: 0 0 var(--g-space-md);
  }
}

.demo-row {
  display: flex;
  gap: var(--g-space-md);
  align-items: center;

  &.wrap {
    flex-wrap: wrap;
  }
}

.result-output {
  background: var(--g-bg-tertiary);
  border-radius: var(--g-radius-md);
  border: 1px solid var(--g-border-primary);
  padding: var(--g-padding-md);
  color: var(--g-text-secondary);
  font-size: var(--g-font-size-sm);
  overflow: auto;
}
</style>
