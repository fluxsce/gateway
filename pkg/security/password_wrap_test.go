package security

import (
	"strings"
	"testing"
	"time"
)

func TestPasswordWrap_publicPEMParses(t *testing.T) {
	w, err := NewPasswordWrap()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(w.PublicPEM(), "BEGIN PUBLIC KEY") {
		t.Fatal(w.PublicPEM())
	}
}

func TestPasswordWrap_roundTripAndRejectsPlain(t *testing.T) {
	w, err := NewPasswordWrap()
	if err != nil {
		t.Fatal(err)
	}
	cipher, err := w.Wrap("S3cret!pass")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(cipher, PasswordWrapPrefix) {
		t.Fatalf("prefix: %s", cipher)
	}
	got, err := w.Unwrap(cipher)
	if err != nil || got != "S3cret!pass" {
		t.Fatalf("got %q err=%v", got, err)
	}
	if _, err := w.Unwrap("S3cret!pass"); err != ErrPasswordWrapRequired {
		t.Fatalf("plain should be rejected, err=%v", err)
	}
}

func TestPasswordWrap_pemRoundTrip(t *testing.T) {
	issuer, err := NewPasswordWrap()
	if err != nil {
		t.Fatal(err)
	}
	pemStr, err := issuer.PrivatePEM()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadPasswordWrapPEM(pemStr)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Kid() != issuer.Kid() {
		t.Fatalf("kid %s != %s", loaded.Kid(), issuer.Kid())
	}
	cipher, err := issuer.Wrap("reuse-key")
	if err != nil {
		t.Fatal(err)
	}
	got, err := loaded.Unwrap(cipher)
	if err != nil || got != "reuse-key" {
		t.Fatalf("got %q err=%v", got, err)
	}
}

func TestUnwrapPassword_passthroughWhenUnset(t *testing.T) {
	SetPasswordWrap(nil)
	t.Cleanup(func() { SetPasswordWrap(nil) })
	got, err := UnwrapPassword("plain-login")
	if err != nil || got != "plain-login" {
		t.Fatalf("got %q err=%v", got, err)
	}
	if _, err := UnwrapPassword(PasswordWrapPrefix + "dead"); err != ErrPasswordWrapInvalid {
		t.Fatalf("cipher without key should fail, err=%v", err)
	}
}

func TestUnwrapPassword_requiresWrapWhenSet(t *testing.T) {
	w, err := NewPasswordWrap()
	if err != nil {
		t.Fatal(err)
	}
	SetPasswordWrap(w)
	t.Cleanup(func() { SetPasswordWrap(nil) })
	if _, err := UnwrapPassword("plain-login"); err != ErrPasswordWrapRequired {
		t.Fatalf("plain should be rejected, err=%v", err)
	}
	cipher, err := w.Wrap("ok")
	if err != nil {
		t.Fatal(err)
	}
	got, err := UnwrapPassword(cipher)
	if err != nil || got != "ok" {
		t.Fatalf("got %q err=%v", got, err)
	}
}

func TestPasswordWrap_expired(t *testing.T) {
	w, err := NewPasswordWrap()
	if err != nil {
		t.Fatal(err)
	}
	oldTTL, oldSkew := passwordWrapTTL, passwordWrapSkew
	passwordWrapTTL = time.Millisecond
	passwordWrapSkew = 0
	t.Cleanup(func() {
		passwordWrapTTL = oldTTL
		passwordWrapSkew = oldSkew
	})
	cipher, err := w.Wrap("late")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	if _, err := w.Unwrap(cipher); err != ErrPasswordWrapInvalid {
		t.Fatalf("expired should fail, err=%v", err)
	}
}
