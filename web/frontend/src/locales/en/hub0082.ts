/**
 * Alert Log Management Module English locale
 * hub0082 - Alert Log Management
 */
export default {
  moduleName: 'Alert Log Management',

  common: {
    all: 'All',
    delete: 'Delete',
    cancel: 'Cancel',
    viewDetail: 'View Details',
  },

  level: {
    info: 'Info',
    warn: 'Warning',
    error: 'Error',
    critical: 'Critical',
  },

  sendStatus: {
    pending: 'Pending',
    sending: 'Sending',
    success: 'Success',
    failed: 'Failed',
  },

  shortcuts: {
    today: 'Today',
    yesterday: 'Yesterday',
    lastHour: 'Last 1 Hour',
    last6Hours: 'Last 6 Hours',
    last24Hours: 'Last 24 Hours',
    last7Days: 'Last 7 Days',
  },

  search: {
    timeRange: 'Time Range',
    timeRangePlaceholder: 'Please select a time range',
    timeRangeRequired: 'Please select a time range',
    alertLogId: 'Log ID',
    alertLogIdPlaceholder: 'Please enter log ID',
    alertLevel: 'Alert Level',
    alertLevelPlaceholder: 'Please select alert level',
    alertType: 'Alert Type',
    alertTypePlaceholder: 'Please enter alert type',
    alertTitle: 'Alert Title',
    alertTitlePlaceholder: 'Please enter alert title',
    channelName: 'Channel Name',
    channelNamePlaceholder: 'Please enter channel name',
    sendStatus: 'Send Status',
    sendStatusPlaceholder: 'Please select send status',
  },

  toolbar: {
    delete: 'Delete',
    deleteTooltip: 'Batch delete selected logs',
  },

  columns: {
    alertLogId: 'Log ID',
    alertLevel: 'Alert Level',
    alertType: 'Alert Type',
    alertTitle: 'Alert Title',
    alertContent: 'Alert Content',
    channelName: 'Channel Name',
    sendStatus: 'Send Status',
    alertTimestamp: 'Alert Time',
    sendTime: 'Send Time',
    sendErrorMessage: 'Error Message',
    addTime: 'Created At',
    addWho: 'Created By',
    editTime: 'Updated At',
    editWho: 'Updated By',
  },

  dialog: {
    title: 'Alert Log Details',
    titleWithId: 'Alert Log Details - {id}',
    basicInfo: 'Basic Information',
    alertTitle: 'Alert Title',
    alertContent: 'Alert Content',
    alertTags: 'Alert Tags (JSON)',
    alertExtra: 'Extra Data (JSON)',
    tableData: 'Table Data (JSON)',
    sendResult: 'Send Result (JSON)',
    empty: 'No log data',
  },

  confirm: {
    deleteTitle: 'Confirm Delete',
    deleteContent: 'Delete log "{id}"?',
    batchDeleteTitle: 'Confirm Batch Delete',
    batchDeleteContent: 'Delete the selected {count} log(s)?',
    confirmText: 'Delete',
    cancelText: 'Cancel',
  },

  message: {
    gridRefMissing: 'Grid reference is not set',
    selectToDelete: 'Please check logs to delete, or click a row first',
    alertLogIdRequired: 'Log ID is required',
    queryFailed: 'Failed to query alert logs',
    detailFailed: 'Failed to get alert log details',
    loadDetailFailed: 'Failed to load alert log details',
    deleteSuccess: 'Alert log deleted',
    deleteFailed: 'Failed to delete alert log',
    batchDeleteSuccess: 'Deleted {count} alert log(s)',
    batchDeleteFailed: 'Failed to batch delete alert logs',
  },
}
