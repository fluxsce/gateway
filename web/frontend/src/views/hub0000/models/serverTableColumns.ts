/**
 * 服务器信息表格列配置（RsTable / RsGrid 列定义）。
 */

import { h } from 'vue'
import { formatBytes, formatDate } from '@/utils/format'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { RsButton, RsTag, RsTooltip, type RsTableColumn } from '@/ui'
import type { ServerInfo, ServerStatus } from '../types'
import { ServerStatus as ServerStatusEnum } from '../types'

/**
 * 服务器行操作回调集合。
 */
export interface ServerActionHandlers {
  onView: (row: ServerInfo) => void
  onEdit: (row: ServerInfo) => void
  onDelete: (row: ServerInfo) => void
  onMonitor: (row: ServerInfo) => void
  onConnect: (row: ServerInfo) => void
}

/**
 * 将时间格式化为相对描述（如「3 分钟前」），用于最近更新列。
 */
function formatRelativeTime(input: Date | string | number): string {
  const ms = Date.now() - new Date(input).getTime()
  if (!Number.isFinite(ms)) return '-'
  const sec = Math.max(0, Math.floor(ms / 1000))
  if (sec < 60) return `${sec} 秒前`
  const min = Math.floor(sec / 60)
  if (min < 60) return `${min} 分钟前`
  const hour = Math.floor(min / 60)
  if (hour < 24) return `${hour} 小时前`
  const day = Math.floor(hour / 24)
  return `${day} 天前`
}

/**
 * 获取服务器状态标签。
 */
const getServerStatusTag = (status: ServerStatus) => {
  const { t } = useModuleI18n('hub0000')

  switch (status) {
    case ServerStatusEnum.ONLINE:
      return h(RsTag, { variant: 'success', size: 'sm' }, () => t('server.status.online'))
    case ServerStatusEnum.OFFLINE:
      return h(RsTag, { variant: 'danger', size: 'sm' }, () => t('server.status.offline'))
    case ServerStatusEnum.WARNING:
      return h(RsTag, { variant: 'warning', size: 'sm' }, () => t('server.status.warning'))
    case ServerStatusEnum.CRITICAL:
      return h(RsTag, { variant: 'danger', size: 'sm' }, () => t('server.status.critical'))
    default:
      return h(RsTag, { variant: 'default', size: 'sm' }, () => t('server.status.unknown'))
  }
}

/**
 * 获取服务器类型标签。
 */
const getServerTypeTag = (type?: string) => {
  const { t } = useModuleI18n('hub0000')

  switch (type) {
    case 'physical':
      return h(RsTag, { variant: 'primary', size: 'sm' }, () => t('server.type.physical'))
    case 'virtual':
      return h(RsTag, { variant: 'info', size: 'sm' }, () => t('server.type.virtual'))
    default:
      return h(RsTag, { variant: 'default', size: 'sm' }, () => t('server.type.unknown'))
  }
}

/**
 * 按使用率选择 Tag 语义色。
 */
const usageVariant = (usage: number): 'danger' | 'warning' | 'success' => {
  if (usage > 80) return 'danger'
  if (usage > 60) return 'warning'
  return 'success'
}

/**
 * 创建服务器表格列配置。
 * 行勾选请在 RsTable / RsGrid 上开启 selectable，不再使用独立 selection 列。
 */
export const createServerTableColumns = (
  handlers: ServerActionHandlers,
): RsTableColumn<ServerInfo>[] => {
  const { t } = useModuleI18n('hub0000')

  return [
    {
      title: t('server.hostname'),
      key: 'hostname',
      width: 150,
      fixed: 'left',
      sortable: true,
      render: (row) =>
        h(
          RsTooltip,
          { content: `${row.hostname} (${row.ipAddress || '-'})` },
          { default: () => h('span', { class: 'font-medium' }, row.hostname) },
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
          RsTooltip,
          { content: `${row.osType} ${row.osVersion}` },
          { default: () => h('span', {}, row.osType) },
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
      render: (row) => formatDate(row.bootTime),
    },
    {
      title: t('server.lastUpdateTime'),
      key: 'lastUpdateTime',
      width: 160,
      sortable: true,
      render: (row) => formatRelativeTime(row.lastUpdateTime),
    },
    {
      title: t('server.location'),
      key: 'serverLocation',
      width: 150,
      ellipsis: true,
      render: (row) => row.serverLocation || '-',
    },
    {
      title: t('common.actions'),
      key: 'actions',
      width: 200,
      fixed: 'right',
      render: (row) =>
        h(
          'div',
          { style: { display: 'flex', gap: '4px', flexWrap: 'wrap' } },
          [
            h(
              RsButton,
              {
                size: 'sm',
                variant: 'ghost',
                tone: 'primary',
                onClick: () => handlers.onMonitor(row),
              },
              () => t('server.actions.monitor'),
            ),
            h(
              RsButton,
              {
                size: 'sm',
                variant: 'ghost',
                tone: 'info',
                onClick: () => handlers.onView(row),
              },
              () => t('common.view'),
            ),
            h(
              RsButton,
              {
                size: 'sm',
                variant: 'ghost',
                tone: 'warning',
                disabled: row.activeFlag !== 'Y',
                onClick: () => handlers.onEdit(row),
              },
              () => t('common.edit'),
            ),
            h(
              RsButton,
              {
                size: 'sm',
                variant: 'ghost',
                tone: 'danger',
                disabled: row.activeFlag !== 'Y',
                onClick: () => handlers.onDelete(row),
              },
              () => t('common.delete'),
            ),
          ],
        ),
    },
  ]
}

/**
 * 创建服务器监控概览表格列配置。
 */
export const createServerMonitorTableColumns = (): RsTableColumn<Record<string, any>>[] => {
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
          RsTooltip,
          { content: `负载: ${row.cpu.loadAvg.join(', ')}` },
          {
            default: () =>
              h(
                RsTag,
                { variant: usageVariant(row.cpu.usage), size: 'sm' },
                () => `${row.cpu.usage.toFixed(1)}%`,
              ),
          },
        ),
    },
    {
      title: t('monitor.memory'),
      key: 'memory',
      width: 120,
      render: (row) =>
        h(
          RsTooltip,
          {
            content: `${formatBytes(row.memory.usage)} / ${formatBytes(row.memory.total)}`,
          },
          {
            default: () =>
              h(
                RsTag,
                { variant: usageVariant(row.memory.usagePercent), size: 'sm' },
                () => `${row.memory.usagePercent.toFixed(1)}%`,
              ),
          },
        ),
    },
    {
      title: t('monitor.disk'),
      key: 'disk',
      width: 120,
      render: (row) =>
        h(
          RsTooltip,
          {
            content: `${formatBytes(row.disk.totalSpace - row.disk.freeSpace)} / ${formatBytes(row.disk.totalSpace)}`,
          },
          {
            default: () =>
              h(
                RsTag,
                { variant: usageVariant(row.disk.usage), size: 'sm' },
                () => `${row.disk.usage.toFixed(1)}%`,
              ),
          },
        ),
    },
    {
      title: t('monitor.network'),
      key: 'network',
      width: 140,
      render: (row) =>
        h(
          RsTooltip,
          { content: `总流量: ${formatBytes(row.network.totalBytes)}` },
          {
            default: () =>
              h(
                'span',
                { class: 'font-mono text-sm' },
                `↑${formatBytes(row.network.sendRate)}/s ↓${formatBytes(row.network.receiveRate)}/s`,
              ),
          },
        ),
    },
    {
      title: t('monitor.processes'),
      key: 'processes',
      width: 100,
      render: (row) =>
        h(
          RsTooltip,
          {
            content: `运行: ${row.processes.running}, 睡眠: ${row.processes.sleeping}, 僵尸: ${row.processes.zombie}`,
          },
          { default: () => h('span', {}, row.processes.total) },
        ),
    },
    {
      title: t('monitor.temperature'),
      key: 'temperature',
      width: 100,
      render: (row) => {
        if (!row.temperature) return '-'

        const variant =
          row.temperature.status === 'critical'
            ? 'danger'
            : row.temperature.status === 'warning'
              ? 'warning'
              : 'success'

        return h(
          RsTag,
          { variant, size: 'sm' },
          () => `${row.temperature.value.toFixed(1)}°C`,
        )
      },
    },
    {
      title: t('monitor.lastUpdate'),
      key: 'timestamp',
      width: 160,
      render: (row) => formatRelativeTime(row.timestamp),
    },
  ]
}
