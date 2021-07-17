package ovirtclient

import (
	"context"
)

func (m *mockClient) CreateVM(
	ctx context.Context,
	clusterID string,
	cpuTopo VMCPUTopo,
	templateID string,
	blockDevices []VMBlockDevice,
) {
	// TODO implement create VM
	panic("implement me")
}
