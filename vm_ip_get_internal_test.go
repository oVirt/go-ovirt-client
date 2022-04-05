package ovirtclient

import (
	"net"
	"regexp"
	"testing"
)

var filterReportedIPListInput = map[string][]net.IP{
	"lo": {
		net.ParseIP("127.0.0.1"),
		net.ParseIP("::1"),
	},
	"eth0": {
		net.ParseIP("192.168.0.2"),
		net.ParseIP("fe80::2"),
	},
}

var filterReportedIPListCases = map[string]filterReportedIPListTestCase{
	"empty": {
		params:         NewVMIPSearchParams(),
		expectedResult: filterReportedIPListInput,
	},
	"include_interface_name": {
		params: NewVMIPSearchParams().WithIncludedInterface("eth0"),
		expectedResult: map[string][]net.IP{
			"eth0": filterReportedIPListInput["eth0"],
		},
	},
	"include_interface_pattern": {
		params: NewVMIPSearchParams().WithIncludedInterfacePattern(regexp.MustCompile("^eth[0-9]+$")),
		expectedResult: map[string][]net.IP{
			"eth0": filterReportedIPListInput["eth0"],
		},
	},
	"included_interface_name_and_pattern": {
		params: NewVMIPSearchParams().
			WithIncludedInterface("eth0").
			WithIncludedInterfacePattern(regexp.MustCompile("^asdf$")),
		expectedResult: map[string][]net.IP{
			"eth0": filterReportedIPListInput["eth0"],
		},
	},
	"exclude_interface_name": {
		params: NewVMIPSearchParams().WithExcludedInterface("eth0"),
		expectedResult: map[string][]net.IP{
			"lo": filterReportedIPListInput["lo"],
		},
	},
	"exclude_interface_pattern": {
		params: NewVMIPSearchParams().WithExcludedInterfacePattern(regexp.MustCompile("^eth[0-9]+$")),
		expectedResult: map[string][]net.IP{
			"lo": filterReportedIPListInput["lo"],
		},
	},
	"include_ip_range": {
		params: NewVMIPSearchParams().WithIncludedRange(mustParseCIDR("192.168.0.0/16")),
		expectedResult: map[string][]net.IP{
			"eth0": {
				net.ParseIP("192.168.0.2"),
			},
		},
	},
	"exclude_ip_range": {
		params: NewVMIPSearchParams().WithExcludedRange(mustParseCIDR("192.168.0.0/16")),
		expectedResult: map[string][]net.IP{
			"lo": filterReportedIPListInput["lo"],
			"eth0": {
				net.ParseIP("fe80::2"),
			},
		},
	},
}

func TestFilterReportedIPList(t *testing.T) {
	for testCaseName, testCase := range filterReportedIPListCases {
		t.Run(
			testCaseName, func(t *testing.T) {
				testCase.run(t, filterReportedIPListInput)
			},
		)
	}
}

type filterReportedIPListTestCase struct {
	params         VMIPSearchParams
	expectedResult map[string][]net.IP
}

func (c filterReportedIPListTestCase) run(t *testing.T, input map[string][]net.IP) {
	result := filterReportedIPList(input, c.params)
	if len(result) != len(c.expectedResult) {
		t.Fatalf("Incorrect number of results (expected: %d, got: %d)", len(c.expectedResult), len(result))
	}
	for expectedIfName, expectedIPs := range c.expectedResult {
		gotIPs, ok := result[expectedIfName]
		if !ok {
			t.Fatalf("Expected interface %s not found in result.", expectedIfName)
		}
		if len(gotIPs) != len(expectedIPs) {
			t.Fatalf(
				"Incorrect number of IP addresses on interface %s (expected: %d, got: %d)",
				expectedIfName,
				len(expectedIPs),
				len(gotIPs),
			)
		}
		for i, ip := range expectedIPs {
			if ip.String() != gotIPs[i].String() {
				t.Fatalf(
					"Incorrect IP on interface %s position %d (expected: %s, got: %s)",
					expectedIfName,
					i,
					ip.String(),
					gotIPs[i].String(),
				)
			}
		}
	}
}
