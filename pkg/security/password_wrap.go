package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	// PasswordWrapPrefix 控制台口令密文前缀。抓包可见此前缀，看不到口令原文。
	PasswordWrapPrefix = "enc.rsa1."
	passwordWrapBits   = 2048
)

var (
	passwordWrapTTL  = 2 * time.Minute
	passwordWrapSkew = 5 * time.Second
)

var (
	// ErrPasswordWrapRequired 已开启密文传输但未按 RSA-OAEP 包装。
	ErrPasswordWrapRequired = errors.New("密码必须加密后提交")
	// ErrPasswordWrapInvalid 密文损坏、过期或无法解密。
	ErrPasswordWrapInvalid = errors.New("密码密文无效或已过期")
)

type passwordWrapPayload struct {
	V int    `json:"v"`
	T int64  `json:"t"`
	P string `json:"p"`
}

// PasswordWrap RSA-OAEP 敏感字段包装。登录、改密、建用户以及其它模块的口令/密钥都用这一套。
// 公钥给控制台，私钥只留服务端。运行时钥匙由环境设置灌入，不要各自 New。
type PasswordWrap struct {
	key *rsa.PrivateKey
	kid string
	pem string
}

var (
	activeWrapMu sync.RWMutex
	activeWrap   *PasswordWrap
)

// SetPasswordWrap 安装或清空进程内包装钥。密文传输关闭时传入 nil。
func SetPasswordWrap(w *PasswordWrap) {
	activeWrapMu.Lock()
	activeWrap = w
	activeWrapMu.Unlock()
}

// DefaultPasswordWrap 返回当前已安装的包装钥。未开启密文传输时为 nil。
func DefaultPasswordWrap() *PasswordWrap {
	activeWrapMu.RLock()
	defer activeWrapMu.RUnlock()
	return activeWrap
}

// NewPasswordWrap 生成一把 RSA 钥匙。生产路径由环境设置首次开启时调用并入库。
func NewPasswordWrap() (*PasswordWrap, error) {
	key, err := rsa.GenerateKey(rand.Reader, passwordWrapBits)
	if err != nil {
		return nil, fmt.Errorf("generate password wrap key: %w", err)
	}
	return passwordWrapFromKey(key)
}

// LoadPasswordWrapPEM 从 PKCS#8 或 PKCS#1 私钥 PEM 恢复包装钥。
func LoadPasswordWrapPEM(pemStr string) (*PasswordWrap, error) {
	pemStr = strings.TrimSpace(pemStr)
	if pemStr == "" {
		return nil, ErrPasswordWrapInvalid
	}
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, ErrPasswordWrapInvalid
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, ErrPasswordWrapInvalid
		}
		return passwordWrapFromKey(rsaKey)
	}
	rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, ErrPasswordWrapInvalid
	}
	return passwordWrapFromKey(rsaKey)
}

func passwordWrapFromKey(key *rsa.PrivateKey) (*PasswordWrap, error) {
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(der)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
	return &PasswordWrap{
		key: key,
		kid: fmt.Sprintf("%x", sum[:8]),
		pem: string(pemBytes),
	}, nil
}

// Kid 公钥指纹，控制台可据此缓存。
func (w *PasswordWrap) Kid() string {
	if w == nil {
		return ""
	}
	return w.kid
}

// PublicPEM 返回 PKIX PEM 公钥。
func (w *PasswordWrap) PublicPEM() string {
	if w == nil {
		return ""
	}
	return w.pem
}

// PrivatePEM 返回 PKCS#8 私钥 PEM，仅供环境设置入库。
func (w *PasswordWrap) PrivatePEM() (string, error) {
	if w == nil || w.key == nil {
		return "", ErrPasswordWrapInvalid
	}
	der, err := x509.MarshalPKCS8PrivateKey(w.key)
	if err != nil {
		return "", err
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})), nil
}

// Wrap 把口令打成带前缀的 RSA-OAEP 密文，供单测与联调。
func (w *PasswordWrap) Wrap(plain string) (string, error) {
	if w == nil || w.key == nil {
		return "", ErrPasswordWrapInvalid
	}
	plain = strings.TrimSpace(plain)
	if plain == "" {
		return "", ErrPasswordWrapRequired
	}
	raw, err := json.Marshal(passwordWrapPayload{V: 1, T: time.Now().Unix(), P: plain})
	if err != nil {
		return "", err
	}
	cipher, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &w.key.PublicKey, raw, nil)
	if err != nil {
		return "", err
	}
	return PasswordWrapPrefix + base64.RawURLEncoding.EncodeToString(cipher), nil
}

// Unwrap 解开控制台提交的口令密文。无前缀视为明文，直接拒绝。
func (w *PasswordWrap) Unwrap(value string) (string, error) {
	if w == nil || w.key == nil {
		return "", ErrPasswordWrapInvalid
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrPasswordWrapRequired
	}
	if !strings.HasPrefix(value, PasswordWrapPrefix) {
		return "", ErrPasswordWrapRequired
	}
	blob := strings.TrimPrefix(value, PasswordWrapPrefix)
	raw, err := decodeWrapBlob(blob)
	if err != nil {
		return "", ErrPasswordWrapInvalid
	}
	plain, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, w.key, raw, nil)
	if err != nil {
		return "", ErrPasswordWrapInvalid
	}
	var payload passwordWrapPayload
	if json.Unmarshal(plain, &payload) != nil || payload.V != 1 || payload.P == "" {
		return "", ErrPasswordWrapInvalid
	}
	issued := time.Unix(payload.T, 0)
	if time.Now().After(issued.Add(passwordWrapTTL + passwordWrapSkew)) {
		return "", ErrPasswordWrapInvalid
	}
	if issued.After(time.Now().Add(passwordWrapSkew)) {
		return "", ErrPasswordWrapInvalid
	}
	return payload.P, nil
}

// UnwrapPassword 解开敏感字段。未安装钥匙时原样返回明文；已安装则必须带前缀。
func UnwrapPassword(value string) (string, error) {
	w := DefaultPasswordWrap()
	if w == nil {
		if strings.HasPrefix(strings.TrimSpace(value), PasswordWrapPrefix) {
			return "", ErrPasswordWrapInvalid
		}
		return value, nil
	}
	return w.Unwrap(value)
}

// ClonePasswordWrap 从已有钥匙复制一份，用于验证「同一把钥、另一处解包」。
func ClonePasswordWrap(src *PasswordWrap) (*PasswordWrap, error) {
	if src == nil || src.key == nil {
		return nil, ErrPasswordWrapInvalid
	}
	return passwordWrapFromKey(src.key)
}

func decodeWrapBlob(blob string) ([]byte, error) {
	if raw, err := base64.RawURLEncoding.DecodeString(blob); err == nil {
		return raw, nil
	}
	return base64.StdEncoding.DecodeString(blob)
}
