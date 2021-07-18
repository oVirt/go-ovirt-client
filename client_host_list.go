package ovirtclient

func (o *oVirtClient) ListHosts() ([]Host, error) {
	response, err := o.conn.SystemService().HostsService().List().Send()
	if err != nil {
		return nil, wrap(err, EUnidentified, "failed to list hosts")
	}
	sdkHosts, ok := response.Hosts()
	if !ok {
		return []Host{}, nil
	}
	result := make([]Host, len(sdkHosts.Slice()))
	for i, sdkHost := range sdkHosts.Slice() {
		result[i], err = convertSDKHost(sdkHost)
		if err != nil {
			return nil, wrap(err, EBug, "failed to convert host item %d", i)
		}
	}
	return result, nil
}
