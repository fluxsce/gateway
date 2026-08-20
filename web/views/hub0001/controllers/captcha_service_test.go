package controllers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image/png"
	"strings"
	"testing"
	"time"

	"gateway/web/views/hub0001/models"
)

func newTestCaptchaService(ttl time.Duration) *CaptchaService {
	return &CaptchaService{
		secret: []byte("test-captcha-secret"),
		ttl:    ttl,
	}
}

func TestGenerateCaptcha_doesNotLeakCode(t *testing.T) {
	s := newTestCaptchaService(2 * time.Minute)
	resp, err := s.GenerateCaptcha(context.Background(), &models.CaptchaRequest{})
	if err != nil {
		t.Fatalf("GenerateCaptcha: %v", err)
	}
	if resp.CaptchaId == "" {
		t.Fatal("empty captchaId")
	}
	if !strings.HasPrefix(resp.Image, "data:image/png;base64,") {
		t.Fatalf("image is not png data uri: %s", resp.Image[:min(32, len(resp.Image))])
	}

	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToLower(string(raw)), `"code"`) {
		t.Fatalf("response json leaked code field: %s", raw)
	}

	payload, err := base64.RawURLEncoding.DecodeString(resp.CaptchaId)
	if err != nil {
		t.Fatalf("captchaId is not a ticket: %v", err)
	}
	if len(payload) != captchaTicketSize {
		t.Fatalf("ticket size %d, want %d", len(payload), captchaTicketSize)
	}
}

func TestGenerateCaptcha_imageIsPNG(t *testing.T) {
	s := newTestCaptchaService(time.Minute)
	resp, err := s.GenerateCaptcha(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	b64 := strings.TrimPrefix(resp.Image, "data:image/png;base64,")
	bin, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(bin))
	if err != nil {
		t.Fatalf("png decode: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != captchaImageWidth || b.Dy() != captchaImageHeight {
		t.Fatalf("image size %dx%d, want %dx%d", b.Dx(), b.Dy(), captchaImageWidth, captchaImageHeight)
	}
}

func TestVerifyCaptcha_roundTripAndRejectsWrongCode(t *testing.T) {
	s := newTestCaptchaService(time.Minute)
	code := "123456"
	expireAt := time.Now().Add(time.Minute)
	ticket, err := s.signTicket(code, expireAt)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.VerifyCaptcha(context.Background(), ticket, code); err != nil {
		t.Fatalf("valid code rejected: %v", err)
	}
	if err := s.VerifyCaptcha(context.Background(), ticket, "000000"); !errors.Is(err, ErrCaptchaInvalid) {
		t.Fatalf("wrong code err=%v, want ErrCaptchaInvalid", err)
	}
	if err := s.VerifyCaptcha(context.Background(), ticket, ""); !errors.Is(err, ErrCaptchaRequired) {
		t.Fatalf("empty code err=%v, want ErrCaptchaRequired", err)
	}
}

func TestVerifyCaptcha_expiredAndTampered(t *testing.T) {
	s := newTestCaptchaService(-time.Minute)
	resp, err := s.GenerateCaptcha(context.Background(), &models.CaptchaRequest{Type: "random"})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.VerifyCaptcha(context.Background(), resp.CaptchaId, "000000"); !errors.Is(err, ErrCaptchaExpired) {
		t.Fatalf("expired ticket err=%v, want ErrCaptchaExpired", err)
	}

	s2 := newTestCaptchaService(time.Minute)
	ticket, err := s2.signTicket("654321", time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	tampered := ticket[:len(ticket)-2] + "aa"
	if err := s2.VerifyCaptcha(context.Background(), tampered, "654321"); !errors.Is(err, ErrCaptchaExpired) && !errors.Is(err, ErrCaptchaInvalid) {
		t.Fatalf("tampered ticket err=%v, want expired or invalid", err)
	}
}

func TestVerifyCaptcha_clusterSameSecret(t *testing.T) {
	issuer := newTestCaptchaService(time.Minute)
	verifier := newTestCaptchaService(time.Minute)
	ticket, err := issuer.signTicket("112233", time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if err := verifier.VerifyCaptcha(context.Background(), ticket, "112233"); err != nil {
		t.Fatalf("peer node rejected ticket: %v", err)
	}

	other := &CaptchaService{secret: []byte("other-node-secret"), ttl: time.Minute}
	if err := other.VerifyCaptcha(context.Background(), ticket, "112233"); !errors.Is(err, ErrCaptchaInvalid) {
		t.Fatalf("different secret err=%v, want ErrCaptchaInvalid", err)
	}
}

func TestGenerateCaptcha_unsupportedType(t *testing.T) {
	s := newTestCaptchaService(time.Minute)
	_, err := s.GenerateCaptcha(context.Background(), &models.CaptchaRequest{Type: "sms"})
	if !errors.Is(err, ErrCaptchaType) {
		t.Fatalf("err=%v, want ErrCaptchaType", err)
	}
}
