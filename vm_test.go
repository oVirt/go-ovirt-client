package ovirtclient_test

import (
	"fmt"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func TestVMListShouldNotFail(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	client := helper.GetClient()

	if _, err := client.ListVMs(); err != nil {
		t.Fatal(err)
	}
}
func TestGetVMByName(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	client := helper.GetClient()
	vmName := fmt.Sprintf("test-%s", helper.GenerateRandomID(5))

	assertCanCreateVM(
		t,
		helper,
		vmName,
		nil,
	)

	fetchedVM, err := client.GetVMByName(vmName)
	if err != nil {
		t.Fatal(err)
	}
	if fetchedVM == nil {
		t.Fatal("returned VM is nil")
	}

	t.Logf("fetched VM Name %s mismatches original created VM Name %s", fetchedVM.Name(), vmName)
	if fetchedVM.Name() != vmName {
		t.Fatalf("fetched VM Name %s mismatches original created VM Name %s", fetchedVM.Name(), vmName)
	}

}
func TestAfterVMCreationShouldBePresent(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	client := helper.GetClient()

	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
		nil,
	)
	fetchedVM, err := client.GetVM(vm.ID())
	if err != nil {
		t.Fatal(err)
	}
	if fetchedVM == nil {
		t.Fatal("returned VM is nil")
	}
	if fetchedVM.ID() != vm.ID() {
		t.Fatalf("fetched VM ID %s mismatches original created VM ID %s", fetchedVM.ID(), vm.ID())
	}

	updatedVM, err := fetchedVM.Update(
		ovirtclient.UpdateVMParams().MustWithName("new_name").MustWithComment("new comment").MustWithDescription("new description"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if updatedVM.ID() != vm.ID() {
		t.Fatalf("updated VM ID %s mismatches original created VM ID %s", updatedVM.ID(), vm.ID())
	}
	if updatedVM.Name() != "new_name" {
		t.Fatalf("updated VM name %s does not match update parameters", updatedVM.Name())
	}
	if updatedVM.Comment() != "new comment" {
		t.Fatalf("updated VM comment %s does not match update parameters", updatedVM.Comment())
	}
	if updatedVM.Description() != "new description" {
		t.Fatalf("updated VM description %s does not match update parameters", updatedVM.Description())
	}

	fetchedVM, err = client.GetVM(vm.ID())
	if err != nil {
		t.Fatal(err)
	}
	if fetchedVM == nil {
		t.Fatal("returned VM is nil")
	}
	if fetchedVM.Name() != "new_name" {
		t.Fatalf("updated VM name %s does not match update parameters", fetchedVM.Name())
	}
	if fetchedVM.Comment() != "new comment" {
		t.Fatalf("updated VM comment %s does not match update parameters", fetchedVM.Comment())
	}
	if fetchedVM.Description() != "new description" {
		t.Fatalf("updated VM description %s does not match update parameters", fetchedVM.Description())
	}
}

func TestVMCreationWithCPU(t *testing.T) {

	params := map[string]ovirtclient.OptionalVMParameters{
		"nocpu":   ovirtclient.CreateVMParams(),
		"withcpu": ovirtclient.CreateVMParams().MustWithCPUParameters(1, 1, 1),
	}
	for name, param := range params {
		t.Run(name, func(t *testing.T) {
			param := param
			t.Parallel()
			helper := getHelper(t)
			vm := assertCanCreateVM(
				t,
				helper,
				fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
				param,
			)

			cpu := vm.CPU()
			if cpu == nil {
				t.Fatalf("Creating a VM with CPU settings did not return a VM with CPU.")
			}

			topo := cpu.Topo()
			if topo == nil {
				t.Fatalf("Creating a VM with CPU settings did not return a CPU topology.")
			}

			if cores := topo.Cores(); cores != 1 {
				t.Fatalf("Creating a VM with 1 CPU core returned a topology with %d cores.", cores)
			}

			if threads := topo.Threads(); threads != 1 {
				t.Fatalf("Creating a VM with 1 CPU thread returned a topology with %d threads.", threads)
			}

			if sockets := topo.Sockets(); sockets != 1 {
				t.Fatalf("Creating a VM with 1 CPU socket returned a topology with %d sockets.", sockets)
			}

		})
	}
}

func TestVMCreationWithMemory(t *testing.T) {
	t.Parallel()
	testMemory := int64(1048576)
	helper := getHelper(t)
	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams().MustWithMemory(testMemory),
	)

	memory := vm.Memory()
	if memory != testMemory {
		t.Fatalf("Creating a VM with Memory settings did not return a VM with expected Memory , %d , %d.", memory, testMemory)
	}

}

func TestVMCreationWithDefaultMemory(t *testing.T) {
	t.Parallel()
	defaultMemory := int64(1073741824)
	helper := getHelper(t)
	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams(),
	)

	memory := vm.Memory()
	if memory != defaultMemory {
		t.Fatalf("Creating a VM with a Default Memory settings did not return a VM with expected Memory , %d , %d.", memory, defaultMemory)
	}

}

func TestVMCreationFromTemplateChangedCPUValues(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)
	vm1 := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams().MustWithCPUParameters(2, 2, 2),
	)
	tpl := assertCanCreateTemplate(t, helper, vm1)
	vm2 := assertCanCreateVMFromTemplate(
		t,
		helper,
		"test2",
		tpl.ID(),
		ovirtclient.CreateVMParams().MustWithCPUParameters(3, 3, 3),
	)
	if vm2.CPU().Topo().Cores() != 3 {
		t.Fatalf("Invalid number of cores: %d", vm2.CPU().Topo().Cores())
	}
	if vm2.CPU().Topo().Threads() != 3 {
		t.Fatalf("Invalid number of cores: %d", vm2.CPU().Topo().Threads())
	}
	if vm2.CPU().Topo().Sockets() != 3 {
		t.Fatalf("Invalid number of cores: %d", vm2.CPU().Topo().Sockets())
	}
}

func assertNIC(t *testing.T, nic, shouldBe ovirtclient.NicConfiguration) {
	if nic == nil {
		t.Fatalf(
			"got Unexpected output from the NicConfiguration (%s) init field ",
			nic,
		)
	}

	if nic.Name() != shouldBe.Name() {
		t.Fatalf(
			"got Unexpected output from the NicConfiguration.Name (%s) init field ",
			nic.Name(),
		)
	}

	if nic.IP().Address != shouldBe.IP().Address {
		t.Fatalf(
			"got Unexpected output from the NicConfiguration.IP.Address (%s) init field ",
			nic.IP().Address,
		)
	}
	if nic.IP().Gateway != shouldBe.IP().Gateway {
		t.Fatalf(
			"got Unexpected output from the NicConfiguration.IP.Gateway (%s) init field ",
			nic.IP().Gateway,
		)
	}
	if nic.IP().Netmask != shouldBe.IP().Netmask {
		t.Fatalf(
			"got Unexpected output from the NicConfiguration.IP.Netmask (%s) init field ",
			nic.IP().Netmask,
		)
	}

	if shouldBe.IPV6() != nil {
		if nic.IPV6().Address != shouldBe.IPV6().Address {
			t.Fatalf(
				"got Unexpected output from the NicConfiguration.IPV6.Address (%s) init field ",
				nic.IPV6().Address,
			)
		}
		if nic.IPV6().Gateway != shouldBe.IPV6().Gateway {
			t.Fatalf(
				"got Unexpected output from the NicConfiguration.IPV6.Gateway (%s) init field ",
				nic.IPV6().Gateway,
			)
		}
		if nic.IPV6().Netmask != shouldBe.IPV6().Netmask {
			t.Fatalf(
				"got Unexpected output from the NicConfiguration.IPV6.Netmask (%s) init field ",
				nic.IPV6().Netmask,
			)
		}

	}
}

func TestVMCreationWithInit(t *testing.T) { //nolint:funlen,gocognit,gocyclo
	t.Parallel()
	helper := getHelper(t)

	testCases := []struct {
		name         string
		customScript string
		hostName     string
		optional     map[string]interface{}
	}{
		{
			"only custom script and hostname",
			"customScript",
			"hostName",
			nil,
		},
		{
			"all string and boolean parameters",
			"customScript",
			"hostName",
			map[string]interface{}{
				"activeDirectoryOu": "ActiveDirectoryOu",
				"authorizedSshKeys": "AuthorizedSshKeys",
				"dnsSearch":         "DnsSearch",
				"dnsServers":        "DnsServers",
				"domain":            "Domain",
				"inputLocale":       "InputLocale",
				"orgName":           "OrgName",
				"regenerateIds":     false,
				"regenerateSshKeys": false,
				"rootPassword":      "rootpass",
				"systemLocale":      "SystemLocale",
				"timezone":          "Timezone",
				"uiLanguage":        "UiLanguage",
				"userLocale":        "UserLocale",
				"userName":          "UserName",
				"windowsLicenseKey": "WindowsLicenseKey",
			},
		},
		{
			"with ipv4 nic configuration",
			"customScript",
			"hostName",
			map[string]interface{}{
				"nicConfiguration": ovirtclient.NewNicConfiguration("custom-nic", ovirtclient.IP{
					Version: ovirtclient.IPVERSION_V4,
					Address: "192.168.178.15",
					Gateway: "192.168.19.1",
					Netmask: "255.255.255.0",
				})},
		},

		{
			"with ipv6 nic configuration",
			"customScript",
			"hostName",
			map[string]interface{}{
				"nicConfiguration": ovirtclient.NewNicConfiguration("custom-nic", ovirtclient.IP{
					Version: ovirtclient.IPVERSION_V4,
					Address: "192.168.178.15",
					Gateway: "192.168.19.1",
					Netmask: "255.255.255.0",
				}).WithIPV6(ovirtclient.IP{
					Version: ovirtclient.IPVERSION_V6,
					Address: "fe80::bfb6:1c6c:f541:1aa564",
					Gateway: "fe80::",
					Netmask: "64",
				})},
		},
	}

	vm1 := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams(),
	)
	tpl := assertCanCreateTemplate(t, helper, vm1)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			customScript := testCase.customScript
			hostName := testCase.hostName

			init := ovirtclient.NewInitialization(customScript, hostName)

			nicConfiguration, nicConfigurationExists := testCase.optional["nicConfiguration"]
			if nicConfigurationExists {
				init = init.WithNicConfiguration(nicConfiguration.(ovirtclient.NicConfiguration))

			}

			activeDirectoryOu, activeDirectoryOuExists := testCase.optional["activeDirectoryOu"]
			if activeDirectoryOuExists {
				init = init.WithActiveDirectoryOu(activeDirectoryOu.(string))
			}

			authorizedSshKeys, authorizedSshKeysExists := testCase.optional["authorizedSshKeys"]
			if authorizedSshKeysExists {
				init = init.WithAuthorizedSshKeys(authorizedSshKeys.(string))
			}

			dnsSearch, dnsSearchExists := testCase.optional["dnsSearch"]
			if dnsSearchExists {
				init = init.WithDnsSearch(dnsSearch.(string))
			}

			dnsServers, dnsServersExists := testCase.optional["dnsServers"]
			if dnsServersExists {
				init = init.WithDnsServers(dnsServers.(string))
			}

			domain, domainExists := testCase.optional["domain"]
			if domainExists {
				init = init.WithDomain(domain.(string))
			}

			inputLocale, inputLocaleExists := testCase.optional["inputLocale"]
			if inputLocaleExists {
				init = init.WithInputLocale(inputLocale.(string))
			}

			orgName, orgNameExists := testCase.optional["orgName"]
			if orgNameExists {
				init = init.WithOrgName(orgName.(string))
			}

			regenerateIds, regenerateIdsExists := testCase.optional["regenerateIds"]
			if regenerateIdsExists {
				init = init.WithRegenerateIds(regenerateIds.(bool))
			}

			regenerateSshKeys, regenerateSshKeysExists := testCase.optional["regenerateSshKeys"]
			if regenerateSshKeysExists {
				init = init.WithRegenerateSshKeys(regenerateSshKeys.(bool))
			}

			rootPassword, rootPasswordExists := testCase.optional["rootPassword"]
			if rootPasswordExists {
				init = init.WithRootPassword(rootPassword.(string))
			}

			systemLocale, systemLocaleExists := testCase.optional["systemLocale"]
			if systemLocaleExists {
				init = init.WithSystemLocale(systemLocale.(string))
			}

			timezone, timezoneExists := testCase.optional["timezone"]
			if timezoneExists {
				init = init.WithTimezone(timezone.(string))
			}

			uiLanguage, uiLanguageExists := testCase.optional["uiLanguage"]
			if uiLanguageExists {
				init = init.WithUiLanguage(uiLanguage.(string))
			}

			userLocale, userLocaleExists := testCase.optional["userLocale"]
			if userLocaleExists {
				init = init.WithUserLocale(userLocale.(string))
			}

			userName, userNameExists := testCase.optional["userName"]
			if userNameExists {
				init = init.WithUserName(userName.(string))
			}

			windowsLicenseKey, windowsLicenseKeyExists := testCase.optional["windowsLicenseKey"]
			if windowsLicenseKeyExists {
				init = init.WithWindowsLicenseKey(windowsLicenseKey.(string))
			}

			vm2 := assertCanCreateVMFromTemplate(
				t,
				helper,
				fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
				tpl.ID(),
				ovirtclient.CreateVMParams().MustWithInitialization(init),
			)

			// vm2.Initialization().RegenerateIds() always nil in response, so skip this parameter

			if vm2.Initialization().CustomScript() != testCase.customScript {
				t.Fatalf("got Unexpected output from the CustomScript (%s) init field ", vm2.Initialization().CustomScript())
			}

			if vm2.Initialization().HostName() != testCase.hostName {
				t.Fatalf("got Unexpected output from the HostName (%s) init field ", vm2.Initialization().HostName())
			}

			if nicConfigurationExists {
				nic := vm2.Initialization().NicConfiguration()
				assertNIC(t, nic, nicConfiguration.(ovirtclient.NicConfiguration))
			}

			if activeDirectoryOuExists && vm2.Initialization().ActiveDirectoryOu() != activeDirectoryOu {
				t.Fatalf("got Unexpected output from the ActiveDirectoryOu (%s) init field ", vm2.Initialization().ActiveDirectoryOu())
			}

			if authorizedSshKeysExists && vm2.Initialization().AuthorizedSshKeys() != authorizedSshKeys {
				t.Fatalf("got Unexpected output from the AuthorizedSshKeys (%s) init field ", vm2.Initialization().AuthorizedSshKeys())
			}

			if dnsSearchExists && vm2.Initialization().DnsSearch() != dnsSearch {
				t.Fatalf("got Unexpected output from the DnsSearch (%s) init field ", vm2.Initialization().DnsSearch())
			}

			if dnsServersExists && vm2.Initialization().DnsServers() != dnsServers {
				t.Fatalf("got Unexpected output from the DnsServers (%s) init field ", vm2.Initialization().DnsServers())
			}

			if domainExists && vm2.Initialization().Domain() != domain {
				t.Fatalf("got Unexpected output from the Domain (%s) init field ", vm2.Initialization().Domain())
			}

			if inputLocaleExists && vm2.Initialization().InputLocale() != inputLocale {
				t.Fatalf("got Unexpected output from the InputLocale (%s) init field ", vm2.Initialization().InputLocale())
			}

			if orgNameExists && vm2.Initialization().OrgName() != orgName {
				t.Fatalf("got Unexpected output from the OrgName (%s) init field ", vm2.Initialization().OrgName())
			}

			if regenerateSshKeysExists && *vm2.Initialization().RegenerateSshKeys() != regenerateSshKeys {
				t.Fatalf("got Unexpected output from the RegenerateSshKeys (%t)", *vm2.Initialization().RegenerateSshKeys())
			}

			if rootPasswordExists && vm2.Initialization().RootPassword() != "******" {
				t.Fatalf("got Unexpected output from the RootPassword (%s) init field ", vm2.Initialization().RootPassword())
			}

			if systemLocaleExists && vm2.Initialization().SystemLocale() != systemLocale {
				t.Fatalf("got Unexpected output from the SystemLocale (%s) init field ", vm2.Initialization().SystemLocale())
			}

			if timezoneExists && vm2.Initialization().Timezone() != timezone {
				t.Fatalf("got Unexpected output from the Timezone (%s) init field ", vm2.Initialization().Timezone())
			}

			if uiLanguageExists && vm2.Initialization().UiLanguage() != uiLanguage {
				t.Fatalf("got Unexpected output from the UiLanguage (%s) init field ", vm2.Initialization().UiLanguage())
			}

			if userLocaleExists && vm2.Initialization().UserLocale() != userLocale {
				t.Fatalf("got Unexpected output from the UserLocale (%s) init field ", vm2.Initialization().UserLocale())
			}

			if userNameExists && vm2.Initialization().UserName() != userName {
				t.Fatalf("got Unexpected output from the UserName (%s) init field ", vm2.Initialization().UserName())
			}

			if windowsLicenseKeyExists && vm2.Initialization().WindowsLicenseKey() != windowsLicenseKey {
				t.Fatalf("got Unexpected output from the WindowsLicenseKey (%s) init field ", vm2.Initialization().WindowsLicenseKey())
			}

		},
		)
	}

}

func TestVMCreationWithDescription(t *testing.T) {
	t.Parallel()
	testDescription := "test description"
	helper := getHelper(t)
	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams().MustWithDescription(testDescription),
	)

	description := vm.Description()
	if description != testDescription {
		t.Fatalf("Creating a VM with Description settings did not return a VM with expected Description , %s , %s.", description, testDescription)
	}

}

// TestVMStartStop creates a micro VM with a tiny operating system, starts it and then stops it. The OS doesn't support
// ACPI, so shutdown cannot be tested.
func TestVMStartStop(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
		nil,
	)
	disk := assertCanCreateDisk(t, helper)
	assertCanAttachDiskWithParams(t, vm, disk, ovirtclient.CreateDiskAttachmentParams().MustWithBootable(true).MustWithActive(true))
	assertCanUploadDiskImage(t, helper, disk)
	assertCanStartVM(t, helper, vm)
	assertVMWillStart(t, vm)
	assertCanStopVM(t, vm)
	assertVMWillStop(t, vm)
}

func TestVMHugePages(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("test-%s", helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams().MustWithHugePages(ovirtclient.VMHugePages2M),
	)
	if hugePages := vm.HugePages(); hugePages == nil || *hugePages != ovirtclient.VMHugePages2M {
		if hugePages == nil {
			t.Fatalf("Hugepages not set on VM")
		} else {
			t.Fatalf("Incorrect value for Hugepages: %d", hugePages)
		}
	}
}

func TestCanRemoveTemplateIfVMIsCloned(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	templateVM := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		nil,
	)
	template := assertCanCreateTemplate(t, helper, templateVM)
	vm := assertCanCreateVMFromTemplate(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		template.ID(),
		ovirtclient.CreateVMParams().MustWithClone(true),
	)
	if vm.TemplateID() != ovirtclient.DefaultBlankTemplateID {
		t.Fatalf("Template ID is not set correctly on cloned VM (%s vs %s)", vm.TemplateID(), template.ID())
	}
	assertCanRemoveTemplate(t, helper, template.ID())
}

func TestCannotRemoveTemplateIfVMIsNotCloned(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	templateVM := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		nil,
	)
	template := assertCanCreateTemplate(t, helper, templateVM)
	vm := assertCanCreateVMFromTemplate(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		template.ID(),
		ovirtclient.CreateVMParams().MustWithClone(false),
	)
	if vm.TemplateID() != template.ID() {
		t.Fatalf("Template ID is not set correctly on non-cloned VM (%s vs %s)", vm.TemplateID(), template.ID())
	}
	assertCannotRemoveTemplate(t, helper, template.ID())
}

func TestCannotRemoveTemplateIfVMCloneIsNotSet(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	templateVM := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		nil,
	)
	template := assertCanCreateTemplate(t, helper, templateVM)
	vm := assertCanCreateVMFromTemplate(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		template.ID(),
		nil,
	)
	if vm.TemplateID() != template.ID() {
		t.Fatalf("Template ID is not set correctly on VM if cloning is not specified (%s vs %s)", vm.TemplateID(), template.ID())
	}
	assertCannotRemoveTemplate(t, helper, template.ID())
}

func TestVMCreationWithInstanceTypeID(t *testing.T) {
	t.Parallel()
	helper := getHelper(t)

	instanceTypes, err := helper.GetClient().ListInstanceTypes()
	if err != nil {
		t.Fatalf("Failed to list instance types (%v)", err)
	}
	if len(instanceTypes) == 0 {
		t.Fatalf("No instance types defined in engine.")
	}
	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		ovirtclient.NewCreateVMParams().MustWithInstanceTypeID(instanceTypes[0].ID()),
	)
	if vm.InstanceTypeID() == nil {
		t.Fatalf("VM %s has empty instance type ID.", vm.ID())
	}
	if *vm.InstanceTypeID() != instanceTypes[0].ID() {
		t.Fatalf("Incorrect instance type ID returned (expected: %s, got: %s)", instanceTypes[0].ID(), *vm.InstanceTypeID())
	}
}

func TestVMType(t *testing.T) {
	helper := getHelper(t)
	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		ovirtclient.NewCreateVMParams(),
	)
	if vm.VMType() != ovirtclient.VMTypeServer {
		t.Fatalf("Incorrect default VM type: %s", vm.VMType())
	}

	for _, vmType := range ovirtclient.VMTypeValues() {
		vm := assertCanCreateVM(
			t,
			helper,
			fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
			ovirtclient.NewCreateVMParams().MustWithVMType(vmType),
		)
		if vm.VMType() != vmType {
			t.Fatalf("Incorrect VM type (expected: %s, got: %s)", vmType, vm.VMType())
		}
	}
}

func assertCanCreateVM(
	t *testing.T,
	helper ovirtclient.TestHelper,
	name string,
	params ovirtclient.OptionalVMParameters,
) ovirtclient.VM {
	return assertCanCreateVMFromTemplate(t, helper, name, helper.GetBlankTemplateID(), params)
}

func TestVMWithoutInitialization(t *testing.T) {
	helper := getHelper(t)

	vm := assertCanCreateVM(t, helper, fmt.Sprintf("test_%s", helper.GenerateRandomID(5)), nil)

	vm, err := helper.GetClient().GetVM(vm.ID())
	if err != nil {
		t.Fatalf("Failed to re-fetch VM after creation (%v)", err)
	}

	if vm.Initialization() == nil {
		t.Fatalf("Initialization is nil on VM without initialization.")
	}
}

func TestVMCreationWithSparseDisks(t *testing.T) {
	helper := getHelper(t)
	disk := assertCanCreateDiskWithParameters(
		t,
		helper,
		ovirtclient.ImageFormatRaw,
		ovirtclient.
			CreateDiskParams().
			MustWithSparse(false),
	)
	assertCanUploadDiskImage(t, helper, disk)
	startVM := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		nil,
	)
	assertCanAttachDisk(t, startVM, disk)

	checkVMDiskSparseness(
		t,
		startVM,
		false,
		"Start VM disk is sparse despite being created non-sparse.",
	)

	tpl := assertCanCreateTemplate(t, helper, startVM)
	diskAttachments, err := tpl.ListDiskAttachments()
	if err != nil {
		t.Fatalf("Failed to list disk attachments for template %s (%v).", tpl.ID(), err)
	}
	templateDisk := diskAttachments[0]
	var diskList = []ovirtclient.OptionalVMDiskParameters{
		// We must create a sparse / cow config here, otherwise the disk creation can horribly fail on NFS SDs and lock
		// up the disk.
		ovirtclient.
			MustNewBuildableVMDiskParameters(templateDisk.DiskID()).
			MustWithSparse(true),
	}
	vm := assertCanCreateVMFromTemplate(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		tpl.ID(),
		ovirtclient.CreateVMParams().MustWithDisks(diskList).MustWithClone(true),
	)

	checkVMDiskSparseness(
		t,
		vm,
		true,
		"VM from template disk is non-sparse despite being created with a sparse override.",
	)
}

func TestMemoryPolicyDefaults(t *testing.T) {
	helper := getHelper(t)
	vm := assertCanCreateVM(
		t,
		helper,
		helper.GenerateTestResourceName(t),
		nil,
	)
	// Test if memory policy is correctly set
	memoryPolicy := vm.MemoryPolicy()
	if memoryPolicy == nil {
		t.Fatalf("Memory policy was not set.")
	}

}

func TestGuaranteedMemory(t *testing.T) {
	helper := getHelper(t)
	expectedGuaranteed := int64(2 * 1024 * 1024 * 1024)
	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		ovirtclient.
			CreateVMParams().
			WithMemoryPolicy(
				ovirtclient.
					NewMemoryPolicyParameters().
					MustWithGuaranteed(expectedGuaranteed),
			).MustWithMemory(expectedGuaranteed),
	)
	memoryPolicy := vm.MemoryPolicy()
	guaranteed := memoryPolicy.Guaranteed()
	if guaranteed == nil {
		t.Fatalf("Guaranteed memory is not set on VM.")
	}
	if *guaranteed != expectedGuaranteed {
		t.Fatalf("Incorrect guaranteed memory value (expected: %d, got: %d)", expectedGuaranteed, *guaranteed)
	}
}

func TestMaxMemory(t *testing.T) {
	helper := getHelper(t)
	expectedMax := int64(2 * 1024 * 1024 * 1024)
	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		ovirtclient.
			NewCreateVMParams().
			WithMemoryPolicy(
				ovirtclient.
					NewMemoryPolicyParameters().
					MustWithMax(expectedMax),
			).MustWithMemory(expectedMax),
	)
	memoryPolicy := vm.MemoryPolicy()
	max := memoryPolicy.Max()
	if max == nil {
		t.Fatalf("Guaranteed memory is not set on VM.")
	}
	if *max != expectedMax {
		t.Fatalf("Incorrect max memory value (expected: %d, got: %d)", expectedMax, *max)
	}
}

func TestBallooning(t *testing.T) {
	truePointer := true
	falsePointer := false
	testCases := []struct {
		name     string
		set      *bool
		expected bool
	}{
		{
			"empty",
			nil,
			true,
		},
		{
			"true",
			&truePointer,
			true,
		},
		{
			"false",
			&falsePointer,
			false,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("ballooning=%s", testCase.name), func(t *testing.T) {
			helper := getHelper(t)
			params := ovirtclient.NewCreateVMParams()
			if testCase.set != nil {
				params = params.WithMemoryPolicy(
					ovirtclient.
						NewMemoryPolicyParameters().
						MustWithBallooning(*testCase.set),
				)
			}
			vm := assertCanCreateVM(
				t,
				helper,
				helper.GenerateTestResourceName(t),
				params,
			)
			memoryPolicy := vm.MemoryPolicy()
			ballooning := memoryPolicy.Ballooning()
			if ballooning != testCase.expected {
				t.Fatalf("Incorrect ballooning value")
			}
		})
	}

}

func checkVMDiskSparseness(t *testing.T, checkVM ovirtclient.VM, sparse bool, message string) {
	t.Helper()
	diskAttachments, err := checkVM.ListDiskAttachments()
	if err != nil {
		t.Fatalf("Failed to list disk attachments for VM %s (%v).", checkVM.ID(), err)
	}
	if len(diskAttachments) != 1 {
		t.Fatalf("Incorrect number of disk attachments on VM %s (%d).", checkVM.ID(), len(diskAttachments))
	}
	d, err := diskAttachments[0].Disk()
	if err != nil {
		t.Fatalf("Failed to fetch disk for VM %s (%v).", checkVM.ID(), err)
	}
	if d.Sparse() != sparse {
		t.Fatalf(message)
	}
}

func TestPlacementPolicy(t *testing.T) {
	helper := getHelper(t)

	hosts, err := helper.GetClient().ListHosts()
	if err != nil {
		t.Fatalf("Failed to list hosts (%v).", err)
	}
	if len(hosts) == 0 {
		t.Fatalf("No hosts found in Engine!")
	}
	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		ovirtclient.CreateVMParams().WithPlacementPolicy(
			ovirtclient.
				NewVMPlacementPolicyParameters().
				MustWithAffinity(ovirtclient.VMAffinityPinned).
				MustWithHostIDs([]ovirtclient.HostID{hosts[0].ID()}),
		),
	)
	pp, ok := vm.PlacementPolicy()
	if !ok {
		t.Fatalf("No placement policy returned after creating a VM with a set placement policy.")
	}
	if pp.Affinity() == nil {
		t.Fatalf("Placement policy has no affinity even though it was set on VM creation.")
	}
	if affinity := *pp.Affinity(); affinity != ovirtclient.VMAffinityPinned {
		t.Fatalf(
			"Incorrect affinity on placement policy (expected: %s, got: %s)",
			ovirtclient.VMAffinityPinned,
			affinity,
		)
	}
	hostIDs := pp.HostIDs()
	if len(hostIDs) != 1 {
		t.Fatalf("Incorrect number of host IDs in host list (expected: %d, got: %d)", len(hostIDs), 1)
	}
	if hostIDs[0] != hosts[0].ID() {
		t.Fatalf("Incorrect host ID in host list (expected: %s, got: %s)", hosts[0].ID(), hostIDs[0])
	}
}

func TestVMOSParameter(t *testing.T) {
	helper := getHelper(t)

	vm := assertCanCreateVM(
		t,
		helper,
		fmt.Sprintf("%s-%s", t.Name(), helper.GenerateRandomID(5)),
		ovirtclient.NewCreateVMParams().WithOS(ovirtclient.NewVMOSParameters().MustWithType("rhcos_x64")),
	)
	os := vm.OS()
	if os.Type() != "rhcos_x64" {
		t.Fatalf("Incorrect OS type (expected: %s, got: %s)", "rhcos_x64", os.Type())
	}
}

func TestVMCPUMode(t *testing.T) {
	helper := getHelper(t)

	vm := assertCanCreateVM(
		t,
		helper,
		helper.GenerateTestResourceName(t),
		nil,
	)
	if vm.CPU().Mode() != nil {
		t.Fatalf("Incorrect CPU mode: %s", *vm.CPU().Mode())
	}

	m := ovirtclient.CPUModeHostPassthrough
	vm = assertCanCreateVM(
		t,
		helper,
		helper.GenerateTestResourceName(t),
		ovirtclient.NewCreateVMParams().MustWithCPU(ovirtclient.NewVMCPUParams().MustWithMode(m)),
	)
	if vm.CPU().Mode() == nil {
		t.Fatalf("CPU mode is nil, expected %s.", m)
	}
	if *vm.CPU().Mode() != m {
		t.Fatalf("Incorrect CPU mode (expected: %s, got: %s)", m, *vm.CPU().Mode())
	}
}

func TestVMDiskStorageDomain(t *testing.T) {
	helper := getHelper(t)

	storageDomain2 := helper.GetSecondaryStorageDomainID(t)
	disk := assertCanCreateDiskWithParameters(t, helper, ovirtclient.ImageFormatCow, nil)
	vm := assertCanCreateVM(t, helper, helper.GenerateTestResourceName(t), nil)
	assertCanAttachDisk(t, vm, disk)
	tpl := assertCanCreateTemplate(t, helper, vm)
	diskAttachments, err := tpl.ListDiskAttachments()
	if err != nil {
		t.Fatalf("Failed to list template %s diks attachments (%v)", tpl.ID(), err)
	}
	vm2, err := helper.GetClient().CreateVM(
		helper.GetClusterID(),
		tpl.ID(),
		helper.GenerateTestResourceName(t),
		ovirtclient.
			NewCreateVMParams().
			MustWithClone(true).
			MustWithDisks(
				[]ovirtclient.OptionalVMDiskParameters{
					ovirtclient.
						MustNewBuildableVMDiskParameters(diskAttachments[0].DiskID()).
						MustWithStorageDomainID(storageDomain2),
				},
			),
	)
	if err != nil {
		t.Fatalf("Failed to create VM from template %s (%v)", tpl.ID(), err)
	}
	t.Cleanup(func() {
		if err := vm2.Remove(); err != nil && !ovirtclient.HasErrorCode(err, ovirtclient.ENotFound) {
			t.Fatalf("Failed to clean up VM %s after test (%v)", vm2.ID(), err)
		}
	})

	vmDiskAttachments, err := vm2.ListDiskAttachments()
	if err != nil {
		t.Fatalf("Failed to list VM %s disk attachments (%v)", vm2.ID(), err)
	}
	disk2, err := vmDiskAttachments[0].Disk()
	if err != nil {
		t.Fatalf("Failed to fetch disk %s (%v)", vmDiskAttachments[0].DiskID(), err)
	}
	found := false
	for _, storageDomainID := range disk2.StorageDomainIDs() {
		if storageDomainID == storageDomain2 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Disk %s is not on the required storage domain %s.", disk2.ID(), storageDomain2)
	}
}

func TestVMSerialConsole(t *testing.T) {
	testCases := getSerialConsoleTestCases()
	nilString := "nil"

	for _, tc := range testCases {
		vmType := nilString
		if tc.vmType != nil {
			vmType = string(*tc.vmType)
		}
		set := nilString
		if tc.set != nil {
			set = fmt.Sprintf("%t", *tc.set)
		}
		t.Run(
			fmt.Sprintf("vmType=%s,console=%s", vmType, set),
			func(t *testing.T) {
				var err error
				helper := getHelper(t)

				t.Logf(
					"Creating VM with vmType=%s and console=%s, expecting console to be %t...",
					vmType,
					set,
					tc.expected,
				)

				params := ovirtclient.NewCreateVMParams()
				if tc.vmType != nil {
					params, err = params.WithVMType(*tc.vmType)
					if err != nil {
						t.Fatalf("Failed to set VM type (%v)", err)
					}
				}
				if tc.set != nil {
					params = params.WithSerialConsole(*tc.set)
				}

				vm := assertCanCreateVM(
					t,
					helper,
					helper.GenerateTestResourceName(t),
					params,
				)
				if vm.SerialConsole() != tc.expected {
					t.Fatalf(
						"Incorrect value for serial console (expected: %t, got: %t)",
						tc.expected,
						vm.SerialConsole(),
					)
				}
				t.Logf("Found correct value for serial console: %t", tc.expected)
			},
		)
	}
}

func TestVMSoundcardEnabled(t *testing.T) {
	testCases := getSoundcardEnabledTestCases()
	nilString := "nil"

	for _, tc := range testCases {
		vmType := nilString
		if tc.vmType != nil {
			vmType = string(*tc.vmType)
		}
		set := nilString
		if tc.set != nil {
			set = fmt.Sprintf("%t", *tc.set)
		}
		t.Run(
			fmt.Sprintf("vmType=%s,enabled=%s", vmType, set),
			func(t *testing.T) {
				var err error
				helper := getHelper(t)

				params := ovirtclient.NewCreateVMParams()
				if tc.vmType != nil {
					params, err = params.WithVMType(*tc.vmType)
					if err != nil {
						t.Fatalf("Failed to set VM type (%v)", err)
					}
				}
				if tc.set != nil {
					params = params.WithSoundcardEnabled(*tc.set)
				}

				vm := assertCanCreateVM(
					t,
					helper,
					helper.GenerateTestResourceName(t),
					params,
				)
				if vm.SoundcardEnabled() != tc.expected {
					t.Fatalf(
						"Incorrect value for soundcard (expected: %t, got: %t)",
						tc.expected,
						vm.SoundcardEnabled(),
					)
				}
			},
		)
	}
}

func getSerialConsoleTestCases() []struct {
	vmType   *ovirtclient.VMType
	set      *bool
	expected bool
} {
	yes := true
	no := false
	desktop := ovirtclient.VMTypeDesktop
	server := ovirtclient.VMTypeServer
	hp := ovirtclient.VMTypeHighPerformance
	testCases := []struct {
		vmType   *ovirtclient.VMType
		set      *bool
		expected bool
	}{
		{nil, &yes, true},
		{nil, &no, false},
		{&desktop, nil, false},
		{&server, nil, false},
		{&hp, nil, false},
		{&desktop, &yes, true},
		{&server, &yes, true},
		{&hp, &yes, true},
		{&desktop, &no, false},
		{&server, &no, false},
		{&hp, &no, false},
	}
	return testCases
}

func getSoundcardEnabledTestCases() []struct {
	vmType   *ovirtclient.VMType
	set      *bool
	expected bool
} {
	yes := true
	no := false
	desktop := ovirtclient.VMTypeDesktop
	server := ovirtclient.VMTypeServer
	hp := ovirtclient.VMTypeHighPerformance
	testCases := []struct {
		vmType   *ovirtclient.VMType
		set      *bool
		expected bool
	}{
		{nil, &yes, true},
		{nil, &no, false},
		{&desktop, nil, true},
		{&server, nil, true},
		{&hp, nil, true},
		{&desktop, &yes, true},
		{&server, &yes, true},
		{&hp, &yes, true},
		{&desktop, &no, false},
		{&server, &no, false},
		{&hp, &no, false},
	}
	return testCases
}

func assertCanCreateVMFromTemplate(
	t *testing.T,
	helper ovirtclient.TestHelper,
	name string,
	templateID ovirtclient.TemplateID,
	params ovirtclient.OptionalVMParameters,
) ovirtclient.VM {
	t.Logf("Creating VM %s from template %s...", name, templateID)
	client := helper.GetClient()
	vm, err := client.CreateVM(
		helper.GetClusterID(),
		templateID,
		name,
		params,
	)
	if err != nil {
		t.Fatalf("Failed to create test VM (%v)", err)
	}
	t.Cleanup(
		func() {
			if vm != nil {
				t.Logf("Cleaning up test VM %s...", vm.ID())
				if err := vm.Remove(); err != nil && !ovirtclient.HasErrorCode(err, ovirtclient.ENotFound) {
					t.Fatalf("Failed to remove test VM %s (%v)", vm.ID(), err)
				}
			}
		},
	)
	t.Logf("Created VM %s.", vm.ID())
	return vm
}

func assertCanStartVM(t *testing.T, helper ovirtclient.TestHelper, vm ovirtclient.VM) {
	if err := vm.Start(); err != nil {
		t.Fatalf("Failed to start VM (%v)", err)
	}
	t.Cleanup(func() {
		vmID := vm.ID()
		vm, err := helper.GetClient().GetVM(vmID)
		if err != nil {
			if !ovirtclient.HasErrorCode(err, ovirtclient.ENotFound) {
				t.Fatalf("Failed to update VM %s status (%v)", vmID, err)
			} else {
				return
			}
		}
		if vm.Status() == ovirtclient.VMStatusDown {
			return
		}
		t.Logf("Stopping test VM %s...", vm.ID())
		if err := vm.Stop(true); err != nil {
			t.Fatalf("Failed to stop VM %s after test (%v)", vm.ID(), err)
		}
		if _, err := vm.WaitForStatus(ovirtclient.VMStatusDown); err != nil {
			t.Fatalf("Failed to wait for VM %s to stop (%v)", vm.ID(), err)
		}
	})
}

func assertVMWillStart(t *testing.T, vm ovirtclient.VM) ovirtclient.VM {
	vm, err := vm.WaitForStatus(ovirtclient.VMStatusUp)
	if err != nil {
		t.Fatalf("Failed to wait for VM status to reach \"up\". (%v)", err)
	}
	return vm
}

func assertCanStopVM(t *testing.T, vm ovirtclient.VM) {
	if err := vm.Stop(false); err != nil {
		t.Fatalf("Failed to stop VM (%v)", err)
	}
}

func assertVMWillStop(t *testing.T, vm ovirtclient.VM) {
	if _, err := vm.WaitForStatus(ovirtclient.VMStatusDown); err != nil {
		t.Fatalf("Failed to wait for VM status to reach \"down\". (%v)", err)
	}
}

func assertCanCreateBootableVM(t *testing.T, helper ovirtclient.TestHelper) ovirtclient.VM {
	vm1 := assertCanCreateVM(
		t,
		helper,
		helper.GenerateRandomID(5),
		nil,
	)
	disk1 := assertCanCreateDisk(t, helper)
	assertCanUploadDiskImage(t, helper, disk1)
	assertCanAttachDiskWithParams(t, vm1, disk1, ovirtclient.CreateDiskAttachmentParams().MustWithBootable(true).MustWithActive(true))
	return vm1
}
