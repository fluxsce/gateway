package testutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// TLSMaterial 测试用 TLS 证书材料（PEM 文件路径）。
type TLSMaterial struct {
	Dir        string
	CACertFile string
	CAKeyFile  string
	ServerCert string
	ServerKey  string
	ClientCert string
	ClientKey  string
	CACertPEM  []byte
}

// GenerateTLSMaterial 在 dir 下生成自签 CA、服务端证书与客户端证书（含 127.0.0.1 / localhost SAN）。
// 私钥为 PKCS#8 PEM，便于 Java Netty 加载。
func GenerateTLSMaterial(dir string) (*TLSMaterial, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("创建证书目录失败: %w", err)
	}

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	now := time.Now().Add(-time.Minute)
	serial := func() *big.Int {
		n, _ := rand.Int(rand.Reader, big.NewInt(1<<62))
		return n
	}

	caTpl := &x509.Certificate{
		SerialNumber:          serial(),
		Subject:               pkix.Name{CommonName: "ServiceCenter E2E CA", Organization: []string{"Flux E2E"}},
		NotBefore:             now,
		NotAfter:              now.Add(24 * time.Hour * 365),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTpl, caTpl, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("创建 CA 证书失败: %w", err)
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		return nil, err
	}

	serverTpl := &x509.Certificate{
		SerialNumber: serial(),
		Subject:      pkix.Name{CommonName: "service-center-e2e", Organization: []string{"Flux E2E"}},
		NotBefore:    now,
		NotAfter:     now.Add(24 * time.Hour * 365),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost", "127.0.0.1"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}
	serverDER, err := x509.CreateCertificate(rand.Reader, serverTpl, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("创建服务端证书失败: %w", err)
	}

	clientTpl := &x509.Certificate{
		SerialNumber: serial(),
		Subject:      pkix.Name{CommonName: "service-center-e2e-client", Organization: []string{"Flux E2E"}},
		NotBefore:    now,
		NotAfter:     now.Add(24 * time.Hour * 365),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	clientDER, err := x509.CreateCertificate(rand.Reader, clientTpl, caCert, &clientKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("创建客户端证书失败: %w", err)
	}

	m := &TLSMaterial{
		Dir:        dir,
		CACertFile: filepath.Join(dir, "ca.crt"),
		CAKeyFile:  filepath.Join(dir, "ca.key"),
		ServerCert: filepath.Join(dir, "server.crt"),
		ServerKey:  filepath.Join(dir, "server.key"),
		ClientCert: filepath.Join(dir, "client.crt"),
		ClientKey:  filepath.Join(dir, "client.key"),
	}

	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})
	m.CACertPEM = caPEM

	if err := writePEM(m.CACertFile, "CERTIFICATE", caDER); err != nil {
		return nil, err
	}
	if err := writePKCS8Key(m.CAKeyFile, caKey); err != nil {
		return nil, err
	}
	if err := writePEM(m.ServerCert, "CERTIFICATE", serverDER); err != nil {
		return nil, err
	}
	if err := writePKCS8Key(m.ServerKey, serverKey); err != nil {
		return nil, err
	}
	if err := writePEM(m.ClientCert, "CERTIFICATE", clientDER); err != nil {
		return nil, err
	}
	if err := writePKCS8Key(m.ClientKey, clientKey); err != nil {
		return nil, err
	}
	return m, nil
}

func writePEM(path, typ string, der []byte) error {
	return os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: der}), 0o644)
}

func writePKCS8Key(path string, key *ecdsa.PrivateKey) error {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return err
	}
	return os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), 0o600)
}
