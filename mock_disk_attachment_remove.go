package ovirtclient

func (m *mockClient) RemoveDiskAttachment(vmID string, diskAttachmentID string, _ ...RetryStrategy) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	vm, ok := m.diskAttachmentsByVM[vmID]
	if !ok {
		return newError(ENotFound, "VM %s doesn't exist", vmID)
	}

	diskAttachment, ok := vm[diskAttachmentID]
	if !ok {
		return newError(ENotFound, "Disk attachment %s not found on VM %s", diskAttachmentID, vmID)
	}

	delete(m.diskAttachmentsByDisk, diskAttachment.DiskID())
	delete(m.diskAttachmentsByVM[vmID], diskAttachmentID)

	return nil
}
