package ovirtclient

func (o *oVirtClient) ListClusters() ([]Cluster, error) {
	clustersResponse, err := o.conn.SystemService().ClustersService().List().Send()
	if err != nil {
		return nil, wrap(
			err,
			EUnidentified,
			"failed to list oVirt clusters",
		)
	}
	sdkClusters, ok := clustersResponse.Clusters()
	if !ok {
		return nil, nil
	}
	clusters := make([]Cluster, len(sdkClusters.Slice()))
	for i, sdkCluster := range sdkClusters.Slice() {
		clusters[i], err = convertSDKCluster(sdkCluster)
		if err != nil {
			return nil, wrap(err, EBug, "failed to convert cluster during cluster listing item %d", i)
		}
	}
	return clusters, nil
}
