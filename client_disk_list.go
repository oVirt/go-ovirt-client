package ovirtclient

func (o *oVirtClient) ListDisks() ([]Disk, error) {
	response, err := o.conn.SystemService().DisksService().List().Send()
	if err != nil {
		return nil, wrap(
			err,
			EUnidentified,
			"failed to list disks",
		)
	}
	sdkDisks, ok := response.Disks()
	if !ok {
		return nil, nil
	}
	result := make([]Disk, len(sdkDisks.Slice()))
	for i, sdkDisk := range sdkDisks.Slice() {
		disk, err := convertSDKDisk(sdkDisk)
		if err != nil {
			return nil, wrap(err, EBug, "failed to convert disk item %d", i)
		}
		result[i] = disk
	}
	return result, nil
}
