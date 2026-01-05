/**
 * Hub0041 Service Registry Management Module - English Language Pack
 * 
 * Contains all English text for service registry management
 * 
 * @author System Architecture Team
 * @version 1.0.0
 * @since 2024-01-01
 */

export default {
  // Page titles
  title: 'Service Registry Management',
  pageTitle: 'Service Registry Center',
  subtitle: 'Manage services and service instances registered by third-party applications',

  // Menu and navigation
  menu: {
    serviceRegistry: 'Service Registry Management',
    serviceList: 'Service List',
    instanceList: 'Instance List'
  },

  // Search and filters
  searchServiceName: 'Search service name',
  selectProtocol: 'Select protocol',
  selectStatus: 'Select status',

  // Service detail related
  basicInfo: 'Basic Information',
  serviceDetail: 'Service Details',
  instanceList: 'Instance List',
  healthCheckConfig: 'Health Check Configuration',
  noServices: 'No service data',
  noDescription: 'No description',
  serviceNotFound: 'Service not found',
  
  // New table fields
  serviceName: 'Service Name',
  serviceDescription: 'Service Description',
  healthCheckUrl: 'Health Check URL',
  healthCheckInterval: 'Health Check Interval',
  healthCheckTimeout: 'Health Check Timeout',
  healthCheckType: 'Health Check Type',
  healthCheckMode: 'Health Check Mode',
  selectHealthCheckType: 'Please select health check type',
  selectHealthCheckMode: 'Please select health check mode',
  lastUpdateTime: 'Last Update Time',
  instanceId: 'Instance ID',
  instanceCount: 'Instance Count',
  healthyCount: 'Healthy Count',
  instances: 'Instances',
  healthRate: 'Health Rate',
  totalInstances: 'Total Instances',
  healthyInstances: 'Healthy Instances',
  serviceHealth: 'Service Health',
  serviceGroup: 'Service Group',
  namespace: 'Namespace',
  port: 'Port',
  weight: 'Weight',
  lastHeartbeat: 'Last Heartbeat',
  
  // Load balance strategy texts
  roundRobin: 'Round Robin',
  weightedRoundRobin: 'Weighted Round Robin',
  leastConnections: 'Least Connections',
  random: 'Random',
  ipHash: 'IP Hash',
  
  // Actions related
  viewDetail: 'View Details',
  viewMetadata: 'View Metadata',
  refreshService: 'Refresh Service',
  refreshInstances: 'Refresh Instances',
  
  // Message texts
  fetchServicesFailed: 'Failed to fetch service list',
  fetchDetailFailed: 'Failed to fetch service details',
  fetchMetadataFailed: 'Failed to fetch metadata',
  refreshServiceSuccess: 'Refresh service successful',
  refreshServiceFailed: 'Refresh service failed',
  refreshInstancesSuccess: 'Refresh instances successful',
  refreshInstancesFailed: 'Refresh instances failed',
  healthCheckSuccess: 'Health check completed',
  healthCheckFailed: 'Health check failed',
  bringUpSuccess: 'Instance brought up successfully',
  bringUpFailed: 'Failed to bring up instance',
  takeDownSuccess: 'Instance taken down successfully',
  takeDownFailed: 'Failed to take down instance',
  
  // Status texts
  healthy: 'Healthy',
  unhealthy: 'Unhealthy',
  unknown: 'Unknown',

  // Service form dialog
  addService: 'Add Service',
  editService: 'Edit Service',
  addInstance: 'Add Instance',
  editInstance: 'Edit Instance',
  instanceConfig: 'Service Instance Configuration',
  networkConfig: 'Network Configuration',
  selectServiceGroup: 'Please select service group',
  selectLoadBalance: 'Please select load balance strategy',
  serviceNamePlaceholder: 'Please enter service name',
  serviceDescriptionPlaceholder: 'Please enter service description',
  contextPathPlaceholder: 'Please enter context path, e.g.: /api',
  healthCheckUrlPlaceholder: 'Please enter health check URL, e.g.: /health',
  healthCheckIntervalPlaceholder: 'Health check interval',
  healthCheckTimeoutPlaceholder: 'Health check timeout',
  seconds: 'seconds',
  cancel: 'Cancel',
  reset: 'Reset',
  submit: 'Submit',

  // Form validation
  serviceNameRequired: 'Please enter service name',
  serviceNameLength: 'Service name length should be between 2-50 characters',
  serviceGroupRequired: 'Please select service group',
  instanceHostRequired: 'Please enter instance host address',
  instancePortRequired: 'Please enter instance port number',
  instanceStatusRequired: 'Please select instance status',
  healthStatusRequired: 'Please select health status',
  weightRequired: 'Please enter weight value',
  serviceDescriptionLength: 'Service description length cannot exceed 200 characters',
  protocolTypeRequired: 'Please select protocol type',
  contextPathRequired: 'Please enter context path',
  loadBalanceStrategyRequired: 'Please select load balance strategy',
  healthCheckUrlRequired: 'Please enter health check URL',
  healthCheckIntervalRequired: 'Please enter health check interval',
  healthCheckTimeoutRequired: 'Please enter health check timeout',
  healthCheckTypeRequired: 'Please select health check type',
  healthCheckModeRequired: 'Please select health check mode',

  // Operation results
  addServiceSuccess: 'Service added successfully',
  addServiceFailed: 'Failed to add service',
  updateServiceSuccess: 'Service updated successfully',
  updateServiceFailed: 'Failed to update service',
  deleteServiceSuccess: 'Service deleted successfully',
  deleteServiceFailed: 'Failed to delete service',
  saveServiceBeforeAddingInstance: 'Please save the service before adding instances',
  fetchServiceGroupsFailed: 'Failed to fetch service groups',
  loadInstancesSuccess: 'Instances loaded successfully',
  loadInstancesFailed: 'Failed to load instances',
  addInstanceSuccess: 'Instance added successfully',
  addInstanceFailed: 'Failed to add instance',
  updateInstanceSuccess: 'Instance updated successfully',
  updateInstanceFailed: 'Failed to update instance',
  deleteInstanceSuccess: 'Instance deleted successfully',
  deleteInstanceFailed: 'Failed to delete instance',

  // Page related
  serviceManagement: 'Service Management',
  advancedConfig: 'Advanced Configuration',
  maxInstances: 'Max Instances',
  maxInstancesPlaceholder: 'Please enter max instances',
  activeFlagRequired: 'Please select service status',
  maxInstancesRequired: 'Please enter max instances',

  // Service group selection dialog
  searchGroupName: 'Search group name',
  selectGroupType: 'Select group type',
  noServiceGroups: 'No service groups available',
  selectedGroup: 'Selected group',
  groupType: {
    BUSINESS: 'Business Group',
    SYSTEM: 'System Group',
    TEST: 'Test Group'
  },
  
  // Registry types
  registryTypes: {
    INTERNAL: 'Internal Management',
    NACOS: 'Nacos Registry',
    CONSUL: 'Consul Registry',
    EUREKA: 'Eureka Registry',
    ETCD: 'ETCD Registry',
    ZOOKEEPER: 'ZooKeeper Registry'
  },
  selectRegistryType: 'Please select registry type',
  registryTypeSelection: 'Registry Type Selection',
  serviceConfig: 'Service Configuration',
  registryTypeRequired: 'Please select registry type',

  // Table column headers
  columns: {
    // Service table
    serviceName: 'Service Name',
    serviceGroupId: 'Group ID',
    groupName: 'Group Name', 
    protocolType: 'Protocol Type',
    loadBalanceStrategy: 'Load Balance',
    instanceCount: 'Instance Count',
    healthyInstanceCount: 'Healthy Instances',
    activeFlag: 'Status',
    addTime: 'Created Time',
    editTime: 'Updated Time',
    actions: 'Actions',

    // Service instance table
    serviceInstanceId: 'Instance ID',
    hostAddress: 'Host Address',
    portNumber: 'Port',
    contextPath: 'Context Path',
    instanceStatus: 'Instance Status',
    healthStatus: 'Health Status',
    weightValue: 'Weight',
    clientId: 'Client ID',
    clientVersion: 'Client Version', 
    clientType: 'Client Type',
    tempInstanceFlag: 'Temporary Flag',
    registerTime: 'Register Time',
    lastHeartbeatTime: 'Last Heartbeat',
    lastHealthCheckTime: 'Last Health Check'
  },

  // Status enums
  status: {
    // Instance status
    UP: 'Up',
    DOWN: 'Down',
    STARTING: 'Starting',
    OUT_OF_SERVICE: 'Out of Service',

    // Health status
    HEALTHY: 'Healthy',
    UNHEALTHY: 'Unhealthy', 
    CHECKING: 'Checking',
    UNKNOWN: 'Unknown',

    // Client types
    JAVA: 'Java',
    DOTNET: '.NET',
    NODEJS: 'Node.js',
    PYTHON: 'Python',
    GO: 'Go',
    OTHER: 'Other',

    // Protocol types
    HTTP: 'HTTP',
    HTTPS: 'HTTPS',
    TCP: 'TCP',
    UDP: 'UDP', 
    GRPC: 'gRPC',

    // Load balance strategies
    ROUND_ROBIN: 'Round Robin',
    WEIGHTED_ROUND_ROBIN: 'Weighted Round Robin',
    LEAST_CONNECTIONS: 'Least Connections',
    RANDOM: 'Random',
    HASH: 'Hash',

    // General status
    Y: 'Enabled',
    N: 'Disabled',
    temporary: 'Temporary Instance',
    permanent: 'Permanent Instance'
  },

  // Action buttons
  actions: {
    view: 'View',
    refresh: 'Refresh',
    edit: 'Edit',
    delete: 'Delete',
    healthCheck: 'Health Check',
    batchRefresh: 'Batch Refresh',
    batchHealthCheck: 'Batch Health Check',
    export: 'Export',
    search: 'Search',
    reset: 'Reset',
    add: 'Add Service',
    addInstance: 'Add Instance',
    back: 'Back',
    create: 'Create',
    update: 'Update',
    confirm: 'Confirm',
    cancel: 'Cancel',
    bringUp: 'Bring Up',
    takeDown: 'Take Down',
    viewEvents: 'View Events'
  },

  // Search form
  search: {
    placeholder: {
      serviceName: 'Enter service name',
      groupName: 'Enter group name',
      hostAddress: 'Enter host address',
      clientId: 'Enter client ID'
    }
  },

  // Statistics
  statistics: {
    totalServices: 'Total Services',
    activeServices: 'Active Services',
    inactiveServices: 'Inactive Services',
    totalInstances: 'Total Instances',
    healthyInstances: 'Healthy Instances',
    unhealthyInstances: 'Unhealthy Instances',
    upInstances: 'Up Instances',
    downInstances: 'Down Instances'
  },

  // Message prompts
  messages: {
    // Success messages
    refreshSuccess: 'Refresh successful',
    healthCheckSuccess: 'Health check completed',
    batchRefreshSuccess: 'Batch refresh successful',
    batchHealthCheckSuccess: 'Batch health check completed',
    exportSuccess: 'Export successful',

    // Error messages
    loadError: 'Failed to load data',
    refreshError: 'Refresh failed',
    healthCheckError: 'Health check failed',
    batchRefreshError: 'Batch refresh failed',
    batchHealthCheckError: 'Batch health check failed',
    exportError: 'Export failed',
    noSelection: 'Please select items to operate',

    // Confirmation messages
    confirmRefresh: 'Confirm to refresh selected services?',
    confirmHealthCheck: 'Confirm to perform health check?',
    confirmDeleteInstance: 'Are you sure you want to delete this instance?',
    confirmDeleteService: 'Are you sure you want to delete this service? This action will also delete all instances under this service and cannot be undone.'
  },

  // Dialogs
  dialogs: {
    serviceDetail: {
      title: 'Service Details',
      tabs: {
        basic: 'Basic Info',
        instances: 'Instance List',
        metadata: 'Metadata',
        audit: 'Audit Info'
      }
    }
  },

  // Table related
  table: {
    empty: 'No data',
    loading: 'Loading...',
    selectAll: 'Select All',
    selected: '{count} items selected',
    heartbeatTimeout: 'Heartbeat timeout',
    noInstances: 'No instances',
    noHealthCheck: 'Not checked yet'
  },

  // Time related
  time: {
    ago: '{time} ago',
    justNow: 'Just now',
    seconds: 'seconds',
    minutes: 'minutes',
    hours: 'hours',
    days: 'days'
  },

  // Extension configuration related
  extensionConfig: 'Extension Configuration',
  metadataAndTags: 'Metadata and Tags',
  notesAndExtProperty: 'Notes and Extension Property',
  reservedFields: 'Reserved Fields',
  forFutureExpansion: 'Future Expansion',
  groupName: 'Group Name',
  groupNamePlaceholder: 'Auto-filled from selected group',
  metadataJson: 'Service Metadata',
  metadataJsonPlaceholder: 'Please enter JSON format metadata, e.g.: {"version":"1.0","env":"prod"}',
  tagsJson: 'Service Tags',
  tagsJsonPlaceholder: 'Please enter JSON format tags, e.g.: ["web","api","microservice"]',
  noteText: 'Notes',
  noteTextPlaceholder: 'Please enter service notes',
  noteTextLength: 'Notes cannot exceed 500 characters',
  extProperty: 'Extension Property',
  extPropertyPlaceholder: 'Please enter JSON format extension properties for future functionality',
  reservedField: 'Reserved Field {number}',
  reservedFieldPlaceholder: 'Reserved field {number} for future expansion',
  invalidJsonFormat: 'Please enter valid JSON format',
  expand: 'Expand',
  collapse: 'Collapse',
  
  // Health status texts
  healthStatus: {
    excellent: 'Excellent',
    good: 'Good',
    warning: 'Warning', 
    critical: 'Critical',
    offline: 'Offline'
  },
  
  // Health check mode texts
  healthCheckModes: {
    active: 'Active Detection',
    passive: 'Passive Reporting'
  },
  
  // Service event related
  serviceEventLog: 'Service Event Log',
  
  // Event related messages
  fetchServiceEventsFailed: 'Failed to fetch service events',
  getServiceEventFailed: 'Failed to get service event details'
}
