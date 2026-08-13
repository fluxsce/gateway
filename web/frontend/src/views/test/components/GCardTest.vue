<template>
  <div class="gcard-test-page">
    <div class="page-header">
      <h1>GCard 卡片测试</h1>
      <p class="page-description">
        RsCard 适配层：验证 showTitle / title、header / header-extra、footer / action、hoverable、bordered
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
              <RsButton size="sm" variant="ghost">更多</RsButton>
            </template>
            <div class="card-placeholder">header-extra → RsCard #actions（标题右侧）</div>
          </GCard>
        </div>
      </section>

      <section class="test-section">
        <h2>Footer / Action 插槽</h2>
        <div class="card-grid two-cols">
          <GCard show-title title="带底部">
            <div class="card-placeholder">内容区域</div>
            <template #footer>
              <div class="footer-placeholder">底部信息</div>
            </template>
          </GCard>
          <GCard show-title title="带操作区">
            <div class="card-placeholder">action 渲染在卡片底部（非头部）</div>
            <template #action>
              <RsButton size="sm" variant="primary">确认</RsButton>
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
        <div class="card-grid two-cols">
          <GCard show-title title="无边框（默认）" :bordered="false">
            <div class="card-placeholder">plain + borderless</div>
          </GCard>
          <GCard show-title title="带边框" bordered>
            <div class="card-placeholder">outlined</div>
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
          >
            <template #header-extra>
              <RsButton size="sm" variant="ghost">设置</RsButton>
            </template>
            <div class="card-placeholder">内容</div>
            <template #footer>
              <div class="footer-placeholder">底部</div>
            </template>
            <template #action>
              <RsButton size="sm" variant="primary">保存</RsButton>
            </template>
          </GCard>
        </div>
      </section>

      <section class="test-section test-section--gap">
        <h2>未透传能力（当前无效）</h2>
        <p class="gap-note">
          size / embedded / cover 在类型或旧测试中存在，但 GCard → RsCard 未实现；以下用例用于对照确认「无视觉差异」。
        </p>
        <div class="card-grid three-cols">
          <GCard show-title title="size=small" size="small">
            <div class="card-placeholder">size 未透传</div>
          </GCard>
          <GCard show-title title="embedded" embedded>
            <div class="card-placeholder">embedded 未透传</div>
          </GCard>
          <GCard show-title title="cover 插槽">
            <template #cover>
              <div class="cover-placeholder">封面（不会渲染）</div>
            </template>
            <div class="card-placeholder">#cover 未转发</div>
          </GCard>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { GCard } from '@/components/gcard'
import { GIcon } from '@/components/gicon'
import { RsButton } from '@/ui'

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

.test-section--gap {
  border-style: dashed;
  opacity: 0.92;
}

.gap-note {
  margin: calc(var(--g-space-md) * -0.5) 0 var(--g-space-md);
  font-size: var(--g-font-size-sm);
  color: var(--g-text-tertiary);
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

@media (max-width: 768px) {
  .card-grid.two-cols,
  .card-grid.three-cols {
    grid-template-columns: 1fr;
  }
}
</style>
