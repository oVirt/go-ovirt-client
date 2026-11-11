package ovirtclient

import (
	"fmt"

	ovirtsdk4 "github.com/ovirt/go-ovirt/v4"
)

func (o *oVirtClient) UpdateDatacenter(
	id DatacenterID,
	params UpdateDatacenterParams,
	retries ...RetryStrategy,
) (result Datacenter, err error) {
	retries = defaultRetries(retries, defaultWriteTimeouts(o))
	if params == nil {
		params = NewUpdateDatacenterParams()
	}

	err = retry(
		fmt.Sprintf("updating datacenter %s", id),
		o.logger,
		retries,
		func() error {
			builder := ovirtsdk4.NewDataCenterBuilder().Id(string(id))
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
			if qm := params.QuotaMode(); qm != nil {
				builder.QuotaMode(ovirtsdk4.QuotaModeType(*qm))
				hasChanges = true
			}
			if !hasChanges {
				return nil
			}

			response, e := o.conn.SystemService().DataCentersService().DataCenterService(string(id)).Update().DataCenter(builder.MustBuild()).Send()
			if e != nil {
				return e
			}
			sdkDC, ok := response.DataCenter()
			if !ok {
				return newFieldNotFound("response from datacenter update", "data_center")
			}
			result, err = convertSDKDatacenter(sdkDC, o)
			return err
		},
	)
	if err == nil && result == nil {
		return o.GetDatacenter(id, retries...)
	}
	return result, err
}

func (m *mockClient) UpdateDatacenter(
	id DatacenterID,
	params UpdateDatacenterParams,
	_ ...RetryStrategy,
) (result Datacenter, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	dc, ok := m.dataCenters[id]
	if !ok {
		return nil, newError(ENotFound, "datacenter with ID %s not found", id)
	}

	name := dc.name
	if newName := params.Name(); newName != nil {
		if *newName == "" {
			return nil, newError(EBadArgument, "name cannot be empty")
		}
		for _, existing := range m.dataCenters {
			if existing.id != id && existing.name == *newName {
				return nil, newError(EConflict, "datacenter with name %s already exists", *newName)
			}
		}
		name = *newName
	}

	description := dc.description
	if newDesc := params.Description(); newDesc != nil {
		description = newDesc
	}

	comment := dc.comment
	if newComment := params.Comment(); newComment != nil {
		comment = newComment
	}

	quotaMode := dc.quotaMode
	if newQM := params.QuotaMode(); newQM != nil {
		quotaMode = newQM
	}

	dc.name = name
	dc.description = description
	dc.comment = comment
	dc.quotaMode = quotaMode

	return dc, nil
}
