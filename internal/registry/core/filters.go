package core

import (
	"strconv"
	"strings"
	"time"
)

// ================== 实例过滤器实现 ==================

// StatusFilter 状态过滤器
type StatusFilter struct {
	status string
}

func NewStatusFilter(status string) *StatusFilter {
	return &StatusFilter{status: status}
}

func (f *StatusFilter) Filter(instances []*ServiceInstance) []*ServiceInstance {
	var result []*ServiceInstance
	for _, instance := range instances {
		if instance.InstanceStatus == f.status {
			result = append(result, instance)
		}
	}
	return result
}

func (f *StatusFilter) Name() string {
	return "StatusFilter"
}

// HealthFilter 健康状态过滤器
type HealthFilter struct {
	healthStatus string
}

func NewHealthFilter(healthStatus string) *HealthFilter {
	return &HealthFilter{healthStatus: healthStatus}
}

func (f *HealthFilter) Filter(instances []*ServiceInstance) []*ServiceInstance {
	var result []*ServiceInstance
	for _, instance := range instances {
		if instance.HealthStatus == f.healthStatus {
			result = append(result, instance)
		}
	}
	return result
}

func (f *HealthFilter) Name() string {
	return "HealthFilter"
}

// EnabledFilter 启用状态过滤器
type EnabledFilter struct {
	enabled bool
}

func NewEnabledFilter(enabled bool) *EnabledFilter {
	return &EnabledFilter{enabled: enabled}
}

func (f *EnabledFilter) Filter(instances []*ServiceInstance) []*ServiceInstance {
	var result []*ServiceInstance
	expectedFlag := FlagNo
	if f.enabled {
		expectedFlag = FlagYes
	}

	for _, instance := range instances {
		if instance.ActiveFlag == expectedFlag {
			result = append(result, instance)
		}
	}
	return result
}

func (f *EnabledFilter) Name() string {
	return "EnabledFilter"
}

// TagFilter 标签过滤器
type TagFilter struct {
	tag string
}

func NewTagFilter(tag string) *TagFilter {
	return &TagFilter{tag: tag}
}

func (f *TagFilter) Filter(instances []*ServiceInstance) []*ServiceInstance {
	var result []*ServiceInstance
	for _, instance := range instances {
		tags := instance.GetTags()
		for _, tag := range tags {
			if tag == f.tag {
				result = append(result, instance)
				break
			}
		}
	}
	return result
}

func (f *TagFilter) Name() string {
	return "TagFilter"
}

// MetadataFilter 元数据过滤器
type MetadataFilter struct {
	key   string
	value string
}

func NewMetadataFilter(key, value string) *MetadataFilter {
	return &MetadataFilter{key: key, value: value}
}

func (f *MetadataFilter) Filter(instances []*ServiceInstance) []*ServiceInstance {
	var result []*ServiceInstance
	for _, instance := range instances {
		metadata := instance.GetMetadata()
		if metadata[f.key] == f.value {
			result = append(result, instance)
		}
	}
	return result
}

func (f *MetadataFilter) Name() string {
	return "MetadataFilter"
}

// VersionFilter 版本过滤器
type VersionFilter struct {
	version string
}

func NewVersionFilter(version string) *VersionFilter {
	return &VersionFilter{version: version}
}

func (f *VersionFilter) Filter(instances []*ServiceInstance) []*ServiceInstance {
	var result []*ServiceInstance
	for _, instance := range instances {
		metadata := instance.GetMetadata()
		if metadata["version"] == f.version {
			result = append(result, instance)
		}
	}
	return result
}

func (f *VersionFilter) Name() string {
	return "VersionFilter"
}

// HostFilter 主机过滤器
type HostFilter struct {
	host string
}

func NewHostFilter(host string) *HostFilter {
	return &HostFilter{host: host}
}

func (f *HostFilter) Filter(instances []*ServiceInstance) []*ServiceInstance {
	var result []*ServiceInstance
	for _, instance := range instances {
		if instance.HostAddress == f.host {
			result = append(result, instance)
		}
	}
	return result
}

func (f *HostFilter) Name() string {
	return "HostFilter"
}

// PortFilter 端口过滤器
type PortFilter struct {
	port int
}

func NewPortFilter(port int) *PortFilter {
	return &PortFilter{port: port}
}

func (f *PortFilter) Filter(instances []*ServiceInstance) []*ServiceInstance {
	var result []*ServiceInstance
	for _, instance := range instances {
		if instance.PortNumber == f.port {
			result = append(result, instance)
		}
	}
	return result
}

func (f *PortFilter) Name() string {
	return "PortFilter"
}

// WeightRangeFilter 权重范围过滤器
type WeightRangeFilter struct {
	minWeight int
	maxWeight int
}

func NewWeightRangeFilter(minWeight, maxWeight int) *WeightRangeFilter {
	return &WeightRangeFilter{minWeight: minWeight, maxWeight: maxWeight}
}

func (f *WeightRangeFilter) Filter(instances []*ServiceInstance) []*ServiceInstance {
	var result []*ServiceInstance
	for _, instance := range instances {
		if instance.WeightValue >= f.minWeight && instance.WeightValue <= f.maxWeight {
			result = append(result, instance)
		}
	}
	return result
}

func (f *WeightRangeFilter) Name() string {
	return "WeightRangeFilter"
}

// ProtocolFilter 协议过滤器
type ProtocolFilter struct {
	protocol string
}

func NewProtocolFilter(protocol string) *ProtocolFilter {
	return &ProtocolFilter{protocol: protocol}
}

func (f *ProtocolFilter) Filter(instances []*ServiceInstance) []*ServiceInstance {
	var result []*ServiceInstance
	for _, instance := range instances {
		metadata := instance.GetMetadata()
		if metadata["protocol"] == f.protocol {
			result = append(result, instance)
		}
	}
	return result
}

func (f *ProtocolFilter) Name() string {
	return "ProtocolFilter"
}

// ================== 复合过滤器 ==================

// AndFilter AND逻辑过滤器
type AndFilter struct {
	filters []InstanceFilter
}

func NewAndFilter(filters ...InstanceFilter) *AndFilter {
	return &AndFilter{filters: filters}
}

func (f *AndFilter) Filter(instances []*ServiceInstance) []*ServiceInstance {
	result := instances
	for _, filter := range f.filters {
		result = filter.Filter(result)
		if len(result) == 0 {
			break
		}
	}
	return result
}

func (f *AndFilter) Name() string {
	var names []string
	for _, filter := range f.filters {
		names = append(names, filter.Name())
	}
	return "AndFilter(" + strings.Join(names, ",") + ")"
}

// OrFilter OR逻辑过滤器
type OrFilter struct {
	filters []InstanceFilter
}

func NewOrFilter(filters ...InstanceFilter) *OrFilter {
	return &OrFilter{filters: filters}
}

func (f *OrFilter) Filter(instances []*ServiceInstance) []*ServiceInstance {
	var result []*ServiceInstance
	instanceMap := make(map[string]*ServiceInstance)

	for _, filter := range f.filters {
		filtered := filter.Filter(instances)
		for _, instance := range filtered {
			if _, exists := instanceMap[instance.ServiceInstanceId]; !exists {
				instanceMap[instance.ServiceInstanceId] = instance
				result = append(result, instance)
			}
		}
	}
	return result
}

func (f *OrFilter) Name() string {
	var names []string
	for _, filter := range f.filters {
		names = append(names, filter.Name())
	}
	return "OrFilter(" + strings.Join(names, ",") + ")"
}

// NotFilter NOT逻辑过滤器
type NotFilter struct {
	filter InstanceFilter
}

func NewNotFilter(filter InstanceFilter) *NotFilter {
	return &NotFilter{filter: filter}
}

func (f *NotFilter) Filter(instances []*ServiceInstance) []*ServiceInstance {
	filtered := f.filter.Filter(instances)
	filteredMap := make(map[string]bool)

	for _, instance := range filtered {
		filteredMap[instance.ServiceInstanceId] = true
	}

	var result []*ServiceInstance
	for _, instance := range instances {
		if !filteredMap[instance.ServiceInstanceId] {
			result = append(result, instance)
		}
	}
	return result
}

func (f *NotFilter) Name() string {
	return "NotFilter(" + f.filter.Name() + ")"
}

// ================== 预定义过滤器 ==================

// GetHealthyInstancesFilter 获取健康实例过滤器
func GetHealthyInstancesFilter() InstanceFilter {
	return NewAndFilter(
		NewEnabledFilter(true),
		NewStatusFilter(InstanceStatusUp),
		NewHealthFilter(HealthStatusHealthy),
	)
}

// GetAvailableInstancesFilter 获取可用实例过滤器
func GetAvailableInstancesFilter() InstanceFilter {
	return GetHealthyInstancesFilter()
}

// GetUnhealthyInstancesFilter 获取不健康实例过滤器
func GetUnhealthyInstancesFilter() InstanceFilter {
	return NewOrFilter(
		NewStatusFilter(InstanceStatusDown),
		NewStatusFilter(InstanceStatusOutOfService),
		NewHealthFilter(HealthStatusUnhealthy),
		NewEnabledFilter(false),
	)
}

// ================== 服务过滤器实现 ==================

// ServiceActiveFilter 服务活跃状态过滤器
type ServiceActiveFilter struct {
	active bool
}

func NewServiceActiveFilter(active bool) *ServiceActiveFilter {
	return &ServiceActiveFilter{active: active}
}

func (f *ServiceActiveFilter) Filter(services []*Service) []*Service {
	var result []*Service
	expectedFlag := FlagNo
	if f.active {
		expectedFlag = FlagYes
	}

	for _, service := range services {
		if service.ActiveFlag == expectedFlag {
			result = append(result, service)
		}
	}
	return result
}

func (f *ServiceActiveFilter) Name() string {
	return "ServiceActiveFilter"
}

// ServiceGroupFilter 服务分组过滤器
type ServiceGroupFilter struct {
	groupName string
}

func NewServiceGroupFilter(groupName string) *ServiceGroupFilter {
	return &ServiceGroupFilter{groupName: groupName}
}

func (f *ServiceGroupFilter) Filter(services []*Service) []*Service {
	var result []*Service
	for _, service := range services {
		if service.GroupName == f.groupName {
			result = append(result, service)
		}
	}
	return result
}

func (f *ServiceGroupFilter) Name() string {
	return "ServiceGroupFilter"
}

// ServiceProtocolFilter 服务协议过滤器
type ServiceProtocolFilter struct {
	protocol string
}

func NewServiceProtocolFilter(protocol string) *ServiceProtocolFilter {
	return &ServiceProtocolFilter{protocol: protocol}
}

func (f *ServiceProtocolFilter) Filter(services []*Service) []*Service {
	var result []*Service
	for _, service := range services {
		if service.ProtocolType == f.protocol {
			result = append(result, service)
		}
	}
	return result
}

func (f *ServiceProtocolFilter) Name() string {
	return "ServiceProtocolFilter"
}

// ================== 事件过滤器实现 ==================

// EventTypeFilter 事件类型过滤器
type EventTypeFilter struct {
	eventType string
}

func NewEventTypeFilter(eventType string) *EventTypeFilter {
	return &EventTypeFilter{eventType: eventType}
}

func (f *EventTypeFilter) Filter(events []*ServiceEvent) []*ServiceEvent {
	var result []*ServiceEvent
	for _, event := range events {
		if event.EventType == f.eventType {
			result = append(result, event)
		}
	}
	return result
}

func (f *EventTypeFilter) Name() string {
	return "EventTypeFilter"
}

// EventServiceFilter 事件服务过滤器
type EventServiceFilter struct {
	serviceName string
}

func NewEventServiceFilter(serviceName string) *EventServiceFilter {
	return &EventServiceFilter{serviceName: serviceName}
}

func (f *EventServiceFilter) Filter(events []*ServiceEvent) []*ServiceEvent {
	var result []*ServiceEvent
	for _, event := range events {
		if event.ServiceName == f.serviceName {
			result = append(result, event)
		}
	}
	return result
}

func (f *EventServiceFilter) Name() string {
	return "EventServiceFilter"
}

// EventTimeRangeFilter 事件时间范围过滤器
type EventTimeRangeFilter struct {
	startTime time.Time
	endTime   time.Time
}

func NewEventTimeRangeFilter(startTime, endTime time.Time) *EventTimeRangeFilter {
	return &EventTimeRangeFilter{startTime: startTime, endTime: endTime}
}

func (f *EventTimeRangeFilter) Filter(events []*ServiceEvent) []*ServiceEvent {
	var result []*ServiceEvent
	for _, event := range events {
		if event.EventTime.After(f.startTime) && event.EventTime.Before(f.endTime) {
			result = append(result, event)
		}
	}
	return result
}

func (f *EventTimeRangeFilter) Name() string {
	return "EventTimeRangeFilter"
}

// ================== 工具函数 ==================

// ApplyInstanceFilters 应用实例过滤器
func ApplyInstanceFilters(instances []*ServiceInstance, filters ...InstanceFilter) []*ServiceInstance {
	result := instances
	for _, filter := range filters {
		result = filter.Filter(result)
		if len(result) == 0 {
			break
		}
	}
	return result
}

// ApplyServiceFilters 应用服务过滤器
func ApplyServiceFilters(services []*Service, filters ...ServiceFilter) []*Service {
	result := services
	for _, filter := range filters {
		result = filter.Filter(result)
		if len(result) == 0 {
			break
		}
	}
	return result
}

// ApplyEventFilters 应用事件过滤器
func ApplyEventFilters(events []*ServiceEvent, filters ...EventFilter) []*ServiceEvent {
	result := events
	for _, filter := range filters {
		result = filter.Filter(result)
		if len(result) == 0 {
			break
		}
	}
	return result
}

// ParseInstanceFilters 解析实例过滤器参数
func ParseInstanceFilters(params map[string]string) []InstanceFilter {
	var filters []InstanceFilter

	// 状态过滤器
	if status, ok := params["status"]; ok {
		filters = append(filters, NewStatusFilter(status))
	}

	// 健康状态过滤器
	if health, ok := params["health"]; ok {
		filters = append(filters, NewHealthFilter(health))
	}

	// 启用状态过滤器
	if enabled, ok := params["enabled"]; ok {
		if enabled == "true" {
			filters = append(filters, NewEnabledFilter(true))
		} else if enabled == "false" {
			filters = append(filters, NewEnabledFilter(false))
		}
	}

	// 标签过滤器
	if tag, ok := params["tag"]; ok {
		filters = append(filters, NewTagFilter(tag))
	}

	// 版本过滤器
	if version, ok := params["version"]; ok {
		filters = append(filters, NewVersionFilter(version))
	}

	// 主机过滤器
	if host, ok := params["host"]; ok {
		filters = append(filters, NewHostFilter(host))
	}

	// 端口过滤器
	if portStr, ok := params["port"]; ok {
		if port, err := strconv.Atoi(portStr); err == nil {
			filters = append(filters, NewPortFilter(port))
		}
	}

	// 协议过滤器
	if protocol, ok := params["protocol"]; ok {
		filters = append(filters, NewProtocolFilter(protocol))
	}

	// 权重范围过滤器
	if minWeightStr, ok := params["minWeight"]; ok {
		if maxWeightStr, ok := params["maxWeight"]; ok {
			if minWeight, err := strconv.Atoi(minWeightStr); err == nil {
				if maxWeight, err := strconv.Atoi(maxWeightStr); err == nil {
					filters = append(filters, NewWeightRangeFilter(minWeight, maxWeight))
				}
			}
		}
	}

	// 元数据过滤器
	for key, value := range params {
		if strings.HasPrefix(key, "metadata.") {
			metaKey := strings.TrimPrefix(key, "metadata.")
			filters = append(filters, NewMetadataFilter(metaKey, value))
		}
	}

	return filters
}
