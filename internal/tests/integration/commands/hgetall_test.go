package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHGetAll(t *testing.T) {
	exec, err := NewHTTPCommandExecutor()
	if err != nil {
		t.Fatal(err)
	}

	defer exec.FlushDB()

	testCases := []TestCase{
		{
			Name: "HGETALL with a non-existent key",
			Commands: []HTTPCommand{
				{Command: "HGETALL", Body: []string{"user", "name"}},
			},
			Result: []TestCaseResult{
				{Expected: ""},
			},
		},
		{
			Name: "HGETALL with a valid key",
			Commands: []HTTPCommand{
				{Command: "HSET", Body: []string{"user", "name", "John Doe"}},
				{Command: "HGETALL", Body: []string{"user"}},
			},
			Result: []TestCaseResult{
				{Expected: "1"},
				{Expected: "1) \"name\"\n2) \"John Doe\"\n"},
			},
		},
		{
			Name: "HGETALL with an invalid key",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"user1", "John Doe"}},
				{Command: "HGETALL", Body: []string{"user1"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{ErrorExpected: true, Expected: "(error) WRONGTYPE Operation against a key holding the wrong kind of value"},
			},
		},
		{
			Name: "HGETALL with invalid number of arguments",
			Commands: []HTTPCommand{
				{Command: "HGETALL", Body: []string{"user", "name"}},
			},
			Result: []TestCaseResult{
				{ErrorExpected: true, Expected: "(error) ERR wrong number of arguments for 'hgetall' command"},
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
				if result.ErrorExpected {
					assert.NotNil(t, err)
					assert.Equal(t, result.Expected, err.Error())
				} else {
					assert.Equal(t, result.Expected, response)
				}
			}
		})
	}
}
