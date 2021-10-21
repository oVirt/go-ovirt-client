package ovirtclient

import (
	ovirtsdk4 "github.com/ovirt/go-ovirt"
	"sync"
)

//go:generate go run scripts/rest.go -i "Template" -n "template"

// TemplateClient represents the portion of the client that deals with VM templates.
type TemplateClient interface {
	ListTemplates(retries ...RetryStrategy) ([]Template, error)
	GetTemplate(id string, retries ...RetryStrategy) (Template, error)
	CopyTemplateDiskToStorageDomain(diskId string,
		storageDomainId string,
		retries ...RetryStrategy) (result Disk, err error)
}

// Template is a set of prepared configurations for VMs.
type Template interface {
	// ID returns the identifier of the template. This is typically a UUID.
	ID() string
	// Name is the human-readable name for the template.
	Name() string
	// Description is a longer description for the template.
	Description() string
}

func convertSDKTemplate(sdkTemplate *ovirtsdk4.Template, client Client) (Template, error) {
	id, ok := sdkTemplate.Id()
	if !ok {
		return nil, newError(EFieldMissing, "template does not contain ID")
	}
	name, ok := sdkTemplate.Name()
	if !ok {
		return nil, newError(EFieldMissing, "template does not contain a name")
	}
	description, ok := sdkTemplate.Description()
	if !ok {
		return nil, newError(EFieldMissing, "template does not contain a description")
	}
	return &template{
		client:      client,
		id:          id,
		name:        name,
		description: description,
	}, nil
}

type template struct {
	client      Client
	id          string
	name        string
	description string
}

func (t template) ID() string {
	return t.id
}

func (t template) Name() string {
	return t.name
}

func (t template) Description() string {
	return t.description
}


type templateDiskCopyWait struct {
	client        *oVirtClient
	disk          Disk
	template	  Template
	correlationID string
	lock          *sync.Mutex
}

func (d *templateDiskCopyWait) Disk() Disk {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.disk
}

func (d *templateDiskCopyWait) Wait(retries ...RetryStrategy) (Disk, error) {
	retries = defaultRetries(retries, defaultWriteTimeouts())
	if err := d.client.waitForJobFinished(d.correlationID, retries); err != nil {
		return d.disk, err
	}

	d.lock.Lock()
	diskID := d.disk.ID()
	d.lock.Unlock()

	disk, err := d.client.GetDisk(diskID, retries...)

	d.lock.Lock()
	defer d.lock.Unlock()
	if disk != nil {
		d.disk = disk
	}
	return disk, err
}

