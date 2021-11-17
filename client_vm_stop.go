package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) StopVM(id string, retries ...RetryStrategy) (VM, error) {
	retries = defaultRetries(retries, defaultReadTimeouts())
	vm, err := o.GetVM(id, retries...)
	if err != nil {
		return nil, newError(
			ENotFound,
			"vm with ID %s was not found",
			id,
		)
	}
	err = retry(
		fmt.Sprintf("stopping vm %s", id),
		o.logger,
		retries,
		func() error {
			_, err = o.conn.SystemService().VmsService().VmService(id).Stop().Send()
			if err != nil {
				return wrap(
					err,
					EUnidentified,
					"vm with ID %s failed to stop",
					id,
				)
			}
			err = vm.WaitForStatus(VMStatusDown, retries...)
			if err != nil {
				return wrap(err, EUnidentified, "failed waiting for VM %s to reach status Down", id)
			}
			return nil
		})
	return vm, err
}
