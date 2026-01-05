<template>
  <div class="g-drawer">
    <n-drawer
    :show="props.show"
    :width="drawerWidth"
    :placement="props.placement"
    :mask-closable="props.maskClosable"
    :closable="props.closable"
    :auto-focus="props.autoFocus"
    :show-mask="props.mask"
    :block-scroll="props.blockScroll"
    :to="props.to === false ? undefined : props.to"
    :resizable="props.resizable"
    :style="drawerStyle"
    :trap-focus="false"
    @update:show="handleUpdateShow"
    @after-enter="emit('after-enter')"
    @after-leave="emit('after-leave')"
  >
    <n-drawer-content
      :title="props.title"
      :closable="props.closable"
      :header-class="props.headerClass"
      :header-style="headerStyleComputed"
      :footer-class="props.footerClass"
      :footer-style="footerStyleComputed"
      :body-class="props.bodyClass"
      :body-style="props.bodyStyle"
      :body-content-class="props.bodyContentClass"
      :body-content-style="props.bodyContentStyle"
    >
      <!-- 头部插槽：与 GModal 保持一致的布局结构 -->
      <template #header>
        <div class="g-drawer__header">
          <div class="g-drawer__header-main">
            <slot name="header">
              <span class="g-drawer__title">{{ props.title }}</span>
            </slot>
          </div>
        </div>
      </template>

      <!-- 主体内容：固定在头部和底部之间，仅内容区域滚动 -->
      <div class="g-drawer__body">
        <slot />
      </div>

      <!-- 底部操作区，支持关闭 / 确认按钮，或完全自定义 -->
      <template #footer>
        <slot name="footer">
          <div v-if="props.showFooter" class="g-drawer__footer">
            <n-space justify="end" :size="8">
              <n-button v-if="props.showCancel" size="small" @click="handleCancel">
                {{ props.cancelText }}
              </n-button>
              <n-button
                v-if="props.showConfirm"
                type="primary"
                size="small"
                :loading="props.confirmLoading"
                @click="handleConfirm"
              >
                {{ props.confirmText }}
              </n-button>
            </n-space>
          </div>
        </slot>
      </template>
    </n-drawer-content>
  </n-drawer>
  </div>
</template>

<script setup lang="ts">
import { NButton, NDrawer, NDrawerContent, NSpace } from 'naive-ui';
import { computed } from 'vue';
import type { GDrawerEmits, GDrawerProps } from './types';

defineOptions({
  name: 'GDrawer'
})

const props = withDefaults(defineProps<GDrawerProps>(), {
  show: false,
  title: '',
  // 默认宽度为 400px
  width: 400,
  placement: 'right',
  mask: true,
  // 默认阻止背景滚动
  blockScroll: true,
  maskClosable: true,
  closable: true,
  showFooter: true,
  showCancel: true,
  showConfirm: true,
  cancelText: '取消',
  confirmText: '确定',
  confirmLoading: false,
  autoFocus: true,
  resizable: false
})

const emit = defineEmits<GDrawerEmits>()

const drawerWidth = computed(() => {
  if (typeof props.width === 'number') {
    return `${props.width}px`
  }
  return props.width
})

const drawerStyle = computed(() => {
  if (!props.style) return undefined

  if (Array.isArray(props.style)) {
    return props.style
  }

  return props.style
})

// 计算 header 样式，合并默认样式和用户自定义样式
const headerStyleComputed = computed(() => {
  const defaultStyle = {
    height: 'var(--g-drawer-header-height)',
    minHeight: 'var(--g-drawer-header-height)',
    maxHeight: 'var(--g-drawer-header-height)',
  }

  if (!props.headerStyle) {
    return defaultStyle
  }

  if (typeof props.headerStyle === 'string') {
    return props.headerStyle
  }

  // 合并对象样式
  return {
    ...defaultStyle,
    ...props.headerStyle
  }
})

// 计算 footer 样式，合并默认样式和用户自定义样式
const footerStyleComputed = computed(() => {
  const defaultStyle = {
    height: 'var(--g-drawer-footer-height)',
    minHeight: 'var(--g-drawer-footer-height)',
    maxHeight: 'var(--g-drawer-footer-height)'
  }

  if (!props.footerStyle) {
    return defaultStyle
  }

  if (typeof props.footerStyle === 'string') {
    return props.footerStyle
  }

  // 合并对象样式
  return {
    ...defaultStyle,
    ...props.footerStyle
  }
})

const handleUpdateShow = (value: boolean) => {
  emit('update:show', value)
  if (!value) {
    emit('close')
  }
}

const handleCancel = () => {
  emit('cancel')
  emit('update:show', false)
}

const handleConfirm = () => {
  emit('confirm')
}
</script>

<style scoped lang="scss">
/* 
 * 注意：header 和 footer 的高度和样式现在通过 n-drawer-content 的 
 * header-style 和 footer-style props 来设置，这样更可靠且不需要
 * 依赖 CSS 选择器覆盖。如果用户需要额外的样式覆盖，可以使用
 * header-class 和 footer-class props。
 */

.g-drawer__header {
  display: flex;
  align-items: center;
  height: 100%;
  width: 100%;
  /* 去掉左侧内边距，让标题更靠近左侧图标，仅保留右侧间距 */
  padding: 0;
  box-sizing: border-box;
}

.g-drawer__header-main {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
  height: 100%;
}

.g-drawer__title {
  font-size: var(--g-font-size-lg);
  font-weight: 500;
  color: var(--g-text-primary);
}

.g-drawer__body {
  padding: var(--g-space-sm);
  overflow-y: auto;
}

.g-drawer__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  height: 100%;
  width: 100%;
  padding: 0;
  box-sizing: border-box;
}
</style>

