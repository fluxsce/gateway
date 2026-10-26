<template>
  <div class="tool-page">
    <RsLoading v-if="loading" block size="lg" />
    <RsEmpty v-else-if="denied" :description="t('work.denied')" fill />
    <RsAlert v-else-if="failed" type="danger" :description="failed" />

    <template v-else>
      <header class="tool-page__head">
        <span class="tool-page__mark" aria-hidden="true">
          <RsIcon :name="icon" :size="20" />
        </span>
        <div class="tool-page__titles">
          <h1>{{ title }}</h1>
          <p>{{ description }}</p>
        </div>
        <div class="tool-page__actions">
          <slot name="actions" />
        </div>
      </header>

      <RsAlert v-if="error" type="danger" size="sm" :description="error" />

      <div class="tool-page__body">
        <slot />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { RsAlert, RsEmpty, RsIcon, RsLoading } from '@/ui'

defineProps<{
  title: string
  description: string
  icon: string
  error?: string
  loading?: boolean
  denied?: boolean
  failed?: string
}>()

const { t } = useModuleI18n('hub0010')
</script>

<style lang="scss" scoped>
.tool-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  box-sizing: border-box;
  padding: 16px 20px 20px;
  gap: 12px;
  background:
    radial-gradient(1200px 240px at 0% 0%, color-mix(in srgb, var(--rs-primary) 8%, transparent), transparent 60%),
    var(--rs-bg);
}

.tool-page__head {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 14px;
  flex: none;
  width: 100%;
}

.tool-page__mark {
  display: grid;
  place-items: center;
  width: 40px;
  height: 40px;
  flex: none;
  border-radius: 12px;
  color: var(--rs-primary);
  background: color-mix(in srgb, var(--rs-primary) 12%, var(--rs-surface));
  border: 1px solid color-mix(in srgb, var(--rs-primary) 22%, var(--rs-border));
}

.tool-page__titles {
  min-width: 0;

  h1 {
    margin: 0;
    font-size: 18px;
    font-weight: 650;
    letter-spacing: -0.02em;
    line-height: 1.3;
  }

  p {
    margin: 2px 0 0;
    color: var(--rs-muted);
    font-size: 13px;
    line-height: 1.45;
  }
}

.tool-page__actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  flex-wrap: wrap;
}

@media (max-width: 860px) {
  .tool-page__head {
    grid-template-columns: auto minmax(0, 1fr);
  }

  .tool-page__actions {
    grid-column: 1 / -1;
    justify-content: flex-start;
  }
}

.tool-page__body {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  min-width: 0;

  > :deep(*) {
    flex: 1 1 auto;
    min-height: 0;
    min-width: 0;
  }
}
</style>
