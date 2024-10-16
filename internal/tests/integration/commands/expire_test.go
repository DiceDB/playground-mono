package commands

import (
	"server/internal/tests/integration/commands/assertions"
	"testing"
	"time"
)

func TestExpire(t *testing.T) {
	exec, err := NewHTTPCommandExecutor()
	if err != nil {
		t.Fatal(err)
	}

	defer exec.FlushDB()

	testCases := []TestCase{
		{
			Name: "EXPIRE on an existing key",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"key_to_expire_1", "SomeValue"}},
				{Command: "EXPIRE", Body: []string{"key_to_expire_1", "1"}}, // 1 second expiration
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "1"},
			},
		},
		{
			Name: "Check expiration after delay",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"key_to_expire_2", "SomeValue"}},
				{Command: "EXPIRE", Body: []string{"key_to_expire_2", "1"}},
				{Command: "GET", Body: []string{"key_to_expire_2"}}, // Check if key is still accessible
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "1"},
				{Expected: "SomeValue"},
			},
		},
		{
			Name: "Check key after waiting for expiration",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"key_to_expire_3", "SomeValue"}},
				{Command: "EXPIRE", Body: []string{"key_to_expire_3", "3"}},
				{Command: "GET", Body: []string{"key_to_expire_3"}}, // Check if key is still accessible after waiting
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "1"},
				{Expected: "(nil)"}, // Expecting (nil) after waiting
			},
		},
		{
			Name: "EXPIRE on a non-existent key",
			Commands: []HTTPCommand{
				{Command: "EXPIRE", Body: []string{"non_existent_key", "1"}},
			},
			Result: []TestCaseResult{
				{Expected: "0"},
			},
		},
		{
			Name: "EXPIRE with invalid number of arguments",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"key_to_expire_4", "SomeValue"}},
				{Command: "EXPIRE", Body: []string{"key_to_expire_4", "1", "extra_argument"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{ErrorExpected: true, Expected: "(error) ERR Unsupported option extra_argument"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			for i, cmd := range tc.Commands {
				if tc.Name == "Check key after waiting for expiration" && cmd.Command == "GET" { // Wait longer for expiration check
					time.Sleep(4 * time.Second) // Longer than the expiration time
				}
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
