<template>
  <div
    class="g-toolbar"
    :class="[
      `g-toolbar--${align}`,
      {
        'g-toolbar--bordered': bordered,
        'g-toolbar--shadow': shadow
      }
    ]"
    :style="toolbarStyle"
  >
    <!-- 标题区域 -->
    <div v-if="title || slots.title" class="g-toolbar__title">
      <slot name="title">
        <span class="g-toolbar__title-text">{{ title }}</span>
      </slot>
    </div>

    <!-- 左侧插槽 -->
    <div v-if="slots.left" class="g-toolbar__left">
      <slot name="left" />
    </div>

    <!-- 主内容区域 -->
    <div class="g-toolbar__content">
      <slot>
        <!-- 渲染分组按钮（非尾部） -->
        <template v-if="groups && groups.length > 0">
          <div
            v-for="(group, groupIndex) in contentGroups"
            :key="group.key"
            class="g-toolbar__group"
            :class="{ 'g-toolbar__group--divider': group.divider && groupIndex > 0 }"
          >
            <span v-if="group.title" class="g-toolbar__group-title">{{ group.title }}</span>
            <div class="g-toolbar__group-buttons">
              <template v-for="(button, index) in group.buttons" :key="button.key">
                <toolbar-button-component
                  :button="button"
                  :module-id="props.moduleId"
                  :class="{ 'g-toolbar__button--divider': index > 0 }"
                  @click="handleButtonClick"
                  @dropdown-select="handleDropdownSelect"
                />
              </template>
            </div>
          </div>
        </template>

        <!-- 渲染扁平按钮（非尾部） -->
        <template v-else-if="buttons && buttons.length > 0">
          <template v-for="(button, index) in contentButtons" :key="button.key">
            <toolbar-button-component
              :button="button"
              :module-id="props.moduleId"
              :class="{ 'g-toolbar__button--divider': index > 0 }"
              @click="handleButtonClick"
              @dropdown-select="handleDropdownSelect"
            />
          </template>
        </template>
      </slot>
    </div>

    <!-- 中间插槽 -->
    <div v-if="slots.center" class="g-toolbar__center">
      <slot name="center" />
    </div>

    <!-- 右侧插槽（包含尾部按钮和自定义内容） -->
    <div v-if="(endButtons.length > 0 || endGroups.length > 0) || slots.right" class="g-toolbar__right">
      <!-- 渲染尾部分组按钮 -->
      <template v-if="groups && groups.length > 0">
        <div
          v-for="(group, groupIndex) in endGroups"
          :key="group.key"
          class="g-toolbar__group"
          :class="{ 'g-toolbar__group--divider': group.divider && groupIndex > 0 }"
        >
          <span v-if="group.title" class="g-toolbar__group-title">{{ group.title }}</span>
          <div class="g-toolbar__group-buttons">
            <template v-for="(button, index) in group.buttons" :key="button.key">
              <toolbar-button-component
                :button="button"
                :module-id="props.moduleId"
                :class="{ 'g-toolbar__button--divider': index > 0 }"
                @click="handleButtonClick"
                @dropdown-select="handleDropdownSelect"
              />
            </template>
          </div>
        </div>
      </template>

      <!-- 渲染尾部扁平按钮 -->
      <template v-else-if="buttons && buttons.length > 0">
        <template v-for="(button, index) in endButtons" :key="button.key">
          <toolbar-button-component
            :button="button"
            :module-id="props.moduleId"
            :class="{ 'g-toolbar__button--divider': index > 0 }"
            @click="handleButtonClick"
            @dropdown-select="handleDropdownSelect"
          />
        </template>
      </template>

      <!-- 自定义右侧插槽内容 -->
      <slot name="right" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, useSlots } from 'vue';
import ToolbarButtonComponent from './ToolbarButton.vue';
import type { ToolbarButton, ToolbarEmits, ToolbarProps } from './types';

// 定义组件名称
defineOptions({
  name: 'GToolbar'
})

// Props
const props = withDefaults(defineProps<ToolbarProps>(), {
  align: 'left',
  bordered: true,
  shadow: false
})

// Emits
const emit = defineEmits<ToolbarEmits>()

// Slots
const slots = useSlots()

// 计算工具栏样式
const toolbarStyle = computed(() => {
  return {
    height: props.height || 'var(--g-toolbar-height)'
  }
})


// 过滤可见按钮（包含权限检查）
const visibleButtons = computed(() => {
  return props.buttons?.filter(button => {
    // 如果设置了 show 为 false，则不显示
    if (button.show === false) {
      return false
    }
    return true
  }) || []
})

// 过滤尾部按钮（atEnd 为 true 的按钮）
const endButtons = computed(() => {
  return visibleButtons.value.filter(button => button.atEnd === true)
})

// 过滤非尾部按钮（atEnd 不为 true 的按钮）
const contentButtons = computed(() => {
  return visibleButtons.value.filter(button => button.atEnd !== true)
})

// 过滤可见分组（包含权限检查）
const visibleGroups = computed(() => {
  return props.groups?.map(group => ({
    ...group,
    buttons: group.buttons.filter(button => {
      // 如果设置了 show 为 false，则不显示
      if (button.show === false) {
        return false
      }
      return true
    })
  })).filter(group => group.buttons.length > 0) || []
})

// 过滤尾部分组（包含 atEnd 为 true 的按钮）
const endGroups = computed(() => {
  return visibleGroups.value.map(group => ({
    ...group,
    buttons: group.buttons.filter(button => button.atEnd === true)
  })).filter(group => group.buttons.length > 0)
})

// 过滤非尾部分组（不包含 atEnd 为 true 的按钮）
const contentGroups = computed(() => {
  return visibleGroups.value.map(group => ({
    ...group,
    buttons: group.buttons.filter(button => button.atEnd !== true)
  })).filter(group => group.buttons.length > 0)
})

// 处理按钮点击
const handleButtonClick = async (key: string) => {
  emit('button-click', key)
  
  // 查找按钮并执行其 onClick
  const button = findButton(key)
  if (button?.onClick) {
    await button.onClick(key)
  }
}

// 处理下拉菜单选择
const handleDropdownSelect = async (buttonKey: string, optionKey: string) => {
  emit('dropdown-select', buttonKey, optionKey)
  
  // 查找按钮和选项并执行 onClick
  const button = findButton(buttonKey)
  const option = button?.dropdownOptions?.find(opt => opt.key === optionKey)
  if (option?.onClick) {
    await option.onClick(optionKey)
  }
}

// 查找按钮
const findButton = (key: string): ToolbarButton | undefined => {
  if (props.buttons) {
    return props.buttons.find(btn => btn.key === key)
  }
  if (props.groups) {
    for (const group of props.groups) {
      const button = group.buttons.find(btn => btn.key === key)
      if (button) return button
    }
  }
  return undefined
}

// 暴露方法供外部调用
defineExpose({
  findButton
})
</script>

<style lang="scss" scoped>
.g-toolbar {
  display: flex;
  align-items: center;
  width: 100%;
  box-sizing: border-box;
  transition: all var(--g-transition-base) var(--g-transition-ease);
  gap: var(--g-space-xs);
  border-bottom: 1px solid var(--g-border-primary);

  // 对齐方式
  &--left {
    justify-content: flex-start;
  }

  &--center {
    justify-content: center;
  }

  &--right {
    justify-content: flex-end;
  }

  &--space-between {
    justify-content: space-between;
  }

  // 边框
  &--bordered {
    border-bottom: 1px solid var(--g-border-primary);
  }

  // 阴影
  &--shadow {
    box-shadow: var(--g-shadow-sm);
  }

  // 按钮竖线分隔符
  &__button--divider {
    position: relative;

    &::before {
      content: '';
      position: absolute;
      left: calc(var(--g-space-xs) * -1);
      top: 50%;
      transform: translateY(-50%);
      width: 1px;
      height: 60%;
      background-color: var(--g-border-primary);
    }
  }

  // 标题
  &__title {
    display: flex;
    align-items: center;
    font-size: var(--g-font-size-lg);
    font-weight: 600;
    color: var(--g-text-primary);
    white-space: nowrap;

    &-text {
      line-height: 1;
    }
  }

  // 主内容区域
  &__content {
    display: flex;
    align-items: center;
    gap: var(--g-space-xs);
    flex: 1;
    min-width: 0;
  }

  // 左侧、中间、右侧插槽
  &__left,
  &__center,
  &__right {
    display: flex;
    align-items: center;
    gap: var(--g-space-xs);
  }

  &__left {
    margin-right: auto;
  }

  &__center {
    margin: 0 auto;
  }

  &__right {
    margin-left: auto;
  }

  // 分组
  &__group {
    display: flex;
    align-items: center;
    gap: var(--g-space-xs);

    &--divider {
      position: relative;
      padding-left: var(--g-space-xs);

      &::before {
        content: '';
        position: absolute;
        left: 0;
        top: 50%;
        transform: translateY(-50%);
        width: 1px;
        height: 60%;
        background-color: var(--g-border-primary);
      }
    }

    &-title {
      font-size: var(--g-font-size-sm);
      color: var(--g-text-secondary);
      margin-right: var(--g-space-xs);
      white-space: nowrap;
    }

    &-buttons {
      display: flex;
      align-items: center;
      gap: var(--g-space-xs);
    }
  }
}

// 响应式
@media (max-width: 768px) {
  .g-toolbar {
    gap: var(--g-space-xs);

    &__title {
      font-size: var(--g-font-size-base);
    }

    &__group {
      &-title {
        display: none;
      }
    }
  }
}
</style>

