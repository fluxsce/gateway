/**
 * 登录认证相关的业务逻辑Hook
 * 封装登录、验证码等功能，与视图层分离
 */
import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { store, type UserPermissionResponse } from '@/stores'
import type { RsFormRules, RsFormValidationResult } from '@/ui'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { logger } from '@/utils/logger'
import type { User } from '@/views/hub0002/types'
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { hub0001Api } from '../api'
import type { LoginFormData, PhoneLoginFormData } from '../types'

/** 登录表单实例（RsForm expose） */
interface LoginFormExpose {
  validate: () => Promise<RsFormValidationResult>
}

/** 登录接口返回的用户信息（扩展自 User 接口） */
interface LoginUserInfo extends Pick<User, 'userId' | 'userName' | 'realName' | 'tenantId' | 'avatar' | 'email' | 'mobile' | 'deptId' | 'tenantAdminFlag'> {
  permissions?: UserPermissionResponse
  timeout?: number
}

// 移除 parseBizData，改用 format.ts 中的 parseJsonData

/**
 * 登录认证相关的Hook
 * 专用于hub0001模块的登录认证逻辑
 *
 * @returns 登录认证相关的方法和状态
 */
export function useLoginAuth() {
  const router = useRouter()
  const message = useAppMessage()

  // 获取模块化i18n实例
  const { t } = useModuleI18n('hub0001')
  const { t: tCommon } = useModuleI18n('common')

  // 表单引用
  const formRef = ref<LoginFormExpose | null>(null)
  const phoneFormRef = ref<LoginFormExpose | null>(null)

  // 登录状态
  const loading = ref(false)
  const phoneLoading = ref(false)

  // 数字验证码相关：只保存签名票与服务端图片，答案不出前端变量
  const captchaId = ref('')
  const captchaUrl = ref('')

  // 登录冷却剩余秒数，>0 时禁止提交
  const lockRemainSeconds = ref(0)
  let lockTimer: ReturnType<typeof setInterval> | null = null

  // 手机验证码相关
  const codeSending = ref(false)
  const countdown = ref(60)

  // 版本信息
  const appVersion = ref('')

  // 登录表单数据
  const formData = reactive<LoginFormData>({
    userId: '',
    password: '',
    captchaCode: '',
    rememberMe: false,
  })

  // 手机登录表单数据
  const phoneFormData = reactive<PhoneLoginFormData>({
    phone: '',
    code: '',
    rememberMe: false,
  })

  /** 账号密码登录：RsForm 集中 rules（按字段 name 匹配） */
  const accountRules = computed<RsFormRules>(() => ({
    userId: [
      { required: true, message: t('validation.userIdRequired'), trigger: 'blur' },
      { min: 3, max: 20, message: t('validation.userIdLength'), trigger: 'blur' },
    ],
    password: [
      { required: true, message: t('validation.passwordRequired'), trigger: 'blur' },
      { min: 6, max: 32, message: t('validation.passwordLength'), trigger: 'blur' },
    ],
    captchaCode: [
      { required: true, message: t('validation.captchaRequired'), trigger: 'blur' },
      { len: 6, message: t('validation.captchaLength'), trigger: 'blur' },
    ],
  }))

  /** 手机验证码登录 rules */
  const phoneRules = computed<RsFormRules>(() => ({
    phone: [
      { required: true, message: t('validation.phoneRequired'), trigger: 'blur' },
      {
        pattern: /^1[3-9]\d{9}$/,
        message: t('validation.phoneFormat'),
        trigger: 'blur',
      },
    ],
    code: [
      { required: true, message: t('validation.codeRequired'), trigger: 'blur' },
      { len: 6, message: t('validation.codeLength'), trigger: 'blur' },
    ],
  }))

  // 防抖计时器
  let refreshTimer: ReturnType<typeof setTimeout> | null = null

  /**
   * 刷新验证码：使用服务端返回的 PNG，不再在浏览器用答案绘图。
   */
  const refreshCaptcha = async () => {
    if (refreshTimer) {
      clearTimeout(refreshTimer)
    }

    refreshTimer = setTimeout(async () => {
      try {
        const response = await hub0001Api.getCaptcha()
        if (isApiSuccess(response)) {
          const captchaData = parseJsonData<{ captchaId?: string; image?: string } | null>(
            response,
            null,
          )
          if (captchaData?.captchaId && captchaData.image) {
            captchaId.value = captchaData.captchaId
            captchaUrl.value = captchaData.image
            formData.captchaCode = ''
          } else {
            logger.warn('验证码数据解析失败', response)
          }
        } else {
          logger.warn('验证码响应无效', response)
        }
      } catch (error) {
        logger.error('刷新验证码失败:', error)
      } finally {
        refreshTimer = null
      }
    }, 300)
  }

  const stopLockCountdown = () => {
    if (lockTimer) {
      clearInterval(lockTimer)
      lockTimer = null
    }
  }

  const startLockCountdown = (seconds: number) => {
    stopLockCountdown()
    lockRemainSeconds.value = Math.max(0, Math.ceil(seconds))
    if (lockRemainSeconds.value <= 0) {
      return
    }
    lockTimer = setInterval(() => {
      if (lockRemainSeconds.value <= 1) {
        lockRemainSeconds.value = 0
        stopLockCountdown()
        return
      }
      lockRemainSeconds.value -= 1
    }, 1000)
  }

  const readRemainSeconds = (response: { extObj?: unknown }): number => {
    const ext = response.extObj
    if (ext && typeof ext === 'object' && 'remainSeconds' in ext) {
      const value = Number((ext as { remainSeconds?: unknown }).remainSeconds)
      if (Number.isFinite(value) && value > 0) {
        return value
      }
    }
    return 0
  }

  /**
   * 验证表单并登录
   */
  const handleLogin = async () => {
    if (lockRemainSeconds.value > 0) {
      message.warning(t('login.cooldownHint', { seconds: lockRemainSeconds.value }))
      return
    }
    if (!formRef.value) return

    try {
      logger.info('开始表单验证')
      // RsForm.validate 返回 { valid }，不抛错
      const result = await formRef.value.validate()
      if (!result?.valid) {
        logger.warn('表单验证失败')
        return
      }
      logger.info('表单验证通过')

      await login(formData)
    } catch (errors) {
      logger.warn('表单验证异常:', errors)
    }
  }

  /**
   * 执行登录
   * @param formData 登录表单数据
   */
  const login = async (formData: LoginFormData) => {
    loading.value = true
    logger.info('开始执行登录', { userId: formData.userId })

    try {
      // 添加验证码ID到表单数据
      const loginData = {
        ...formData,
        captchaId: captchaId.value,
      }

      // 发送登录请求
      const response = await hub0001Api.login(loginData)

      // 使用 format.ts 中的工具类处理响应
      if (!isApiSuccess(response)) {
        const remain = readRemainSeconds(response)
        if (remain > 0) {
          startLockCountdown(remain)
        }
        const errorMsg =
          remain > 0
            ? t('login.cooldownHint', { seconds: remain })
            : getApiMessage(response, t('login.loginFailed'))
        message.error(errorMsg)
        logger.warn('登录失败', {
          errMsg: response.errMsg,
          popMsg: response.popMsg,
        })
        refreshCaptcha()
        return false
      }

      // 解析登录响应数据
      const loginResult = parseJsonData<LoginUserInfo>(response, {} as LoginUserInfo)

      if (loginResult && loginResult.userId) {
        // 登录成功，更新状态
        logger.info('登录返回的用户对象:', loginResult)

        // 设置登录状态
        await store.user.setLoginState(
          loginResult.userId,
          loginResult.userName,
          loginResult.realName,
          loginResult.tenantId,
          {
            avatar: loginResult.avatar,
            email: loginResult.email,
            mobile: loginResult.mobile,
            deptId: loginResult.deptId,
            tenantAdminFlag: loginResult.tenantAdminFlag,
            timeout: loginResult.timeout,
            remember: formData.rememberMe,
          }
        )

        // 设置权限信息到用户 store
        if (loginResult.permissions) {
          store.user.setPermissions(loginResult.permissions)
          logger.info('权限信息已设置到用户 store', {
            modules: loginResult.permissions.modules?.length || 0,
            buttons: loginResult.permissions.buttons?.length || 0,
          })
        }

        logger.info('设置后的store用户信息', {
          userId: store.user.userId,
          userName: store.user.displayName,
          isAuthenticated: store.user.isAuthenticated,
        })

        const successMsg = getApiMessage(response, t('login.loginSuccess'))
        message.success(successMsg)
        logger.info('登录成功', { userId: loginResult.userId })

        // 设置全局页面标题
        store.global.setPageTitle('首页')

        // 记录登录成功，转到主界面
        router.push({ path: '/' })
        return true
      } else {
        // 登录失败 - 数据解析异常
        const errorMsg = getApiMessage(response, t('login.loginFailed'))
        message.error(errorMsg)
        logger.warn('登录失败 - 数据解析异常', {
          errMsg: response.errMsg,
          popMsg: response.popMsg,
        })
        refreshCaptcha() // 刷新验证码
        return false
      }
    } catch (error: any) {
      logger.error('登录请求异常:', error)
      message.error(error.message || t('login.networkError'))
      refreshCaptcha() // 刷新验证码
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 发送手机验证码
   */
  const sendVerificationCode = async () => {
    if (!phoneFormData.phone) {
      message.error(t('validation.phoneRequired'))
      return
    }

    if (!/^1[3-9]\d{9}$/.test(phoneFormData.phone)) {
      message.error(t('validation.phoneFormat'))
      return
    }

    codeSending.value = true
    logger.info('发送手机验证码', { phone: phoneFormData.phone })

    try {
      // 假设这里调用发送验证码API
      // const response = await hub0001Api.sendPhoneCode(phoneFormData.phone)

      // 模拟API调用
      await new Promise((resolve) => setTimeout(resolve, 500))

      message.success(t('login.codeSent'))
      logger.info('验证码发送成功')

      // 倒计时
      countdown.value = 60
      const timer = setInterval(() => {
        countdown.value--
        if (countdown.value <= 0) {
          clearInterval(timer)
          codeSending.value = false
        }
      }, 1000)
    } catch (error) {
      logger.error('验证码发送失败:', error)
      message.error(t('login.codeSendFailed'))
      codeSending.value = false
    }
  }

  /**
   * 验证手机登录表单并登录
   */
  const handlePhoneLogin = async () => {
    if (!phoneFormRef.value) return

    try {
      logger.info('开始手机登录表单验证')
      const result = await phoneFormRef.value.validate()
      if (!result?.valid) {
        logger.warn('手机登录表单验证失败')
        return
      }

      await phoneLogin()
    } catch (errors) {
      logger.warn('手机登录表单验证异常:', errors)
    }
  }

  /**
   * 执行手机验证码登录
   */
  const phoneLogin = async () => {
    phoneLoading.value = true
    logger.info('开始执行手机登录', { phone: phoneFormData.phone })

    try {
      // 假设这里调用登录API
      return true
    } catch (error: any) {
      logger.error('手机登录失败:', error)
      message.error(error.message || t('login.loginFailed'))
      return false
    } finally {
      phoneLoading.value = false
    }
  }

  /**
   * 微信登录
   */
  const handleWechatLogin = () => {
    logger.info('尝试微信登录')
    message.info(t('login.wechatRedirect'))
    // 微信登录实现逻辑
    // 通常会重定向到微信授权页面
  }

  /**
   * 跳转到忘记密码页面
   */
  const goToForgotPassword = () => {
    logger.info('跳转到忘记密码页面')
    router.push({ path: '/hub0001/forgot-password' })
  }

  /**
   * 检查登录状态并进行重定向
   */
  const checkLoginStatus = () => {
    // 如果已登录，直接跳转到首页
    if (store.user.isAuthenticated) {
      logger.info('用户已登录，重定向到首页')
      window.location.href = '/dashboard'
    }
  }

  /**
   * 获取系统版本信息
   */
  const fetchVersion = async () => {
    try {
      logger.info('开始获取系统版本信息')
      const response = await hub0001Api.getVersion()
      
      // 使用 format.ts 中的工具类处理响应
      if (isApiSuccess(response)) {
        const versionData = parseJsonData<{ version: string }>(response, { version: '' })
        if (versionData && versionData.version) {
        appVersion.value = versionData.version || ''
        // 更新全局store中的版本信息
        store.global.setAppVersion(appVersion.value)
          logger.info('版本信息获取成功', { version: appVersion.value })
        } else {
          logger.warn('版本信息数据解析失败', response)
        }
      } else {
        logger.warn('版本信息响应无效', response)
      }
    } catch (error) {
      logger.error('获取版本信息失败:', error)
      // 使用默认值
      appVersion.value = ''
    }
  }

  onUnmounted(() => {
    stopLockCountdown()
  })

  // 初始化验证码 - 延迟执行，优先保证LCP
  onMounted(() => {
    logger.info('LoginAuth组件挂载，开始初始化')

    // 立即检查登录状态
    checkLoginStatus()

    // 立即获取版本信息（不阻塞主渲染）
    fetchVersion()

    // 延迟验证码获取，确保LCP元素优先渲染
    requestAnimationFrame(() => {
      setTimeout(() => {
        refreshCaptcha()
      }, 200) // 200ms延迟，确保LCP完成
    })
  })

  return {
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
    handleWechatLogin,
    refreshCaptcha,
    goToForgotPassword,
    tCommon,
    t,
  }
}
