package controllers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"gohub/pkg/cache"
	"gohub/pkg/logger"
	"gohub/web/views/hub0001/models"
	mathRand "math/rand"
	"strconv"
	"strings"
	"time"
)

// CaptchaService 验证码服务
type CaptchaService struct {
	cacheManager *cache.Manager
}

// NewCaptchaService 创建验证码服务
func NewCaptchaService() *CaptchaService {
	return &CaptchaService{
		cacheManager: cache.GetGlobalManager(),
	}
}

// GenerateCaptcha 生成验证码
func (s *CaptchaService) GenerateCaptcha(ctx context.Context, req *models.CaptchaRequest) (*models.CaptchaResponse, error) {
	// 默认类型为随机验证码（数字+字母）
	if req.Type == "" {
		req.Type = "random"
	}

	// 生成验证码ID
	captchaId, err := s.generateCaptchaId()
	if err != nil {
		logger.ErrorWithTrace(ctx, "生成验证码ID失败", "error", err)
		return nil, fmt.Errorf("生成验证码ID失败: %w", err)
	}

	var code string
	var captchaResp *models.CaptchaResponse
	expireTime := time.Now().Add(1 * time.Minute) // 1分钟过期

	switch req.Type {
	case "random":
		// 生成6位随机验证码（数字+字母）
		code = s.generateRandomCode(6)
		captchaResp = &models.CaptchaResponse{
			CaptchaId: captchaId,
			Code:      code,
			ExpireAt:  expireTime.Unix(),
		}
	case "math":
		// 生成数学验证码（加减乘除）
		mathExpression, answer := s.generateMathExpression()
		code = strconv.Itoa(answer) // 存储答案用于验证
		captchaResp = &models.CaptchaResponse{
			CaptchaId: captchaId,
			Code:      mathExpression, // 返回数学表达式给前端显示
			ExpireAt:  expireTime.Unix(),
		}
	case "sms":
		// 短信验证码（扩展功能，当前只是预留）
		if req.Mobile == "" {
			return nil, fmt.Errorf("手机号不能为空")
		}
		// 生成6位随机验证码（数字+字母）
		code = s.generateRandomCode(6)
		
		// TODO: 在这里添加短信发送逻辑
		// err := s.sendSMSCode(req.Mobile, code)
		// if err != nil {
		//     logger.ErrorWithTrace(ctx, "发送短信验证码失败", "error", err, "mobile", req.Mobile)
		//     return nil, fmt.Errorf("短信发送失败: %w", err)
		// }
		
		captchaResp = &models.CaptchaResponse{
			CaptchaId: captchaId,
			Code:      "", // 短信验证码不返回code
			ExpireAt:  expireTime.Unix(),
		}
		
		logger.Info("短信验证码发送", "mobile", req.Mobile, "captchaId", captchaId)
	default:
		return nil, fmt.Errorf("不支持的验证码类型: %s", req.Type)
	}

	// 将验证码存储到Redis缓存中
	redisCache := s.cacheManager.GetCache("default")
	if redisCache == nil {
		logger.ErrorWithTrace(ctx, "Redis缓存未初始化")
		return nil, fmt.Errorf("Redis缓存未初始化")
	}

	// 缓存key格式：captcha:验证码ID
	cacheKey := fmt.Sprintf("captcha:%s", captchaId)
	err = redisCache.SetString(ctx, cacheKey, code, 5*time.Minute)
	if err != nil {
		logger.ErrorWithTrace(ctx, "验证码存储到缓存失败", "error", err, "captchaId", captchaId)
		return nil, fmt.Errorf("验证码存储失败: %w", err)
	}

	logger.Info("验证码生成成功", "type", req.Type, "captchaId", captchaId)
	return captchaResp, nil
}

// VerifyCaptcha 验证验证码
func (s *CaptchaService) VerifyCaptcha(ctx context.Context, captchaId, code string) error {
	if captchaId == "" || code == "" {
		return fmt.Errorf("验证码ID和验证码不能为空")
	}

	// 从Redis缓存中获取验证码
	redisCache := s.cacheManager.GetCache("default")
	if redisCache == nil {
		logger.ErrorWithTrace(ctx, "Redis缓存未初始化")
		return fmt.Errorf("Redis缓存未初始化")
	}

	cacheKey := fmt.Sprintf("captcha:%s", captchaId)
	storedCode, err := redisCache.GetString(ctx, cacheKey)
	if err != nil {
		logger.ErrorWithTrace(ctx, "从缓存获取验证码失败", "error", err, "captchaId", captchaId)
		return fmt.Errorf("获取验证码失败: %w", err)
	}

	// 验证码不存在或已过期
	if storedCode == "" {
		return fmt.Errorf("验证码不存在或已过期")
	}

	// 验证码错误（不区分大小写）
	if !strings.EqualFold(code, storedCode) {
		logger.Info("验证码错误", "captchaId", captchaId, "input", code, "stored", storedCode)
		return fmt.Errorf("验证码错误")
	}

	// 验证成功，删除验证码（一次性使用）
	err = redisCache.Delete(ctx, cacheKey)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除已使用的验证码失败", "error", err, "captchaId", captchaId)
		// 继续执行，不影响验证结果
	}

	logger.Info("验证码验证成功", "captchaId", captchaId)
	return nil
}

// ValidateCaptcha 验证验证码的公共方法
// 这个方法可以被其他服务调用来验证验证码，返回bool值更直观
func (s *CaptchaService) ValidateCaptcha(ctx context.Context, captchaId, code string) (bool, error) {
	err := s.VerifyCaptcha(ctx, captchaId, code)
	if err != nil {
		return false, err
	}
	return true, nil
}

// generateCaptchaId 生成验证码ID
func (s *CaptchaService) generateCaptchaId() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateRandomCode 生成指定长度的随机验证码
// 包含数字0-9和英文字母大小写，但排除容易混淆的字母O
func (s *CaptchaService) generateRandomCode(length int) string {
	// 字符集：数字 + 大写字母（排除O） + 小写字母（排除o）
	const charset = "0123456789"
	randSource := mathRand.New(mathRand.NewSource(time.Now().UnixNano()))
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[randSource.Intn(len(charset))]
	}
	return string(code)
}

// generateMathExpression 生成一个数学验证码表达式
func (s *CaptchaService) generateMathExpression() (string, int) {
	randSource := mathRand.New(mathRand.NewSource(time.Now().UnixNano()))
	operators := []string{"+", "-", "*", "/"}
	operator := operators[randSource.Intn(len(operators))]

	var num1, num2, answer int
	var expression string

	switch operator {
	case "+":
		// 加法：1-50范围内的数字
		num1 = randSource.Intn(50) + 1
		num2 = randSource.Intn(50) + 1
		answer = num1 + num2
		expression = fmt.Sprintf("%d + %d", num1, num2)
	case "-":
		// 减法：确保结果为正数，num1 > num2
		num1 = randSource.Intn(50) + 10  // 10-59
		num2 = randSource.Intn(num1-1) + 1 // 1 到 num1-1，确保结果为正
		answer = num1 - num2
		expression = fmt.Sprintf("%d - %d", num1, num2)
	case "*":
		// 乘法：较小的数字避免结果过大
		num1 = randSource.Intn(12) + 1  // 1-12
		num2 = randSource.Intn(12) + 1  // 1-12
		answer = num1 * num2
		expression = fmt.Sprintf("%d × %d", num1, num2)
	case "/":
		// 除法：确保能够整除
		// 先生成答案，再生成被除数
		answer = randSource.Intn(20) + 1    // 答案1-20
		num2 = randSource.Intn(10) + 2      // 除数2-11
		num1 = answer * num2                // 被除数 = 答案 × 除数
		expression = fmt.Sprintf("%d ÷ %d", num1, num2)
	}

	return expression, answer
}

// sendSMSCode 发送短信验证码（扩展功能，当前为预留接口）
// func (s *CaptchaService) sendSMSCode(mobile, code string) error {
//     // TODO: 实现短信发送逻辑
//     // 1. 调用短信服务提供商API
//     // 2. 发送验证码到指定手机号
//     // 3. 处理发送结果
//     return nil
// } 