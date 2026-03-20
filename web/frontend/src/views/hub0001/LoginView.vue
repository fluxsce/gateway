<template>
    <div class="login-view">
        <!-- 与语言切换器对称：固定视口角标 -->
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
                    <!-- 登录卡片为浅色磨砂底：与全局暗色主题解耦，避免 Naive 暗色 token（浅字）叠在浅底上不可读 -->
                    <div class="form-container">
                        <n-config-provider :theme="null" :theme-overrides="lightThemeOverrides">
                        <div class="form-header">
                            <h2>{{ t('login.loginTitle') || t('login.title') }}</h2>
                            <p>{{ t('login.loginSubtitle') || t('login.subtitle') }}</p>
                        </div>

                        <!-- 登录类型选择tabs -->
                        <n-tabs v-model:value="activeTab" :animated="false" justify-content="space-evenly" size="small"
                            class="login-tabs" type="line">
                            <n-tab-pane name="account" :tab="t('login.accountLogin')">
                                <!-- 懒加载：只渲染当前激活的表单 -->
                                <n-form v-if="activeTab === 'account'" ref="formRef" :model="formData" :rules="rules"
                                    size="large" class="login-form">
                                    <n-form-item path="userId">
                                        <n-input v-model:value="formData.userId" :placeholder="t('login.userId')"
                                            @keyup.enter="handleLogin">
                                            <template #prefix>
                                                <n-icon>
                                                    <PersonOutline />
                                                </n-icon>
                                            </template>
                                        </n-input>
                                    </n-form-item>

                                    <n-form-item path="password">
                                        <n-input v-model:value="formData.password" type="password" show-password-on="mousedown"
                                            :placeholder="t('login.password')" @keyup.enter="handleLogin">
                                            <template #prefix>
                                                <n-icon>
                                                    <LockClosedOutline />
                                                </n-icon>
                                            </template>
                                        </n-input>
                                    </n-form-item>

                                    <n-form-item path="captcha">
                                        <div class="captcha-area">
                                            <n-input v-model:value="formData.captchaCode" :placeholder="t('login.captcha')"
                                                @keyup.enter="handleLogin">
                                                <template #prefix>
                                                    <n-icon>
                                                        <ShieldOutline />
                                                    </n-icon>
                                                </template>
                                            </n-input>
                                            <div class="captcha-img" @click="refreshCaptcha">
                                                <img v-if="captchaUrl" :src="captchaUrl" alt="Captcha" />
                                                <div v-else class="captcha-loading">
                                                    <n-spin size="small" />
                                                </div>
                                            </div>
                                        </div>
                                    </n-form-item>

                                    <div class="form-options">
                                        <n-button text @click="goToForgotPassword" class="forgot-btn">
                                            {{ t('login.forgotPassword') }}
                                        </n-button>
                                    </div>

                                    <n-button type="primary" size="large" block :loading="loading" @click="handleLogin"
                                        class="login-btn">
                                        {{ t('login.loginButton') }}
                                    </n-button>
                                </n-form>
                            </n-tab-pane>

                            <!-- 手机验证码登录 -->
                            <n-tab-pane name="phone" :tab="t('login.phoneLogin')">
                                <!-- 懒加载：只渲染当前激活的表单 -->
                                <n-form v-if="activeTab === 'phone'" ref="phoneFormRef" :model="phoneFormData"
                                    :rules="phoneRules" size="large" class="login-form">
                                    <n-form-item path="phone">
                                        <n-input v-model:value="phoneFormData.phone" :placeholder="t('login.phoneNumber')"
                                            @keyup.enter="handlePhoneLogin">
                                            <template #prefix>
                                                <n-icon>
                                                    <PhonePortraitOutline />
                                                </n-icon>
                                            </template>
                                        </n-input>
                                    </n-form-item>

                                    <n-form-item path="code">
                                        <div class="verification-area">
                                            <n-input v-model:value="phoneFormData.code"
                                                :placeholder="t('login.verificationCode')" @keyup.enter="handlePhoneLogin">
                                                <template #prefix>
                                                    <n-icon>
                                                        <KeyOutline />
                                                    </n-icon>
                                                </template>
                                            </n-input>
                                            <n-button class="verification-btn" :disabled="codeSending"
                                                @click="sendVerificationCode">
                                                {{ codeSending ? `${countdown}s` : t('login.sendCode') }}
                                            </n-button>
                                        </div>
                                    </n-form-item>

                                    <n-button type="primary" size="large" block :loading="phoneLoading"
                                        @click="handlePhoneLogin" class="login-btn">
                                        {{ t('login.loginButton') }}
                                    </n-button>
                                </n-form>
                            </n-tab-pane>
                        </n-tabs>

                        <div class="form-footer">
                            <p class="copyright">
                                © {{ new Date().getFullYear() }} {{ tCommon('common.companyName') }}.
                                {{ tCommon('common.copyright') }}
                                {{ tCommon('common.version') }}: {{ appVersion }}
                            </p>
                        </div>
                        </n-config-provider>
                    </div>
                </div>
            </main>
        </div>
    </div>
</template>

<script setup lang="ts">
// Icons import
import {
    DocumentTextOutline,
    KeyOutline,
    LockClosedOutline,
    PeopleOutline,
    PersonOutline,
    PhonePortraitOutline,
    ShieldCheckmarkOutline,
    ShieldOutline
} from '@vicons/ionicons5'
import { lightThemeOverrides } from '@/config/theme'
import { NConfigProvider, NIcon } from 'naive-ui'
import { onMounted, onUnmounted, ref } from 'vue'

// Import the useLoginAuth hook which now contains all login logic
import LanguageSwitcher from '@/components/common/LanguageSwitcher.vue'
import { useLoginAuth } from './hooks/useLoginAuth'

const activeTab = ref('account')

const LOGIN_BODY_CLASS = 'is-login-page'

onMounted(() => {
    document.body.classList.add(LOGIN_BODY_CLASS)
})

onUnmounted(() => {
    document.body.classList.remove(LOGIN_BODY_CLASS)
})

// Use the hook to get all login-related functionality
const {
    formRef,
    phoneFormRef,
    formData,
    phoneFormData,
    rules,
    phoneRules,
    loading,
    phoneLoading,
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
    tCommon
} = useLoginAuth()
</script>

<style lang="scss" scoped>
.login-view {
    /* 脱离文档流，避免把 #app / body 撑高而出现浏览器级右侧滚动条 */
    position: fixed;
    inset: 0;
    z-index: 50;
    display: flex;
    flex-direction: column;
    width: 100%;
    /* inset:0 已铺满视口；dvh 兜底部分移动浏览器地址栏伸缩 */
    min-height: 100vh;
    min-height: 100dvh;
    isolation: isolate;
    overflow: hidden;
    overscroll-behavior: none;
    color-scheme: dark;

    /* ========== 全屏科技风底图（左右两栏共用，避免仅半屏渐变） ========== */
    background:
        radial-gradient(100% 70% at 85% 15%, rgba(120, 80, 255, 0.38) 0%, transparent 55%),
        radial-gradient(80% 55% at 8% 88%, rgba(0, 200, 255, 0.18) 0%, transparent 52%),
        linear-gradient(152deg, #0a0f1c 0%, #121638 38%, #1e1b4b 72%, #312e81 100%);
    /* 与渐变末端一致，避免层与层之间露缝 */
    background-color: #121638;

    /*
     * 全屏氛围层 + 网格：原先仅右侧 .form-area 有 slate 渐变，logo/语言栏在 absolute 层，
     * 未盖住顶区，会像「另一截背景」。此处与 form-area 同款渐变铺满整视口，再叠网格。
     */
    &::before {
        content: '';
        position: absolute;
        inset: 0;
        z-index: 0;
        background-image:
            linear-gradient(rgba(255, 255, 255, 0.04) 1px, transparent 1px),
            linear-gradient(90deg, rgba(255, 255, 255, 0.04) 1px, transparent 1px),
            linear-gradient(105deg, rgba(15, 23, 42, 0.2) 0%, rgba(15, 23, 42, 0.45) 100%);
        background-size: 28px 28px, 28px 28px, auto;
        background-repeat: repeat, repeat, no-repeat;
        background-position: 0 0, 0 0, center;
        pointer-events: none;
    }

    &::after {
        content: '';
        position: absolute;
        inset: -45%;
        z-index: 0;
        background: conic-gradient(from 200deg at 50% 50%, transparent 0deg, rgba(99, 102, 241, 0.1) 100deg, transparent 220deg);
        animation: login-brand-glow 22s linear infinite;
        pointer-events: none;
    }

    /* 左上角：固定 px/rem，窗口拉伸时不跟 vw 一起变大变小（浏览器缩放仍会按比例缩放，符合预期） */
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
        /* 不在整栏上纵向滚动，避免视口级/靠左误导滚动条；仅左右栏内部滚动 */
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
        /* 低于约 640px 宽时整行横向滚动，保持左右栏而非上下堆叠 */
        min-width: min(320px, 100%);
        min-height: 0;
        max-width: 50%;
        box-sizing: border-box;
    }

    /* ========== 左侧品牌区（透底，叠在全屏背景上） ========== */
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

    /* 与右侧登录区同为「主内容垂直居中」，避免左轻右重 */
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
        transition: border-color 0.2s ease, box-shadow 0.2s ease, transform 0.2s ease;

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

    /* ========== 右侧表单：仅在栏内滚动，不撑开整页 ========== */
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
        /* 氛围渐变已铺在 .login-view::before 全屏，避免右侧与顶区重复叠色 */
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

    .form-container {
        flex-shrink: 0;
        width: 100%;
        max-width: 420px;
        margin: 0 auto;
        padding: 22px 24px 24px;
        border-radius: 18px;
        color-scheme: light;
        background: rgba(255, 255, 255, 0.9);
        border: 1px solid rgba(255, 255, 255, 0.55);
        backdrop-filter: blur(22px) saturate(1.35);
        -webkit-backdrop-filter: blur(22px) saturate(1.35);
        box-shadow:
            0 8px 40px rgba(0, 0, 0, 0.22),
            0 0 0 1px rgba(255, 255, 255, 0.35) inset;
    }

    .form-header {
        text-align: center;
        margin-bottom: clamp(20px, 3vw, 28px);

        h2 {
            font-size: clamp(1.25rem, 2.5vw, 1.65rem);
            margin: 0 0 8px;
            color: #0f172a;
            font-weight: 700;
            letter-spacing: -0.02em;
        }

        p {
            margin: 0;
            color: #475569;
            font-size: clamp(0.875rem, 1.4vw, 1rem);
            line-height: 1.5;
        }
    }

    :deep(.login-tabs.n-tabs) {
        .n-tabs-nav {
            margin-bottom: 8px;
        }

        .n-tabs-bar {
            height: 3px;
            border-radius: 3px;
            background: linear-gradient(90deg, #6366f1, #8b5cf6);
        }

        .n-tabs-tab {
            font-weight: 500;
        }

        .n-tabs-tab--active {
            color: #0f172a;
        }
    }

    :deep(.login-form .n-input) {
        border-radius: 10px;
        transition: box-shadow 0.2s ease;

        &:hover {
            box-shadow: 0 0 0 1px rgba(99, 102, 241, 0.2);
        }
    }

    :deep(.login-form .n-input--focus) {
        box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.25);
    }

    .captcha-area,
    .verification-area {
        display: flex;
        gap: 10px;
        width: 100%;

        :deep(.n-input) {
            flex: 1;
            min-width: 0;
        }

        .captcha-img {
            width: 120px;
            height: 42px;
            border-radius: 10px;
            overflow: hidden;
            cursor: pointer;
            /* 固定浅色，避免全局 data-theme=dark 时 --g-* 与浅色卡片冲突 */
            background-color: #f1f5f9;
            border: 1px solid #e2e8f0;
            display: flex;
            align-items: center;
            justify-content: center;
            flex-shrink: 0;
            transition: box-shadow 0.2s ease, border-color 0.2s ease;

            &:hover {
                border-color: rgba(99, 102, 241, 0.4);
                box-shadow: 0 2px 12px rgba(99, 102, 241, 0.12);
            }

            img {
                width: 100%;
                height: 100%;
                object-fit: cover;
            }
        }

        .verification-btn {
            min-width: 120px;
            height: 42px;
            flex-shrink: 0;
        }
    }

    .form-options {
        display: flex;
        justify-content: flex-end;
        margin: 12px 0;
    }

    :deep(.login-btn.n-button) {
        margin-top: 8px;
        height: 44px;
        border-radius: 10px;
        font-weight: 600;
        box-shadow: 0 4px 14px rgba(99, 102, 241, 0.22);
        transition: transform 0.15s ease, box-shadow 0.15s ease;

        &:hover {
            box-shadow: 0 6px 20px rgba(99, 102, 241, 0.32);
        }

        &:active {
            transform: scale(0.98);
        }
    }

    .form-footer {
        margin-top: 12px;
        text-align: center;
        font-size: 11px;
        color: #94a3b8;
        line-height: 1.4;
    }

    .form-footer .copyright {
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

    .login-view .feature-item {
        transition: none;
    }

    .login-view :deep(.login-btn.n-button):active {
        transform: none;
    }
}
</style>

<style lang="scss">
/* 登录页 fixed 层之下若透出 #app/body 主题色，顶栏会与主背景不一致 */
body.is-login-page,
body.is-login-page #app,
body.is-login-page #app-container {
    background-color: #121638 !important;
}

/* 禁止登录路由下 body 出现整页滚动条（缩放/内容过高时仅在 .login-view 内分栏滚动） */
html:has(body.is-login-page),
body.is-login-page {
    overflow: hidden !important;
    height: 100% !important;
    overscroll-behavior: none;
}
</style>