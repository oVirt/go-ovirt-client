package ovirtclient

func (m *mockClient) RemoveVM(id string, _ ...RetryStrategy) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.vms[id]; !ok {
		return newError(ENotFound, "VM with ID %s not found", id)
	}

	for _, diskAttachment := range m.vmDiskAttachmentsByVM[id] {
		if m.disks[diskAttachment.DiskID()].status == DiskStatusLocked {
			return newError(EConflict, "Cannot delete VM, disk %s is locked.", diskAttachment.DiskID())
		}
		delete(m.disks, diskAttachment.DiskID())
		delete(m.vmDiskAttachmentsByDisk, diskAttachment.DiskID())
	}
	for nicID, nic := range m.nics {
		if nic.VMID() == id {
			delete(m.nics, nicID)
		}
	}
	delete(m.vmDiskAttachmentsByVM, id)
	delete(m.vms, id)

	return nil
}
