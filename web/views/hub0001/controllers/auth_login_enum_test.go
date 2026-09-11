package controllers

import (
	"errors"
	"testing"

	"gateway/web/utils/constants"
)

func TestLoginPublicFailure_hidesMissingUser(t *testing.T) {
	t.Parallel()

	missingMsg, missingID := loginPublicFailure(ErrUserNotFound)
	wrongMsg, wrongID := loginPublicFailure(ErrInvalidCredentials)
	if missingMsg != wrongMsg || missingID != wrongID {
		t.Fatalf("missing=%q/%s wrong=%q/%s", missingMsg, missingID, wrongMsg, wrongID)
	}
	if missingMsg != ErrInvalidCredentials.Error() {
		t.Fatalf("msg=%q", missingMsg)
	}
	if missingID != constants.ED00103 {
		t.Fatalf("id=%s", missingID)
	}
}

func TestLoginPublicFailure_otherErrorsStayDistinct(t *testing.T) {
	t.Parallel()

	if _, id := loginPublicFailure(ErrUserDisabled); id != constants.ED00104 {
		t.Fatalf("disabled id=%s", id)
	}
	if _, id := loginPublicFailure(ErrUserExpired); id != constants.ED00105 {
		t.Fatalf("expired id=%s", id)
	}
	if _, id := loginPublicFailure(errors.New("db down")); id != constants.ED00101 {
		t.Fatalf("unknown id=%s", id)
	}
}
