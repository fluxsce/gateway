<template>
  <div class="user-settings-container">
    <n-spin :show="loading">
      <n-card :title="t('title')" :bordered="false">
        <n-tabs v-model:value="activeTab" type="line" animated>
        <!-- 个人资料 -->
        <n-tab-pane name="profile" :tab="t('tabs.profile')">
          <n-form
            ref="profileFormRef"
            :model="profileForm"
            :rules="profileRules"
            label-placement="left"
            label-width="120"
            size="medium"
            class="profile-form"
          >
            <div class="avatar-section">
              <n-form-item :label="t('profile.avatar')">
                <div class="avatar-upload">
                  <n-avatar
                    :size="80"
                    :src="avatarPreview || profileForm.avatar || defaultAvatar"
                    class="user-avatar"
                  />
                  <n-upload
                    :show-file-list="false"
                    :custom-request="handleAvatarUpload"
                    accept="image/*"
                    @before-upload="beforeAvatarUpload"
                  >
                    <n-button size="small" class="upload-btn">
                      {{ t('profile.changeAvatar') }}
                    </n-button>
                  </n-upload>
                </div>
              </n-form-item>
            </div>

            <n-form-item :label="t('profile.userId')" path="userId">
              <n-input v-model:value="profileForm.userId" disabled />
            </n-form-item>

            <n-form-item :label="t('profile.userName')" path="userName">
              <n-input v-model:value="profileForm.userName" disabled />
            </n-form-item>

            <n-form-item :label="t('profile.realName')" path="realName">
              <n-input
                v-model:value="profileForm.realName"
                :placeholder="t('profile.realNamePlaceholder')"
              />
            </n-form-item>

            <n-form-item :label="t('profile.email')" path="email">
              <n-input
                v-model:value="profileForm.email"
                :placeholder="t('profile.emailPlaceholder')"
              />
            </n-form-item>

            <n-form-item :label="t('profile.mobile')" path="mobile">
              <n-input
                v-model:value="profileForm.mobile"
                :placeholder="t('profile.mobilePlaceholder')"
              />
            </n-form-item>

            <n-form-item :label="t('profile.gender')" path="gender">
              <n-radio-group v-model:value="profileForm.gender">
                <n-space>
                  <n-radio :value="1">{{ t('profile.male') }}</n-radio>
                  <n-radio :value="2">{{ t('profile.female') }}</n-radio>
                  <n-radio :value="0">{{ t('profile.unknown') }}</n-radio>
                </n-space>
              </n-radio-group>
            </n-form-item>

            <n-form-item :label="t('profile.deptName')">
              <n-input v-model:value="profileForm.deptName" disabled />
            </n-form-item>

            <n-form-item>
              <n-space>
                <n-button
                  type="primary"
                  :loading="savingProfile"
                  @click="handleSaveProfile"
                >
                  {{ t('common.save') }}
                </n-button>
                <n-button @click="handleResetProfile">
                  {{ t('common.reset') }}
                </n-button>
              </n-space>
            </n-form-item>
          </n-form>
        </n-tab-pane>

        <!-- 修改密码 -->
        <n-tab-pane name="password" :tab="t('tabs.password')">
          <n-form
            ref="passwordFormRef"
            :model="passwordForm"
            :rules="passwordRules"
            label-placement="left"
            label-width="120"
            size="medium"
            class="password-form"
          >
            <n-form-item :label="t('password.oldPassword')" path="oldPassword">
              <n-input
                v-model:value="passwordForm.oldPassword"
                type="password"
                show-password-on="click"
                :placeholder="t('password.oldPasswordPlaceholder')"
              />
            </n-form-item>

            <n-form-item :label="t('password.newPassword')" path="newPassword">
              <n-input
                v-model:value="passwordForm.newPassword"
                type="password"
                show-password-on="click"
                :placeholder="t('password.newPasswordPlaceholder')"
              />
            </n-form-item>

            <n-form-item :label="t('password.confirmPassword')" path="confirmPassword">
              <n-input
                v-model:value="passwordForm.confirmPassword"
                type="password"
                show-password-on="click"
                :placeholder="t('password.confirmPasswordPlaceholder')"
              />
            </n-form-item>

            <n-alert type="info" :show-icon="false" class="password-tips">
              <div v-html="t('password.tips')"></div>
            </n-alert>

            <n-form-item>
              <n-space>
                <n-button
                  type="primary"
                  :loading="changingPassword"
                  @click="handleChangePassword"
                >
                  {{ t('password.changeButton') }}
                </n-button>
                <n-button @click="handleResetPassword">
                  {{ t('common.reset') }}
                </n-button>
              </n-space>
            </n-form-item>
          </n-form>
        </n-tab-pane>

        <!-- 系统设置 -->
        <n-tab-pane name="settings" :tab="t('tabs.settings')">
          <n-form
            :model="settingsForm"
            label-placement="left"
            label-width="120"
            size="medium"
            class="settings-form"
          >
            <n-form-item :label="t('settings.theme')">
              <n-select
                v-model:value="settingsForm.theme"
                :options="themeOptions"
                @update:value="handleThemeChange"
              />
            </n-form-item>

            <n-form-item :label="t('settings.language')">
              <n-select
                v-model:value="settingsForm.language"
                :options="languageOptions"
                @update:value="handleLanguageChange"
              />
            </n-form-item>

            <n-form-item :label="t('settings.notification')">
              <n-switch
                v-model:value="settingsForm.notificationEnabled"
                @update:value="handleNotificationChange"
              />
              <span class="setting-desc">{{ t('settings.notificationDesc') }}</span>
            </n-form-item>

            <n-form-item :label="t('settings.showGuide')">
              <n-switch
                v-model:value="settingsForm.showGuide"
                @update:value="handleGuideChange"
              />
              <span class="setting-desc">{{ t('settings.showGuideDesc') }}</span>
            </n-form-item>
          </n-form>
        </n-tab-pane>

        <!-- 账号信息 -->
        <n-tab-pane name="account" :tab="t('tabs.account')">
          <n-descriptions
            :column="1"
            label-placement="left"
            label-style="width: 120px; font-weight: 500;"
            bordered
            class="account-info"
          >
            <n-descriptions-item :label="t('account.userId')">
              {{ userInfo?.userId }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('account.tenantId')">
              {{ userInfo?.tenantId }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('account.roles')">
              <n-space>
                <n-tag
                  v-for="role in (userInfo?.roles || [])"
                  :key="role"
                  type="info"
                  size="small"
                >
                  {{ role }}
                </n-tag>
              </n-space>
            </n-descriptions-item>
            <n-descriptions-item :label="t('account.tenantAdmin')">
              <n-tag :type="isTenantAdmin ? 'success' : 'default'" size="small">
                {{ isTenantAdmin ? t('common.yes') : t('common.no') }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item :label="t('account.deptAdmin')">
              <n-tag :type="isDeptAdmin ? 'success' : 'default'" size="small">
                {{ isDeptAdmin ? t('common.yes') : t('common.no') }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item :label="t('account.status')">
              <n-tag :type="statusFlag === 'Y' ? 'success' : 'error'" size="small">
                {{ statusFlag === 'Y' ? t('account.enabled') : t('account.disabled') }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item :label="t('account.lastLoginTime')">
              {{ userInfo?.lastLoginTime || t('common.notAvailable') }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('account.lastLoginIp')">
              {{ userInfo?.lastLoginIp || t('common.notAvailable') }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('account.userExpireDate')">
              {{ userInfo?.userExpireDate || t('common.notAvailable') }}
            </n-descriptions-item>
          </n-descriptions>
        </n-tab-pane>
      </n-tabs>
    </n-card>
    </n-spin>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { useUserSettings } from './hooks'
import defaultAvatar from '@/assets/images/default-avatar.png'
import type { UploadCustomRequestOptions, UploadFileInfo } from 'naive-ui'

const { t } = useModuleI18n('hub0002')
const { t: tCommon } = useModuleI18n('common')
const message = useMessage()

// Active tab
const activeTab = ref('profile')

// Avatar preview
const avatarPreview = ref<string>('')

// Use user settings hook
const {
  userInfo,
  loading,
  fetchUserInfo,
  profileForm,
  profileFormRef,
  profileRules,
  savingProfile,
  passwordForm,
  passwordFormRef,
  passwordRules,
  changingPassword,
  settingsForm,
  themeOptions,
  languageOptions,
  handleSaveProfile: originalHandleSaveProfile,
  handleResetProfile: originalHandleResetProfile,
  handleChangePassword,
  handleResetPassword,
  handleThemeChange,
  handleLanguageChange,
  handleNotificationChange,
  handleGuideChange,
  convertFileToBase64,
} = useUserSettings()

// Computed properties
const isTenantAdmin = computed(() => userInfo.value?.tenantAdminFlag === 'Y')
const isDeptAdmin = computed(() => userInfo.value?.deptAdminFlag === 'Y')
const statusFlag = computed(() => userInfo.value?.statusFlag)

// 头像上传前验证
const beforeAvatarUpload = async (data: {
  file: UploadFileInfo
  fileList: UploadFileInfo[]
}): Promise<boolean> => {
  const file = data.file.file as File
  
  // 验证文件大小 (max 500KB for base64)
  if (file.size > 500 * 1024) {
    message.error(t('profile.avatarSizeLimitBase64'))
    return false
  }

  // 验证文件类型
  const allowedTypes = ['image/jpeg', 'image/png', 'image/gif']
  if (!allowedTypes.includes(file.type)) {
    message.error(t('profile.avatarTypeInvalid'))
    return false
  }

  return true
}

// 处理头像上传（预览）
const handleAvatarUpload = async ({ file }: UploadCustomRequestOptions) => {
  try {
    if (!file.file) return

    // 转换为 base64
    const base64 = await convertFileToBase64(file.file as File)
    
    // 检查 base64 大小（约为原文件的 4/3，应小于 500KB）
    if (base64.length > 500 * 1024) {
      message.error(t('profile.avatarSizeLimitBase64'))
      return
    }

    // 设置预览
    avatarPreview.value = base64
    profileForm.avatar = base64
  } catch (error) {
    console.error('Avatar upload error:', error)
    message.error(t('profile.avatarUploadFailed'))
  }
}

// 保存个人资料（包含头像）
const handleSaveProfile = async () => {
  await originalHandleSaveProfile()
  // 保存成功后清除预览
  avatarPreview.value = ''
}

// 重置表单
const handleResetProfile = () => {
  originalHandleResetProfile()
  avatarPreview.value = ''
}

// Initialize
onMounted(async () => {
  // 加载用户信息
  await fetchUserInfo()
})
</script>

<style lang="scss" scoped>
.user-settings-container {
  padding: 16px;
  max-width: 900px;
  margin: 0 auto;

  :deep(.n-card) {
    border-radius: 8px;
  }

  .profile-form,
  .password-form,
  .settings-form {
    max-width: 600px;
    margin-top: 24px;
  }

  .avatar-section {
    margin-bottom: 24px;

    .avatar-upload {
      display: flex;
      align-items: center;
      gap: 16px;

      .user-avatar {
        border: 2px solid var(--n-border-color);
        border-radius: 50%;
      }

      .upload-btn {
        margin-left: 8px;
      }
    }
  }

  .password-tips {
    margin-bottom: 16px;
    font-size: 13px;

    :deep(.n-alert__content) {
      line-height: 1.6;
    }
  }

  .setting-desc {
    margin-left: 12px;
    font-size: 13px;
    color: var(--n-text-color-3);
  }

  .account-info {
    margin-top: 24px;
  }

  :deep(.n-form-item-label) {
    font-weight: 500;
  }
}
</style>

