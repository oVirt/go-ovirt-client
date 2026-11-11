package ovirtclient

import (
	"strings"
	"sync"

	ovirtsdk4 "github.com/ovirt/go-ovirt/v4"
)

// StorageDomainID is a specialized type for storage domain IDs.
type StorageDomainID string

// StorageDomainClient contains the portion of the goVirt API that deals with storage domains.
type StorageDomainClient interface {
	// CreateNFSStorageDomain creates a new NFS storage domain.
	CreateNFSStorageDomain(
		name string,
		hostID HostID,
		address string,
		path string,
		params CreateNFSStorageDomainParams,
		retries ...RetryStrategy,
	) (StorageDomain, error)
	// ListStorageDomains lists all storage domains.
	ListStorageDomains(retries ...RetryStrategy) (StorageDomainList, error)
	// GetStorageDomain returns a single storage domain, or an error if the storage domain could not be found.
	GetStorageDomain(id StorageDomainID, retries ...RetryStrategy) (StorageDomain, error)
	// GetDiskFromStorageDomain returns a single disk from a specific storage domain, or an error if no disk can be found.
	GetDiskFromStorageDomain(id StorageDomainID, diskID DiskID, retries ...RetryStrategy) (result Disk, err error)
	// RemoveDiskFromStorageDomain removes a disk from a specific storage domain, but leaves the disk on other storage
	// domains if any. If the disk is not present on any more storage domains, the entire disk will be removed.
	RemoveDiskFromStorageDomain(id StorageDomainID, diskID DiskID, retries ...RetryStrategy) error
	// UpdateStorageDomain updates an existing storage domain.
	UpdateStorageDomain(id StorageDomainID, params UpdateStorageDomainParams, retries ...RetryStrategy) (StorageDomain, error)
	// RemoveStorageDomain removes a storage domain.
	RemoveStorageDomain(id StorageDomainID, params RemoveStorageDomainParams, retries ...RetryStrategy) error
}

// StorageDomainData is the core of StorageDomain, providing only data access functions.
type StorageDomainData interface {
	// ID is the unique identified for the storage system connected to oVirt.
	ID() StorageDomainID
	// Name is the user-given name for the storage domain.
	Name() string
	// Available returns the number of available bytes on the storage domain
	Available() uint64
	// StorageType returns the type of the storage domain
	StorageType() StorageDomainType
	// Status returns the status of the storage domain. This status may be unknown if the storage domain is external.
	// Check ExternalStatus as well.
	Status() StorageDomainStatus
	// ExternalStatus returns the external status of a storage domain.
	ExternalStatus() StorageDomainExternalStatus
	// Description returns the user-given description for the storage domain.
	Description() *string
	// Comment returns the user-given comment for the storage domain.
	Comment() *string
	// StorageAddress returns the host address of the storage server.
	StorageAddress() *string
	// StoragePath returns the path or export of the storage server.
	StoragePath() *string
	// StorageFormat returns the storage format (e.g. v5).
	StorageFormat() *StorageFormat
	// WipeAfterDelete returns whether wipe-after-delete is enabled by default on the storage domain.
	WipeAfterDelete() *bool
	// Backup returns whether the storage domain is marked for backup.
	Backup() *bool
	// CriticalSpaceActionBlocker returns the percentage threshold to block actions.
	CriticalSpaceActionBlocker() *int64
	// WarningLowSpaceIndicator returns the percentage threshold for low space warning.
	WarningLowSpaceIndicator() *int64
}

// StorageDomain represents a storage domain returned from the oVirt Engine API.
type StorageDomain interface {
	StorageDomainData

	Update(params UpdateStorageDomainParams, retries ...RetryStrategy) (StorageDomain, error)
	Remove(params RemoveStorageDomainParams, retries ...RetryStrategy) error
}

// StorageDomainList represents a list of storage domains.
type StorageDomainList []StorageDomain

// Filter applies the passed in filter function to all items in the current list and returns a new list with items that returned true.
func (list StorageDomainList) Filter(apply func(StorageDomain) bool) StorageDomainList {
	var filtered StorageDomainList
	for _, sd := range list {
		if apply(sd) {
			filtered = append(filtered, sd)
		}
	}
	return filtered
}

// StorageDomainType represents the type of the storage domain.
type StorageDomainType string

const (
	// StorageDomainTypeCinder represents a cinder host storage type.
	StorageDomainTypeCinder StorageDomainType = "cinder"
	// StorageDomainTypeFCP represents a fcp host storage type.
	StorageDomainTypeFCP StorageDomainType = "fcp"
	// StorageDomainTypeGlance represents a glance host storage type.
	StorageDomainTypeGlance StorageDomainType = "glance"
	// StorageDomainTypeGlusterFS represents a glusterfs host storage type.
	StorageDomainTypeGlusterFS StorageDomainType = "glusterfs"
	// StorageDomainTypeISCSI represents a iscsi host storage type.
	StorageDomainTypeISCSI StorageDomainType = "iscsi"
	// StorageDomainTypeLocalFS represents a localfs host storage type.
	StorageDomainTypeLocalFS StorageDomainType = "localfs"
	// StorageDomainTypeManagedBlockStorage represents a managed block storage host storage type.
	StorageDomainTypeManagedBlockStorage StorageDomainType = "managed_block_storage"
	// StorageDomainTypeNFS represents a nfs host storage type.
	StorageDomainTypeNFS StorageDomainType = "nfs"
	// StorageDomainTypePosixFS represents a posixfs host storage type.
	StorageDomainTypePosixFS StorageDomainType = "posixfs"
)

// StorageDomainTypeList is a list of possible StorageDomainTypes.
type StorageDomainTypeList []StorageDomainType

// FileStorageDomainTypeList is a list of possible StorageDomainTypes which are considered file storage.
type FileStorageDomainTypeList []StorageDomainType

// FileStorageDomainTypeValues returns all the StorageDomainTypes values which are considered file storage.
func FileStorageDomainTypeValues() FileStorageDomainTypeList {
	return []StorageDomainType{
		StorageDomainTypeGlusterFS,
		StorageDomainTypeLocalFS,
		StorageDomainTypeNFS,
		StorageDomainTypePosixFS,
	}
}

// StorageDomainTypeValues returns all possible StorageDomainTypeValues values.
func StorageDomainTypeValues() StorageDomainTypeList {
	return []StorageDomainType{
		StorageDomainTypeCinder,
		StorageDomainTypeFCP,
		StorageDomainTypeGlance,
		StorageDomainTypeGlusterFS,
		StorageDomainTypeISCSI,
		StorageDomainTypeLocalFS,
		StorageDomainTypeManagedBlockStorage,
		StorageDomainTypeNFS,
		StorageDomainTypePosixFS,
	}
}

// StorageDomainStatus represents the status a domain can be in. Either this status field, or the
// StorageDomainExternalStatus must be set.
//
// Note: this is not well documented due to missing source documentation. If you know something about these statuses
// please contribute here:
// https://github.com/oVirt/ovirt-engine-api-model/blob/master/src/main/java/types/StorageDomainStatus.java
type StorageDomainStatus string

// Validate returns an error if the storage domain status doesn't have a valid value.
func (s StorageDomainStatus) Validate() error {
	for _, format := range StorageDomainStatusValues() {
		if format == s {
			return nil
		}
	}
	return newError(
		EBadArgument,
		"invalid storage domain status: %s must be one of: %s",
		s,
		strings.Join(ImageFormatValues().Strings(), ", "),
	)
}

const (
	// StorageDomainStatusActivating indicates that the storage domain is currently activating and will soon be active.
	StorageDomainStatusActivating StorageDomainStatus = "activating"
	// StorageDomainStatusActive is the normal status for a storage domain when it's working.
	StorageDomainStatusActive StorageDomainStatus = "active"
	// StorageDomainStatusDetaching is the status when it is being disconnected.
	StorageDomainStatusDetaching StorageDomainStatus = "detaching"
	// StorageDomainStatusInactive is an undocumented status of the storage domain.
	StorageDomainStatusInactive StorageDomainStatus = "inactive"
	// StorageDomainStatusLocked is an undocumented status of the storage domain.
	StorageDomainStatusLocked StorageDomainStatus = "locked"
	// StorageDomainStatusMaintenance is an undocumented status of the storage domain.
	StorageDomainStatusMaintenance StorageDomainStatus = "maintenance"
	// StorageDomainStatusMixed is an undocumented status of the storage domain.
	StorageDomainStatusMixed StorageDomainStatus = "mixed"
	// StorageDomainStatusPreparingForMaintenance is an undocumented status of the storage domain.
	StorageDomainStatusPreparingForMaintenance StorageDomainStatus = "preparing_for_maintenance"
	// StorageDomainStatusUnattached is an undocumented status of the storage domain.
	StorageDomainStatusUnattached StorageDomainStatus = "unattached"
	// StorageDomainStatusUnknown is an undocumented status of the storage domain.
	StorageDomainStatusUnknown StorageDomainStatus = "unknown"
	// StorageDomainStatusNA indicates that the storage domain does not have a status. Please check the external status
	// instead.
	StorageDomainStatusNA StorageDomainStatus = ""
)

// StorageDomainStatusList is a list of StorageDomainStatus.
type StorageDomainStatusList []StorageDomainStatus

// StorageDomainStatusValues returns all possible StorageDomainStatus values.
func StorageDomainStatusValues() StorageDomainStatusList {
	return []StorageDomainStatus{
		StorageDomainStatusActivating,
		StorageDomainStatusActive,
		StorageDomainStatusDetaching,
		StorageDomainStatusInactive,
		StorageDomainStatusLocked,
		StorageDomainStatusMaintenance,
		StorageDomainStatusMixed,
		StorageDomainStatusPreparingForMaintenance,
		StorageDomainStatusUnattached,
		StorageDomainStatusUnknown,
		StorageDomainStatusNA,
	}
}

// Strings creates a string list of the values.
func (l StorageDomainStatusList) Strings() []string {
	result := make([]string, len(l))
	for i, status := range l {
		result[i] = string(status)
	}
	return result
}

// StorageDomainExternalStatus represents the status of an external storage domain. This status is updated externally.
//
// Note: this is not well-defined as the oVirt model has only a very generic description. See
// https://github.com/oVirt/ovirt-engine-api-model/blob/9869596c298925538d510de5019195b488970738/src/main/java/types/ExternalStatus.java
// for details.
type StorageDomainExternalStatus string

const (
	// StorageDomainExternalStatusNA represents an external status that is not applicable.
	// Most likely, the status should be obtained from StorageDomainStatus, since the
	// storage domain in question is not an external storage.
	StorageDomainExternalStatusNA StorageDomainExternalStatus = ""
	// StorageDomainExternalStatusError indicates an error state.
	StorageDomainExternalStatusError StorageDomainExternalStatus = "error"
	// StorageDomainExternalStatusFailure indicates a failure state.
	StorageDomainExternalStatusFailure StorageDomainExternalStatus = "failure"
	// StorageDomainExternalStatusInfo indicates an OK status, but there is information available for the administrator
	// that might be relevant.
	StorageDomainExternalStatusInfo StorageDomainExternalStatus = "info"
	// StorageDomainExternalStatusOk indicates a working status.
	StorageDomainExternalStatusOk StorageDomainExternalStatus = "ok"
	// StorageDomainExternalStatusWarning indicates that the storage domain has warnings that may be relevant for the
	// administrator.
	StorageDomainExternalStatusWarning StorageDomainExternalStatus = "warning"
)

// StorageDomainExternalStatusList is a list of StorageDomainStatus.
type StorageDomainExternalStatusList []StorageDomainExternalStatus

// StorageDomainExternalStatusValues returns all possible StorageDomainExternalStatus values.
func StorageDomainExternalStatusValues() StorageDomainExternalStatusList {
	return []StorageDomainExternalStatus{
		StorageDomainExternalStatusNA,
		StorageDomainExternalStatusError,
		StorageDomainExternalStatusFailure,
		StorageDomainExternalStatusInfo,
		StorageDomainExternalStatusOk,
		StorageDomainExternalStatusWarning,
	}
}

// Strings creates a string list of the values.
func (l StorageDomainExternalStatusList) Strings() []string {
	result := make([]string, len(l))
	for i, status := range l {
		result[i] = string(status)
	}
	return result
}

func convertSDKStorageDomain(sdkStorageDomain *ovirtsdk4.StorageDomain, client Client) (StorageDomain, error) {
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
	storage, ok := sdkStorageDomain.Storage()
	if !ok {
		return nil, newError(EFieldMissing, "failed to fetch hostStorage of storage domain")
	}
	storageType, ok := storage.Type()
	if !ok {
		return nil, newError(EFieldMissing, "failed to fetch storage type of storage domain")
	}
	// It is OK for the storage domain status to not be present if the external status is present.
	status, _ := sdkStorageDomain.Status()
	// It is OK for the storage domain external status to not be present if the status is present.
	externalStatus, _ := sdkStorageDomain.ExternalStatus()
	if status == "" && externalStatus == "" {
		return nil, newError(EFieldMissing, "neither the status nor the external status is set for storage domain %s", id)
	}

	var description *string
	if d, ok := sdkStorageDomain.Description(); ok {
		description = &d
	}
	var comment *string
	if c, ok := sdkStorageDomain.Comment(); ok {
		comment = &c
	}
	var storageAddress *string
	if a, ok := storage.Address(); ok {
		storageAddress = &a
	}
	var storagePath *string
	if p, ok := storage.Path(); ok {
		storagePath = &p
	}
	var storageFormat *StorageFormat
	if sf, ok := sdkStorageDomain.StorageFormat(); ok {
		s := StorageFormat(sf)
		storageFormat = &s
	}
	var wipeAfterDelete *bool
	if wad, ok := sdkStorageDomain.WipeAfterDelete(); ok {
		wipeAfterDelete = &wad
	}
	var backup *bool
	if b, ok := sdkStorageDomain.Backup(); ok {
		backup = &b
	}
	var csab *int64
	if c, ok := sdkStorageDomain.CriticalSpaceActionBlocker(); ok {
		csab = &c
	}
	var wlsi *int64
	if w, ok := sdkStorageDomain.WarningLowSpaceIndicator(); ok {
		wlsi = &w
	}

	return &storageDomain{
		client: client,

		id:                         StorageDomainID(id),
		name:                       name,
		available:                  uint64(available),
		storageType:                StorageDomainType(storageType),
		status:                     StorageDomainStatus(status),
		externalStatus:             StorageDomainExternalStatus(externalStatus),
		description:                description,
		comment:                    comment,
		storageAddress:             storageAddress,
		storagePath:                storagePath,
		storageFormat:              storageFormat,
		wipeAfterDelete:            wipeAfterDelete,
		backup:                     backup,
		criticalSpaceActionBlocker: csab,
		warningLowSpaceIndicator:   wlsi,
	}, nil
}

type storageDomain struct {
	client Client

	id                         StorageDomainID
	name                       string
	available                  uint64
	storageType                StorageDomainType
	status                     StorageDomainStatus
	externalStatus             StorageDomainExternalStatus
	description                *string
	comment                    *string
	storageAddress             *string
	storagePath                *string
	storageFormat              *StorageFormat
	wipeAfterDelete            *bool
	backup                     *bool
	criticalSpaceActionBlocker *int64
	warningLowSpaceIndicator   *int64
}

func (s storageDomain) ID() StorageDomainID {
	return s.id
}

func (s storageDomain) Name() string {
	return s.name
}

func (s storageDomain) Available() uint64 {
	return s.available
}

func (s storageDomain) StorageType() StorageDomainType {
	return s.storageType
}

func (s storageDomain) Status() StorageDomainStatus {
	return s.status
}

func (s storageDomain) ExternalStatus() StorageDomainExternalStatus {
	return s.externalStatus
}

func (s storageDomain) Description() *string {
	return s.description
}

func (s storageDomain) Comment() *string {
	return s.comment
}

func (s storageDomain) StorageAddress() *string {
	return s.storageAddress
}

func (s storageDomain) StoragePath() *string {
	return s.storagePath
}

func (s storageDomain) StorageFormat() *StorageFormat {
	return s.storageFormat
}

func (s storageDomain) WipeAfterDelete() *bool {
	return s.wipeAfterDelete
}

func (s storageDomain) Backup() *bool {
	return s.backup
}

func (s storageDomain) CriticalSpaceActionBlocker() *int64 {
	return s.criticalSpaceActionBlocker
}

func (s storageDomain) WarningLowSpaceIndicator() *int64 {
	return s.warningLowSpaceIndicator
}

func (s storageDomain) Update(params UpdateStorageDomainParams, retries ...RetryStrategy) (StorageDomain, error) {
	return s.client.UpdateStorageDomain(s.id, params, retries...)
}

func (s storageDomain) Remove(params RemoveStorageDomainParams, retries ...RetryStrategy) error {
	return s.client.RemoveStorageDomain(s.id, params, retries...)
}

type storageDomainDiskWait struct {
	client        *oVirtClient
	disk          Disk
	storageDomain StorageDomain
	correlationID string
	lock          *sync.Mutex
}

func (d *storageDomainDiskWait) Disk() Disk {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.disk
}

func (d *storageDomainDiskWait) Wait(retries ...RetryStrategy) (Disk, error) {
	retries = defaultRetries(retries, defaultWriteTimeouts(d.client))
	d.lock.Lock()
	diskID := d.disk.ID()
	storageDomainID := d.storageDomain.ID()
	d.lock.Unlock()

	if _, err := d.client.WaitForDiskOK(diskID, retries...); err != nil {
		return nil, err
	}

	if err := d.client.waitForJobFinished(d.correlationID, retries); err != nil {
		return nil, err
	}

	disk, err := d.client.GetDiskFromStorageDomain(storageDomainID, diskID)

	if disk != nil {
		d.disk = disk
	}
	return disk, err
}

// StorageFormat represents the storage format version of a storage domain.
type StorageFormat string

const (
	StorageFormatV1 StorageFormat = "v1"
	StorageFormatV2 StorageFormat = "v2"
	StorageFormatV3 StorageFormat = "v3"
	StorageFormatV4 StorageFormat = "v4"
	StorageFormatV5 StorageFormat = "v5"
)

// CreateNFSStorageDomainParams contains optional parameters for NFS storage domain creation.
type CreateNFSStorageDomainParams interface {
	Description() *string
	Comment() *string
	StorageFormat() *StorageFormat
	WipeAfterDelete() *bool
	Backup() *bool
	CriticalSpaceActionBlocker() *int64
	WarningLowSpaceIndicator() *int64
	MountOptions() *string
}

// BuildableNFSStorageDomainParams is a buildable version of CreateNFSStorageDomainParams.
type BuildableNFSStorageDomainParams interface {
	CreateNFSStorageDomainParams

	WithDescription(description string) (BuildableNFSStorageDomainParams, error)
	MustWithDescription(description string) BuildableNFSStorageDomainParams
	WithComment(comment string) (BuildableNFSStorageDomainParams, error)
	MustWithComment(comment string) BuildableNFSStorageDomainParams
	WithStorageFormat(format StorageFormat) (BuildableNFSStorageDomainParams, error)
	MustWithStorageFormat(format StorageFormat) BuildableNFSStorageDomainParams
	WithWipeAfterDelete(wipe bool) (BuildableNFSStorageDomainParams, error)
	MustWithWipeAfterDelete(wipe bool) BuildableNFSStorageDomainParams
	WithBackup(backup bool) (BuildableNFSStorageDomainParams, error)
	MustWithBackup(backup bool) BuildableNFSStorageDomainParams
	WithCriticalSpaceActionBlocker(percent int64) (BuildableNFSStorageDomainParams, error)
	MustWithCriticalSpaceActionBlocker(percent int64) BuildableNFSStorageDomainParams
	WithWarningLowSpaceIndicator(percent int64) (BuildableNFSStorageDomainParams, error)
	MustWithWarningLowSpaceIndicator(percent int64) BuildableNFSStorageDomainParams
	WithMountOptions(options string) (BuildableNFSStorageDomainParams, error)
	MustWithMountOptions(options string) BuildableNFSStorageDomainParams
}

// NewCreateNFSStorageDomainParams creates a buildable set of CreateNFSStorageDomainParams.
func NewCreateNFSStorageDomainParams() BuildableNFSStorageDomainParams {
	return &createNFSStorageDomainParams{}
}

type createNFSStorageDomainParams struct {
	description                *string
	comment                    *string
	storageFormat              *StorageFormat
	wipeAfterDelete            *bool
	backup                     *bool
	criticalSpaceActionBlocker *int64
	warningLowSpaceIndicator   *int64
	mountOptions               *string
}

func (c *createNFSStorageDomainParams) Description() *string {
	return c.description
}

func (c *createNFSStorageDomainParams) Comment() *string {
	return c.comment
}

func (c *createNFSStorageDomainParams) StorageFormat() *StorageFormat {
	return c.storageFormat
}

func (c *createNFSStorageDomainParams) WipeAfterDelete() *bool {
	return c.wipeAfterDelete
}

func (c *createNFSStorageDomainParams) Backup() *bool {
	return c.backup
}

func (c *createNFSStorageDomainParams) CriticalSpaceActionBlocker() *int64 {
	return c.criticalSpaceActionBlocker
}

func (c *createNFSStorageDomainParams) WarningLowSpaceIndicator() *int64 {
	return c.warningLowSpaceIndicator
}

func (c *createNFSStorageDomainParams) MountOptions() *string {
	return c.mountOptions
}

func (c *createNFSStorageDomainParams) WithDescription(description string) (BuildableNFSStorageDomainParams, error) {
	c.description = &description
	return c, nil
}

func (c *createNFSStorageDomainParams) MustWithDescription(description string) BuildableNFSStorageDomainParams {
	b, err := c.WithDescription(description)
	if err != nil {
		panic(err)
	}
	return b
}

func (c *createNFSStorageDomainParams) WithComment(comment string) (BuildableNFSStorageDomainParams, error) {
	c.comment = &comment
	return c, nil
}

func (c *createNFSStorageDomainParams) MustWithComment(comment string) BuildableNFSStorageDomainParams {
	b, err := c.WithComment(comment)
	if err != nil {
		panic(err)
	}
	return b
}

func (c *createNFSStorageDomainParams) WithStorageFormat(format StorageFormat) (BuildableNFSStorageDomainParams, error) {
	c.storageFormat = &format
	return c, nil
}

func (c *createNFSStorageDomainParams) MustWithStorageFormat(format StorageFormat) BuildableNFSStorageDomainParams {
	b, err := c.WithStorageFormat(format)
	if err != nil {
		panic(err)
	}
	return b
}

func (c *createNFSStorageDomainParams) WithWipeAfterDelete(wipe bool) (BuildableNFSStorageDomainParams, error) {
	c.wipeAfterDelete = &wipe
	return c, nil
}

func (c *createNFSStorageDomainParams) MustWithWipeAfterDelete(wipe bool) BuildableNFSStorageDomainParams {
	b, err := c.WithWipeAfterDelete(wipe)
	if err != nil {
		panic(err)
	}
	return b
}

func (c *createNFSStorageDomainParams) WithBackup(backup bool) (BuildableNFSStorageDomainParams, error) {
	c.backup = &backup
	return c, nil
}

func (c *createNFSStorageDomainParams) MustWithBackup(backup bool) BuildableNFSStorageDomainParams {
	b, err := c.WithBackup(backup)
	if err != nil {
		panic(err)
	}
	return b
}

func (c *createNFSStorageDomainParams) WithCriticalSpaceActionBlocker(percent int64) (BuildableNFSStorageDomainParams, error) {
	if percent < 0 || percent > 100 {
		return nil, newError(EBadArgument, "critical space action blocker must be between 0 and 100")
	}
	c.criticalSpaceActionBlocker = &percent
	return c, nil
}

func (c *createNFSStorageDomainParams) MustWithCriticalSpaceActionBlocker(percent int64) BuildableNFSStorageDomainParams {
	b, err := c.WithCriticalSpaceActionBlocker(percent)
	if err != nil {
		panic(err)
	}
	return b
}

func (c *createNFSStorageDomainParams) WithWarningLowSpaceIndicator(percent int64) (BuildableNFSStorageDomainParams, error) {
	if percent < 0 || percent > 100 {
		return nil, newError(EBadArgument, "warning low space indicator must be between 0 and 100")
	}
	c.warningLowSpaceIndicator = &percent
	return c, nil
}

func (c *createNFSStorageDomainParams) MustWithWarningLowSpaceIndicator(percent int64) BuildableNFSStorageDomainParams {
	b, err := c.WithWarningLowSpaceIndicator(percent)
	if err != nil {
		panic(err)
	}
	return b
}

func (c *createNFSStorageDomainParams) WithMountOptions(options string) (BuildableNFSStorageDomainParams, error) {
	c.mountOptions = &options
	return c, nil
}

func (c *createNFSStorageDomainParams) MustWithMountOptions(options string) BuildableNFSStorageDomainParams {
	b, err := c.WithMountOptions(options)
	if err != nil {
		panic(err)
	}
	return b
}

// UpdateStorageDomainParams contains parameters for updating an existing storage domain.
type UpdateStorageDomainParams interface {
	Name() *string
	Description() *string
	Comment() *string
	WipeAfterDelete() *bool
	Backup() *bool
	CriticalSpaceActionBlocker() *int64
	WarningLowSpaceIndicator() *int64
}

// BuildableUpdateStorageDomainParams is a buildable version of UpdateStorageDomainParams.
type BuildableUpdateStorageDomainParams interface {
	UpdateStorageDomainParams

	WithName(name string) (BuildableUpdateStorageDomainParams, error)
	MustWithName(name string) BuildableUpdateStorageDomainParams
	WithDescription(description string) (BuildableUpdateStorageDomainParams, error)
	MustWithDescription(description string) BuildableUpdateStorageDomainParams
	WithComment(comment string) (BuildableUpdateStorageDomainParams, error)
	MustWithComment(comment string) BuildableUpdateStorageDomainParams
	WithWipeAfterDelete(wipe bool) (BuildableUpdateStorageDomainParams, error)
	MustWithWipeAfterDelete(wipe bool) BuildableUpdateStorageDomainParams
	WithBackup(backup bool) (BuildableUpdateStorageDomainParams, error)
	MustWithBackup(backup bool) BuildableUpdateStorageDomainParams
	WithCriticalSpaceActionBlocker(percent int64) (BuildableUpdateStorageDomainParams, error)
	MustWithCriticalSpaceActionBlocker(percent int64) BuildableUpdateStorageDomainParams
	WithWarningLowSpaceIndicator(percent int64) (BuildableUpdateStorageDomainParams, error)
	MustWithWarningLowSpaceIndicator(percent int64) BuildableUpdateStorageDomainParams
}

// NewUpdateStorageDomainParams creates a buildable set of UpdateStorageDomainParams.
func NewUpdateStorageDomainParams() BuildableUpdateStorageDomainParams {
	return &updateStorageDomainParams{}
}

type updateStorageDomainParams struct {
	name                       *string
	description                *string
	comment                    *string
	wipeAfterDelete            *bool
	backup                     *bool
	criticalSpaceActionBlocker *int64
	warningLowSpaceIndicator   *int64
}

func (u *updateStorageDomainParams) Name() *string {
	return u.name
}

func (u *updateStorageDomainParams) Description() *string {
	return u.description
}

func (u *updateStorageDomainParams) Comment() *string {
	return u.comment
}

func (u *updateStorageDomainParams) WipeAfterDelete() *bool {
	return u.wipeAfterDelete
}

func (u *updateStorageDomainParams) Backup() *bool {
	return u.backup
}

func (u *updateStorageDomainParams) CriticalSpaceActionBlocker() *int64 {
	return u.criticalSpaceActionBlocker
}

func (u *updateStorageDomainParams) WarningLowSpaceIndicator() *int64 {
	return u.warningLowSpaceIndicator
}

func (u *updateStorageDomainParams) WithName(name string) (BuildableUpdateStorageDomainParams, error) {
	if name == "" {
		return nil, newError(EBadArgument, "name cannot be empty")
	}
	u.name = &name
	return u, nil
}

func (u *updateStorageDomainParams) MustWithName(name string) BuildableUpdateStorageDomainParams {
	b, err := u.WithName(name)
	if err != nil {
		panic(err)
	}
	return b
}

func (u *updateStorageDomainParams) WithDescription(description string) (BuildableUpdateStorageDomainParams, error) {
	u.description = &description
	return u, nil
}

func (u *updateStorageDomainParams) MustWithDescription(description string) BuildableUpdateStorageDomainParams {
	b, err := u.WithDescription(description)
	if err != nil {
		panic(err)
	}
	return b
}

func (u *updateStorageDomainParams) WithComment(comment string) (BuildableUpdateStorageDomainParams, error) {
	u.comment = &comment
	return u, nil
}

func (u *updateStorageDomainParams) MustWithComment(comment string) BuildableUpdateStorageDomainParams {
	b, err := u.WithComment(comment)
	if err != nil {
		panic(err)
	}
	return b
}

func (u *updateStorageDomainParams) WithWipeAfterDelete(wipe bool) (BuildableUpdateStorageDomainParams, error) {
	u.wipeAfterDelete = &wipe
	return u, nil
}

func (u *updateStorageDomainParams) MustWithWipeAfterDelete(wipe bool) BuildableUpdateStorageDomainParams {
	b, err := u.WithWipeAfterDelete(wipe)
	if err != nil {
		panic(err)
	}
	return b
}

func (u *updateStorageDomainParams) WithBackup(backup bool) (BuildableUpdateStorageDomainParams, error) {
	u.backup = &backup
	return u, nil
}

func (u *updateStorageDomainParams) MustWithBackup(backup bool) BuildableUpdateStorageDomainParams {
	b, err := u.WithBackup(backup)
	if err != nil {
		panic(err)
	}
	return b
}

func (u *updateStorageDomainParams) WithCriticalSpaceActionBlocker(percent int64) (BuildableUpdateStorageDomainParams, error) {
	if percent < 0 || percent > 100 {
		return nil, newError(EBadArgument, "critical space action blocker must be between 0 and 100")
	}
	u.criticalSpaceActionBlocker = &percent
	return u, nil
}

func (u *updateStorageDomainParams) MustWithCriticalSpaceActionBlocker(percent int64) BuildableUpdateStorageDomainParams {
	b, err := u.WithCriticalSpaceActionBlocker(percent)
	if err != nil {
		panic(err)
	}
	return b
}

func (u *updateStorageDomainParams) WithWarningLowSpaceIndicator(percent int64) (BuildableUpdateStorageDomainParams, error) {
	if percent < 0 || percent > 100 {
		return nil, newError(EBadArgument, "warning low space indicator must be between 0 and 100")
	}
	u.warningLowSpaceIndicator = &percent
	return u, nil
}

func (u *updateStorageDomainParams) MustWithWarningLowSpaceIndicator(percent int64) BuildableUpdateStorageDomainParams {
	b, err := u.WithWarningLowSpaceIndicator(percent)
	if err != nil {
		panic(err)
	}
	return b
}

// RemoveStorageDomainParams contains options for removing a storage domain.
type RemoveStorageDomainParams interface {
	Destroy() *bool
	Format() *bool
}

// BuildableRemoveStorageDomainParams is a buildable version of RemoveStorageDomainParams.
type BuildableRemoveStorageDomainParams interface {
	RemoveStorageDomainParams

	WithDestroy(destroy bool) (BuildableRemoveStorageDomainParams, error)
	MustWithDestroy(destroy bool) BuildableRemoveStorageDomainParams
	WithFormat(format bool) (BuildableRemoveStorageDomainParams, error)
	MustWithFormat(format bool) BuildableRemoveStorageDomainParams
}

// NewRemoveStorageDomainParams creates a buildable set of RemoveStorageDomainParams.
func NewRemoveStorageDomainParams() BuildableRemoveStorageDomainParams {
	return &removeStorageDomainParams{}
}

type removeStorageDomainParams struct {
	destroy *bool
	format  *bool
}

func (r *removeStorageDomainParams) Destroy() *bool {
	return r.destroy
}

func (r *removeStorageDomainParams) Format() *bool {
	return r.format
}

func (r *removeStorageDomainParams) WithDestroy(destroy bool) (BuildableRemoveStorageDomainParams, error) {
	r.destroy = &destroy
	return r, nil
}

func (r *removeStorageDomainParams) MustWithDestroy(destroy bool) BuildableRemoveStorageDomainParams {
	b, err := r.WithDestroy(destroy)
	if err != nil {
		panic(err)
	}
	return b
}

func (r *removeStorageDomainParams) WithFormat(format bool) (BuildableRemoveStorageDomainParams, error) {
	r.format = &format
	return r, nil
}

func (r *removeStorageDomainParams) MustWithFormat(format bool) BuildableRemoveStorageDomainParams {
	b, err := r.WithFormat(format)
	if err != nil {
		panic(err)
	}
	return b
}
