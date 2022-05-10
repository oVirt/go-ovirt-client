package ovirtclient_test

import (
	"flag"
	"fmt"
	"os"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v3"
)

var getHelper func(t *testing.T) ovirtclient.TestHelper

func getHelperLive(t *testing.T) ovirtclient.TestHelper {
	helper, err := ovirtclient.NewLiveTestHelperFromEnv(ovirtclientlog.NewTestLogger(t))
	if err != nil {
		t.Fatal(fmt.Errorf("failed to create live test helper (%w)", err))
	}
	return helper
}

func getHelperMock(t *testing.T) ovirtclient.TestHelper {
	helper, err := ovirtclient.NewMockTestHelper(ovirtclientlog.NewTestLogger(t))
	if err != nil {
		t.Fatal(fmt.Errorf("failed to create mock test helper (%w)", err))
	}
	return helper
}

func TestMain(m *testing.M) {

	flagValueClientMock := "mock"
	flagValueClientLive := "live"
	flagValueClientAll := "all"

	clientFlag := flag.String("client", flagValueClientAll,
		"Client to use for running the tests. \n"+
			"Supported values: \n"+
			fmt.Sprintf("\t%s\t: Run tests with mock client \n", flagValueClientMock)+
			fmt.Sprintf("\t%s\t: Run tests with live client \n", flagValueClientLive)+
			fmt.Sprintf("\t%s\t: Run tests with mock and live client \n", flagValueClientAll),
	)
	flag.Parse()

	switch *clientFlag {
	case flagValueClientLive:
		getHelper = getHelperLive
		exitVal := m.Run()
		os.Exit(exitVal)
	case flagValueClientMock:
		getHelper = getHelperMock
		exitVal := m.Run()
		os.Exit(exitVal)
	case flagValueClientAll:
		getHelper = getHelperMock
		exitVal := m.Run()
		if exitVal != 0 {
			os.Exit(exitVal)
		}
		getHelper = getHelperLive
		exitVal = m.Run()
		os.Exit(exitVal)
	default:
		panic(fmt.Errorf("Unsupported client '%s'", *clientFlag))
	}
}
