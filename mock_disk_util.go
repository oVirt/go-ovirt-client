package ovirtclient

import "time"

// WaitForDiskOK waits for a disk to be in the OK status, then additionally queries the job that was in progress with
// the correlation ID. This is necessary because the disk returns OK status before the job has actually finished,
// resulting in a "disk locked" error on subsequent operations. It uses checkDiskOk as an underlying function.
func (m *mockClient) WaitForDiskOK(diskID string, correlationID string, retries []RetryStrategy) (err error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	disk, ok := m.disks[diskID]

	if !ok {
		return newError(ENotFound, "Disk with ID %s not found", diskID)
	}
	time.Sleep(2 * time.Second)
	disk.status = DiskStatusOK

	return nil
}
