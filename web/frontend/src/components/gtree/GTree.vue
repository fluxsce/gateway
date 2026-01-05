<template>
  <div class="g-tree-wrapper">
    <n-tree
      :data="treeData"
      :default-expanded-keys="props.defaultExpandedKeys"
      :checked-keys="props.checkedKeys !== undefined ? props.checkedKeys : undefined"
      :default-checked-keys="props.checkedKeys === undefined ? props.defaultCheckedKeys : undefined"
      :checkable="props.checkable"
      :cascade="props.cascade"
      :check-strategy="props.checkStrategy"
      :draggable="props.draggable"
      :show-line="props.showLine"
      :show-icon="props.showIcon"
      :block-line="props.blockLine"
      :node-props="nodeProps"
      :render-label="renderLabelWrapperComputed"
      :render-prefix="renderPrefixWrapper"
      :render-suffix="renderSuffixWrapper"
      :virtual-scroll="props.virtualScroll"
      @update:checked-keys="handleUpdateCheckedKeys"
      @update:expanded-keys="handleUpdateExpandedKeys"
      @select="handleSelect"
      @dblclick="handleDblclick"
    />
    <!-- 右键菜单：使用 x 和 y 属性定位 -->
    <n-dropdown
      v-if="props.menuConfig && props.menuConfig.enabled !== false"
      :options="dropdownOptions"
      :show="showContextMenu"
      :x="contextMenuX"
      :y="contextMenuY"
      trigger="manual"
      placement="bottom-start"
      @clickoutside="handleContextMenuClose"
      @select="handleDropdownSelect"
    />
  </div>
</template>

<script setup lang="ts">
import { GEllipsis } from '@/components/gellipsis'
import { useContextMenu } from '@/components/gmenu/useContextMenu'
import { renderIconVNode } from '@/utils'
import type { TreeOption } from 'naive-ui'
import { NDropdown, NIcon, NTree } from 'naive-ui'
import { computed, h, ref } from 'vue'
import type { GTreeEmits, GTreeProps } from './types'

defineOptions({
  name: 'GTree'
})

const props = withDefaults(defineProps<GTreeProps>(), {
  data: () => [],
  defaultExpandedKeys: () => [],
  defaultCheckedKeys: () => [],
  checkable: false,
  cascade: true,
  checkStrategy: 'all',
  draggable: false,
  showLine: false,
  showIcon: false,
  blockLine: true,
  keyField: 'key',
  labelField: 'label',
  childrenField: 'children',
  nodeHeight: 28,
  virtualScroll: false,
  ellipsis: false,
  ellipsisLineClamp: 1,
  ellipsisTooltip: true
})

const emit = defineEmits<GTreeEmits>()

// 转换数据格式，确保符合 NTree 的要求
const treeData = computed(() => {
  if (!props.data || props.data.length === 0) {
    return []
  }

  // 递归转换函数
  const convertItem = (item: any): TreeOption => {
    const option: TreeOption = {
      key: props.keyField === 'key' ? item.key : (item as any)[props.keyField],
      label: props.labelField === 'label' ? item.label : (item as any)[props.labelField]
    }

    // 处理子节点
    const children = (item as any)[props.childrenField] || item.children
    if (children && Array.isArray(children) && children.length > 0) {
      option.children = children.map(convertItem)
    }

    // 保留其他属性
    Object.keys(item).forEach((key) => {
      if (key !== props.keyField && key !== props.labelField && key !== props.childrenField && key !== 'key' && key !== 'label' && key !== 'children') {
        ;(option as any)[key] = item[key]
      }
    })

    return option
  }

  return props.data.map(convertItem)
})

// 右键菜单相关
const showContextMenu = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const currentContextNode = ref<TreeOption | null>(null)


// 关闭右键菜单
const handleContextMenuClose = () => {
  showContextMenu.value = false
  currentContextNode.value = null
}

// 使用右键菜单 Hook
const { dropdownOptions, handleMenuClick: handleMenuClickFromHook } = useContextMenu(
  props.menuConfig,
  props.moduleId,
  (menu) => {
    // 触发菜单点击事件，传递关联的数据
    emit('menu-click', {
      code: menu.code,
      node: menu.node ?? currentContextNode.value
    })
  }
)

// 处理下拉菜单选择
const handleDropdownSelect = (key: string | number) => {
  // 调用 hook 的处理方法（会触发复制等默认操作）
  handleMenuClickFromHook(
    key,
    currentContextNode.value, // node
    undefined, // row
    undefined  // column
  )
  
  // 关闭菜单
  handleContextMenuClose()
}

// 节点属性
const nodeProps = computed(() => {
  return ({ option }: { option: TreeOption }) => {
    const nodePropsObj: Record<string, any> = {
      style: {
        height: `${props.nodeHeight}px`
      },
      onClick: () => {
        // 手动触发 select 事件
        handleSelect([option.key as string], option)
      },
      'data-key': option.key
    }
    
    // 如果有菜单配置，添加右键菜单处理
    if (props.menuConfig && props.menuConfig.enabled !== false) {
      nodePropsObj.onContextmenu = (e: MouseEvent) => {
        e.preventDefault()
        e.stopPropagation()
        
        currentContextNode.value = option
        contextMenuX.value = e.clientX
        contextMenuY.value = e.clientY
        showContextMenu.value = true
      }
    }
    
    return nodePropsObj
  }
})

// 前缀渲染包装器：支持函数或图标名称字符串
const renderPrefixWrapper = props.renderPrefix
  ? typeof props.renderPrefix === 'string'
    ? (() => {
        const iconRenderFn = renderIconVNode(props.renderPrefix as string, NIcon)
        return () => iconRenderFn()
      })()
    : props.renderPrefix
  : undefined

// 后缀渲染包装器：支持函数或图标名称字符串
const renderSuffixWrapper = props.renderSuffix
  ? typeof props.renderSuffix === 'string'
    ? (() => {
        const iconRenderFn = renderIconVNode(props.renderSuffix as string, NIcon)
        return () => iconRenderFn()
      })()
    : props.renderSuffix
  : undefined

// Label 渲染包装器：支持省略显示
// 只有当用户提供了 renderLabel prop 或启用了 ellipsis 时才使用
const renderLabelWrapperComputed = computed(() => {
  // 如果用户没有提供 renderLabel prop 且没有启用 ellipsis，返回 undefined，让 naive-ui 使用默认渲染或 slot
  if (!props.renderLabel && !props.ellipsis) {
    return undefined
  }
  
  // 返回渲染函数
  return ({ option }: { option: TreeOption }) => {
    // 如果用户提供了 renderLabel，优先使用用户的
    if (props.renderLabel) {
      const userLabel = props.renderLabel({ option })
      // 如果启用了省略显示，用 GEllipsis 包装
      if (props.ellipsis && userLabel) {
        return h(GEllipsis, {
          lineClamp: props.ellipsisLineClamp,
          tooltip: props.ellipsisTooltip
        }, {
          default: () => userLabel
        })
      }
      return userLabel
    }
    
    // 如果没有自定义 renderLabel，但启用了省略显示，用 GEllipsis 包装默认 label
    if (props.ellipsis) {
      return h(GEllipsis, {
        text: option.label as string,
        lineClamp: props.ellipsisLineClamp,
        tooltip: props.ellipsisTooltip
      })
    }
    
    return undefined
  }
})

// 处理选中变化
const handleUpdateCheckedKeys = (keys: string[]) => {
  emit('update:checkedKeys', keys)
}

// 处理展开变化
const handleUpdateExpandedKeys = (keys: string[]) => {
  emit('update:expandedKeys', keys)
}

// 处理节点选择
const handleSelect = (keys: string[], option: TreeOption) => {
  emit('select', keys, option)
}

// 处理节点双击
const handleDblclick = (option: TreeOption) => {
  emit('dblclick', option)
}
</script>

<style scoped lang="scss">
.g-tree-wrapper {
  width: 100%;
  height: 100%;
}

:deep(.n-tree-node) {
  .n-tree-node-content {
    padding: 4px 8px;
    overflow: hidden;
    width: 100%;
    max-width: 100%;
    
    // 确保文本不换行，超出部分显示省略号
    .n-tree-node-content__text {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      max-width: 100%;
    }
    
    // 当使用自定义 label 时，确保内容不会溢出
    > * {
      max-width: 100%;
      overflow: hidden;
    }
  }
}
</style>

