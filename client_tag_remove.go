// Code generated automatically using go:generate. DO NOT EDIT.

package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) RemoveTag(tagID string, retries ...RetryStrategy) (err error) {
	retries = defaultRetries(retries, defaultWriteTimeouts())
	err = retry(
		fmt.Sprintf("removing tag %s", tagID),
		o.logger,
		retries,
		func() error {
			_, err := o.conn.SystemService().TagsService().TagService(string(tagID)).Remove().Send()
			return err
		})
	return
}
