<template>
    <div class="login-view">
        <!-- 全局语言切换器 -->
        <div class="global-language-switch">
            <LanguageSwitcher />
        </div>

        <GPane direction="horizontal" :no-resize="true"  defaultSize="50%" class="login-pane">
            <template #1>
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
            </template>

            <template #2>
                <div class="form-area">
                    <div class="form-container">
                        <div class="form-header">
                            <h2>{{ t('login.loginTitle') || t('login.title') }}</h2>
                            <p>{{ t('login.loginSubtitle') || t('login.subtitle') }}</p>
                        </div>

                        <!-- 登录类型选择tabs -->
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
            </template>
        </GPane>
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
import { GPane } from '@/components/gpane'
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
    width: 100%;
    height: 100vh;
    background-color: #f7f9fc;
    position: relative;
    overflow: hidden;

    .global-language-switch {
        position: absolute;
        top: 20px;
        right: 20px;
        z-index: 20;
    }

    :deep(.login-pane) {
        width: 100%;
        height: 100%;

        .g-pane__flex-container--horizontal {
            .g-pane__flex-pane--1,
            .g-pane__flex-pane--2 {
                flex: 1 1 0;
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
        height: 100%;
        width: 100%;
        background: linear-gradient(135deg, #2b5876, #4e4376);
        color: white;

        .logo-area {
            position: absolute;
            top: 30px;
            left: 50%;
            transform: translateX(-50%);
            display: flex;
            align-items: center;
            gap: 10px;
            z-index: 10;

            img {
                height: 40px;
            }

            .logo-text {
                font-size: 1.5rem;
                font-weight: 600;
                margin: 0;
            }
        }

        .welcome-content {
            max-width: 600px;
            width: 100%;
            margin-top: 40px;
            text-align: center;

            .title {
                font-size: 2.8rem;
                margin-bottom: 16px;
                font-weight: 600;
            }

            .subtitle {
                font-size: 1.3rem;
                margin-bottom: 40px;
                opacity: 0.95;
            }

            .features {
                display: grid;
                grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
                gap: 30px;

                .feature-item {
                    display: flex;
                    align-items: flex-start;
                    background: rgba(255, 255, 255, 0.05);
                    border-radius: 16px;
                    padding: 20px;

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
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 40px 30px;
        background-color: white;
        height: 100%;
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
                    border-radius: 8px;
                    overflow: hidden;
                    cursor: pointer;
                    background-color: #f5f7fa;
                    border: 1px solid #e2e8f0;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    flex-shrink: 0;

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

            .form-footer {
                margin-top: 30px;
                text-align: center;
                font-size: 13px;
                color: #a0aec0;
            }
        }
    }
}

</style>