/**
 * 登录认证相关的业务逻辑Hook
 * 封装登录、验证码等功能，与视图层分离
 */
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { store, type UserPermissionResponse } from '@/stores'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { logger } from '@/utils/logger'
import type { User } from '@/views/hub0002/types'
import type { FormInst, FormRules } from 'naive-ui'
import { useMessage } from 'naive-ui'
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { hub0001Api } from '../api'
import type { LoginFormData, PhoneLoginFormData } from '../types'

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
  const message = useMessage()

  // 获取模块化i18n实例
  const { t } = useModuleI18n('hub0001')
  const { t: tCommon } = useModuleI18n('common')

  // 表单引用
  const formRef = ref<FormInst | null>(null)
  const phoneFormRef = ref<FormInst | null>(null)

  // 登录状态
  const loading = ref(false)
  const phoneLoading = ref(false)

  // 数字验证码相关
  const captchaId = ref('')
  const captchaCode = ref('')
  const captchaExpireAt = ref(0)
  const captchaUrl = ref('')

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

  // 表单校验规则 - 使用计算属性确保i18n加载后再构建规则
  const rules = computed<FormRules>(() => {
    return {
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
    }
  })

  // 手机登录表单校验规则
  const phoneRules = computed<FormRules>(() => {
    return {
      phone: [
        { required: true, message: t('validation.phoneRequired'), trigger: 'blur' },
        { pattern: /^1[3-9]\d{9}$/, message: t('validation.phoneFormat'), trigger: 'blur' },
      ],
      code: [
        { required: true, message: t('validation.codeRequired'), trigger: 'blur' },
        { len: 6, message: t('validation.codeLength'), trigger: 'blur' },
      ],
    }
  })

  /**
   * 复杂验证码Canvas生成器 - 异步处理避免阻塞
   * @param code 验证码字符串
   * @returns Promise<string> 生成的 Data URL
   */
  const generateComplexCaptchaCanvas = async (code: string): Promise<string> => {
    return new Promise((resolve) => {
      // 使用requestAnimationFrame确保在下一帧执行，避免阻塞当前帧
      requestAnimationFrame(() => {
        try {
          const canvas = document.createElement('canvas')
          const ctx = canvas.getContext('2d')
          if (!ctx) {
            resolve('')
            return
          }

          // 设置画布尺寸
          canvas.width = 140
          canvas.height = 50

          // 创建复杂背景渐变
          const bgGradient = ctx.createLinearGradient(0, 0, canvas.width, canvas.height)
          bgGradient.addColorStop(0, '#f8f9fa')
          bgGradient.addColorStop(0.3, '#e9ecef')
          bgGradient.addColorStop(0.7, '#dee2e6')
          bgGradient.addColorStop(1, '#ced4da')
          ctx.fillStyle = bgGradient
          ctx.fillRect(0, 0, canvas.width, canvas.height)

          // 优化纹理背景 - 减少数量提升性能
          for (let i = 0; i < 100; i++) {
            ctx.fillStyle = `rgba(${Math.random() * 50 + 200}, ${Math.random() * 50 + 200}, ${Math.random() * 50 + 200}, 0.1)`
            ctx.fillRect(Math.random() * canvas.width, Math.random() * canvas.height, 2, 2)
          }

          // 绘制复杂干扰线 - 贝塞尔曲线
          ctx.strokeStyle = 'rgba(78, 67, 118, 0.4)'
          ctx.lineWidth = 2
          for (let i = 0; i < 5; i++) {
            ctx.beginPath()
            const startX = Math.random() * canvas.width
            const startY = Math.random() * canvas.height
            const cp1X = Math.random() * canvas.width
            const cp1Y = Math.random() * canvas.height
            const cp2X = Math.random() * canvas.width
            const cp2Y = Math.random() * canvas.height
            const endX = Math.random() * canvas.width
            const endY = Math.random() * canvas.height

            ctx.moveTo(startX, startY)
            ctx.bezierCurveTo(cp1X, cp1Y, cp2X, cp2Y, endX, endY)
            ctx.stroke()
          }

          // 绘制波浪干扰线
          ctx.strokeStyle = 'rgba(78, 67, 118, 0.3)'
          ctx.lineWidth = 1.5
          for (let i = 0; i < 3; i++) {
            ctx.beginPath()
            ctx.moveTo(0, canvas.height / 2 + Math.sin(i) * 10)
            for (let x = 0; x <= canvas.width; x += 5) {
              const y =
                canvas.height / 2 + Math.sin((x + i * 50) * 0.02) * 15 + Math.random() * 10 - 5
              ctx.lineTo(x, y)
            }
            ctx.stroke()
          }

          // 绘制多种形状干扰
          const shapes = ['circle', 'rect', 'triangle', 'star']
          for (let i = 0; i < 15; i++) {
            const shape = shapes[Math.floor(Math.random() * shapes.length)]
            const x = Math.random() * canvas.width
            const y = Math.random() * canvas.height
            const size = Math.random() * 8 + 2

            ctx.fillStyle = `rgba(78, 67, 118, ${Math.random() * 0.3 + 0.1})`
            ctx.beginPath()

            switch (shape) {
              case 'circle':
                ctx.arc(x, y, size, 0, 2 * Math.PI)
                break
              case 'rect':
                ctx.rect(x - size / 2, y - size / 2, size, size)
                break
              case 'triangle':
                ctx.moveTo(x, y - size)
                ctx.lineTo(x - size, y + size)
                ctx.lineTo(x + size, y + size)
                ctx.closePath()
                break
              case 'star':
                // 简化的星形
                ctx.moveTo(x, y - size)
                ctx.lineTo(x + size / 3, y + size / 3)
                ctx.lineTo(x - size, y)
                ctx.lineTo(x + size, y)
                ctx.lineTo(x - size / 3, y + size / 3)
                ctx.closePath()
                break
            }
            ctx.fill()
          }

          // 绘制验证码文字 - 复杂样式
          const chars = code.split('')
          const colors = ['#4e4376', '#5a4e8c', '#6b5b95', '#7b68a2', '#8b789f', '#9b88b2']
          const fonts = [
            'bold 24px Georgia',
            'bold 22px "Times New Roman"',
            'bold 26px Arial',
            'bold 23px Verdana',
          ]

          chars.forEach((char, index) => {
            const x = 15 + index * 20
            const y = canvas.height / 2 + (Math.random() - 0.5) * 8

            // 随机字体和颜色
            ctx.font = fonts[Math.floor(Math.random() * fonts.length)]
            ctx.textAlign = 'center'
            ctx.textBaseline = 'middle'

            // 随机旋转角度
            const rotation = (Math.random() - 0.5) * 0.6

            ctx.save()
            ctx.translate(x, y)
            ctx.rotate(rotation)

            // 绘制文字阴影
            ctx.fillStyle = 'rgba(0, 0, 0, 0.3)'
            ctx.fillText(char, 2, 2)

            // 绘制文字描边
            ctx.strokeStyle = 'rgba(255, 255, 255, 0.8)'
            ctx.lineWidth = 3
            ctx.strokeText(char, 0, 0)

            // 绘制主文字
            ctx.fillStyle = colors[index % colors.length]
            ctx.fillText(char, 0, 0)

            // 添加发光效果
            ctx.shadowColor = colors[index % colors.length]
            ctx.shadowBlur = 3
            ctx.fillText(char, 0, 0)

            ctx.restore()
          })

          // 添加整体扭曲效果
          const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height)
          const data = imageData.data

          // 简单的波浪扭曲
          for (let y = 0; y < canvas.height; y++) {
            for (let x = 0; x < canvas.width; x++) {
              const offset = Math.sin(x * 0.1) * 2
              const newY = Math.min(Math.max(y + offset, 0), canvas.height - 1)

              if (newY !== y) {
                const sourceIndex = (y * canvas.width + x) * 4
                const targetIndex = (Math.floor(newY) * canvas.width + x) * 4

                // 交换像素
                if (targetIndex < data.length && sourceIndex < data.length) {
                  for (let i = 0; i < 4; i++) {
                    const temp = data[sourceIndex + i]
                    data[sourceIndex + i] = data[targetIndex + i]
                    data[targetIndex + i] = temp
                  }
                }
              }
            }
          }
          ctx.putImageData(imageData, 0, 0)

          logger.debug('复杂验证码Canvas生成完成')
          resolve(canvas.toDataURL('image/png'))
        } catch (error) {
          logger.error('Canvas验证码生成失败:', error)
          resolve('')
        }
      })
    })
  }

  // 防抖计时器
  let refreshTimer: ReturnType<typeof setTimeout> | null = null

  /**
   * 刷新验证码 - 完全异步处理
   */
  const refreshCaptcha = async () => {
    // 防抖处理 - 300ms内只允许一次刷新
    if (refreshTimer) {
      clearTimeout(refreshTimer)
    }

    refreshTimer = setTimeout(async () => {
      try {
        logger.info('开始刷新验证码')

        // 异步获取验证码数据
        const response = await hub0001Api.getCaptcha()
        logger.debug('验证码API响应:', response)

        // 使用 format.ts 中的工具类处理响应
        if (isApiSuccess(response)) {
          const captchaData = parseJsonData<any>(response, null)
          if (captchaData) {
            // 立即更新验证码数据，不等待Canvas生成
            captchaCode.value = captchaData.code
            captchaId.value = captchaData.captchaId
            captchaExpireAt.value = captchaData.expireAt

            logger.info('验证码数据更新成功', {
              captchaId: captchaData.captchaId,
              hasCode: !!captchaData.code,
            })

            // 异步生成Canvas - 完全不阻塞主线程
            if (captchaData.code) {
              // 使用Web Worker的思路，完全异步
              const generateCanvas = async () => {
                try {
                  const canvasDataUrl = await generateComplexCaptchaCanvas(captchaData.code)
                  captchaUrl.value = canvasDataUrl
                  logger.debug('验证码Canvas生成并更新完成')
                } catch (error) {
                  logger.error('Canvas生成异常:', error)
                  // 降级处理 - 使用简单的文本显示
                  captchaUrl.value = ''
                }
              }

              // 延迟Canvas生成，确保LCP优先完成
              if ('requestIdleCallback' in window) {
                requestIdleCallback(generateCanvas, { timeout: 3000 })
              } else {
                setTimeout(generateCanvas, 500) // 延迟500ms，确保页面LCP完成
              }
            }
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

  /**
   * 验证表单并登录
   */
  const handleLogin = async () => {
    if (!formRef.value) return

    try {
      logger.info('开始表单验证')
      // 验证表单
      await formRef.value.validate()
      logger.info('表单验证通过')

      // 调用登录方法
      await login(formData)
    } catch (errors) {
      logger.warn('表单验证失败:', errors)
      // 表单验证错误，不处理，由表单自动显示错误信息
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
        // 登录失败
        const errorMsg = getApiMessage(response, t('login.loginFailed'))
        message.error(errorMsg)
        logger.warn('登录失败', {
          errMsg: response.errMsg,
          popMsg: response.popMsg,
        })
        refreshCaptcha() // 刷新验证码
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
        router.push({ path: '/dashboard' })
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
      // 验证表单
      await phoneFormRef.value.validate()

      // 调用手机登录方法
      await phoneLogin()
    } catch (errors) {
      logger.warn('手机登录表单验证失败:', errors)
      // 表单验证错误，不处理，由表单自动显示错误信息
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
    rules,
    phoneRules,
    loading,
    phoneLoading,
    captchaCode,
    captchaUrl,
    captchaExpireAt,
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
