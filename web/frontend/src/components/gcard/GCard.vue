<template>
  <RsCard
    :class="[
      'g-card',
      {
        'g-card--no-title': !showTitle,
        'g-card--hoverable': hoverable === true || hoverable === 'hover',
        'g-card--shadow-always': hoverable === 'always',
      },
      props.class,
    ]"
    :title="showTitle ? title || '' : undefined"
    :padding="false"
    :borderless="!bordered"
    :variant="bordered ? 'outlined' : 'plain'"
    :style="style"
  >
    <template v-if="showTitle && $slots.header" #header>
      <slot name="header" />
    </template>
    <template v-if="showTitle && $slots['header-extra']" #actions>
      <slot name="header-extra" />
    </template>
    <div class="g-card__body" :style="computedContentStyle">
      <slot />
    </div>
    <div v-if="$slots.footer || $slots.action" class="g-card__footer">
      <slot name="footer" />
      <slot name="action" />
    </div>
  </RsCard>
</template>

<script setup lang="ts">
import { RsCard } from '@/ui'
import { computed } from 'vue'
import type { GCardEmits, GCardProps } from './types'

defineOptions({
  name: 'GCard',
})

const props = withDefaults(defineProps<GCardProps>(), {
  showTitle: false,
  bordered: false,
  hoverable: false,
  size: 'medium',
  embedded: false,
})

defineEmits<GCardEmits>()

const computedContentStyle = computed(() => {
  if (props.contentStyle) return props.contentStyle
  return { padding: 0 }
})
</script>

<style scoped lang="scss">
.g-card {
  width: 100%;
}

.g-card__body {
  min-width: 0;
}

.g-card__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border-top: 1px solid var(--g-border-primary, var(--rs-border));
}

.g-card--hoverable:hover {
  box-shadow: var(--g-shadow-md, 0 4px 12px rgba(0, 0, 0, 0.08));
}

.g-card--shadow-always {
  box-shadow: var(--g-shadow-md, 0 4px 12px rgba(0, 0, 0, 0.08));
}
</style>
