package commands

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIncr(t *testing.T) {
	exec, err := NewHTTPCommandExecutor()
	if err != nil {
		t.Fatal(err)
	}

	defer exec.FlushDB()

	testCases := []TestCase{
		{
			Name: "Increment multiple keys",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"key1", "0"}},
				{Command: "INCR", Body: []string{"key1"}},
				{Command: "INCR", Body: []string{"key1"}},
				{Command: "INCR", Body: []string{"key2"}},
				{Command: "GET", Body: []string{"key1"}},
				{Command: "GET", Body: []string{"key2"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "(integer) 1"},
				{Expected: "(integer) 2"},
				{Expected: "(integer) 1"},
				{Expected: "(integer) 2"},
				{Expected: "(integer) 1"},
			},
		},
		{
			Name: "Increment from min int64",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"min_int", strconv.Itoa(math.MinInt64)}},
				{Command: "INCR", Body: []string{"min_int"}},
				{Command: "INCR", Body: []string{"min_int"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "(integer) " + strconv.FormatInt(math.MinInt64+1, 10)},
				{Expected: "(integer) " + strconv.FormatInt(math.MinInt64+2, 10)},
			},
		},
		{
			Name: "Increment non-existent key",
			Commands: []HTTPCommand{
				{Command: "INCR", Body: []string{"non_existent"}},
				{Command: "GET", Body: []string{"non_existent"}},
				{Command: "INCR", Body: []string{"non_existent"}},
			},
			Result: []TestCaseResult{
				{Expected: "(integer) 1"},
				{Expected: "(integer) 1"},
				{Expected: "(integer) 2"},
			},
		},
		{
			Name: "Increment string representing integers",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"str_int1", "42"}},
				{Command: "INCR", Body: []string{"str_int1"}},
				{Command: "SET", Body: []string{"str_int2", "-10"}},
				{Command: "INCR", Body: []string{"str_int2"}},
				{Command: "SET", Body: []string{"str_int3", "0"}},
				{Command: "INCR", Body: []string{"str_int3"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "(integer) 43"},
				{Expected: "OK"},
				{Expected: "(integer) -9"},
				{Expected: "OK"},
				{Expected: "(integer) 1"},
			},
		},
		{
			Name: "Increment with expiry",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"expiry_key", "0", "EX", "1"}},
				{Command: "INCR", Body: []string{"expiry_key"}},
				{Command: "INCR", Body: []string{"expiry_key"}},
				{Command: "GET", Body: []string{"expiry_key"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "(integer) 1"},
				{Expected: "(integer) 2"},
				{Expected: "(nil)"},
			},
			Delays: []time.Duration{0, 0, 0, 2 * time.Second},
		},
		{
			Name: "Increment non-integer values",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"float_key", "3.14"}},
				{Command: "INCR", Body: []string{"float_key"}},
				{Command: "SET", Body: []string{"string_key", "hello"}},
				{Command: "INCR", Body: []string{"string_key"}},
				{Command: "SET", Body: []string{"bool_key", "true"}},
				{Command: "INCR", Body: []string{"bool_key"}},
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{ErrorExpected: true, Expected: "(error) ERR value is not an integer or out of range"},
				{Expected: "OK"},
				{ErrorExpected: true, Expected: "(error) ERR value is not an integer or out of range"},
				{Expected: "OK"},
				{ErrorExpected: true, Expected: "(error) ERR value is not an integer or out of range"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			for i, cmd := range tc.Commands {
				if tc.Delays != nil && tc.Delays[i] > 0 {
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
