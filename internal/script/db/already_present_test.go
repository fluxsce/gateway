package db

import (
	"errors"
	"testing"
)

func TestIsAlreadyPresentError(t *testing.T) {
	present := []string{
		"Error 1050: Table 'hub.HUB_GW_ACCESS_LOG' already exists",
		"Error 1062: Duplicate entry '1' for key 'PRIMARY'",
		"Error 1060: Duplicate column name 'routeName'",
		"Error 1061: Duplicate key name 'idx_route'",
		"ORA-00001: unique constraint (HUB.PK) violated",
		"ORA-00955: name is already used by an existing object",
		"There is already an object named 'HUB_GW_ACCESS_LOG' in the database.",
		"Violation of PRIMARY KEY constraint 'PK'. Cannot insert duplicate key",
		"UNIQUE constraint failed: HUB_USER.userId",
	}
	for _, msg := range present {
		if !isAlreadyPresentError(errors.New(msg)) {
			t.Fatalf("应视为已存在: %s", msg)
		}
	}

	absent := []error{
		nil,
		errors.New("ORA-00942: table or view does not exist"),
		errors.New("Error 1146: Table 'hub.missing' doesn't exist"),
		errors.New("dial tcp: connection refused"),
	}
	for _, err := range absent {
		if isAlreadyPresentError(err) {
			t.Fatalf("不应视为已存在: %v", err)
		}
	}
}
