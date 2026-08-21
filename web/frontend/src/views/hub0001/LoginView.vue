<template>
  <div class="login-view">
    <div class="global-brand-lockup" aria-label="FLUX Datahub Gateway">
      <img src="@/assets/images/logo.png" alt="" class="global-brand-lockup__logo" />
      <span class="global-brand-lockup__text">FLUX Datahub Gateway</span>
    </div>

    <div class="global-language-switch">
      <LanguageSwitcher variant="dark-surface" />
    </div>

    <div class="login-split">
      <aside class="login-split__brand" aria-label="Branding">
        <div class="welcome-area">
          <div class="welcome-area__body">
            <h1 class="welcome-area__title">{{ t('login.welcomeTitle') }}</h1>
            <p class="welcome-area__subtitle">{{ t('login.welcomeSubtitle') }}</p>

            <div class="features">
              <div class="feature-item">
                <div class="feature-icon">
                  <ShieldCheckmarkOutline />
                </div>
                <div class="feature-text">
                  <h3>{{ t('login.featureSecurityTitle') }}</h3>
                  <p>{{ t('login.featureSecurityDesc') }}</p>
                </div>
              </div>
              <div class="feature-item">
                <div class="feature-icon">
                  <DocumentTextOutline />
                </div>
                <div class="feature-text">
                  <h3>{{ t('login.featureAnalyticsTitle') }}</h3>
                  <p>{{ t('login.featureAnalyticsDesc') }}</p>
                </div>
              </div>
              <div class="feature-item">
                <div class="feature-icon">
                  <PeopleOutline />
                </div>
                <div class="feature-text">
                  <h3>{{ t('login.featureCollaborationTitle') }}</h3>
                  <p>{{ t('login.featureCollaborationDesc') }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </aside>

      <main class="login-split__form" aria-label="Login">
        <div class="form-area">
          <!-- 跟随 App 全局主题；控件 radius 由组件 props 透传 -->
          <RsCard class="login-card" variant="glass" elevated radius="lg">
            <div v-if="showForgotPassword" class="forgot-panel">
              <h2 class="forgot-panel__title">{{ t('forgotPassword.title') }}</h2>
              <p class="forgot-panel__subtitle">{{ t('forgotPassword.subtitle') }}</p>
              <RsAlert type="info" class="forgot-panel__hint">
                {{ t('forgotPassword.contactAdmin') }}
              </RsAlert>
              <RsButton
                variant="primary"
                size="lg"
                radius="sm"
                class="login-btn"
                @click="backToLogin"
              >
                {{ t('forgotPassword.backToLogin') }}
              </RsButton>
            </div>
            <RsTabs
              v-else
              v-model="activeTab"
              :items="loginTabItems"
              variant="line"
              size="md"
              borderless
              content-gap="xl"
              justify="stretch"
              class="login-tabs"
            >
              <template #account>
                <RsForm
                  v-if="activeTab === 'account'"
                  ref="formRef"
                  size="md"
                  gap="md"
                  class="login-form"
                  :rules="accountRules"
                >
                  <RsInput
                    v-model="formData.userId"
                    name="userId"
                    radius="sm"
                    size="lg"
                    :placeholder="t('login.userId')"
                    @keyup.enter="handleLogin"
                  >
                    <template #prefix>
                      <GIcon>
                        <PersonOutline />
                      </GIcon>
                    </template>
                  </RsInput>

                  <RsInput
                    v-model="formData.password"
                    name="password"
                    type="password"
                    radius="sm"
                    size="lg"
                    :placeholder="t('login.password')"
                    @keyup.enter="handleLogin"
                  >
                    <template #prefix>
                      <GIcon>
                        <LockClosedOutline />
                      </GIcon>
                    </template>
                  </RsInput>

                  <div class="captcha-area">
                    <div class="captcha-area__field">
                      <RsInput
                        v-model="formData.captchaCode"
                        name="captchaCode"
                        radius="sm"
                        size="lg"
                        :maxlength="6"
                        autocomplete="off"
                        :placeholder="t('login.captchaPlaceholder')"
                        @keyup.enter="handleLogin"
                      >
                        <template #prefix>
                          <GIcon>
                            <ShieldOutline />
                          </GIcon>
                        </template>
                      </RsInput>
                    </div>
                    <button
                      type="button"
                      class="captcha-img"
                      tabindex="-1"
                      :aria-label="t('login.captcha')"
                      @click="refreshCaptcha"
                    >
                      <img v-if="captchaUrl" :src="captchaUrl" alt="" />
                      <span v-else class="captcha-loading">
                        <RsLoading size="sm" />
                      </span>
                    </button>
                  </div>

                  <div class="form-options">
                    <RsButton variant="link" size="sm" radius="sm" @click="goToForgotPassword">
                      {{ t('login.forgotPassword') }}
                    </RsButton>
                  </div>

                  <p v-if="lockRemainSeconds > 0" class="login-cooldown">
                    {{ t('login.cooldownHint', { seconds: lockRemainSeconds }) }}
                  </p>

                  <RsButton
                    variant="primary"
                    size="lg"
                    radius="sm"
                    class="login-btn"
                    :loading="loading"
                    :disabled="lockRemainSeconds > 0"
                    @click="handleLogin"
                  >
                    {{
                      lockRemainSeconds > 0
                        ? t('login.cooldownButton', { seconds: lockRemainSeconds })
                        : t('login.loginButton')
                    }}
                  </RsButton>
                </RsForm>
              </template>

              <template #phone>
                <RsForm
                  v-if="activeTab === 'phone'"
                  ref="phoneFormRef"
                  size="md"
                  gap="md"
                  class="login-form"
                  :rules="phoneRules"
                >
                  <RsInput
                    v-model="phoneFormData.phone"
                    name="phone"
                    radius="sm"
                    size="lg"
                    :placeholder="t('login.phoneNumber')"
                    @keyup.enter="handlePhoneLogin"
                  >
                    <template #prefix>
                      <GIcon>
                        <PhonePortraitOutline />
                      </GIcon>
                    </template>
                  </RsInput>

                  <div class="verification-area">
                    <div class="verification-area__field">
                      <RsInput
                        v-model="phoneFormData.code"
                        name="code"
                        radius="sm"
                        size="lg"
                        :placeholder="t('login.verificationCode')"
                        @keyup.enter="handlePhoneLogin"
                      >
                        <template #prefix>
                          <GIcon>
                            <KeyOutline />
                          </GIcon>
                        </template>
                      </RsInput>
                    </div>
                    <RsButton
                      class="verification-btn"
                      radius="sm"
                      size="lg"
                      :disabled="codeSending"
                      @click="sendVerificationCode"
                    >
                      {{ codeSending ? `${countdown}s` : t('login.sendCode') }}
                    </RsButton>
                  </div>

                  <RsButton
                    variant="primary"
                    size="lg"
                    radius="sm"
                    class="login-btn"
                    :loading="phoneLoading"
                    @click="handlePhoneLogin"
                  >
                    {{ t('login.loginButton') }}
                  </RsButton>
                </RsForm>
              </template>
            </RsTabs>

            <div class="form-footer">
              <p class="copyright">
                © {{ new Date().getFullYear() }} {{ tCommon('common.companyName') }}.
                {{ tCommon('common.copyright') }}
                {{ tCommon('common.version') }}: {{ appVersion }}
              </p>
            </div>
          </RsCard>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import LanguageSwitcher from '@/components/common/LanguageSwitcher.vue'
import GIcon from '@/components/gicon/GIcon.vue'
import { preloadModules } from '@/locales'
import {
  RsAlert,
  RsButton,
  RsCard,
  RsForm,
  RsInput,
  RsLoading,
  RsTabs,
} from '@/ui'
import {
  DocumentTextOutline,
  KeyOutline,
  LockClosedOutline,
  PeopleOutline,
  PersonOutline,
  PhonePortraitOutline,
  ShieldCheckmarkOutline,
  ShieldOutline,
} from '@vicons/ionicons5'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useLoginAuth } from './hooks/useLoginAuth'

const activeTab = ref('account')
const route = useRoute()
const router = useRouter()
const showForgotPassword = computed(() => route.name === 'forgotPassword')

const LOGIN_BODY_CLASS = 'is-login-page'

onMounted(() => {
  document.body.classList.add(LOGIN_BODY_CLASS)
  void preloadModules(['common', 'hub0001'], ['zh-CN', 'en'])
})

onUnmounted(() => {
  document.body.classList.remove(LOGIN_BODY_CLASS)
})

const {
  formRef,
  phoneFormRef,
  formData,
  phoneFormData,
  accountRules,
  phoneRules,
  loading,
  phoneLoading,
  lockRemainSeconds,
  captchaUrl,
  codeSending,
  countdown,
  appVersion,
  handleLogin,
  handlePhoneLogin,
  sendVerificationCode,
  refreshCaptcha,
  goToForgotPassword,
  t,
  tCommon,
} = useLoginAuth()

const loginTabItems = computed(() => [
  { value: 'account', label: t('login.accountLogin') },
  { value: 'phone', label: t('login.phoneLogin') },
])

const backToLogin = () => {
  router.push({ name: 'login' })
}
</script>

<style lang="scss" scoped>
.login-view {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  flex-direction: column;
  width: 100%;
  min-height: 100vh;
  min-height: 100dvh;
  isolation: isolate;
  overflow: hidden;
  overscroll-behavior: none;
  color-scheme: dark;
  background:
    radial-gradient(90% 60% at 82% 12%, rgba(129, 140, 248, 0.32) 0%, transparent 58%),
    radial-gradient(70% 50% at 12% 86%, rgba(56, 189, 248, 0.14) 0%, transparent 55%),
    linear-gradient(155deg, #070b16 0%, #10152b 42%, #1a1840 78%, #241f4f 100%);
  background-color: #0c1020;

  &::before {
    content: '';
    position: absolute;
    inset: 0;
    z-index: 0;
    opacity: 0.45;
    background-image:
      linear-gradient(rgba(255, 255, 255, 0.028) 1px, transparent 1px),
      linear-gradient(90deg, rgba(255, 255, 255, 0.028) 1px, transparent 1px);
    background-size: 40px 40px;
    mask-image: radial-gradient(ellipse 80% 70% at 50% 45%, #000 20%, transparent 85%);
    -webkit-mask-image: radial-gradient(ellipse 80% 70% at 50% 45%, #000 20%, transparent 85%);
    pointer-events: none;
  }

  &::after {
    content: '';
    position: absolute;
    inset: -40%;
    z-index: 0;
    background: conic-gradient(
      from 210deg at 50% 50%,
      transparent 0deg,
      rgba(129, 140, 248, 0.08) 90deg,
      transparent 200deg
    );
    animation: login-brand-glow 28s linear infinite;
    pointer-events: none;
  }
}

.global-brand-lockup {
  position: absolute;
  top: 20px;
  left: 20px;
  z-index: 30;
  display: flex;
  align-items: center;
  gap: 12px;
  max-width: calc(100vw - 140px);
  pointer-events: none;

  &__logo {
    height: 40px;
    width: auto;
    object-fit: contain;
    flex-shrink: 0;
    filter: drop-shadow(0 2px 8px rgba(0, 0, 0, 0.25));
  }

  &__text {
    font-size: 1.0625rem;
    font-weight: 600;
    letter-spacing: 0.02em;
    color: rgba(255, 255, 255, 0.92);
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
    line-height: 1.2;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

.global-language-switch {
  position: absolute;
  top: 20px;
  right: 20px;
  z-index: 30;
}

.login-split {
  position: relative;
  z-index: 1;
  flex: 1;
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  align-items: stretch;
  width: 100%;
  min-width: 0;
  min-height: 0;
  padding-top: 56px;
  box-sizing: border-box;
  overflow-x: auto;
  overflow-y: hidden;
  overscroll-behavior: contain;
  scrollbar-gutter: stable;
}

.login-split__brand,
.login-split__form {
  flex: 1 1 50%;
  display: flex;
  flex-direction: column;
  min-width: min(320px, 100%);
  min-height: 0;
  max-width: 50%;
  box-sizing: border-box;
}

.welcome-area {
  position: relative;
  z-index: 1;
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
  padding: 12px 20px 20px;
  color: #f0f4ff;
  overflow: hidden;
  background: transparent;
}

.welcome-area__body {
  position: relative;
  z-index: 2;
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: safe center;
  align-items: center;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  text-align: center;
  max-width: min(420px, 100%);
  width: 100%;
  margin: 0 auto;
  padding: 8px 0 16px;
  box-sizing: border-box;
  scrollbar-width: thin;
  scrollbar-color: rgba(148, 163, 184, 0.5) transparent;

  &::-webkit-scrollbar {
    width: 6px;
  }

  &::-webkit-scrollbar-thumb {
    background: rgba(148, 163, 184, 0.45);
    border-radius: 6px;
  }
}

.welcome-area__title {
  font-size: clamp(1.5rem, 2.5vmin + 0.35rem, 2.125rem);
  line-height: 1.2;
  margin: 0 0 10px;
  font-weight: 700;
  letter-spacing: -0.02em;
  background: linear-gradient(120deg, #fff 0%, #c7d2fe 45%, #a5b4fc 100%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.welcome-area__subtitle {
  font-size: clamp(0.875rem, 1.2vmin + 0.5rem, 1.0625rem);
  margin: 0 0 20px;
  opacity: 0.88;
  line-height: 1.55;
  padding: 0 4px;
  max-width: 36rem;
}

.features {
  display: flex;
  flex-direction: column;
  gap: 12px;
  text-align: left;
  width: 100%;
}

.feature-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 16px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.12);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.2s ease;

  &:hover {
    border-color: rgba(165, 180, 252, 0.45);
    box-shadow: 0 12px 40px rgba(79, 70, 229, 0.15);
    transform: translateY(-1px);
  }
}

.feature-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: linear-gradient(145deg, rgba(99, 102, 241, 0.35), rgba(56, 189, 248, 0.2));
  border: 1px solid rgba(255, 255, 255, 0.15);
  font-size: 22px;
  color: #e0e7ff;
}

.feature-text {
  min-width: 0;

  h3 {
    font-size: 1rem;
    margin: 0 0 6px;
    font-weight: 600;
    color: #fff;
  }

  p {
    margin: 0;
    opacity: 0.88;
    line-height: 1.55;
    font-size: 0.8125rem;
  }
}

.form-area {
  position: relative;
  z-index: 1;
  display: flex;
  flex: 1;
  flex-direction: column;
  align-items: center;
  justify-content: safe center;
  min-height: 0;
  padding: 12px 20px 20px;
  background: transparent;
  overflow-x: hidden;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  box-sizing: border-box;
  scrollbar-width: thin;
  scrollbar-color: rgba(148, 163, 184, 0.5) transparent;

  &::-webkit-scrollbar {
    width: 6px;
  }

  &::-webkit-scrollbar-thumb {
    background: rgba(148, 163, 184, 0.45);
    border-radius: 6px;
  }
}

/* 跟随全局主题 token；仅关掉厚焦点环 */
.login-card {
  width: 100%;
  max-width: 420px;
  --rs-focus-ring: transparent;
  --rs-focus-ring-width: 0px;
  --rs-input-shadow: none;
}

.login-tabs {
  width: 100%;
}

.login-form {
  width: 100%;
}

.forgot-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
}

.forgot-panel__title {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 650;
  color: var(--rs-text);
}

.forgot-panel__subtitle {
  margin: 0;
  font-size: 0.9375rem;
  line-height: 1.55;
  color: var(--rs-muted);
}

.forgot-panel__hint {
  margin: 0;
}

.captcha-area,
.verification-area {
  display: flex;
  /* 错误文案在输入框下方增高时，验证码图/发送按钮仍与控件顶对齐，避免垂直错位 */
  align-items: flex-start;
  gap: 10px;
  width: 100%;
}

.captcha-area__field,
.verification-area__field {
  flex: 1 1 auto;
  min-width: 0;
}

.captcha-img {
  width: 118px;
  /* 与旁侧 RsInput size=lg 同高，避免验证码偏高/偏低 */
  height: var(--rs-control-height-lg);
  min-height: var(--rs-control-height-lg);
  box-sizing: border-box;
  padding: 0;
  border-radius: 10px;
  overflow: hidden;
  cursor: pointer;
  background: var(--rs-surface-hover);
  border: 1px solid var(--rs-border);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;

  img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.captcha-loading {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.verification-btn {
  min-width: 118px;
  height: var(--rs-control-height-lg);
  min-height: var(--rs-control-height-lg);
  flex-shrink: 0;
}

.form-options {
  display: flex;
  justify-content: flex-end;
}

.login-cooldown {
  margin: 0;
  font-size: 13px;
  color: var(--rs-muted);
  text-align: center;
}

.login-btn {
  width: 100%;
}

.form-footer {
  margin-top: 16px;
  text-align: center;
  font-size: 12px;
  font-weight: 500;
  color: var(--rs-muted);
  line-height: 1.5;
  letter-spacing: 0.01em;

  .copyright {
    margin: 0;
  }
}

@keyframes login-brand-glow {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .login-view::after {
    animation: none;
  }

  .feature-item {
    transition: none;
  }
}
</style>

<style lang="scss">
body.is-login-page,
body.is-login-page #app,
body.is-login-page #app-container {
  background-color: #0c1020 !important;
}

html:has(body.is-login-page),
body.is-login-page {
  overflow: hidden !important;
  height: 100% !important;
  overscroll-behavior: none;
}
</style>
