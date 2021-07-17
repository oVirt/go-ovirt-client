package ovirtclient

import (
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

//go:generate go run scripts/rest.go -i "StorageDomain" -n "storage domain"

// StorageDomainClient contains the portion of the goVirt API that deals with storage domains.
type StorageDomainClient interface {
	// ListStorageDomains lists all storage domains.
	ListStorageDomains(retries ...RetryStrategy) ([]StorageDomain, error)
	// GetStorageDomain returns a single storage domain, or an error if the storage domain could not be found.
	GetStorageDomain(id string, retries ...RetryStrategy) (StorageDomain, error)
}

// StorageDomain represents a storage domain returned from the oVirt Engine API.
type StorageDomain interface {
	// ID is the unique identified for the storage system connected to oVirt.
	ID() string
	// Name is the user-given name for the storage domain.
	Name() string
	// Available returns the number of available bytes on the storage domain
	Available() uint64
	// Status returns the status of the storage domain. This status may be unknown if the storage domain is external.
	// Check ExternalStatus as well.
	Status() StorageDomainStatus
	// ExternalStatus returns the external status of a storage domain.
	ExternalStatus() StorageDomainExternalStatus
}

// StorageDomainStatus represents the status a domain can be in. Either this status field, or the
// StorageDomainExternalStatus must be set.
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
	StorageDomainStatusNA                      StorageDomainStatus = ""
)

// StorageDomainExternalStatus represents the status of an external storage domain.
type StorageDomainExternalStatus string

const (
	StorageDomainExternalStatusNA      StorageDomainExternalStatus = ""
	StorageDomainExternalStatusError   StorageDomainExternalStatus = "error"
	StorageDomainExternalStatusFailure StorageDomainExternalStatus = "failure"
	StorageDomainExternalStatusInfo    StorageDomainExternalStatus = "info"
	StorageDomainExternalStatusOk      StorageDomainExternalStatus = "ok"
	StorageDomainExternalStatusWarning StorageDomainExternalStatus = "warning"
)

func convertSDKStorageDomain(sdkStorageDomain *ovirtsdk4.StorageDomain) (StorageDomain, error) {
	id, ok := sdkStorageDomain.Id()
	if !ok {
		return nil, newError(EFieldMissing, "failed to fetch ID of storage domain")
	}
	name, ok := sdkStorageDomain.Name()
	if !ok {
		return nil, newError(EFieldMissing, "failed to fetch name of storage domain")
	}
	available, ok := sdkStorageDomain.Available()
	if !ok {
		// If this is not OK the status probably doesn't allow for reading disk space (e.g. unattached), so we return 0.
		available = 0
	}
	if available < 0 {
		return nil, newError(EBug, "invalid available bytes returned from storage domain: %d", available)
	}
	// It is OK for the storage domain status to not be present if the external status is present.
	status, _ := sdkStorageDomain.Status()
	// It is OK for the storage domain external status to not be present if the status is present.
	externalStatus, _ := sdkStorageDomain.ExternalStatus()
	if status == "" && externalStatus == "" {
		return nil, newError(EFieldMissing, "neither the status nor the external status is set for storage domain %s", id)
	}

	return &storageDomain{
		id:             id,
		name:           name,
		available:      uint64(available),
		status:         StorageDomainStatus(status),
		externalStatus: StorageDomainExternalStatus(externalStatus),
	}, nil
}

type storageDomain struct {
	id             string
	name           string
	available      uint64
	status         StorageDomainStatus
	externalStatus StorageDomainExternalStatus
}

func (s storageDomain) ID() string {
	return s.id
}

func (s storageDomain) Name() string {
	return s.name
}

func (s storageDomain) Available() uint64 {
	return s.available
}

func (s storageDomain) Status() StorageDomainStatus {
	return s.status
}

func (s storageDomain) ExternalStatus() StorageDomainExternalStatus {
	return s.externalStatus
}
