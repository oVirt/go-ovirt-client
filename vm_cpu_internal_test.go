package ovirtclient

import "testing"

func TestVMCPUTopoNil(t *testing.T) {
	t.Parallel()

	if topo := (vmCPU{}).Topo(); topo != nil {
		t.Fatalf("expected a nil topology for a CPU without one, got %#v", topo)
	}
}
