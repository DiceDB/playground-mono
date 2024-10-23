package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	exec, err := NewHTTPCommandExecutor()
	if err != nil {
		t.Fatal(err)
	}

	defer exec.FlushDB()

	testCases := []TestCase{
		{
			Name: "Set and Get Simple Value",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"k", "v"}},
				{Command: "GET", Body: []string{"k"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "\"v\""},
			},
		},
		{
			Name: "Set and Get Integer Value",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"k", "123456789"}},
				{Command: "GET", Body: []string{"k"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "(integer) 123456789"}, // This is Redis' scientific notation for large numbers
			},
		},
		{
			Name: "Overwrite Existing Key",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"k", "v1"}},
				{Command: "SET", Body: []string{"k", "5"}},
				{Command: "GET", Body: []string{"k"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "OK"},
				{Expected: "(integer) 5"}, // As the value 5 is stored as a string
			},
		},
		{
			Name: "Set with EX and PX option",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"k", "v", "EX", "2", "PX", "2000"}},
			},
			Result: []TestCaseResult{
				{ErrorExpected: true, Expected: "(error) ERR syntax error"},
			},
		},
		{
			Name: "XX on non-existing key",
			Commands: []HTTPCommand{
				{Command: "DEL", Body: []string{"a"}},
				{Command: "SET", Body: []string{"a", "v", "XX"}},
				{Command: "GET", Body: []string{"a"}},
			},
			Result: []TestCaseResult{
				{Expected: "(integer) 0"}, // DEL returns number of deleted keys
				{Expected: "(nil)"},
				{Expected: "(nil)"},
			},
		},
		{
			Name: "NX on non-existing key",
			Commands: []HTTPCommand{
				{Command: "DEL", Body: []string{"c"}},
				{Command: "SET", Body: []string{"c", "v", "NX"}},
				{Command: "GET", Body: []string{"c"}},
			},
			Result: []TestCaseResult{
				{Expected: "(integer) 0"}, // DEL returns number of deleted keys
				{Expected: "OK"},
				{Expected: "\"v\""},
			},
		},
		{
			Name: "NX on existing key",
			Commands: []HTTPCommand{
				{Command: "DEL", Body: []string{"b"}},
				{Command: "SET", Body: []string{"b", "v", "NX"}},
				{Command: "GET", Body: []string{"b"}},
				{Command: "SET", Body: []string{"b", "v", "NX"}},
			},
			Result: []TestCaseResult{
				{Expected: "(integer) 0"}, // DEL returns number of deleted keys
				{Expected: "OK"},
				{Expected: "\"v\""},
				{Expected: "(nil)"}, // NX fails because the key already exists
			},
		},
		{
			Name: "PXAT option with invalid unix time ms",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"k2", "v2", "PXAT", "123123"}},
				{Command: "GET", Body: []string{"k2"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "(nil)"}, // Invalid time causes key not to be set properly
			},
		},
		{
			Name: "XX on existing key",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"k", "v1"}},
				{Command: "SET", Body: []string{"k", "v2", "XX"}},
				{Command: "GET", Body: []string{"k"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "OK"},
				{Expected: "\"v2\""},
			},
		},
		{
			Name: "Multiple XX operations",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"k", "v1"}},
				{Command: "SET", Body: []string{"k", "v2", "XX"}},
				{Command: "SET", Body: []string{"k", "v3", "XX"}},
				{Command: "GET", Body: []string{"k"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "OK"},
				{Expected: "OK"},
				{Expected: "\"v3\""},
			},
		},
		{
			Name: "EX option",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"k", "v", "EX", "1"}},
				{Command: "GET", Body: []string{"k"}},
				{Command: "SLEEP", Body: []string{"2"}},
				{Command: "GET", Body: []string{"k"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "\"v\""},
				{Expected: "OK"},
				{Expected: "(nil)"}, // After expiration
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
