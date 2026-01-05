/**
 * 用户设置管理hook
 * 处理用户个人资料、密码修改、系统设置等
 */
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { store } from '@/stores'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import type { FormInst, FormRules } from 'naive-ui'
import { useMessage } from 'naive-ui'
import { computed, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { changePassword, editUser, getUserInfo } from '../api'
import type { User } from '../types'

// 前端扩展类型：在后端 User 结构体基础上增加 deptName（来自关联查询或缓存）
type UserWithDeptName = User & { deptName?: string }

export function useUserSettings() {
  const { t } = useModuleI18n('hub0002')
  const { t: tCommon } = useModuleI18n('common')
  const message = useMessage()
  const i18n = useI18n()

  // Form refs
  const profileFormRef = ref<FormInst | null>(null)
  const passwordFormRef = ref<FormInst | null>(null)

  // Loading states
  const loading = ref(false)
  const savingProfile = ref(false)
  const changingPassword = ref(false)

  // User info - 直接查询而不是从 store 获取
  const userInfo = ref<UserWithDeptName | null>(null)

  // Profile form
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

  // Fetch user info from API
  const fetchUserInfo = async () => {
    try {
      loading.value = true
      
      // 检查用户是否已登录
      if (!store.user.isAuthenticated) {
        message.error(t('profile.fetchFailed'))
        return
      }

      const result = await getUserInfo(store.user.userId, store.user.tenantId)
      
      if (isApiSuccess(result)) {
        // 使用 parseJsonData 从 bizData 中解析用户信息
        const userData = parseJsonData<UserWithDeptName>(result)
        
        userInfo.value = userData
        
        // 初始化表单
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
    } catch (error) {
      console.error('Fetch user info error:', error)
      message.error(t('profile.fetchFailed'))
    } finally {
      loading.value = false
    }
  }

  // Initialize profile form
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

  // Profile validation rules
  const profileRules: FormRules = {
    realName: [
      {
        required: true,
        message: () => t('profile.realNameRequired'),
        trigger: ['blur', 'input'],
      },
      {
        min: 2,
        max: 50,
        message: () => t('profile.realNameLength'),
        trigger: ['blur', 'input'],
      },
    ],
    email: [
      {
        type: 'email',
        message: () => t('profile.emailInvalid'),
        trigger: ['blur', 'input'],
      },
    ],
    mobile: [
      {
        pattern: /^1[3-9]\d{9}$/,
        message: () => t('profile.mobileInvalid'),
        trigger: ['blur', 'input'],
      },
    ],
  }

  // Password form
  const passwordForm = reactive({
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
  })

  // Password validation rules
  const passwordRules: FormRules = {
    oldPassword: [
      {
        required: true,
        message: () => t('password.oldPasswordRequired'),
        trigger: ['blur', 'input'],
      },
    ],
    newPassword: [
      {
        required: true,
        message: () => t('password.newPasswordRequired'),
        trigger: ['blur', 'input'],
      },
      {
        min: 8,
        max: 20,
        message: () => t('password.passwordLength'),
        trigger: ['blur', 'input'],
      },
      {
        pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]/,
        message: () => t('password.passwordPattern'),
        trigger: ['blur', 'input'],
      },
    ],
    confirmPassword: [
      {
        required: true,
        message: () => t('password.confirmPasswordRequired'),
        trigger: ['blur', 'input'],
      },
      {
        validator: (rule: any, value: string) => {
          if (value !== passwordForm.newPassword) {
            return new Error(t('password.passwordMismatch'))
          }
          return true
        },
        trigger: ['blur', 'input'],
      },
    ],
  }

  // Settings form
  const settingsForm = reactive({
    theme: store.user.theme,
    language: store.user.language,
    showGuide: false, // 简化版 store 中没有此字段，使用默认值
    notificationEnabled: true, // 简化版 store 中没有此字段，使用默认值
  })

  // Theme options
  const themeOptions = computed(() => [
    { label: t('settings.lightTheme'), value: 'light' },
    { label: t('settings.darkTheme'), value: 'dark' },
    { label: t('settings.autoTheme'), value: 'auto' },
  ])

  // Language options
  const languageOptions = computed(() => [
    { label: '简体中文', value: 'zh-CN' },
    { label: 'English', value: 'en' },
  ])

  // Handle save profile
  const handleSaveProfile = async () => {
    try {
      await profileFormRef.value?.validate()

      savingProfile.value = true

      // 使用 editUser API，需要传递完整的用户信息
      if (!userInfo.value) {
        message.error(t('profile.saveFailed'))
        return
      }

      // 从 extObj 中获取完整的用户数据（如果有的话）
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

        // 重新查询用户信息
        await fetchUserInfo()

        // 同时更新 store 中的用户信息（保持同步）
        store.user.update({
          realName: profileForm.realName!,
          email: profileForm.email,
          mobile: profileForm.mobile,
          avatar: profileForm.avatar,
        })
      } else {
        message.error(getApiMessage(result, t('profile.saveFailed')))
      }
    } catch (error: any) {
      if (error?.errorFields) {
        message.error(t('profile.validationFailed'))
      } else {
        console.error('Save profile error:', error)
        message.error(t('profile.saveFailed'))
      }
    } finally {
      savingProfile.value = false
    }
  }

  // Handle reset profile
  const handleResetProfile = () => {
    initProfileForm()
    profileFormRef.value?.restoreValidation()
  }

  // Handle change password
  const handleChangePassword = async () => {
    try {
      await passwordFormRef.value?.validate()

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

        // Optionally redirect to login after password change
        setTimeout(() => {
          store.user.clearUserInfo()
          window.location.href = '/'
        }, 2000)
      } else {
        message.error(getApiMessage(result, t('password.changeFailed')))
      }
    } catch (error: any) {
      if (error?.errorFields) {
        message.error(t('password.validationFailed'))
      } else {
        console.error('Change password error:', error)
        message.error(t('password.changeFailed'))
      }
    } finally {
      changingPassword.value = false
    }
  }

  // Handle reset password form
  const handleResetPassword = () => {
    passwordForm.oldPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''
    passwordFormRef.value?.restoreValidation()
  }

  // Handle theme change
  const handleThemeChange = (value: string) => {
    store.user.updateSettings({ theme: value })
    message.success(t('settings.themeChanged'))
  }

  // Handle language change
  const handleLanguageChange = (value: string) => {
    store.user.updateSettings({ language: value })
    i18n.locale.value = value
    message.success(t('settings.languageChanged'))
  }

  // Handle notification change
  const handleNotificationChange = (value: boolean) => {
    // 简化版 store 暂不支持 notificationEnabled，仅显示消息
    settingsForm.notificationEnabled = value
    message.success(
      value ? t('settings.notificationEnabled') : t('settings.notificationDisabled')
    )
  }

  // Handle guide change
  const handleGuideChange = (value: boolean) => {
    // 简化版 store 暂不支持 showGuide，仅更新本地状态
    settingsForm.showGuide = value
  }

  // 将文件转换为 base64
  const convertFileToBase64 = (file: File): Promise<string> => {
    return new Promise((resolve, reject) => {
      const reader = new FileReader()
      reader.readAsDataURL(file)
      reader.onload = () => resolve(reader.result as string)
      reader.onerror = (error) => reject(error)
    })
  }

  return {
    // User info
    userInfo,
    loading,
    fetchUserInfo,

    // Profile
    profileForm,
    profileFormRef,
    profileRules,
    savingProfile,
    handleSaveProfile,
    handleResetProfile,

    // Password
    passwordForm,
    passwordFormRef,
    passwordRules,
    changingPassword,
    handleChangePassword,
    handleResetPassword,

    // Settings
    settingsForm,
    themeOptions,
    languageOptions,
    handleThemeChange,
    handleLanguageChange,
    handleNotificationChange,
    handleGuideChange,

    // Avatar
    convertFileToBase64,
  }
}

