package ovirtclient

import (
	"sync"
)

// diskWithData adds the ability to store the data directly in the disk for mocking purposes.
type diskWithData struct {
	disk
	lock *sync.Mutex
	data []byte
}

func (d *diskWithData) Lock() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.status != DiskStatusOK {
		return newError(EDiskLocked, "disk %s is %s", d.id, d.status)
	}
	d.status = DiskStatusLocked
	return nil
}

func (d *diskWithData) Unlock() {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.status = DiskStatusOK
}

func (d *diskWithData) WithAlias(alias *string) *diskWithData {
	return &diskWithData{
		disk{
			client:          d.client,
			id:              d.id,
			alias:           *alias,
			provisionedSize: d.provisionedSize,
			format:          d.format,
			storageDomainID: d.storageDomainID,
			status:          d.status,
			totalSize:       d.totalSize,
			sparse:          d.sparse,
		},
		d.lock,
		d.data,
	}
}
