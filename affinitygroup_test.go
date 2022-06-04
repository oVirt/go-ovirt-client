package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func assertCanCreateAffinityGroup(t *testing.T, helper ovirtclient.TestHelper, params ovirtclient.CreateAffinityGroupOptionalParams) ovirtclient.AffinityGroup {
	client := helper.GetClient()

	testClusterID := helper.GetClusterID()

	affinityGroup, err := client.CreateAffinityGroup(
		testClusterID,
		fmt.Sprintf("test_ag_%s", helper.GenerateRandomID(5)),
		params,
	)
	if err != nil {
		t.Fatalf("Failed to create affinity group (%v)", err)
	}
	t.Cleanup(
		func() {
			if err := affinityGroup.Remove(); err != nil {
				if !ovirtclient.HasErrorCode(err, ovirtclient.ENotFound) {
					t.Fatalf("Failed to clean up affinity group %s (%v)", affinityGroup.ID(), err)
				}
			}
		},
	)
	return affinityGroup
}

func assertCanRemoveAffinityGroup(
	t *testing.T,
	helper ovirtclient.TestHelper,
	clusterID ovirtclient.ClusterID,
	id ovirtclient.AffinityGroupID,
) {
	client := helper.GetClient()

	if err := client.RemoveAffinityGroup(clusterID, id); err != nil {
		t.Fatalf("Failed to remove affinity group %s from cluster %s (%v)", id, clusterID, err)
	}
}

func assertCanGetAffinityGroup(
	t *testing.T,
	helper ovirtclient.TestHelper,
	clusterID ovirtclient.ClusterID,
	id ovirtclient.AffinityGroupID,
) ovirtclient.AffinityGroup {
	ag, err := helper.GetClient().GetAffinityGroup(clusterID, id)
	if err != nil {
		t.Fatalf("Failed to fetch affinity group %s from cluster %s (%v)", id, clusterID, err)
	}
	return ag
}

func assertAffinityGroupListHasAffinityGroup(
	t *testing.T,
	helper ovirtclient.TestHelper,
	clusterID ovirtclient.ClusterID,
	id ovirtclient.AffinityGroupID,
) {
	agList, err := helper.GetClient().ListAffinityGroups(clusterID)
	if err != nil {
		t.Fatalf("Failed to list affinity groups in cluster %s (%v)", clusterID, err)
	}
	for _, ag := range agList {
		if ag.ID() == id {
			return
		}
	}
	t.Fatalf("Failed to find affinity group %s in cluster %s", id, clusterID)
}

func TestAffinityGroupCreation(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	ag := assertCanCreateAffinityGroup(t, helper, nil)

	if ag.ID() == "" {
		t.Fatalf("Returned affinity group ID is empty.")
	}

	if ag.Name() == "" {
		t.Fatalf("Returned affinity group name is empty.")
	}

	if ag.Enforcing() == true {
		t.Fatalf("Returned affinity group enforcing flag is true.")
	}

	if ag.Priority() != 1 {
		t.Fatalf("Returned affinity group priority is %f, not 1.", ag.Priority())
	}

	if ag.HostsRule().Enabled() != false {
		t.Fatalf("Hosts rule is enabled on returned affinity group.")
	}
	if ag.VMsRule().Enabled() != false {
		t.Fatalf("VMs rule is enabled on returned affinity group.")
	}

	secondAG := assertCanGetAffinityGroup(t, helper, ag.ClusterID(), ag.ID())

	if ag.ID() != secondAG.ID() {
		t.Fatalf("Affinity group IDs don't match (%s != %s)", ag.ID(), secondAG.ID())
	}

	if ag.Name() != secondAG.Name() {
		t.Fatalf("Affinity group names don't match (%s != %s)", ag.Name(), secondAG.Name())
	}

	assertAffinityGroupListHasAffinityGroup(t, helper, ag.ClusterID(), ag.ID())

	assertCanRemoveAffinityGroup(t, helper, ag.ClusterID(), ag.ID())
}

func TestAffinityGroupCreationWithDescription(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	params := ovirtclient.CreateAffinityGroupParams()
	params = params.MustWithDescription("test")

	ag := assertCanCreateAffinityGroup(t, helper, params)

	if ag.Description() != "test" {
		t.Fatalf("Incorrect affinity group description: %s", ag.Description())
	}
}

func TestNegativeVMAffinityShouldResultInDifferentHosts(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	ag := assertCanCreateAffinityGroup(
		t,
		helper,
		ovirtclient.
			CreateAffinityGroupParams().
			MustWithVMsRuleParameters(true, ovirtclient.AffinityNegative, true),
	)
	vm1 := assertCanCreateBootableVM(t, helper)
	vm2 := assertCanCreateBootableVM(t, helper)
	assertCanAddVMToAffinityGroup(t, vm1, ag)
	assertCanAddVMToAffinityGroup(t, vm2, ag)
	assertCanStartVM(t, helper, vm1)
	vm1 = assertVMWillStart(t, vm1)
	if err := vm2.Start(); err != nil {
		t.Logf("VM 2 failed to start, we assume there are not enough hosts available. (%v)", err)
		return
	}
	t.Cleanup(
		func() {
			if err := vm2.Stop(true); err != nil {
				t.Fatalf("Failed to stop VM %s (%v)", vm2.ID(), err)
			}
			if _, err := vm2.WaitForStatus(ovirtclient.VMStatusDown); err != nil {
				t.Fatalf("Failed to wait for VM %s to stop (%v)", vm2.ID(), err)
			}
		})

	vm2, err := vm2.WaitForStatus(ovirtclient.VMStatusUp)
	if err != nil {
		t.Logf("VM 2 failed to start, we assume there are not enough hosts available. (%v)", err)
		return
	}
	if *vm1.HostID() == *vm2.HostID() {
		t.Fatalf("Identical hosts for both VMs despite affinity group.")
	}
}

func assertCanAddVMToAffinityGroup(t *testing.T, vm1 ovirtclient.VM, ag ovirtclient.AffinityGroup) {
	if err := ag.AddVM(vm1.ID()); err != nil {
		t.Fatalf("Failed to add VM %s to affinity group %s (%v)", vm1.ID(), ag.ID(), err)
	}
}
