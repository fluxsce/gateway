/**
 * 服务实例表格列配置
 * 用于显示服务实例列表的表格配置
 */

import { h } from 'vue'
import { NTag, NButton, NButtonGroup, NIcon, NTooltip, NTime } from 'naive-ui'
import { 
  CheckmarkCircleOutline, CloseCircleOutline, TimeOutline, 
  RefreshOutline, ArrowUpOutline, ArrowDownOutline, HeartOutline 
} from '@vicons/ionicons5'
import type { DataTableColumns } from 'naive-ui'
import type { ServiceInstance, InstanceStatus, HealthStatus } from '../types'

export const createInstanceTableColumns = (
  t: (key: string) => string, 
  handleInstanceAction?: (action: string, instance: ServiceInstance) => void
): DataTableColumns<ServiceInstance> => [
  {
    title: t('columns.serviceInstanceId'),
    key: 'serviceInstanceId',
    width: 280,
    ellipsis: {
      tooltip: true
    },
    render: (row) => h('span', { 
      style: { fontFamily: 'monospace', fontSize: '13px' } 
    }, row.serviceInstanceId)
  },
  {
    title: t('columns.hostAddress'),
    key: 'hostAddress',
    width: 140,
    render: (row) => h('span', { 
      style: { fontFamily: 'monospace' } 
    }, row.hostAddress)
  },
  {
    title: t('columns.portNumber'),
    key: 'portNumber',
    width: 80,
    align: 'center',
    render: (row) => h('span', { 
      style: { fontFamily: 'monospace', fontWeight: 'bold' } 
    }, row.portNumber.toString())
  },
  {
    title: t('columns.instanceStatus'),
    key: 'instanceStatus',
    width: 100,
    align: 'center',
    render: (row) => {
      const statusConfig = getInstanceStatusConfig(row.instanceStatus)
      return h(NTag, {
        type: statusConfig.type,
        size: 'small',
        class: `status-${row.instanceStatus.toLowerCase()}`
      }, {
        default: () => t(`status.${row.instanceStatus}`),
        icon: () => h(NIcon, { size: 14, component: statusConfig.icon })
      })
    }
  },
  {
    title: t('columns.healthStatus'),
    key: 'healthStatus',
    width: 100,
    align: 'center',
    render: (row) => {
      const healthConfig = getHealthStatusConfig(row.healthStatus)
      return h(NTag, {
        type: healthConfig.type,
        size: 'small',
        class: `health-${row.healthStatus.toLowerCase()}`
      }, {
        default: () => t(`status.${row.healthStatus}`),
        icon: () => h(NIcon, { size: 14, component: healthConfig.icon })
      })
    }
  },
  {
    title: t('columns.weightValue'),
    key: 'weightValue',
    width: 80,
    align: 'center',
    render: (row) => h('span', { 
      style: { fontWeight: '600' } 
    }, row.weightValue.toString())
  },
  {
    title: t('columns.clientType'),
    key: 'clientType',
    width: 100,
    align: 'center',
    render: (row) => h(NTag, {
      type: 'info',
      size: 'small'
    }, { default: () => row.clientType })
  },
  {
    title: t('columns.tempInstanceFlag'),
    key: 'tempInstanceFlag',
    width: 100,
    align: 'center',
    render: (row) => {
      const isTemp = row.tempInstanceFlag === 'Y'
      return h(NTag, {
        type: isTemp ? 'warning' : 'success',
        size: 'small'
      }, { default: () => isTemp ? t('status.temporary') : t('status.permanent') })
    }
  },
  {
    title: t('columns.contextPath'),
    key: 'contextPath',
    width: 120,
    ellipsis: {
      tooltip: true
    },
    render: (row) => h('span', { 
      style: { fontFamily: 'monospace' } 
    }, row.contextPath || '/')
  },
  {
    title: t('columns.registerTime'),
    key: 'registerTime',
    width: 160,
    render: (row) => h(NTime, { time: new Date(row.registerTime) })
  },
  {
    title: t('columns.lastHeartbeatTime'),
    key: 'lastHeartbeatTime',
    width: 160,
    render: (row) => {
      if (!row.lastHeartbeatTime) {
        return h('span', { style: { color: 'var(--text-color-3)' } }, t('table.heartbeatTimeout'))
      }
      return h(NTime, { time: new Date(row.lastHeartbeatTime) })
    }
  },
  {
    title: t('columns.lastHealthCheckTime'),
    key: 'lastHealthCheckTime',
    width: 160,
    render: (row) => {
      if (!row.lastHealthCheckTime) {
        return h('span', { style: { color: 'var(--text-color-3)' } }, t('table.noHealthCheck'))
      }
      return h(NTime, { time: new Date(row.lastHealthCheckTime) })
    }
  },
  {
    title: t('columns.actions'),
    key: 'actions',
    width: 180,
    align: 'center',
    fixed: 'right',
    render: (row) => {
      return h(NButtonGroup, { size: 'small' }, {
        default: () => [
          // 健康检查按钮
          h(NTooltip, {
            trigger: 'hover'
          }, {
            trigger: () => h(NButton, {
              size: 'small',
              type: 'primary',
              quaternary: true,
              onClick: () => handleInstanceAction?.('health-check', row)
            }, {
              icon: () => h(NIcon, { component: HeartOutline })
            }),
            default: () => t('actions.healthCheck')
          }),
          // 上线按钮 - 仅当实例状态为 DOWN 时显示
          row.instanceStatus === 'DOWN' ? h(NTooltip, {
            trigger: 'hover'
          }, {
            trigger: () => h(NButton, {
              size: 'small',
              type: 'success',
              quaternary: true,
              onClick: () => handleInstanceAction?.('up', row)
            }, {
              icon: () => h(NIcon, { component: ArrowUpOutline })
            }),
            default: () => t('actions.bringUp')
          }) : null,
          // 下线按钮 - 仅当实例状态为 UP 时显示
          row.instanceStatus === 'UP' ? h(NTooltip, {
            trigger: 'hover'
          }, {
            trigger: () => h(NButton, {
              size: 'small',
              type: 'error',
              quaternary: true,
              onClick: () => handleInstanceAction?.('down', row)
            }, {
              icon: () => h(NIcon, { component: ArrowDownOutline })
            }),
            default: () => t('actions.takeDown')
          }) : null
        ].filter(Boolean) // 过滤掉 null 元素
      })
    }
  }
]

/**
 * 获取实例状态配置
 */
const getInstanceStatusConfig = (status: InstanceStatus) => {
  switch (status) {
    case 'UP':
      return {
        type: 'success' as const,
        text: 'UP',
        icon: CheckmarkCircleOutline
      }
    case 'DOWN':
      return {
        type: 'error' as const,
        text: 'DOWN',
        icon: CloseCircleOutline
      }
    case 'STARTING':
      return {
        type: 'warning' as const,
        text: 'STARTING',
        icon: TimeOutline
      }
    case 'OUT_OF_SERVICE':
      return {
        type: 'default' as const,
        text: 'OUT_OF_SERVICE',
        icon: CloseCircleOutline
      }
    default:
      return {
        type: 'default' as const,
        text: status,
        icon: TimeOutline
      }
  }
}

/**
 * 获取健康状态配置
 */
const getHealthStatusConfig = (status: HealthStatus) => {
  switch (status) {
    case 'HEALTHY':
      return {
        type: 'success' as const,
        text: 'HEALTHY',
        icon: CheckmarkCircleOutline
      }
    case 'UNHEALTHY':
      return {
        type: 'error' as const,
        text: 'UNHEALTHY',
        icon: CloseCircleOutline
      }
    case 'UNKNOWN':
      return {
        type: 'default' as const,
        text: 'UNKNOWN',
        icon: TimeOutline
      }
    default:
      return {
        type: 'default' as const,
        text: status,
        icon: TimeOutline
      }
  }
}


