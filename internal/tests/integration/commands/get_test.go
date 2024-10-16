package commands

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	exec, err := NewHTTPCommandExecutor()
	if err != nil {
		t.Fatal(err)
	}

	defer exec.FlushDB()

	testCases := []TestCase{
		{
			Name: "Get with expiration",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"k", "v", "EX", "4"}},
				{Command: "GET", Body: []string{"k"}},
				{Command: "GET", Body: []string{"k"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "v"},
				{Expected: "(nil)"},
			},
			Delays: []time.Duration{0, 0, 5 * time.Second},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			for i, cmd := range tc.Commands {
				if tc.Delays[i] > 0 {
					time.Sleep(tc.Delays[i])
				}
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
