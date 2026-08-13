/**
 * Cluster Event Management Module English locale
 * hub0008 - Cluster Event Management
 */
export default {
  moduleName: 'Cluster Event Management',

  common: {
    all: 'All',
    active: 'Active',
    inactive: 'Inactive',
    close: 'Close',
    viewDetail: 'View Details',
  },

  event: {
    search: {
      eventType: 'Event Type',
      eventTypePlaceholder: 'Please enter event type',
      eventAction: 'Event Action',
      eventActionPlaceholder: 'Please select event action',
      activeFlag: 'Active Status',
      activeFlagPlaceholder: 'Please select active status',
      sourceNodeId: 'Source Node ID',
      sourceNodeIdPlaceholder: 'Please enter source node ID',
      sourceNodeIp: 'Source Node IP',
      sourceNodeIpPlaceholder: 'Please enter source node IP',
    },
    toolbar: {
      collapseAckList: 'Collapse Ack List',
      expandAckList: 'Expand Ack List',
      toggleAckListTooltip: 'Collapse / expand ack list',
    },
    columns: {
      eventId: 'Event ID',
      eventType: 'Event Type',
      eventAction: 'Event Action',
      sourceNodeId: 'Source Node',
      sourceNodeIp: 'Source Node IP',
      eventTime: 'Event Time',
      expireTime: 'Expire Time',
      activeFlag: 'Active Status',
    },
    dialog: {
      title: 'Event Details',
      payloadTitle: 'Event Payload (JSON)',
    },
    message: {
      queryFailed: 'Failed to query cluster events',
      loadFailed: 'Failed to load cluster events',
    },
  },

  ack: {
    search: {
      nodeId: 'Node ID',
      nodeIdPlaceholder: 'Please enter node ID',
      nodeIp: 'Node IP',
      nodeIpPlaceholder: 'Please enter node IP',
      ackStatus: 'Ack Status',
      ackStatusPlaceholder: 'Please select ack status',
      activeFlag: 'Active Status',
      activeFlagPlaceholder: 'Please select active status',
    },
    status: {
      pending: 'Pending',
      success: 'Success',
      failed: 'Failed',
      skipped: 'Skipped',
    },
    columns: {
      ackId: 'Ack ID',
      nodeId: 'Node ID',
      nodeIp: 'Node IP',
      ackStatus: 'Ack Status',
      processTime: 'Process Time',
      retryCount: 'Retry Count',
      resultMessage: 'Result Message',
      activeFlag: 'Active Status',
    },
    dialog: {
      title: 'Ack Details',
      eventId: 'Event ID',
      addTime: 'Created At',
      addWho: 'Created By',
      editTime: 'Updated At',
      editWho: 'Updated By',
      resultTitle: 'Result Message',
      noteTitle: 'Notes',
      extTitle: 'Extended Properties (JSON)',
    },
    message: {
      queryFailed: 'Failed to query event ack list',
      loadFailed: 'Failed to load event ack list',
      ackIdRequired: 'Ack ID is required',
      detailFailed: 'Failed to get ack details',
    },
  },
}
