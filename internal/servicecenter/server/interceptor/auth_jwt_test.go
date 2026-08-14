package interceptor

import (
	"testing"
	"time"
)

func TestSignAndValidateJWT(t *testing.T) {
	secret := "unit-test-secret"
	issuer := "service-center"
	token, err := SignJWT("u1", "default", secret, issuer, time.Hour)
	if err != nil {
		t.Fatalf("SignJWT: %v", err)
	}
	claims, err := ValidateJWT(token, secret, issuer)
	if err != nil {
		t.Fatalf("ValidateJWT: %v", err)
	}
	if claims.UserId != "u1" || claims.TenantId != "default" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestValidateJWT_RejectsWrongSecret(t *testing.T) {
	token, err := SignJWT("u1", "default", "secret-a", "service-center", time.Hour)
	if err != nil {
		t.Fatalf("SignJWT: %v", err)
	}
	if _, err := ValidateJWT(token, "secret-b", "service-center"); err == nil {
		t.Fatal("expected signature failure")
	}
}

func TestValidateJWT_RejectsExpired(t *testing.T) {
	token, err := SignJWT("u1", "default", "secret", "service-center", -time.Minute)
	if err != nil {
		t.Fatalf("SignJWT: %v", err)
	}
	if _, err := ValidateJWT(token, "secret", "service-center"); err == nil {
		t.Fatal("expected expired failure")
	}
}

func TestLooksLikeJWT(t *testing.T) {
	if !looksLikeJWT("a.b.c") {
		t.Fatal("expected jwt shape")
	}
	if looksLikeJWT("opaque-token-value") {
		t.Fatal("opaque token should not look like jwt")
	}
}
