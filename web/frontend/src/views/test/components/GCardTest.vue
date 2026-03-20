<template>
  <div class="gcard-test-page">
    <div class="page-header">
      <h1>GCard 卡片测试</h1>
      <p class="page-description">
        基于 Naive NCard：标题、插槽（header/header-extra/action/cover/footer）、hoverable、bordered、size、embedded
      </p>
    </div>

    <div class="test-sections">
      <section class="test-section">
        <h2>无标题（默认）</h2>
        <div class="card-grid">
          <GCard>
            <div class="card-placeholder">默认内容区域，无 padding</div>
          </GCard>
        </div>
      </section>

      <section class="test-section">
        <h2>带标题</h2>
        <div class="card-grid">
          <GCard show-title title="卡片标题">
            <div class="card-placeholder">内容区域</div>
          </GCard>
        </div>
      </section>

      <section class="test-section">
        <h2>Header 插槽</h2>
        <div class="card-grid">
          <GCard show-title title="自定义标题">
            <template #header>
              <span class="custom-header">
                <GIcon icon="DocumentOutline" size="small" />
                自定义 Header
              </span>
            </template>
            <div class="card-placeholder">使用 header 插槽覆盖默认标题</div>
          </GCard>
        </div>
      </section>

      <section class="test-section">
        <h2>Header-Extra 插槽</h2>
        <div class="card-grid">
          <GCard show-title title="标题 + 右侧操作">
            <template #header-extra>
              <NButton size="tiny" quaternary>更多</NButton>
            </template>
            <div class="card-placeholder">header-extra 显示在标题右侧</div>
          </GCard>
        </div>
      </section>

      <section class="test-section">
        <h2>Action 插槽</h2>
        <div class="card-grid">
          <GCard show-title title="带操作区">
            <template #action>
              <NButton size="tiny" type="primary">确认</NButton>
            </template>
            <div class="card-placeholder">action 插槽在头部右侧</div>
          </GCard>
        </div>
      </section>

      <section class="test-section">
        <h2>Cover 插槽</h2>
        <div class="card-grid">
          <GCard show-title title="带封面">
            <template #cover>
              <div class="cover-placeholder">封面区域</div>
            </template>
            <div class="card-placeholder">cover 插槽在内容上方</div>
          </GCard>
        </div>
      </section>

      <section class="test-section">
        <h2>Footer 插槽</h2>
        <div class="card-grid">
          <GCard show-title title="带底部">
            <div class="card-placeholder">内容区域</div>
            <template #footer>
              <div class="footer-placeholder">底部信息</div>
            </template>
          </GCard>
        </div>
      </section>

      <section class="test-section">
        <h2>Hoverable</h2>
        <div class="card-grid two-cols">
          <GCard show-title title="hover 时阴影" :hoverable="true">
            <div class="card-placeholder">鼠标悬停显示阴影</div>
          </GCard>
          <GCard show-title title="always 阴影" hoverable="always">
            <div class="card-placeholder">始终显示阴影</div>
          </GCard>
        </div>
      </section>

      <section class="test-section">
        <h2>Bordered</h2>
        <div class="card-grid">
          <GCard show-title title="带边框" bordered>
            <div class="card-placeholder">bordered 显示边框</div>
          </GCard>
        </div>
      </section>

      <section class="test-section">
        <h2>Size</h2>
        <div class="card-grid three-cols">
          <GCard show-title title="small" size="small">
            <div class="card-placeholder">size="small"</div>
          </GCard>
          <GCard show-title title="medium" size="medium">
            <div class="card-placeholder">size="medium"（默认）</div>
          </GCard>
          <GCard show-title title="large" size="large">
            <div class="card-placeholder">size="large"</div>
          </GCard>
        </div>
      </section>

      <section class="test-section">
        <h2>Embedded</h2>
        <div class="card-grid">
          <GCard show-title title="嵌入模式" embedded>
            <div class="card-placeholder">embedded 降低视觉层级</div>
          </GCard>
        </div>
      </section>

      <section class="test-section">
        <h2>组合示例</h2>
        <div class="card-grid">
          <GCard
            show-title
            title="完整示例"
            bordered
            hoverable="hover"
            size="medium"
          >
            <template #header-extra>
              <NButton size="tiny" quaternary>设置</NButton>
            </template>
            <template #cover>
              <div class="cover-placeholder">封面</div>
            </template>
            <div class="card-placeholder">内容</div>
            <template #footer>
              <div class="footer-placeholder">底部</div>
            </template>
          </GCard>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { GCard, GIcon } from '@/components'
import { NButton } from 'naive-ui'

defineOptions({ name: 'GCardTest' })
</script>

<style scoped lang="scss">
.gcard-test-page {
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

.card-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--g-space-md);

  &.two-cols {
    grid-template-columns: repeat(2, 1fr);
  }

  &.three-cols {
    grid-template-columns: repeat(3, 1fr);
  }
}

.card-placeholder {
  padding: var(--g-padding-md);
  background: var(--g-bg-tertiary);
  border-radius: var(--g-radius-md);
  color: var(--g-text-secondary);
  font-size: var(--g-font-size-sm);
}

.cover-placeholder {
  padding: var(--g-padding-lg);
  background: linear-gradient(135deg, var(--g-primary-light) 0%, var(--g-bg-tertiary) 100%);
  color: var(--g-text-secondary);
  font-size: var(--g-font-size-sm);
  text-align: center;
}

.footer-placeholder {
  padding: var(--g-padding-sm);
  color: var(--g-text-tertiary);
  font-size: var(--g-font-size-xs);
}

.custom-header {
  display: inline-flex;
  align-items: center;
  gap: var(--g-space-xs);
}
</style>
