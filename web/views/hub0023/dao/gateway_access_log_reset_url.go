package dao

import (
	"context"
	"net"
	"net/url"
	"strconv"
	"strings"

	"gateway/pkg/config"
	"gateway/pkg/database"
	hub0020dao "gateway/web/views/hub0020/dao"
	hub0020models "gateway/web/views/hub0020/models"
	"gateway/web/views/hub0023/models"
)

// FillGatewayAccessLogResetURL 根据 HUB_GW_INSTANCE 的本机绑定地址（bindAddress）与 HTTP/HTTPS 监听端口拼装 ResetUrl，
// 与日志中的 requestPath、requestQuery 组合。端口始终取自实例配置，不使用请求 Host 头中的端口。
// bindAddress 为 0.0.0.0 或空时表示监听所有网卡，此时使用 config.GetNodeIP()（与 random 包启动时探测的节点 IPv4 一致）；仍为空时再回退 gatewayNodeIp。
// 仅用于 API 响应填充，不写入 HUB_GW_ACCESS_LOG；Mongo/ClickHouse 日志查询路径需传入同一套关系库连接以查实例表。
func FillGatewayAccessLogResetURL(ctx context.Context, db database.Database, log *models.GatewayAccessLog) {
	if db == nil || log == nil {
		return
	}
	gatewayInstanceID := strings.TrimSpace(log.GatewayInstanceId)
	tenantID := strings.TrimSpace(log.TenantId)
	if gatewayInstanceID == "" || tenantID == "" {
		return
	}

	instDAO := hub0020dao.NewGatewayInstanceDAO(db)
	inst, err := instDAO.GetGatewayInstanceById(ctx, gatewayInstanceID, tenantID)
	if err != nil || inst == nil {
		return
	}

	scheme, port, ok := pickGatewayInstanceHTTPSchemeAndPort(inst)
	if !ok {
		return
	}

	host := pickResetURLHostFromInstance(log, inst)
	if host == "" {
		return
	}

	path := strings.TrimSpace(log.RequestPath)
	if path == "" {
		path = "/"
	} else if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	u := &url.URL{
		Scheme: scheme,
		Path:   path,
	}
	rawQ := strings.TrimSpace(log.RequestQuery)
	rawQ = strings.TrimPrefix(rawQ, "?")
	if rawQ != "" {
		u.RawQuery = rawQ
	}

	// 标准端口省略显示，与常见浏览器行为一致
	if (scheme == "http" && port == 80) || (scheme == "https" && port == 443) {
		u.Host = host
	} else {
		u.Host = net.JoinHostPort(host, strconv.Itoa(port))
	}

	log.ResetUrl = u.String()
}

// pickResetURLHostFromInstance 优先使用实例绑定的监听 IP（bindAddress）；为 0.0.0.0 或空时使用 config.GetNodeIP()，最后再回退 gatewayNodeIp。
func pickResetURLHostFromInstance(log *models.GatewayAccessLog, inst *hub0020models.GatewayInstance) string {
	if inst != nil {
		bind := strings.TrimSpace(inst.BindAddress)
		if bind != "" && bind != "0.0.0.0" {
			return bind
		}
	}
	if node := strings.TrimSpace(config.GetNodeIP()); node != "" {
		return node
	}
	return strings.TrimSpace(log.GatewayNodeIp)
}

// pickGatewayInstanceHTTPSchemeAndPort 按 TLS 与端口配置选择对外 HTTP(S) 方案。
func pickGatewayInstanceHTTPSchemeAndPort(inst *hub0020models.GatewayInstance) (scheme string, port int, ok bool) {
	if inst == nil {
		return "", 0, false
	}
	if inst.TlsEnabled == "Y" && inst.HttpsPort != nil && *inst.HttpsPort > 0 {
		return "https", *inst.HttpsPort, true
	}
	if inst.HttpPort != nil && *inst.HttpPort > 0 {
		return "http", *inst.HttpPort, true
	}
	if inst.HttpsPort != nil && *inst.HttpsPort > 0 {
		return "https", *inst.HttpsPort, true
	}
	return "", 0, false
}
