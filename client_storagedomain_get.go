package ovirtclient

func (o *oVirtClient) GetStorageDomain(id string) (storageDomain StorageDomain, err error) {
	response, err := o.conn.SystemService().StorageDomainsService().StorageDomainService(id).Get().Send()
	if err != nil {
		return nil, wrap(err, EUnidentified, "failed to get storage domain %s", id)
	}
	sdkStorageDomain, ok := response.StorageDomain()
	if !ok {
		return nil, newError(ENotFound, "response did not contain a storage domain")
	}
	storageDomain, err = convertSDKStorageDomain(sdkStorageDomain)
	if err != nil {
		return nil, wrap(err, EUnidentified, "failed to convert storage domain")
	}
	return storageDomain, nil
}
