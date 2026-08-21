package controllers

import "testing"

func TestApplyTenantAdminFlagChange(t *testing.T) {
	t.Parallel()

	t.Run("tenant admin can grant on create", func(t *testing.T) {
		got, err := applyTenantAdminFlagChange(true, "Y", "", true)
		if err != nil || got != "Y" {
			t.Fatalf("got %q err=%v", got, err)
		}
	})

	t.Run("non admin create with Y is rejected", func(t *testing.T) {
		_, err := applyTenantAdminFlagChange(false, "Y", "", true)
		if err != errCannotGrantTenantAdmin {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("non admin create defaults to N", func(t *testing.T) {
		got, err := applyTenantAdminFlagChange(false, "", "", true)
		if err != nil || got != "N" {
			t.Fatalf("got %q err=%v", got, err)
		}
	})

	t.Run("non admin edit keeps existing Y", func(t *testing.T) {
		got, err := applyTenantAdminFlagChange(false, "N", "Y", false)
		if err != nil || got != "Y" {
			t.Fatalf("got %q err=%v", got, err)
		}
	})

	t.Run("non admin cannot escalate", func(t *testing.T) {
		_, err := applyTenantAdminFlagChange(false, "Y", "N", false)
		if err != errCannotGrantTenantAdmin {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("tenant admin can demote", func(t *testing.T) {
		got, err := applyTenantAdminFlagChange(true, "N", "Y", false)
		if err != nil || got != "N" {
			t.Fatalf("got %q err=%v", got, err)
		}
	})
}
