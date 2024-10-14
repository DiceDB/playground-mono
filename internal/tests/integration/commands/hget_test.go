package commands

import (
	"server/internal/tests/integration/commands/assertions"
	"testing"
)

func TestHGet(t *testing.T) {
	exec, err := NewHTTPCommandExecutor()
	if err != nil {
		t.Fatal(err)
	}

	defer exec.FlushDB()

	testCases := []TestCase{
		{
			Name: "HGET with a non-existent key",
			Commands: []HTTPCommand{
				{Command: "HGET", Body: []string{"user", "name"}},
			},
			Result: []TestCaseResult{
				{Expected: "(nil)"},
			},
		},
		{
			Name: "HGET with a valid field in the key",
			Commands: []HTTPCommand{
				{Command: "HSET", Body: []string{"user", "name", "John Doe", "age", "30"}},
				{Command: "HGET", Body: []string{"user", "name"}},
			},
			Result: []TestCaseResult{
				{Expected: "2"},
				{Expected: "John Doe"},
			},
		},
		{
			Name: "HGET with an invalid field in the key",
			Commands: []HTTPCommand{
				{Command: "HSET", Body: []string{"user1", "name", "John Doe", "age", "30"}},
				{Command: "HGET", Body: []string{"user1", "gender"}},
			},
			Result: []TestCaseResult{
				{Expected: "2"},
				{Expected: "(nil)"},
			},
		},
		{
			Name: "HGET with an invalid key",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"user2", "John Doe"}},
				{Command: "HGET", Body: []string{"user2", "name"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{ErrorExpected: true, Expected: "(error) WRONGTYPE Operation against a key holding the wrong kind of value"},
			},
		},
		{
			Name: "HGET with invalid number of arguments",
			Commands: []HTTPCommand{
				{Command: "HGET", Body: []string{"user2", "name", "age"}},
			},
			Result: []TestCaseResult{
				{ErrorExpected: true, Expected: "(error) ERR wrong number of arguments for 'hget' command"},
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
