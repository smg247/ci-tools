package jobruntestcaseanalyzer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestJobRunsTestCaseAnalyzerFlagsValidateTimeout(t *testing.T) {
	testCases := []struct {
		name        string
		timeout     time.Duration
		expectedErr string
	}{
		{
			name:    "release analysis timeout is accepted",
			timeout: 5*time.Hour + 50*time.Minute,
		},
		{
			name:    "maximum timeout is accepted",
			timeout: 6 * time.Hour,
		},
		{
			name:        "timeout above maximum is rejected",
			timeout:     6*time.Hour + time.Nanosecond,
			expectedErr: "timeout value of 6h0m0.000000001s is out of range, valid value should be at most 6h0m0s",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			flags := NewJobRunsTestCaseAnalyzerFlags()
			flags.Authentication.GoogleServiceAccountCredentialFile = "credentials.json"
			flags.PayloadTag = "5.1.0-0.nightly-test"
			flags.TestGroup = installTestGroup
			flags.Timeout = testCase.timeout

			err := flags.Validate()
			if testCase.expectedErr == "" {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, testCase.expectedErr)
		})
	}
}
