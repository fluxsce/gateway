package controllers

import (
	"errors"
	"testing"

	"gateway/web/views/hub0001/models"
)

func TestValidateOwnProfile(t *testing.T) {
	t.Parallel()

	if err := validateOwnProfile(models.ProfileUpdateRequest{RealName: "张三", Gender: 1}); err != nil {
		t.Fatalf("valid profile rejected: %v", err)
	}
	if err := validateOwnProfile(models.ProfileUpdateRequest{RealName: " ", Gender: 0}); !errors.Is(err, errProfileRealNameRequired) {
		t.Fatalf("empty name err=%v", err)
	}
	if err := validateOwnProfile(models.ProfileUpdateRequest{RealName: "A", Gender: 0}); !errors.Is(err, errProfileRealNameLength) {
		t.Fatalf("short name err=%v", err)
	}
	if err := validateOwnProfile(models.ProfileUpdateRequest{RealName: "张三", Email: "bad", Gender: 0}); !errors.Is(err, errProfileEmailInvalid) {
		t.Fatalf("bad email err=%v", err)
	}
	if err := validateOwnProfile(models.ProfileUpdateRequest{RealName: "张三", Mobile: "123", Gender: 0}); !errors.Is(err, errProfileMobileInvalid) {
		t.Fatalf("bad mobile err=%v", err)
	}
	if err := validateOwnProfile(models.ProfileUpdateRequest{RealName: "张三", Gender: 9}); !errors.Is(err, errProfileGenderInvalid) {
		t.Fatalf("bad gender err=%v", err)
	}
}

func TestValidateOwnProfile_ignoresRequestUserId(t *testing.T) {
	t.Parallel()
	req := models.ProfileUpdateRequest{
		UserId:   "ag-test001",
		RealName: "测试用户",
		Gender:   0,
	}
	if err := validateOwnProfile(req); err != nil {
		t.Fatalf("profile with foreign userId should still validate: %v", err)
	}
}
