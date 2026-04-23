<template>
  <div class="route-view-loading-mask" aria-live="polite" aria-busy="true">
    <div class="route-view-loading-card">
      <div class="route-view-loading-title">加载中</div>
      <div class="route-view-loading-bar" />
    </div>
  </div>
</template>

<script setup lang="ts"></script>

<style scoped>
.route-view-loading-mask {
  position: absolute;
  inset: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  /*
    主题适配说明：
    - 不直接写死 light/dark 的颜色，全部从 --g-* 语义变量推导
    - 使用 color-mix 做“主色光晕 + 背景玻璃”组合，浅/深色都能保持对比度
  */
  background:
    radial-gradient(
      90% 70% at 85% 18%,
      color-mix(in srgb, var(--g-primary) 26%, transparent),
      transparent 60%
    ),
    radial-gradient(
      85% 60% at 10% 88%,
      color-mix(in srgb, var(--g-info) 16%, transparent),
      transparent 56%
    ),
    linear-gradient(
      152deg,
      color-mix(in srgb, var(--g-bg-tertiary) 68%, transparent) 0%,
      color-mix(in srgb, var(--g-bg-secondary) 58%, transparent) 55%,
      color-mix(in srgb, var(--g-bg-tertiary) 64%, transparent) 100%
    );
  backdrop-filter: blur(2px);
}

.route-view-loading-mask::before {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background-image:
    linear-gradient(color-mix(in srgb, var(--g-text-primary) 10%, transparent) 1px, transparent 1px),
    linear-gradient(90deg, color-mix(in srgb, var(--g-text-primary) 10%, transparent) 1px, transparent 1px);
  background-size: 28px 28px, 28px 28px;
  opacity: 0.28;
}

.route-view-loading-card {
  position: relative;
  width: min(420px, calc(100% - 44px));
  border-radius: 16px;
  padding: 14px 14px 12px;
  background: color-mix(in srgb, var(--g-bg-primary) 58%, transparent);
  border: 1px solid color-mix(in srgb, var(--g-border-primary) 78%, transparent);
  box-shadow: var(--g-shadow-md);
}

.route-view-loading-title {
  font-size: 13px;
  font-weight: 650;
  color: var(--g-text-primary);
  letter-spacing: 0.04em;
}

.route-view-loading-bar {
  margin-top: 10px;
  height: 3px;
  border-radius: 999px;
  overflow: hidden;
  background: color-mix(in srgb, var(--g-text-primary) 12%, transparent);
  position: relative;
}

.route-view-loading-bar::after {
  content: '';
  position: absolute;
  inset: 0;
  width: 32%;
  border-radius: 999px;
  background: linear-gradient(
    90deg,
    var(--g-primary) 0%,
    color-mix(in srgb, var(--g-primary) 60%, var(--g-info)) 55%,
    var(--g-info) 100%
  );
  animation: route-view-loading-slide 0.85s cubic-bezier(0.2, 0.9, 0.2, 1) infinite;
}

@keyframes route-view-loading-slide {
  0% {
    transform: translateX(-120%);
    opacity: 0.85;
  }
  50% {
    opacity: 1;
  }
  100% {
    transform: translateX(220%);
    opacity: 0.85;
  }
}
</style>

