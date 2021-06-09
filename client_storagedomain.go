package govirt

import (
	"fmt"

	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

// StorageDomainClient contains the portion of the goVirt API that deals with storage domains.
type StorageDomainClient interface {
	ListStorageDomains() ([]StorageDomain, error)
	GetStorageDomain(id string) (StorageDomain, error)
}

// StorageDomain represents a storage domain returned from the oVirt Engine API.
type StorageDomain interface {
	ID() string
	Name() string
	// Available returns the number of available bytes on the storage domain
	Available() uint
	// Status returns the status of the storage domain. This status may be unknown if the storage domain is external.
	// Check ExternalStatus as well.
	Status() StorageDomainStatus
	// ExternalStatus returns the external status of a storage domain.
	ExternalStatus() StorageDomainExternalStatus
}

type StorageDomainStatus string

const (
	StorageDomainStatusActivating              StorageDomainStatus = "activating"
	StorageDomainStatusActive                  StorageDomainStatus = "active"
	StorageDomainStatusDetaching               StorageDomainStatus = "detaching"
	StorageDomainStatusInactive                StorageDomainStatus = "inactive"
	StorageDomainStatusLocked                  StorageDomainStatus = "locked"
	StorageDomainStatusMaintenance             StorageDomainStatus = "maintenance"
	StorageDomainStatusMixed                   StorageDomainStatus = "mixed"
	StorageDomainStatusPreparingForMaintenance StorageDomainStatus = "preparing_for_maintenance"
	StorageDomainStatusUnattached              StorageDomainStatus = "unattached"
	StorageDomainStatusUnknown                 StorageDomainStatus = "unknown"
)

type StorageDomainExternalStatus string

const (
	StorageDomainExternalStatusError   StorageDomainExternalStatus = "error"
	StorageDomainExternalStatusFailure StorageDomainExternalStatus = "failure"
	StorageDomainExternalStatusInfo    StorageDomainExternalStatus = "info"
	StorageDomainExternalStatusOk      StorageDomainExternalStatus = "ok"
	StorageDomainExternalStatusWarning StorageDomainExternalStatus = "warning"
)

func convertSDKStorageDomain(sdkStorageDomain *ovirtsdk4.StorageDomain) (StorageDomain, error) {
	id, ok := sdkStorageDomain.Id()
	if !ok {
		return nil, fmt.Errorf("failed to fetch ID of storage domain")
	}
	name, ok := sdkStorageDomain.Name()
	if !ok {
		return nil, fmt.Errorf("failed to fetch name of storage domain")
	}
	available, ok := sdkStorageDomain.Available()
	if !ok {
		return nil, fmt.Errorf("failed to fetch name of storage domain")
	}
	if available < 0 {
		return nil, fmt.Errorf("invalid available bytes returned from storage domain: %d", available)
	}
	status, ok := sdkStorageDomain.Status()
	if !ok {
		return nil, fmt.Errorf("failed to fetch status of storage domain")
	}
	externalStatus, ok := sdkStorageDomain.ExternalStatus()
	if !ok {
		return nil, fmt.Errorf("failed to fetch external status of storage domain")
	}

	return &storageDomain{
		id:             id,
		name:           name,
		available:      uint(available),
		status:         StorageDomainStatus(status),
		externalStatus: StorageDomainExternalStatus(externalStatus),
	}, nil
}

type storageDomain struct {
	id string
	name string
	available uint
	status StorageDomainStatus
	externalStatus StorageDomainExternalStatus
}

func (s storageDomain) ID() string {
	return s.id
}

func (s storageDomain) Name() string {
	return s.name
}

func (s storageDomain) Available() uint {
	return s.available
}

func (s storageDomain) Status() StorageDomainStatus {
	return s.status
}

func (s storageDomain) ExternalStatus() StorageDomainExternalStatus {
	return s.externalStatus
}

