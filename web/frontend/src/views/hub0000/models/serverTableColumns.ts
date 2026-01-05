/**
 * 服务器信息表格列配置
 */

import { h } from 'vue'
import { NTag, NButton, NSpace, NTooltip, NIcon, NTime } from 'naive-ui'
import { formatBytes } from '@/utils/format'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import type { DataTableColumns } from 'naive-ui'
import type { ServerInfo, ServerStatus } from '../types'
import { ServerStatus as ServerStatusEnum } from '../types'

// 操作处理函数接口
export interface ServerActionHandlers {
  onView: (row: ServerInfo) => void
  onEdit: (row: ServerInfo) => void
  onDelete: (row: ServerInfo) => void
  onMonitor: (row: ServerInfo) => void
  onConnect: (row: ServerInfo) => void
}

/**
 * 获取服务器状态标签
 */
const getServerStatusTag = (status: ServerStatus) => {
  const { t } = useModuleI18n('hub0000')

  switch (status) {
    case ServerStatusEnum.ONLINE:
      return h(
        NTag,
        { type: 'success', size: 'small' },
        { default: () => t('server.status.online') },
      )
    case ServerStatusEnum.OFFLINE:
      return h(
        NTag,
        { type: 'error', size: 'small' },
        { default: () => t('server.status.offline') },
      )
    case ServerStatusEnum.WARNING:
      return h(
        NTag,
        { type: 'warning', size: 'small' },
        { default: () => t('server.status.warning') },
      )
    case ServerStatusEnum.CRITICAL:
      return h(
        NTag,
        { type: 'error', size: 'small' },
        { default: () => t('server.status.critical') },
      )
    default:
      return h(
        NTag,
        { type: 'default', size: 'small' },
        { default: () => t('server.status.unknown') },
      )
  }
}

/**
 * 获取服务器类型标签
 */
const getServerTypeTag = (type?: string) => {
  const { t } = useModuleI18n('hub0000')

  switch (type) {
    case 'physical':
      return h(
        NTag,
        { type: 'primary', size: 'small' },
        { default: () => t('server.type.physical') },
      )
    case 'virtual':
      return h(NTag, { type: 'info', size: 'small' }, { default: () => t('server.type.virtual') })
    default:
      return h(
        NTag,
        { type: 'default', size: 'small' },
        { default: () => t('server.type.unknown') },
      )
  }
}

/**
 * 创建服务器表格列配置
 */
export const createServerTableColumns = (
  handlers: ServerActionHandlers,
): DataTableColumns<ServerInfo> => {
  const { t } = useModuleI18n('hub0000')

  return [
    {
      type: 'selection',
      disabled: (row: ServerInfo) => row.activeFlag !== 'Y',
    },
    {
      title: t('server.hostname'),
      key: 'hostname',
      width: 150,
      fixed: 'left',
      sorter: true,
      render: (row) =>
        h(
          NTooltip,
          { trigger: 'hover' },
          {
            trigger: () => h('span', { class: 'font-medium' }, row.hostname),
            default: () => `${row.hostname} (${row.ipAddress || '-'})`,
          },
        ),
    },
    {
      title: t('server.status'),
      key: 'status',
      width: 100,
      render: (row) => {
        // 根据最后更新时间判断状态
        const lastUpdate = new Date(row.lastUpdateTime)
        const now = new Date()
        const diffMinutes = (now.getTime() - lastUpdate.getTime()) / 60000

        let status: ServerStatus
        if (diffMinutes > 10) {
          status = ServerStatusEnum.OFFLINE
        } else if (diffMinutes > 5) {
          status = ServerStatusEnum.WARNING
        } else {
          status = ServerStatusEnum.ONLINE
        }

        return getServerStatusTag(status)
      },
    },
    {
      title: t('server.osType'),
      key: 'osType',
      width: 120,
      render: (row) =>
        h(
          NTooltip,
          { trigger: 'hover' },
          {
            trigger: () => h('span', {}, row.osType),
            default: () => `${row.osType} ${row.osVersion}`,
          },
        ),
    },
    {
      title: t('server.serverType'),
      key: 'serverType',
      width: 100,
      render: (row) => getServerTypeTag(row.serverType),
    },
    {
      title: t('server.ipAddress'),
      key: 'ipAddress',
      width: 140,
      render: (row) => h('span', { class: 'font-mono' }, row.ipAddress || '-'),
    },
    {
      title: t('server.architecture'),
      key: 'architecture',
      width: 100,
      render: (row) => h('span', {}, row.architecture),
    },
    {
      title: t('server.bootTime'),
      key: 'bootTime',
      width: 160,
      render: (row) => h(NTime, { time: new Date(row.bootTime) }),
    },
    {
      title: t('server.lastUpdateTime'),
      key: 'lastUpdateTime',
      width: 160,
      sorter: true,
      render: (row) => h(NTime, { time: new Date(row.lastUpdateTime), type: 'relative' }),
    },
    {
      title: t('server.location'),
      key: 'serverLocation',
      width: 150,
      ellipsis: {
        tooltip: true,
      },
      render: (row) => row.serverLocation || '-',
    },
    {
      title: t('common.actions'),
      key: 'actions',
      width: 200,
      fixed: 'right',
      render: (row) =>
        h(
          NSpace,
          { size: 'small' },
          {
            default: () => [
              h(
                NButton,
                {
                  size: 'small',
                  type: 'primary',
                  ghost: true,
                  onClick: () => handlers.onMonitor(row),
                },
                { default: () => t('server.actions.monitor') },
              ),
              h(
                NButton,
                {
                  size: 'small',
                  type: 'info',
                  ghost: true,
                  onClick: () => handlers.onView(row),
                },
                { default: () => t('common.view') },
              ),
              h(
                NButton,
                {
                  size: 'small',
                  type: 'warning',
                  ghost: true,
                  onClick: () => handlers.onEdit(row),
                  disabled: row.activeFlag !== 'Y',
                },
                { default: () => t('common.edit') },
              ),
              h(
                NButton,
                {
                  size: 'small',
                  type: 'error',
                  ghost: true,
                  onClick: () => handlers.onDelete(row),
                  disabled: row.activeFlag !== 'Y',
                },
                { default: () => t('common.delete') },
              ),
            ],
          },
        ),
    },
  ]
}

/**
 * 创建服务器监控概览表格列配置
 */
export const createServerMonitorTableColumns = (): DataTableColumns<any> => {
  const { t } = useModuleI18n('hub0000')

  return [
    {
      title: t('server.hostname'),
      key: 'hostname',
      width: 150,
      fixed: 'left',
      render: (row) => h('span', { class: 'font-medium' }, row.hostname),
    },
    {
      title: t('monitor.cpu'),
      key: 'cpu',
      width: 120,
      render: (row) =>
        h(
          NTooltip,
          { trigger: 'hover' },
          {
            trigger: () =>
              h(
                NTag,
                {
                  type: row.cpu.usage > 80 ? 'error' : row.cpu.usage > 60 ? 'warning' : 'success',
                  size: 'small',
                },
                { default: () => `${row.cpu.usage.toFixed(1)}%` },
              ),
            default: () => `负载: ${row.cpu.loadAvg.join(', ')}`,
          },
        ),
    },
    {
      title: t('monitor.memory'),
      key: 'memory',
      width: 120,
      render: (row) =>
        h(
          NTooltip,
          { trigger: 'hover' },
          {
            trigger: () =>
              h(
                NTag,
                {
                  type:
                    row.memory.usagePercent > 80
                      ? 'error'
                      : row.memory.usagePercent > 60
                        ? 'warning'
                        : 'success',
                  size: 'small',
                },
                { default: () => `${row.memory.usagePercent.toFixed(1)}%` },
              ),
            default: () => `${formatBytes(row.memory.usage)} / ${formatBytes(row.memory.total)}`,
          },
        ),
    },
    {
      title: t('monitor.disk'),
      key: 'disk',
      width: 120,
      render: (row) =>
        h(
          NTooltip,
          { trigger: 'hover' },
          {
            trigger: () =>
              h(
                NTag,
                {
                  type: row.disk.usage > 80 ? 'error' : row.disk.usage > 60 ? 'warning' : 'success',
                  size: 'small',
                },
                { default: () => `${row.disk.usage.toFixed(1)}%` },
              ),
            default: () =>
              `${formatBytes(row.disk.totalSpace - row.disk.freeSpace)} / ${formatBytes(row.disk.totalSpace)}`,
          },
        ),
    },
    {
      title: t('monitor.network'),
      key: 'network',
      width: 140,
      render: (row) =>
        h(
          NTooltip,
          { trigger: 'hover' },
          {
            trigger: () =>
              h(
                'span',
                { class: 'font-mono text-sm' },
                `↑${formatBytes(row.network.sendRate)}/s ↓${formatBytes(row.network.receiveRate)}/s`,
              ),
            default: () => `总流量: ${formatBytes(row.network.totalBytes)}`,
          },
        ),
    },
    {
      title: t('monitor.processes'),
      key: 'processes',
      width: 100,
      render: (row) =>
        h(
          NTooltip,
          { trigger: 'hover' },
          {
            trigger: () => h('span', {}, row.processes.total),
            default: () =>
              `运行: ${row.processes.running}, 睡眠: ${row.processes.sleeping}, 僵尸: ${row.processes.zombie}`,
          },
        ),
    },
    {
      title: t('monitor.temperature'),
      key: 'temperature',
      width: 100,
      render: (row) => {
        if (!row.temperature) return '-'

        return h(
          NTag,
          {
            type:
              row.temperature.status === 'critical'
                ? 'error'
                : row.temperature.status === 'warning'
                  ? 'warning'
                  : 'success',
            size: 'small',
          },
          { default: () => `${row.temperature.value.toFixed(1)}°C` },
        )
      },
    },
    {
      title: t('monitor.lastUpdate'),
      key: 'timestamp',
      width: 160,
      render: (row) => h(NTime, { time: new Date(row.timestamp), type: 'relative' }),
    },
  ]
}
