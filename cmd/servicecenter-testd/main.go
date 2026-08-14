// servicecenter-testd 启动基于 SQLite 的服务中心测试守护进程，供 Java SDK E2E 连接。
//
// 用法：
//
//	go run ./cmd/servicecenter-testd
//
// 标准输出就绪信号示例：
//
//	SERVICE_CENTER_READY host=127.0.0.1 port=xxxxx namespace=ns_e2e_xxx group=e2e-group auth=off tls=off
//
// 环境变量：
//   - SC_E2E_NAMESPACE / SC_E2E_GROUP
//   - SC_E2E_ENABLE_AUTH：开启 Basic/API Token/JWT
//   - SC_E2E_ENABLE_TLS：开启单向 TLS（自动生成自签证书）
//   - SC_E2E_ENABLE_MTLS：开启 mTLS（隐含 TLS）
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"gateway/internal/servicecenter/testutil"
)

func main() {
	ctx := context.Background()
	enableAuth := truthy(os.Getenv("SC_E2E_ENABLE_AUTH"))
	enableTLS := truthy(os.Getenv("SC_E2E_ENABLE_TLS"))
	enableMTLS := truthy(os.Getenv("SC_E2E_ENABLE_MTLS"))

	env, err := testutil.Start(ctx, testutil.Options{
		NamespaceID: os.Getenv("SC_E2E_NAMESPACE"),
		GroupName:   os.Getenv("SC_E2E_GROUP"),
		EnableAuth:  enableAuth,
		EnableTLS:   enableTLS,
		EnableMTLS:  enableMTLS,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "start service center failed: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = env.Stop(context.Background())
	}()

	// 字段顺序：auth 细节紧跟 auth=on，tls 细节紧跟 tls=，保持既有 Java Auth 正则兼容
	ready := fmt.Sprintf("SERVICE_CENTER_READY host=%s port=%d namespace=%s group=%s",
		env.Host, env.Port, env.NamespaceID, env.GroupName)
	if env.EnableAuth {
		ready += fmt.Sprintf(" auth=on user=%s password=%s token=%s jwtSecret=%s jwtIssuer=%s",
			env.AuthUserID, env.AuthPassword, env.AuthAPIToken, env.AuthJwtSecret, env.AuthJwtIssuer)
	} else {
		ready += " auth=off"
	}
	if env.EnableMTLS {
		ready += fmt.Sprintf(" tls=mtls ca=%s clientCert=%s clientKey=%s",
			env.TLS.CACertFile, env.TLS.ClientCert, env.TLS.ClientKey)
	} else if env.EnableTLS {
		ready += fmt.Sprintf(" tls=on ca=%s", env.TLS.CACertFile)
	} else {
		ready += " tls=off"
	}
	fmt.Println(ready)
	fmt.Fprintf(os.Stderr, "service-center testd listening on %s (auth=%v tls=%v mtls=%v), ctrl+c to stop\n",
		env.Endpoint(), env.EnableAuth, env.EnableTLS, env.EnableMTLS)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	fmt.Fprintln(os.Stderr, "shutting down...")
}

func truthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
