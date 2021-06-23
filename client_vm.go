package ovirtclient

import (
	"context"
	"fmt"
)

type VMClient interface {
	CreateVM(
		ctx context.Context,
		clusterID string,
		cpuTopo VMCPUTopo,
		templateID string,
		blockDevices []VMBlockDevice,
	)
}

// NewVMCPUTopo creates a new CPU topology with the given parameters. It returns an error if cores, threads, or sockets
// is 0. If the parameters are guaranteed to be non-zero MustNewVMCPUTopo should be used.
func NewVMCPUTopo(cores int64, threads int64, sockets int64) (VMCPUTopo, error) {
	if cores == 0 {
		return nil, fmt.Errorf("BUG: cores cannot be zero")
	}
	if threads == 0 {
		return nil, fmt.Errorf("BUG: threads cannot be zero")
	}
	if sockets == 0 {
		return nil, fmt.Errorf("BUG: sockets cannot be zero")
	}
	return &vmCPUTopo{
		cores:   cores,
		threads: threads,
		sockets: sockets,
	}, nil
}

// MustNewVMCPUTopo is identical to NewVMCPUTopo, but panics instead of returning an error if cores, threads, or
// sockets is zero.
func MustNewVMCPUTopo(cores int64, threads int64, sockets int64) VMCPUTopo {
	topo, err := NewVMCPUTopo(cores, threads, sockets)
	if err != nil {
		panic(err)
	}
	return topo
}

type VMCPUTopo interface {
	Cores() int64
	Threads() int64
	Sockets() int64
}

type vmCPUTopo struct {
	cores   int64
	threads int64
	sockets int64
}

func (v *vmCPUTopo) Cores() int64 {
	return v.cores
}

func (v *vmCPUTopo) Threads() int64 {
	return v.threads
}

func (v *vmCPUTopo) Sockets() int64 {
	return v.sockets
}

type VMBlockDevice interface {
	DiskID() string
	Bootable() bool

	StorageDomainID() string
}

type VMInitialization interface {
	CustomScript() string
	HostName() string
}

type vmInitialization struct {
	customScript string
	hostName string
}

func (v *vmInitialization) CustomScript() string {
	return v.customScript
}

func (v *vmInitialization) HostName() string {
	return v.hostName
}