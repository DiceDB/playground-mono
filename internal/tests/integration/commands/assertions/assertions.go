package assertions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertResult checks the result of command execution against expected results.
// Parameters:
// - t: the testing context used for reporting errors.
// - err: the error returned from the command execution.
// - response: the response obtained from the command.
// - expected: the expected response string.
// - errorExpected: a flag indicating whether an error is expected.
func AssertResult(t *testing.T, err error, response, expected string, errorExpected bool) {
	if errorExpected {
		// Assert that an error occurred and check the error message.
		assert.Error(t, err, "Expected an error but got none")
		assert.EqualError(t, err, expected, "Error message does not match the expected message")
	} else {
		// Assert that no error occurred and check the response.
		assert.NoError(t, err, "Expected no error but got one")
		assert.Equal(t, expected, response, "Response does not match the expected value")
	}
}
