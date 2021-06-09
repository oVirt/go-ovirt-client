package govirt

import (
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

// TestHelper is a helper to run tests against an oVirt engine. When created it scans the oVirt Engine and tries to find
// working resources (hosts, clusters, etc) for running tests against. Tests should clean up after themselves.
type TestHelper interface {
	// GetClient returns the goVirt client.
	GetClient() Client

	// GetSDKClient returns the oVirt SDK client.
	GetSDKClient() *ovirtsdk4.Connection

	// GetCluster returns the oVirt cluster for testing purposes.
	GetCluster() *ovirtsdk4.Cluster

	// GetBlankTemplateID returns the ID of the blank template that can be used for creating dummy VMs.
	GetBlankTemplateID() string

	// GetStorageDomainID returns the ID of the storage domain to create the images on.
	GetStorageDomainID() string

	// GetHostCount returns the number of hosts in the oVirt cluster used for testing.
	GetHostCount() uint

	// GetMACPoolList returns a list of MAC pool names.
	GetMACPoolList() ([]string, error)

	// GetAuthzName returns the name of an authz that can be used for testing purposes.
	GetAuthzName() string

	// GetDatacenterName returns the name of a datacenter that can be used for testing purposes.
	GetDatacenterName() string

	// GetDatacenterID returns the ID of the test datacenter
	GetDatacenterID() string

	// GenerateRandomID generates a random ID for testing.
	GenerateRandomID(length uint) string
}

func NewTestHelper(
	url string,
	username string,
	password string,
	caFile string,
	caBundle []byte,
	insecure bool,
	clusterID string,
	blankTemplateID string,
	storageDomainID string,
	logger Logger,
) (TestHelper, error) {
	client, err := New(
		url,
		username,
		password,
		caFile,
		caBundle,
		insecure,
		nil,
		logger,
	)
	if err != nil {
		return nil, err
	}
	return &testHelper{
		client: client,
	}, nil
}

func MustNewTestHelper(
	username string,
	password string,
	url string,
	insecure bool,
	caFile string,
	caBundle []byte,
	clusterID string,
	blankTemplateID string,
	storageDomainID string,
	logger Logger,
) TestHelper {
	helper, err := NewTestHelper(
		url,
		username,
		password,
		caFile,
		caBundle,
		insecure,
		clusterID,
		blankTemplateID,
		storageDomainID,
		logger,
	)
	if err != nil {
		panic(err)
	}
	return helper
}

type testHelper struct {
	client Client
}

func (t *testHelper) GetClient() Client {
	return t.client
}

func (t *testHelper) GetSDKClient() *ovirtsdk4.Connection {
	return t.client.GetSDKClient()
}

func (t *testHelper) GetCluster() *ovirtsdk4.Cluster {
	panic("implement me")
}

func (t *testHelper) GetBlankTemplateID() string {
	panic("implement me")
}

func (t *testHelper) GetStorageDomainID() string {
	panic("implement me")
}

func (t *testHelper) GetHostCount() uint {
	panic("implement me")
}

func (t *testHelper) GetMACPoolList() ([]string, error) {
	panic("implement me")
}

func (t *testHelper) GetAuthzName() string {
	panic("implement me")
}

func (t *testHelper) GetDatacenterName() string {
	panic("implement me")
}

func (t *testHelper) GetDatacenterID() string {
	panic("implement me")
}

func (t *testHelper) GenerateRandomID(length uint) string {
	panic("implement me")
}
