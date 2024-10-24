package commands

import (
	"server/internal/tests/integration/commands/assertions"
	"testing"
)

func TestPfCount(t *testing.T) {
	exec, err := NewHTTPCommandExecutor()
	if err != nil {
		t.Fatal(err)
	}
	testCases := []TestCase{
		{
			Name: "PFCOUNT on non-existent key",
			Commands: []HTTPCommand{
				{Command: "PFCOUNT", Body: []string{"non_existent_key"}},
			},
			Result: []TestCaseResult{
				{Expected: "(integer) 0"},
			},
		},
		{
			Name: "PFCOUNT on wrong type of key",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"mykey", "value"}},
				{Command: "PFCOUNT", Body: []string{"mykey"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{ErrorExpected: true, Expected: "(error) WRONGTYPE Key is not a valid HyperLogLog string value"},
			},
		},
		{
			Name: "PFCOUNT with invalid arguments (no arguments)",
			Commands: []HTTPCommand{
				{Command: "PFCOUNT", Body: []string{}},
			},
			Result: []TestCaseResult{
				{ErrorExpected: true, Expected: "(error) ERR wrong number of arguments for 'pfcount' command"},
			},
		},
		{
			Name: "PFCOUNT on single key",
			Commands: []HTTPCommand{
				{Command: "PFADD", Body: []string{"hll1", "foo", "bar", "baz"}},
				{Command: "PFCOUNT", Body: []string{"hll1"}},
			},
			Result: []TestCaseResult{
				{Expected: "(integer) 1"},
				{Expected: "(integer) 3"},
			},
		},
		{
			Name: "PFCOUNT on multiple keys",
			Commands: []HTTPCommand{
				{Command: "PFADD", Body: []string{"hll1", "foo", "bar"}},
				{Command: "PFADD", Body: []string{"hll2", "baz", "qux"}},
				{Command: "PFCOUNT", Body: []string{"hll1", "hll2"}},
			},
			Result: []TestCaseResult{
				{Expected: "(integer) 0"},
				{Expected: "(integer) 1"},
				{Expected: "(integer) 4"},
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
