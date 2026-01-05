<template>
    <div class="login-view">
        <!-- 全局语言切换器 -->
        <div class="global-language-switch">
            <LanguageSwitcher />
        </div>

        <div class="welcome-area">
            <div class="logo-area">
                <img src="@/assets/images/logo.png" alt="Logo" />
                <h3 class="logo-text">FLUX Datahub Gateway</h3>
            </div>

            <div class="welcome-content">
                <h1 class="title">{{ t('login.welcomeTitle') }}</h1>
                <p class="subtitle">{{ t('login.welcomeSubtitle') }}</p>

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

        <div class="form-area">
            <div class="form-container">
                <div class="form-header">
                    <h2>{{ t('login.loginTitle') || t('login.title') }}</h2>
                    <p>{{ t('login.loginSubtitle') || t('login.subtitle') }}</p>
                </div>

                <!-- 登录类型选择tabs - 优化INP性能 -->
                <n-tabs v-model:value="activeTab" :animated="false" justify-content="space-evenly" size="small"
                    class="login-tabs" type="line" @update:value="handleTabChange">
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
            </div>
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
import { NIcon } from 'naive-ui'
import { nextTick, ref } from 'vue'

// Import the useLoginAuth hook which now contains all login logic
import LanguageSwitcher from '@/components/common/LanguageSwitcher.vue'
import { useLoginAuth } from './hooks/useLoginAuth'

// Tab切换状态管理 - 优化INP性能
const activeTab = ref('account')

// 防抖计时器
let tabChangeTimer: ReturnType<typeof setTimeout> | null = null

// 优化的Tab切换处理函数
const handleTabChange = (value: string) => {
    // 防抖处理 - 避免快速切换导致的性能问题
    if (tabChangeTimer) {
        clearTimeout(tabChangeTimer)
    }

    // 使用requestAnimationFrame确保在下一帧执行
    requestAnimationFrame(() => {
        tabChangeTimer = setTimeout(() => {
            activeTab.value = value
            tabChangeTimer = null

            // 异步更新，避免阻塞主线程
            nextTick(() => {
                // Tab切换完成后的处理
                if (value === 'account') {
                    // 可以在这里做一些账号登录相关的初始化
                } else if (value === 'phone') {
                    // 可以在这里做一些手机登录相关的初始化
                }
            })
        }, 50) // 50ms防抖
    })
}

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
    display: flex;
    min-height: 100vh;
    background-color: #f7f9fc;
    font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    position: relative;
    contain: layout paint;
    will-change: auto;
    font-display: swap;

    // 全局语言切换器
    .global-language-switch {
        position: absolute;
        top: 20px;
        right: 20px;
        z-index: 20;

        // 暗色背景上的样式
        @media (max-width: 1199px) {
            :deep(.language-selector) {
                color: white;

                &:hover {
                    background-color: rgba(255, 255, 255, 0.1);
                }
            }
        }
    }

    .welcome-area {
        position: relative;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        padding: 50px 40px;
        width: 50%;
        background: linear-gradient(135deg, #2b5876, #4e4376);
        color: white;
        overflow: hidden;
        contain: layout paint;
        will-change: auto;

        .logo-area {
            position: absolute;
            top: 30px;
            left: 50%;
            transform: translateX(-50%);
            display: flex;
            align-items: center;
            z-index: 10;

            img {
                height: 40px;
                filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.2));
                margin-right: 10px;
            }

            .logo-text {
                font-size: 1.5rem;
                font-weight: 600;
                margin: 0;
                text-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
            }
        }

        .welcome-content {
            position: relative;
            z-index: 5;
            max-width: 600px;
            width: 100%;
            margin-top: 40px;

            .title {
                font-size: 2.8rem;
                margin-bottom: 16px;
                font-weight: 600;
                letter-spacing: -0.5px;
                color: #ffffff;
                text-shadow: 0 2px 10px rgba(0, 0, 0, 0.2);
                will-change: transform;
                contain: layout;
                font-display: swap;
                transform: translateZ(0);
            }

            .subtitle {
                font-size: 1.3rem;
                margin-bottom: 40px;
                opacity: 0.95;
                line-height: 1.5;
                text-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
            }

            .features {
                display: grid;
                grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
                gap: 30px;
                width: 100%;

                .feature-item {
                    display: flex;
                    align-items: flex-start;
                    margin-bottom: 20px;
                    background: rgba(255, 255, 255, 0.05);
                    border-radius: 16px;
                    padding: 20px;
                    border: 1px solid rgba(255, 255, 255, 0.1);
                    contain: layout paint;

                    .feature-icon {
                        display: flex;
                        align-items: center;
                        justify-content: center;
                        min-width: 50px;
                        height: 50px;
                        background: rgba(255, 255, 255, 0.15);
                        border-radius: 14px;
                        margin-right: 15px;
                        font-size: 24px;
                    }

                    .feature-text {
                        h3 {
                            font-size: 1.2rem;
                            margin-bottom: 8px;
                            font-weight: 600;
                        }

                        p {
                            opacity: 0.9;
                            line-height: 1.5;
                            font-size: 0.95rem;
                        }
                    }
                }
            }
        }
    }

    .form-area {
        width: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 40px 30px;
        background-color: white;
        box-shadow: 0 5px 30px rgba(0, 0, 0, 0.05);
        overflow-y: auto;

        .form-container {
            width: 100%;
            max-width: 420px;

            .form-header {
                text-align: center;
                margin-bottom: 30px;

                h2 {
                    font-size: 1.8rem;
                    margin-bottom: 8px;
                    color: #1a202c;
                    font-weight: 700;
                }

                p {
                    color: #718096;
                    font-size: 1rem;
                }
            }

            .login-form {
                max-height: calc(100vh - 240px);
                overflow-y: auto;
            }

            // Tab性能优化样式
            :deep(.login-tabs) {
                contain: layout;
                will-change: auto;

                .n-tabs-nav {
                    contain: layout paint;
                    transform: translateZ(0);
                }

                .n-tabs-tab {
                    contain: layout;
                    transition: none !important; // 移除transition减少重绘

                    &__label {
                        will-change: auto;
                        contain: layout;
                        user-select: none;
                        pointer-events: auto;
                        transform: translateZ(0);
                    }

                    // 优化激活状态切换
                    &--active {
                        .n-tabs-tab__label {
                            font-weight: 600;
                        }
                    }
                }

                .n-tabs-bar {
                    contain: layout;
                    transform: translateZ(0);
                    transition: transform 0.15s ease-out !important; // 简化transition
                }

                .n-tabs-pane-wrapper {
                    contain: layout;
                    will-change: auto;
                }
            }

            :deep(.n-form) {
                display: flex;
                flex-direction: column;
                width: 100%;
                contain: layout;
            }

            :deep(.n-form-item) {
                width: 100%;
                margin-bottom: 16px;
                contain: layout;
            }

            :deep(.n-input) {
                --n-border: #e2e8f0;
                --n-border-hover: #cbd5e0;
                --n-border-focus: var(--primary-color, #4e4376);
                --n-color-focus: #f7f9fc;
                --n-border-radius: 8px;
                --n-height: 42px;
                --n-icon-size: 18px;
                --n-font-size: 15px;
                contain: layout;
                will-change: auto;
            }

            // 进一步优化输入框性能
            :deep(.n-input__input-el) {
                font-size: 15px;
                will-change: auto;
            }

            :deep(.n-input__prefix) {
                margin-right: 10px;
                color: #a0aec0;
                contain: layout;
            }

            :deep(.n-input__placeholder) {
                font-size: 15px;
                color: #a0aec0;
            }

            :deep(.n-icon) {
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 18px;
                contain: layout;
            }

            .login-btn {
                --n-color: var(--primary-color, #4e4376);
                --n-color-hover: var(--primary-color-hover, #5a4e8c);
                --n-color-pressed: var(--primary-color-active, #423866);
                --n-border: var(--primary-color, #4e4376);
                --n-border-hover: var(--primary-color-hover, #5a4e8c);
                --n-border-pressed: var(--primary-color-active, #423866);
                --n-border-focus: var(--primary-color, #4e4376);
                --n-ripple-color: rgba(78, 67, 118, 0.2);
                --n-text-color: #fff;
                --n-height: 44px;
                --n-font-size: 16px;
                margin-top: 8px;
                border-radius: 8px;
                font-weight: 500;
                contain: layout;
                will-change: auto;

                // 优化按钮点击性能
                :deep(.n-button__content) {
                    contain: layout;
                }
            }

            .captcha-area,
            .verification-area {
                display: flex;
                gap: 10px;
                width: 100%;

                .n-input {
                    flex: 1;
                }

                .captcha-img {
                    width: 120px;
                    height: 42px;
                    border-radius: 8px;
                    overflow: hidden;
                    cursor: pointer;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    background-color: #f5f7fa;
                    border: 1px solid #e2e8f0;
                    contain: layout paint;

                    img {
                        width: 100%;
                        height: 100%;
                        object-fit: cover;
                    }

                    .captcha-loading {
                        width: 100%;
                        height: 100%;
                        display: flex;
                        align-items: center;
                        justify-content: center;
                    }
                }

                .verification-btn {
                    min-width: 120px;
                    font-size: 14px;
                    height: 42px;
                    border-radius: 8px;
                    background-color: #f8f9fa;
                    color: var(--primary-color, #4e4376);
                    border: 1px solid #e2e8f0;

                    &:hover:not(:disabled) {
                        background-color: #f1f1f1;
                        border-color: #cbd5e0;
                    }

                    &:disabled {
                        background-color: #f8f9fa;
                        color: #a0aec0;
                    }
                }
            }

            .form-options {
                display: flex;
                align-items: center;
                justify-content: flex-end;
                margin: 12px 0;
                width: 100%;

                .forgot-btn {
                    color: var(--primary-color, #4e4376);
                    font-weight: 500;

                    &:hover {
                        color: var(--primary-color-hover, #5a4e8c);
                    }
                }
            }

            .form-footer {
                margin-top: 30px;
                text-align: center;
                font-size: 13px;
                color: #a0aec0;

                .copyright {
                    margin-top: 8px;
                }
            }
        }
    }
}

// 响应式设计
@media (max-width: 1199px) {
    .login-view {
        flex-direction: column;

        .global-language-switch {
            top: 20px;
            right: 20px;
            z-index: 20;

            :deep(.language-selector) {
                color: white;

                &:hover {
                    background-color: rgba(255, 255, 255, 0.1);
                }
            }
        }

        .welcome-area {
            padding: 40px 30px;
            min-height: 40vh;
            width: 100%;
            transform: translateZ(0);

            .welcome-content {
                text-align: center;
                margin: 30px auto 0;

                .features {
                    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
                    margin: 0 auto;
                    max-width: 700px;

                    .feature-item {
                        justify-content: flex-start;
                        text-align: left;
                    }
                }
            }
        }

        .form-area {
            width: 100%;
            max-width: 100%;
            padding: 30px 20px;
            max-height: 60vh;
            overflow-y: auto;

            .form-container {
                max-width: 420px;
                margin: 0 auto;
            }
        }
    }
}

@media (max-width: 767px) {
    .login-view {
        .global-language-switch {
            top: 15px;
            right: 15px;
        }

        .welcome-area {
            padding: 30px 15px;
            max-height: 40vh;
            overflow-y: auto;

            .logo-area {
                top: 15px;

                img {
                    height: 32px;
                }

                .logo-text {
                    font-size: 1.2rem;
                }
            }

            .welcome-content {
                .title {
                    font-size: 2rem;
                }

                .subtitle {
                    font-size: 1rem;
                    margin-bottom: 25px;
                }

                .features {
                    gap: 15px;

                    .feature-item {
                        padding: 15px;
                        flex-direction: column;
                        text-align: center;

                        .feature-icon {
                            margin-right: 0;
                            margin-bottom: 10px;
                        }
                    }
                }
            }
        }

        .form-area {
            padding: 25px 15px;
            max-height: 60vh;

            .form-container {
                .form-header {
                    h2 {
                        font-size: 1.6rem;
                    }
                }

                .captcha-area,
                .verification-area {
                    flex-direction: column;
                    gap: 8px;

                    .captcha-img,
                    .verification-btn {
                        width: 100%;
                    }
                }

                .form-options {
                    justify-content: center;
                }
            }
        }
    }
}
</style>