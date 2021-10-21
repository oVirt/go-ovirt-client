package ovirtclient_test

import (
	"fmt"
	ovirtclient "github.com/ovirt/go-ovirt-client"
	"testing"
)

func (o *oVirtClient) CreateTemplate(
	vmid string,
	vnicProfileID string,
	name string,
	_ OptionalNICParameters,
	retries ...RetryStrategy,
) (result NIC, err error) {


}