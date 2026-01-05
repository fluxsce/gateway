/**
 * JVM资源列表表格列定义
 */

import type { DataTableColumns } from 'naive-ui'
import { h } from 'vue'
import { NTag, NSpace, NButton } from 'naive-ui'
import { formatDate } from '@/utils/format'
import type { JvmResource } from '../types'

export const createJvmResourceTableColumns = (t: (key: string) => string): DataTableColumns<JvmResource> => {
  return [
    {
      title: t('jvmResourceId'),
      key: 'jvmResourceId',
      width: 200,
      ellipsis: {
        tooltip: true
      }
    },
    {
      title: t('applicationName'),
      key: 'applicationName',
      width: 150,
      ellipsis: {
        tooltip: true
      }
    },
    {
      title: t('groupName'),
      key: 'groupName',
      width: 120,
      ellipsis: {
        tooltip: true
      }
    },
    {
      title: t('hostName'),
      key: 'hostName',
      width: 150,
      ellipsis: {
        tooltip: true
      },
      render: (row) => row.hostName || '-'
    },
    {
      title: t('hostIpAddress'),
      key: 'hostIpAddress',
      width: 140,
      ellipsis: {
        tooltip: true
      },
      render: (row) => row.hostIpAddress || '-'
    },
    {
      title: t('healthStatus'),
      key: 'healthyFlag',
      width: 120,
      render: (row) => {
        const isHealthy = row.healthyFlag === 'Y'
        const needsAttention = row.requiresAttentionFlag === 'Y'
        
        return h(NSpace, { size: 4 }, {
          default: () => [
            h(NTag, {
              type: isHealthy ? 'success' : 'error',
              size: 'small'
            }, {
              default: () => isHealthy ? t('healthy') : t('unhealthy')
            }),
            needsAttention && h(NTag, {
              type: 'warning',
              size: 'small'
            }, {
              default: () => t('attention')
            })
          ]
        })
      }
    },
    {
      title: t('healthGrade'),
      key: 'healthGrade',
      width: 100,
      render: (row) => {
        if (!row.healthGrade) return '-'
        
        const gradeMap: Record<string, { type: 'success' | 'info' | 'warning' | 'error', text: string }> = {
          EXCELLENT: { type: 'success', text: t('excellent') },
          GOOD: { type: 'success', text: t('good') },
          FAIR: { type: 'warning', text: t('fair') },
          POOR: { type: 'error', text: t('poor') },
          CRITICAL: { type: 'error', text: t('critical') }
        }
        
        const grade = gradeMap[row.healthGrade] || { type: 'info', text: row.healthGrade }
        
        return h(NTag, {
          type: grade.type,
          size: 'small'
        }, {
          default: () => grade.text
        })
      }
    },
    {
      title: t('jvmUptime'),
      key: 'jvmUptimeMs',
      width: 130,
      render: (row) => {
        const ms = row.jvmUptimeMs
        const hours = Math.floor(ms / (1000 * 60 * 60))
        const minutes = Math.floor((ms % (1000 * 60 * 60)) / (1000 * 60))
        
        if (hours > 24) {
          const days = Math.floor(hours / 24)
          const remainHours = hours % 24
          return `${days}${t('day')} ${remainHours}${t('hour')}`
        }
        
        return `${hours}${t('hour')} ${minutes}${t('minute')}`
      }
    },
    {
      title: t('jvmStartTime'),
      key: 'jvmStartTime',
      width: 180,
      ellipsis: {
        tooltip: true
      },
      render: (row) => {
        if (!row.jvmStartTime) return '-'
        try {
          return formatDate(row.jvmStartTime, 'YYYY-MM-DD HH:mm:ss')
        } catch (error) {
          return row.jvmStartTime
        }
      }
    },
    {
      title: t('collectionTime'),
      key: 'collectionTime',
      width: 180,
      ellipsis: {
        tooltip: true
      },
      render: (row) => {
        if (!row.collectionTime) return '-'
        try {
          return formatDate(row.collectionTime, 'YYYY-MM-DD HH:mm:ss')
        } catch (error) {
          return row.collectionTime
        }
      }
    },
    {
      title: t('summaryText'),
      key: 'summaryText',
      width: 200,
      ellipsis: {
        tooltip: true
      }
    },
    {
      title: t('actions'),
      key: 'actions',
      width: 150,
      fixed: 'right',
      render: (row, index) => {
        return h(NSpace, { size: 4 }, {
          default: () => []
        })
      }
    }
  ]
}

