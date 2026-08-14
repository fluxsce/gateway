<template>
  <div class="rs-card-test-page">
    <div class="page-header">
      <h1>RsCard 卡片测试</h1>
      <p class="page-description">
        覆盖 title / description、header / actions / footer / cover、variant、size、hoverable、elevated、borderless、padding。
      </p>
    </div>

    <div class="test-sections">
      <section class="test-section">
        <h2>无标题（默认）</h2>
        <div class="card-grid">
          <RsCard :padding="false">
            <div class="card-placeholder">默认 grouped，无标题</div>
          </RsCard>
        </div>
      </section>

      <section class="test-section">
        <h2>标题 / 描述</h2>
        <div class="card-grid two-cols">
          <RsCard title="卡片标题">
            <div class="card-placeholder">仅标题</div>
          </RsCard>
          <RsCard title="卡片标题" description="辅助说明文字">
            <div class="card-placeholder">标题 + description</div>
          </RsCard>
        </div>
      </section>

      <section class="test-section">
        <h2>Header / Actions</h2>
        <div class="card-grid two-cols">
          <RsCard title="自定义标题">
            <template #header>
              <span class="custom-header">
                <GIcon icon="DocumentOutline" size="small" />
                自定义 Header
              </span>
            </template>
            <div class="card-placeholder">#header 覆盖默认标题</div>
          </RsCard>
          <RsCard title="标题 + 右侧操作">
            <template #actions>
              <RsButton size="sm" variant="ghost">更多</RsButton>
            </template>
            <div class="card-placeholder">#actions 在标题右侧</div>
          </RsCard>
        </div>
      </section>

      <section class="test-section">
        <h2>Footer / Cover</h2>
        <div class="card-grid two-cols">
          <RsCard title="带底部">
            <div class="card-placeholder">内容区域</div>
            <template #footer>
              <RsButton size="sm" variant="secondary">取消</RsButton>
              <RsButton size="sm" variant="primary">确认</RsButton>
            </template>
          </RsCard>
          <RsCard title="带封面" :padding="false" clip>
            <template #cover>
              <div class="cover-placeholder">封面</div>
            </template>
            <div class="card-placeholder">#cover + clip 裁切圆角</div>
          </RsCard>
        </div>
      </section>

      <section class="test-section">
        <h2>Hoverable / Elevated</h2>
        <div class="card-grid two-cols">
          <RsCard title="hover 抬升" hoverable>
            <div class="card-placeholder">鼠标悬停显示阴影</div>
          </RsCard>
          <RsCard title="始终抬升" elevated>
            <div class="card-placeholder">elevated 常驻阴影</div>
          </RsCard>
        </div>
      </section>

      <section class="test-section">
        <h2>Borderless / Padding</h2>
        <div class="card-grid two-cols">
          <RsCard title="无外框" borderless variant="plain">
            <div class="card-placeholder">borderless + plain，适合嵌在内容区</div>
          </RsCard>
          <RsCard title="贴边正文" :padding="false">
            <div class="card-placeholder">padding=false，适合图表/表格</div>
          </RsCard>
        </div>
      </section>

      <section class="test-section">
        <h2>Variant</h2>
        <div class="card-grid three-cols">
          <RsCard title="grouped" variant="grouped">
            <div class="card-placeholder">macOS 分组面板</div>
          </RsCard>
          <RsCard title="outlined" variant="outlined">
            <div class="card-placeholder">Material 描边</div>
          </RsCard>
          <RsCard title="filled" variant="filled">
            <div class="card-placeholder">Material 填充</div>
          </RsCard>
        </div>
      </section>

      <section class="test-section">
        <h2>Size</h2>
        <div class="card-grid three-cols">
          <RsCard title="size=sm" size="sm">
            <div class="card-placeholder">紧凑</div>
          </RsCard>
          <RsCard title="size=md" size="md">
            <div class="card-placeholder">默认</div>
          </RsCard>
          <RsCard title="size=lg" size="lg">
            <div class="card-placeholder">宽松</div>
          </RsCard>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { GIcon } from '@/components/gicon'
import { RsButton, RsCard } from '@/ui'

defineOptions({ name: 'RsCardTest' })
</script>

<style scoped lang="scss">
.rs-card-test-page {
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
