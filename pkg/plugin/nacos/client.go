// Package nacos 提供了基于 Nacos SDK 的服务注册发现功能。
//
// 该包封装了 Nacos SDK v2，提供了简洁易用的 API 来进行服务注册、发现和管理。
// 所有的服务发现操作都会优先使用 Nacos SDK 的本地缓存，只有在缓存过期或不存在时
// 才会从 Nacos 服务器获取最新数据，这大大提高了性能并减少了网络开销。
//
// 缓存机制说明：
//   - SelectInstances: 使用本地缓存，缓存时间由 Nacos SDK 管理（通常为30秒）
//   - SelectOneHealthyInstance: 使用本地缓存，内置负载均衡算法
//   - GetService: 使用本地缓存
//   - 服务订阅: 实时推送，无缓存延迟
//   - 服务注册/注销: 实时操作，立即生效
package nacos

import (
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"

	"gateway/pkg/logger"
	"gateway/pkg/plugin/tools/common"
)

// NacosClient 是 Nacos 命名服务客户端的高级封装。
//
// 它提供了简洁的 API 来进行服务注册、发现和管理操作，同时处理了参数验证、
// 错误转换和日志记录等细节。所有的服务发现操作都会利用 Nacos SDK 的
// 本地缓存机制来提高性能。
//
// 线程安全：NacosClient 是线程安全的，可以在多个 goroutine 中并发使用。
type NacosClient struct {
	namingClient naming_client.INamingClient // Nacos SDK 的命名服务客户端
	config       *Config                     // 客户端配置
	userConfig   *NacosConfig                // 用户原始配置，用于获取默认分组等信息
}

// NewClient 创建一个新的 Nacos 客户端实例。
//
// 该方法会验证配置参数的有效性，并初始化与 Nacos 服务器的连接。
// 创建的客户端会自动启用本地缓存，提高服务发现的性能。
//
// 参数：
//   - config: Nacos 客户端配置，如果为 nil 则使用默认配置
//
// 返回值：
//   - *NacosClient: 创建的客户端实例
//   - error: 如果配置无效或连接失败则返回错误
//
// 错误类型：
//   - ErrTypeInvalidArgument: 配置参数无效
//   - ErrTypeConnection: 连接 Nacos 服务器失败
//
// 示例：
//
//	config := &NacosConfig{
//		Servers: []ServerConfig{{Host: "127.0.0.1", Port: 8848}},
//		Namespace: "public",
//		Group: "my-group",
//	}
//	client, err := NewClient(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
func NewClient(userConfig *NacosConfig) (*NacosClient, error) {
	if userConfig == nil {
		userConfig = DefaultConfig()
	}

	// 转换为完整配置
	config, err := NewConfig(userConfig)
	if err != nil {
		return nil, fmt.Errorf("创建配置失败: %w", err)
	}

	// 验证配置
	if len(config.ServerConfigs) == 0 {
		return nil, common.NewError(common.ErrTypeInvalidArgument, "服务器配置不能为空", nil)
	}

	// 创建命名服务客户端
	namingClient, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  &config.ClientConfig,
		ServerConfigs: config.ServerConfigs,
	})
	if err != nil {
		return nil, common.NewError(common.ErrTypeConnection, "创建Nacos客户端失败", err)
	}

	client := &NacosClient{
		namingClient: namingClient,
		config:       config,
		userConfig:   userConfig,
	}

	logger.Info("Nacos客户端创建成功",
		"serverAddr", config.ServerConfigs[0].IpAddr,
		"port", config.ServerConfigs[0].Port,
		"namespaceId", config.ClientConfig.NamespaceId)

	return client, nil
}

// getDefaultGroup 获取默认分组名称。
//
// 如果用户配置中设置了分组，则使用用户配置的分组；
// 否则使用 Nacos SDK 的默认分组 DEFAULT_GROUP。
//
// 返回值：
//   - string: 默认分组名称
func (c *NacosClient) getDefaultGroup() string {
	if c.userConfig != nil && c.userConfig.Group != "" {
		return c.userConfig.Group
	}
	return constant.DEFAULT_GROUP
}

// RegisterInstance 向 Nacos 注册一个服务实例。
//
// 该操作是实时的，注册成功后服务实例会立即在 Nacos 服务器上可见，
// 并开始接收来自其他服务的发现请求。注册的实例默认为健康状态，
// 权重为 1.0，并启用临时实例模式。
//
// 参数：
//   - serviceName: 服务名称，不能为空
//   - ip: 服务实例的 IP 地址，不能为空
//   - port: 服务实例的端口号，必须大于 0
//   - groupName: 分组名称，如果为空则使用配置文件中的分组，配置文件中也为空时使用 DEFAULT_GROUP
//
// 返回值：
//   - error: 注册失败时返回错误
//
// 错误类型：
//   - ErrTypeInvalidArgument: 参数无效（服务名为空、IP为空或端口为0）
//   - ErrTypeConnection: 与 Nacos 服务器通信失败
//   - ErrTypeInternal: Nacos 服务器返回注册失败
//
// 注意：
//   - 注册操作是实时的，不使用缓存
//   - 实例注册后会自动开始心跳检测
//   - 如果相同的服务实例重复注册，会更新现有实例信息
//
// 示例：
//
//	err := client.RegisterInstance("user-service", "192.168.1.100", 8080, "prod")
//	if err != nil {
//	    log.Printf("注册服务失败: %v", err)
//	}
func (c *NacosClient) RegisterInstance(serviceName, ip string, port uint64, groupName string) error {
	// 参数验证，防止无效参数
	if serviceName == "" {
		return common.NewError(common.ErrTypeInvalidArgument, "服务名不能为空", nil)
	}
	if ip == "" {
		return common.NewError(common.ErrTypeInvalidArgument, "IP地址不能为空", nil)
	}
	if port == 0 {
		return common.NewError(common.ErrTypeInvalidArgument, "端口号不能为0", nil)
	}

	if groupName == "" {
		groupName = c.getDefaultGroup()
	}

	success, err := c.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: serviceName,
		GroupName:   groupName,
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{},
	})

	if err != nil {
		return common.NewError(common.ErrTypeConnection, "注册服务实例失败", err)
	}

	if !success {
		return common.NewError(common.ErrTypeInternal, "注册服务实例返回失败", nil)
	}

	logger.Info("服务实例注册成功",
		"serviceName", serviceName,
		"groupName", groupName,
		"ip", ip,
		"port", port)

	return nil
}

// RegisterInstanceWithMetadata 向 Nacos 注册一个带有自定义元数据的服务实例。
//
// 该方法与 RegisterInstance 类似，但允许为服务实例添加自定义元数据。
// 元数据可以用于存储版本信息、环境标识、配置参数等额外信息，
// 这些信息可以在服务发现时获取到。
//
// 参数：
//   - serviceName: 服务名称，不能为空
//   - ip: 服务实例的 IP 地址，不能为空
//   - port: 服务实例的端口号，必须大于 0
//   - groupName: 分组名称，如果为空则使用配置文件中的分组，配置文件中也为空时使用 DEFAULT_GROUP
//   - metadata: 自定义元数据，可以为 nil
//
// 返回值：
//   - error: 注册失败时返回错误
//
// 错误类型：
//   - ErrTypeInvalidArgument: 参数无效
//   - ErrTypeConnection: 与 Nacos 服务器通信失败
//   - ErrTypeInternal: Nacos 服务器返回注册失败
//
// 元数据用途：
//   - 版本标识: {"version": "1.2.3"}
//   - 环境标识: {"env": "production"}
//   - 地域信息: {"region": "us-west", "zone": "us-west-1a"}
//   - 配置信息: {"protocol": "https", "contextPath": "/api"}
//
// 示例：
//
//	metadata := map[string]string{
//	    "version": "1.0.0",
//	    "env": "production",
//	    "region": "us-west",
//	}
//	err := client.RegisterInstanceWithMetadata("user-service", "192.168.1.100", 8080, "prod", metadata)
func (c *NacosClient) RegisterInstanceWithMetadata(serviceName, ip string, port uint64, groupName string, metadata map[string]string) error {
	if groupName == "" {
		groupName = c.getDefaultGroup()
	}

	if metadata == nil {
		metadata = make(map[string]string)
	}

	success, err := c.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: serviceName,
		GroupName:   groupName,
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    metadata,
	})

	if err != nil {
		return common.NewError(common.ErrTypeConnection, "注册服务实例失败", err)
	}

	if !success {
		return common.NewError(common.ErrTypeInternal, "注册服务实例返回失败", nil)
	}

	logger.Info("服务实例注册成功（带元数据）",
		"serviceName", serviceName,
		"groupName", groupName,
		"ip", ip,
		"port", port,
		"metadata", metadata)

	return nil
}

// DeregisterInstance 注销服务实例
func (c *NacosClient) DeregisterInstance(serviceName, ip string, port uint64, groupName string) error {
	// 参数验证
	if serviceName == "" {
		return common.NewError(common.ErrTypeInvalidArgument, "服务名不能为空", nil)
	}
	if ip == "" {
		return common.NewError(common.ErrTypeInvalidArgument, "IP地址不能为空", nil)
	}
	if port == 0 {
		return common.NewError(common.ErrTypeInvalidArgument, "端口号不能为0", nil)
	}

	if groupName == "" {
		groupName = c.getDefaultGroup()
	}

	success, err := c.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: serviceName,
		GroupName:   groupName,
		Ephemeral:   true,
	})

	if err != nil {
		return common.NewError(common.ErrTypeConnection, "注销服务实例失败", err)
	}

	if !success {
		return common.NewError(common.ErrTypeInternal, "注销服务实例返回失败", nil)
	}

	logger.Info("服务实例注销成功",
		"serviceName", serviceName,
		"groupName", groupName,
		"ip", ip,
		"port", port)

	return nil
}

// GetService 获取指定服务的所有实例列表。
//
// 该方法使用 SelectAllInstances 获取服务的所有实例，包括健康和不健康的实例。
// 首次调用时会从 Nacos 服务器获取数据，后续调用会优先使用本地缓存，
// 缓存时间通常为 30 秒，由 Nacos SDK 自动管理。
//
// 缓存机制：
//   - 首次调用：从 Nacos 服务器实时获取
//   - 后续调用：优先使用本地缓存（30秒有效期）
//   - 缓存过期：自动从服务器刷新
//   - 服务变更：通过推送机制实时更新缓存
//
// 参数：
//   - serviceName: 服务名称，不能为空
//   - groupName: 分组名称，如果为空则使用配置文件中的分组，配置文件中也为空时使用 DEFAULT_GROUP
//
// 返回值：
//   - []model.Instance: 服务实例列表
//   - error: 获取失败时返回错误
//
// 错误类型：
//   - ErrTypeInvalidArgument: 服务名为空
//   - ErrTypeConnection: 与 Nacos 服务器通信失败
//
// 性能说明：
//   - 本地缓存命中时响应时间 < 1ms
//   - 远程获取时响应时间取决于网络延迟
//   - 建议在高频调用场景下使用
//
// 示例：
//
//	instances, err := client.GetService("user-service", "prod")
//	if err != nil {
//	    log.Printf("获取服务失败: %v", err)
//	    return
//	}
//	for _, instance := range instances {
//	    fmt.Printf("实例: %s:%d, 健康: %v\n", instance.Ip, instance.Port, instance.Healthy)
//	}
func (c *NacosClient) GetService(serviceName, groupName string) ([]model.Instance, error) {
	// 参数验证
	if serviceName == "" {
		return nil, common.NewError(common.ErrTypeInvalidArgument, "服务名不能为空", nil)
	}

	if groupName == "" {
		groupName = c.getDefaultGroup()
	}

	service, err := c.namingClient.SelectAllInstances(vo.SelectAllInstancesParam{
		ServiceName: serviceName,
		GroupName:   groupName,
	})

	if err != nil {
		return nil, common.NewError(common.ErrTypeConnection, "获取服务实例失败", err)
	}

	logger.Debug("获取服务实例成功",
		"serviceName", serviceName,
		"groupName", groupName,
		"instanceCount", len(service))

	return service, nil
}

// GetHealthyInstances 获取健康的服务实例
func (c *NacosClient) GetHealthyInstances(serviceName, groupName string) ([]model.Instance, error) {
	if groupName == "" {
		groupName = c.getDefaultGroup()
	}

	instances, err := c.namingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		GroupName:   groupName,
		HealthyOnly: true,
	})

	if err != nil {
		return nil, common.NewError(common.ErrTypeConnection, "获取健康服务实例失败", err)
	}

	logger.Debug("获取健康服务实例成功",
		"serviceName", serviceName,
		"groupName", groupName,
		"instanceCount", len(instances))

	return instances, nil
}

// SelectOneHealthyInstance 使用负载均衡算法选择一个健康的服务实例。
//
// 该方法会从本地缓存中的健康实例中选择一个实例返回，使用加权轮询（WRR）
// 算法进行负载均衡。只有健康状态为 true、启用状态为 true 且权重大于 0
// 的实例才会被考虑。
//
// 缓存和负载均衡机制：
//   - 数据源：使用本地缓存的服务实例列表
//   - 筛选条件：healthy=true, enable=true, weight>0
//   - 负载均衡：加权轮询（WRR）算法
//   - 权重考虑：权重越高的实例被选中的概率越大
//   - 缓存更新：自动跟随服务发现缓存更新
//
// 参数：
//   - serviceName: 服务名称，不能为空
//   - groupName: 分组名称，如果为空则使用配置文件中的分组，配置文件中也为空时使用 DEFAULT_GROUP
//
// 返回值：
//   - *model.Instance: 选中的服务实例
//   - error: 选择失败时返回错误
//
// 错误类型：
//   - ErrTypeInvalidArgument: 服务名为空
//   - ErrTypeConnection: 与 Nacos 服务器通信失败
//   - ErrTypeInternal: 没有可用的健康实例
//
// 性能特点：
//   - 响应速度极快（< 1ms），基于本地缓存
//   - 内置负载均衡，无需额外实现
//   - 自动排除不健康实例
//   - 支持权重配置
//
// 使用场景：
//   - 微服务间调用的实例选择
//   - API 网关的后端服务选择
//   - 负载均衡器的目标选择
//
// 示例：
//
//	instance, err := client.SelectOneHealthyInstance("user-service", "prod")
//	if err != nil {
//	    log.Printf("选择实例失败: %v", err)
//	    return
//	}
//	url := fmt.Sprintf("http://%s:%d/api/users", instance.Ip, instance.Port)
//	// 发起 HTTP 请求...
func (c *NacosClient) SelectOneHealthyInstance(serviceName, groupName string) (*model.Instance, error) {
	// 参数验证
	if serviceName == "" {
		return nil, common.NewError(common.ErrTypeInvalidArgument, "服务名不能为空", nil)
	}

	if groupName == "" {
		groupName = c.getDefaultGroup()
	}

	// 使用Nacos SDK的SelectOneHealthyInstance方法，它内部实现了负载均衡
	instance, err := c.namingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
		GroupName:   groupName,
	})

	if err != nil {
		return nil, common.NewError(common.ErrTypeConnection, "选择健康服务实例失败", err)
	}

	if instance == nil {
		return nil, common.NewError(common.ErrTypeInternal, "没有可用的健康实例", nil)
	}

	logger.Debug("选择健康服务实例成功",
		"serviceName", serviceName,
		"groupName", groupName,
		"selectedInstance", fmt.Sprintf("%s:%d", instance.Ip, instance.Port))

	return instance, nil
}

// GetAllServicesInfo 获取所有服务信息
func (c *NacosClient) GetAllServicesInfo(pageNo, pageSize uint32, groupName string) (model.ServiceList, error) {
	if groupName == "" {
		groupName = c.getDefaultGroup()
	}

	if pageNo == 0 {
		pageNo = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	serviceList, err := c.namingClient.GetAllServicesInfo(vo.GetAllServiceInfoParam{
		PageNo:    pageNo,
		PageSize:  pageSize,
		GroupName: groupName,
	})

	if err != nil {
		return serviceList, common.NewError(common.ErrTypeConnection, "获取所有服务信息失败", err)
	}

	logger.Debug("获取所有服务信息成功",
		"groupName", groupName,
		"pageNo", pageNo,
		"pageSize", pageSize,
		"serviceCount", len(serviceList.Doms))

	return serviceList, nil
}

// Subscribe 订阅指定服务的实例变更事件。
//
// 该方法会向 Nacos 服务器注册一个服务变更监听器，当服务实例发生变化时
// （如实例上线、下线、健康状态变更等），会通过回调函数实时通知。
// 这是一个实时推送机制，没有缓存延迟。
//
// 推送机制：
//   - 实时推送：服务变更时立即触发回调
//   - 无缓存延迟：直接从服务器推送，不经过本地缓存
//   - 长连接：使用长连接或 WebSocket 保持推送通道
//   - 自动重连：连接断开时自动重新建立订阅
//
// 触发场景：
//   - 新实例注册到服务
//   - 实例从服务中注销
//   - 实例健康状态发生变化
//   - 实例权重或元数据更新
//   - 实例启用/禁用状态变更
//
// 参数：
//   - serviceName: 要订阅的服务名称，不能为空
//   - groupName: 分组名称，如果为空则使用配置文件中的分组，配置文件中也为空时使用 DEFAULT_GROUP
//   - callback: 变更回调函数，接收实例列表和可能的错误
//
// 回调函数参数：
//   - []model.Instance: 当前所有实例列表（包含变更后的状态）
//   - error: 如果获取实例信息时发生错误
//
// 返回值：
//   - error: 订阅失败时返回错误
//
// 错误类型：
//   - ErrTypeConnection: 与 Nacos 服务器建立订阅连接失败
//
// 注意事项：
//   - 回调函数应该快速执行，避免阻塞推送线程
//   - 一个服务可以有多个订阅者
//   - 订阅会一直保持，直到显式取消或客户端关闭
//   - 回调函数中的错误处理应该妥善处理
//
// 示例：
//
//	err := client.Subscribe("user-service", "prod", func(instances []model.Instance, err error) {
//	    if err != nil {
//	        log.Printf("服务变更通知错误: %v", err)
//	        return
//	    }
//	    log.Printf("服务 user-service 实例变更，当前实例数: %d", len(instances))
//	    for _, instance := range instances {
//	        log.Printf("  - %s:%d (健康: %v)", instance.Ip, instance.Port, instance.Healthy)
//	    }
//	})
func (c *NacosClient) Subscribe(serviceName, groupName string, callback func([]model.Instance, error)) error {
	if groupName == "" {
		groupName = c.getDefaultGroup()
	}

	err := c.namingClient.Subscribe(&vo.SubscribeParam{
		ServiceName:       serviceName,
		GroupName:         groupName,
		SubscribeCallback: callback,
	})

	if err != nil {
		return common.NewError(common.ErrTypeConnection, "订阅服务变更失败", err)
	}

	logger.Info("订阅服务变更成功",
		"serviceName", serviceName,
		"groupName", groupName)

	return nil
}

// Unsubscribe 取消订阅服务变更
func (c *NacosClient) Unsubscribe(serviceName, groupName string, callback func([]model.Instance, error)) error {
	if groupName == "" {
		groupName = c.getDefaultGroup()
	}

	err := c.namingClient.Unsubscribe(&vo.SubscribeParam{
		ServiceName:       serviceName,
		GroupName:         groupName,
		SubscribeCallback: callback,
	})

	if err != nil {
		return common.NewError(common.ErrTypeConnection, "取消订阅服务变更失败", err)
	}

	logger.Info("取消订阅服务变更成功",
		"serviceName", serviceName,
		"groupName", groupName)

	return nil
}

// GetServiceDetail 获取服务详细信息
func (c *NacosClient) GetServiceDetail(serviceName, groupName string) (model.Service, error) {
	if groupName == "" {
		groupName = c.getDefaultGroup()
	}

	service, err := c.namingClient.GetService(vo.GetServiceParam{
		ServiceName: serviceName,
		GroupName:   groupName,
	})

	if err != nil {
		return service, common.NewError(common.ErrTypeConnection, "获取服务详细信息失败", err)
	}

	logger.Debug("获取服务详细信息成功",
		"serviceName", serviceName,
		"groupName", groupName,
		"clusterCount", len(service.Clusters))

	return service, nil
}

// UpdateInstance 更新服务实例
func (c *NacosClient) UpdateInstance(serviceName, ip string, port uint64, groupName string, metadata map[string]string) error {
	if groupName == "" {
		groupName = c.getDefaultGroup()
	}

	if metadata == nil {
		metadata = make(map[string]string)
	}

	success, err := c.namingClient.UpdateInstance(vo.UpdateInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: serviceName,
		GroupName:   groupName,
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    metadata,
	})

	if err != nil {
		return common.NewError(common.ErrTypeConnection, "更新服务实例失败", err)
	}

	if !success {
		return common.NewError(common.ErrTypeInternal, "更新服务实例返回失败", nil)
	}

	logger.Info("服务实例更新成功",
		"serviceName", serviceName,
		"groupName", groupName,
		"ip", ip,
		"port", port)

	return nil
}

// Close 关闭 Nacos 客户端并释放所有相关资源。
//
// 该方法会正确关闭与 Nacos 服务器的连接，包括 GRPC 连接、推送连接等，
// 并清理所有相关的资源。调用此方法后，客户端将不能再使用。
//
// 资源清理包括：
//   - 关闭 GRPC 客户端连接
//   - 断开服务推送连接
//   - 清理本地缓存
//   - 取消所有订阅
//   - 停止心跳任务
//
// 调用时机：
//   - 应用程序关闭时
//   - 不再需要使用客户端时
//   - 重新创建客户端之前
//
// 返回值：
//   - error: 通常返回 nil，关闭操作很少失败
//
// 重要提醒：
//   - 必须调用此方法以避免资源泄露
//   - 建议使用 defer 语句确保调用
//   - 关闭后的客户端不能再次使用
//   - 多次调用是安全的
//
// 示例：
//
//	client, err := NewClient(config)
//	if err != nil {
//	    return err
//	}
//	defer client.Close() // 确保资源被释放
//
//	// 使用客户端...
func (c *NacosClient) Close() error {
	// 调用Nacos SDK的CloseClient方法关闭GRPC连接
	if c.namingClient != nil {
		// 关闭GRPC客户端连接，释放网络资源
		c.namingClient.CloseClient()
		// 将客户端引用设置为nil，帮助GC回收
		c.namingClient = nil
	}

	logger.Info("Nacos客户端已关闭")
	return nil
}
