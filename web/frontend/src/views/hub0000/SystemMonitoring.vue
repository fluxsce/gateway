<template>
    <div class="system-monitoring">
        <!-- 服务器选择器 -->
        <div class="server-selector">
            <n-card title="服务器选择" embedded>
                <div class="selector-content">
                    <n-select v-model:value="selectedServerId" :options="serverOptions" placeholder="选择服务器"
                        :loading="serverListLoading" @update:value="handleServerChange" clearable />
                    <n-button @click="refreshAllData" :loading="operationLoading">
                        <template #icon>
                            <n-icon>
                                <ReloadOutlined />
                            </n-icon>
                        </template>
                        刷新数据
                    </n-button>
                </div>
            </n-card>
        </div>

        <!-- 服务器信息概览卡片 -->
        <div class="overview-cards" v-if="selectedServerInfo">
            <n-card title="服务器信息" embedded>
                <div class="overview-grid">
                    <div class="overview-item">
                        <div class="overview-icon hostname">
                            <n-icon size="24">
                                <DatabaseOutlined />
                            </n-icon>
                        </div>
                        <div class="overview-content">
                            <div class="overview-label">主机名</div>
                            <n-tooltip :show-arrow="false">
                                <template #trigger>
                                    <div class="overview-value text-truncate">{{ selectedServerInfo.hostname }}</div>
                                </template>
                                {{ selectedServerInfo.hostname }}
                            </n-tooltip>
                        </div>
                    </div>

                    <div class="overview-item">
                        <div class="overview-icon os">
                            <n-icon size="24">
                                <component :is="getOSIcon(selectedServerInfo.osType)" />
                            </n-icon>
                        </div>
                        <div class="overview-content">
                            <div class="overview-label">操作系统</div>
                            <n-tooltip :show-arrow="false">
                                <template #trigger>
                                    <div class="overview-value text-truncate">{{ selectedServerInfo.osType }}</div>
                                </template>
                                {{ selectedServerInfo.osType }}
                            </n-tooltip>
                        </div>
                    </div>

                    <div class="overview-item">
                        <div class="overview-icon version">
                            <n-icon size="24">
                                <AndroidOutlined />
                            </n-icon>
                        </div>
                        <div class="overview-content">
                            <div class="overview-label">系统版本</div>
                            <n-tooltip :show-arrow="false">
                                <template #trigger>
                                    <div class="overview-value text-truncate">{{
                                        getShortVersion(selectedServerInfo.osVersion) }}</div>
                                </template>
                                {{ selectedServerInfo.osVersion }}
                            </n-tooltip>
                        </div>
                    </div>

                    <div class="overview-item">
                        <div class="overview-icon architecture">
                            <n-icon size="24">
                                <DesktopOutlined />
                            </n-icon>
                        </div>
                        <div class="overview-content">
                            <div class="overview-label">系统架构</div>
                            <n-tooltip :show-arrow="false">
                                <template #trigger>
                                    <div class="overview-value text-truncate">{{ selectedServerInfo.architecture }}
                                    </div>
                                </template>
                                {{ selectedServerInfo.architecture }}
                            </n-tooltip>
                        </div>
                    </div>

                    <div class="overview-item">
                        <div class="overview-icon server-type">
                            <n-icon size="24">
                                <CloudServerOutlined />
                            </n-icon>
                        </div>
                        <div class="overview-content">
                            <div class="overview-label">服务器类型</div>
                            <n-tooltip :show-arrow="false">
                                <template #trigger>
                                    <div class="overview-value text-truncate">{{
                                        getServerTypeLabel(selectedServerInfo.serverType) }}</div>
                                </template>
                                {{ getServerTypeLabel(selectedServerInfo.serverType) }}
                            </n-tooltip>
                        </div>
                    </div>

                    <div class="overview-item">
                        <div class="overview-icon ip">
                            <n-icon size="24">
                                <GlobalOutlined />
                            </n-icon>
                        </div>
                        <div class="overview-content">
                            <div class="overview-label">IP地址</div>
                            <n-tooltip :show-arrow="false">
                                <template #trigger>
                                    <div class="overview-value text-truncate">{{ selectedServerInfo.ipAddress || 'N/A'
                                    }}</div>
                                </template>
                                {{ selectedServerInfo.ipAddress || 'N/A' }}
                            </n-tooltip>
                        </div>
                    </div>
                </div>
            </n-card>
        </div>

        <!-- 数据加载中状态 -->
        <div v-if="initialDataLoading" class="loading-container">
            <n-spin size="large" />
            <p>正在加载服务器监控数据，请稍候...</p>
        </div>

        <!-- 图表展示区域 - 只有在初始数据加载完成后才显示 -->
        <div v-if="!initialDataLoading && selectedServerId" class="charts-container">
            <!-- 第一行：CPU、内存使用率趋势 -->
            <div class="chart-row">
                <div class="chart-item">
                    <CpuMonitor :data="model.cpuMetrics.value" :loading="cpuLoading" :warning-threshold="80"
                        :danger-threshold="90" :cpu-detail-data="model.cpuMetrics.value" @refresh="refreshCpuData"
                        @time-range-change="handleCpuTimeRangeChange" />
                </div>

                <div class="chart-item">
                    <MemoryMonitor :data="model.memoryMetrics.value" :loading="memoryLoading" :warning-threshold="80"
                        :danger-threshold="90" :memory-detail-data="model.memoryMetrics.value"
                        @refresh="refreshMemoryData" @time-range-change="handleMemoryTimeRangeChange" />
                </div>
            </div>

            <!-- 第二行：磁盘使用率、磁盘IO监控 -->
            <div class="chart-row">
                <div class="chart-item">
                    <DiskMonitor :data="model.diskMetrics.value" :loading="diskLoading" :warning-threshold="80"
                        :danger-threshold="90" :disk-detail-data="model.diskMetrics.value" @refresh="refreshDiskData"
                        @time-range-change="handleDiskTimeRangeChange" />
                </div>

                <div class="chart-item">
                    <DiskIOMonitor :data="model.diskIOMetrics.value" :loading="diskIOLoading"
                        :disk-io-detail-data="model.diskIOMetrics.value" @refresh="refreshDiskIOData"
                        @time-range-change="handleDiskIOTimeRangeChange" />
                </div>
            </div>

            <!-- 第三行：网络流量监控、进程监控 -->
            <div class="chart-row">
                <div class="chart-item">
                    <NetworkMonitor :data="model.networkMetrics.value" :loading="networkLoading"
                        :network-detail-data="model.networkMetrics.value" upload-color="#ff4d4f"
                        download-color="#52c41a" @refresh="refreshNetworkData"
                        @time-range-change="handleNetworkTimeRangeChange" />
                </div>

                <div class="chart-item">
                    <ProcessMonitor :data="model.processMetrics.value" :loading="processLoading"
                        :process-detail-data="model.processMetrics.value" @refresh="refreshProcessData"
                        @time-range-change="handleProcessTimeRangeChange" />
                </div>
            </div>
        </div>

        <!-- 无数据提示 -->
        <div v-if="!initialDataLoading && !selectedServerId && model.serverList.value.length > 0"
            class="no-data-container">
            <n-empty description="请选择一个服务器查看监控数据" />
        </div>

        <!-- 无服务器提示 -->
        <div v-if="!initialDataLoading && model.serverList.value.length === 0" class="no-data-container">
            <n-empty description="暂无可用的服务器" />
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { NCard, NSelect, NButton, NIcon, NTooltip, useMessage, NSpin, NEmpty } from 'naive-ui'
import {
    DatabaseOutlined,
    DesktopOutlined,
    WindowsOutlined,
    AndroidOutlined,
    CloudServerOutlined,
    GlobalOutlined,
    ReloadOutlined,
    AppleOutlined,
} from '@vicons/antd'
import {
    CpuMonitor,
    MemoryMonitor,
    DiskMonitor,
    DiskIOMonitor,
    NetworkMonitor,
    ProcessMonitor,
} from './components/metrics'
import { useSystemMonitorModel } from './hooks/useSystemMonitorModel'
import { useSystemMonitorManagement } from './hooks/useSystemMonitorManagement'
import { formatDate } from '@/utils/format'
import type { ServerInfo } from './types'

const message = useMessage()

// 使用系统监控模型和管理
const model = useSystemMonitorModel()
const management = useSystemMonitorManagement(model)

// 组件状态
const selectedServerId = ref('')
const initialDataLoading = ref(true) // 初始数据加载状态

// 计算属性 - 选中服务器信息
const selectedServerInfo = computed<ServerInfo | null>(() => {
    if (!selectedServerId.value) return null
    return model.serverList.value.find(server => server.metricServerId === selectedServerId.value) || null
})

// 服务器类型标签转换
const getServerTypeLabel = (serverType?: string): string => {
    const typeMap: Record<string, string> = {
        'physical': '物理机',
        'virtual': '虚拟机',
        'unknown': '未知'
    }
    return typeMap[serverType || 'unknown'] || '未知'
}

// 根据操作系统类型获取图标
const getOSIcon = (osType: string) => {
    const osLower = osType.toLowerCase()
    if (osLower.includes('windows')) {
        return WindowsOutlined
    } else if (osLower.includes('linux')) {
        return AndroidOutlined  // 使用Android图标代表Linux
    } else if (osLower.includes('mac') || osLower.includes('darwin')) {
        return AppleOutlined
    } else {
        return DesktopOutlined
    }
}

// 获取简化的系统版本信息
const getShortVersion = (version: string): string => {
    if (!version) return 'N/A'

    // 针对Windows系统版本的特殊处理
    if (version.toLowerCase().includes('windows')) {
        // 提取关键信息：Windows 版本号
        const match = version.match(/Windows (\d+(?:\.\d+)?)/i)
        if (match) {
            const windowsVersion = match[1]
            // 如果有额外信息（如 Home, Pro等），也提取出来
            const editionMatch = version.match(/Windows \d+(?:\.\d+)?\s+(\w+)/i)
            if (editionMatch) {
                return `Windows ${windowsVersion} ${editionMatch[1]}`
            }
            return `Windows ${windowsVersion}`
        }
    }

    // 对于其他系统，如果版本信息太长，进行截断
    if (version.length > 20) {
        return version.substring(0, 17) + '...'
    }

    return version
}

// 服务器选项
const serverOptions = computed(() => {
    const options = [
        { label: '全部服务器', value: '' }
    ]

    return [
        ...options,
        ...model.serverList.value.map(server => ({
            label: `${server.hostname} (${server.ipAddress || 'N/A'})`,
            value: server.metricServerId,
        }))
    ]
})

// 时间范围变化处理
const handleCpuTimeRangeChange = async (timeRange: [number, number] | null) => {
    if (timeRange) {
        const [startTime, endTime] = timeRange
        model.updateQueryParams({
            startTime: formatDate(startTime, 'YYYY-MM-DD HH:mm:ss'),
            endTime: formatDate(endTime, 'YYYY-MM-DD HH:mm:ss')
        })
        await management.loadCPUMetrics(selectedServerId.value)
    }
}

const handleMemoryTimeRangeChange = async (timeRange: [number, number] | null) => {
    if (timeRange) {
        const [startTime, endTime] = timeRange
        model.updateQueryParams({
            startTime: formatDate(startTime, 'YYYY-MM-DD HH:mm:ss'),
            endTime: formatDate(endTime, 'YYYY-MM-DD HH:mm:ss')
        })
        await management.loadMemoryMetrics(selectedServerId.value)
    }
}

const handleDiskTimeRangeChange = async (timeRange: [number, number] | null) => {
    if (timeRange) {
        const [startTime, endTime] = timeRange
        model.updateQueryParams({
            startTime: formatDate(startTime, 'YYYY-MM-DD HH:mm:ss'),
            endTime: formatDate(endTime, 'YYYY-MM-DD HH:mm:ss')
        })
        await management.loadDiskMetrics(selectedServerId.value)
    }
}

const handleNetworkTimeRangeChange = async (timeRange: [number, number] | null) => {
    if (timeRange) {
        const [startTime, endTime] = timeRange
        model.updateQueryParams({
            startTime: formatDate(startTime, 'YYYY-MM-DD HH:mm:ss'),
            endTime: formatDate(endTime, 'YYYY-MM-DD HH:mm:ss')
        })
        await management.loadNetworkMetrics(selectedServerId.value)
    }
}

const handleDiskIOTimeRangeChange = async (timeRange: [number, number] | null) => {
    if (timeRange) {
        const [startTime, endTime] = timeRange
        model.updateQueryParams({
            startTime: formatDate(startTime, 'YYYY-MM-DD HH:mm:ss'),
            endTime: formatDate(endTime, 'YYYY-MM-DD HH:mm:ss')
        })
        await management.loadDiskIOMetrics(selectedServerId.value)
    }
}

const handleProcessTimeRangeChange = async (timeRange: [number, number] | null) => {
    if (timeRange) {
        const [startTime, endTime] = timeRange
        model.updateQueryParams({
            startTime: formatDate(startTime, 'YYYY-MM-DD HH:mm:ss'),
            endTime: formatDate(endTime, 'YYYY-MM-DD HH:mm:ss')
        })
        await management.loadProcessMetrics(selectedServerId.value)
    }
}

// 监听分页变化
watch(
    () => model.pagination.pagination.page,
    (newPage) => {
        model.queryParams.pageNum = newPage
        loadServerList()
    }
)

watch(
    () => model.pagination.pagination.pageSize,
    (newPageSize) => {
        model.queryParams.pageSize = newPageSize
        loadServerList()
    }
)

// 监听查询参数变化
watch(
    () => ({
        tenantId: model.queryParams.tenantId,
        hostname: model.queryParams.hostname,
        osType: model.queryParams.osType,
        serverType: model.queryParams.serverType,
        activeFlag: model.queryParams.activeFlag,
    }),
    () => {
        if (!initialDataLoading.value) {
            loadServerList()
        }
    },
    { deep: true }
)

// 加载服务器列表
const loadServerList = async () => {
    try {
        const servers = await management.loadServerList()

        // 如果当前没有选中的服务器，且有可用服务器，则选择第一个
        if (!selectedServerId.value && servers.length > 0) {
            selectedServerId.value = servers[0].metricServerId
            await loadServerMonitorData(servers[0].metricServerId)
        }
    } catch (error) {
        console.error('加载服务器列表失败', error)
        message.error('加载服务器列表失败')
    }
}

// 加载服务器监控数据
const loadServerMonitorData = async (serverId: string) => {
    try {
        // 更新查询参数
        model.updateQueryParams({
            metricServerId: serverId,
            startTime: model.queryParams.startTime,
            endTime: model.queryParams.endTime
        })

        // 加载监控数据
        await management.loadAllMetrics(serverId)
    } catch (error) {
        console.error('加载服务器监控数据失败', error)
        message.error('加载服务器监控数据失败')
    }
}

// 初始化数据
const initializeData = async () => {
    try {
        initialDataLoading.value = true

        // 设置默认的时间范围查询参数
        const now = new Date()
        const oneHourAgo = new Date(now.getTime() - 3600 * 1000)
        model.updateQueryParams({
            startTime: formatDate(oneHourAgo, 'YYYY-MM-DD HH:mm:ss'),
            endTime: formatDate(now, 'YYYY-MM-DD HH:mm:ss')
        })

        // 加载服务器列表
        await loadServerList()
    } catch (error) {
        console.error('初始化数据失败', error)
        message.error('加载监控数据失败，请刷新页面重试')
    } finally {
        initialDataLoading.value = false
    }
}

// 处理服务器选择变更
const handleServerChange = async (serverId: string) => {
    if (!serverId) return

    try {
        operationLoading.value = true
        // 清空旧的图表数据，防止新服务器无数据时显示旧数据
        model.clearAllMetrics()
        await loadServerMonitorData(serverId)
        message.success('服务器监控数据加载成功')
    } catch (error) {
        console.error('加载服务器监控数据失败', error)
        message.error('加载服务器监控数据失败')
    } finally {
        operationLoading.value = false
    }
}

// 刷新所有数据
const refreshAllData = async () => {
    if (!selectedServerId.value) {
        message.warning('请先选择一个服务器')
        return
    }

    try {
        operationLoading.value = true
        await loadServerMonitorData(selectedServerId.value)
        message.success('所有监控数据刷新成功')
    } catch (error) {
        console.error('刷新监控数据失败', error)
        message.error('刷新监控数据失败')
    } finally {
        operationLoading.value = false
    }
}


const refreshCpuData = async () => {
    await management.loadCPUMetrics(selectedServerId.value)
    message.success('CPU数据刷新成功')
}

const refreshMemoryData = async () => {
    await management.loadMemoryMetrics(selectedServerId.value)
    message.success('内存数据刷新成功')
}

const refreshDiskData = async () => {
    await management.loadDiskMetrics(selectedServerId.value)
    message.success('磁盘数据刷新成功')
}

const refreshNetworkData = async () => {
    await management.loadNetworkMetrics(selectedServerId.value)
    message.success('网络数据刷新成功')
}

const refreshDiskIOData = async () => {
    await management.loadDiskIOMetrics(selectedServerId.value)
    message.success('磁盘IO数据刷新成功')
}

const refreshProcessData = async () => {
    await management.loadProcessMetrics(selectedServerId.value)
    message.success('进程数据刷新成功')
}

// 监听服务器列表变化，自动选择第一个服务器并加载监控数据
watch(
    () => model.serverList.value,
    (newList) => {
        if (newList.length > 0 && !selectedServerId.value) {
            selectedServerId.value = newList[0].metricServerId
        }
    }
)

// 生命周期钩子
onMounted(() => {
    initializeData()
})

onUnmounted(() => {
    // 清理资源已经在 management 中处理
    console.log('SystemMonitoring组件已卸载')
})

// 解构需要的响应式数据 - 从model中解构
const {
    // 加载状态
    serverListLoading,
    cpuLoading,
    memoryLoading,
    diskLoading,
    diskIOLoading,
    networkLoading,
    processLoading,
} = model

// 解构需要的方法和状态 - 从management中解构
const {
    // 操作状态
    operationLoading,
} = management
</script>

<style lang="scss" scoped>
.system-monitoring {
    padding: 16px;

    .server-selector {
        margin-bottom: 16px;

        .selector-content {
            display: flex;
            gap: 12px;

            .n-select {
                flex: 1;
            }
        }
    }

    .overview-cards {
        margin-bottom: 16px;

        .overview-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
            gap: 16px;

            .overview-item {
                display: flex;
                align-items: center;
                gap: 12px;

                .overview-icon {
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    width: 40px;
                    height: 40px;
                    border-radius: 50%;
                    color: #fff;

                    &.hostname {
                        background-color: #1890ff;
                    }

                    &.os {
                        background-color: #52c41a;
                    }

                    &.version {
                        background-color: #fa8c16;
                    }

                    &.architecture {
                        background-color: #722ed1;
                    }

                    &.server-type {
                        background-color: #eb2f96;
                    }

                    &.ip {
                        background-color: #faad14;
                    }
                }

                .overview-content {
                    flex: 1;

                    .overview-label {
                        font-size: 12px;
                        color: #999;
                        margin-bottom: 4px;
                    }

                    .overview-value {
                        font-size: 14px;
                        font-weight: 500;
                    }

                    .text-truncate {
                        max-width: 180px;
                        white-space: nowrap;
                        overflow: hidden;
                        text-overflow: ellipsis;
                    }
                }
            }
        }
    }

    .charts-container {
        .chart-row {
            display: flex;
            gap: 16px;
            margin-bottom: 16px;

            @media (max-width: 1200px) {
                flex-direction: column;
            }

            .chart-item {
                flex: 1;
                min-height: 360px;
            }
        }
    }

    .loading-container,
    .no-data-container {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        min-height: 400px;

        p {
            margin-top: 16px;
            color: #666;
        }
    }
}
</style>