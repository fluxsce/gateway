<template>
  <div class="service-form-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <n-breadcrumb>
        <n-breadcrumb-item @click="handleGoBack">
          {{ t('serviceManagement') }}
        </n-breadcrumb-item>
        <n-breadcrumb-item>
          {{ isEdit ? t('editService') : t('addService') }}
        </n-breadcrumb-item>
      </n-breadcrumb>
      
      <div class="header-actions">
        <n-button @click="handleGoBack" quaternary>
          <template #icon>
            <n-icon><ArrowBackOutline /></n-icon>
          </template>
          {{ t('actions.back') }}
        </n-button>
      </div>
    </div>

    <!-- 表单内容 -->
    <n-card size="large" class="form-card">
        <n-form
          ref="formRef"
          :model="formData"
          :rules="formRules"
          label-placement="left"
          :label-width="140"
          require-mark-placement="left"
        >
          <!-- 基本信息 -->
          <n-card :title="t('basicInfo')" size="small" style="margin-bottom: 24px;">
            <n-grid :cols="3" :x-gap="24" :y-gap="16">
              <n-gi>
                <n-form-item :label="t('serviceName')" path="serviceName">
                  <n-input 
                    v-model:value="formData.serviceName"
                    :placeholder="t('serviceNamePlaceholder')"
                    maxlength="50"
                    show-count
                    :disabled="isEdit"
                  />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item :label="t('columns.groupName')" path="serviceGroupId">
                  <n-input
                    :value="selectedGroupName"
                    :placeholder="t('selectServiceGroup')"
                    readonly
                    @click="handleOpenGroupSelection"
                    style="cursor: pointer;"
                  >
                    <template #suffix>
                      <n-button text @click="handleOpenGroupSelection">
                        <n-icon><ChevronDownOutline /></n-icon>
                      </n-button>
                    </template>
                  </n-input>
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item :label="t('groupName')" path="groupName">
                  <n-input 
                    v-model:value="formData.groupName"
                    :placeholder="t('groupNamePlaceholder')"
                    readonly
                    style="background-color: var(--input-color-disabled);"
                  />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item :label="t('columns.registryType')" path="registryType">
                  <n-select
                    v-model:value="formData.registryType"
                    :options="registryTypeOptions"
                    :placeholder="t('selectRegistryType')"
                  />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item :label="t('columns.activeFlag')" path="activeFlag">
                  <n-switch
                    v-model:value="formData.activeFlag"
                    :checked-value="'Y'"
                    :unchecked-value="'N'"
                  >
                    <template #checked>{{ t('status.Y') }}</template>
                    <template #unchecked>{{ t('status.N') }}</template>
                  </n-switch>
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item :label="t('maxInstances')" path="maxInstances">
                  <n-input-number
                    v-model:value="formData.maxInstances"
                    :placeholder="t('maxInstancesPlaceholder')"
                    :min="1"
                    :max="100"
                    style="width: 100%"
                  />
                </n-form-item>
              </n-gi>
            </n-grid>

            <n-form-item :label="t('serviceDescription')" path="serviceDescription">
              <n-input 
                v-model:value="formData.serviceDescription"
                type="textarea"
                :placeholder="t('serviceDescriptionPlaceholder')"
                :rows="4"
                maxlength="200"
                show-count
              />
            </n-form-item>
          </n-card>

          <!-- 网络配置 -->
          <n-card :title="t('networkConfig')" size="small" style="margin-bottom: 24px;">
            <n-grid :cols="3" :x-gap="24" :y-gap="16">
              <n-gi>
                <n-form-item :label="t('columns.protocolType')" path="protocolType">
                  <n-select
                    v-model:value="formData.protocolType"
                    :options="protocolOptions"
                    :placeholder="t('selectProtocol')"
                  />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item :label="t('columns.contextPath')" path="contextPath">
                  <n-input 
                    v-model:value="formData.contextPath"
                    :placeholder="t('contextPathPlaceholder')"
                  />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item :label="t('columns.loadBalanceStrategy')" path="loadBalanceStrategy">
                  <n-select
                    v-model:value="formData.loadBalanceStrategy"
                    :options="loadBalanceOptions"
                    :placeholder="t('selectLoadBalance')"
                  />
                </n-form-item>
              </n-gi>
            </n-grid>
          </n-card>

          <!-- Nacos配置 -->
          <fieldset v-if="showNacosConfig" style="border: none; padding: 0; margin: 0;">
            <n-card :title="t('nacosConfig.title')" size="small" style="margin-bottom: 24px;">
              <!-- 服务器配置 -->
              <div style="margin-bottom: 16px;">
                <n-text strong style="display: block; margin-bottom: 12px;">{{ t('nacosConfig.serverConfig') }}</n-text>
                <div v-for="(server, index) in formData.nacosConfig.servers" :key="index" style="margin-bottom: 16px; padding: 16px; border: 1px solid var(--border-color); border-radius: 6px; background-color: var(--color-embedded);">
                  <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px;">
                    <n-text strong>{{ t('nacosConfig.serverTitle', { number: index + 1 }) }}</n-text>
                    <n-button 
                      v-if="formData.nacosConfig.servers.length > 1"
                      size="small" 
                      type="error" 
                      quaternary 
                      @click="removeNacosServer(index)"
                    >
                      <template #icon>
                        <n-icon><TrashOutline /></n-icon>
                      </template>
                    </n-button>
                  </div>
                  <n-grid :cols="3" :x-gap="16" :y-gap="12">
                    <n-gi>
                      <n-form-item :label="t('nacosConfig.host')" :path="`nacosConfig.servers[${index}].host`">
                        <n-input 
                          v-model:value="server.host"
                          :placeholder="t('nacosConfig.hostPlaceholder')"
                        />
                      </n-form-item>
                    </n-gi>
                    <n-gi>
                      <n-form-item :label="t('nacosConfig.port')" :path="`nacosConfig.servers[${index}].port`">
                        <n-input-number
                          v-model:value="server.port"
                          :placeholder="t('nacosConfig.portPlaceholder')"
                          :min="1"
                          :max="65535"
                          style="width: 100%"
                        />
                      </n-form-item>
                    </n-gi>
                    <n-gi>
                      <n-form-item :label="t('nacosConfig.scheme')" :path="`nacosConfig.servers[${index}].scheme`">
                        <n-select
                          v-model:value="server.scheme"
                          :options="[{label: 'HTTP', value: 'http'}, {label: 'HTTPS', value: 'https'}]"
                        />
                      </n-form-item>
                    </n-gi>
                    <n-gi>
                      <n-form-item :label="t('nacosConfig.grpcPort')" :path="`nacosConfig.servers[${index}].grpcPort`">
                        <n-input-number
                          v-model:value="server.grpcPort"
                          :placeholder="t('nacosConfig.grpcPortPlaceholder')"
                          :min="1"
                          :max="65535"
                          style="width: 100%"
                        />
                      </n-form-item>
                    </n-gi>
                    <n-gi>
                      <n-form-item label="上下文路径" :path="`nacosConfig.servers[${index}].contextPath`">
                        <n-input 
                          v-model:value="server.contextPath"
                          placeholder="默认: /nacos"
                        />
                      </n-form-item>
                    </n-gi>
                  </n-grid>
                </div>
                <n-button size="small" type="primary" dashed @click="addNacosServer" style="width: 100%;">
                  <template #icon>
                    <n-icon><AddOutline /></n-icon>
                  </template>
                  {{ t('nacosConfig.addServer') }}
                </n-button>
              </div>

              <!-- 基本配置 -->
              <div style="margin-bottom: 16px;">
                <n-text strong style="display: block; margin-bottom: 12px;">{{ t('nacosConfig.basicConfig') }}</n-text>
                <n-grid :cols="3" :x-gap="16" :y-gap="12">
                  <n-gi>
                    <n-form-item :label="t('nacosConfig.namespace')" path="nacosConfig.namespace">
                      <n-input 
                        v-model:value="formData.nacosConfig.namespace"
                        :placeholder="t('nacosConfig.namespacePlaceholder')"
                      />
                    </n-form-item>
                  </n-gi>
                  <n-gi>
                    <n-form-item :label="t('nacosConfig.group')" path="nacosConfig.group">
                      <n-input 
                        v-model:value="formData.nacosConfig.group"
                        :placeholder="t('nacosConfig.groupPlaceholder')"
                      />
                    </n-form-item>
                  </n-gi>
                  <n-gi>
                    <n-form-item :label="t('nacosConfig.timeout')" path="nacosConfig.timeout">
                      <n-input-number
                        v-model:value="formData.nacosConfig.timeout"
                        :placeholder="t('nacosConfig.timeoutPlaceholder')"
                        :min="1"
                        :max="60"
                        style="width: 100%"
                      >
                        <template #suffix>{{ t('seconds') }}</template>
                      </n-input-number>
                    </n-form-item>
                  </n-gi>
                  <n-gi>
                    <n-form-item :label="t('nacosConfig.beatInterval')" path="nacosConfig.beatInterval">
                      <n-input-number
                        v-model:value="formData.nacosConfig.beatInterval"
                        :placeholder="t('nacosConfig.beatIntervalPlaceholder')"
                        :min="1"
                        :max="60"
                        style="width: 100%"
                      >
                        <template #suffix>{{ t('seconds') }}</template>
                      </n-input-number>
                    </n-form-item>
                  </n-gi>
                </n-grid>
              </div>

              <!-- 认证配置 -->
              <div style="margin-bottom: 16px;">
                <n-text strong style="display: block; margin-bottom: 12px;">{{ t('nacosConfig.authConfig') }}</n-text>
                <n-grid :cols="2" :x-gap="16" :y-gap="12">
                  <n-gi>
                    <n-form-item :label="t('nacosConfig.username')">
                      <n-input 
                        v-model:value="formData.nacosConfig.username"
                        :placeholder="t('nacosConfig.usernamePlaceholder')"
                      />
                    </n-form-item>
                  </n-gi>
                  <n-gi>
                    <n-form-item :label="t('nacosConfig.password')">
                      <n-input 
                        v-model:value="formData.nacosConfig.password"
                       
                        :placeholder="t('nacosConfig.passwordPlaceholder')"
                        show-password-on="click"
                        :input-props="{ autocomplete: 'off' }"
                      />
                    </n-form-item>
                  </n-gi>
                  <n-gi>
                    <n-form-item :label="t('nacosConfig.accessKey')">
                      <n-input 
                        v-model:value="formData.nacosConfig.accessKey"
                        :placeholder="t('nacosConfig.accessKeyPlaceholder')"
                      />
                    </n-form-item>
                  </n-gi>
                  <n-gi>
                    <n-form-item :label="t('nacosConfig.secretKey')">
                      <n-input 
                        v-model:value="formData.nacosConfig.secretKey"
                       
                        :placeholder="t('nacosConfig.secretKeyPlaceholder')"
                        show-password-on="click"
                        :input-props="{ autocomplete: 'off' }"
                      />
                    </n-form-item>
                  </n-gi>
                </n-grid>
              </div>

              <!-- 高级配置 -->
              <n-collapse>
                <n-collapse-item :title="t('nacosConfig.advancedConfig')" name="advanced">
                  <n-space vertical :size="16">
                    <!-- 日志配置 -->
                    <div>
                      <n-text strong style="display: block; margin-bottom: 12px;">日志配置</n-text>
                      <n-grid :cols="3" :x-gap="16" :y-gap="12">
                        <n-gi>
                          <n-form-item :label="t('nacosConfig.logLevel')" path="nacosConfig.logLevel">
                            <n-select
                              v-model:value="formData.nacosConfig.logLevel"
                              :options="[
                                {label: 'Debug', value: 'debug'},
                                {label: 'Info', value: 'info'},
                                {label: 'Warn', value: 'warn'},
                                {label: 'Error', value: 'error'}
                              ]"
                            />
                          </n-form-item>
                        </n-gi>
                        <n-gi>
                          <n-form-item label="日志目录" path="nacosConfig.logDir">
                            <n-input 
                              v-model:value="formData.nacosConfig.logDir"
                              placeholder="默认: /tmp/nacos/log"
                            />
                          </n-form-item>
                        </n-gi>
                        <n-gi>
                          <n-form-item label="输出到控制台" path="nacosConfig.appendToStdout">
                            <n-switch v-model:value="formData.nacosConfig.appendToStdout" />
                          </n-form-item>
                        </n-gi>
                      </n-grid>
                    </div>

                    <!-- 缓存配置 -->
                    <div>
                      <n-text strong style="display: block; margin-bottom: 12px;">缓存配置</n-text>
                      <n-grid :cols="2" :x-gap="16" :y-gap="12">
                        <n-gi>
                          <n-form-item label="缓存目录">
                            <n-input 
                              v-model:value="formData.nacosConfig.cacheDir"
                              placeholder="默认: /tmp/nacos/cache"
                            />
                          </n-form-item>
                        </n-gi>
                        <n-gi>
                          <n-form-item :label="t('nacosConfig.updateThreadNum')">
                            <n-input-number
                              v-model:value="formData.nacosConfig.updateThreadNum"
                              :placeholder="t('nacosConfig.updateThreadNumPlaceholder')"
                              :min="1"
                              :max="100"
                              style="width: 100%"
                            />
                          </n-form-item>
                        </n-gi>
                        <n-gi>
                          <n-form-item :label="t('nacosConfig.notLoadCacheAtStart')">
                            <n-switch v-model:value="formData.nacosConfig.notLoadCacheAtStart" />
                          </n-form-item>
                        </n-gi>
                        <n-gi>
                          <n-form-item :label="t('nacosConfig.disableUseSnapShot')">
                            <n-switch v-model:value="formData.nacosConfig.disableUseSnapShot" />
                          </n-form-item>
                        </n-gi>
                      </n-grid>
                    </div>

                    <!-- TLS配置 -->
                    <div>
                      <n-text strong style="display: block; margin-bottom: 12px;">TLS配置</n-text>
                      <n-grid :cols="2" :x-gap="16" :y-gap="12">
                        <n-gi>
                          <n-form-item :label="t('nacosConfig.enableTLS')">
                            <n-switch v-model:value="formData.nacosConfig.enableTLS" />
                          </n-form-item>
                        </n-gi>
                        <n-gi>
                          <n-form-item :label="t('nacosConfig.trustAll')">
                            <n-switch v-model:value="formData.nacosConfig.trustAll" />
                          </n-form-item>
                        </n-gi>
                      </n-grid>
                    </div>
                  </n-space>
                </n-collapse-item>
              </n-collapse>
            </n-card>
          </fieldset>

          <!-- 健康检查配置 -->
          <n-card v-if="showHealthCheckConfig" :title="t('healthCheckConfig')" size="small" style="margin-bottom: 24px;">
            <n-grid :cols="3" :x-gap="24" :y-gap="16">
              <n-gi>
                <n-form-item :label="t('healthCheckUrl')" path="healthCheckUrl">
                  <n-input 
                    v-model:value="formData.healthCheckUrl"
                    :placeholder="t('healthCheckUrlPlaceholder')"
                  />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item :label="t('healthCheckInterval')" path="healthCheckIntervalSeconds">
                  <n-input-number
                    v-model:value="formData.healthCheckIntervalSeconds"
                    :placeholder="t('healthCheckIntervalPlaceholder')"
                    :min="5"
                    :max="300"
                    style="width: 100%"
                  >
                    <template #suffix>{{ t('seconds') }}</template>
                  </n-input-number>
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item :label="t('healthCheckTimeout')" path="healthCheckTimeoutSeconds">
                  <n-input-number
                    v-model:value="formData.healthCheckTimeoutSeconds"
                    :placeholder="t('healthCheckTimeoutPlaceholder')"
                    :min="1"
                    :max="60"
                    style="width: 100%"
                  >
                    <template #suffix>{{ t('seconds') }}</template>
                  </n-input-number>
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item :label="t('healthCheckType')" path="healthCheckType">
                  <n-select
                    v-model:value="formData.healthCheckType"
                    :options="healthCheckTypeOptions"
                    :placeholder="t('selectHealthCheckType')"
                  />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item :label="t('healthCheckMode')" path="healthCheckMode">
                  <n-select
                    v-model:value="formData.healthCheckMode"
                    :options="healthCheckModeOptions"
                    :placeholder="t('selectHealthCheckMode')"
                  />
                </n-form-item>
              </n-gi>
            </n-grid>
          </n-card>

          
          <!-- 服务实例配置 -->
          <n-card :title="t('instanceList')" size="small" style="margin-bottom: 24px;">
            <div class="instance-header">
              <div class="instance-header-left">
                <n-text>{{ t('instanceConfig') }}</n-text>
                <n-text v-if="formData.instances.length > 0" depth="3" style="margin-left: 8px;">
                  ({{ formData.instances.length }} {{ t('instances') }})
                </n-text>
              </div>
              <div class="instance-header-actions">
                <n-button 
                  size="small" 
                  @click="handleRefreshInstances" 
                  :loading="loading"
                  :disabled="!isEdit || !formData.serviceName"
                >
                  <template #icon>
                    <n-icon><RefreshOutline /></n-icon>
                  </template>
                  {{ t('actions.refresh') }}
                </n-button>
                <n-tooltip v-if="!isEdit" placement="top" trigger="hover">
                  <template #trigger>
                    <n-button size="small" type="primary" @click="handleAddInstanceClick" :disabled="!isEdit">
                      <template #icon>
                        <n-icon><AddOutline /></n-icon>
                      </template>
                      {{ t('actions.addInstance') }}
                    </n-button>
                  </template>
                  {{ t('saveServiceBeforeAddingInstance') }}
                </n-tooltip>
                <n-button v-else size="small" type="primary" @click="handleAddInstanceClick">
                  <template #icon>
                    <n-icon><AddOutline /></n-icon>
                  </template>
                  {{ t('actions.addInstance') }}
                </n-button>
              </div>
            </div>
            
            <div class="instance-table" style="margin-top: 16px;">
              <n-spin :show="loading">
                <n-data-table
                  v-if="formData.instances && formData.instances.length > 0"
                  :columns="instanceColumns"
                  :data="formData.instances"
                  :pagination="false"
                  size="small"
                  :max-height="300"
                  :bordered="true"
                />
                <n-empty 
                  v-else
                  :description="t('table.noInstances')"
                  size="small"
                  style="padding: 20px 0;"
                />
              </n-spin>
            </div>
            
            <n-drawer
              v-model:show="instanceDrawerVisible"
              :width="500"
              placement="right"
            >
              <n-drawer-content :title="isEditInstance ? t('editInstance') : t('addInstance')">
                <n-form
                  ref="instanceFormRef"
                  :model="instanceForm"
                  :rules="instanceRules"
                  label-placement="left"
                  :label-width="120"
                >
                  <n-form-item :label="t('columns.hostAddress')" path="hostAddress">
                    <n-input 
                      v-model:value="instanceForm.hostAddress" 
                      :placeholder="t('search.placeholder.hostAddress')"
                    />
                  </n-form-item>
                  
                  <n-form-item :label="t('columns.portNumber')" path="portNumber">
                    <n-input-number
                      v-model:value="instanceForm.portNumber"
                      :min="1"
                      :max="65535"
                      style="width: 100%"
                    />
                  </n-form-item>
                  
                  <n-form-item :label="t('columns.weightValue')" path="weightValue">
                    <n-input-number
                      v-model:value="instanceForm.weightValue"
                      :min="1"
                      :max="100"
                      :default-value="1"
                      style="width: 100%"
                    />
                  </n-form-item>
                  
                  <n-form-item :label="t('columns.instanceStatus')" path="instanceStatus">
                    <n-select
                      v-model:value="instanceForm.instanceStatus"
                      :options="instanceStatusOptions"
                    />
                  </n-form-item>
                  
                  <n-form-item :label="t('columns.healthStatus')" path="healthStatus">
                    <n-select
                      v-model:value="instanceForm.healthStatus"
                      :options="healthStatusOptions"
                    />
                  </n-form-item>
                  
                  <n-form-item :label="t('columns.clientType')" path="clientType">
                    <n-select
                      v-model:value="instanceForm.clientType"
                      :options="clientTypeOptions"
                    />
                  </n-form-item>
                  
                  <n-form-item :label="t('columns.tempInstanceFlag')" path="tempInstanceFlag">
                    <n-switch
                      v-model:value="instanceForm.tempInstanceFlag"
                      :checked-value="'Y'"
                      :unchecked-value="'N'"
                    >
                      <template #checked>{{ t('status.temporary') }}</template>
                      <template #unchecked>{{ t('status.permanent') }}</template>
                    </n-switch>
                  </n-form-item>
                  
                  <div style="margin-top: 24px; display: flex; justify-content: center; gap: 12px;">
                    <n-button @click="instanceDrawerVisible = false">
                      {{ t('cancel') }}
                    </n-button>
                    <n-button type="primary" @click="handleInstanceSubmit">
                      {{ t('submit') }}
                    </n-button>
                  </div>
                </n-form>
              </n-drawer-content>
            </n-drawer>
          </n-card>

          <!-- 扩展配置 (可折叠，用于后期扩展) -->
          <n-card v-if="showExtensionConfigSection" :title="t('extensionConfig')" size="small" style="margin-bottom: 32px;">
            <template #header-extra>
              <n-button 
                text 
                @click="showExtensionConfig = !showExtensionConfig"
                style="padding: 0;"
              >
                <template #icon>
                  <n-icon>
                    <component :is="showExtensionConfig ? 'ChevronUpOutline' : 'ChevronDownOutline'" />
                  </n-icon>
                </template>
                {{ showExtensionConfig ? t('collapse') : t('expand') }}
              </n-button>
            </template>
            
            <n-collapse-transition :show="showExtensionConfig">
              <n-space vertical :size="24">
                <!-- 元数据和标签 -->
                <n-card :title="t('metadataAndTags')" size="small" embedded>
                  <n-grid :cols="1" :y-gap="16">
                    <n-gi>
                      <n-form-item :label="t('metadataJson')" path="metadataJson">
                        <n-input 
                          v-model:value="formData.metadataJson"
                          type="textarea"
                          :placeholder="t('metadataJsonPlaceholder')"
                          :rows="4"
                        />
                      </n-form-item>
                    </n-gi>
                    <n-gi>
                      <n-form-item :label="t('tagsJson')" path="tagsJson">
                        <n-input 
                          v-model:value="formData.tagsJson"
                          type="textarea"
                          :placeholder="t('tagsJsonPlaceholder')"
                          :rows="3"
                        />
                      </n-form-item>
                    </n-gi>
                  </n-grid>
                </n-card>

                <!-- 备注和扩展属性 -->
                <n-card :title="t('notesAndExtProperty')" size="small" embedded>
                  <n-grid :cols="1" :y-gap="16">
                    <n-gi>
                      <n-form-item :label="t('noteText')" path="noteText">
                        <n-input 
                          v-model:value="formData.noteText"
                          type="textarea"
                          :placeholder="t('noteTextPlaceholder')"
                          :rows="3"
                          maxlength="500"
                          show-count
                        />
                      </n-form-item>
                    </n-gi>
                    <n-gi>
                      <n-form-item :label="t('extProperty')" path="extProperty">
                        <n-input 
                          v-model:value="formData.extProperty"
                          type="textarea"
                          :placeholder="t('extPropertyPlaceholder')"
                          :rows="4"
                        />
                      </n-form-item>
                    </n-gi>
                  </n-grid>
                </n-card>

                <!-- 预留字段 (用于后期扩展Nacos等服务) -->
                <n-card :title="t('reservedFields')" size="small" embedded>
                  <template #header-extra>
                    <n-tag size="small" type="info">{{ t('forFutureExpansion') }}</n-tag>
                  </template>
                  <n-grid :cols="2" :x-gap="16" :y-gap="12">
                    <n-gi>
                      <n-form-item :label="t('reservedField', { number: 1 })" path="reserved1">
                        <n-input v-model:value="formData.reserved1" :placeholder="t('reservedFieldPlaceholder', { number: 1 })" />
                      </n-form-item>
                    </n-gi>
                    <n-gi>
                      <n-form-item :label="t('reservedField', { number: 2 })" path="reserved2">
                        <n-input v-model:value="formData.reserved2" :placeholder="t('reservedFieldPlaceholder', { number: 2 })" />
                      </n-form-item>
                    </n-gi>
                    <n-gi>
                      <n-form-item :label="t('reservedField', { number: 3 })" path="reserved3">
                        <n-input v-model:value="formData.reserved3" :placeholder="t('reservedFieldPlaceholder', { number: 3 })" />
                      </n-form-item>
                    </n-gi>
                    <n-gi>
                      <n-form-item :label="t('reservedField', { number: 4 })" path="reserved4">
                        <n-input v-model:value="formData.reserved4" :placeholder="t('reservedFieldPlaceholder', { number: 4 })" />
                      </n-form-item>
                    </n-gi>
                    <n-gi>
                      <n-form-item :label="t('reservedField', { number: 5 })" path="reserved5">
                        <n-input v-model:value="formData.reserved5" :placeholder="t('reservedFieldPlaceholder', { number: 5 })" />
                      </n-form-item>
                    </n-gi>
                  </n-grid>
                  <n-divider style="margin: 12px 0;" />
                  <n-grid :cols="2" :x-gap="16" :y-gap="12">
                    <n-gi>
                      <n-form-item :label="t('reservedField', { number: 6 })" path="reserved6">
                        <n-input v-model:value="formData.reserved6" :placeholder="t('reservedFieldPlaceholder', { number: 6 })" />
                      </n-form-item>
                    </n-gi>
                    <n-gi>
                      <n-form-item :label="t('reservedField', { number: 7 })" path="reserved7">
                        <n-input v-model:value="formData.reserved7" :placeholder="t('reservedFieldPlaceholder', { number: 7 })" />
                      </n-form-item>
                    </n-gi>
                    <n-gi>
                      <n-form-item :label="t('reservedField', { number: 8 })" path="reserved8">
                        <n-input v-model:value="formData.reserved8" :placeholder="t('reservedFieldPlaceholder', { number: 8 })" />
                      </n-form-item>
                    </n-gi>
                    <n-gi>
                      <n-form-item :label="t('reservedField', { number: 9 })" path="reserved9">
                        <n-input v-model:value="formData.reserved9" :placeholder="t('reservedFieldPlaceholder', { number: 9 })" />
                      </n-form-item>
                    </n-gi>
                    <n-gi>
                      <n-form-item :label="t('reservedField', { number: 10 })" path="reserved10">
                        <n-input v-model:value="formData.reserved10" :placeholder="t('reservedFieldPlaceholder', { number: 10 })" />
                      </n-form-item>
                    </n-gi>
                  </n-grid>
                </n-card>
              </n-space>
            </n-collapse-transition>
          </n-card>
        </n-form>

        <!-- 表单操作按钮 -->
        <div class="form-actions">
          <n-space size="large" justify="center">
            <n-button size="large" @click="handleGoBack">
              {{ t('cancel') }}
            </n-button>
            <n-button size="large" @click="resetForm">
              {{ t('reset') }}
            </n-button>
            <n-button 
              type="primary" 
              size="large"
              @click="handleSubmit" 
              :loading="submitLoading"
            >
              {{ isEdit ? t('actions.update') : t('actions.create') }}
            </n-button>
          </n-space>
        </div>
      </n-card>

    <!-- 服务分组选择对话框 -->
    <ServiceGroupSelectionDialog
      v-model:show="groupSelectionVisible"
      :selected-group-id="formData.serviceGroupId"
      @confirm="handleGroupSelected"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, h } from 'vue'
import { 
  NCard, NForm, NFormItem, NInput, NInputNumber, NSelect, NSwitch,
  NButton, NButtonGroup, NSpace, NGrid, NGi, NBreadcrumb, NBreadcrumbItem, 
  NIcon, NCollapseTransition, NDivider, NTag, useMessage, 
  NDataTable, NEmpty, NDrawer, NDrawerContent, NText, NPopconfirm,
  NSpin, NTooltip, NCollapse, NCollapseItem
} from 'naive-ui'
import { 
  ArrowBackOutline, ChevronDownOutline, ChevronUpOutline, 
  AddOutline, TrashOutline, CreateOutline, RefreshOutline
} from '@vicons/ionicons5'
import ServiceGroupSelectionDialog from './ServiceGroupSelectionDialog.vue'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { 
  createService, updateService, getServiceGroups, 
  createServiceInstance, updateServiceInstance, deleteServiceInstance,
  queryServiceInstances
} from '../api'
import { safeParseJsonArray } from '@/utils/format'
import type { 
  Service, ServiceInstance, ProtocolType, LoadBalanceStrategy,
  InstanceStatus, HealthStatus, ClientType, HealthCheckType, HealthCheckMode,
  RegistryType
} from '../types'

interface Props {
  editData?: Service | null
}

interface Emits {
  (e: 'back'): void
  (e: 'success'): void
}

const props = withDefaults(defineProps<Props>(), {
  editData: null
})
const emit = defineEmits<Emits>()

// 国际化
const { t } = useModuleI18n('hub0041')

// 消息提示
const message = useMessage()

// 表单引用
const formRef = ref()

// 是否编辑模式
const isEdit = computed<boolean>(() => !!props.editData)

// 加载和提交状态
const loading = ref<boolean>(false)
const submitLoading = ref<boolean>(false)

// 服务分组相关
const serviceGroups = ref<any[]>([])
const groupSelectionVisible = ref(false)
const selectedGroupName = computed(() => {
  if (!formData.serviceGroupId) return ''
  const group = serviceGroups.value.find(g => g.serviceGroupId === formData.serviceGroupId)
  return group ? group.groupName : ''
})

// 扩展配置显示控制
const showExtensionConfig = ref(false)

// 根据注册类型显示不同配置
const showHealthCheckConfig = computed(() => formData.registryType !== 'NACOS')
const showExtensionConfigSection = computed(() => formData.registryType !== 'NACOS')
const showNacosConfig = computed(() => formData.registryType === 'NACOS')

// 表单数据
const formData = reactive({
  serviceName: '',
  serviceGroupId: '',
  groupName: '',
  serviceDescription: '',
  registryType: 'INTERNAL' as RegistryType,
  protocolType: 'HTTP' as ProtocolType,
  contextPath: '/',
  loadBalanceStrategy: 'ROUND_ROBIN' as LoadBalanceStrategy,
  healthCheckUrl: '/health',
  healthCheckIntervalSeconds: 30,
  healthCheckTimeoutSeconds: 5,
  healthCheckType: 'HTTP' as HealthCheckType,
  healthCheckMode: 'ACTIVE' as HealthCheckMode,
  activeFlag: 'Y',
  maxInstances: 10,
  // Nacos配置
  nacosConfig: {
    servers: [{
      host: '',
      port: 8848,
      grpcPort: undefined as number | undefined,
      contextPath: '/nacos',
      scheme: 'http' as 'http' | 'https'
    }],
    namespace: 'public',
    group: 'DEFAULT_GROUP',
    username: '',
    password: '',
    accessKey: '',
    secretKey: '',
    timeout: 5,
    beatInterval: 5,
    cacheDir: '',
    logDir: '',
    logLevel: 'info' as 'debug' | 'info' | 'warn' | 'error',
    updateThreadNum: 20,
    notLoadCacheAtStart: true,
    disableUseSnapShot: false,
    updateCacheWhenEmpty: false,
    appendToStdout: false,
    enableTLS: false,
    trustAll: false,
    caFile: '',
    certFile: '',
    keyFile: '',
    appName: '',
    appKey: '',
    openKMS: false,
    regionId: ''
  },
  // 实例列表
  instances: [] as ServiceInstance[],
  // 扩展字段
  metadataJson: '',
  tagsJson: '',
  noteText: '',
  extProperty: '',
  // 预留字段 (用于后期扩展Nacos等服务)
  reserved1: '',
  reserved2: '',
  reserved3: '',
  reserved4: '',
  reserved5: '',
  reserved6: '',
  reserved7: '',
  reserved8: '',
  reserved9: '',
  reserved10: ''
})

// 注册类型选项
const registryTypeOptions = computed(() => [
  { label: t('registryTypes.INTERNAL'), value: 'INTERNAL' },
  { label: t('registryTypes.NACOS'), value: 'NACOS' },
  { label: t('registryTypes.CONSUL'), value: 'CONSUL' },
  { label: t('registryTypes.EUREKA'), value: 'EUREKA' },
  { label: t('registryTypes.ETCD'), value: 'ETCD' },
  { label: t('registryTypes.ZOOKEEPER'), value: 'ZOOKEEPER' }
])

// 协议类型选项
const protocolOptions = [
  { label: 'HTTP', value: 'HTTP' },
  { label: 'HTTPS', value: 'HTTPS' },
  { label: 'TCP', value: 'TCP' },
  { label: 'UDP', value: 'UDP' },
  { label: 'GRPC', value: 'GRPC' }
]

// 健康检查类型选项
const healthCheckTypeOptions = [
  { label: 'HTTP', value: 'HTTP' },
  { label: 'TCP', value: 'TCP' }
]

// 健康检查模式选项
const healthCheckModeOptions = computed(() => [
  { label: t('healthCheckModes.active'), value: 'ACTIVE' },
  { label: t('healthCheckModes.passive'), value: 'PASSIVE' }
])

// 负载均衡策略选项
const loadBalanceOptions = computed(() => [
  { label: t('roundRobin'), value: 'ROUND_ROBIN' },
  { label: t('weightedRoundRobin'), value: 'WEIGHTED_ROUND_ROBIN' },
  { label: t('leastConnections'), value: 'LEAST_CONNECTIONS' },
  { label: t('random'), value: 'RANDOM' },
  { label: t('ipHash'), value: 'IP_HASH' }
])

// 表单验证规则
const formRules = computed(() => ({
  serviceName: [
    { required: true, message: t('serviceNameRequired'), trigger: 'blur' },
    { min: 2, max: 50, message: t('serviceNameLength'), trigger: 'blur' }
  ],
  serviceGroupId: [
    { required: true, message: t('serviceGroupRequired'), trigger: 'change' }
  ],
  groupName: [
    // groupName是只读字段，无需验证
  ],
  registryType: [
    { required: true, message: t('registryTypeRequired'), trigger: 'change' }
  ],
  serviceDescription: [
    { max: 200, message: t('serviceDescriptionLength'), trigger: 'blur' }
  ],
  protocolType: [
    { required: true, message: t('protocolTypeRequired'), trigger: 'change' }
  ],
  contextPath: [
    { required: true, message: t('contextPathRequired'), trigger: 'blur' }
  ],
  loadBalanceStrategy: [
    { required: true, message: t('loadBalanceStrategyRequired'), trigger: 'change' }
  ],
  healthCheckUrl: [
    { 
      validator: (rule: any, value: string) => {
        if (showHealthCheckConfig.value && !value) {
          return new Error(t('healthCheckUrlRequired'))
        }
        return true
      },
      trigger: 'blur' 
    }
  ],
  healthCheckIntervalSeconds: [
    { 
      validator: (rule: any, value: number) => {
        if (showHealthCheckConfig.value && !value) {
          return new Error(t('healthCheckIntervalRequired'))
        }
        return true
      },
      trigger: 'blur' 
    }
  ],
  healthCheckTimeoutSeconds: [
    { 
      validator: (rule: any, value: number) => {
        if (showHealthCheckConfig.value && !value) {
          return new Error(t('healthCheckTimeoutRequired'))
        }
        return true
      },
      trigger: 'blur' 
    }
  ],
  healthCheckType: [
    { 
      validator: (rule: any, value: string) => {
        if (showHealthCheckConfig.value && !value) {
          return new Error(t('healthCheckTypeRequired'))
        }
        return true
      },
      trigger: 'change' 
    }
  ],
  healthCheckMode: [
    { 
      validator: (rule: any, value: string) => {
        if (showHealthCheckConfig.value && !value) {
          return new Error(t('healthCheckModeRequired'))
        }
        return true
      },
      trigger: 'change' 
    }
  ],
  activeFlag: [
    { required: true, message: t('activeFlagRequired'), trigger: 'change' }
  ],
  maxInstances: [
    { required: true, type: 'number' as const, message: t('maxInstancesRequired'), trigger: 'blur' }
  ],
  // 扩展字段验证 (可选)
  metadataJson: [
    { 
      validator: (rule: any, value: string) => {
        if (value && value.trim()) {
          try {
            JSON.parse(value)
            return true
          } catch {
            return new Error(t('invalidJsonFormat'))
          }
        }
        return true
      },
      trigger: 'blur'
    }
  ],
  tagsJson: [
    { 
      validator: (rule: any, value: string) => {
        if (value && value.trim()) {
          try {
            JSON.parse(value)
            return true
          } catch {
            return new Error(t('invalidJsonFormat'))
          }
        }
        return true
      },
      trigger: 'blur'
    }
  ],
  noteText: [
    { max: 500, message: t('noteTextLength'), trigger: 'blur' }
  ],
  extProperty: [
    { 
      validator: (rule: any, value: string) => {
        if (value && value.trim()) {
          try {
            JSON.parse(value)
            return true
          } catch {
            return new Error(t('invalidJsonFormat'))
          }
        }
        return true
      },
      trigger: 'blur'
    }
  ],
  // Nacos配置验证
  'nacosConfig.namespace': [
    {
      validator: (rule: any, value: string) => {
        if (showNacosConfig.value && !value) {
          return new Error('请输入命名空间')
        }
        return true
      },
      trigger: 'blur'
    }
  ],
  'nacosConfig.group': [
    {
      validator: (rule: any, value: string) => {
        if (showNacosConfig.value && !value) {
          return new Error('请输入默认分组')
        }
        return true
      },
      trigger: 'blur'
    }
  ]
}))

// 加载服务实例列表
const loadServiceInstances = async (serviceName: string) => {
  if (!serviceName) return
  
  try {
    loading.value = true
    const response = await queryServiceInstances({ 
      serviceName,
      pageIndex: 1,
      pageSize: 1000 
    })
    
    if (response.oK) {
      // 使用format工具类解析实例数据
      const instancesData = safeParseJsonArray<ServiceInstance>(response.bizData)
      
      // 更新实例列表
      formData.instances = instancesData
      message.success(t('loadInstancesSuccess'))
    } else {
      message.error(response.errMsg || t('loadInstancesFailed'))
    }
  } catch (error) {
    console.error('Failed to load instances:', error)
    message.error(t('loadInstancesFailed'))
  } finally {
    loading.value = false
  }
}

// 初始化表单数据
const initFormData = (data?: Service | null) => {
  if (data) {
    // 填充编辑数据
    Object.assign(formData, {
      serviceName: data.serviceName || '',
      serviceGroupId: data.serviceGroupId || '',
      groupName: data.groupName || '',
      serviceDescription: data.serviceDescription || '',
      registryType: data.registryType || 'INTERNAL',
      protocolType: data.protocolType || 'HTTP',
      contextPath: data.contextPath || '/',
      loadBalanceStrategy: data.loadBalanceStrategy || 'ROUND_ROBIN',
      healthCheckUrl: data.healthCheckUrl || '/health',
      healthCheckIntervalSeconds: data.healthCheckIntervalSeconds || 30,
      healthCheckTimeoutSeconds: data.healthCheckTimeoutSeconds || 5,
      healthCheckType: data.healthCheckType || 'HTTP',
      healthCheckMode: data.healthCheckMode || 'ACTIVE',
      activeFlag: data.activeFlag || 'Y',
      maxInstances: 10,
      // Nacos配置
      nacosConfig: (() => {
        if (data.externalRegistryConfig && data.registryType === 'NACOS') {
          try {
            const config = JSON.parse(data.externalRegistryConfig)
            return {
              servers: config.servers || [{ host: '', port: 8848, grpcPort: undefined, contextPath: '/nacos', scheme: 'http' }],
              namespace: config.namespace || 'public',
              group: config.group || 'DEFAULT_GROUP',
              username: config.username || '',
              password: config.password || '',
              accessKey: config.accessKey || '',
              secretKey: config.secretKey || '',
              timeout: config.timeout || 5,
              beatInterval: config.beatInterval || 5,
              cacheDir: config.cacheDir || '',
              logDir: config.logDir || '',
              logLevel: config.logLevel || 'info',
              updateThreadNum: config.updateThreadNum || 20,
              notLoadCacheAtStart: config.notLoadCacheAtStart ?? true,
              disableUseSnapShot: config.disableUseSnapShot ?? false,
              updateCacheWhenEmpty: config.updateCacheWhenEmpty ?? false,
              appendToStdout: config.appendToStdout ?? false,
              enableTLS: config.enableTLS ?? false,
              trustAll: config.trustAll ?? false,
              caFile: config.caFile || '',
              certFile: config.certFile || '',
              keyFile: config.keyFile || '',
              appName: config.appName || '',
              appKey: config.appKey || '',
              openKMS: config.openKMS ?? false,
              regionId: config.regionId || ''
            }
          } catch {
            return {
              servers: [{ host: '', port: 8848, grpcPort: undefined, contextPath: '/nacos', scheme: 'http' }],
              namespace: 'public',
              group: 'DEFAULT_GROUP',
              username: '',
              password: '',
              accessKey: '',
              secretKey: '',
              timeout: 5,
              beatInterval: 5,
              cacheDir: '',
              logDir: '',
              logLevel: 'info',
              updateThreadNum: 20,
              notLoadCacheAtStart: true,
              disableUseSnapShot: false,
              updateCacheWhenEmpty: false,
              appendToStdout: false,
              enableTLS: false,
              trustAll: false,
              caFile: '',
              certFile: '',
              keyFile: '',
              appName: '',
              appKey: '',
              openKMS: false,
              regionId: ''
            }
          }
        }
        return {
          servers: [{ host: '', port: 8848, grpcPort: undefined, contextPath: '/nacos', scheme: 'http' }],
          namespace: 'public',
          group: 'DEFAULT_GROUP',
          username: '',
          password: '',
          accessKey: '',
          secretKey: '',
          timeout: 5,
          beatInterval: 5,
          cacheDir: '',
          logDir: '',
          logLevel: 'info',
          updateThreadNum: 20,
          notLoadCacheAtStart: true,
          disableUseSnapShot: false,
          updateCacheWhenEmpty: false,
          appendToStdout: false,
          enableTLS: false,
          trustAll: false,
          caFile: '',
          certFile: '',
          keyFile: '',
          appName: '',
          appKey: '',
          openKMS: false,
          regionId: ''
        }
      })(),
      // 实例列表 - 初始化为空数组，后续通过API加载
      instances: [],
      // 扩展字段
      metadataJson: data.metadataJson || '',
      tagsJson: data.tagsJson || '',
      noteText: data.noteText || '',
      extProperty: data.extProperty || '',
      // 预留字段
      reserved1: data.reserved1 || '',
      reserved2: data.reserved2 || '',
      reserved3: data.reserved3 || '',
      reserved4: data.reserved4 || '',
      reserved5: data.reserved5 || '',
      reserved6: data.reserved6 || '',
      reserved7: data.reserved7 || '',
      reserved8: data.reserved8 || '',
      reserved9: data.reserved9 || '',
      reserved10: data.reserved10 || ''
    })
    
    // 如果有扩展字段数据，默认展开扩展配置
    if (data.metadataJson || data.tagsJson || data.noteText || data.extProperty ||
        data.reserved1 || data.reserved2 || data.reserved3 || data.reserved4 || data.reserved5 ||
        data.reserved6 || data.reserved7 || data.reserved8 || data.reserved9 || data.reserved10) {
      showExtensionConfig.value = true
    }
    
    // 如果是编辑模式，加载实例列表
    if (isEdit.value && data.serviceName) {
      loadServiceInstances(data.serviceName)
    }
  } else {
    // 重置为默认值
    resetForm()
  }
}

// 监听编辑数据变化
watch(() => props.editData, (newData) => {
  initFormData(newData)
})

// 初始化
onMounted(() => {
  fetchServiceGroups()
  // 初始化表单数据
  initFormData(props.editData)
})

// 获取服务分组列表
const fetchServiceGroups = async () => {
  try {
    const response = await getServiceGroups()
    
    if (response.oK) {
      let responseData: any = {}
      try {
        responseData = typeof response.bizData === 'string' 
          ? JSON.parse(response.bizData) 
          : response.bizData
        
        serviceGroups.value = responseData || []
      } catch (error) {
        console.error('Failed to parse service groups data:', error)
        serviceGroups.value = []
      }
    } else {
      message.error(t('fetchServiceGroupsFailed'))
      serviceGroups.value = []
    }
  } catch (error) {
    console.error('Failed to fetch service groups:', error)
    message.error(t('fetchServiceGroupsFailed'))
    serviceGroups.value = []
  }
}

// 重置表单
const resetForm = () => {
  formRef.value?.restoreValidation()
  if (!isEdit.value) {
    Object.assign(formData, {
      serviceName: '',
      serviceGroupId: '',
      groupName: '',
      serviceDescription: '',
      registryType: 'INTERNAL',
      protocolType: 'HTTP',
      contextPath: '/',
      loadBalanceStrategy: 'ROUND_ROBIN',
      healthCheckUrl: '/health',
      healthCheckIntervalSeconds: 30,
      healthCheckTimeoutSeconds: 5,
      healthCheckType: 'HTTP',
      healthCheckMode: 'ACTIVE',
      activeFlag: 'Y',
      maxInstances: 10,
      // Nacos配置重置
      nacosConfig: {
        servers: [{ host: '', port: 8848, grpcPort: undefined, contextPath: '/nacos', scheme: 'http' }],
        namespace: 'public',
        group: 'DEFAULT_GROUP',
        username: '',
        password: '',
        accessKey: '',
        secretKey: '',
        timeout: 5,
        beatInterval: 5,
        cacheDir: '',
        logDir: '',
        logLevel: 'info',
        updateThreadNum: 20,
        notLoadCacheAtStart: true,
        disableUseSnapShot: false,
        updateCacheWhenEmpty: false,
        appendToStdout: false,
        enableTLS: false,
        trustAll: false,
        caFile: '',
        certFile: '',
        keyFile: '',
        appName: '',
        appKey: '',
        openKMS: false,
        regionId: ''
      },
      // 实例列表
      instances: [],
      // 扩展字段
      metadataJson: '',
      tagsJson: '',
      noteText: '',
      extProperty: '',
      // 预留字段
      reserved1: '',
      reserved2: '',
      reserved3: '',
      reserved4: '',
      reserved5: '',
      reserved6: '',
      reserved7: '',
      reserved8: '',
      reserved9: '',
      reserved10: ''
    })
    // 重置时收起扩展配置
    showExtensionConfig.value = false
  }
}

// 打开分组选择对话框
const handleOpenGroupSelection = () => {
  groupSelectionVisible.value = true
}

// 处理分组选择
const handleGroupSelected = (group: any) => {
  formData.serviceGroupId = group.serviceGroupId
  formData.groupName = group.groupName
}

// Nacos服务器管理
const addNacosServer = () => {
  formData.nacosConfig.servers.push({
    host: '',
    port: 8848,
    grpcPort: undefined,
    contextPath: '/nacos',
    scheme: 'http' as 'http' | 'https'
  })
}

const removeNacosServer = (index: number) => {
  if (formData.nacosConfig.servers.length > 1) {
    formData.nacosConfig.servers.splice(index, 1)
  }
}

// 返回列表页面
const handleGoBack = () => {
  emit('back')
}

// 提交表单
const handleSubmit = async () => {
  try {
    // 表单验证
    await formRef.value?.validate()
    
    submitLoading.value = true
    
    // 构建提交数据，排除不属于Service接口的字段
    const { maxInstances, nacosConfig, ...serviceData } = formData
    
    // 处理外部注册中心配置
    let externalRegistryConfig: string | undefined
    if (formData.registryType === 'NACOS') {
      externalRegistryConfig = JSON.stringify(formData.nacosConfig)
    }
    
    const submitData: Partial<Service> = {
      ...serviceData,
      externalRegistryConfig,
      // 确保将实例数组正确格式化为后端所需的格式
      instances: formData.instances.map(instance => ({
        ...instance,
        serviceName: formData.serviceName // 确保每个实例都有正确的服务名称
      }))
    }
    
    let response
    if (isEdit.value) {
      response = await updateService(submitData)
    } else {
      response = await createService(submitData)
    }
    
    if (response.oK) {
      // 只有成功时才显示成功消息
      message.success(isEdit.value ? t('updateServiceSuccess') : t('addServiceSuccess'))
      emit('success')
      emit('back')
    } else {
      // 失败时显示错误消息，优先显示后端返回的具体错误信息
      const errorMessage = response.popMsg || response.errMsg || (isEdit.value ? t('updateServiceFailed') : t('addServiceFailed'))
      message.error(errorMessage)
    }
  } catch (error) {
    console.error('Failed to submit service:', error)
    message.error(isEdit.value ? t('updateServiceFailed') : t('addServiceFailed'))
  } finally {
    submitLoading.value = false
  }
}
// 实例管理相关
// 实例表单引用
const instanceFormRef = ref()
// 实例抽屉可见性
const instanceDrawerVisible = ref(false)
// 是否为编辑实例模式
const isEditInstance = ref(false)
// 当前编辑的实例索引
const currentInstanceIndex = ref(-1)

// 实例表单数据
const instanceForm = reactive<Partial<ServiceInstance>>({
  serviceInstanceId: '',
  hostAddress: '',
  portNumber: 8080,
  contextPath: '/',
  instanceStatus: 'UP' as InstanceStatus,
  healthStatus: 'HEALTHY' as HealthStatus,
  weightValue: 1,
  clientType: 'OTHER' as ClientType,
  tempInstanceFlag: 'N'
})

// 实例状态选项
const instanceStatusOptions = computed(() => [
  { label: t('status.UP'), value: 'UP' },
  { label: t('status.DOWN'), value: 'DOWN' },
  { label: t('status.STARTING'), value: 'STARTING' },
  { label: t('status.OUT_OF_SERVICE'), value: 'OUT_OF_SERVICE' }
])

// 健康状态选项
const healthStatusOptions = computed(() => [
  { label: t('status.HEALTHY'), value: 'HEALTHY' },
  { label: t('status.UNHEALTHY'), value: 'UNHEALTHY' },
  { label: t('status.UNKNOWN'), value: 'UNKNOWN' }
])

// 客户端类型选项
const clientTypeOptions = computed(() => [
  { label: t('status.JAVA'), value: 'JAVA' },
  { label: t('status.DOTNET'), value: 'DOTNET' },
  { label: t('status.NODEJS'), value: 'NODEJS' },
  { label: t('status.PYTHON'), value: 'PYTHON' },
  { label: t('status.GO'), value: 'GO' },
  { label: t('status.OTHER'), value: 'OTHER' }
])

// 实例表单验证规则
const instanceRules = {
  hostAddress: [
    { required: true, message: t('instanceHostRequired'), trigger: 'blur' }
  ],
  portNumber: [
    { required: true, type: 'number' as const, message: t('instancePortRequired'), trigger: 'blur' }
  ],
  instanceStatus: [
    { required: true, message: t('instanceStatusRequired'), trigger: 'change' }
  ],
  healthStatus: [
    { required: true, message: t('healthStatusRequired'), trigger: 'change' }
  ],
  weightValue: [
    { required: true, type: 'number' as const, message: t('weightRequired'), trigger: 'blur' }
  ]
}

// 实例表格列定义
const instanceColumns = computed(() => [
  {
    title: t('columns.hostAddress'),
    key: 'hostAddress'
  },
  {
    title: t('columns.portNumber'),
    key: 'portNumber'
  },
  {
    title: t('columns.weightValue'),
    key: 'weightValue',
    width: 80,
    align: "center" as const,
    render: (row: ServiceInstance) => {
      return h('span', { 
        style: { fontWeight: '600' } 
      }, row.weightValue?.toString() || '1')
    }
  },
  {
    title: t('columns.instanceStatus'),
    key: 'instanceStatus',
    render: (row: ServiceInstance) => {
      const status = row.instanceStatus
      const colorMap: Record<string, string> = {
        'UP': 'success',
        'DOWN': 'error',
        'STARTING': 'warning',
        'OUT_OF_SERVICE': 'default'
      }
      return h(NTag, {
        type: colorMap[status] as any,
        size: 'small'
      }, { default: () => t(`status.${status}`) })
    }
  },
  {
    title: t('columns.healthStatus'),
    key: 'healthStatus',
    render: (row: ServiceInstance) => {
      const status = row.healthStatus
      const colorMap: Record<string, string> = {
        'HEALTHY': 'success',
        'UNHEALTHY': 'error',
        'UNKNOWN': 'default'
      }
      return h(NTag, {
        type: colorMap[status] as any,
        size: 'small'
      }, { default: () => t(`status.${status}`) })
    }
  },
  {
    title: t('columns.tempInstanceFlag'),
    key: 'tempInstanceFlag',
    render: (row: ServiceInstance) => {
      const isTemp = row.tempInstanceFlag === 'Y'
      return h(NTag, {
        type: isTemp ? 'warning' : 'success',
        size: 'small'
      }, { default: () => isTemp ? t('status.temporary') : t('status.permanent') })
    }
  },
  {
    title: t('columns.actions'),
    key: 'actions',
    width: 100,
    align: "center" as const,
    fixed: "right" as const,
    render: (row: ServiceInstance, index: number) => {
      return h(NButtonGroup, { size: 'small' }, {
        default: () => [
          h(NButton, {
            size: 'small',
            type: 'info',
            quaternary: true,
            onClick: () => handleEditInstance(index)
          }, {
            icon: () => h(NIcon, { component: CreateOutline })
          }),
          h(NPopconfirm, {
            onPositiveClick: () => handleDeleteInstance(index)
          }, {
            trigger: () => h(NButton, {
              size: 'small',
              type: 'error',
              quaternary: true
            }, {
              icon: () => h(NIcon, { component: TrashOutline })
            }),
            default: () => t('messages.confirmDeleteInstance')
          })
        ]
      })
    }
  }
])

// 点击添加实例按钮
const handleAddInstanceClick = () => {
  if (!isEdit.value) {
    // 如果是新建服务，提示用户需要先保存服务
    message.warning(t('saveServiceBeforeAddingInstance'))
    return
  }
  
  // 如果是编辑服务，正常添加实例
  handleAddInstance()
}

// 添加实例
const handleAddInstance = () => {
  isEditInstance.value = false
  currentInstanceIndex.value = -1
  
  // 重置表单数据
  Object.assign(instanceForm, {
    serviceInstanceId: `${formData.serviceName}-${Date.now()}`,
    hostAddress: '',
    portNumber: 8080,
    contextPath: formData.contextPath || '/',
    instanceStatus: 'UP',
    healthStatus: 'HEALTHY',
    weightValue: 1,
    clientType: 'OTHER',
    tempInstanceFlag: 'N'
  })
  
  instanceDrawerVisible.value = true
}

// 编辑实例
const handleEditInstance = (index: number) => {
  isEditInstance.value = true
  currentInstanceIndex.value = index
  
  const instance = formData.instances[index]
  Object.assign(instanceForm, { ...instance })
  
  instanceDrawerVisible.value = true
}

// 刷新实例列表
const handleRefreshInstances = async () => {
  if (isEdit.value && formData.serviceName) {
    await loadServiceInstances(formData.serviceName)
  }
}

// 删除实例
const handleDeleteInstance = async (index: number) => {
  const instance = formData.instances[index]
  
  // 如果是编辑模式并且有实例ID，尝试调用API删除
  if (isEdit.value && instance.serviceInstanceId) {
    try {
      const response = await deleteServiceInstance(instance.serviceInstanceId)
      if (response.oK) {
        formData.instances.splice(index, 1)
        message.success(t('deleteInstanceSuccess'))
      } else {
        // API调用失败，保持事务一致性，不更新本地状态
        message.error(response.errMsg || t('deleteInstanceFailed'))
      }
    } catch (error) {
      console.error('Failed to delete instance:', error)
      message.error(t('deleteInstanceFailed'))
    }
  } else {
    // 新服务的情况下，直接从本地删除
    formData.instances.splice(index, 1)
    message.success(t('deleteInstanceSuccess'))
  }
}

// 提交实例表单
const handleInstanceSubmit = async () => {
  try {
    await instanceFormRef.value?.validate()
    
    // 构造完整的实例对象
    const newInstance: ServiceInstance = {
      ...instanceForm as ServiceInstance,
      serviceName: formData.serviceName,
      serviceGroupId: formData.serviceGroupId,
      groupName: formData.groupName,
      tenantId: 'default', // 这个可能需要从其他地方获取
      registerTime: new Date().toISOString(),
      lastHeartbeatTime: new Date().toISOString(),
      activeFlag: 'Y',
      tempInstanceFlag: instanceForm.tempInstanceFlag || 'N' // 使用表单中的临时实例标记
    }
    
    if (isEditInstance.value && currentInstanceIndex.value >= 0) {
      // 更新现有实例
      // 如果在编辑模式下且服务已存在，调用API更新实例
      if (isEdit.value && formData.serviceName && instanceForm.serviceInstanceId) {
        try {
          const response = await updateServiceInstance(newInstance)
          if (response.oK) {
            formData.instances[currentInstanceIndex.value] = newInstance
            message.success(t('updateInstanceSuccess'))
            instanceDrawerVisible.value = false
          } else {
            // API调用失败，保持事务一致性，不更新本地状态
            message.error(response.errMsg || t('updateInstanceFailed'))
          }
        } catch (error) {
          console.error('Failed to update instance:', error)
          message.error(t('updateInstanceFailed'))
        }
      } else {
        // 新服务或无实例ID情况下，只更新本地
        formData.instances[currentInstanceIndex.value] = newInstance
        message.success(t('updateInstanceSuccess'))
        instanceDrawerVisible.value = false
      }
    } else {
      // 添加新实例
      // 如果是在编辑模式下且服务已存在，调用后端API创建实例
      if (isEdit.value && formData.serviceName) {
        try {
          const response = await createServiceInstance(newInstance)
          if (response.oK) {
            formData.instances.push(newInstance)
            message.success(t('addInstanceSuccess'))
            instanceDrawerVisible.value = false
          } else {
            // API调用失败，保持事务一致性，不更新本地状态
            message.error(response.errMsg || t('addInstanceFailed'))
          }
        } catch (error) {
          console.error('Failed to create instance:', error)
          message.error(t('addInstanceFailed'))
        }
      } else {
        // 新服务的情况下，只添加到本地实例列表
        formData.instances.push(newInstance)
        message.success(t('addInstanceSuccess'))
        instanceDrawerVisible.value = false
      }
    }
  } catch (error) {
    console.error('Instance validation error:', error)
  }
}
</script>

<style scoped lang="scss">
.service-form-page {
  min-height: 100vh;

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 0;

    .header-actions {
      display: flex;
      gap: 12px;
    }

    :deep(.n-breadcrumb-item__link) {
      cursor: pointer;
      
      &:hover {
        color: var(--primary-color);
      }
    }
  }

  .form-card {
    max-width: 1200px;
    margin: 0 auto;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);

    .form-actions {
      margin-top: 40px;
      padding: 24px 0;
      border-top: 1px solid var(--border-color);
    }

    .instance-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;
      
      .instance-header-left {
        display: flex;
        align-items: center;
      }
      
      .instance-header-actions {
        display: flex;
        gap: 8px;
      }
    }

    .instance-table {
      border: 1px solid var(--border-color);
      border-radius: 4px;
      overflow: hidden;
    }

    :deep(.n-form-item-label) {
      font-weight: 500;
    }

    :deep(.n-input__input-el) {
      font-size: 14px;
    }

    :deep(.n-select) {
      font-size: 14px;
    }
  }
}
</style>
