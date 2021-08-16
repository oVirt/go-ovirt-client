// Code generated automatically using go:generate. DO NOT EDIT.

package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) Get{{ .Object }}(id string, retries ...RetryStrategy) (result {{ .Object }}, err error) {
	retries = defaultRetries(retries, defaultReadTimeouts())
	err = retry(
		fmt.Sprintf("getting {{ .Name }} %s", id),
		o.logger,
		retries,
		func() error {
			response, err := o.conn.SystemService().{{ .ID }}sService().{{ .ID2 }}Service(id).Get().Send()
			if err != nil {
				return err
			}
			sdkObject, ok := response.{{ .ID2 }}()
			if !ok {
				return newError(
					ENotFound,
					"no {{ .Name }} returned when getting {{ .Name }} ID %s",
					id,
				)
			}
			result, err = convertSDK{{ .Object }}(sdkObject, o)
			if err != nil {
				return wrap(
					err,
					EBug,
					"failed to convert {{ .Name }} %s",
					id,
				)
			}
			return nil
		})
	return
}
