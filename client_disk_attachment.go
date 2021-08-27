package ovirtclient

import (
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

// DiskAttachmentClient contains the methods required for handling disk attachments.
type DiskAttachmentClient interface {
	// CreateDiskAttachment attaches a disk to a VM.
	CreateDiskAttachment(vmID string, diskID string, diskInterface DiskInterface, params CreateDiskAttachmentOptionalParams, retries ...RetryStrategy) (DiskAttachment, error)
	// GetDiskAttachment returns a single disk attachment in a virtual machine.
	GetDiskAttachment(vmID string, id string, retries ...RetryStrategy) (DiskAttachment, error)
	// ListDiskAttachments lists all disk attachments for a virtual machine.
	ListDiskAttachments(vmID string, retries ...RetryStrategy) ([]DiskAttachment, error)
	// RemoveDiskAttachment removes the disk attachment in question.
	RemoveDiskAttachment(vmID string, diskAttachmentID string, retries ...RetryStrategy) error
}

// DiskInterface describes the means by which a disk will appear to the VM.
type DiskInterface string

const (
	// DiskInterfaceIDE is a legacy controller device. Works with almost all guest operating systems, so it is good for
	// compatibility. Performance is lower than with the other alternatives.
	DiskInterfaceIDE DiskInterface = "ide"
	// DiskInterfaceSATA is a SATA controller device.
	DiskInterfaceSATA DiskInterface = "sata"
	// DiskInterfacesPAPRvSCSI is a para-virtualized device supported by the IBM pSeries family of machines, using the
	// SCSI protocol.
	DiskInterfacesPAPRvSCSI DiskInterface = "spapr_vscsi"
	// DiskInterfaceVirtIO is a virtualization interface where just the guest's device driver knows it is running in a
	// virtual environment. Enables guests to get high performance disk operations.
	DiskInterfaceVirtIO DiskInterface = "virtio"
	// DiskInterfaceVirtIOSCSI is a para-virtualized SCSI controller device. Fast interface with the guest via direct
	// physical storage device address, using the SCSI protocol.
	DiskInterfaceVirtIOSCSI DiskInterface = "virtio_scsi"
)

// Validate checks if the DiskInterface actually has a valid value.
func (d DiskInterface) Validate() error {
	switch d {
	case DiskInterfaceIDE:
		return nil
	case DiskInterfaceSATA:
		return nil
	case DiskInterfacesPAPRvSCSI:
		return nil
	case DiskInterfaceVirtIO:
		return nil
	case DiskInterfaceVirtIOSCSI:
		return nil
	default:
		return newError(EBadArgument, "invalid disk interface: %s", d)
	}
}

// CreateDiskAttachmentOptionalParams are the optional parameters for creating a disk attachment.
type CreateDiskAttachmentOptionalParams interface{}

// CreateDiskAttachmentBuildableParams is a buildable version of CreateDiskAttachmentOptionalParams.
type CreateDiskAttachmentBuildableParams interface{}

// CreateDiskAttachmentParams creates a buildable set of parameters for creating a disk attachment.
func CreateDiskAttachmentParams() CreateDiskAttachmentBuildableParams {
	return &createDiskAttachmentParams{}
}

type createDiskAttachmentParams struct{}

// DiskAttachment links together a Disk and a VM.
type DiskAttachment interface {
	// ID returns the identifier of the attachment.
	ID() string
	// VMID returns the ID of the virtual machine this attachment belongs to.
	VMID() string
	// DiskID returns the ID of the disk in this attachment.
	DiskID() string
	// DiskInterface describes the means by which a disk will appear to the VM.
	DiskInterface() DiskInterface

	// VM fetches the virtual machine this attachment belongs to.
	VM(retries ...RetryStrategy) (VM, error)
	// Disk fetches the disk this attachment attaches.
	Disk(retries ...RetryStrategy) (Disk, error)

	// Remove removes the current disk attachment.
	Remove(retries ...RetryStrategy) error
}

type diskAttachment struct {
	client Client

	id            string
	vmid          string
	diskID        string
	diskInterface DiskInterface
}

func (d *diskAttachment) DiskInterface() DiskInterface {
	return d.diskInterface
}

func (d *diskAttachment) Remove(retries ...RetryStrategy) error {
	return d.client.RemoveDiskAttachment(d.vmid, d.id, retries...)
}

func (d *diskAttachment) ID() string {
	return d.id
}

func (d *diskAttachment) VMID() string {
	return d.vmid
}

func (d *diskAttachment) DiskID() string {
	return d.diskID
}

func (d *diskAttachment) VM(retries ...RetryStrategy) (VM, error) {
	return d.client.GetVM(d.vmid, retries...)
}

func (d *diskAttachment) Disk(retries ...RetryStrategy) (Disk, error) {
	return d.client.GetDisk(d.diskID, retries...)
}

func convertSDKDiskAttachment(object *ovirtsdk4.DiskAttachment, o *oVirtClient) (DiskAttachment, error) {
	id, ok := object.Id()
	if !ok {
		return nil, newFieldNotFound("disk attachment", "id")
	}
	vm, ok := object.Vm()
	if !ok {
		return nil, newFieldNotFound("disk attachment", "vm")
	}
	vmID, ok := vm.Id()
	if !ok {
		return nil, newFieldNotFound("vm on disk attachment", "id")
	}
	disk, ok := object.Disk()
	if !ok {
		return nil, newFieldNotFound("disk attachment", "disk")
	}
	diskID, ok := disk.Id()
	if !ok {
		return nil, newFieldNotFound("disk on disk attachment", "id")
	}
	diskInterface, ok := object.Interface()
	if !ok {
		return nil, newFieldNotFound("disk attachment", "disk interface")
	}
	return &diskAttachment{
		client: o,

		id:            id,
		vmid:          vmID,
		diskID:        diskID,
		diskInterface: DiskInterface(diskInterface),
	}, nil
}
