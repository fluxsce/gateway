/**
 * 域名访问控制配置列表页面级 Hook
 * - 组合 useDomainAccessConfigService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import { parseArrayField } from '@/utils/format'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import type { DomainAccessConfig } from './types'
import { useDomainAccessConfigService } from './useDomainAccessConfigService'

/**
 * 域名访问控制配置列表页面级 Hook
 * @param moduleId 模块ID（用于权限控制，必填）
 * @param gridRef Grid 组件引用（可选）
 * @param securityConfigId 安全配置ID（可选，用于新增时自动填充）
 */
export function useDomainAccessConfigPage(
  moduleId: Ref<string>,
  gridRef?: Ref<any> | any,
  securityConfigId?: Ref<string | undefined>,
  searchFormRef?: Ref<any> | any
) {
  const message = useMessage()

  // 业务服务（包含 model、增删改查等，传递 moduleId）
  const service = useDomainAccessConfigService(moduleId.value, securityConfigId, searchFormRef)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditConfig = ref<DomainAccessConfig | null>(null)

  /**
   * 将配置对象转换为表单数据格式（解析 JSON 字符串字段为数组）
   */
  const convertToFormData = (config: DomainAccessConfig): any => {
    return {
      ...config,
      whitelistDomains: parseArrayField(config.whitelistDomains as any),
      blacklistDomains: parseArrayField(config.blacklistDomains as any),
    }
  }

  /** 打开新增配置对话框 */
  const openAddDialog = (securityConfigId?: string) => {
    formDialogMode.value = 'create'
    // 如果有传入 securityConfigId，设置为初始数据
    currentEditConfig.value = securityConfigId ? { securityConfigId } as any : null
    formDialogVisible.value = true
  }

  /** 打开编辑配置对话框 */
  const openEditDialog = async (config: DomainAccessConfig) => {
    // 如果有 domainAccessConfigId，通过主键获取最新数据
    if (config.domainAccessConfigId) {
      try {
        const latestConfig = await service.getConfigDetail(config.domainAccessConfigId)
        if (latestConfig) {
          formDialogMode.value = 'edit'
          currentEditConfig.value = convertToFormData(latestConfig)
          formDialogVisible.value = true
          return
        }
      } catch (error) {
        // 获取失败时降级使用传入的配置数据
      }
    }
    // 降级：使用传入的配置数据
    formDialogMode.value = 'edit'
    currentEditConfig.value = convertToFormData(config)
    formDialogVisible.value = true
  }

  /** 关闭表单对话框 */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditConfig.value = null
  }
  
  /** 打开查看详情对话框 */
  const openViewDialog = async (config: DomainAccessConfig) => {
    // 如果有 domainAccessConfigId，通过主键获取最新数据
    if (config.domainAccessConfigId) {
      try {
        const latestConfig = await service.getConfigDetail(config.domainAccessConfigId)
        if (latestConfig) {
          formDialogMode.value = 'view'
          currentEditConfig.value = convertToFormData(latestConfig)
          formDialogVisible.value = true
          return
        }
      } catch (error) {
        // 获取失败时降级使用传入的配置数据
      }
    }
    // 降级：使用传入的配置数据
    formDialogMode.value = 'view'
    currentEditConfig.value = convertToFormData(config)
    formDialogVisible.value = true
  }

  /**
   * 处理搜索（接收 SearchForm 传递的表单数据）
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    await service.handleSearch(formData)
  }

  /** 提交表单（新增/编辑共用，由 GDataFormModal 收集表单数据后回调） */
  const handleFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) return

    // 查看模式下不执行提交
    if (formDialogMode.value === 'view') {
      return
    }

    // 将数组字段转换为 JSON 字符串（后端需要）
    const apiData: any = {
      ...formData,
      whitelistDomains: Array.isArray(formData.whitelistDomains) ? JSON.stringify(formData.whitelistDomains) : formData.whitelistDomains,
      blacklistDomains: Array.isArray(formData.blacklistDomains) ? JSON.stringify(formData.blacklistDomains) : formData.blacklistDomains,
    }

    if (formDialogMode.value === 'create') {
      // 新增模式 - 优先使用 formData 中的 securityConfigId，否则使用传入的 securityConfigId
      const finalSecurityConfigId = apiData.securityConfigId || securityConfigId?.value
      if (!finalSecurityConfigId) {
        message.error('安全配置ID不能为空')
        return
      }
      apiData.securityConfigId = finalSecurityConfigId
      const success = await service.addConfig(apiData as Partial<DomainAccessConfig> & { securityConfigId: string })
      if (success) {
        closeFormDialog()
      }
    } else if (formDialogMode.value === 'edit') {
      // 编辑模式
      if (!currentEditConfig.value) return
      // 合并当前配置ID和租户ID，确保更新的是正确的记录
      const updatedConfig = {
        ...currentEditConfig.value,
        ...apiData,
      } as Partial<DomainAccessConfig> & { securityConfigId: string; domainAccessConfigId: string; tenantId: string }
      const success = await service.editConfig(updatedConfig)
      if (success) {
        closeFormDialog()
      }
    }
  }

  /**
   * 工具栏按钮点击处理
   * @param key 按钮 key
   * @param formData 表单数据（可选，search 操作时会传递）
   */
  const handleToolbarClick = async (key: string, formData?: Record<string, any>) => {
    switch (key) {
      case 'add':
        // 新增时需要 securityConfigId
        const finalSecurityConfigId = securityConfigId?.value
        if (!finalSecurityConfigId) {
          message.warning('请先选择安全配置')
          return
        }
        // 直接打开新增对话框，传入 securityConfigId
        openAddDialog(finalSecurityConfigId)
        break

      case 'edit': {
        // 编辑当前高亮的行（点击选中的行）
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getCurrentRecord()
        if (!currentRow) {
          message.warning('请先点击选择要编辑的配置')
          return
        }
        openEditDialog(currentRow as DomainAccessConfig)
        break
      }

      case 'delete': {
        // 删除当前高亮的行
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getCurrentRecord()
        if (!currentRow) {
          message.warning('请先点击选择要删除的配置')
          return
        }
        await service.deleteConfig(currentRow as DomainAccessConfig)
        break
      }

      case 'search': {
        // 如果传递了表单数据，直接使用它进行查询
        await service.handleSearch(formData)
        break
      }
    }
  }

  /**
   * 右键菜单点击处理
   */
  const handleMenuClick = async ({ code, row }: { code: string; row?: DomainAccessConfig }) => {
    if (!row) return

    switch (code) {
      case 'view':
        openViewDialog(row)
        break

      case 'edit':
        openEditDialog(row)
        break

      case 'delete':
        await service.deleteConfig(row)
        break
    }
  }

  return {
    // 业务服务（包含 model 与增删改查）
    service,

    // 表单对话框（新增/编辑/查看共用）
    formDialogVisible,
    formDialogMode,
    currentEditConfig,
    openAddDialog,
    openEditDialog,
    openViewDialog,
    closeFormDialog,
    handleFormSubmit,

    // 事件处理器
    handleToolbarClick,
    handleMenuClick,
    handleSearch,
  }
}

