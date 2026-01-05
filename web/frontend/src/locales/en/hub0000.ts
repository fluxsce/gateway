/**
 * Hub0000 System Monitoring Module Internationalization Config - English
 */

export default {
  systemMonitoring: {
    title: 'System Monitoring',
    description:
      'Real-time monitoring of server status, view system resource usage and performance metrics',
  },

  overview: {
    totalServers: 'Total Servers',
    onlineServers: 'Online Servers',
    offlineServers: 'Offline Servers',
    avgCpuUsage: 'Average CPU Usage',
    avgMemoryUsage: 'Average Memory Usage',
    avgDiskUsage: 'Average Disk Usage',
    criticalAlerts: 'Critical Alerts',
  },

  server: {
    hostname: 'Hostname',
    serverStatus: 'Status',
    osType: 'Operating System',
    serverType: 'Server Type',
    ipAddress: 'IP Address',
    architecture: 'Architecture',
    bootTime: 'Boot Time',
    lastUpdateTime: 'Last Update',
    location: 'Location',

    type: {
      physical: 'Physical',
      virtual: 'Virtual',
      unknown: 'Unknown',
    },

    status: {
      online: 'Online',
      offline: 'Offline',
      warning: 'Warning',
      critical: 'Critical',
      unknown: 'Unknown',
    },

    actions: {
      monitor: 'Monitor',
      connect: 'Connect',
    },
  },

  monitor: {
    cpu: 'CPU',
    memory: 'Memory',
    disk: 'Disk',
    network: 'Network',
    processes: 'Processes',
    temperature: 'Temperature',
    lastUpdate: 'Last Update',
    realtimeStatus: 'Real-time Status',
    autoRefresh: 'Auto Refresh',
    manualRefresh: 'Manual Refresh',
  },

  alerts: {
    title: 'System Alerts',
    level: 'Level',
    server: 'Server',
    type: 'Type',
    message: 'Message',
    value: 'Current Value',
    time: 'Time',
    alertStatus: 'Status',
    acknowledged: 'Acknowledged',
    pending: 'Pending',
    batchAcknowledge: 'Batch Acknowledge',
    acknowledgeSuccess: 'Alerts acknowledged successfully',

    levels: {
      low: 'Low',
      medium: 'Medium',
      high: 'High',
      critical: 'Critical',
    },

    types: {
      cpu: 'CPU',
      memory: 'Memory',
      disk: 'Disk',
      network: 'Network',
      process: 'Process',
      temperature: 'Temperature',
    },
  },

  buttons: {
    refresh: 'Refresh',
    export: 'Export',
    view: 'View',
    edit: 'Edit',
    delete: 'Delete',
    actions: 'Actions',
    refreshSuccess: 'Refresh successful',
  },

  timeRange: {
    hour: 'Last Hour',
    day: 'Last Day',
    week: 'Last Week',
    month: 'Last Month',
  },

  timeRangeShortcuts: {
    lastHour: 'Last Hour',
    last6Hours: 'Last 6 Hours',
    last12Hours: 'Last 12 Hours',
    last24Hours: 'Last 24 Hours',
    last7Days: 'Last 7 Days',
  },

  common: {
    selectTimeRange: 'Select Time Range',
    refresh: 'Refresh',
    noData: 'No Data',
    loading: 'Loading...',
    unit: {
      bytes: 'B',
      kilobytes: 'KB',
      megabytes: 'MB',
      gigabytes: 'GB',
      terabytes: 'TB',
      petabytes: 'PB',
    },
  },

  diskIO: {
    title: 'Disk I/O',
    detailTitle: 'Disk I/O Details',
    deviceName: 'Device Name',
    readBytes: 'Read Bytes',
    writeBytes: 'Write Bytes',
    readCount: 'Read Count',
    writeCount: 'Write Count',
    collectTime: 'Collection Time',
    read: 'Read',
    write: 'Write',
    noData: 'No Disk I/O Data',
  },

  cpu: {
    title: 'CPU Monitor',
    detailTitle: 'CPU Usage Details',
    serverId: 'Server ID',
    usage: 'CPU Usage',
    userUsage: 'User Usage',
    systemUsage: 'System Usage',
    idleUsage: 'Idle Rate',
    ioWaitUsage: 'IO Wait Rate',
    irqUsage: 'IRQ Rate',
    softIrqUsage: 'Soft IRQ Rate',
    coreCount: 'Physical Cores',
    logicalCount: 'Logical Cores',
    loadAvg1: '1 Min Load',
    loadAvg5: '5 Min Load',
    loadAvg15: '15 Min Load',
    collectTime: 'Collection Time',
    noData: 'No CPU Data',
  },

  process: {
    title: 'Process Monitor',
    detailTitle: 'Top Processes',
    processName: 'Process Name',
    processId: 'PID',
    cpuPercent: 'CPU Usage',
    memoryUsage: 'Memory Usage',
    memoryPercent: 'Memory %',
    threadCount: 'Threads',
    status: 'Status',
    collectTime: 'Collection Time',
    avgCpu: 'Avg CPU Usage',
    avgMemory: 'Avg Memory Usage',
    processCount: 'Process Count',
    noData: 'No Process Data',
  },

  memory: {
    title: 'Memory Monitor',
    usage: 'Memory Usage',
    details: 'Memory Details',
    total: 'Total Memory',
    used: 'Used Memory',
    available: 'Available Memory',
    free: 'Free Memory',
    cached: 'Cached Memory',
    buffers: 'Buffers',
    swap: 'Swap',
    swapTotal: 'Total Swap',
    swapUsed: 'Used Swap',
    swapFree: 'Free Swap',
    swapUsage: 'Swap Usage',
    warning: 'Warning',
    danger: 'Danger',
  },

  disk: {
    title: 'Disk Monitor',
    usage: 'Disk Usage',
    details: 'Disk Details',
    partitionCount: 'Partition Count',
    totalSpace: 'Total Space',
    usedSpace: 'Used Space',
    freeSpace: 'Free Space',
    partitionDetails: 'Partition Details',
    morePartitions: 'and {count} more partitions',
    warning: 'Warning',
    danger: 'Danger',
  },
}
