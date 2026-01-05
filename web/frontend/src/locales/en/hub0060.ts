/**
 * hub0060 Tunnel Server Management Module English Language Pack
 */
export default {
  // Page title and description
  pageTitle: 'Tunnel Server Management',
  pageDescription: 'Manage FRP tunnel server configurations, including control ports, authentication, network settings, etc.',

  // Statistics
  stats: {
    totalServers: 'Total Servers',
    runningServers: 'Running',
    stoppedServers: 'Stopped',
    errorServers: 'Error Status',
    totalClients: 'Total Clients',
    totalConnections: 'Total Connections',
    clientUsage: 'Usage Rate',
    connectionLoad: 'Load',
    healthyServers: 'Healthy Servers',
    stoppedServersStat: 'Stopped Servers',
    abnormalServers: 'Abnormal Servers',
    healthOverview: 'Health Overview'
  },

  // Table column headers
  table: {
    serverName: 'Server Name',
    controlAddress: 'Control Address',
    status: 'Status',
    maxClients: 'Client Limit',
    auth: 'Authentication',
    tls: 'TLS',
    heartbeat: 'Heartbeat Config',
    startTime: 'Start Time',
    createTime: 'Create Time',
    actions: 'Actions',
    running: 'Running',
    stopped: 'Stopped',
    error: 'Error',
    enabled: 'Enabled',
    disabled: 'Disabled'
  },

  // Search and filter
  search: {
    serverName: 'Search server name',
    serverStatus: 'Server Status',
    controlAddress: 'Control Address',
    controlPort: 'Control Port',
    search: 'Search',
    reset: 'Reset'
  },

  // Button actions
  actions: {
    refresh: 'Refresh',
    create: 'Add Server',
    batchDelete: 'Batch Delete',
    start: 'Start',
    stop: 'Stop',
    restart: 'Restart',
    test: 'Test',
    edit: 'Edit',
    delete: 'Delete',
    startServer: 'Start Server',
    stopServer: 'Stop Server',
    restartServer: 'Restart Server',
    testConnection: 'Test Connection',
    editServer: 'Edit Server',
    deleteServer: 'Delete Server'
  },

  // Dialog
  dialog: {
    create: 'Create Tunnel Server',
    edit: 'Edit Tunnel Server',
    basicConfig: 'Basic Configuration',
    networkConfig: 'Network Configuration',
    authConfig: 'Authentication Configuration',
    advancedConfig: 'Advanced Configuration',
    
    // Form fields
    form: {
      serverName: 'Server Name',
      serverNamePlaceholder: 'Please enter server name',
      serverDescription: 'Server Description',
      serverDescriptionPlaceholder: 'Please enter server description',
      controlAddress: 'Control Address',
      controlAddressPlaceholder: 'Please enter control address, e.g.: 0.0.0.0',
      controlPort: 'Control Port',
      controlPortPlaceholder: 'Please enter control port',
      dashboardPort: 'Dashboard Port',
      dashboardPortPlaceholder: 'Please enter dashboard port',
      httpPort: 'HTTP Port',
      httpPortPlaceholder: 'Virtual host HTTP port',
      httpsPort: 'HTTPS Port',
      httpsPortPlaceholder: 'Virtual host HTTPS port',
      maxClients: 'Max Clients',
      maxClientsPlaceholder: 'Maximum client connections',
      maxPortsPerClient: 'Max Ports Per Client',
      maxPortsPerClientPlaceholder: 'Maximum ports per client',
      allowPorts: 'Allowed Port Ranges',
      allowPortsPlaceholder: 'e.g.: 10000-20000,30000-40000',
      enableTokenAuth: 'Enable Token Authentication',
      authToken: 'Authentication Token',
      authTokenPlaceholder: 'Please enter or generate auth token',
      generateToken: 'Generate Token',
      enableTls: 'Enable TLS',
      tlsCertFile: 'TLS Certificate File',
      tlsCertFilePlaceholder: 'Please enter TLS certificate file path',
      tlsKeyFile: 'TLS Private Key File',
      tlsKeyFilePlaceholder: 'Please enter TLS private key file path',
      heartbeatInterval: 'Heartbeat Interval (seconds)',
      heartbeatIntervalPlaceholder: 'Heartbeat interval time',
      heartbeatTimeout: 'Heartbeat Timeout (seconds)',
      heartbeatTimeoutPlaceholder: 'Heartbeat timeout time',
      logLevel: 'Log Level',
      logLevelPlaceholder: 'Please select log level',
      noteText: 'Notes',
      noteTextPlaceholder: 'Please enter notes'
    },

    // Buttons
    cancel: 'Cancel',
    createBtn: 'Create',
    update: 'Update'
  },

  // Confirmation dialogs
  confirm: {
    batchDelete: 'Confirm Batch Delete',
    batchDeleteContent: 'Confirm to delete the selected {count} tunnel servers?',
    delete: 'Confirm to delete this tunnel server?',
    confirmDelete: 'Confirm Delete',
    cancel: 'Cancel'
  },

  // Message prompts
  messages: {
    createSuccess: 'Tunnel server created successfully',
    updateSuccess: 'Tunnel server updated successfully',
    deleteSuccess: 'Tunnel server deleted successfully',
    batchDeleteSuccess: 'Successfully deleted {count} tunnel servers',
    startSuccess: 'Tunnel server started successfully',
    stopSuccess: 'Tunnel server stopped successfully',
    restartSuccess: 'Tunnel server restarted successfully',
    testSuccess: 'Connection test successful',
    testSuccessWithLatency: 'Connection test successful (latency: {latency}ms)',
    testFailed: 'Connection test failed: Unable to connect to server',
    generateTokenSuccess: 'Authentication token generated successfully',
    refreshSuccess: 'Data refreshed successfully',
    createFailed: 'Failed to create tunnel server',
    updateFailed: 'Failed to update tunnel server',
    deleteFailed: 'Failed to delete tunnel server',
    batchDeleteFailed: 'Batch delete failed',
    startFailed: 'Start failed',
    stopFailed: 'Stop failed',
    restartFailed: 'Restart failed',
    generateTokenFailed: 'Failed to generate authentication token',
    getListFailed: 'Failed to get tunnel server list',
    getStatsFailed: 'Failed to get statistics'
  },

  // Form validation
  validation: {
    serverNameRequired: 'Please enter server name',
    serverNameLength: 'Server name length should be between 2-100 characters',
    controlAddressRequired: 'Please enter control address',
    controlPortRequired: 'Please enter control port',
    controlPortRange: 'Port range should be between 1-65535',
    maxClientsRequired: 'Please enter maximum clients',
    maxClientsRange: 'Maximum clients should be between 1-10000',
    heartbeatIntervalRequired: 'Please enter heartbeat interval',
    heartbeatIntervalRange: 'Heartbeat interval should be between 10-300 seconds',
    heartbeatTimeoutRequired: 'Please enter heartbeat timeout',
    heartbeatTimeoutRange: 'Heartbeat timeout should be between 30-600 seconds',
    authTokenRequired: 'Authentication token is required when token auth is enabled',
    tlsCertFileRequired: 'Certificate file path is required when TLS is enabled',
    tlsKeyFileRequired: 'Private key file path is required when TLS is enabled'
  },

  // Status options
  options: {
    logLevel: {
      debug: 'Debug',
      info: 'Info',
      warn: 'Warning',
      error: 'Error'
    },
    status: {
      running: 'Running',
      stopped: 'Stopped',
      error: 'Error'
    },
    auth: {
      enabled: 'Enabled',
      disabled: 'Disabled'
    }
  },

  // Pagination
  pagination: {
    total: 'Total {total} records'
  }
}
