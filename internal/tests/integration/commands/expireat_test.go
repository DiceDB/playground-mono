package commands

import (
	"server/internal/tests/integration/commands/assertions"
	"strconv"
	"testing"
	"time"
)

func TestExpireAt(t *testing.T) {
	exec, err := NewHTTPCommandExecutor()
	if err != nil {
		t.Fatal(err)
	}

	defer exec.FlushDB()

	testCases := []TestCase{
		{
			Name: "EXPIREAT on a non-existent key",
			Commands: []HTTPCommand{
				{Command: "EXPIREAT", Body: []string{"non_existent_key", "1660000000"}}, // Arbitrary timestamp
			},
			Result: []TestCaseResult{
				{Expected: "0"}, // Expecting 0 because the key does not exist
			},
		},
		{
			Name: "EXPIREAT on an existing key with future timestamp",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"temp_key", "temp_value"}},
				{Command: "EXPIREAT", Body: []string{"temp_key", strconv.FormatInt(time.Now().Add(30*time.Second).Unix(), 10)}}, // Future timestamp
				{Command: "EXISTS", Body: []string{"temp_key"}},                                                                 // Check if the key exists immediately
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "1"}, // Expecting 1 because EXPIREAT set the expiration successfully
				{Expected: "1"}, // Key should still exist as it hasn't expired yet
			},
		},
		{
			Name: "EXPIREAT on an existing key with past timestamp",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"past_key", "past_value"}},
				{Command: "EXPIREAT", Body: []string{"past_key", "1600000000"}}, // Past timestamp
				{Command: "EXISTS", Body: []string{"past_key"}},                 // Check if the key exists after expiration
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "1"}, // EXPIREAT should execute successfully
				{Expected: "0"}, // Key should not exist as it has already expired
			},
		},
		{
			Name: "EXPIREAT with invalid arguments",
			Commands: []HTTPCommand{
				{Command: "EXPIREAT", Body: []string{"key_only"}}, // Missing timestamp
			},
			Result: []TestCaseResult{
				{ErrorExpected: true, Expected: "(error) ERR wrong number of arguments for 'expireat' command"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			for i, cmd := range tc.Commands {
				response, err := exec.FireCommand(cmd)
				if err != nil {
					t.Logf("Error executing command: %s - %v", cmd.Command, err)
				} else {
					t.Logf("Response for command %s: %s", cmd.Command, response)
				}

				result := tc.Result[i]
				assertions.AssertResult(t, err, response, result.Expected, result.ErrorExpected)

			}
		})
	}
}
