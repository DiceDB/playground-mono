package commands

import (
	"server/internal/tests/integration/commands/assertions"
	"testing"
)

func TestExists(t *testing.T) {
	exec, err := NewHTTPCommandExecutor()
	if err != nil {
		t.Fatal(err)
	}

	defer exec.FlushDB()

	testCases := []TestCase{
		{
			Name: "EXISTS with a non-existent key",
			Commands: []HTTPCommand{
				{Command: "EXISTS", Body: []string{"non_existent_key"}},
			},
			Result: []TestCaseResult{
				{Expected: "0"}, // Expecting 0 because the key should not exist
			},
		},
		{
			Name: "EXISTS with an existing key",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"existing_key", "SomeValue"}},
				{Command: "EXISTS", Body: []string{"existing_key"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"}, // Expecting "OK" from the SET command
				{Expected: "1"},  // Expecting 1 because the key should exist
			},
		},
		{
			Name: "EXISTS with multiple keys where some exist",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"key1", "Value1"}},
				{Command: "SET", Body: []string{"key2", "Value2"}},
				{Command: "EXISTS", Body: []string{"key1", "key2", "non_existent_key"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "OK"},
				{Expected: "2"}, // Expecting 2 because only key1 and key2 exist
			},
		},
		{
			Name: "EXISTS command with invalid number of arguments",
			Commands: []HTTPCommand{
				{Command: "EXISTS", Body: []string{}}, // No arguments
			},
			Result: []TestCaseResult{
				{ErrorExpected: true, Expected: "(error) ERR wrong number of arguments for 'exists' command"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			for i, cmd := range tc.Commands {
				response, err := exec.FireCommand(cmd)
				if err != nil {
					t.Logf("error in executing command: %s - %v", cmd.Command, err)
				}

				result := tc.Result[i]
				assertions.AssertResult(t, err, response, result.Expected, result.ErrorExpected)

			}
		})
	}
}
