package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) RemoveTagFromVM(id string, tagID string, retries ...RetryStrategy) (err error) {
	retries = defaultRetries(retries, defaultWriteTimeouts())
	err = retry(
		fmt.Sprintf("removing tag from VM %s", id),
		o.logger,
		retries,
		func() error {
			_, err := o.conn.
				SystemService().
				VmsService().
				VmService(id).
				TagsService().
				TagService(tagID).
				Remove().
				Send()
			return err
		})
	return
}

func (m *mockClient) RemoveTagFromVM(id string, tagID string, retries ...RetryStrategy) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.vms[id]; !ok {
		return newError(ENotFound, "VM with ID %s not found", id)
	}
	if _, ok := m.tags[tagID]; !ok {
		return newError(ENotFound, "tag with ID %s not found", tagID)
	}
	foundIndex := -1
	for removeIndex, foundTagID := range m.vms[id].tagIDs {
		if tagID == foundTagID {
			foundIndex = removeIndex
		}
	}
	if foundIndex == -1 {
		return newError(ENotFound, "tag with ID %s not found on VM %s", tagID, id)
	}
	m.vms[id].tagIDs = append(m.vms[id].tagIDs[:foundIndex], m.vms[id].tagIDs[foundIndex+1:]...)
	return nil
}
