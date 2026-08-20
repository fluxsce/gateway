/**
 * 用户设置管理 hook
 * 处理用户个人资料、密码修改、系统设置等
 */
import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { store } from '@/stores'
import type {
  RsFormRules,
  RsFormValidationResult,
  RsSelectModelValue,
  RsSwitchValue,
} from '@/ui'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { computed, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { changePassword, editUser, getUserInfo } from '../api'
import type { User } from '../types'

/** RsForm 暴露的校验与重置方法 */
interface RsFormExpose {
  validate: () => Promise<RsFormValidationResult>
  clearValidation: (names?: string | string[]) => void
  resetFields: (names?: string | string[]) => void
}

/** 前端扩展类型：在后端 User 结构体基础上增加 deptName（来自关联查询或缓存） */
type UserWithDeptName = User & { deptName?: string }

/**
 * 用户设置页业务逻辑：资料编辑、改密、主题/语言等本地偏好。
 */
export function useUserSettings() {
  const { t } = useModuleI18n('hub0002')
  const message = useAppMessage()
  const i18n = useI18n()

  const profileFormRef = ref<RsFormExpose | null>(null)
  const passwordFormRef = ref<RsFormExpose | null>(null)

  const loading = ref(false)
  const savingProfile = ref(false)
  const changingPassword = ref(false)

  // 用户信息：直接查询而不是从 store 获取
  const userInfo = ref<UserWithDeptName | null>(null)

  const profileForm = reactive<Partial<UserWithDeptName>>({
    userId: '',
    userName: '',
    realName: '',
    email: '',
    mobile: '',
    gender: 0,
    avatar: '',
    deptName: '',
    tenantId: '',
  })

  /**
   * 从接口拉取当前登录用户信息并回填资料表单。
   */
  const fetchUserInfo = async () => {
    try {
      loading.value = true

      if (!store.user.isAuthenticated) {
        message.error(t('profile.fetchFailed'))
        return
      }

      const result = await getUserInfo(store.user.userId, store.user.tenantId)

      if (isApiSuccess(result)) {
        const userData = parseJsonData<UserWithDeptName>(result)

        userInfo.value = userData

        Object.assign(profileForm, {
          userId: userData.userId,
          userName: userData.userName,
          realName: userData.realName,
          email: userData.email || '',
          mobile: userData.mobile || '',
          gender: userData.gender || 0,
          avatar: userData.avatar || '',
          deptName: userData.deptName || '',
          tenantId: userData.tenantId,
        })
      } else {
        message.error(getApiMessage(result, t('profile.fetchFailed')))
      }
    } catch {
      message.error(t('profile.fetchFailed'))
    } finally {
      loading.value = false
    }
  }

  /** 用已加载的 userInfo 重置资料表单字段 */
  const initProfileForm = () => {
    if (userInfo.value) {
      Object.assign(profileForm, {
        userId: userInfo.value.userId,
        userName: userInfo.value.userName,
        realName: userInfo.value.realName,
        email: userInfo.value.email || '',
        mobile: userInfo.value.mobile || '',
        gender: userInfo.value.gender || 0,
        avatar: userInfo.value.avatar || '',
        deptName: userInfo.value.deptName || '',
        tenantId: userInfo.value.tenantId,
      })
    }
  }

  const profileRules = computed<RsFormRules>(() => ({
    realName: [
      {
        required: true,
        message: t('profile.realNameRequired'),
        trigger: 'blur',
      },
      {
        min: 2,
        max: 50,
        message: t('profile.realNameLength'),
        trigger: 'blur',
      },
    ],
    email: [
      {
        type: 'email',
        message: t('profile.emailInvalid'),
        trigger: 'blur',
      },
    ],
    mobile: [
      {
        pattern: /^1[3-9]\d{9}$/,
        message: t('profile.mobileInvalid'),
        trigger: 'blur',
      },
    ],
  }))

  const passwordForm = reactive({
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
  })

  const passwordRules = computed<RsFormRules>(() => ({
    oldPassword: [
      {
        required: true,
        message: t('password.oldPasswordRequired'),
        trigger: 'blur',
      },
    ],
    newPassword: [
      {
        required: true,
        message: t('password.newPasswordRequired'),
        trigger: 'blur',
      },
      {
        min: 8,
        max: 20,
        message: t('password.passwordLength'),
        trigger: 'blur',
      },
      {
        pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]/,
        message: t('password.passwordPattern'),
        trigger: 'blur',
      },
    ],
    confirmPassword: [
      {
        required: true,
        message: t('password.confirmPasswordRequired'),
        trigger: 'blur',
      },
      {
        validator: (value) => {
          if (value !== passwordForm.newPassword) {
            return t('password.passwordMismatch')
          }
          return true
        },
        trigger: 'blur',
      },
    ],
  }))

  const settingsForm = reactive({
    theme: store.user.theme,
    language: store.user.language,
    showGuide: false,
    notificationEnabled: true,
  })

  const themeOptions = computed(() => [
    { label: t('settings.lightTheme'), value: 'light' },
    { label: t('settings.darkTheme'), value: 'dark' },
    { label: t('settings.autoTheme'), value: 'auto' },
  ])

  const languageOptions = computed(() => [
    { label: '简体中文', value: 'zh-CN' },
    { label: 'English', value: 'en' },
  ])

  /**
   * 校验并保存个人资料；成功后刷新用户信息与 store。
   */
  const handleSaveProfile = async () => {
    try {
      const validateResult = await profileFormRef.value?.validate()
      if (validateResult && !validateResult.valid) {
        message.error(t('profile.validationFailed'))
        return
      }

      savingProfile.value = true

      if (!userInfo.value) {
        message.error(t('profile.saveFailed'))
        return
      }

      const result = await editUser({
        userId: profileForm.userId!,
        tenantId: profileForm.tenantId!,
        userName: profileForm.userName!,
        realName: profileForm.realName!,
        deptId: userInfo.value.deptId || '',
        email: profileForm.email,
        mobile: profileForm.mobile,
        gender: profileForm.gender,
        avatar: profileForm.avatar,
        statusFlag: userInfo.value.statusFlag || 'Y',
        deptAdminFlag: userInfo.value.deptAdminFlag || 'N',
        tenantAdminFlag: userInfo.value.tenantAdminFlag || 'N',
        userExpireDate: userInfo.value.userExpireDate || '',
        addTime: userInfo.value.addTime || new Date().toISOString(),
        addWho: userInfo.value.addWho || userInfo.value.userId,
        editTime: new Date().toISOString(),
        editWho: userInfo.value.userId,
        oprSeqFlag: userInfo.value.oprSeqFlag || '',
        currentVersion: userInfo.value.currentVersion || 1,
        activeFlag: userInfo.value.activeFlag || 'Y',
      })

      if (isApiSuccess(result)) {
        message.success(t('profile.saveSuccess'))

        await fetchUserInfo()

        store.user.update({
          realName: profileForm.realName!,
          email: profileForm.email,
          mobile: profileForm.mobile,
          avatar: profileForm.avatar,
        })
      } else {
        message.error(getApiMessage(result, t('profile.saveFailed')))
      }
    } catch {
      message.error(t('profile.saveFailed'))
    } finally {
      savingProfile.value = false
    }
  }

  /** 重置资料表单并清除校验态 */
  const handleResetProfile = () => {
    initProfileForm()
    profileFormRef.value?.clearValidation()
  }

  /**
   * 校验并提交密码修改；成功后清空表单并引导重新登录。
   */
  const handleChangePassword = async () => {
    try {
      const validateResult = await passwordFormRef.value?.validate()
      if (validateResult && !validateResult.valid) {
        message.error(t('password.validationFailed'))
        return
      }

      changingPassword.value = true

      const result = await changePassword({
        userId: store.user.userId,
        tenantId: store.user.tenantId,
        oldPassword: passwordForm.oldPassword,
        newPassword: passwordForm.newPassword,
      })

      if (isApiSuccess(result)) {
        message.success(t('password.changeSuccess'))
        handleResetPassword()

        setTimeout(() => {
          store.user.clearPersistedSession()
          window.location.href = '/'
        }, 2000)
      } else {
        message.error(getApiMessage(result, t('password.changeFailed')))
      }
    } catch {
      message.error(t('password.changeFailed'))
    } finally {
      changingPassword.value = false
    }
  }

  /** 清空密码表单并清除校验态 */
  const handleResetPassword = () => {
    passwordForm.oldPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''
    passwordFormRef.value?.clearValidation()
  }

  /** 从 RsSelect 取值中取出单个字符串 */
  const pickSelectString = (value: RsSelectModelValue): string => {
    const raw = Array.isArray(value) ? value[0] : value
    if (raw == null || raw === '') return ''
    if (typeof raw === 'object' && 'value' in raw) return String(raw.value)
    return String(raw)
  }

  /** 切换主题并持久化到用户设置 */
  const handleThemeChange = (value: RsSelectModelValue) => {
    const theme = pickSelectString(value)
    if (!theme) return
    settingsForm.theme = theme
    store.user.updateSettings({ theme })
    message.success(t('settings.themeChanged'))
  }

  /** 切换语言并同步 i18n locale */
  const handleLanguageChange = (value: RsSelectModelValue) => {
    const language = pickSelectString(value)
    if (!language) return
    settingsForm.language = language
    store.user.updateSettings({ language })
    i18n.locale.value = language
    message.success(t('settings.languageChanged'))
  }

  /** 通知开关：当前仅更新本地状态并提示 */
  const handleNotificationChange = (value: RsSwitchValue) => {
    const enabled = value === true || value === 'Y' || value === 1
    settingsForm.notificationEnabled = enabled
    message.success(
      enabled ? t('settings.notificationEnabled') : t('settings.notificationDisabled'),
    )
  }

  /** 引导开关：当前仅更新本地状态 */
  const handleGuideChange = (value: RsSwitchValue) => {
    settingsForm.showGuide = value === true || value === 'Y' || value === 1
  }

  /**
   * 将本地文件读为 Data URL（base64）。
   * @param file - 原始文件
   * @returns Data URL 字符串
   */
  const convertFileToBase64 = (file: File): Promise<string> => {
    return new Promise((resolve, reject) => {
      const reader = new FileReader()
      reader.readAsDataURL(file)
      reader.onload = () => resolve(reader.result as string)
      reader.onerror = () => reject(new Error('Failed to read avatar file'))
    })
  }

  return {
    userInfo,
    loading,
    fetchUserInfo,

    profileForm,
    profileFormRef,
    profileRules,
    savingProfile,
    handleSaveProfile,
    handleResetProfile,

    passwordForm,
    passwordFormRef,
    passwordRules,
    changingPassword,
    handleChangePassword,
    handleResetPassword,

    settingsForm,
    themeOptions,
    languageOptions,
    handleThemeChange,
    handleLanguageChange,
    handleNotificationChange,
    handleGuideChange,

    convertFileToBase64,
  }
}
