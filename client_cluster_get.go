package ovirtclient

func (o *oVirtClient) GetCluster(id string) (cluster Cluster, err error) {
	response, err := o.conn.SystemService().ClustersService().ClusterService(id).Get().Send()
	if err != nil {
		return nil, wrap(err, "failed to fetch cluster ID %s", id)
	}
	sdkCluster, ok := response.Cluster()
	if !ok {
		return nil, newError(
			ENotFound,
			"no cluster returned when getting cluster ID %s",
			id,
		)
	}
	cluster, err = convertSDKCluster(sdkCluster)
	if err != nil {
		return nil, wrap(
			err,
			EBug,
			"failed to convert cluster %s",
			id,
		)
	}
	return cluster, nil
}
