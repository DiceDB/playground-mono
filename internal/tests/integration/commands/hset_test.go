package commands

import (
	"server/internal/tests/integration/commands/assertions"
	"testing"
)

func TestHSet(t *testing.T) {
	exec, err := NewHTTPCommandExecutor()
	if err != nil {
		t.Fatal(err)
	}

	testCases := []TestCase{
		{
			Name: "HSET with simple key value pairs in the hash",
			Commands: []HTTPCommand{
				{Command: "HSET", Body: []string{"user", "name", "John Doe", "age", "30"}},
			},
			Result: []TestCaseResult{
				{Expected: "(integer) 2"},
			},
		},
		{
			Name: "HSET update one key and set a new key",
			Commands: []HTTPCommand{
				{Command: "HSET", Body: []string{"user1", "name", "John Doe", "age", "30"}},
				{Command: "HSET", Body: []string{"user1", "name", "John Loe", "gender", "Male"}},
			},
			Result: []TestCaseResult{
				{Expected: "(integer) 2"},
				{Expected: "(integer) 1"},
			},
		},
		{
			Name: "HSET with invalid number of arguments",
			Commands: []HTTPCommand{
				{Command: "HSET", Body: []string{"user", "name", "John Loe", "gender"}},
			},
			Result: []TestCaseResult{
				{ErrorExpected: true, Expected: "(error) ERR wrong number of arguments for 'HSET' command"},
			},
		},
		{
			Name: "HSET with invalid key",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"user2", "John Doe"}},
				{Command: "HSET", Body: []string{"user2", "name", "John Doe"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{ErrorExpected: true, Expected: "(error) WRONGTYPE Operation against a key holding the wrong kind of value"},
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
