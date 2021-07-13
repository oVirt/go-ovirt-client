package ovirtclient

func (o *oVirtClient) ListStorageDomains() (storageDomains []StorageDomain, err error) {
	response, err := o.conn.SystemService().StorageDomainsService().List().Send()
	if err != nil {
		return nil, wrap(err, EUnidentified, "failed to list storage domains")
	}
	sdkStorageDomains, ok := response.StorageDomains()
	if !ok {
		return nil, nil
	}
	storageDomains = make([]StorageDomain, len(sdkStorageDomains.Slice()))
	for i, sdkStorageDomain := range sdkStorageDomains.Slice() {
		storageDomain, err := convertSDKStorageDomain(sdkStorageDomain)
		if err != nil {
			return nil, wrap(err, EBug, "failed to convert storage domain %d in listing", i)
		}
		storageDomains[i] = storageDomain
	}
	return storageDomains, nil
}
