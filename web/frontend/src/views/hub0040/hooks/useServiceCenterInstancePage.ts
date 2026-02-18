/**
 * 服务中心实例管理页面级 Hook
 * - 组合 useServiceCenterInstanceService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import { useGDialog } from '@/components/gdialog'
import { flattenExtProperty, unflattenExtProperty } from '@/utils/format'
import { PlayCircleOutline, RefreshOutline, StopCircleOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import type { ServiceCenterInstance } from '../types'
import { useServiceCenterInstanceService } from './useServiceCenterInstanceService'

/**
 * 服务中心实例管理页面级 Hook
 */
export function useServiceCenterInstancePage(gridRef?: Ref<any> | any, searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 业务服务（包含 model、增删改查等）
  const service = useServiceCenterInstanceService(searchFormRef)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditInstance = ref<ServiceCenterInstance | null>(null)
  const submitting = ref(false)

  /**
   * 打开新增实例对话框
   */
  const openAddDialog = () => {
    formDialogMode.value = 'create'
    currentEditInstance.value = null
    formDialogVisible.value = true
  }

  /**
   * 打开编辑实例对话框
   */
  const openEditDialog = async (instance: ServiceCenterInstance) => {
    try {
      // 获取完整详情
      const detailInstance = await service.getInstanceDetail(
        instance.instanceName,
        instance.environment
      )
      
      if (!detailInstance) {
        message.error('获取实例详情失败')
        return
      }

      // 将证书和私钥内容转换为文件列表格式（用于回显）
      const formData: any = { ...detailInstance }
      
      // 如果有证书内容，转换为文件列表格式
      if (detailInstance.certContent) {
        formData.certFileList = [{
          id: 'cert-file',
          name: detailInstance.certFilePath || 'certificate.pem',
          content: detailInstance.certContent,
          status: 'finished',
        }]
      } else {
        formData.certFileList = []
      }

      // 如果有私钥内容，转换为文件列表格式
      if (detailInstance.keyContent) {
        formData.keyFileList = [{
          id: 'key-file',
          name: detailInstance.keyFilePath || 'private-key.pem',
          content: detailInstance.keyContent,
          status: 'finished',
        }]
      } else {
        formData.keyFileList = []
      }

      // 处理 IP 白名单和黑名单（JSON 字符串转数组）
      if (formData.ipWhitelist && typeof formData.ipWhitelist === 'string') {
        try {
          formData.ipWhitelist = JSON.parse(formData.ipWhitelist)
        } catch {
          formData.ipWhitelist = []
        }
      }
      if (formData.ipBlacklist && typeof formData.ipBlacklist === 'string') {
        try {
          formData.ipBlacklist = JSON.parse(formData.ipBlacklist)
        } catch {
          formData.ipBlacklist = []
        }
      }

      // 展开 extProperty 为扁平字段
      flattenExtProperty(formData)

      formDialogMode.value = 'edit'
      currentEditInstance.value = formData
      formDialogVisible.value = true
    } catch (error) {
      message.error('获取实例详情失败')
    }
  }

  /**
   * 关闭表单对话框
   */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditInstance.value = null
  }
  
  /**
   * 打开查看详情对话框
   * 从后端获取最新数据，确保显示的是最新状态
   */
  const openViewDialog = async (instance: ServiceCenterInstance) => {
    try {
      // 从后端获取最新数据，确保显示的是最新状态
      const detailInstance = await service.getInstanceDetail(
        instance.instanceName,
        instance.environment
      )
      
      if (!detailInstance) {
        message.error('获取实例详情失败')
        return
      }

      // 将证书和私钥内容转换为文件列表格式（用于回显）
      const formData: any = { ...detailInstance }
      
      // 如果有证书内容，转换为文件列表格式
      if (detailInstance.certContent) {
        formData.certFileList = [{
          id: 'cert-file',
          name: detailInstance.certFilePath || 'certificate.pem',
          content: detailInstance.certContent,
          status: 'finished',
        }]
      } else {
        formData.certFileList = []
      }

      // 如果有私钥内容，转换为文件列表格式
      if (detailInstance.keyContent) {
        formData.keyFileList = [{
          id: 'key-file',
          name: detailInstance.keyFilePath || 'private-key.pem',
          content: detailInstance.keyContent,
          status: 'finished',
        }]
      } else {
        formData.keyFileList = []
      }

      // 处理 IP 白名单和黑名单（JSON 字符串转数组）
      if (formData.ipWhitelist && typeof formData.ipWhitelist === 'string') {
        try {
          formData.ipWhitelist = JSON.parse(formData.ipWhitelist)
        } catch {
          formData.ipWhitelist = []
        }
      }
      if (formData.ipBlacklist && typeof formData.ipBlacklist === 'string') {
        try {
          formData.ipBlacklist = JSON.parse(formData.ipBlacklist)
        } catch {
          formData.ipBlacklist = []
        }
      }

      // 展开 extProperty 为扁平字段
      flattenExtProperty(formData)

      formDialogMode.value = 'view'
      currentEditInstance.value = formData
      formDialogVisible.value = true
    } catch (error) {
      message.error('获取实例详情失败')
    }
  }

  /**
   * 处理搜索（接收 SearchForm 传递的表单数据）
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    await service.handleSearch(formData)
  }

  /**
   * 提交表单（新增/编辑共用，由 GdataFormModal 收集表单数据后回调）
   */
  const handleFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) return

    // 查看模式下不执行提交
    if (formDialogMode.value === 'view') {
      return
    }

    submitting.value = true
    try {
      // 处理文件内容：将文件列表中的 content 和 name 提取到对应的字段
      const processedData = { ...formData }
      
      // 处理证书文件：从 certFileList 中提取 content 和 name
      if (processedData.certFileList && Array.isArray(processedData.certFileList) && processedData.certFileList.length > 0) {
        const certFile = processedData.certFileList[0]
        if (certFile) {
          if (certFile.content) {
            processedData.certContent = certFile.content
          }
          // 提取文件名到 certFilePath（如果用户上传了新文件，使用新文件名）
          if (certFile.name) {
            processedData.certFilePath = certFile.name
          }
        }
      } else if (formDialogMode.value === 'edit' && currentEditInstance.value) {
        // 编辑模式下，如果用户没有上传新文件，保留原有的文件名
        if (currentEditInstance.value.certFilePath && !processedData.certFilePath) {
          processedData.certFilePath = currentEditInstance.value.certFilePath
        }
      }
      // 删除 certFileList，不提交给后端
      delete processedData.certFileList

      // 处理私钥文件：从 keyFileList 中提取 content 和 name
      if (processedData.keyFileList && Array.isArray(processedData.keyFileList) && processedData.keyFileList.length > 0) {
        const keyFile = processedData.keyFileList[0]
        if (keyFile) {
          if (keyFile.content) {
            processedData.keyContent = keyFile.content
          }
          // 提取文件名到 keyFilePath（如果用户上传了新文件，使用新文件名）
          if (keyFile.name) {
            processedData.keyFilePath = keyFile.name
          }
        }
      } else if (formDialogMode.value === 'edit' && currentEditInstance.value) {
        // 编辑模式下，如果用户没有上传新文件，保留原有的文件名
        if (currentEditInstance.value.keyFilePath && !processedData.keyFilePath) {
          processedData.keyFilePath = currentEditInstance.value.keyFilePath
        }
      }
      // 删除 keyFileList，不提交给后端
      delete processedData.keyFileList

      // 处理 IP 白名单和黑名单（数组转 JSON 字符串）
      if (Array.isArray(processedData.ipWhitelist)) {
        processedData.ipWhitelist = processedData.ipWhitelist.length > 0
          ? JSON.stringify(processedData.ipWhitelist)
          : ''
      }
      if (Array.isArray(processedData.ipBlacklist)) {
        processedData.ipBlacklist = processedData.ipBlacklist.length > 0
          ? JSON.stringify(processedData.ipBlacklist)
          : ''
      }

      // 将扁平字段打包回 extProperty JSON 字符串
      unflattenExtProperty(processedData)

      if (formDialogMode.value === 'create') {
        // 新增模式
        const success = await service.addInstance(processedData as ServiceCenterInstance)
        if (success) {
          closeFormDialog()
        }
      } else if (formDialogMode.value === 'edit') {
        // 编辑模式
        if (!currentEditInstance.value) return
        const success = await service.editInstance({
          ...processedData,
          instanceName: currentEditInstance.value.instanceName,
          environment: currentEditInstance.value.environment,
        } as Partial<ServiceCenterInstance> & { instanceName: string; environment: string })
        if (success) {
          closeFormDialog()
        }
      }
    } finally {
      submitting.value = false
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
        // 直接打开新增对话框
        openAddDialog()
        break

      case 'edit': {
        // 编辑当前高亮的行（点击选中的行）
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getCurrentRecord()
        if (!currentRow) {
          message.warning('请先点击选择要编辑的实例')
          return
        }
        await openEditDialog(currentRow as ServiceCenterInstance)
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
          message.warning('请先点击选择要删除的实例')
          return
        }
        await service.deleteInstance(currentRow as ServiceCenterInstance)
        break
      }

      case 'search': {
        // 如果传递了表单数据，直接使用它进行查询
        // formData 参数在 SearchForm 的 handleToolbarClick 中传递
        await service.handleSearch(formData)
        break
      }
    }
  }

  /**
   * 处理启动实例
   */
  const handleStartInstance = async (instance: ServiceCenterInstance) => {
    const confirmed = await gDialog.warning({
      title: '确认启动',
      subtitle: '启动后将开始处理请求',
      content: `确定要启动实例"${instance.instanceName}" (${instance.environment}) 吗？`,
      icon: PlayCircleOutline,
      headerStyle: 'gradient',
      positiveText: '确定启动',
      negativeText: '取消',
      width: 500
    })
    
    if (confirmed) {
      service.startInstance(instance)
    }
  }

  /**
   * 处理停止实例
   */
  const handleStopInstance = async (instance: ServiceCenterInstance) => {
    const confirmed = await gDialog.warning({
      title: '确认停止',
      subtitle: '停止后将无法处理请求',
      content: `确定要停止实例"${instance.instanceName}" (${instance.environment}) 吗？`,
      icon: StopCircleOutline,
      headerStyle: 'gradient',
      positiveText: '确定停止',
      negativeText: '取消',
      width: 500
    })
    
    if (confirmed) {
      service.stopInstance(instance)
    }
  }

  /**
   * 处理配置重载
   */
  const handleReloadInstance = async (instance: ServiceCenterInstance) => {
    const confirmed = await gDialog.warning({
      title: '确认重载配置',
      subtitle: '重载将重新加载配置',
      content: `确定要对实例"${instance.instanceName}" (${instance.environment}) 执行配置重载操作吗？`,
      icon: RefreshOutline,
      headerStyle: 'gradient',
      positiveText: '确定重载',
      negativeText: '取消',
      width: 500
    })
    
    if (confirmed) {
      service.reloadInstance(instance)
    }
  }

  /**
   * 右键菜单点击处理
   */
  const handleMenuClick = async ({ code, row }: { code: string; row?: ServiceCenterInstance }) => {
    if (!row) return

    switch (code) {
      case 'view':
        await openViewDialog(row)
        break

      case 'edit':
        await openEditDialog(row)
        break

      case 'delete':
        await service.deleteInstance(row)
        break

      case 'start':
        handleStartInstance(row)
        break

      case 'stop':
        handleStopInstance(row)
        break

      case 'reload':
        handleReloadInstance(row)
        break
    }
  }

  return {
    // 业务服务（包含 model 与增删改查）
    service,

    // 表单对话框（新增/编辑/查看共用）
    formDialogVisible,
    formDialogMode,
    currentEditInstance,
    submitting,
    openAddDialog,
    openEditDialog,
    openViewDialog,
    handleFormSubmit,

    // 事件处理器
    handleToolbarClick,
    handleMenuClick,
    handleSearch,
    handleStartInstance,
    handleStopInstance,
    handleReloadInstance,
  }
}

export type ServiceCenterInstancePage = ReturnType<typeof useServiceCenterInstancePage>

