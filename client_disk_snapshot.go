package ovirtclient

type DiskSnapshotClient interface {
	// StartCreateSnapshot starts the creation of a disk snapshot. This takes time, so the returned DiskSnapshotProgress
	// object can be used to monitor the progress.
	StartCreateSnapshot(diskID string, retries ...RetryStrategy) (DiskSnapshotProgress, error)
	// CreateSnapshot starts the creation of a disk snapshot, and then waits for its completion. It may return an error
	// AND also a snapshot that is still in progress.
	CreateSnapshot(diskID string, retries ...RetryStrategy) (DiskSnapshot, error)
	// ListSnapshots lists all snapshots for a disk.
	ListSnapshots(diskID string, retries ...RetryStrategy) ([]DiskSnapshot, error)
	// GetSnapshot returns a single snapshot for the specified disk.
	GetSnapshot(diskID string, snapshotID string, retries ...RetryStrategy) (DiskSnapshot, error)
	// RemoveSnapshot removes a specific snapshot.
	RemoveSnapshot(diskID string, snapshotID string, retries ...RetryStrategy) error

	// StartDownloadSnapshot starts the download of the image file the specified snapshot.
	// The caller can then wait for the initialization using the Initialized() call:
	//
	//     <-download.Initialized()
	//
	// Alternatively, the downloader can use the Read() function to wait for the download to become available
	// and then read immediately.
	//
	// The caller MUST close the returned reader, otherwise the disk will remain locked in the oVirt engine.
	// The passed context is observed only for the initialization phase.
	StartDownloadSnapshot(diskID string, snapshotID string, retries ...RetryStrategy) (ImageDownload, error)

	// DownloadSnapshot runs StartDownloadSnapshot, then waits for the download to be ready before returning the reader.
	// The caller MUST close the ImageDownloadReader in order to properly unlock the disk in the oVirt engine.
	DownloadSnapshot(diskID string, snapshotID string, retries ...RetryStrategy) (ImageDownloadReader, error)
}

// DiskSnapshotProgress is an object that lets you monitor the progress of a snapshot.
type DiskSnapshotProgress interface {
	// DiskSnapshot returns the disk snapshot as it was during the last update call.
	DiskSnapshot() DiskSnapshot
	// Wait waits until the disk snapshot is complete and returns when it is done. It returns the created snapshot and
	// an error if one happened.
	Wait(retries ...RetryStrategy) (DiskSnapshot, error)
}

// DiskSnapshotData is the data portion of the disk snapshot.
type DiskSnapshotData interface {
	// ID returns the ID of the snapshot. This can be used in conjunction with the DiskID to work on the snapshot.
	ID() string
	// DiskID returns the ID of the disk this snapshot was made from and belongs to.
	DiskID() string
}

// DiskSnapshot declares the data access functions, as well as convenience functions to handle the snapshots.
type DiskSnapshot interface {
	DiskSnapshotData

	GetDisk(retries ...RetryStrategy) (Disk, error)
	// StartDownload starts the download of the image file the current snapshot.
	// The caller can then wait for the initialization using the Initialized() call:
	//
	//     <-download.Initialized()
	//
	// Alternatively, the downloader can use the Read() function to wait for the download to become available
	// and then read immediately.
	//
	// The caller MUST close the returned reader, otherwise the disk will remain locked in the oVirt engine.
	// The passed context is observed only for the initialization phase.
	StartDownload(
		format ImageFormat,
		retries ...RetryStrategy,
	) (ImageDownload, error)

	// Download runs StartDownload, then waits for the download to be ready before returning the reader.
	// The caller MUST close the ImageDownloadReader in order to properly unlock the disk in the oVirt engine.
	Download(
		format ImageFormat,
		retries ...RetryStrategy,
	) (ImageDownloadReader, error)
}
