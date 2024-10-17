package commands

import (
	"server/internal/tests/integration/commands/assertions"
	"strconv"
	"testing"
	"time"
)

func TestExpireTime(t *testing.T) {
	exec, err := NewHTTPCommandExecutor()
	if err != nil {
		t.Fatal(err)
	}

	defer exec.FlushDB()

	testCases := []TestCase{
		{
			Name: "EXPIRETIME on a non-existent key",
			Commands: []HTTPCommand{
				{Command: "EXPIRETIME", Body: []string{"non_existent_key"}},
			},
			Result: []TestCaseResult{
				{Expected: "(integer) -2"}, // Expecting -2 because the key does not exist
			},
		},
		{
			Name: "EXPIRETIME on an existing key with future expiration",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"temp_key", "temp_value"}},
				{Command: "EXPIREAT", Body: []string{"temp_key", strconv.FormatInt(time.Now().Add(30*time.Second).Unix(), 10)}}, // Set future expiration
				{Command: "EXPIRETIME", Body: []string{"temp_key"}},                                                             // Retrieve expiration time
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "(integer) 1"}, // Indicating the EXPIREAT command was successful
				{Expected: "(integer) " + strconv.FormatInt(time.Now().Add(30*time.Second).Unix(), 10)}, // Future timestamp in seconds
			},
		},
		{
			Name: "EXPIRETIME on an existing key without expiration",
			Commands: []HTTPCommand{
				{Command: "SET", Body: []string{"persist_key", "persistent_value"}},
				{Command: "EXPIRETIME", Body: []string{"persist_key"}}, // Check expiration time
			},
			Result: []TestCaseResult{
				{Expected: "OK"},
				{Expected: "(integer) -1"}, // Expecting -1 because no expiration is set
			},
		},
		{
			Name: "EXPIRETIME with invalid arguments",
			Commands: []HTTPCommand{
				{Command: "EXPIRETIME", Body: []string{}}, // Missing key argument
			},
			Result: []TestCaseResult{
				{ErrorExpected: true, Expected: "(error) ERR wrong number of arguments for 'expiretime' command"},
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
