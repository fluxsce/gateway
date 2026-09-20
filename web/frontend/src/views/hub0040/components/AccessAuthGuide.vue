<template>
  <div class="access-auth-guide">
    <p class="access-auth-guide__lead">
      开启后两种方式同时有效，接入方任选一种即可，不必在本页二选一。
    </p>
    <div class="access-auth-guide__grid">
      <section class="access-auth-guide__card">
        <h4>用户名密码</h4>
        <p>使用本平台已有用户账号。接入方填写用户名和密码即可，这里不用再配置账号。</p>
      </section>
      <section class="access-auth-guide__card">
        <h4>访问令牌</h4>
        <p>在实例右键「访问令牌」由服务端颁发，明文只出现一次。Java SDK 用 setAuthToken 带上即可。</p>
      </section>
    </div>
  </div>
</template>

<script lang="ts" setup>
defineOptions({
  name: 'AccessAuthGuide',
})
</script>

<style lang="scss" scoped>
.access-auth-guide {
  width: 100%;
}

.access-auth-guide__lead {
  margin: 0 0 10px;
  font-size: 13px;
  line-height: 20px;
  color: var(--g-text-secondary);
}

.access-auth-guide__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.access-auth-guide__card {
  min-width: 0;
  padding: 12px;
  border-radius: 12px;
  background: rgba(120, 120, 128, 0.1);
}

.access-auth-guide__card h4 {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  line-height: 18px;
  letter-spacing: -0.01em;
  color: var(--g-text-primary);
}

.access-auth-guide__card p {
  margin: 6px 0 0;
  font-size: 12px;
  line-height: 18px;
  color: var(--g-text-secondary);
}

@media (max-width: 720px) {
  .access-auth-guide__grid {
    grid-template-columns: 1fr;
  }
}

[data-theme='dark'] .access-auth-guide__card {
  background: rgba(235, 235, 245, 0.08);
}
</style>
