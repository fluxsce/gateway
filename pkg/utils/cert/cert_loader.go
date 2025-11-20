package cert

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"

	"gateway/pkg/logger"

	"github.com/youmark/pkcs8"
)

// CertConfig 证书配置
type CertConfig struct {
	CertFile     string   // 证书文件路径
	KeyFile      string   // 私钥文件路径
	KeyPassword  string   // 私钥密码（用于解密加密的私钥）
	TLSVersions  []string // TLS版本列表，支持多个版本，如: ["TLS1.2", "TLS1.3"]
	CipherSuites []string // 加密套件列表
}

// CertLoader 证书加载器
type CertLoader struct {
	config *CertConfig
	cert   *tls.Certificate
}

// NewCertLoader 创建证书加载器
func NewCertLoader(config *CertConfig) *CertLoader {
	return &CertLoader{
		config: config,
	}
}

// LoadCertificate 加载证书
func (loader *CertLoader) LoadCertificate() (*tls.Certificate, error) {
	if loader.config.CertFile == "" || loader.config.KeyFile == "" {
		return nil, fmt.Errorf("证书文件路径或私钥文件路径为空")
	}

	// 检查文件是否存在
	if _, err := os.Stat(loader.config.CertFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("证书文件不存在: %s", loader.config.CertFile)
	}
	if _, err := os.Stat(loader.config.KeyFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("私钥文件不存在: %s", loader.config.KeyFile)
	}

	// 读取证书和私钥文件
	certPEM, err := os.ReadFile(loader.config.CertFile)
	if err != nil {
		return nil, fmt.Errorf("读取证书文件失败: %w", err)
	}

	keyPEM, err := os.ReadFile(loader.config.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("读取私钥文件失败: %w", err)
	}

	// 检查私钥是否加密，如果加密则解密
	keyPEM, err = loader.decryptPrivateKey(keyPEM)
	if err != nil {
		return nil, fmt.Errorf("解密私钥失败: %w", err)
	}

	// 加载证书和私钥
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("加载证书和私钥失败: %w", err)
	}

	// 验证证书
	if err := loader.validateCertificate(&cert); err != nil {
		logger.Warn("证书验证警告", "error", err)
	}

	loader.cert = &cert
	logger.Info("证书加载成功", "certFile", loader.config.CertFile, "keyFile", loader.config.KeyFile)

	return &cert, nil
}

// decryptPrivateKey 解密加密的私钥
func (loader *CertLoader) decryptPrivateKey(keyPEM []byte) ([]byte, error) {
	// 解析PEM块
	block, rest := pem.Decode(keyPEM)
	if block == nil {
		// 如果解析失败，可能是格式问题，尝试直接返回原始数据
		// 让 tls.X509KeyPair 自己处理
		logger.Debug("PEM解析失败，尝试直接使用原始数据", "dataLen", len(keyPEM), "restLen", len(rest))
		return keyPEM, nil
	}

	// 检查是否为加密的 PKCS#8 私钥
	if block.Type == "ENCRYPTED PRIVATE KEY" {
		if loader.config.KeyPassword == "" {
			return nil, fmt.Errorf("私钥已加密（PKCS#8）但未提供密码")
		}

		logger.Debug("检测到PKCS#8加密私钥，使用密码解密")

		// 使用 github.com/youmark/pkcs8 包解密 PKCS#8 私钥
		privateKey, err := pkcs8.ParsePKCS8PrivateKey(block.Bytes, []byte(loader.config.KeyPassword))
		if err != nil {
			return nil, fmt.Errorf("解密PKCS#8私钥失败: %w", err)
		}

		// 将私钥重新编码为 PKCS#8 未加密格式
		decryptedDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			return nil, fmt.Errorf("编码私钥失败: %w", err)
		}

		// 重新编码为未加密的 PRIVATE KEY
		decryptedBlock := &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: decryptedDER,
		}
		decryptedPEM := pem.EncodeToMemory(decryptedBlock)
		logger.Info("PKCS#8私钥解密成功")
		return decryptedPEM, nil
	}

	// 检查传统加密格式（PEM 加密，如 RSA PRIVATE KEY with DEK-Info）
	if block.Headers != nil && block.Headers["DEK-Info"] != "" {
		if loader.config.KeyPassword == "" {
			return nil, fmt.Errorf("私钥已加密（传统格式）但未提供密码")
		}

		logger.Debug("检测到传统格式加密私钥，使用密码解密")

		// 解密传统格式私钥
		decryptedDER, err := decryptPEMBlock(block, []byte(loader.config.KeyPassword))
		if err != nil {
			return nil, fmt.Errorf("解密传统格式私钥失败: %w", err)
		}

		// 重新编码为未加密的私钥
		decryptedBlock := &pem.Block{
			Type:  block.Type,
			Bytes: decryptedDER,
		}
		decryptedPEM := pem.EncodeToMemory(decryptedBlock)
		logger.Info("传统格式私钥解密成功")
		return decryptedPEM, nil
	}

	// 私钥未加密，直接返回
	logger.Debug("私钥未加密，直接使用", "type", block.Type)
	return keyPEM, nil
}

// decryptPEMBlock 解密传统 PEM 加密块（替代废弃的 x509.DecryptPEMBlock）
func decryptPEMBlock(block *pem.Block, password []byte) ([]byte, error) {
	dekInfo := block.Headers["DEK-Info"]
	if dekInfo == "" {
		return nil, errors.New("未找到 DEK-Info 头")
	}

	// 解析 DEK-Info: algorithm,iv
	parts := strings.Split(dekInfo, ",")
	if len(parts) != 2 {
		return nil, errors.New("无效的 DEK-Info 格式")
	}

	algorithm := parts[0]
	ivHex := parts[1]

	// 解析 IV
	iv := make([]byte, len(ivHex)/2)
	for i := 0; i < len(iv); i++ {
		fmt.Sscanf(ivHex[i*2:i*2+2], "%02X", &iv[i])
	}

	// 根据算法解密
	switch algorithm {
	case "DES-CBC":
		return decryptDESCBC(block.Bytes, password, iv)
	case "DES-EDE3-CBC", "DES3":
		return decrypt3DESCBC(block.Bytes, password, iv)
	case "AES-128-CBC":
		return decryptAESCBC(block.Bytes, password, iv, 16)
	case "AES-192-CBC":
		return decryptAESCBC(block.Bytes, password, iv, 24)
	case "AES-256-CBC":
		return decryptAESCBC(block.Bytes, password, iv, 32)
	default:
		return nil, fmt.Errorf("不支持的加密算法: %s", algorithm)
	}
}

// deriveKey 从密码派生密钥（OpenSSL 兼容）
func deriveKey(password, salt []byte, keyLen int) []byte {
	// OpenSSL 使用 EVP_BytesToKey 派生密钥
	// 算法：MD5(password + salt) + MD5(MD5(password + salt) + password + salt) + ...
	var derived []byte
	var digest []byte

	for len(derived) < keyLen {
		h := md5.New()
		h.Write(digest)
		h.Write(password)
		h.Write(salt)
		digest = h.Sum(nil)
		derived = append(derived, digest...)
	}

	return derived[:keyLen]
}

// decryptDESCBC 使用 DES-CBC 解密
func decryptDESCBC(data, password, iv []byte) ([]byte, error) {
	key := deriveKey(password, iv[:8], 8)
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(data)%block.BlockSize() != 0 {
		return nil, errors.New("密文长度不是块大小的倍数")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(data))
	mode.CryptBlocks(decrypted, data)

	// 移除 PKCS#5 填充
	return removePKCS5Padding(decrypted)
}

// decrypt3DESCBC 使用 3DES-CBC 解密
func decrypt3DESCBC(data, password, iv []byte) ([]byte, error) {
	key := deriveKey(password, iv[:8], 24)
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}

	if len(data)%block.BlockSize() != 0 {
		return nil, errors.New("密文长度不是块大小的倍数")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(data))
	mode.CryptBlocks(decrypted, data)

	// 移除 PKCS#5 填充
	return removePKCS5Padding(decrypted)
}

// decryptAESCBC 使用 AES-CBC 解密
func decryptAESCBC(data, password, iv []byte, keyLen int) ([]byte, error) {
	key := deriveKey(password, iv, keyLen)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(data)%block.BlockSize() != 0 {
		return nil, errors.New("密文长度不是块大小的倍数")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(data))
	mode.CryptBlocks(decrypted, data)

	// 移除 PKCS#7 填充
	return removePKCS5Padding(decrypted)
}

// removePKCS5Padding 移除 PKCS#5/PKCS#7 填充
func removePKCS5Padding(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("数据为空")
	}

	padding := int(data[len(data)-1])
	if padding > len(data) || padding > aes.BlockSize {
		return nil, errors.New("无效的填充")
	}

	// 验证填充
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, errors.New("无效的填充")
		}
	}

	return data[:len(data)-padding], nil
}

// validateCertificate 验证证书
func (loader *CertLoader) validateCertificate(cert *tls.Certificate) error {
	if len(cert.Certificate) == 0 {
		return fmt.Errorf("证书链为空")
	}

	// 解析证书
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("解析证书失败: %w", err)
	}

	logger.Debug("证书信息",
		"subject", x509Cert.Subject.CommonName,
		"issuer", x509Cert.Issuer.CommonName,
		"notBefore", x509Cert.NotBefore,
		"notAfter", x509Cert.NotAfter,
		"dnsNames", x509Cert.DNSNames)

	return nil
}

// CreateTLSConfig 创建TLS配置
func (loader *CertLoader) CreateTLSConfig() (*tls.Config, error) {
	cert, err := loader.LoadCertificate()
	if err != nil {
		return nil, err
	}

	// 解析TLS版本范围
	minVersion, maxVersion := loader.parseTLSVersionRange()

	// 解析加密套件（如果未配置则为 nil，让 Go 自动选择）
	var cipherSuites []uint16
	if len(loader.config.CipherSuites) > 0 {
		cipherSuites = loader.parseCipherSuites()
	} else {
		// nil 表示使用 Go 的默认安全加密套件（推荐）
		// Go 会根据 TLS 版本自动选择最安全的加密套件
		cipherSuites = nil
		logger.Debug("使用 Go 默认加密套件（自动协商）")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*cert},
		MinVersion:   minVersion,
		MaxVersion:   maxVersion,
		CipherSuites: cipherSuites, // nil 表示使用默认值
	}

	// 设置其他安全选项
	// PreferServerCipherSuites 在 TLS 1.3 中已废弃，Go 会自动处理
	tlsConfig.CurvePreferences = []tls.CurveID{
		tls.X25519,    // 推荐：现代、快速、安全
		tls.CurveP256, // 兼容性
		tls.CurveP384, // 高安全性
	}

	cipherInfo := "自动协商"
	if cipherSuites != nil {
		cipherInfo = fmt.Sprintf("%d个自定义套件", len(cipherSuites))
	}

	logger.Info("TLS配置已创建",
		"minVersion", getTLSVersionName(minVersion),
		"maxVersion", getTLSVersionName(maxVersion),
		"cipherSuites", cipherInfo)

	return tlsConfig, nil
}

// parseTLSVersionRange 解析TLS版本范围
// 支持多个版本配置，自动计算最小和最大版本
func (loader *CertLoader) parseTLSVersionRange() (minVersion, maxVersion uint16) {
	// 默认值：TLS 1.2 到最新版本
	minVersion = tls.VersionTLS12
	maxVersion = 0 // 0表示使用最新版本

	if len(loader.config.TLSVersions) == 0 {
		return minVersion, maxVersion
	}

	// 解析所有配置的TLS版本
	versions := make([]uint16, 0, len(loader.config.TLSVersions))
	for _, versionStr := range loader.config.TLSVersions {
		if version := ParseTLSVersion(versionStr); version > 0 {
			versions = append(versions, version)
		}
	}

	if len(versions) == 0 {
		logger.Warn("未找到有效的TLS版本配置，使用默认值", "configured", loader.config.TLSVersions)
		return minVersion, maxVersion
	}

	// 找出最小和最大版本
	minVersion = versions[0]
	maxVersion = versions[0]
	for _, v := range versions {
		if v < minVersion {
			minVersion = v
		}
		if v > maxVersion {
			maxVersion = v
		}
	}

	logger.Debug("TLS版本范围已解析",
		"configured", loader.config.TLSVersions,
		"minVersion", getTLSVersionName(minVersion),
		"maxVersion", getTLSVersionName(maxVersion))

	return minVersion, maxVersion
}

// parseCipherSuites 解析加密套件
func (loader *CertLoader) parseCipherSuites() []uint16 {
	cipherSuites := ParseCipherSuites(loader.config.CipherSuites)
	if len(cipherSuites) == 0 {
		logger.Warn("未找到有效的加密套件配置", "configured", loader.config.CipherSuites)
		return nil // 返回 nil 使用 Go 默认值
	}

	logger.Debug("加密套件已解析", "count", len(cipherSuites), "configured", loader.config.CipherSuites)
	return cipherSuites
}

// ParseTLSVersion 解析TLS版本字符串
// 支持多种格式：TLS1.2, TLSv1.2, 1.2, v1.2
func ParseTLSVersion(version string) uint16 {
	// 转换为大写并去除空格
	version = strings.ToUpper(strings.TrimSpace(version))

	// 移除可能的前缀
	version = strings.TrimPrefix(version, "TLS")
	version = strings.TrimPrefix(version, "V")

	switch version {
	case "1.0", "10":
		return tls.VersionTLS10
	case "1.1", "11":
		return tls.VersionTLS11
	case "1.2", "12":
		return tls.VersionTLS12
	case "1.3", "13":
		return tls.VersionTLS13
	default:
		logger.Warn("无法识别的TLS版本", "version", version)
		return 0
	}
}

// getTLSVersionName 获取TLS版本名称
func getTLSVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	case 0:
		return "Latest"
	default:
		return fmt.Sprintf("Unknown(0x%04x)", version)
	}
}

// ParseCipherSuites 解析加密套件列表
func ParseCipherSuites(suites []string) []uint16 {
	var result []uint16
	cipherSuiteMap := getCipherSuiteMap()

	for _, suite := range suites {
		// 转换为大写并去除空格
		suite = strings.ToUpper(strings.TrimSpace(suite))
		if id, ok := cipherSuiteMap[suite]; ok {
			result = append(result, id)
		} else {
			logger.Warn("无法识别的加密套件", "suite", suite)
		}
	}

	return result
}

// getCipherSuiteMap 获取加密套件映射表（完整列表）
func getCipherSuiteMap() map[string]uint16 {
	return map[string]uint16{
		// === TLS 1.3 加密套件（推荐） ===
		"TLS_AES_128_GCM_SHA256":       tls.TLS_AES_128_GCM_SHA256,
		"TLS_AES_256_GCM_SHA384":       tls.TLS_AES_256_GCM_SHA384,
		"TLS_CHACHA20_POLY1305_SHA256": tls.TLS_CHACHA20_POLY1305_SHA256,

		// === TLS 1.2 ECDHE-RSA 加密套件（推荐：支持前向保密） ===
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256": tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384": tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305":  tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA":    tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA":    tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256": tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,

		// === TLS 1.2 ECDHE-ECDSA 加密套件（推荐：支持前向保密，ECC证书） ===
		"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384": tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305":  tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
		"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA":    tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA":    tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,

		// === TLS 1.2 RSA 加密套件（兼容性：不支持前向保密） ===
		"TLS_RSA_WITH_AES_128_GCM_SHA256": tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		"TLS_RSA_WITH_AES_256_GCM_SHA384": tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		"TLS_RSA_WITH_AES_128_CBC_SHA":    tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		"TLS_RSA_WITH_AES_256_CBC_SHA":    tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		"TLS_RSA_WITH_AES_128_CBC_SHA256": tls.TLS_RSA_WITH_AES_128_CBC_SHA256,

		// === TLS 1.2 其他加密套件 ===
		"TLS_RSA_WITH_3DES_EDE_CBC_SHA": tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA, // 不推荐：3DES已过时
	}
}
