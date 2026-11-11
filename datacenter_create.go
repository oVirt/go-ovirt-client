package ovirtclient

import (
	"fmt"

	ovirtsdk4 "github.com/ovirt/go-ovirt/v4"
)

func (o *oVirtClient) CreateDatacenter(
	name string,
	params CreateDatacenterParams,
	retries ...RetryStrategy,
) (result Datacenter, err error) {
	retries = defaultRetries(retries, defaultWriteTimeouts(o))
	if params == nil {
		params = NewCreateDatacenterParams()
	}
	if name == "" {
		return nil, newError(EBadArgument, "name cannot be empty for datacenter creation")
	}

	err = retry(
		fmt.Sprintf("creating datacenter %s", name),
		o.logger,
		retries,
		func() error {
			builder := ovirtsdk4.NewDataCenterBuilder().Name(name)
			if desc := params.Description(); desc != nil {
				builder.Description(*desc)
			}
			if comment := params.Comment(); comment != nil {
				builder.Comment(*comment)
			}
			if local := params.Local(); local != nil {
				builder.Local(*local)
			}
			if qm := params.QuotaMode(); qm != nil {
				builder.QuotaMode(ovirtsdk4.QuotaModeType(*qm))
			}

			response, e := o.conn.SystemService().DataCentersService().Add().DataCenter(builder.MustBuild()).Send()
			if e != nil {
				return e
			}
			sdkDC, ok := response.DataCenter()
			if !ok {
				return newFieldNotFound("response from datacenter creation", "data_center")
			}
			result, err = convertSDKDatacenter(sdkDC, o)
			return err
		},
	)
	return result, err
}

func (m *mockClient) CreateDatacenter(
	name string,
	params CreateDatacenterParams,
	_ ...RetryStrategy,
) (result Datacenter, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if params == nil {
		params = NewCreateDatacenterParams()
	}
	if name == "" {
		return nil, newError(EBadArgument, "name cannot be empty for datacenter creation")
	}

	for _, existing := range m.dataCenters {
		if existing.name == name {
			return nil, newError(EConflict, "datacenter with name %s already exists", name)
		}
	}

	local := false
	if l := params.Local(); l != nil {
		local = *l
	}

	id := DatacenterID(m.GenerateUUID())
	dc := &datacenterWithClusters{
		datacenter: datacenter{
			client:      m,
			id:          id,
			name:        name,
			description: params.Description(),
			comment:     params.Comment(),
			local:       local,
			quotaMode:   params.QuotaMode(),
		},
		clusters: []ClusterID{},
	}
	m.dataCenters[id] = dc
	return dc, nil
}
