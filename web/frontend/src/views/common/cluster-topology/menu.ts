import type { RsGridMenuItem } from '@/components/rs-grid'

/** 右键项。页面模块还要有按钮 `{模块号}:clusterTopology`，并在 clusterstatus.ButtonCodes 里登记。 */
export const clusterTopologyMenuItem: RsGridMenuItem = {
  key: 'clusterTopology',
  label: '集群健康与拓扑',
  icon: 'network',
  requireRow: true,
}
