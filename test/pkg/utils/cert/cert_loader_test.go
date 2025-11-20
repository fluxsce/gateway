package cert

import (
	"crypto/tls"
	"os"
	"path/filepath"
	"testing"

	"gateway/pkg/utils/cert"
)

// 测试用的自签名证书和私钥（PEM格式）
const (
	testCertPEM = `-----BEGIN CERTIFICATE-----
MIID3zCCAsegAwIBAgIJALmoG9obpWhkMA0GCSqGSIb3DQEBCwUAMIGFMQswCQYD
VQQGEwJDTjEOMAwGA1UECAwFSFVCRUkxDjAMBgNVBAcMBVdVSEFOMQ0wCwYDVQQK
DARGTFVYMRAwDgYDVQQLDAdEQVRBSFVCMRIwEAYDVQQDDAkxMjcuMC4wLjExITAf
BgkqhkiG9w0BCQEWEnNoYW5namlhbkBmbHV4LmNvbTAeFw0yNTExMjAwNDE4MzFa
Fw0yNjExMjAwNDE4MzFaMIGFMQswCQYDVQQGEwJDTjEOMAwGA1UECAwFSFVCRUkx
DjAMBgNVBAcMBVdVSEFOMQ0wCwYDVQQKDARGTFVYMRAwDgYDVQQLDAdEQVRBSFVC
MRIwEAYDVQQDDAkxMjcuMC4wLjExITAfBgkqhkiG9w0BCQEWEnNoYW5namlhbkBm
bHV4LmNvbTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKQPI0DGrljk
i04bj+7rtaMUEYPd5oH6dfZrKe1mA1vi73AtLLDIXDAIPWKJE9J9R4Am0hRQwyyT
cQfVAOngWvv2GyRsfrHb6+6rEBH9Z6mKvYa8kRY6/mKUN6hg9GdAkPecw1J9/eXr
dNWXrnVARiQx2ijhXVcLhOIxFdYROChV5Yey9uRsIxVZIsJyhbWCigaiwA9SnUal
MBeA3GtLT/VdYTGseHPV9SdxVU9dwbCrOuzGD4TP86LbI85CcGlL9xDAZVroDzdD
/Vrr05salB8ZIT7xkcBYKt4RMf7kQl7+LfPTbeZn59gUjblJy+nj62Fea6kp1t0t
UwX66Ra3ptUCAwEAAaNQME4wHQYDVR0OBBYEFKZbx2F3nM1afySmTFA5Vd8O0heH
MB8GA1UdIwQYMBaAFKZbx2F3nM1afySmTFA5Vd8O0heHMAwGA1UdEwQFMAMBAf8w
DQYJKoZIhvcNAQELBQADggEBAFu2UAw/Kdqdiiv0BN9pgyQ5GeNT1EGquoNNwEg0
Cl2BKOP00KvFOo9w7TxJroU5pnE3rPPgYeL8jdbkgJuPROCHhfKQZ1whqp8L4xLJ
j9jAGa1iJcIuue1yqpNksTeflMD5UCqMFnKtaEOG50SV9VEZ8kMQ9Vi1WEfAsIBL
LXUxgMjYoHvD5TnQ8dWYd5225+AgITndbc8nFePOZSMxg4ArQciQTYNQw844Uv62
myMD7i5cPGJf/fon5g2xoGrp0dxIY4F21zm0e14DpSuw/5o4RlK5DL1fQnjSyz73
syGvjtlf3jxqqoFEIHuWvIukNXU7EzEzmhyJLqEcNN45hX0=
-----END CERTIFICATE-----`

	testKeyPEM = `-----BEGIN ENCRYPTED PRIVATE KEY-----
MIIFDjBABgkqhkiG9w0BBQ0wMzAbBgkqhkiG9w0BBQwwDgQIeqqa4y6PmEsCAggA
MBQGCCqGSIb3DQMHBAgLC6YyiIJ8UQSCBMhxua+9rSUW3S0d8XDT1XtMccIOqgt/
MxSfWqj/+vHZQCK1e1n2uzZuGyA19w4tHaEIKDo5cGYyxk69kp156BxxcpJA5s9Q
28GFEVhfx/suEokdUNLjbYf5Dft6H/sWnKPyhjN0ZWAPCTkkOMiXJkdYzTKjtmZC
HkwO+tj73R9VwcvWZA+cRJ24HBWvxDvj09jVdv/fRm4cagSre1qDbhKVLQluDuxy
G9WW7iM+MjiieC5sONGGLKB4hCbmflMgZQiE1b3XQKerLAhz8bW2KvI9I5a+K99k
By7rtiobkA7bG82YQMr+kTazcem3J+Lu97ptcqWNZb9fMGScp90DluyuzxGlTpeD
WKg400nrVFRP3jS8O89jLrDB282vDBrym91N1IPVVWZ4LYPHTNsrCrLDwfmhS7Go
s8uJVW32k2jYYfoc0dJJM5L/vfMMDixr85rxS7/GjKsVmtYuy8Z87NYBy7ABY6Kh
CKJjUG7LXsJrEWe8AVYhHmoZEvBHpovriwSOnl/ib6p5D+s2PVIQuhvK3xtqzh6M
LIqCSqJ+9c54Yt0pYsJNn9/z0CBT5IZ2b+wrhSkRjoasf4TsL7chr62HY+ZN70pY
e+rUqUzFhPK3A4AVqNOPH5eUWyKitylJU8MsAwcfu5mPoHyJvaXakHuXsKZaVWBW
/hlNsneEDLeCVp8FKIiqlsaYo/SnIyDY18dd81h1wAFcIg2m3epnjnHGPvIXUWM4
mBv4Li403d5ZCbc/9X8Z/EGFCb0kQdNPMA9Mvx2hdv2TjfR6zyrh6+ImZbC517FU
h1XMtE+yIF3jJl6zt9qZ3PG9p2EsmrI1vUg3kViy2quaFBEm7y/43c2cEN5t2rA6
dRzvsFGqi42UsWYPrxAzU8nn4kh4pcDUJkmnJ5ICrWy4e79G6vThaepWtGyFTAoW
1+tas+jKb50ZYs0zvA/TpnRsAWLcriqrpof6ZIkJuU0mJiP2lXKn7t2HiD6zMUED
PxzyJbwtGSPLX1cnLIxznSWg/WEoESgGpZizmJeFdJP4Ua01DEtoC6x1fcKDuRDe
4qf7xQaRy+f1M/ip8VEoTTcgm8BbTcmVRL+zEi7mlxkX7Re8XF3SX+PUA0vrgs96
xZdlXLiVn4GNTkdqxRWOCNsvwB19roHatRSthgisM89uWhQp7xZqjxMER/f9HFM5
/Jl6827eijUdLWFuv4pjGzHcdUdj9OznsL08AlmM2LWB64mlIqNIYX+n+o1B7/Zk
gETtwPGO9TdyTD5EH5fxuU/Jo+KcXkYlqCOG4gKdrf/Vs0in6iQJVpikDLD7msIo
QLZUh+DWGJ9KKaCrYTjr+YtqUod/Zk8YYV4MfQYbvUOyuSbKON6PYqyqF0y4I44Q
IeU4jeMyl+xn9eujayGj18524cEAZQVa2+E6nZvQKrE7j4YtXqcJRPX0SfaY8HOK
uZVUkKn17OJ4a0AFv/6qgY+HPtoJwUziJeLW4dXBHbCDOFAj7CFxocsWfZOj2xmh
e7Gm/tsXaElLOKUaCfwxUrgBC0bAwCO8HnvzZ3qXRctFcbsE8GVjydTe9rOqa5wM
FbDe7Io2dMI7qw2bVHYEH6l/FkwW1LF9J7QnU9VbJzTtK2+sMqgfNh8qykwURjPS
mrs=
-----END ENCRYPTED PRIVATE KEY-----`
)

func TestNewCertLoader(t *testing.T) {
	config := &cert.CertConfig{
		CertFile: "test.crt",
		KeyFile:  "test.key",
	}

	loader := cert.NewCertLoader(config)
	if loader == nil {
		t.Fatal("NewCertLoader returned nil")
	}
}

func TestParseTLSVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected uint16
	}{
		// 标准格式
		{"TLS1.0", tls.VersionTLS10},
		{"TLS1.1", tls.VersionTLS11},
		{"TLS1.2", tls.VersionTLS12},
		{"TLS1.3", tls.VersionTLS13},
		// 小写
		{"tls1.2", tls.VersionTLS12},
		{"tls1.3", tls.VersionTLS13},
		// 带v前缀
		{"TLSv1.2", tls.VersionTLS12},
		{"v1.2", tls.VersionTLS12},
		// 简短格式
		{"1.2", tls.VersionTLS12},
		{"1.3", tls.VersionTLS13},
		// 无效格式
		{"unknown", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := cert.ParseTLSVersion(tt.input)
			if result != tt.expected {
				t.Errorf("ParseTLSVersion(%s) = 0x%04x, want 0x%04x", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseCipherSuites(t *testing.T) {
	suites := []string{
		"TLS_RSA_WITH_AES_128_CBC_SHA",
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		"TLS_AES_128_GCM_SHA256",
		"INVALID_SUITE", // 应该被忽略
	}

	result := cert.ParseCipherSuites(suites)

	// 应该返回3个有效的加密套件
	if len(result) != 3 {
		t.Errorf("ParseCipherSuites returned %d suites, want 3", len(result))
	}

	// 验证第一个套件
	if result[0] != tls.TLS_RSA_WITH_AES_128_CBC_SHA {
		t.Error("First cipher suite not parsed correctly")
	}
}

func TestParseCipherSuites_CaseInsensitive(t *testing.T) {
	suites := []string{
		"tls_rsa_with_aes_128_cbc_sha",          // 小写
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256", // 大写
	}

	result := cert.ParseCipherSuites(suites)

	if len(result) != 2 {
		t.Errorf("ParseCipherSuites returned %d suites, want 2", len(result))
	}
}

func TestLoadCertificate_FileNotExist(t *testing.T) {
	config := &cert.CertConfig{
		CertFile: "/nonexistent/cert.pem",
		KeyFile:  "/nonexistent/key.pem",
	}

	loader := cert.NewCertLoader(config)
	_, err := loader.LoadCertificate()

	if err == nil {
		t.Error("Expected error for non-existent files, got nil")
	}
}

func TestLoadCertificate_Success(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "certtest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 写入测试证书和私钥
	certFile := filepath.Join(tmpDir, "test.crt")
	keyFile := filepath.Join(tmpDir, "test.key")

	if err := os.WriteFile(certFile, []byte(testCertPEM), 0600); err != nil {
		t.Fatalf("Failed to write cert file: %v", err)
	}

	if err := os.WriteFile(keyFile, []byte(testKeyPEM), 0600); err != nil {
		t.Fatalf("Failed to write key file: %v", err)
	}

	// 测试加载证书
	config := &cert.CertConfig{
		CertFile:    certFile,
		KeyFile:     keyFile,
		KeyPassword: "123456",
	}

	loader := cert.NewCertLoader(config)
	certificate, err := loader.LoadCertificate()

	if err != nil {
		t.Fatalf("LoadCertificate failed: %v", err)
	}

	if certificate == nil {
		t.Fatal("Certificate is nil")
	}

	if len(certificate.Certificate) == 0 {
		t.Error("Certificate chain is empty")
	}
}

func TestCreateTLSConfig_WithDefaultSettings(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "certtest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 写入测试证书和私钥
	certFile := filepath.Join(tmpDir, "test.crt")
	keyFile := filepath.Join(tmpDir, "test.key")

	if err := os.WriteFile(certFile, []byte(testCertPEM), 0600); err != nil {
		t.Fatalf("Failed to write cert file: %v", err)
	}

	if err := os.WriteFile(keyFile, []byte(testKeyPEM), 0600); err != nil {
		t.Fatalf("Failed to write key file: %v", err)
	}

	// 测试创建TLS配置（使用默认设置）
	config := &cert.CertConfig{
		CertFile:    certFile,
		KeyFile:     keyFile,
		KeyPassword: "123456",
		// 不配置TLSVersions和CipherSuites，使用默认值
	}

	loader := cert.NewCertLoader(config)
	tlsConfig, err := loader.CreateTLSConfig()

	if err != nil {
		t.Fatalf("CreateTLSConfig failed: %v", err)
	}

	if tlsConfig == nil {
		t.Fatal("TLS config is nil")
	}

	if len(tlsConfig.Certificates) == 0 {
		t.Error("TLS config has no certificates")
	}

	// 验证默认最小版本为TLS 1.2
	if tlsConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = 0x%04x, want 0x%04x (TLS 1.2)", tlsConfig.MinVersion, tls.VersionTLS12)
	}

	// 验证使用了默认加密套件（nil表示使用Go的默认安全套件）
	if tlsConfig.CipherSuites != nil {
		t.Error("CipherSuites should be nil to use Go's default secure cipher suites")
	}
}

func TestCreateTLSConfig_WithMultipleTLSVersions(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "certtest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 写入测试证书和私钥
	certFile := filepath.Join(tmpDir, "test.crt")
	keyFile := filepath.Join(tmpDir, "test.key")

	if err := os.WriteFile(certFile, []byte(testCertPEM), 0600); err != nil {
		t.Fatalf("Failed to write cert file: %v", err)
	}

	if err := os.WriteFile(keyFile, []byte(testKeyPEM), 0600); err != nil {
		t.Fatalf("Failed to write key file: %v", err)
	}

	// 测试创建TLS配置（配置多个TLS版本）
	config := &cert.CertConfig{
		CertFile:    certFile,
		KeyFile:     keyFile,
		KeyPassword: "123456",
		TLSVersions: []string{"TLS1.2", "TLS1.3"}, // 支持TLS 1.2和1.3
	}

	loader := cert.NewCertLoader(config)
	tlsConfig, err := loader.CreateTLSConfig()

	if err != nil {
		t.Fatalf("CreateTLSConfig failed: %v", err)
	}

	// 验证版本范围
	if tlsConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = 0x%04x, want 0x%04x (TLS 1.2)", tlsConfig.MinVersion, tls.VersionTLS12)
	}

	if tlsConfig.MaxVersion != tls.VersionTLS13 {
		t.Errorf("MaxVersion = 0x%04x, want 0x%04x (TLS 1.3)", tlsConfig.MaxVersion, tls.VersionTLS13)
	}
}

func TestCreateTLSConfig_WithCustomCipherSuites(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "certtest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 写入测试证书和私钥
	certFile := filepath.Join(tmpDir, "test.crt")
	keyFile := filepath.Join(tmpDir, "test.key")

	if err := os.WriteFile(certFile, []byte(testCertPEM), 0600); err != nil {
		t.Fatalf("Failed to write cert file: %v", err)
	}

	if err := os.WriteFile(keyFile, []byte(testKeyPEM), 0600); err != nil {
		t.Fatalf("Failed to write key file: %v", err)
	}

	// 测试创建TLS配置（自定义加密套件）
	config := &cert.CertConfig{
		CertFile:    certFile,
		KeyFile:     keyFile,
		KeyPassword: "123456",
		CipherSuites: []string{
			"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
			"TLS_AES_128_GCM_SHA256",
		},
	}

	loader := cert.NewCertLoader(config)
	tlsConfig, err := loader.CreateTLSConfig()

	if err != nil {
		t.Fatalf("CreateTLSConfig failed: %v", err)
	}

	// 验证加密套件数量
	if len(tlsConfig.CipherSuites) != 2 {
		t.Errorf("CipherSuites count = %d, want 2", len(tlsConfig.CipherSuites))
	}

	// 验证第一个加密套件
	if tlsConfig.CipherSuites[0] != tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 {
		t.Error("First cipher suite not set correctly")
	}
}

func TestParseTLSVersionRange_SingleVersion(t *testing.T) {
	config := &cert.CertConfig{
		TLSVersions: []string{"TLS1.3"},
	}

	loader := cert.NewCertLoader(config)
	tlsConfig, err := loader.CreateTLSConfig()
	if err == nil {
		// 单个版本时，min和max应该相同
		if tlsConfig.MinVersion != tls.VersionTLS13 || tlsConfig.MaxVersion != tls.VersionTLS13 {
			t.Errorf("Expected both min and max to be TLS 1.3, got min=0x%04x, max=0x%04x",
				tlsConfig.MinVersion, tlsConfig.MaxVersion)
		}
	}
}

func TestParseTLSVersionRange_MultipleVersions(t *testing.T) {
	config := &cert.CertConfig{
		TLSVersions: []string{"TLS1.2", "TLS1.3", "TLS1.1"}, // 无序
	}

	loader := cert.NewCertLoader(config)
	tlsConfig, err := loader.CreateTLSConfig()
	if err == nil {
		// 应该自动计算出min=1.1, max=1.3
		if tlsConfig.MinVersion != tls.VersionTLS11 {
			t.Errorf("MinVersion = 0x%04x, want 0x%04x (TLS 1.1)", tlsConfig.MinVersion, tls.VersionTLS11)
		}
		if tlsConfig.MaxVersion != tls.VersionTLS13 {
			t.Errorf("MaxVersion = 0x%04x, want 0x%04x (TLS 1.3)", tlsConfig.MaxVersion, tls.VersionTLS13)
		}
	}
}
