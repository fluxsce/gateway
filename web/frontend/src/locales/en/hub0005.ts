/**
 * Permission Management Module English Internationalization File
 * hub0005 - Permission Management Module
 */

export default {
  // Module Name
  moduleName: 'Permission Management',
  
  // Menu Items
  menu: {
    roleManagement: 'Role Management',
    resourceManagement: 'Resource Management',
    userRoleAssignment: 'User Role Assignment',
    dataPermissionManagement: 'Data Permission Management',
    operationLog: 'Operation Log'
  },

  // Role Management
  role: {
    title: 'Role Management',
    list: 'Role List',
    add: 'Add Role',
    edit: 'Edit Role',
    delete: 'Delete Role',
    batchDelete: 'Batch Delete',
    assignPermissions: 'Assign Permissions',
    
    // Form Fields
    form: {
      roleName: 'Role Name',
      roleCode: 'Role Code',
      roleDescription: 'Role Description',
      roleType: 'Role Type',
      roleLevel: 'Role Level',
      roleStatus: 'Role Status',
      dataScope: 'Data Scope',
      dataScopeDeptIds: 'Department Permissions',
      noteText: 'Note',
      
      // Placeholders
      placeholder: {
        roleName: 'Please enter role name',
        roleCode: 'Please enter role code',
        roleDescription: 'Please enter role description',
        roleLevel: 'Please enter role level',
        noteText: 'Please enter note'
      },
      
      // Validation Messages
      validation: {
        roleNameRequired: 'Please enter role name',
        roleCodeRequired: 'Please enter role code',
        roleCodePattern: 'Role code must start with uppercase letter and contain only uppercase letters, numbers and underscores',
        roleTypeRequired: 'Please select role type',
        roleLevelRequired: 'Please enter role level',
        roleLevelRange: 'Role level should be between 1-999',
        dataScopeRequired: 'Please select data scope',
        roleStatusRequired: 'Please select role status'
      }
    },

    // Role Types
    type: {
      system: 'System Role',
      custom: 'Custom Role'
    },

    // Data Scopes
    dataScope: {
      all: 'All Data',
      tenant: 'Tenant Data',
      dept: 'Department Data',
      self: 'Personal Data'
    },

    // Status
    status: {
      enabled: 'Enabled',
      disabled: 'Disabled'
    },

    // Operation Messages
    message: {
      createSuccess: 'Role created successfully',
      updateSuccess: 'Role updated successfully',
      deleteSuccess: 'Role deleted successfully',
      batchDeleteSuccess: 'Batch delete successful',
      deleteConfirm: 'Are you sure to delete role "{name}"? This operation cannot be undone.',
      batchDeleteConfirm: 'Are you sure to delete selected {count} roles? This operation cannot be undone.',
      permissionAssignSuccess: 'Permissions assigned successfully'
    }
  },

  // Resource Management
  resource: {
    title: 'Resource Management',
    list: 'Resource List',
    tree: 'Resource Tree',
    add: 'Add Resource',
    addChild: 'Add Child Resource',
    edit: 'Edit Resource',
    delete: 'Delete Resource',
    sync: 'Sync Module Resources',
    
    // Form Fields
    form: {
      resourceName: 'Resource Name',
      resourceCode: 'Resource Code',
      resourceType: 'Resource Type',
      resourcePath: 'Resource Path',
      resourceMethod: 'Request Method',
      parentResource: 'Parent Resource',
      resourceLevel: 'Resource Level',
      sortOrder: 'Sort Order',
      moduleCode: 'Module Code',
      moduleName: 'Module Name',
      displayName: 'Display Name',
      iconClass: 'Icon Class',
      description: 'Description',
      resourceStatus: 'Resource Status',
      visibleFlag: 'Visible',
      builtInFlag: 'Built-in',
      noteText: 'Note',
      
      // Placeholders
      placeholder: {
        resourceName: 'Please enter resource name',
        resourceCode: 'Please enter resource code',
        resourcePath: 'Please enter resource path',
        moduleCode: 'Please enter module code',
        moduleName: 'Please enter module name',
        displayName: 'Please enter display name',
        iconClass: 'Please enter icon class',
        description: 'Please enter description',
        noteText: 'Please enter note'
      },
      
      // Validation Messages
      validation: {
        resourceNameRequired: 'Please enter resource name',
        resourceCodeRequired: 'Please enter resource code',
        resourceCodePattern: 'Resource code can only contain letters, numbers, colons, underscores and hyphens, and must start with a letter or number',
        resourceTypeRequired: 'Please select resource type',
        resourceLevelRequired: 'Please enter resource level',
        resourceLevelRange: 'Resource level should be between 1-10',
        sortOrderRequired: 'Please enter sort order',
        sortOrderRange: 'Sort order should be between 0-9999',
        resourceStatusRequired: 'Please select resource status',
        visibleFlagRequired: 'Please select visibility'
      }
    },

    // Resource Types
    type: {
      module: 'Module',
      menu: 'Menu',
      button: 'Button',
      api: 'API'
    },

    // View Modes
    view: {
      list: 'List View',
      tree: 'Tree View'
    },

    // Operation Messages
    message: {
      createSuccess: 'Resource created successfully',
      updateSuccess: 'Resource updated successfully',
      deleteSuccess: 'Resource deleted successfully',
      syncSuccess: 'Module resources synced successfully',
      deleteConfirm: 'Are you sure to delete resource "{name}"? This operation cannot be undone.',
      syncConfirm: 'Are you sure to sync module resources? This operation will automatically scan system modules and create corresponding permission resources.'
    }
  },

  // User Role Assignment
  userRole: {
    title: 'User Role Assignment',
    assign: 'Assign Role',
    revoke: 'Revoke Role',
    
    // Form Fields
    form: {
      userName: 'Username',
      roleName: 'Role Name',
      roleCode: 'Role Code',
      primaryRole: 'Primary Role',
      grantedTime: 'Granted Time',
      expireTime: 'Expire Time',
      neverExpire: 'Never Expire'
    },

    // Operation Messages
    message: {
      assignSuccess: 'Role assigned successfully',
      revokeSuccess: 'Role revoked successfully',
      revokeConfirm: 'Are you sure to revoke this role assignment?'
    }
  },

  // Data Permission Management
  dataPermission: {
    title: 'Data Permission Management',
    add: 'Add Data Permission',
    edit: 'Edit Data Permission',
    delete: 'Delete Data Permission',
    
    // Form Fields
    form: {
      resourceType: 'Resource Type',
      resourceCode: 'Resource Code',
      permissionScope: 'Permission Scope',
      scopeValue: 'Scope Value',
      filterCondition: 'Filter Condition',
      columnPermissions: 'Column Permissions',
      operationPermissions: 'Operation Permissions',
      effectiveTime: 'Effective Time',
      expireTime: 'Expire Time'
    },

    // Permission Scopes
    scope: {
      all: 'All',
      tenant: 'Tenant',
      dept: 'Department',
      self: 'Personal',
      custom: 'Custom'
    },

    // Operation Permissions
    operation: {
      read: 'Read Only',
      write: 'Read Write',
      delete: 'Delete'
    },

    // Operation Messages
    message: {
      createSuccess: 'Data permission created successfully',
      updateSuccess: 'Data permission updated successfully',
      deleteSuccess: 'Data permission deleted successfully',
      deleteConfirm: 'Are you sure to delete this data permission?'
    }
  },

  // Operation Log
  operationLog: {
    title: 'Operation Log',
    view: 'View Log',
    detail: 'Log Detail',
    export: 'Export Log',
    clean: 'Clean Log',
    
    // Form Fields
    form: {
      operationType: 'Operation Type',
      operationTarget: 'Operation Target',
      targetName: 'Target Name',
      operator: 'Operator',
      operatorIp: 'IP Address',
      operationResult: 'Operation Result',
      operationDescription: 'Operation Description',
      operationTime: 'Operation Time',
      beforeData: 'Before Data',
      afterData: 'After Data',
      errorMessage: 'Error Message'
    },

    // Operation Types
    type: {
      roleCreate: 'Create Role',
      roleUpdate: 'Update Role',
      roleDelete: 'Delete Role',
      permissionGrant: 'Grant Permission',
      permissionRevoke: 'Revoke Permission',
      userRoleAssign: 'Assign Role',
      userRoleRevoke: 'Revoke Role',
      dataPermissionCreate: 'Create Data Permission',
      dataPermissionUpdate: 'Update Data Permission',
      dataPermissionDelete: 'Delete Data Permission'
    },

    // Operation Targets
    target: {
      role: 'Role',
      resource: 'Resource',
      userRole: 'User Role',
      dataPermission: 'Data Permission'
    },

    // Operation Results
    result: {
      success: 'Success',
      failed: 'Failed'
    },

    // Operation Messages
    message: {
      exportSuccess: 'Log exported successfully',
      cleanSuccess: 'Log cleaned successfully',
      cleanConfirm: 'Are you sure to clean logs before {date}? This operation cannot be undone.'
    }
  },

  // Permission Verification
  permission: {
    title: 'Permission Verification',
    check: 'Permission Check',
    denied: 'Access Denied',
    required: 'Permission Required',
    
    message: {
      accessDenied: 'Access denied, insufficient permissions',
      loginRequired: 'Please login first'
    }
  },

  // Common
  common: {
    // Action Buttons
    action: {
      search: 'Search',
      reset: 'Reset',
      refresh: 'Refresh',
      add: 'Add',
      edit: 'Edit',
      delete: 'Delete',
      save: 'Save',
      cancel: 'Cancel',
      confirm: 'Confirm',
      submit: 'Submit',
      back: 'Back',
      close: 'Close',
      expand: 'Expand',
      collapse: 'Collapse',
      selectAll: 'Select All',
      unselectAll: 'Unselect All',
      expandAll: 'Expand All',
      collapseAll: 'Collapse All'
    },

    // Status
    status: {
      enabled: 'Enabled',
      disabled: 'Disabled',
      active: 'Active',
      inactive: 'Inactive',
      yes: 'Yes',
      no: 'No',
      visible: 'Visible',
      hidden: 'Hidden',
      builtin: 'Built-in',
      custom: 'Custom'
    },

    // Time
    time: {
      createTime: 'Create Time',
      updateTime: 'Update Time',
      effectiveTime: 'Effective Time',
      expireTime: 'Expire Time',
      neverExpire: 'Never Expire',
      immediately: 'Immediately'
    },

    // Pagination
    pagination: {
      total: 'Total {total} items',
      pageSize: 'Items per page',
      goto: 'Go to',
      page: 'Page'
    },

    // Messages
    message: {
      loading: 'Loading...',
      noData: 'No Data',
      operationSuccess: 'Operation successful',
      operationFailed: 'Operation failed',
      deleteConfirm: 'Are you sure to delete? This operation cannot be undone.',
      selectAtLeast: 'Please select at least one item',
      networkError: 'Network error, please try again later',
      systemError: 'System error'
    },

    // Form Validation
    validation: {
      required: 'This field is required',
      minLength: 'Length cannot be less than {min} characters',
      maxLength: 'Length cannot exceed {max} characters',
      email: 'Please enter a valid email address',
      phone: 'Please enter a valid phone number',
      url: 'Please enter a valid URL',
      number: 'Please enter a number',
      integer: 'Please enter an integer',
      positive: 'Please enter a positive number',
      range: 'Please enter a value between {min} and {max}'
    }
  }
}
