package ovirtclient_test

import (
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client/v2"
)

func TestFeatureFlags(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	testcases := map[string]struct {
		input ovirtclient.Feature
	}{
		"Feature Autopinning": {
			input: ovirtclient.FeatureAutoPinning,
		},
		"Feature Placement Policy": {
			input: ovirtclient.FeatureAutoPinning,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			supported, err := helper.GetClient().SupportsFeature(tc.input)
			if err != nil {
				t.Fatalf("Failed to check '%s' support (%v)", tc.input, err)
			}
			if supported {
				t.Logf("'%s' is supported.", tc.input)
			} else {
				t.Logf("'%s' is not supported.", tc.input)
			}
		})
	}
}
