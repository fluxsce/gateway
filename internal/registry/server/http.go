package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gateway/internal/registry/config"
	"gateway/internal/registry/core"

	"github.com/gorilla/mux"
)

// HTTPServer HTTP API服务器
type HTTPServer struct {
	config  *config.HTTPConfig
	manager core.Manager
	server  *http.Server
	router  *mux.Router
}

// NewHTTPServer 创建HTTP服务器
func NewHTTPServer(config *config.HTTPConfig, manager core.Manager) *HTTPServer {
	s := &HTTPServer{
		config:  config,
		manager: manager,
		router:  mux.NewRouter(),
	}

	s.setupRoutes()
	s.setupMiddleware()

	s.server = &http.Server{
		Addr:           fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:        s.router,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		IdleTimeout:    config.IdleTimeout,
		MaxHeaderBytes: config.MaxHeaderBytes,
	}

	return s
}

// Start 启动服务器
func (s *HTTPServer) Start() error {
	fmt.Printf("Starting HTTP server on %s\n", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop 停止服务器
func (s *HTTPServer) Stop(ctx context.Context) error {
	fmt.Println("Stopping HTTP server...")
	return s.server.Shutdown(ctx)
}

// ================== 路由设置 ==================

// setupRoutes 设置路由
func (s *HTTPServer) setupRoutes() {
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// 健康检查
	s.router.HandleFunc("/health", s.healthCheck).Methods("GET")
	s.router.HandleFunc("/ready", s.readinessCheck).Methods("GET")

	// 服务实例管理
	api.HandleFunc("/instances", s.registerInstance).Methods("POST")
	api.HandleFunc("/instances/{instanceId}", s.deregisterInstance).Methods("DELETE")
	api.HandleFunc("/instances/{instanceId}", s.getInstance).Methods("GET")
	api.HandleFunc("/instances/{instanceId}/heartbeat", s.heartbeat).Methods("PUT")
	api.HandleFunc("/instances/{instanceId}/health", s.updateInstanceHealth).Methods("PUT")

	// 服务发现
	api.HandleFunc("/services", s.listServices).Methods("GET")
	api.HandleFunc("/services/{serviceName}/instances", s.discoverInstances).Methods("GET")

	// 服务管理
	api.HandleFunc("/services", s.createService).Methods("POST")
	api.HandleFunc("/services/{serviceName}", s.getService).Methods("GET")
	api.HandleFunc("/services/{serviceName}", s.updateService).Methods("PUT")
	api.HandleFunc("/services/{serviceName}", s.deleteService).Methods("DELETE")

	// 服务分组管理
	api.HandleFunc("/groups", s.listGroups).Methods("GET")
	api.HandleFunc("/groups", s.createGroup).Methods("POST")
	api.HandleFunc("/groups/{groupName}", s.getGroup).Methods("GET")
	api.HandleFunc("/groups/{groupName}", s.updateGroup).Methods("PUT")
	api.HandleFunc("/groups/{groupName}", s.deleteGroup).Methods("DELETE")
	api.HandleFunc("/groups/{groupName}/services", s.listGroupServices).Methods("GET")

	// 事件管理
	api.HandleFunc("/events", s.getEvents).Methods("GET")
	api.HandleFunc("/events/subscribe", s.subscribeEvents).Methods("GET") // WebSocket

	// 统计信息
	api.HandleFunc("/stats", s.getStats).Methods("GET")
	api.HandleFunc("/stats/instances", s.getInstanceStats).Methods("GET")
	api.HandleFunc("/stats/health", s.getHealthStats).Methods("GET")

	// 外部注册中心管理
	api.HandleFunc("/external/configs", s.listExternalConfigs).Methods("GET")
	api.HandleFunc("/external/configs", s.createExternalConfig).Methods("POST")
	api.HandleFunc("/external/configs/{configId}", s.getExternalConfig).Methods("GET")
	api.HandleFunc("/external/configs/{configId}", s.updateExternalConfig).Methods("PUT")
	api.HandleFunc("/external/configs/{configId}", s.deleteExternalConfig).Methods("DELETE")
	api.HandleFunc("/external/configs/{configId}/status", s.getExternalStatus).Methods("GET")
	api.HandleFunc("/external/configs/{configId}/connect", s.connectExternal).Methods("POST")
	api.HandleFunc("/external/configs/{configId}/disconnect", s.disconnectExternal).Methods("POST")
}

// setupMiddleware 设置中间件
func (s *HTTPServer) setupMiddleware() {
	// CORS中间件
	if s.config.EnableCORS {
		s.router.Use(s.corsMiddleware)
	}

	// 请求日志中间件
	if s.config.EnableRequestLog {
		s.router.Use(s.loggingMiddleware)
	}

	// 压缩中间件
	if s.config.EnableGzip {
		s.router.Use(s.gzipMiddleware)
	}

	// 错误处理中间件
	s.router.Use(s.errorMiddleware)
}

// ================== 健康检查接口 ==================

// healthCheck 健康检查
func (s *HTTPServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	status := s.manager.GetHealthStatus()
	s.writeJSON(w, http.StatusOK, status)
}

// readinessCheck 就绪检查
func (s *HTTPServer) readinessCheck(w http.ResponseWriter, r *http.Request) {
	if s.manager.IsRunning() {
		s.writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	} else {
		s.writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
	}
}

// ================== 服务实例管理接口 ==================

// registerInstance 注册服务实例
func (s *HTTPServer) registerInstance(w http.ResponseWriter, r *http.Request) {
	var instance core.ServiceInstance
	if err := s.readJSON(r, &instance); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 设置默认值
	tenantId := s.getTenantId(r)
	instance.TenantId = tenantId
	instance.AddWho = s.getUserId(r)
	instance.EditWho = instance.AddWho

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if err := s.manager.GetRegistry().Register(ctx, &instance); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Register instance failed", err)
		return
	}

	s.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"instanceId": instance.ServiceInstanceId,
		"message":    "Instance registered successfully",
	})
}

// deregisterInstance 注销服务实例
func (s *HTTPServer) deregisterInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceId := vars["instanceId"]
	tenantId := s.getTenantId(r)

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if err := s.manager.GetRegistry().Deregister(ctx, tenantId, instanceId); err != nil {
		if err == core.ErrInstanceNotFound {
			s.writeError(w, http.StatusNotFound, "Instance not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Deregister instance failed", err)
		}
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "Instance deregistered successfully"})
}

// getInstance 获取服务实例
func (s *HTTPServer) getInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceId := vars["instanceId"]
	tenantId := s.getTenantId(r)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	instance, err := s.manager.GetRegistry().GetInstance(ctx, tenantId, instanceId)
	if err != nil {
		if err == core.ErrInstanceNotFound {
			s.writeError(w, http.StatusNotFound, "Instance not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Get instance failed", err)
		}
		return
	}

	s.writeJSON(w, http.StatusOK, instance)
}

// heartbeat 心跳
func (s *HTTPServer) heartbeat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceId := vars["instanceId"]
	tenantId := s.getTenantId(r)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := s.manager.GetRegistry().Heartbeat(ctx, tenantId, instanceId); err != nil {
		if err == core.ErrInstanceNotFound {
			s.writeError(w, http.StatusNotFound, "Instance not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Heartbeat failed", err)
		}
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "Heartbeat successful"})
}

// updateInstanceHealth 更新实例健康状态
func (s *HTTPServer) updateInstanceHealth(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceId := vars["instanceId"]
	tenantId := s.getTenantId(r)

	var req struct {
		HealthStatus string `json:"healthStatus"`
	}
	if err := s.readJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := s.manager.GetRegistry().UpdateHealth(ctx, tenantId, instanceId, req.HealthStatus); err != nil {
		if err == core.ErrInstanceNotFound {
			s.writeError(w, http.StatusNotFound, "Instance not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Update health failed", err)
		}
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "Health status updated successfully"})
}

// ================== 服务发现接口 ==================

// listServices 列出服务
func (s *HTTPServer) listServices(w http.ResponseWriter, r *http.Request) {
	tenantId := s.getTenantId(r)
	groupName := r.URL.Query().Get("groupName")

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	services, err := s.manager.GetRegistry().ListServices(ctx, tenantId, groupName)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "List services failed", err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"services": services,
		"count":    len(services),
	})
}

// discoverInstances 发现服务实例
func (s *HTTPServer) discoverInstances(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["serviceName"]
	tenantId := s.getTenantId(r)
	groupName := r.URL.Query().Get("groupName")

	// 解析过滤器参数
	filters := s.parseInstanceFilters(r)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	instances, err := s.manager.GetRegistry().Discover(ctx, tenantId, serviceName, groupName, filters...)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Discover instances failed", err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"instances": instances,
		"count":     len(instances),
	})
}

// ================== 服务管理接口 ==================

// createService 创建服务
func (s *HTTPServer) createService(w http.ResponseWriter, r *http.Request) {
	var service core.Service
	if err := s.readJSON(r, &service); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 设置默认值
	tenantId := s.getTenantId(r)
	service.TenantId = tenantId
	service.AddWho = s.getUserId(r)
	service.EditWho = service.AddWho
	service.AddTime = time.Now()
	service.EditTime = time.Now()
	service.CurrentVersion = 1
	service.ActiveFlag = core.FlagYes

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := s.manager.GetStorage().SaveService(ctx, &service); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Create service failed", err)
		return
	}

	s.writeJSON(w, http.StatusCreated, map[string]string{"message": "Service created successfully"})
}

// getService 获取服务
func (s *HTTPServer) getService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["serviceName"]
	tenantId := s.getTenantId(r)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	service, err := s.manager.GetStorage().GetService(ctx, tenantId, serviceName)
	if err != nil {
		if err == core.ErrServiceNotFound {
			s.writeError(w, http.StatusNotFound, "Service not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Get service failed", err)
		}
		return
	}

	s.writeJSON(w, http.StatusOK, service)
}

// updateService 更新服务
func (s *HTTPServer) updateService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["serviceName"]
	tenantId := s.getTenantId(r)

	var service core.Service
	if err := s.readJSON(r, &service); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 设置必要字段
	service.TenantId = tenantId
	service.ServiceName = serviceName
	service.EditWho = s.getUserId(r)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := s.manager.GetStorage().SaveService(ctx, &service); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Update service failed", err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "Service updated successfully"})
}

// deleteService 删除服务
func (s *HTTPServer) deleteService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["serviceName"]
	tenantId := s.getTenantId(r)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := s.manager.GetStorage().DeleteService(ctx, tenantId, serviceName); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Delete service failed", err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "Service deleted successfully"})
}

// ================== 服务分组管理接口 ==================

// listGroups 列出服务分组
func (s *HTTPServer) listGroups(w http.ResponseWriter, r *http.Request) {
	tenantId := s.getTenantId(r)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	groups, err := s.manager.GetStorage().ListServiceGroups(ctx, tenantId)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "List groups failed", err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"groups": groups,
		"count":  len(groups),
	})
}

// createGroup 创建服务分组
func (s *HTTPServer) createGroup(w http.ResponseWriter, r *http.Request) {
	var group core.ServiceGroup
	if err := s.readJSON(r, &group); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 设置默认值
	tenantId := s.getTenantId(r)
	group.TenantId = tenantId
	group.AddWho = s.getUserId(r)
	group.EditWho = group.AddWho
	group.AddTime = time.Now()
	group.EditTime = time.Now()
	group.CurrentVersion = 1
	group.ActiveFlag = core.FlagYes

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := s.manager.GetStorage().SaveServiceGroup(ctx, &group); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Create group failed", err)
		return
	}

	s.writeJSON(w, http.StatusCreated, map[string]string{"message": "Group created successfully"})
}

// getGroup 获取服务分组
func (s *HTTPServer) getGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupName := vars["groupName"]
	tenantId := s.getTenantId(r)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	group, err := s.manager.GetStorage().GetServiceGroup(ctx, tenantId, groupName)
	if err != nil {
		if err == core.ErrGroupNotFound {
			s.writeError(w, http.StatusNotFound, "Group not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Get group failed", err)
		}
		return
	}

	s.writeJSON(w, http.StatusOK, group)
}

// updateGroup 更新服务分组
func (s *HTTPServer) updateGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupName := vars["groupName"]
	tenantId := s.getTenantId(r)

	var group core.ServiceGroup
	if err := s.readJSON(r, &group); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 设置必要字段
	group.TenantId = tenantId
	group.GroupName = groupName
	group.EditWho = s.getUserId(r)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := s.manager.GetStorage().SaveServiceGroup(ctx, &group); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Update group failed", err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "Group updated successfully"})
}

// deleteGroup 删除服务分组
func (s *HTTPServer) deleteGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupName := vars["groupName"]
	tenantId := s.getTenantId(r)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := s.manager.GetStorage().DeleteServiceGroup(ctx, tenantId, groupName); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Delete group failed", err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "Group deleted successfully"})
}

// listGroupServices 列出分组下的服务
func (s *HTTPServer) listGroupServices(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupName := vars["groupName"]
	tenantId := s.getTenantId(r)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	services, err := s.manager.GetStorage().ListServices(ctx, tenantId, groupName)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "List group services failed", err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"services": services,
		"count":    len(services),
	})
}

// ================== 事件管理接口 ==================

// getEvents 获取事件列表
func (s *HTTPServer) getEvents(w http.ResponseWriter, r *http.Request) {
	tenantId := s.getTenantId(r)

	// 解析查询参数
	filters := s.parseEventFilters(r)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	events, err := s.manager.GetStorage().GetEvents(ctx, tenantId, filters...)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Get events failed", err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
		"count":  len(events),
	})
}

// subscribeEvents 订阅事件（WebSocket）
func (s *HTTPServer) subscribeEvents(w http.ResponseWriter, r *http.Request) {
	// TODO: 实现WebSocket事件订阅
	s.writeError(w, http.StatusNotImplemented, "WebSocket subscription not implemented", nil)
}

// ================== 统计信息接口 ==================

// getStats 获取统计信息
func (s *HTTPServer) getStats(w http.ResponseWriter, r *http.Request) {
	stats := s.manager.GetStats()
	s.writeJSON(w, http.StatusOK, stats)
}

// getInstanceStats 获取实例统计信息
func (s *HTTPServer) getInstanceStats(w http.ResponseWriter, r *http.Request) {
	tenantId := s.getTenantId(r)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// 这里需要实现获取实例统计的方法
	// 暂时返回基本统计信息
	storageStats, err := s.manager.GetStorage().GetStats(ctx, tenantId)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Get instance stats failed", err)
		return
	}

	s.writeJSON(w, http.StatusOK, storageStats)
}

// getHealthStats 获取健康检查统计信息
func (s *HTTPServer) getHealthStats(w http.ResponseWriter, r *http.Request) {
	stats := s.manager.GetHealthChecker().GetStats()
	s.writeJSON(w, http.StatusOK, stats)
}

// ================== 外部注册中心管理接口 ==================

// listExternalConfigs 列出外部配置
func (s *HTTPServer) listExternalConfigs(w http.ResponseWriter, r *http.Request) {
	tenantId := s.getTenantId(r)
	registryType := r.URL.Query().Get("registryType")
	environment := r.URL.Query().Get("environment")

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	configs, err := s.manager.GetExternalStorage().ListExternalConfigs(ctx, tenantId, registryType, environment)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "List external configs failed", err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"configs": configs,
		"count":   len(configs),
	})
}

// createExternalConfig 创建外部配置
func (s *HTTPServer) createExternalConfig(w http.ResponseWriter, r *http.Request) {
	var config core.ExternalRegistryConfig
	if err := s.readJSON(r, &config); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 设置默认值
	tenantId := s.getTenantId(r)
	config.TenantId = tenantId
	config.AddWho = s.getUserId(r)
	config.EditWho = config.AddWho
	config.AddTime = time.Now()
	config.EditTime = time.Now()
	config.CurrentVersion = 1
	config.ActiveFlag = core.FlagYes

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := s.manager.GetExternalStorage().SaveExternalConfig(ctx, &config); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Create external config failed", err)
		return
	}

	s.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"configId": config.ExternalConfigId,
		"message":  "External config created successfully",
	})
}

// getExternalConfig 获取外部配置
func (s *HTTPServer) getExternalConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	configId := vars["configId"]
	tenantId := s.getTenantId(r)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	config, err := s.manager.GetExternalStorage().GetExternalConfig(ctx, tenantId, configId)
	if err != nil {
		if err == core.ErrConfigNotFound {
			s.writeError(w, http.StatusNotFound, "External config not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Get external config failed", err)
		}
		return
	}

	s.writeJSON(w, http.StatusOK, config)
}

// updateExternalConfig 更新外部配置
func (s *HTTPServer) updateExternalConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	configId := vars["configId"]
	tenantId := s.getTenantId(r)

	var config core.ExternalRegistryConfig
	if err := s.readJSON(r, &config); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 设置必要字段
	config.TenantId = tenantId
	config.ExternalConfigId = configId
	config.EditWho = s.getUserId(r)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := s.manager.GetExternalStorage().SaveExternalConfig(ctx, &config); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Update external config failed", err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "External config updated successfully"})
}

// deleteExternalConfig 删除外部配置
func (s *HTTPServer) deleteExternalConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	configId := vars["configId"]
	tenantId := s.getTenantId(r)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := s.manager.GetExternalStorage().DeleteExternalConfig(ctx, tenantId, configId); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Delete external config failed", err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "External config deleted successfully"})
}

// getExternalStatus 获取外部状态
func (s *HTTPServer) getExternalStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	configId := vars["configId"]
	tenantId := s.getTenantId(r)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	status, err := s.manager.GetExternalStorage().GetExternalStatus(ctx, tenantId, configId)
	if err != nil {
		if err == core.ErrConfigNotFound {
			s.writeError(w, http.StatusNotFound, "External status not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Get external status failed", err)
		}
		return
	}

	s.writeJSON(w, http.StatusOK, status)
}

// connectExternal 连接外部注册中心
func (s *HTTPServer) connectExternal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	configId := vars["configId"]
	tenantId := s.getTenantId(r)

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// 获取配置
	config, err := s.manager.GetExternalStorage().GetExternalConfig(ctx, tenantId, configId)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "External config not found", err)
		return
	}

	// 连接外部注册中心
	if err := s.manager.GetExternalStorage().Connect(ctx, config); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Connect external registry failed", err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "Connected to external registry successfully"})
}

// disconnectExternal 断开外部注册中心连接
func (s *HTTPServer) disconnectExternal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	configId := vars["configId"]

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := s.manager.GetExternalStorage().Disconnect(ctx, configId); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Disconnect external registry failed", err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "Disconnected from external registry successfully"})
}

// ================== 辅助方法 ==================

// readJSON 读取JSON请求体
func (s *HTTPServer) readJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// writeJSON 写入JSON响应
func (s *HTTPServer) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError 写入错误响应
func (s *HTTPServer) writeError(w http.ResponseWriter, status int, message string, err error) {
	response := map[string]interface{}{
		"error":  message,
		"status": status,
		"time":   time.Now().Format(time.RFC3339),
	}

	if err != nil {
		response["details"] = err.Error()
	}

	s.writeJSON(w, status, response)
}

// getTenantId 获取租户ID
func (s *HTTPServer) getTenantId(r *http.Request) string {
	// 从请求头获取租户ID
	tenantId := r.Header.Get("X-Tenant-Id")
	if tenantId == "" {
		// 从查询参数获取
		tenantId = r.URL.Query().Get("tenantId")
	}
	if tenantId == "" {
		// 默认租户ID
		tenantId = "default"
	}
	return tenantId
}

// getUserId 获取用户ID
func (s *HTTPServer) getUserId(r *http.Request) string {
	// 从请求头获取用户ID
	userId := r.Header.Get("X-User-Id")
	if userId == "" {
		// 从查询参数获取
		userId = r.URL.Query().Get("userId")
	}
	if userId == "" {
		// 默认用户ID
		userId = "system"
	}
	return userId
}

// parseInstanceFilters 解析实例过滤器
func (s *HTTPServer) parseInstanceFilters(r *http.Request) []core.InstanceFilter {
	params := make(map[string]string)

	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	return core.ParseInstanceFilters(params)
}

// parseEventFilters 解析事件过滤器
func (s *HTTPServer) parseEventFilters(r *http.Request) []core.EventFilter {
	var filters []core.EventFilter

	// 事件类型过滤器
	if eventType := r.URL.Query().Get("eventType"); eventType != "" {
		filters = append(filters, core.NewEventTypeFilter(eventType))
	}

	// 服务名过滤器
	if serviceName := r.URL.Query().Get("serviceName"); serviceName != "" {
		filters = append(filters, core.NewEventServiceFilter(serviceName))
	}

	// 时间范围过滤器
	if startTimeStr := r.URL.Query().Get("startTime"); startTimeStr != "" {
		if endTimeStr := r.URL.Query().Get("endTime"); endTimeStr != "" {
			if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
				if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
					filters = append(filters, core.NewEventTimeRangeFilter(startTime, endTime))
				}
			}
		}
	}

	return filters
}

// ================== 中间件 ==================

// corsMiddleware CORS中间件
func (s *HTTPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cors := s.config.CORS

		// 设置CORS头
		if len(cors.AllowedOrigins) > 0 {
			origin := r.Header.Get("Origin")
			for _, allowedOrigin := range cors.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
					break
				}
			}
		}

		if len(cors.AllowedMethods) > 0 {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(cors.AllowedMethods, ", "))
		}

		if len(cors.AllowedHeaders) > 0 {
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(cors.AllowedHeaders, ", "))
		}

		if len(cors.ExposedHeaders) > 0 {
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(cors.ExposedHeaders, ", "))
		}

		if cors.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if cors.MaxAge > 0 {
			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(int(cors.MaxAge.Seconds())))
		}

		// 处理预检请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware 请求日志中间件
func (s *HTTPServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 创建响应记录器
		recorder := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		// 处理请求
		next.ServeHTTP(recorder, r)

		// 记录日志
		duration := time.Since(start)
		fmt.Printf("[%s] %s %s %d %v\n",
			start.Format("2006-01-02 15:04:05"),
			r.Method,
			r.RequestURI,
			recorder.statusCode,
			duration,
		)
	})
}

// gzipMiddleware 压缩中间件
func (s *HTTPServer) gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 简单的压缩实现，实际项目中可以使用更完善的压缩中间件
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
		}
		next.ServeHTTP(w, r)
	})
}

// errorMiddleware 错误处理中间件
func (s *HTTPServer) errorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("Panic in HTTP handler: %v\n", err)
				s.writeError(w, http.StatusInternalServerError, "Internal server error", fmt.Errorf("%v", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// responseRecorder 响应记录器
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
