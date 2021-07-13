package ovirtclient

func (o *oVirtClient) GetHost(id string) (Host, error) {
	response, err := o.conn.SystemService().HostsService().HostService(id).Get().Send()
	if err != nil {
		return nil, wrap(
			err,
			EUnidentified,
			"failed to fetch host %s",
			id,
		)
	}
	sdkHost, ok := response.Host()
	if !ok {
		return nil, wrap(
			err,
			ENotFound,
			"host %s response did not contain a host",
		)
	}
	host, err := convertSDKHost(sdkHost)
	if err != nil {
		return nil, wrap(
			err,
			EBug,
			"failed to convert host %s",
			id,
		)
	}
	return host, nil
}
