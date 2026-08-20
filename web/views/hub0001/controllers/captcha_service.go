package controllers

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"gateway/pkg/config"
	"gateway/pkg/logger"
	"gateway/web/views/hub0001/models"
	"strings"
	"sync"
	"time"
)

const (
	captchaTicketVersion = 1
	captchaNonceSize     = 16
	captchaMACSize       = sha256.Size
	captchaTicketSize    = 1 + 8 + captchaNonceSize + captchaMACSize
	captchaCodeLength    = 6
	captchaTTL           = 2 * time.Minute
	captchaClockSkew     = 5 * time.Second
)

var (
	// ErrCaptchaRequired 未提交验证码。
	ErrCaptchaRequired = errors.New("验证码不能为空")
	// ErrCaptchaExpired 票缺失、损坏或已过期。对外统一该文案，避免帮助探测。
	ErrCaptchaExpired = errors.New("验证码不存在或已过期")
	// ErrCaptchaInvalid 用户输入与票不匹配。
	ErrCaptchaInvalid = errors.New("验证码错误")
	// ErrCaptchaType 不支持的验证码类型。
	ErrCaptchaType = errors.New("不支持的验证码类型")
)

var captchaSecretWarnOnce sync.Once

// CaptchaService 图形验证码服务。
// 使用 HMAC 签名票，服务端不存储明文或哈希；票内不含答案，仅含过期时间、nonce 与 MAC。
// 集群节点共用 web.jwt_secret（或回退密钥）即可互相签发与校验。
type CaptchaService struct {
	secret []byte
	ttl    time.Duration
}

// NewCaptchaService 创建验证码服务，密钥与全节点 JWT 密钥对齐。
func NewCaptchaService() *CaptchaService {
	return &CaptchaService{
		secret: loadCaptchaSecret(),
		ttl:    captchaTTL,
	}
}

// GenerateCaptcha 生成图形验证码。
// 只返回签名票、PNG 图和过期时间，答案不出网、不写缓存。
func (s *CaptchaService) GenerateCaptcha(ctx context.Context, req *models.CaptchaRequest) (*models.CaptchaResponse, error) {
	if req == nil {
		req = &models.CaptchaRequest{}
	}
	captchaType := req.Type
	if captchaType == "" {
		captchaType = "random"
	}
	if captchaType != "random" {
		return nil, fmt.Errorf("%w: %s", ErrCaptchaType, captchaType)
	}

	code, err := generateRandomDigits(captchaCodeLength)
	if err != nil {
		logger.ErrorWithTrace(ctx, "生成验证码失败", "error", err)
		return nil, fmt.Errorf("生成验证码失败: %w", err)
	}

	expireAt := time.Now().Add(s.ttl)
	ticket, err := s.signTicket(code, expireAt)
	if err != nil {
		logger.ErrorWithTrace(ctx, "签发验证码票失败", "error", err)
		return nil, fmt.Errorf("签发验证码票失败: %w", err)
	}

	imageDataURI, err := renderCaptchaPNGDataURI(code)
	if err != nil {
		logger.ErrorWithTrace(ctx, "绘制验证码图片失败", "error", err)
		return nil, fmt.Errorf("绘制验证码图片失败: %w", err)
	}

	logger.DebugWithTrace(ctx, "验证码生成成功", "expireAt", expireAt.Unix())
	return &models.CaptchaResponse{
		CaptchaId: ticket,
		Image:     imageDataURI,
		ExpireAt:  expireAt.Unix(),
	}, nil
}

// VerifyCaptcha 校验用户输入。成功只说明票与输入匹配，不在服务端核销。
// 短 TTL 内同一张票可重复使用，登录失败锁定用于限制撞库。
func (s *CaptchaService) VerifyCaptcha(ctx context.Context, captchaId, code string) error {
	if captchaId == "" || strings.TrimSpace(code) == "" {
		return ErrCaptchaRequired
	}

	raw, err := base64.RawURLEncoding.DecodeString(captchaId)
	if err != nil || len(raw) != captchaTicketSize || raw[0] != captchaTicketVersion {
		return ErrCaptchaExpired
	}

	expUnix := int64(binary.BigEndian.Uint64(raw[1:9]))
	if time.Now().Add(-captchaClockSkew).Unix() > expUnix {
		return ErrCaptchaExpired
	}

	nonce := raw[9 : 9+captchaNonceSize]
	expectedMAC := raw[9+captchaNonceSize:]
	actualMAC := s.computeMAC(raw[0], expUnix, nonce, normalizeCaptchaCode(code))
	if !hmac.Equal(expectedMAC, actualMAC) {
		logger.InfoWithTrace(ctx, "验证码错误")
		return ErrCaptchaInvalid
	}
	return nil
}

// ValidateCaptcha 校验验证码并返回是否通过。
func (s *CaptchaService) ValidateCaptcha(ctx context.Context, captchaId, code string) (bool, error) {
	err := s.VerifyCaptcha(ctx, captchaId, code)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *CaptchaService) signTicket(code string, expireAt time.Time) (string, error) {
	nonce := make([]byte, captchaNonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	expUnix := expireAt.Unix()
	mac := s.computeMAC(captchaTicketVersion, expUnix, nonce, normalizeCaptchaCode(code))

	buf := make([]byte, captchaTicketSize)
	buf[0] = captchaTicketVersion
	binary.BigEndian.PutUint64(buf[1:9], uint64(expUnix))
	copy(buf[9:9+captchaNonceSize], nonce)
	copy(buf[9+captchaNonceSize:], mac)
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (s *CaptchaService) computeMAC(version byte, expUnix int64, nonce []byte, code string) []byte {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte{version})
	var expBuf [8]byte
	binary.BigEndian.PutUint64(expBuf[:], uint64(expUnix))
	mac.Write(expBuf[:])
	mac.Write(nonce)
	mac.Write([]byte(code))
	return mac.Sum(nil)
}

func normalizeCaptchaCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func generateRandomDigits(length int) (string, error) {
	const digits = "0123456789"
	out := make([]byte, length)
	for i := 0; i < length; {
		var b [1]byte
		if _, err := rand.Read(b[:]); err != nil {
			return "", err
		}
		if b[0] >= 250 {
			continue
		}
		out[i] = digits[int(b[0])%10]
		i++
	}
	return string(out), nil
}

func loadCaptchaSecret() []byte {
	secret := strings.TrimSpace(config.GetString("web.jwt_secret", ""))
	if secret == "" {
		secret = strings.TrimSpace(config.GetString("app.jwt_secret", ""))
	}
	if secret == "" {
		secret = strings.TrimSpace(config.GetString("web.encryption_key", ""))
	}
	if secret == "" {
		secret = strings.TrimSpace(config.GetString("app.encryption_key", ""))
	}
	if secret == "" {
		captchaSecretWarnOnce.Do(func() {
			logger.Warn("验证码 HMAC 未配置 jwt_secret，使用内置开发密钥，生产环境必须配置 web.jwt_secret")
		})
		secret = "gateway-captcha-dev-secret-change-me"
	}
	return []byte(secret)
}
