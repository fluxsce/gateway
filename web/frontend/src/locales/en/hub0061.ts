/**
 * hub0061 - Static Port Mapping Management Module English Language Pack
 */
export default {
  // Module Title
  title: 'Static Port Mapping Management',
  subtitle: 'Manage static port mapping configurations for tunnel servers',

  // Statistics
  stats: {
    totalNodes: 'Total Mappings',
    activeNodes: 'Active Mappings',
    inactiveNodes: 'Inactive Mappings',
    errorNodes: 'Error Mappings',
    totalConnections: 'Total Connections',
    totalTraffic: 'Total Traffic'
  },

  // Table Columns
  table: {
    nodeName: 'Mapping Name',
    nodeType: 'Node Type',
    proxyType: 'Proxy Type',
    listenPort: 'Listen Port',
    targetPort: 'Target Port',
    nodeStatus: 'Node Status',
    connectionCount: 'Current Connections',
    totalConnections: 'Total Connections',
    totalBytes: 'Total Traffic',
    maxConnections: 'Max Connections',
    compression: 'Compression',
    encryption: 'Encryption',
    healthCheckType: 'Health Check',
    lastHealthCheck: 'Last Check Time',
    createdTime: 'Created Time'
  },

  // Filters
  filter: {
    proxyType: 'Proxy Type',
    nodeStatus: 'Node Status'
  },

  // Node Types
  nodeType: {
    static: 'Static',
    dynamic: 'Dynamic'
  },

  // Node Status
  nodeStatus: {
    active: 'Active',
    inactive: 'Inactive',
    error: 'Error'
  },

  // Dialog
  dialog: {
    createTitle: 'Create Static Port Mapping',
    editTitle: 'Edit Static Port Mapping',
    tabs: {
      basic: 'Basic Config',
      advanced: 'Advanced Config',
      health: 'Health Check',
      http: 'HTTP Config',
      note: 'Note'
    }
  },

  // Form
  form: {
    tunnelServerId: 'Tunnel Server',
    tunnelServerIdPlaceholder: 'Please select tunnel server',
    tunnelServerIdRequired: 'Please select tunnel server',
    
    nodeName: 'Mapping Name',
    nodeNamePlaceholder: 'Please enter mapping name',
    nodeNameRequired: 'Please enter mapping name',
    nodeNameLength: 'Mapping name length should be between 2-100 characters',
    
    nodeType: 'Node Type',
    nodeTypeRequired: 'Please select node type',
    
    proxyType: 'Proxy Type',
    proxyTypePlaceholder: 'Please select proxy type',
    proxyTypeRequired: 'Please select proxy type',
    
    listenAddress: 'Listen Address',
    listenAddressPlaceholder: 'Please enter listen address, e.g. 0.0.0.0',
    listenAddressRequired: 'Please enter listen address',
    
    listenPort: 'Listen Port',
    listenPortPlaceholder: 'Please enter listen port (public port)',
    listenPortRequired: 'Please enter listen port',
    
    targetAddress: 'Target Address',
    targetAddressPlaceholder: 'Please enter target address (internal address)',
    targetAddressRequired: 'Please enter target address',
    
    targetPort: 'Target Port',
    targetPortPlaceholder: 'Please enter target port (internal port)',
    targetPortRequired: 'Please enter target port',
    
    maxConnections: 'Max Connections',
    maxConnectionsPlaceholder: 'Please enter max connections',
    
    compression: 'Enable Compression',
    encryption: 'Enable Encryption',
    
    secretKey: 'Secret Key',
    secretKeyPlaceholder: 'Please enter secret key',
    
    healthCheckType: 'Health Check Type',
    healthCheckTypePlaceholder: 'Please select health check type',
    
    healthCheckUrl: 'Health Check URL',
    healthCheckUrlPlaceholder: 'Please enter health check URL',
    
    healthCheckInterval: 'Check Interval',
    healthCheckIntervalPlaceholder: 'Please enter check interval (seconds)',
    
    subDomain: 'Sub Domain',
    subDomainPlaceholder: 'Please enter sub domain prefix',
    
    httpUser: 'HTTP Auth Username',
    httpUserPlaceholder: 'Please enter HTTP basic auth username',
    
    httpPassword: 'HTTP Auth Password',
    httpPasswordPlaceholder: 'Please enter HTTP basic auth password',
    
    hostHeaderRewrite: 'Host Header Rewrite',
    hostHeaderRewritePlaceholder: 'Please enter host header to rewrite',
    
    activeFlag: 'Active Flag',
    
    noteText: 'Note',
    noteTextPlaceholder: 'Please enter note'
  },

  // Messages
  message: {
    portConflict: 'Port is already in use, please change port or listen address',
    createSuccess: 'Static port mapping created successfully',
    createFailed: 'Failed to create static port mapping',
    updateSuccess: 'Static port mapping updated successfully',
    updateFailed: 'Failed to update static port mapping',
    deleteSuccess: 'Static port mapping deleted successfully',
    deleteFailed: 'Failed to delete static port mapping',
    enableSuccess: 'Static port mapping enabled successfully',
    enableFailed: 'Failed to enable static port mapping',
    disableSuccess: 'Static port mapping disabled successfully',
    disableFailed: 'Failed to disable static port mapping'
  }
}
