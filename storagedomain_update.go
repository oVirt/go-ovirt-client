package ovirtclient

import (
	"fmt"

	ovirtsdk4 "github.com/ovirt/go-ovirt/v4"
)

func (o *oVirtClient) UpdateStorageDomain(
	id StorageDomainID,
	params UpdateStorageDomainParams,
	retries ...RetryStrategy,
) (result StorageDomain, err error) {
	retries = defaultRetries(retries, defaultWriteTimeouts(o))
	if params == nil {
		params = NewUpdateStorageDomainParams()
	}

	err = retry(
		fmt.Sprintf("updating storage domain %s", id),
		o.logger,
		retries,
		func() error {
			builder := ovirtsdk4.NewStorageDomainBuilder().Id(string(id))
			hasChanges := false
			if name := params.Name(); name != nil {
				builder.Name(*name)
				hasChanges = true
			}
			if desc := params.Description(); desc != nil {
				builder.Description(*desc)
				hasChanges = true
			}
			if comment := params.Comment(); comment != nil {
				builder.Comment(*comment)
				hasChanges = true
			}
			if wad := params.WipeAfterDelete(); wad != nil {
				builder.WipeAfterDelete(*wad)
				hasChanges = true
			}
			if b := params.Backup(); b != nil {
				builder.Backup(*b)
				hasChanges = true
			}
			if csab := params.CriticalSpaceActionBlocker(); csab != nil {
				builder.CriticalSpaceActionBlocker(*csab)
				hasChanges = true
			}
			if wlsi := params.WarningLowSpaceIndicator(); wlsi != nil {
				builder.WarningLowSpaceIndicator(*wlsi)
				hasChanges = true
			}

			if !hasChanges {
				return nil
			}

			response, e := o.conn.SystemService().StorageDomainsService().StorageDomainService(string(id)).Update().StorageDomain(builder.MustBuild()).Send()
			if e != nil {
				return e
			}
			sdkSD, ok := response.StorageDomain()
			if !ok {
				return newFieldNotFound("response from storage domain update", "storage_domain")
			}
			result, err = convertSDKStorageDomain(sdkSD, o)
			return err
		},
	)
	if err == nil && result == nil {
		return o.GetStorageDomain(id, retries...)
	}
	return result, err
}

func (m *mockClient) UpdateStorageDomain(
	id StorageDomainID,
	params UpdateStorageDomainParams,
	_ ...RetryStrategy,
) (StorageDomain, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	sd, ok := m.storageDomains[id]
	if !ok {
		return nil, newError(ENotFound, "storage domain with ID %s not found", id)
	}

	name := sd.name
	if newName := params.Name(); newName != nil {
		if *newName == "" {
			return nil, newError(EBadArgument, "name cannot be empty")
		}
		for _, existing := range m.storageDomains {
			if existing.id != id && existing.name == *newName {
				return nil, newError(EConflict, "storage domain with name %s already exists", *newName)
			}
		}
		name = *newName
	}

	description := sd.description
	if newDesc := params.Description(); newDesc != nil {
		description = newDesc
	}

	comment := sd.comment
	if newComment := params.Comment(); newComment != nil {
		comment = newComment
	}

	wipeAfterDelete := sd.wipeAfterDelete
	if newWipe := params.WipeAfterDelete(); newWipe != nil {
		wipeAfterDelete = newWipe
	}

	backup := sd.backup
	if newBackup := params.Backup(); newBackup != nil {
		backup = newBackup
	}

	csab := sd.criticalSpaceActionBlocker
	if newCSAB := params.CriticalSpaceActionBlocker(); newCSAB != nil {
		csab = newCSAB
	}

	wlsi := sd.warningLowSpaceIndicator
	if newWLSI := params.WarningLowSpaceIndicator(); newWLSI != nil {
		wlsi = newWLSI
	}

	sd.name = name
	sd.description = description
	sd.comment = comment
	sd.wipeAfterDelete = wipeAfterDelete
	sd.backup = backup
	sd.criticalSpaceActionBlocker = csab
	sd.warningLowSpaceIndicator = wlsi

	return sd, nil
}
