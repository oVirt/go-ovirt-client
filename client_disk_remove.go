package ovirtclient

func (o *oVirtClient) RemoveDisk(diskID string) error {
	if _, err := o.conn.SystemService().DisksService().DiskService(diskID).Remove().Send(); err != nil {
		return wrap(
			err,
			EUnidentified,
			"failed to remove disk %s",
			diskID,
		)
	}
	return nil
}
