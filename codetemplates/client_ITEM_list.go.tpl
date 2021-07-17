// Code generated automatically using go:generate. DO NOT EDIT.

package ovirtclient

func (o *oVirtClient) List{{ .ID }}s(retries ...RetryStrategy) (result []{{ .ID }}, err error) {
	retries = defaultRetries(retries, defaultReadTimeouts())
	result = []{{ .ID }}{}
	err = retry(
		"listing {{ .Name }}s",
		o.logger,
		retries,
		func() error {
			response, e := o.conn.SystemService().{{ .ID }}sService().List().Send()
			if e != nil {
				return e
			}
			sdkObjects, ok := response.{{ .ID }}s()
			if !ok {
				return nil
			}
			result = make([]{{ .ID }}, len(sdkObjects.Slice()))
			for i, sdkObject := range sdkObjects.Slice() {
				result[i], e = convertSDK{{ .ID }}(sdkObject)
				if e != nil {
					return wrap(e, EBug, "failed to convert {{ .Name }} during listing item #%d", i)
				}
			}
			return nil
		})
	return
}
