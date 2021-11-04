package ovirtclient_test

import (
	"fmt"
	"reflect"
	"testing"

	ovirtclient "github.com/ovirt/go-ovirt-client"
)

const DummyString = "test"
const DummyMemoryMib = 1024

func TestVMListShouldNotFail(t *testing.T) {
	helper := getHelper(t)
	client := helper.GetClient()

	if _, err := client.ListVMs(); err != nil {
		t.Fatal(err)
	}
}

func TestVMCreation(t *testing.T) {
	type testCase struct {
		name                string
		shouldSucceedCreate bool
		createVMParams      ovirtclient.BuildableVMParameters
		expected            func() map[string]interface{}
		assertFunction      func(vm ovirtclient.VM, getters map[string]interface{}) error
	}

	helper := getHelper(t)
	client := helper.GetClient()

	newBaseCreateParams := func() ovirtclient.BuildableVMParameters {
		return ovirtclient.CreateVMParams().
			MustWithComment(DummyString).
			MustWithVMType(ovirtclient.VMTypeDesktop).
			MustWithCPU(*ovirtclient.NewCPU(1, 8, 1)).
			MustWithMemoryMB(DummyMemoryMib)
	}
	newBaseFieldToExpectedValue := func() map[string]interface{} {
		return map[string]interface{}{
			"comment":  DummyString,
			"vmType":   ovirtclient.VMTypeDesktop,
			"cpu":      *ovirtclient.NewCPU(1, 8, 1),
			"memoryMB": DummyMemoryMib,
		}
	}
	testCases := []testCase{
		{
			name:                "Simple VM",
			shouldSucceedCreate: true,
			createVMParams:      newBaseCreateParams(),
			expected:            newBaseFieldToExpectedValue,
			assertFunction:      baseVMAssertFunction,
		},
		{
			name:                "VM with server type",
			shouldSucceedCreate: true,
			createVMParams:      newBaseCreateParams().MustWithVMType(ovirtclient.VMTypeServer),
			expected: func() map[string]interface{} {
				expected := newBaseFieldToExpectedValue()
				expected["vmType"] = ovirtclient.VMTypeServer
				return expected
			},
			assertFunction: baseVMAssertFunction,
		},
		{
			name:                "VM with high performance type",
			shouldSucceedCreate: true,
			createVMParams:      newBaseCreateParams().MustWithVMType(ovirtclient.VMTypeHighPerformance),
			expected: func() map[string]interface{} {
				expected := newBaseFieldToExpectedValue()
				expected["vmType"] = ovirtclient.VMTypeHighPerformance
				return expected
			},
			assertFunction: baseVMAssertFunction,
		},
		{
			name:                "VM with valid small hugepages",
			shouldSucceedCreate: true,
			createVMParams:      newBaseCreateParams().MustWithHugepages(ovirtclient.VMHugepagesSmall),
			expected: func() map[string]interface{} {
				expected := newBaseFieldToExpectedValue()
				expected["hugepages"] = ovirtclient.VMHugepagesSmall
				return expected
			},
			assertFunction: baseVMAssertFunction,
		},
		{
			name:                "VM with valid large hugepages",
			shouldSucceedCreate: true,
			createVMParams:      newBaseCreateParams().MustWithHugepages(ovirtclient.VMHugepagesLarge),
			expected: func() map[string]interface{} {
				expected := newBaseFieldToExpectedValue()
				expected["hugepages"] = ovirtclient.VMHugepagesLarge
				return expected
			},
			assertFunction: baseVMAssertFunction,
		},
		{
			name:                "VM with valid GuaranteedMemory",
			shouldSucceedCreate: true,
			createVMParams:      newBaseCreateParams().MustWithGuaranteedMemoryMB(DummyMemoryMib),
			expected: func() map[string]interface{} {
				expected := newBaseFieldToExpectedValue()
				expected["guaranteedMemoryMB"] = DummyMemoryMib
				return expected
			},
			assertFunction: baseVMAssertFunction,
		},
		{
			name:                "VM with invalid GuaranteedMemory - GuaranteedMemory bigger than memory",
			shouldSucceedCreate: false,
			createVMParams:      newBaseCreateParams().MustWithGuaranteedMemoryMB(2048),
			expected:            nil,
			assertFunction:      baseVMAssertFunction,
		},
		// TODO: we need an ability to mock hosts for this test
		//{
		//	name:                "VM with valid AutoPinningPolicy",
		//	shouldSucceedCreate: false,
		//	createVMParams:      newBaseCreateParams().
		//		MustWithAutoPinningPolicy("test").
		//		MustWithPlacementPolicy()),
		//	expected: nil,
		//	assertFunction: baseVMAssertFunction,
		//},
		{
			name:                "VM with invalid AutoPinningPolicy - AutoPinningPolicy without placement policy",
			shouldSucceedCreate: false,
			createVMParams: newBaseCreateParams().
				MustWithAutoPinningPolicy(ovirtclient.VMAutoPinningPolicyResizeAndPin),
			expected:       nil,
			assertFunction: baseVMAssertFunction,
		},
		// TODO: we need an ability to mock hosts for this test
		//{
		//	name:                "VM with valid Placement Policy",
		//	shouldSucceedCreate: true,
		//	createVMParams:      newBaseCreateParams().MustWithPlacementPolicy(),
		//	expected: nil,
		//	assertFunction: baseVMAssertFunction,
		//},
		{
			name:                "VM with valid initialization - just hostname",
			shouldSucceedCreate: true,
			createVMParams: newBaseCreateParams().
				MustWithInitialization(*ovirtclient.NewInitialization().WithHostname(DummyString)),
			expected: func() map[string]interface{} {
				expected := newBaseFieldToExpectedValue()
				expected["initialization"] = *ovirtclient.NewInitialization().WithHostname(DummyString)
				return expected
			},
			assertFunction: baseVMAssertFunction,
		},
		// TODO: Need to a tag resource to actually create the tags before
		//{
		//	name:                "VM with valid tags",
		//	shouldSucceedCreate: true,
		//	createVMParams:      newBaseCreateParams().MustWithTags([]string{"test1", "test2"}),
		//	expected: func() map[string]interface{} {
		//		expected := newBaseFieldToExpectedValue()
		//		expected["tags"] = []string{"test1", "test2"}
		//		return expected
		//	},
		//	assertFunction: baseVMAssertFunction,
		//},
	}
	for _, tc := range testCases {
		t.Logf("starting test case: %s", tc.name)
		vm, err := client.CreateVM(
			helper.GetClusterID(),
			helper.GetBlankTemplateID(),
			fmt.Sprintf("%s_%s", DummyString, helper.GenerateRandomID(5)),
			tc.createVMParams)
		if !tc.shouldSucceedCreate {
			if err == nil {
				t.Fatalf("expected to fail creation but error wasn't returned")
			}
			// TODO: how to analyze the error? in case I expect the failure
		} else {
			if err != nil {
				t.Fatalf("expected to succeed creation but error occourd %v", err)
			}
			createdVm, err := client.GetVM(vm.ID())
			err = tc.assertFunction(createdVm, tc.expected())
			if err != nil {
				t.Fatalf("test case: %s failed assertion, err: %v", tc.name, err)
			}
			// TODO: Since we are running in a for loop I can't use defer to call it but
			// I also don't want to do it in each case, need to examin if there is a better way
			err = client.RemoveVM(vm.ID())
			if err != nil {
				t.Fatalf("failed to remove VM after test, please remove manually (%v)", err)
			}
		}
	}
}

func TestVMUpdate(t *testing.T) {
	helper := getHelper(t)
	client := helper.GetClient()

	vm, err := client.CreateVM(
		helper.GetClusterID(),
		helper.GetBlankTemplateID(),
		DummyString,
		ovirtclient.CreateVMParams(),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = client.RemoveVM(vm.ID())
		if err != nil {
			t.Fatalf("failed to remove VM after test, please remove manually (%v)", err)
		}
	}()
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
		ovirtclient.UpdateVMParams().MustWithName("new_name").MustWithComment("new comment"),
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
}

func assertCanCreateVM(
	t *testing.T,
	helper ovirtclient.TestHelper,
	name string,
	params ovirtclient.OptionalVMParameters,
) ovirtclient.VM {
	client := helper.GetClient()
	vm, err := client.CreateVM(
		helper.GetClusterID(),
		helper.GetBlankTemplateID(),
		DummyString,
		params,
	)
	if err != nil {
		t.Fatalf("Failed to create test VM (%v)", err)
	}
	t.Cleanup(
		func() {
			if err := vm.Remove(); err != nil {
				t.Fatalf("Failed to remove test VM %s (%v)", vm.ID(), err)
			}
		},
	)
	return vm
}

func baseVMAssertFunction(vm ovirtclient.VM, fieldToExpectedValue map[string]interface{}) error {
	msg := "unexpected value, field %s expected %v but got %v"
	for field, value := range fieldToExpectedValue {
		switch field {
		case "name":
			if vm.Name() != value {
				return fmt.Errorf(msg, field, value, vm.Name())
			}
		case "comment":
			if vm.Comment() != value {
				return fmt.Errorf(msg, field, value, vm.Comment())
			}
		case "cpu":
			if vm.CPU() != value {
				return fmt.Errorf(msg, field, value, vm.CPU())
			}
		case "memoryMB":
			if vm.MemoryMB() != uint64(value.(int)) {
				return fmt.Errorf(msg, field, value, vm.MemoryMB())
			}
		case "vmType":
			if vm.VMType() != value {
				return fmt.Errorf(msg, field, value, vm.VMType())
			}
		case "autoPiningPolicy":
			if vm.AutoPinningPolicy() != value {
				return fmt.Errorf(msg, field, value, vm.AutoPinningPolicy())
			}
		case "placementPolicy":
			if placementPolicy := vm.PlacementPolicy(); !reflect.DeepEqual(*placementPolicy, value) {
				return fmt.Errorf(msg, field, value, placementPolicy)
			}
		case "hugepages":
			if hugepages := vm.Hugepages(); *hugepages != value {
				return fmt.Errorf(msg, field, value, hugepages)
			}
		case "guaranteedMemoryMB":
			if guaranteedMemoryMB := vm.GuaranteedMemoryMB(); *guaranteedMemoryMB != uint64(value.(int)) {
				return fmt.Errorf(msg, field, value, guaranteedMemoryMB)
			}
		case "initialization":
			if initialization := vm.Initialization(); !reflect.DeepEqual(*initialization, value) {
				return fmt.Errorf(msg, field, value, initialization)
			}
		case "tags":
			if tags := vm.Tags(); !reflect.DeepEqual(tags, value) {
				return fmt.Errorf(msg, field, value, tags)
			}
		default:
			return fmt.Errorf("unrecognized field %s", field)
		}
	}
	return nil
}

// TODO: Maye this is to complex, and we should remove
// baseAssertFunction calls all the functions in getters on VM object vm,
// and compares each returned value with the expected value.
//
// gettersToExpectedValue is a map in which each key represents a getter method on the VM interface
// and each value represents the expected value which should be returned by the getter.
//
// returns an error if a getter method is not found on the VM object, or if an unexpected value is found
//func baseVMAssertFunction(vm ovirtclient.VM, gettersToExpectedValue map[string]interface{}) error {
//	reflectVM := reflect.ValueOf(&vm).Elem()
//	for getter, expectedValue := range gettersToExpectedValue {
//		method := reflectVM.MethodByName(getter)
//		if method.IsZero() {
//			return fmt.Errorf("method %s doesn't exist for type %s", getter, reflectVM.Type().Name())
//		}
//		value := method.Call([]reflect.Value{})[0]
//		if reflect.DeepEqual(value, expectedValue) {
//			return fmt.Errorf(
//				"field %s doesn't match, expected %v but got %v",
//				getter, expectedValue, value)
//		}
//	}
//	return nil
//}
