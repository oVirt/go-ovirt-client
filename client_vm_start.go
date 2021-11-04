package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) StartVM(id string, retries ...RetryStrategy) (VM, error) {
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
		fmt.Sprintf("starting vm %s", id),
		o.logger,
		retries,
		func() error {
			_, err = o.conn.SystemService().VmsService().VmService(id).Start().Send()
			if err != nil {
				return newError(
					EUnidentified,
					"vm with ID %s failed to start",
					id,
				)
			}
			err = vm.WaitForStatus(VMStatusUp, retries...)
			if err != nil {
				return newError(EUnidentified, "failed waiting for VM %s to reach status UP", id)
			}
			return nil
		})
	return vm, err
}
