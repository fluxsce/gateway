<template>
  <div class="user-settings-container">
    <RsLoading :loading="loading" overlay block size="lg" />
    <RsCard :title="t('title')" variant="outlined">
      <RsTabs
        v-model="activeTab"
        :items="tabItems"
        variant="line"
        size="md"
        borderless
        content-gap="lg"
        class="settings-tabs"
      >
        <!-- 个人资料 -->
        <template #profile>
          <RsForm
            ref="profileFormRef"
            class="profile-form"
            label-position="left"
            label-width="7.5rem"
            size="md"
            gap="md"
            :rules="profileRules"
          >
            <div class="avatar-section">
              <div class="field-row">
                <RsLabel class="field-row__label">{{ t('profile.avatar') }}</RsLabel>
                <div class="avatar-upload">
                  <RsAvatar
                    size="lg"
                    :src="avatarPreview || profileForm.avatar || defaultAvatar"
                    class="user-avatar"
                  />
                  <RsButton size="sm" class="upload-btn" @click="triggerAvatarPick">
                    {{ t('profile.changeAvatar') }}
                  </RsButton>
                  <input
                    id="user-settings-avatar-input"
                    ref="avatarInputRef"
                    type="file"
                    accept="image/jpeg,image/png,image/gif"
                    class="avatar-input"
                    :aria-label="t('profile.changeAvatar')"
                    @change="onAvatarFileChange"
                  />
                </div>
              </div>
            </div>

            <RsInput
              v-model="profileForm.userId"
              name="userId"
              :label="t('profile.userId')"
              disabled
            />
            <RsInput
              v-model="profileForm.userName"
              name="userName"
              :label="t('profile.userName')"
              disabled
            />
            <RsInput
              v-model="profileForm.realName"
              name="realName"
              :label="t('profile.realName')"
              :placeholder="t('profile.realNamePlaceholder')"
            />
            <RsInput
              v-model="profileForm.email"
              name="email"
              :label="t('profile.email')"
              :placeholder="t('profile.emailPlaceholder')"
            />
            <RsInput
              v-model="profileForm.mobile"
              name="mobile"
              :label="t('profile.mobile')"
              :placeholder="t('profile.mobilePlaceholder')"
            />

            <div class="field-row">
              <RsLabel class="field-row__label">{{ t('profile.gender') }}</RsLabel>
              <RsRadio v-model="profileForm.gender" name="gender" size="md">
                <RsRadioItem :value="1">{{ t('profile.male') }}</RsRadioItem>
                <RsRadioItem :value="2">{{ t('profile.female') }}</RsRadioItem>
                <RsRadioItem :value="0">{{ t('profile.unknown') }}</RsRadioItem>
              </RsRadio>
            </div>

            <RsInput
              v-model="profileForm.deptName"
              name="deptName"
              :label="t('profile.deptName')"
              disabled
            />

            <div class="form-actions">
              <RsButton
                variant="primary"
                :loading="savingProfile"
                @click="handleSaveProfile"
              >
                {{ t('common.save') }}
              </RsButton>
              <RsButton variant="secondary" @click="handleResetProfile">
                {{ t('common.reset') }}
              </RsButton>
            </div>
          </RsForm>
        </template>

        <!-- 修改密码 -->
        <template #password>
          <RsForm
            ref="passwordFormRef"
            class="password-form"
            label-position="left"
            label-width="7.5rem"
            size="md"
            gap="md"
            :rules="passwordRules"
          >
            <RsInput
              v-model="passwordForm.oldPassword"
              name="oldPassword"
              type="password"
              autocomplete="current-password"
              :label="t('password.oldPassword')"
              :placeholder="t('password.oldPasswordPlaceholder')"
            />
            <RsInput
              v-model="passwordForm.newPassword"
              name="newPassword"
              type="password"
              autocomplete="new-password"
              :label="t('password.newPassword')"
              :placeholder="t('password.newPasswordPlaceholder')"
            />
            <RsInput
              v-model="passwordForm.confirmPassword"
              name="confirmPassword"
              type="password"
              autocomplete="new-password"
              :label="t('password.confirmPassword')"
              :placeholder="t('password.confirmPasswordPlaceholder')"
            />

            <RsAlert type="info" class="password-tips">
              <div class="password-tips__title">{{ t('password.tipsTitle') }}</div>
              <ul class="password-tips__list">
                <li>{{ t('password.tipLength') }}</li>
                <li>{{ t('password.tipUppercase') }}</li>
                <li>{{ t('password.tipLowercase') }}</li>
                <li>{{ t('password.tipNumber') }}</li>
                <li>{{ t('password.tipSpecial') }}</li>
              </ul>
            </RsAlert>

            <div class="form-actions">
              <RsButton
                variant="primary"
                :loading="changingPassword"
                @click="handleChangePassword"
              >
                {{ t('password.changeButton') }}
              </RsButton>
              <RsButton variant="secondary" @click="handleResetPassword">
                {{ t('common.reset') }}
              </RsButton>
            </div>
          </RsForm>
        </template>

        <!-- 系统设置 -->
        <template #settings>
          <RsForm
            class="settings-form"
            label-position="left"
            label-width="7.5rem"
            size="md"
            gap="md"
          >
            <div class="field-row">
              <RsLabel class="field-row__label">{{ t('settings.theme') }}</RsLabel>
              <RsSelect
                v-model="settingsForm.theme"
                :options="themeOptions"
                match-trigger-width
                block
                @update:model-value="handleThemeChange"
              />
            </div>

            <div class="field-row">
              <RsLabel class="field-row__label">{{ t('settings.language') }}</RsLabel>
              <RsSelect
                v-model="settingsForm.language"
                :options="languageOptions"
                match-trigger-width
                block
                @update:model-value="handleLanguageChange"
              />
            </div>

            <div class="field-row">
              <RsLabel class="field-row__label">{{ t('settings.notification') }}</RsLabel>
              <div class="switch-row">
                <RsSwitch
                  v-model="settingsForm.notificationEnabled"
                  @update:model-value="handleNotificationChange"
                />
                <span class="setting-desc">{{ t('settings.notificationDesc') }}</span>
              </div>
            </div>

            <div class="field-row">
              <RsLabel class="field-row__label">{{ t('settings.showGuide') }}</RsLabel>
              <div class="switch-row">
                <RsSwitch
                  v-model="settingsForm.showGuide"
                  @update:model-value="handleGuideChange"
                />
                <span class="setting-desc">{{ t('settings.showGuideDesc') }}</span>
              </div>
            </div>
          </RsForm>
        </template>

        <!-- 账号信息 -->
        <template #account>
          <RsDescriptions
            :columns="1"
            label-placement="left"
            bordered
            class="account-info"
          >
            <RsDescriptionsItem :label="t('account.userId')">
              {{ userInfo?.userId }}
            </RsDescriptionsItem>
            <RsDescriptionsItem :label="t('account.tenantId')">
              {{ userInfo?.tenantId }}
            </RsDescriptionsItem>
            <RsDescriptionsItem :label="t('account.roles')">
              <div class="tag-list">
                <RsTag
                  v-for="role in userInfo?.roles || []"
                  :key="role"
                  variant="info"
                  size="sm"
                >
                  {{ role }}
                </RsTag>
              </div>
            </RsDescriptionsItem>
            <RsDescriptionsItem :label="t('account.tenantAdmin')">
              <RsTag :variant="isTenantAdmin ? 'success' : 'default'" size="sm">
                {{ isTenantAdmin ? t('common.yes') : t('common.no') }}
              </RsTag>
            </RsDescriptionsItem>
            <RsDescriptionsItem :label="t('account.deptAdmin')">
              <RsTag :variant="isDeptAdmin ? 'success' : 'default'" size="sm">
                {{ isDeptAdmin ? t('common.yes') : t('common.no') }}
              </RsTag>
            </RsDescriptionsItem>
            <RsDescriptionsItem :label="t('account.status')">
              <RsTag :variant="statusFlag === 'Y' ? 'success' : 'danger'" size="sm">
                {{ statusFlag === 'Y' ? t('account.enabled') : t('account.disabled') }}
              </RsTag>
            </RsDescriptionsItem>
            <RsDescriptionsItem :label="t('account.lastLoginTime')">
              {{ userInfo?.lastLoginTime || t('common.notAvailable') }}
            </RsDescriptionsItem>
            <RsDescriptionsItem :label="t('account.lastLoginIp')">
              {{ userInfo?.lastLoginIp || t('common.notAvailable') }}
            </RsDescriptionsItem>
            <RsDescriptionsItem :label="t('account.userExpireDate')">
              {{ userInfo?.userExpireDate || t('common.notAvailable') }}
            </RsDescriptionsItem>
          </RsDescriptions>
        </template>
      </RsTabs>
    </RsCard>
  </div>
</template>

<script setup lang="ts">
import defaultAvatar from '@/assets/images/default-avatar.png'
import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import {
  RsAlert,
  RsAvatar,
  RsButton,
  RsCard,
  RsDescriptions,
  RsDescriptionsItem,
  RsForm,
  RsInput,
  RsLabel,
  RsLoading,
  RsRadio,
  RsRadioItem,
  RsSelect,
  RsSwitch,
  RsTabs,
  RsTag,
  type RsTabItem,
} from '@/ui'
import { computed, onMounted, ref } from 'vue'
import { useUserSettings } from './hooks'

const { t } = useModuleI18n('hub0002')
const message = useAppMessage()

const activeTab = ref('profile')
const avatarPreview = ref('')
const avatarInputRef = ref<HTMLInputElement | null>(null)

const tabItems = computed<RsTabItem[]>(() => [
  { value: 'profile', label: t('tabs.profile') },
  { value: 'password', label: t('tabs.password') },
  { value: 'settings', label: t('tabs.settings') },
  { value: 'account', label: t('tabs.account') },
])

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

const isTenantAdmin = computed(() => userInfo.value?.tenantAdminFlag === 'Y')
const isDeptAdmin = computed(() => userInfo.value?.deptAdminFlag === 'Y')
const statusFlag = computed(() => userInfo.value?.statusFlag)

/** 触发隐藏的头像文件选择 */
const triggerAvatarPick = () => {
  avatarInputRef.value?.click()
}

/** 头像文件选择后做校验并转成 base64 预览 */
const onAvatarFileChange = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return

  if (file.size > 500 * 1024) {
    message.error(t('profile.avatarSizeLimitBase64'))
    return
  }

  const allowedTypes = ['image/jpeg', 'image/png', 'image/gif']
  if (!allowedTypes.includes(file.type)) {
    message.error(t('profile.avatarTypeInvalid'))
    return
  }

  try {
    const base64 = await convertFileToBase64(file)
    if (base64.length > 500 * 1024) {
      message.error(t('profile.avatarSizeLimitBase64'))
      return
    }
    avatarPreview.value = base64
    profileForm.avatar = base64
  } catch {
    message.error(t('profile.avatarUploadFailed'))
  }
}

const handleSaveProfile = async () => {
  await originalHandleSaveProfile()
  avatarPreview.value = ''
}

const handleResetProfile = () => {
  originalHandleResetProfile()
  avatarPreview.value = ''
}

onMounted(async () => {
  await fetchUserInfo()
})
</script>

<style lang="scss" scoped>
.user-settings-container {
  position: relative;
  padding: 16px;
  max-width: 1100px;
  margin: 0 auto;
}

.settings-tabs {
  margin-top: 4px;
}

.profile-form,
.password-form,
.settings-form {
  max-width: 720px;
  margin-top: 8px;
}

.avatar-section {
  margin-bottom: 8px;
}

.field-row {
  display: grid;
  grid-template-columns: 7.5rem minmax(0, 1fr);
  align-items: center;
  column-gap: 0.75rem;
}

.field-row__label {
  justify-self: start;
}

.avatar-upload {
  display: flex;
  align-items: center;
  gap: 16px;
}



.avatar-input {
  display: none;
}

.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 8px;
  padding-left: calc(7.5rem + 0.75rem);
}

.password-tips {
  margin-bottom: 8px;
  font-size: 13px;
  line-height: 1.6;
}

.password-tips__title {
  font-weight: 500;
  margin-bottom: 4px;
}

.password-tips__list {
  margin: 8px 0 0;
  padding-left: 20px;
}

.switch-row {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.setting-desc {
  font-size: 13px;
  color: var(--rs-muted);
}

.account-info {
  margin-top: 16px;
}

.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>
