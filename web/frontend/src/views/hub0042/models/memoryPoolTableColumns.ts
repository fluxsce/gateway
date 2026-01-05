/**
 * 内存池列表表格列定义
 */

import type { DataTableColumns } from 'naive-ui'
import { h } from 'vue'
import { NTag, NProgress } from 'naive-ui'
import type { MemoryPool } from '../types'

export const createMemoryPoolTableColumns = (t: (key: string) => string): DataTableColumns<MemoryPool> => {
  
  const formatMemorySize = (bytes: number): string => {
    if (bytes < 0) return 'N/A'
    if (bytes === 0) return '0 B'
    
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    
    return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
  }
  
  return [
    {
      title: t('memoryPoolName'),
      key: 'poolName',
      width: 200,
      ellipsis: {
        tooltip: true
      }
    },
    {
      title: t('memoryPoolType'),
      key: 'poolType',
      width: 100,
      render: (row) => {
        return h(NTag, {
          type: row.poolType === 'HEAP' ? 'success' : 'info',
          size: 'small'
        }, {
          default: () => row.poolType
        })
      }
    },
    {
      title: t('poolCategory'),
      key: 'poolCategory',
      width: 120
    },
    {
      title: t('currentUsage'),
      key: 'currentUsagePercent',
      width: 180,
      render: (row) => {
        const percent = row.currentUsagePercent
        let status: 'success' | 'warning' | 'error' = 'success'
        
        if (percent >= 90) status = 'error'
        else if (percent >= 75) status = 'warning'
        
        return h('div', [
          h(NProgress, {
            type: 'line',
            percentage: percent,
            status,
            height: 10,
            showIndicator: false
          }),
          h('div', { style: 'margin-top: 4px; font-size: 12px; color: #666;' }, 
            `${formatMemorySize(row.currentUsedBytes)} / ${formatMemorySize(row.currentMaxBytes)} (${percent.toFixed(2)}%)`
          )
        ])
      }
    },
    {
      title: t('currentCommitted'),
      key: 'currentCommittedBytes',
      width: 120,
      render: (row) => formatMemorySize(row.currentCommittedBytes)
    },
    {
      title: t('peakUsage'),
      key: 'peakUsagePercent',
      width: 100,
      render: (row) => {
        if (row.peakUsagePercent === undefined || row.peakUsagePercent === null) return '-'
        return `${row.peakUsagePercent.toFixed(2)}%`
      }
    },
    {
      title: t('peakUsedBytes'),
      key: 'peakUsedBytes',
      width: 120,
      render: (row) => {
        if (row.peakUsedBytes === undefined || row.peakUsedBytes === null) return '-'
        return formatMemorySize(row.peakUsedBytes)
      }
    },
    {
      title: t('healthStatus'),
      key: 'healthyFlag',
      width: 80,
      render: (row) => {
        const isHealthy = row.healthyFlag === 'Y'
        return h(NTag, {
          type: isHealthy ? 'success' : 'error',
          size: 'small'
        }, {
          default: () => isHealthy ? t('healthy') : t('unhealthy')
        })
      }
    },
    {
      title: t('collectionTime'),
      key: 'collectionTime',
      width: 160
    }
  ]
}

