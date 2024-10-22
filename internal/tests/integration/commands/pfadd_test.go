package commands

import (
	"server/internal/tests/integration/commands/assertions"
	"testing"
)

func TestPfAdd(t *testing.T) {
	exec, err := NewHTTPCommandExecutor()

	if err != nil {
		t.Fatal(err)
	}

	testCases := []TestCase{
		{
			Name: "Adding a single element to a HyperLogLog",
			Commands: []HTTPCommand{
				{Command: "PFADD", Body: []string{"myhyperloglog", "element1"}},
			},
			Result: []TestCaseResult{
				{Expected: "(integer) 1"},
			},
		},
		{
			Name: "Adding multiple elements to a HyperLogLog",
			Commands: []HTTPCommand{
				{Command: "PFADD", Body: []string{"myhyperloglog", "element1", "element2", "element3"}},
			},
			Result: []TestCaseResult{
				{Expected: "(integer) 1"},
			},
		},
		{
			Name: "Checking if HyperLogLog was modified (element doesn't alter internal registers)",
			Commands: []HTTPCommand{
				{Command: "PFADD", Body: []string{"myhyperloglog", "element1"}},
			},
			Result: []TestCaseResult{
				{Expected: "(integer) 0"},
			},
		},
		{
			Name: "Adding to a key that is not a HyperLogLog",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"mykey", "notahyperloglog"}},
				{Command: "PFADD", Body: []string{"mykey", "element1"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{ErrorExpected: true, Expected: "(error) WRONGTYPE Key is not a valid HyperLogLog string value"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			for i, cmd := range tc.Commands {
				response, err := exec.FireCommand(cmd)
				if err != nil {
					t.Logf("Error in executing command: %s - %v", cmd.Command, err)
				} else {
					t.Logf("Response for command %s: %s", cmd.Command, response)
				}

				result := tc.Result[i]
				assertions.AssertResult(t, err, response, result.Expected, result.ErrorExpected)
			}
		})
	}
}
