<!-- 测试页面入口，提供所有测试页面的快速访问 -->
<template>
  <div class="test-index-page">
    <div class="page-header">
      <h1>Gateway 组件测试中心</h1>
      <p>快速访问各个组件的测试页面</p>
    </div>

    <div class="test-cards">
      <router-link
        v-for="test in testPages"
        :key="test.path"
        :to="test.path"
        class="test-card"
      >
        <div class="card-icon">
          <GIcon :icon="test.icon" :size="32" />
        </div>
        <div class="card-content">
          <h3>{{ test.title }}</h3>
          <p>{{ test.description }}</p>
        </div>
        <div class="card-arrow">
          <GIcon icon="ChevronForwardOutline" size="small" />
        </div>
      </router-link>
    </div>

    <div class="quick-actions">
      <h2>快速操作</h2>
      <div class="action-buttons">
        <NButton quaternary @click="goBack">
          <GIcon icon="ArrowBackOutline" size="small" class="btn-icon" />
          返回首页
        </NButton>
        <NButton quaternary @click="openDocs">
          <GIcon icon="BookOutline" size="small" class="btn-icon" />
          查看文档
        </NButton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { NButton } from 'naive-ui'
import { GIcon } from '@/components/gicon'
import { useRouter } from 'vue-router'
import { getCurrentInstance } from 'vue'

defineOptions({ name: 'TestIndex' })

const router = useRouter()
const instance = getCurrentInstance()
const $gMessage = instance?.appContext.config.globalProperties.$gMessage

interface TestPage {
  path: string
  title: string
  description: string
  icon: string
}

const testPages: TestPage[] = [
  {
    path: '/test/message',
    title: 'Message 消息组件',
    description: '测试 $gMessage 消息提示：info / success / error / warning / loading',
    icon: 'ChatbubbleOutline',
  },
  {
    path: '/test/custom-render',
    title: '全局自定义渲染',
    description: '测试 $gRender.show()：TS 内直接打开弹窗，无需在模板中引入组件',
    icon: 'CodeOutline',
  },
  {
    path: '/test/gtabs',
    title: 'GTabs 标签页',
    description: '多标签、拖拽排序、关闭、右键菜单、溢出下拉；line / card 类型',
    icon: 'LayersOutline',
  },
  {
    path: '/test/gtext-show',
    title: 'GTextShow 文本展示',
    description: '多格式文本展示：JSON / XML / 纯文本，复制、格式化、行号',
    icon: 'DocumentTextOutline',
  },
  {
    path: '/test/gdropdown',
    title: 'GDropdown 下拉菜单',
    description: '测试 GDropdown：options、placement、click/hover 触发与 @select',
    icon: 'EllipsisHorizontalOutline',
  },
  {
    path: '/test/gcard',
    title: 'GCard 卡片',
    description: '测试 GCard：标题、插槽、hoverable、bordered、size、embedded',
    icon: 'CardOutline',
  },
  {
    path: '/test/gselect',
    title: 'GSelect 选择器',
    description: '测试 GSelect：单选、多选、filterable、disabled、size、分组选项',
    icon: 'ChevronDownOutline',
  },
  {
    path: '/test/gdialog',
    title: 'GDialog 对话框',
    description: '测试 GDialog：标题/副标题/图标、渐变头部、宽度、滚动、拖拽、confirmLoading、自定义插槽',
    icon: 'ChatboxOutline',
  },
]

function goBack() {
  router.push('/')
}

function openDocs() {
  $gMessage?.info('文档功能开发中...')
}
</script>

<style scoped lang="scss">
.test-index-page {
  padding: var(--g-padding-xxl, 24px);
  max-width: 1200px;
  margin: 0 auto;
  background: var(--g-bg-primary);
  min-height: 100%;
}

.page-header {
  text-align: center;
  margin-bottom: var(--g-space-xxl, 32px);
  padding-bottom: var(--g-space-xl, 24px);
  border-bottom: 2px solid var(--g-border-primary);

  h1 {
    font-size: 28px;
    font-weight: 600;
    color: var(--g-text-primary);
    margin: 0 0 var(--g-space-md);
    background: linear-gradient(135deg, var(--g-primary) 0%, #8b5cf6 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  p {
    font-size: var(--g-font-size-lg, 16px);
    color: var(--g-text-secondary);
    margin: 0;
  }
}

.test-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--g-space-lg, 16px);
  margin-bottom: var(--g-space-xxl);
}

.test-card {
  display: flex;
  align-items: center;
  gap: var(--g-space-md);
  padding: var(--g-padding-lg);
  background: var(--g-bg-secondary);
  border: 1px solid var(--g-border-primary);
  border-radius: 12px;
  text-decoration: none;
  transition: all var(--g-transition-base, 0.2s ease);
  cursor: pointer;

  &:hover {
    transform: translateY(-4px);
    border-color: var(--g-primary);
    box-shadow: 0 8px 24px rgba(124, 58, 237, 0.15);

    .card-icon {
      background: linear-gradient(135deg, var(--g-primary) 0%, #8b5cf6 100%);
      color: white;
    }

    .card-arrow {
      transform: translateX(4px);
      color: var(--g-primary);
    }
  }

  .card-icon {
    flex-shrink: 0;
    width: 56px;
    height: 56px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 12px;
    background: var(--g-primary-light);
    color: var(--g-primary);
    transition: all var(--g-transition-base);
  }

  .card-content {
    flex: 1;
    min-width: 0;

    h3 {
      font-size: var(--g-font-size-lg);
      font-weight: 600;
      color: var(--g-text-primary);
      margin: 0 0 var(--g-space-xs);
    }

    p {
      font-size: var(--g-font-size-sm);
      color: var(--g-text-secondary);
      margin: 0;
      line-height: 1.5;
    }
  }

  .card-arrow {
    flex-shrink: 0;
    color: var(--g-text-tertiary);
    transition: all var(--g-transition-base);
  }
}

.quick-actions {
  padding: var(--g-padding-xl);
  background: var(--g-bg-secondary);
  border-radius: 12px;
  border: 1px solid var(--g-border-primary);

  h2 {
    font-size: var(--g-font-size-lg);
    font-weight: 600;
    color: var(--g-text-primary);
    margin: 0 0 var(--g-space-md);
  }

  .action-buttons {
    display: flex;
    gap: var(--g-space-sm);
    flex-wrap: wrap;

    .btn-icon {
      margin-right: var(--g-space-xs);
      vertical-align: middle;
    }
  }
}

@media (max-width: 768px) {
  .test-cards {
    grid-template-columns: 1fr;
  }

  .page-header h1 {
    font-size: 22px;
  }
}
</style>
