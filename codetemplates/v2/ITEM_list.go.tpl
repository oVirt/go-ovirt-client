// Code generated automatically using go:generate. DO NOT EDIT.

package ovirtclient

func (o *oVirtClient) List{{ .Object }}s(retries ...RetryStrategy) (result []{{ .Object }}, err error) {
	retries = defaultRetries(retries, defaultReadTimeouts(o))
	result = []{{ .Object }}{}
	err = retry(
		"listing {{ .Name }}s",
		o.logger,
		retries,
		func() error {
			response, e := o.conn.SystemService().{{ .ID }}sService().List().Send()
			if e != nil {
				return e
			}
			sdkObjects, ok := response.{{ .SecondaryID }}s()
			if !ok {
				return nil
			}
			result = make([]{{ .Object }}, len(sdkObjects.Slice()))
			for i, sdkObject := range sdkObjects.Slice() {
				result[i], e = convertSDK{{ .Object }}(sdkObject, o)
				if e != nil {
					return wrap(e, EBug, "failed to convert {{ .Name }} during listing item #%d", i)
				}
			}
			return nil
		})
	return
}

func (m *mockClient) List{{ .Object }}s(_ ...RetryStrategy) ([]{{ .Object }}, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	result := make([]{{ .Object }}, len(m.{{ .ID | toLower }}s))
	i := 0
	for _, item := range m.{{ .ID | toLower }}s {
		result[i] = item
		i++
	}
	return result, nil
}
