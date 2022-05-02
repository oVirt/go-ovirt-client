package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) GetNIC(vmid string, id string, retries ...RetryStrategy) (result NIC, err error) {
	retries = defaultRetries(retries, defaultReadTimeouts(o))
	err = retry(
		fmt.Sprintf("getting NIC %s for VM %s", id, vmid),
		o.logger,
		retries,
		func() error {
			response, err := o.conn.SystemService().VmsService().VmService(vmid).NicsService().NicService(id).Get().Send()
			if err != nil {
				return err
			}
			sdkObject, ok := response.Nic()
			if !ok {
				return newError(
					ENotFound,
					"no NIC returned when getting NIC %s on VM %s",
					id,
					vmid,
				)
			}
			result, err = convertSDKNIC(sdkObject, o)
			if err != nil {
				return wrap(
					err,
					EBug,
					"failed to convert NIC %s on VM %s",
					id,
					vmid,
				)
			}
			return nil
		},
	)
	return result, err
}

func (m *mockClient) GetNIC(vmid string, id string, _ ...RetryStrategy) (NIC, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if nic, ok := m.nics[id]; ok {
		if nic.vmid != vmid {
			return nil, newError(ENotFound, "nic with ID %s not found", id)
		}
		return nic, nil
	}
	return nil, newError(ENotFound, "nic with ID %s not found", id)
}
