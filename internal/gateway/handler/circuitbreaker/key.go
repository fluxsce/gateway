package circuitbreaker

// NodeCircuitKey 返回上游实例的熔断统计键。
// 网关摘除按「服务 + 节点」隔离，避免不同服务的同名节点撞键。
// serviceID 或 nodeID 为空时返回空串，调用方应跳过统计。
func NodeCircuitKey(serviceID, nodeID string) string {
	if serviceID == "" || nodeID == "" {
		return ""
	}
	return "cb_node:" + serviceID + ":" + nodeID
}
