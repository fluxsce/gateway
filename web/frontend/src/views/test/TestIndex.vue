<!-- 测试页面入口，提供组件测试页的快速访问 -->
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
        <RsButton variant="ghost" size="sm" icon="arrow-left" @click="goBack">
          返回首页
        </RsButton>
        <RsButton variant="ghost" size="sm" icon="book-open" @click="openDocs">
          查看文档
        </RsButton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { moduleApiPrefix, requestPathHelper } from '@/api/requestPath'
import { GIcon } from '@/components/gicon'
import { useAppMessage } from '@/composables/useAppMessage'
import { RsButton } from '@/ui'
import { useRouter } from 'vue-router'

defineOptions({ name: 'TestIndex' })

const router = useRouter()
const message = useAppMessage()

interface TestPage {
  path: string
  title: string
  description: string
  icon: string
}

const testPages: TestPage[] = [
  {
    path: '/test/gtoolbar',
    title: 'GToolbar 工具栏',
    description: '测试 GToolbar / ToolbarButton：扁平按钮、分组、tooltip、dropdown、权限 key',
    icon: 'ConstructOutline',
  },
  {
    path: '/test/rs-search-form',
    title: 'RsSearchForm 查询表单',
    description: '字段栅格、更多条件、内置查询/重置、校验、label 布局',
    icon: 'SearchOutline',
  },
  {
    path: '/test/rs-data-form',
    title: 'RsDataForm 数据表单',
    description: '页签、fieldset、主键禁用、mode、校验；弹窗基于 RsDialog',
    icon: 'DocumentTextOutline',
  },
  {
    path: '/test/rs-grid',
    title: 'RsGrid 表格',
    description: '勾选 + 序号、工具栏、分页、排序筛选、右键菜单、暴露方法',
    icon: 'GridOutline',
  },
  {
    path: '/test/rs-button',
    title: 'RsButton 按钮',
    description: 'primary/secondary 主次、边框对比、尺寸、loading、图标、圆角',
    icon: 'EllipseOutline',
  },
  {
    path: '/test/rs-card',
    title: 'RsCard 卡片',
    description: '标题、插槽、variant、size、hoverable、elevated、borderless、padding',
    icon: 'GridOutline',
  },
  {
    path: '/test/http-probe',
    title: 'HTTP 探测（监控列表）',
    description: `POST ${requestPathHelper.join(moduleApiPrefix('hub0000'), 'server/query')}：对比后端原文与 createApi 解包结果`,
    icon: 'CloudOutline',
  },
]

function goBack() {
  router.push('/')
}

function openDocs() {
  message.info('文档功能开发中...')
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
  border-bottom: 1px solid var(--g-border-primary);

  h1 {
    font-size: 28px;
    font-weight: 600;
    color: var(--g-text-primary);
    margin: 0 0 var(--g-space-md);
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
  border-radius: var(--g-radius-md, 8px);
  text-decoration: none;
  transition: border-color var(--g-transition-base, 0.2s ease);
  cursor: pointer;

  &:hover {
    border-color: var(--g-primary);

    .card-icon {
      background: var(--g-primary);
      color: white;
    }

    .card-arrow {
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
    border-radius: var(--g-radius-md, 8px);
    background: var(--g-primary-light);
    color: var(--g-primary);
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
  }
}

.quick-actions {
  padding: var(--g-padding-xl);
  background: var(--g-bg-secondary);
  border-radius: var(--g-radius-md, 8px);
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
