package ovirtclient

import (
	"fmt"

	ovirtsdk4 "github.com/ovirt/go-ovirt/v4"
)

func (o *oVirtClient) CreateNFSStorageDomain(
	name string,
	hostID HostID,
	address string,
	path string,
	params CreateNFSStorageDomainParams,
	retries ...RetryStrategy,
) (result StorageDomain, err error) {
	retries = defaultRetries(retries, defaultWriteTimeouts(o))
	if params == nil {
		params = NewCreateNFSStorageDomainParams()
	}
	if name == "" {
		return nil, newError(EBadArgument, "name cannot be empty for storage domain creation")
	}
	if hostID == "" {
		return nil, newError(EBadArgument, "hostID cannot be empty for storage domain creation")
	}
	if address == "" {
		return nil, newError(EBadArgument, "address cannot be empty for NFS storage domain creation")
	}
	if path == "" {
		return nil, newError(EBadArgument, "path cannot be empty for NFS storage domain creation")
	}

	err = retry(
		fmt.Sprintf("creating NFS storage domain %s", name),
		o.logger,
		retries,
		func() error {
			hostStorageBuilder := ovirtsdk4.NewHostStorageBuilder().
				Type(ovirtsdk4.STORAGETYPE_NFS).
				Address(address).
				Path(path)

			if mo := params.MountOptions(); mo != nil {
				hostStorageBuilder.MountOptions(*mo)
			}

			builder := ovirtsdk4.NewStorageDomainBuilder().
				Name(name).
				Type(ovirtsdk4.STORAGEDOMAINTYPE_DATA).
				Host(ovirtsdk4.NewHostBuilder().Id(string(hostID)).MustBuild()).
				Storage(hostStorageBuilder.MustBuild())

			if desc := params.Description(); desc != nil {
				builder.Description(*desc)
			}
			if comment := params.Comment(); comment != nil {
				builder.Comment(*comment)
			}
			if sf := params.StorageFormat(); sf != nil {
				builder.StorageFormat(ovirtsdk4.StorageFormat(*sf))
			}
			if wad := params.WipeAfterDelete(); wad != nil {
				builder.WipeAfterDelete(*wad)
			}
			if b := params.Backup(); b != nil {
				builder.Backup(*b)
			}
			if csab := params.CriticalSpaceActionBlocker(); csab != nil {
				builder.CriticalSpaceActionBlocker(*csab)
			}
			if wlsi := params.WarningLowSpaceIndicator(); wlsi != nil {
				builder.WarningLowSpaceIndicator(*wlsi)
			}

			response, e := o.conn.SystemService().StorageDomainsService().Add().StorageDomain(builder.MustBuild()).Send()
			if e != nil {
				return e
			}
			sdkSD, ok := response.StorageDomain()
			if !ok {
				return newFieldNotFound("response from storage domain creation", "storage_domain")
			}
			result, err = convertSDKStorageDomain(sdkSD, o)
			return err
		},
	)
	return result, err
}

func (m *mockClient) CreateNFSStorageDomain(
	name string,
	hostID HostID,
	address string,
	path string,
	params CreateNFSStorageDomainParams,
	_ ...RetryStrategy,
) (StorageDomain, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if params == nil {
		params = NewCreateNFSStorageDomainParams()
	}
	if name == "" {
		return nil, newError(EBadArgument, "name cannot be empty for storage domain creation")
	}
	if hostID == "" {
		return nil, newError(EBadArgument, "hostID cannot be empty for storage domain creation")
	}
	if address == "" {
		return nil, newError(EBadArgument, "address cannot be empty for NFS storage domain creation")
	}
	if path == "" {
		return nil, newError(EBadArgument, "path cannot be empty for NFS storage domain creation")
	}

	if _, ok := m.hosts[hostID]; !ok {
		return nil, newError(ENotFound, "host with ID %s not found", hostID)
	}

	for _, existing := range m.storageDomains {
		if existing.name == name {
			return nil, newError(EConflict, "storage domain with name %s already exists", name)
		}
	}

	id := StorageDomainID(m.GenerateUUID())
	format := StorageFormatV5
	if sf := params.StorageFormat(); sf != nil {
		format = *sf
	}

	sd := &storageDomain{
		client:                     m,
		id:                         id,
		name:                       name,
		available:                  50 * 1024 * 1024 * 1024,
		storageType:                StorageDomainTypeNFS,
		status:                     StorageDomainStatusActive,
		externalStatus:             StorageDomainExternalStatusNA,
		description:                params.Description(),
		comment:                    params.Comment(),
		storageAddress:             &address,
		storagePath:                &path,
		storageFormat:              &format,
		wipeAfterDelete:            params.WipeAfterDelete(),
		backup:                     params.Backup(),
		criticalSpaceActionBlocker: params.CriticalSpaceActionBlocker(),
		warningLowSpaceIndicator:   params.WarningLowSpaceIndicator(),
	}
	m.storageDomains[id] = sd
	return sd, nil
}
