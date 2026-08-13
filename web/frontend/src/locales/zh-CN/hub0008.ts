/**
 * 集群事件管理模块中文语言包
 * hub0008 - Cluster Event Management
 */
export default {
  moduleName: '集群事件管理',

  common: {
    all: '全部',
    active: '活动',
    inactive: '非活动',
    close: '关闭',
    viewDetail: '查看详情',
  },

  event: {
    search: {
      eventType: '事件类型',
      eventTypePlaceholder: '请输入事件类型',
      eventAction: '事件动作',
      eventActionPlaceholder: '请选择事件动作',
      activeFlag: '活动状态',
      activeFlagPlaceholder: '请选择活动状态',
      sourceNodeId: '发布节点ID',
      sourceNodeIdPlaceholder: '请输入发布节点ID',
      sourceNodeIp: '发布节点IP',
      sourceNodeIpPlaceholder: '请输入发布节点IP',
    },
    toolbar: {
      collapseAckList: '收起处理列表',
      expandAckList: '展开处理列表',
      toggleAckListTooltip: '收起/展开处理列表',
    },
    columns: {
      eventId: '事件ID',
      eventType: '事件类型',
      eventAction: '事件动作',
      sourceNodeId: '发布节点',
      sourceNodeIp: '发布节点IP',
      eventTime: '事件时间',
      expireTime: '过期时间',
      activeFlag: '活动状态',
    },
    dialog: {
      title: '事件详情',
      payloadTitle: '事件负载（JSON）',
    },
    message: {
      queryFailed: '查询集群事件列表失败',
      loadFailed: '加载集群事件列表失败',
    },
  },

  ack: {
    search: {
      nodeId: '处理节点ID',
      nodeIdPlaceholder: '请输入处理节点ID',
      nodeIp: '处理节点IP',
      nodeIpPlaceholder: '请输入处理节点IP',
      ackStatus: '确认状态',
      ackStatusPlaceholder: '请选择确认状态',
      activeFlag: '活动状态',
      activeFlagPlaceholder: '请选择活动状态',
    },
    status: {
      pending: '待处理',
      success: '成功',
      failed: '失败',
      skipped: '跳过',
    },
    columns: {
      ackId: '确认ID',
      nodeId: '处理节点ID',
      nodeIp: '处理节点IP',
      ackStatus: '确认状态',
      processTime: '处理时间',
      retryCount: '重试次数',
      resultMessage: '结果信息',
      activeFlag: '活动状态',
    },
    dialog: {
      title: '事件确认详情',
      eventId: '事件ID',
      addTime: '创建时间',
      addWho: '创建人',
      editTime: '修改时间',
      editWho: '修改人',
      resultTitle: '结果信息',
      noteTitle: '备注信息',
      extTitle: '扩展属性（JSON）',
    },
    message: {
      queryFailed: '查询事件处理节点列表失败',
      loadFailed: '加载事件处理节点列表失败',
      ackIdRequired: '确认ID不能为空',
      detailFailed: '获取事件确认详情失败',
    },
  },
}
