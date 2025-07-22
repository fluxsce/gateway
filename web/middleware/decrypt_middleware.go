package middleware

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gateway/pkg/config"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/response"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// DecryptRequest 请求数据解密中间件
// 对前端发送的加密数据进行解密处理
func DecryptRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否启用加密功能
		if !isEncryptionEnabled() {
			logger.Debug("加密功能已禁用，跳过解密处理")
			c.Next()
			return
		}

		// 只处理POST、PUT、PATCH请求
		if c.Request.Method != "POST" && c.Request.Method != "PUT" && c.Request.Method != "PATCH" {
			c.Next()
			return
		}

		// 检查Content-Type是否为支持的类型
		contentType := c.GetHeader("Content-Type")
		if !isSupportedContentType(contentType) {
			c.Next()
			return
		}

		// 检查是否为加密请求
		isEncrypted := c.GetHeader("X-Encrypted")
		if isEncrypted != "true" {
			c.Next()
			return
		}

		// 读取请求体
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Error("读取请求体失败", "error", err)
			response.ErrorJSON(c, "请求数据读取失败", constants.ED00001, http.StatusBadRequest)
			c.Abort()
			return
		}

		// 如果请求体为空，直接继续
		if len(body) == 0 {
			c.Next()
			return
		}

		// 根据Content-Type解析和解密数据
		contentType = c.GetHeader("Content-Type")
		decryptedData, newContentType, err := decryptRequestData(body, contentType)
		if err != nil {
			logger.Error("请求数据解密失败", "error", err, "contentType", contentType)
			response.ErrorJSON(c, "数据解密失败", constants.ED00001, http.StatusBadRequest)
			c.Abort()
			return
		}

		// 将解密后的数据重新设置到请求体
		c.Request.Body = io.NopCloser(strings.NewReader(decryptedData))
		c.Request.ContentLength = int64(len(decryptedData))

		// 更新Content-Type（如果需要）
		if newContentType != "" {
			c.Request.Header.Set("Content-Type", newContentType)
		}

		// 设置解密标记
		c.Set("decrypted", true)

		logger.Debug("请求数据解密成功", "originalSize", len(body), "decryptedSize", len(decryptedData))
		c.Next()
	}
}

// EncryptResponse 响应数据加密中间件
// 对返回给前端的数据进行加密处理
func EncryptResponse() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否启用加密功能
		if !isEncryptionEnabled() {
			logger.Debug("加密功能已禁用，跳过响应加密处理")
			c.Next()
			return
		}

		// 检查是否需要加密响应
		needEncrypt := c.GetHeader("X-Encrypt-Response")
		if needEncrypt != "true" {
			c.Next()
			return
		}

		// 创建自定义ResponseWriter
		writer := &encryptResponseWriter{
			ResponseWriter: c.Writer,
			context:        c,
		}
		c.Writer = writer

		c.Next()
	}
}

// encryptResponseWriter 自定义ResponseWriter用于加密响应
type encryptResponseWriter struct {
	gin.ResponseWriter
	context *gin.Context
	body    *bytes.Buffer
}

// Write 重写Write方法
func (w *encryptResponseWriter) Write(data []byte) (int, error) {
	if w.body == nil {
		w.body = &bytes.Buffer{}
	}
	return w.body.Write(data)
}

// WriteString 重写WriteString方法
func (w *encryptResponseWriter) WriteString(s string) (int, error) {
	if w.body == nil {
		w.body = &bytes.Buffer{}
	}
	return w.body.WriteString(s)
}

// WriteHeader 响应头写入时触发加密
func (w *encryptResponseWriter) WriteHeader(code int) {
	if w.body != nil && w.body.Len() > 0 {
		// 加密响应数据
		originalData := w.body.String()

		// 检查是否为JSON格式
		if isValidJSON(originalData) {
			encryptedData, iv, err := encryptAES(originalData)
			if err != nil {
				logger.Error("响应数据加密失败", "error", err)
				w.ResponseWriter.WriteHeader(code)
				w.ResponseWriter.Write(w.body.Bytes())
				return
			}

			// 构造加密响应
			response := map[string]interface{}{
				"encrypted": true,
				"data":      encryptedData,
				"iv":        iv,
			}

			jsonData, err := json.Marshal(response)
			if err != nil {
				logger.Error("加密响应序列化失败", "error", err)
				w.ResponseWriter.WriteHeader(code)
				w.ResponseWriter.Write(w.body.Bytes())
				return
			}

			// 设置响应头
			w.ResponseWriter.Header().Set("Content-Type", "application/json")
			w.ResponseWriter.Header().Set("X-Encrypted", "true")
			w.ResponseWriter.WriteHeader(code)
			w.ResponseWriter.Write(jsonData)

			logger.Debug("响应数据加密成功", "originalSize", len(originalData), "encryptedSize", len(jsonData))
			return
		}
	}

	// 不加密，直接返回原数据
	w.ResponseWriter.WriteHeader(code)
	if w.body != nil {
		w.ResponseWriter.Write(w.body.Bytes())
	}
}

// getEncryptionKey 获取加密密钥
func getEncryptionKey() []byte {
	// 从配置文件获取密钥
	key := config.GetString("app.encryption_key", "")
	if key == "" {
		// 使用默认密钥
		key = "gateway-default-encryption-key-32chars"

		// 根据环境给出不同的提示
		if IsProductionEnvironment() {
			logger.Error("生产环境必须配置app.encryption_key，当前使用默认密钥存在安全风险")
		} else if IsDevEnvironment() {
			logger.Info("开发环境使用默认加密密钥，生产环境请配置app.encryption_key")
		} else {
			logger.Warn("使用默认加密密钥，生产环境请配置app.encryption_key")
		}
	}

	// 使用SHA256确保密钥长度为32字节
	hash := sha256.Sum256([]byte(key))
	return hash[:]
}

// decryptAES AES解密
func decryptAES(encryptedText, ivText string) (string, error) {
	// Base64解码
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %w", err)
	}

	iv, err := base64.StdEncoding.DecodeString(ivText)
	if err != nil {
		return "", fmt.Errorf("IV解码失败: %w", err)
	}

	// 创建AES cipher
	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建AES cipher失败: %w", err)
	}

	// 检查IV长度
	if len(iv) != aes.BlockSize {
		return "", fmt.Errorf("IV长度错误")
	}

	// CBC模式解密
	mode := cipher.NewCBCDecrypter(block, iv)

	// 解密
	mode.CryptBlocks(ciphertext, ciphertext)

	// 去除PKCS7填充
	plaintext, err := removePKCS7Padding(ciphertext)
	if err != nil {
		return "", fmt.Errorf("去除填充失败: %w", err)
	}

	return string(plaintext), nil
}

// encryptAES AES加密
func encryptAES(plaintext string) (string, string, error) {
	// 生成随机IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", "", fmt.Errorf("生成IV失败: %w", err)
	}

	// 创建AES cipher
	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", fmt.Errorf("创建AES cipher失败: %w", err)
	}

	// PKCS7填充
	paddedText := addPKCS7Padding([]byte(plaintext), aes.BlockSize)

	// CBC模式加密
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(paddedText))
	mode.CryptBlocks(ciphertext, paddedText)

	// Base64编码
	encryptedText := base64.StdEncoding.EncodeToString(ciphertext)
	ivText := base64.StdEncoding.EncodeToString(iv)

	return encryptedText, ivText, nil
}

// addPKCS7Padding 添加PKCS7填充
func addPKCS7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// removePKCS7Padding 移除PKCS7填充
func removePKCS7Padding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("数据为空")
	}

	unpadding := int(data[length-1])
	if unpadding > length {
		return nil, fmt.Errorf("填充数据错误")
	}

	return data[:(length - unpadding)], nil
}

// isValidJSON 检查字符串是否为有效的JSON
func isValidJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// isEncryptionEnabled 检查是否启用加密功能
// 开发环境下可以禁用加密，生产环境建议启用
func isEncryptionEnabled() bool {
	// 1. 从配置文件获取加密开关
	encryptionEnabled := config.GetBool("app.encryption_enabled", true)
	if !encryptionEnabled {
		return false
	}

	// 2. 检查环境变量
	env := config.GetString("app.env", "development")

	// 开发环境下检查是否强制禁用加密
	if env == "development" {
		devEncryptionDisabled := config.GetBool("app.dev_disable_encryption", false)
		if devEncryptionDisabled {
			return false
		}
	}

	return true
}

// IsDevEnvironment 检查是否为开发环境
func IsDevEnvironment() bool {
	env := config.GetString("app.env", "development")
	return env == "development" || env == "dev"
}

// IsProductionEnvironment 检查是否为生产环境
func IsProductionEnvironment() bool {
	env := config.GetString("app.env", "development")
	return env == "production" || env == "prod"
}

// isSupportedContentType 检查是否支持的Content-Type
func isSupportedContentType(contentType string) bool {
	return strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "application/x-www-form-urlencoded") ||
		strings.Contains(contentType, "multipart/form-data")
}

// decryptRequestData 根据Content-Type解密请求数据
func decryptRequestData(body []byte, contentType string) (string, string, error) {
	if strings.Contains(contentType, "application/json") {
		return decryptJSONData(body)
	} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		return decryptFormURLEncodedData(body)
	} else if strings.Contains(contentType, "multipart/form-data") {
		return decryptMultipartData(body, contentType)
	}

	return "", "", fmt.Errorf("不支持的Content-Type: %s", contentType)
}

// decryptJSONData 解密JSON数据
func decryptJSONData(body []byte) (string, string, error) {
	// 解析加密数据结构
	var encryptedData struct {
		Data string `json:"data"` // 加密的数据
		IV   string `json:"iv"`   // 初始化向量
	}

	err := json.Unmarshal(body, &encryptedData)
	if err != nil {
		return "", "", fmt.Errorf("解析JSON加密数据失败: %w", err)
	}

	// 解密数据
	decryptedData, err := decryptAES(encryptedData.Data, encryptedData.IV)
	if err != nil {
		return "", "", fmt.Errorf("JSON数据解密失败: %w", err)
	}

	// 验证解密后的数据是否为有效JSON
	if !isValidJSON(decryptedData) {
		return "", "", fmt.Errorf("解密后的数据不是有效的JSON格式")
	}

	return decryptedData, "", nil
}

// decryptFormURLEncodedData 解密form-urlencoded数据
func decryptFormURLEncodedData(body []byte) (string, string, error) {
	// 解析form数据
	bodyStr := string(body)

	// 查找加密数据字段
	values, err := parseFormData(bodyStr)
	if err != nil {
		return "", "", fmt.Errorf("解析form数据失败: %w", err)
	}

	// 获取加密数据和IV
	encryptedDataStr, hasData := values["data"]
	ivStr, hasIV := values["iv"]

	if !hasData || !hasIV {
		return "", "", fmt.Errorf("form数据中缺少data或iv字段")
	}

	// 解密数据
	decryptedData, err := decryptAES(encryptedDataStr, ivStr)
	if err != nil {
		return "", "", fmt.Errorf("form数据解密失败: %w", err)
	}

	return decryptedData, "application/x-www-form-urlencoded", nil
}

// decryptMultipartData 解密multipart/form-data数据
func decryptMultipartData(body []byte, contentType string) (string, string, error) {
	// 对于multipart/form-data，我们需要解析边界
	boundary := extractBoundary(contentType)
	if boundary == "" {
		return "", "", fmt.Errorf("无法从Content-Type中提取boundary")
	}

	// 解析multipart数据
	values, err := parseMultipartData(body, boundary)
	if err != nil {
		return "", "", fmt.Errorf("解析multipart数据失败: %w", err)
	}

	// 获取加密数据和IV
	encryptedDataStr, hasData := values["data"]
	ivStr, hasIV := values["iv"]

	if !hasData || !hasIV {
		return "", "", fmt.Errorf("multipart数据中缺少data或iv字段")
	}

	// 解密数据
	decryptedData, err := decryptAES(encryptedDataStr, ivStr)
	if err != nil {
		return "", "", fmt.Errorf("multipart数据解密失败: %w", err)
	}

	return decryptedData, contentType, nil
}

// parseFormData 解析form-urlencoded数据
func parseFormData(data string) (map[string]string, error) {
	values := make(map[string]string)

	pairs := strings.Split(data, "&")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}

		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := kv[0]
		value := kv[1]

		// URL解码
		decodedKey, err := decodeURLComponent(key)
		if err != nil {
			continue
		}

		decodedValue, err := decodeURLComponent(value)
		if err != nil {
			continue
		}

		values[decodedKey] = decodedValue
	}

	return values, nil
}

// extractBoundary 从Content-Type中提取boundary
func extractBoundary(contentType string) string {
	parts := strings.Split(contentType, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "boundary=") {
			return strings.TrimPrefix(part, "boundary=")
		}
	}
	return ""
}

// parseMultipartData 解析multipart数据
func parseMultipartData(body []byte, boundary string) (map[string]string, error) {
	values := make(map[string]string)

	// 分割数据
	delimiter := "--" + boundary
	parts := strings.Split(string(body), delimiter)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == "--" {
			continue
		}

		// 分离头部和内容
		sections := strings.SplitN(part, "\r\n\r\n", 2)
		if len(sections) != 2 {
			continue
		}

		headers := sections[0]
		content := strings.TrimRight(sections[1], "\r\n")

		// 解析Content-Disposition头
		name := extractFieldName(headers)
		if name != "" {
			values[name] = content
		}
	}

	return values, nil
}

// extractFieldName 从headers中提取字段名
func extractFieldName(headers string) string {
	lines := strings.Split(headers, "\r\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "content-disposition:") {
			// 解析 name="fieldname"
			parts := strings.Split(line, ";")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "name=") {
					name := strings.TrimPrefix(part, "name=")
					name = strings.Trim(name, `"`)
					return name
				}
			}
		}
	}
	return ""
}

// decodeURLComponent URL解码
func decodeURLComponent(s string) (string, error) {
	// 简单的URL解码实现
	s = strings.ReplaceAll(s, "+", " ")

	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) {
			hex := s[i+1 : i+3]
			if b, err := parseHex(hex); err == nil {
				result = append(result, b)
				i += 2
			} else {
				result = append(result, s[i])
			}
		} else {
			result = append(result, s[i])
		}
	}

	return string(result), nil
}

// parseHex 解析十六进制字符串
func parseHex(s string) (byte, error) {
	if len(s) != 2 {
		return 0, fmt.Errorf("invalid hex length")
	}

	var result byte
	for i, c := range s {
		var digit byte
		switch {
		case '0' <= c && c <= '9':
			digit = byte(c - '0')
		case 'a' <= c && c <= 'f':
			digit = byte(c - 'a' + 10)
		case 'A' <= c && c <= 'F':
			digit = byte(c - 'A' + 10)
		default:
			return 0, fmt.Errorf("invalid hex character")
		}

		if i == 0 {
			result = digit << 4
		} else {
			result |= digit
		}
	}

	return result, nil
}
