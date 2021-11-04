package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) WaitForStatus(id string, desiredStatus VMStatus, retries ...RetryStrategy) (VM, error) {
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
		fmt.Sprintf("waiting for vm %s to reach state %s", id, desiredStatus),
		o.logger,
		retries,
		func() error {
			status := vm.Status()
			if status != desiredStatus {
				return newError(EPending, "vm %s is at status %s, withing till it reaches status %s",
					id, status, desiredStatus)
			}
			return nil
		})
	return vm, err
}
