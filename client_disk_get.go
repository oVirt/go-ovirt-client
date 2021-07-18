package ovirtclient

func (o *oVirtClient) GetDisk(diskID string) (Disk, error) {
	response, err := o.conn.SystemService().DisksService().DiskService(diskID).Get().Send()
	if err != nil {
		return nil, wrap(
			err,
			EUnidentified,
			"failed to fetch disk %s",
			diskID,
		)
	}
	sdkDisk, ok := response.Disk()
	if !ok {
		return nil, wrap(
			err,
			ENotFound,
			"disk %s response did not contain a disk",
			diskID,
		)
	}
	disk, err := convertSDKDisk(sdkDisk)
	if err != nil {
		return nil, wrap(
			err,
			EBug,
			"failed to convert disk %s",
			diskID,
		)
	}
	return disk, nil
}
